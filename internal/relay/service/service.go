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
	"path/filepath"
	"runtime"

	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/file"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/handler"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/router/callback"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/router/download"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/router/healthz"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/version"
	"github.com/gin-gonic/gin"
)

const (
	relayInfoSvcName     = "relay-info"
	relayAdminSvcName    = "relay-admin"
	relayCallbackSvcName = "relay-callback"
	relayDownloadSvcName = "relay-download"

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

	// authIdentityMap is the map of auth identities.
	authIdentityMap map[config.AuthIdentity]restserver.IAuthIdentity
}

// NewService creates a new relay service.
// nolint: funlen
func NewService(conf *config.RelayService) (*Service, error) {
	svc := &Service{
		conf: conf,
		Cap:  &options.Capability{},
	}

	svc.ctx, svc.cancelFunc = contextx.WithCancel(contextx.New(context.Background()))

	if err := svc.initialStaticsConfigs(); err != nil {
		return nil, fmt.Errorf("failed to initialize static configs: %w", err)
	}

	if err := svc.initialCapability(); err != nil {
		return nil, fmt.Errorf("failed to initialize capability: %w", err)
	}

	if err := svc.registerRestServer(); err != nil {
		return nil, fmt.Errorf("failed to register http rest server: %w", err)
	}

	return svc, nil
}

func (svc *Service) initialCapability() error {
	var err error
	// initial message
	svc.Cap.Messager, err = relayhandler.NewClientMessager(relayhandler.ClientMessagerConfig{
		PluginVersion:          version.Version().Version,
		DomainSocketPath:       svc.conf.Plugin.MessageDomainSocketPath,
		LocalSocketPort:        svc.conf.Plugin.MessageLocalSocketPort,
		MessageTrackerFullPath: filepath.Join(svc.conf.RelayWorkspaceFileGroup.FullPath, messagetrackerDirName),
		PluginName:             string(svc.conf.PluginName),
	})
	if err != nil {
		return fmt.Errorf("failed to create messager: %w", err)
	}

	// initial file manager
	svc.Cap.FileManager, err = file.NewFileManager(
		svc.ctx,
		filepath.Join(svc.conf.RelayWorkspaceFileGroup.FullPath, fileManagerStorageDirName))
	if err != nil {
		return fmt.Errorf("failed to create file manager: %w", err)
	}

	// initial client handler
	clientHandler := handler.NewClientHandler(svc.Cap.FileManager,
		svc.Cap.Messager,
		svc.conf)

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

// nolint: unparam
func (svc *Service) initialStaticsConfigs() error {
	svc.authIdentityMap = map[config.AuthIdentity]restserver.IAuthIdentity{
		config.AuthIdentityNone: restserver.NewNodeAuthIdentity(),
	}

	return nil
}

func (svc *Service) registerRestServer() error {
	if err := svc.registerInfoServer(); err != nil {
		return fmt.Errorf("failed to register info server: %w", err)
	}

	if err := svc.registerAdminServer(); err != nil {
		return fmt.Errorf("failed to register admin server: %w", err)
	}

	if err := svc.registerCallbackServer(); err != nil {
		return fmt.Errorf("failed to register callback server: %w", err)
	}

	if err := svc.registerDownloadServer(); err != nil {
		return fmt.Errorf("failed to register download server: %w", err)
	}

	return nil
}

// nolint: unparam
func (svc *Service) registerInfoServer() error {
	requestIDSetter := restserver.NewRequestIDSetter()
	tenantIDSetter := restserver.NewTenantIDSetter()

	server := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:            string(relayInfoSvcName),
			IP:              svc.conf.InfoServer.BindIP,
			Port:            svc.conf.InfoServer.Port,
			RequestIDSetter: requestIDSetter,
			TenantIDSetter:  tenantIDSetter,
		},
		restserver.WithPing(),
		restserver.WithMetrics(),
		withHealthz(svc.Cap),
	)

	svc.servers = append(svc.servers, server)

	return nil
}

