/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package service provides relay service.
package service

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"runtime"

	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/file"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/handler"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/router/callback"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/router/download"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/blog"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/version"
	"github.com/gin-gonic/gin"
)

const (
	messagetrackerDirName     = "messagetracker"
	fileManagerStorageDirName = "filemanager"
)

// Service defines a server that provides relay service.
type Service struct {
	// conf holds the configuration for the service.
	conf *config.RelayService

	// ctx is used to control the service lifecycle (cancellation and timeouts).
	ctx contextx.IContext

	// cancelFunc is used to cancel the service and all associated operations.
	cancelFunc context.CancelFunc

	// Note: Cap is initialized in the Start() and could not be used in other package.
	// Cap is the capability of the service.
	Cap *options.Capability

	// router is the entry point of the service, routing requests to different capabilities.
	servers []*restserver.Server

	// instance is the discover instance of the service.
	instance discover.Instance
}

// NewService creates a new relay service.
// nolint: funlen
func NewService(conf *config.RelayService) (*Service, error) {
	svc := &Service{
		conf: conf,
		Cap: &options.Capability{
			Logger: blog.GlobalLogger{},
		},
		instance: discover.NewInstance(string(discover.ServiceNameRelay), nil),
	}

	svc.ctx, svc.cancelFunc = contextx.WithCancel(contextx.NewContext(context.Background(), map[string]any{}))

	if err := svc.initialCapability(); err != nil {
		return nil, fmt.Errorf("failed to initialize capability: %w", err)
	}

	if err := svc.registerRestServer(); err != nil {
		return nil, fmt.Errorf("failed to register http rest server: %w", err)
	}

	return svc, nil
}

func (svc *Service) initialCapability() error {
	// initial messager
	svc.Cap.Messager = relayhandler.NewClientMessager(relayhandler.ClientMessagerConfig{
		PluginVersion:          version.Version().Version,
		DomainSocketPath:       svc.conf.Plugin.MessageDomainSocketPath,
		LocalSocketPort:        svc.conf.Plugin.MessageLocalSocketPort,
		Logger:                 svc.Cap.Logger,
		MessageTrackerFullPath: filepath.Join(svc.conf.RelayWorkspaceFileGroup.FullPath, messagetrackerDirName),
		PluginName:             svc.conf.PluginName,
	})

	// initial file manager
	svc.Cap.FileManager = file.NewFileManager(
		svc.ctx,
		filepath.Join(svc.conf.RelayWorkspaceFileGroup.FullPath, fileManagerStorageDirName),
		svc.Cap.Logger)
	if svc.Cap.FileManager == nil {
		return fmt.Errorf("failed to init file manager")
	}

	// initial client handler
	clientHandler := handler.NewClientHandler(svc.Cap.FileManager,
		svc.Cap.Messager,
		svc.conf,
		svc.Cap.Logger)
	if clientHandler == nil {
		return fmt.Errorf("failed to init client handler")
	}

	// register server push event handlers
	dispatcher := svc.Cap.Messager.EventDispatcher()
	dispatcher.RegisterHandler(protoRelay.ServerPushEventTypeCheckPkgState, clientHandler.CheckPkgStats)
	dispatcher.RegisterHandler(protoRelay.ServerPushEventTypeNotifyReceive, clientHandler.StoragePkg)
	dispatcher.RegisterHandler(protoRelay.ServerPushEventTypeDetectInfoBySSH, clientHandler.DetectInfoBySSH)
	dispatcher.RegisterHandler(protoRelay.ServerPushEventTypeInstallBySSH, clientHandler.InstallPagentBySSH)
	dispatcher.RegisterHandler(protoRelay.ServerPushEventTypeDetectInfoByWMI, clientHandler.DetectInfoByWMI)
	dispatcher.RegisterHandler(protoRelay.ServerPushEventTypeInstallByWMI, clientHandler.InstallPagentByWMI)

	return nil
}

func (svc *Service) registerRestServer() error {
	if err := svc.registerCallbackServer(); err != nil {
		return fmt.Errorf("failed to register callback server: %w", err)
	}

	if err := svc.registerFileServer(); err != nil {
		return fmt.Errorf("failed to register file server: %w", err)
	}

	return nil
}

func (svc *Service) registerCallbackServer() error {
	requestIDSetter := restserver.NewRequestIDSetter()
	tenantIDSetter := restserver.NewTenantIDSetter()

	server := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:            string(discover.EndpointNameRelayCallback),
			IP:              svc.conf.CallbackServer.BindIP,
			Port:            svc.conf.CallbackServer.Port,
			LogWriter:       loggerWriterAdaptor{},
			RequestIDSetter: requestIDSetter,
			TenantIDSetter:  tenantIDSetter,
		},
		restserver.WithPing(),
		withCallbackServer(svc.Cap),
	)

	if server == nil {
		return fmt.Errorf("failed to create callback server")
	}

	svc.servers = append(svc.servers, server)
	svc.instance.Update(discover.EndpointNameRelayCallback, discover.Endpoint{
		IPV4: svc.conf.CallbackServer.AdvertiseIPV4,
		Port: svc.conf.CallbackServer.Port,
	})

	return nil
}

func (svc *Service) registerFileServer() error {
	requestIDSetter := restserver.NewRequestIDSetter()
	tenantIDSetter := restserver.NewTenantIDSetter()

	server := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:            string(discover.EndpointNameRelayFile),
			IP:              svc.conf.DownloadServer.BindIP,
			Port:            svc.conf.DownloadServer.Port,
			LogWriter:       loggerWriterAdaptor{},
			RequestIDSetter: requestIDSetter,
			TenantIDSetter:  tenantIDSetter,
		},
		restserver.WithPing(),
		withDownload(svc.Cap),
	)

	if server == nil {
		return fmt.Errorf("failed to create file server")
	}

	svc.servers = append(svc.servers, server)
	svc.instance.Update(discover.EndpointNameRelayFile, discover.Endpoint{
		IPV4: svc.conf.DownloadServer.AdvertiseIPV4,
		Port: svc.conf.DownloadServer.Port,
	})

	return nil
}

// withCallbackServer load callback api.
func withCallbackServer(capability *options.Capability) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		callback.Load(rg, capability)
	}
}

// withDownload load download.
func withDownload(capability *options.Capability) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		download.Load(rg, capability)
	}
}

// loggerWriterAdaptor implements rest.LoggerWriter.
type loggerWriterAdaptor struct{}

func (l loggerWriterAdaptor) InfoWriter() io.Writer {
	return blog.WriterInfo{}
}

func (l loggerWriterAdaptor) ErrorWriter() io.Writer {
	return blog.WriterError{}
}

// Start starts the relay service.
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
