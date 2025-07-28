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
	"fmt"
	"io"
	"runtime"

	"github.com/TencentBlueKing/bk-nodemgr/internal/application/authidentity"
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/frontsetting"
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	apiv3 "github.com/TencentBlueKing/bk-nodemgr/internal/application/router/api-v3"
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/router/healthz"
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/router/web"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/blog"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/etcddiscover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backend"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/bklogin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
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

// NewService creates a new application service.
func NewService(conf *config.ApplicationService) (*Service, error) {
	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("failed to new service: %w", err)
	}

	svc := &Service{
		conf: conf,
		Cap: &options.Capability{
			Logger: blog.GlobalLogger{},
		},
	}

	svc.ctx, svc.cancelFunc = context.WithCancel(context.Background())

	var err error

	svc.Cap.DiscoverProvider = etcddiscover.NewProviderEtcd(&conf.Etcd,
		etcddiscover.WithLogger(svc.Cap.Logger),
		etcddiscover.WithWatch(discover.ServiceNameBackend, discover.ServiceNameFile),
	)

	svc.Cap.BackendHandler, err = newBackendHandler(svc.conf.Backend)
	if err != nil {
		return nil, fmt.Errorf("failed to new service: %w", err)
	}

	svc.Cap.FileHandler, err = newFileHandler(svc.Cap.DiscoverProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to new service: %w", err)
	}

	svc.Cap.AuthIdentity, err = svc.newBKTicketAuthIdentity(conf.BKLogin)
	if err != nil {
		return nil, fmt.Errorf("failed to new service: %w", err)
	}

	svc.Cap.FrontSetting = frontsetting.NewFrontSetting(conf.Front.BKLoginURL, conf.Front.BKSharedResBaseJsUrl, conf.Front.SiteURL)

	if err := svc.registerRestServer(conf); err != nil {
		return nil, fmt.Errorf("failed to new service: %w", err)
	}

	return svc, nil
}

func (svc *Service) registerRestServer(conf *config.ApplicationService) error {
	bkloginHandler, err := newBKLoginHandler(conf.BKLogin, svc.Cap.Logger)
	if err != nil {
		return fmt.Errorf("failed to register rest server: %w", err)
	}

	authIdentity := &authidentity.BKTicketAuthIdentity{
		BKLoginHandler: bkloginHandler,
	}

	httpServer := rest.NewServer(
		svc.ctx,
		rest.ServerOptions{
			Name:         string(discover.EndpointNameApplicationBasic),
			IP:           conf.HTTPServer.BindIP,
			Port:         conf.HTTPServer.Port,
			LogWriter:    loggerWriter{},
			AuthIdentity: authIdentity,
			StaticOptions: rest.NewStaticOptions(conf.HTTPServer.StaticDir).
				WithHTMLs("index.html").
				WithDirs("assets").
				WithDirs("images").
				WithFiles("bk.svg", "favicon.png", "nodeman.png"),
		},
		rest.WithPing(),
		withHealthz(svc.Cap),
		withMetrics(svc.Cap),
		withWeb(svc.Cap),
		withAPIV3(svc.Cap),
	)

	svc.servers = append(svc.servers, httpServer)

	return nil
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

// newFileHandler creates a new file handler.
func newFileHandler(discov discover.Discover) (file.IHandler, error) {
	httpClient, err := client.NewClient(&ssl.TLSConfig{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, err
	}

	clientCap := &client.Capability{
		Client: httpClient,
		Discover: discovery.NewServiceDiscovery(
			discov,
			discover.ServiceNameFile,
			discover.EndpointNameFileAdmin),
		ToleranceLatencyTime: client.ToleranceLatencyTimeDefault,
		MetricOpts:           client.MetricOption{},
		Logger:               logger.LoggerDefault{},
	}

	return file.New(clientCap, &file.Config{})
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

func (svc *Service) newBKTicketAuthIdentity(conf config.BKLogin) (rest.AuthIdentity, error) {
	bkloginHandler, err := newBKLoginHandler(conf, svc.Cap.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to register rest server: %w", err)
	}

	authIdentity := &authidentity.BKTicketAuthIdentity{
		BKLoginHandler: bkloginHandler,
	}

	return authIdentity, nil
}

// newBKLoginHandler
func newBKLoginHandler(conf config.BKLogin, logger logger.Logger) (bklogin.IHandler, error) {
	httpClient, err := client.NewClient(&ssl.TLSConfig{
		InsecureSkipVerify: conf.TLS.InsecureSkipVerify,
		CertFile:           conf.TLS.CertFile,
		KeyFile:            conf.TLS.KeyFile,
		CAFile:             conf.TLS.CAFile,
		Password:           conf.TLS.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to new bklogin handler: %v", err)
	}

	clientCap := &client.Capability{
		Client:               httpClient,
		Discover:             discovery.NewDiscovery(DiscoveryNameApigw, []string{conf.LoginURL}),
		ToleranceLatencyTime: client.ToleranceLatencyTimeDefault,
		MetricOpts:           client.MetricOption{},
		Logger:               logger,
	}

	bkloginHandler, err := bklogin.New(clientCap, &bklogin.Config{LoginURL: conf.LoginURL}, bklogin.WithLogger(logger))
	if err != nil {
		return nil, fmt.Errorf("failed to new bklogin handler: %w", err)
	}

	return bkloginHandler, nil
}

// Start starts the application service.
func (svc *Service) Start() error {
	runtime.GOMAXPROCS(runtime.NumCPU())

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

// GracefulShutdown gracefully shuts down the application service.
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