// nolint: unparam
func (svc *Service) registerAdminServer() error {
	authIdentity := svc.authIdentityMap[svc.conf.AdminServer.AuthIdentity]
	if authIdentity == nil {
		return fmt.Errorf("no support this auth identity, auth-identity(%s), use-one-of(%v)",
			svc.conf.AdminServer.AuthIdentity, conv.MapKeyToSlice(svc.authIdentityMap))
	}

	requestIDSetter := restserver.NewRequestIDSetter()
	tenantIDSetter := restserver.NewTenantIDSetter()

	server := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:            string(relayAdminSvcName),
			IP:              svc.conf.AdminServer.BindIP,
			Port:            svc.conf.AdminServer.Port,
			RequestIDSetter: requestIDSetter,
			TenantIDSetter:  tenantIDSetter,
		},
		restserver.WithPing(),
	)

	svc.servers = append(svc.servers, server)

	return nil
}

// nolint: unparam
func (svc *Service) registerCallbackServer() error {
	authIdentity := svc.authIdentityMap[svc.conf.CallbackServer.AuthIdentity]
	if authIdentity == nil {
		return fmt.Errorf("no support this auth identity, auth-identity(%s), use-one-of(%v)",
			svc.conf.CallbackServer.AuthIdentity, conv.MapKeyToSlice(svc.authIdentityMap))
	}

	requestIDSetter := restserver.NewRequestIDSetter()
	tenantIDSetter := restserver.NewTenantIDSetter()

	server := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:            string(relayCallbackSvcName),
			IP:              svc.conf.CallbackServer.BindIP,
			Port:            svc.conf.CallbackServer.Port,
			RequestIDSetter: requestIDSetter,
			TenantIDSetter:  tenantIDSetter,
		},
		restserver.WithPing(),
		withCallbackServer(svc.Cap),
	)

	svc.servers = append(svc.servers, server)

	return nil
}

// nolint: unparam
func (svc *Service) registerDownloadServer() error {
	authIdentity := svc.authIdentityMap[svc.conf.DownloadServer.AuthIdentity]
	if authIdentity == nil {
		return fmt.Errorf("no support this auth identity, auth-identity(%s), use-one-of(%v)",
			svc.conf.DownloadServer.AuthIdentity, conv.MapKeyToSlice(svc.authIdentityMap))
	}

	requestIDSetter := restserver.NewRequestIDSetter()
	tenantIDSetter := restserver.NewTenantIDSetter()

	server := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:            string(relayDownloadSvcName),
			IP:              svc.conf.DownloadServer.BindIP,
			Port:            svc.conf.DownloadServer.Port,
			RequestIDSetter: requestIDSetter,
			TenantIDSetter:  tenantIDSetter,
		},
		restserver.WithPing(),
		withDownload(svc.Cap),
	)

	svc.servers = append(svc.servers, server)

	return nil
}

func withHealthz(capability *options.Capability) restserver.RouterOptionFunc {
	return func(rg *gin.RouterGroup) restserver.IMiddlewareChain {
		return healthz.Load(rg, capability)
	}
}

// withCallbackServer load callback api.
func withCallbackServer(capability *options.Capability) restserver.RouterOptionFunc {
	return func(rg *gin.RouterGroup) restserver.IMiddlewareChain {
		return callback.Load(rg, capability)
	}
}

// withDownload load download.
func withDownload(capability *options.Capability) restserver.RouterOptionFunc {
	return func(rg *gin.RouterGroup) restserver.IMiddlewareChain {
		return download.Load(rg, capability)
	}
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
			logger.G.Sys().With("name", server.Name(), "ip", server.IP(), "port", server.Port()).Info("started server")

			if err := server.Start(); err != nil {
				return err
			}

			return nil
		}
		gp.Go(fn)
	}

	// wait until all servers stopped or application error.
	if err := gp.Wait(); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to start servers")

		return err
	}

	return nil
}
