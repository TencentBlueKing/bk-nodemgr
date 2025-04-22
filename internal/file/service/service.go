/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package service provides backend service.
package service

import (
	"context"
	"errors"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/filex/local"
	"io"
	"runtime"

	"github.com/TencentBlueKing/bk-nodemgr/internal/file/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/router/download"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/router/healthz"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/router/upload"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/blog"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/etcddiscover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// DiscoveryNameApigw defines the name of apigateway discovery.
	DiscoveryNameApigw = "apigateway"
)

// Service defines a server that provides file services.
// It manages the configuration, lifecycle, and various capabilities (e.g., cmdb, topo storage).
// The service's capabilities are accessed through its 'cap' field, while 'router' is used to route requests.
type Service struct {
	// conf holds the configuration for the service.
	conf *config.FileService

	// ctx is used to control the service lifecycle (cancellation and timeouts).
	ctx context.Context

	// cancelFunc is used to cancel the service and all associated operations.
	cancelFunc context.CancelFunc

	// router is the entry point of the service, routing requests to different capabilities.
	servers []*rest.Server

	// Note: Cap is initialized in the Start() and could not be used in other package.
	// Cap is the capability of the service.
	Cap *options.Capability

	// instance is the discover instance of the service.
	instance discover.Instance
}

// NewService creates a new file service.
func NewService(conf *config.FileService) (*Service, error) {
	svc := &Service{
		conf: conf,
		Cap: &options.Capability{
			Logger: blog.GlobalLogger{},
		},
		instance: discover.NewInstance("file", nil),
	}

	svc.ctx, svc.cancelFunc = context.WithCancel(context.Background())

	var err error
	svc.Cap.AgentFileGroup, err = local.NewLocalDir(conf.AgentFileGroup.FullPath, svc.Cap.Logger)
	if err != nil {
		return nil, errors.New("init agent file group failed")
	}

	svc.Cap.ProxyFileGroup, err = local.NewLocalDir(conf.ProxyFileGroup.FullPath, svc.Cap.Logger)
	if err != nil {
		return nil, errors.New("init proxy file group failed")
	}

	svc.Cap.DiscoverProvider = etcddiscover.NewProviderEtcd(&conf.Etcd,
		etcddiscover.WithLogger(svc.Cap.Logger),
		etcddiscover.WithWatch(discover.ServiceNameBackend, discover.ServiceNameFile),
	)

	httpServer := rest.NewServer(
		svc.ctx,
		rest.ServerOptions{
			Name:      string(discover.EndpointNameFileBasic),
			IP:        conf.HTTPServer.BindIP,
			Port:      conf.HTTPServer.Port,
			LogWriter: loggerWriterAdaptor{},
		},
		rest.WithPing(),
		withHealthz(svc.Cap),
		withMetrics(svc.Cap),
		withDownload(svc.Cap),
	)
	svc.servers = append(svc.servers, httpServer)
	svc.instance.Update(discover.EndpointNameFileBasic, discover.Endpoint{
		IPV4: conf.HTTPServer.AdvertiseIPV4,
		IPV6: conf.HTTPServer.AdvertiseIPV6,
		Port: conf.HTTPServer.Port,
	})

	adminServer := rest.NewServer(
		svc.ctx,
		rest.ServerOptions{
			Name:      string(discover.EndpointNameFileAdmin),
			IP:        conf.AdminServer.BindIP,
			Port:      conf.AdminServer.Port,
			LogWriter: loggerWriterAdaptor{},
		},
		rest.WithPing(),
		withHealthz(svc.Cap),
		withMetrics(svc.Cap),
		withUpload(svc.Cap),
	)
	svc.servers = append(svc.servers, adminServer)

	return svc, nil
}

// loggerWriterAdaptor implements rest.LoggerWriter.
type loggerWriterAdaptor struct{}

// InfoWriter returns the writer for logging info.
func (l loggerWriterAdaptor) InfoWriter() io.Writer {
	return blog.WriterInfo{}
}

// ErrorWriter returns the writer for logging errors.
func (l loggerWriterAdaptor) ErrorWriter() io.Writer {
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

// withDownload load download.
func withDownload(capability *options.Capability) rest.OptionFunc {
	return func(rg *gin.RouterGroup) {
		download.Load(rg, capability)
	}
}

// withUpload load upload.
func withUpload(capability *options.Capability) rest.OptionFunc {
	return func(rg *gin.RouterGroup) {
		upload.Load(rg, capability)
	}
}

// Start starts the file service.
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

	// after all servers brings up, register the instance into discover provider.
	if err := svc.Cap.DiscoverProvider.Register(discover.ServiceNameFile, svc.instance); err != nil {
		blog.Errorf("failed to register instance, err: %v", err)
		return err
	}

	// wait until all servers stopped or application error.
	if err := gp.Wait(); err != nil {
		blog.Errorf("failed to start servers, err: %v", err)
		return err
	}

	return nil
}

// GracefulShutdown ...
func (svc *Service) GracefulShutdown() error {
	if svc.ctx == nil || svc.cancelFunc == nil {
		return errors.New("service is not running")
	}

	defer svc.cancelFunc()

	blog.CloseLogs()

	return nil
}
