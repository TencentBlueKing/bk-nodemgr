/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package service provides application service.
package service

import (
	"context"
	"errors"
	"io"
	"runtime"

	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	apiv3 "github.com/TencentBlueKing/bk-nodemgr/internal/application/router/api-v3"
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/router/healthz"
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/router/web"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/blog"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/ssl"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backend"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// DiscoveryNameApigw defines the name of apigateway discovery.
	DiscoveryNameApigw = "apigateway"
)

// Service defines a server that provides application services.
// It provides a website for user to operate with nodeman.
type Service struct {
	// conf holds the configuration for the service.
	conf *config.ApplicationService

	// ctx is used to control the service lifecycle (cancellation and timeouts).
	ctx context.Context

	// cancelFunc is used to cancel the service and all associated operations.
	cancelFunc context.CancelFunc

	// router is the entry point of the service, routing requests to different capabilities.
	servers []*rest.Server

	// Note: Cap is initialized in the Start() and could not be used in other package.
	// Cap is the capability of the service.
	Cap *options.Capability
}

const (
	// RouterNameHTTPServer defines the name of http server router.
	RouterNameHTTPServer = "http-server"
)

// NewService creates a new application service.
func NewService(conf *config.ApplicationService) (*Service, error) {
	svc := &Service{
		conf: conf,
		Cap: &options.Capability{
			Logger: blog.GlobalLogger{},
		},
	}

	svc.ctx, svc.cancelFunc = context.WithCancel(context.Background())

	var err error

	svc.Cap.BackendHandler, err = newBackendHandler(svc.conf.Backend)
	if err != nil {
		return nil, err
	}

	svc.registerRestServer(conf)

	return svc, nil
}

func (svc *Service) registerRestServer(conf *config.ApplicationService) {
	httpServer := rest.NewServer(svc.ctx, RouterNameHTTPServer, conf.HTTPServer.BindIP, conf.HTTPServer.Port,
		loggerWriter{},
		rest.NewStaticOptions(conf.HTTPServer.StaticDir).
			WithHTMLs("index.html").
			WithDirs("assets").
			WithFiles("bk.svg", "favicon.png", "nodeman.png"),
		rest.WithPing(),
		withHealthz(svc.Cap),
		withMetrics(svc.Cap),
		withWeb(svc.Cap),
		withAPIV3(svc.Cap),
	)

	svc.servers = append(svc.servers, httpServer)
}

// loggerWriter implements rest.LoggerWriter.
type loggerWriter struct{}

func (l loggerWriter) InfoWriter() io.Writer {
	return blog.WriterInfo{}
}

func (l loggerWriter) ErrorWriter() io.Writer {
	return blog.WriterError{}
}

// withHealthz load healthz.
func withHealthz(capability *options.Capability) rest.OptionFunc {
	return func(rg *gin.RouterGroup) {
		healthz.Load(rg, capability)
	}
}

// withMetrics load metrics.
func withMetrics(_ *options.Capability) rest.OptionFunc {
	return func(rg *gin.RouterGroup) {
		rg.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}
}

// withWeb load web page handler.
func withWeb(capability *options.Capability) rest.OptionFunc {
	return func(rg *gin.RouterGroup) {
		web.Load(rg, capability)
	}
}

// withApiV3 load api v3.
func withAPIV3(capability *options.Capability) rest.OptionFunc {
	return func(rg *gin.RouterGroup) {
		apiv3.Load(rg, capability)
	}
}

// newBackendHandler creates a new backend handler.
func newBackendHandler(conf config.BackendGateway) (backend.Handler, error) {
	apiGwHeaderSetter := newAPIGwHeaderSetter(&conf.APIGateway)
	apiGwClientCapability, err := newAPIGwClientCapability(&conf.APIGateway)
	if err != nil {
		return nil, err
	}

	apiGwClientCapability.Name = "backend"
	backendHandler, err := backend.New(apiGwClientCapability, &backend.Config{
		HeaderSetter: apiGwHeaderSetter,
	})
	if err != nil {
		return nil, err
	}

	return backendHandler, nil
}

// newAPIGwClientCapability creates a new api-gateway client capability.
func newAPIGwClientCapability(conf *config.APIGateway) (*client.Capability, error) {
	httpClient, err := client.NewClient(&ssl.TLSConfig{
		InsecureSkipVerify: conf.TLS.InsecureSkipVerify,
		CertFile:           conf.TLS.CertFile,
		KeyFile:            conf.TLS.KeyFile,
		CAFile:             conf.TLS.CAFile,
		Password:           conf.TLS.Password,
	})
	if err != nil {
		return nil, err
	}

	clientCap := &client.Capability{
		Client:               httpClient,
		Discover:             discovery.NewDiscovery(DiscoveryNameApigw, conf.Endpoints),
		ToleranceLatencyTime: client.ToleranceLatencyTimeDefault,
		MetricOpts:           client.MetricOption{},
		Logger:               blog.GlobalLogger{},
	}

	return clientCap, nil
}

// newAPIGwHeaderSetter creates a new api-gateway header setter.
func newAPIGwHeaderSetter(conf *config.APIGateway) apigw.HeaderSetter {
	return &apigw.Config{
		Endpoints:   conf.Endpoints,
		AppCode:     conf.AppCode,
		AppSecret:   conf.AppSecret,
		User:        conf.User,
		AuthMode:    apigw.AuthMode(conf.AuthMode),
		BkTicket:    conf.BkTicket,
		BkToken:     conf.BkToken,
		AccessToken: conf.AccessToken,
	}
}

// Start starts the application service.
func (svc *Service) Start() error {
	runtime.GOMAXPROCS(runtime.NumCPU())

	logConfig := blog.NewLogConfig()
	logConfig.LogDir = svc.conf.Log.Dir
	logConfig.LogMaxSizeMB = svc.conf.Log.MaxSizeMB
	logConfig.LogMaxNum = svc.conf.Log.MaxNum
	logConfig.Level = svc.conf.Log.Level
	logConfig.ToStdErr = svc.conf.Log.ToStdErr
	logConfig.AlsoToStdErr = svc.conf.Log.AlsoToStdErr
	blog.InitLogs(logConfig)

	if err := svc.Cap.Start(svc.ctx); err != nil {
		return err
	}

	// start servers
	gp := gopool.NewPool()
	for idx := range svc.servers {
		server := svc.servers[idx]

		// server start will block until server stop, so we need to run it in a goroutine.
		fn := func() error {
			blog.Infof("started server. name(%s), ip(%s), port(%d)", server.Name(), server.IP(), server.Port())

			if err := server.Start(); err != nil {
				return err
			}

			return nil
		}
		gp.Go(fn)
	}

	// wait until all servers stopped or application error.
	if err := gp.Wait(); err != nil {
		blog.Errorf("failed to start servers, err: %v", err)
		return err
	}

	return nil
}

func (svc *Service) GracefulShutdown() error {
	if svc.ctx == nil || svc.cancelFunc == nil {
		return errors.New("service is not running")
	}

	defer svc.cancelFunc()

	err := svc.Cap.GracefulShutdown()
	if err != nil {
		blog.Errorf("failed to shutdown capability, err: %v", err)
		return err
	}

	blog.CloseLogs()

	return nil
}
