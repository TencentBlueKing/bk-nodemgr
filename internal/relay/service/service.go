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
	"errors"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/file"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/handler"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/router/admin"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/router/callback"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/router/download"
	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/router/healthz"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/version"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
}

// NewService creates a new relay service.
// nolint: funlen
func NewService(conf *config.RelayService) (*Service, error) {
	svc := &Service{
		conf: conf,
		Cap:  &options.Capability{},
	}

	svc.ctx, svc.cancelFunc = contextx.WithCancel(contextx.New(contextx.Background()))

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

func (svc *Service) newAuthIdentity(conf config.HTTPServer) (restserver.IAuthIdentity, error) {
	switch conf.AuthIdentity {
	case config.AuthIdentityNone:
		return restserver.NewNoneAuthIdentity(), nil
	case config.AuthIdentityRestServer:
		return restserver.NewRestServerAuthIdentity(conf.JWTServerConfig.SymmetricKey), nil
	default:
		return nil, fmt.Errorf("no support this auth identity, auth-identity(%s)", conf.AuthIdentity)
	}
}

// nolint: unparam
func (svc *Service) registerInfoServer() error {
	server, err := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:             relayInfoSvcName,
			IP:               svc.conf.InfoServer.BindIP,
			Port:             svc.conf.InfoServer.Port,
			RequestIDSetter:  restserver.NewRequestIDSetter(),
			TraceServiceName: svc.conf.InfoServer.TraceServiceName,
			TraceSampleRate:  svc.conf.InfoServer.TraceSampleRate,
		},
		restserver.WithPing(),
		withHealthz(svc.Cap),
		withMetrics(),
	)
	if err != nil {
		return fmt.Errorf("failed to register info server: %w", err)
	}

	svc.servers = append(svc.servers, server)

	return nil
}

// nolint: unparam
func (svc *Service) registerAdminServer() error {
	if svc.conf.AdminServer.AuthIdentity != config.AuthIdentityNone &&
		svc.conf.AdminServer.AuthIdentity != config.AuthIdentityRestServer {

		return fmt.Errorf("no support this auth identity, auth-identity(%s), support auth-identity(%v, %v)",
			svc.conf.AdminServer.AuthIdentity, config.AuthIdentityNone, config.AuthIdentityRestServer)
	}

	authIdentity, err := svc.newAuthIdentity(svc.conf.AdminServer)
	if err != nil {
		return fmt.Errorf("failed to new auth identity: %w", err)
	}

	requestIDSetter := restserver.NewRequestIDSetter()

	server, err := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:             relayAdminSvcName,
			IP:               svc.conf.AdminServer.BindIP,
			Port:             svc.conf.AdminServer.Port,
			RequestIDSetter:  requestIDSetter,
			TraceServiceName: svc.conf.AdminServer.TraceServiceName,
			TraceSampleRate:  svc.conf.AdminServer.TraceSampleRate,
		},
		restserver.WithPing(),
		withAdmin(svc.Cap,
			restserver.MiddlewareAuth(authIdentity),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to register admin server: %w", err)
	}

	svc.servers = append(svc.servers, server)

	return nil
}

// nolint: unparam
func (svc *Service) registerCallbackServer() error {
	server, err := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:             relayCallbackSvcName,
			IP:               svc.conf.CallbackServer.BindIP,
			Port:             svc.conf.CallbackServer.Port,
			RequestIDSetter:  restserver.NewRequestIDSetter(),
			TraceServiceName: svc.conf.CallbackServer.TraceServiceName,
			TraceSampleRate:  svc.conf.CallbackServer.TraceSampleRate,
		},
		restserver.WithPing(),
		withCallbackServer(svc.Cap),
	)
	if err != nil {
		return fmt.Errorf("failed to register callback server: %w", err)
	}

	svc.servers = append(svc.servers, server)

	return nil
}

// nolint: unparam
func (svc *Service) registerDownloadServer() error {
	if svc.conf.DownloadServer.AuthIdentity != config.AuthIdentityNone {
		return fmt.Errorf("no support this auth identity, auth-identity(%s), support auth-identity(%v)",
			svc.conf.DownloadServer.AuthIdentity, config.AuthIdentityNone)
	}

	server, err := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:             relayDownloadSvcName,
			IP:               svc.conf.DownloadServer.BindIP,
			Port:             svc.conf.DownloadServer.Port,
			RequestIDSetter:  restserver.NewRequestIDSetter(),
			TraceServiceName: svc.conf.DownloadServer.TraceServiceName,
			TraceSampleRate:  svc.conf.DownloadServer.TraceSampleRate,
		},
		restserver.WithPing(),
		withDownload(svc.Cap),
	)
	if err != nil {
		return fmt.Errorf("failed to register download server: %w", err)
	}

	svc.servers = append(svc.servers, server)

	return nil
}

func withHealthz(capability *options.Capability) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		healthz.Load(rg, capability)
	}
}

// withAdmin load admin.
func withAdmin(capability *options.Capability, middleware ...gin.HandlerFunc) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		admin.Load(rg, capability, middleware...)
	}
}

// withMetrics load metrics.
func withMetrics() restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		rg.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}
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

// GracefulShutdown gracefully shuts down the application service.
func (svc *Service) GracefulShutdown() error {
	logger.G.Sys().Info("try to gracefully shutdown relay service")

	if svc.ctx == nil || svc.cancelFunc == nil {
		return errors.New("service is not running")
	}

	defer svc.cancelFunc()

	err := svc.Cap.GracefulShutdown()
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to gracefully shutdown capability")

		return err
	}

	logger.G.Sys().Info("relay service gracefully shutdown")

	return nil
}
