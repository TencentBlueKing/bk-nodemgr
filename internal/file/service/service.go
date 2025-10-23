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
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/TencentBlueKing/bk-nodemgr/internal/file/manager"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/router/download"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/router/healthz"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/router/publish"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/router/transfer"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/router/upload"
	storageRelease "github.com/TencentBlueKing/bk-nodemgr/internal/file/storage/release"
	storageTopo "github.com/TencentBlueKing/bk-nodemgr/internal/file/storage/topo"
	storageUpload "github.com/TencentBlueKing/bk-nodemgr/internal/file/storage/upload"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover/etcddiscover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/bkrepo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
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
	ctx contextx.IContext

	// cancelFunc is used to cancel the service and all associated operations.
	cancelFunc context.CancelFunc

	// router is the entry point of the service, routing requests to different capabilities.
	servers []*restserver.Server

	// Note: Cap is initialized in the Start() and could not be used in other package.
	// Cap is the capability of the service.
	Cap *options.Capability

	// instance is the discover instance of the service.
	instance discover.Instance

	// authIdentityMap defines the mapping between auth identity and auth identity handler.
	authIdentityMap map[config.AuthIdentity]restserver.IAuthIdentity
}

// NewService creates a new file service.
// nolint:funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func NewService(conf *config.FileService) (*Service, error) {
	svc := &Service{
		conf:     conf,
		Cap:      &options.Capability{},
		instance: discover.NewInstance(string(discover.ServiceNameFile), nil),
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

// nolint: unparam
func (svc *Service) initialStaticsConfigs() error {
	svc.authIdentityMap = map[config.AuthIdentity]restserver.IAuthIdentity{
		config.AuthIdentityNone:       restserver.NewNodeAuthIdentity(),
		config.AuthIdentityRestServer: restserver.NewRestServerAuthIdentity(svc.conf.RestServer.JwtSecret),
	}

	return nil
}

// nolint: funlen
func (svc *Service) initialCapability() error {
	var err error

	// discover provider.
	svc.Cap.DiscoverProvider = etcddiscover.NewProviderEtcd(&svc.conf.Etcd)

	// initial gse handler.
	svc.Cap.GSEHandler, err = svc.newGSEHandler()
	if err != nil {
		return fmt.Errorf("failed to create gse handler: %w", err)
	}

	// initial bkrepo.
	svc.Cap.BKRepo, err = svc.newBKRepoHandler()
	if err != nil {
		return fmt.Errorf("failed to create bkrepo handler: %w", err)
	}

	// initial mongo client.
	svc.Cap.MongoClient, err = svc.newMongoClient()
	if err != nil {
		return fmt.Errorf("failed to create mongo client: %w", err)
	}

	// initial serveral storages.
	if err = svc.initialStorages(); err != nil {
		return fmt.Errorf("failed to initial storages: %w", err)
	}

	// initial manager.
	if err = svc.initialManager(); err != nil {
		return fmt.Errorf("failed to initial manager: %w", err)
	}

	return nil
}

func (svc *Service) newGSEHandler() (gse.IHandler, error) {
	apiGWUserConfig := newAPIGWUserConfig(&svc.conf.GSE.APIGatewayClient)
	apiGwClientCapability, err := newAPIGwClientCapability(&svc.conf.GSE.APIGatewayClient)
	if err != nil {
		return nil, fmt.Errorf("failed to new apigw client for gse: %w", err)
	}

	apiGwClientCapability.Name = "gse"
	gseHandler, err := gse.New(
		apiGwClientCapability,
		&gse.Config{
			APIGWUserConfig: apiGWUserConfig,
		},
	)
	if err != nil {
		return nil, err
	}

	return gseHandler, nil
}

func (svc *Service) newBKRepoHandler() (bkrepo.IHandler, error) {
	httpClient, err := restclient.NewHTTPClient(&ssl.TLSConfig{InsecureSkipVerify: true})
	if err != nil {
		return nil, fmt.Errorf("failed to create http client for bkrepo handler: %w", err)
	}

	clientCap := &restclient.Capability{
		Name:                 "bkrepo",
		HTTPClient:           httpClient,
		Discover:             restdiscovery.NewDiscovery("bkrepo", []string{svc.conf.Repo.Endpoint}),
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
		MetricOpts:           restclient.MetricOption{},
	}

	return bkrepo.New(clientCap, &bkrepo.Config{
		ProjectID: svc.conf.Repo.ProjectID,
		RepoName:  svc.conf.Repo.RepoName,
		Username:  svc.conf.Repo.AccessKey,
		Password:  svc.conf.Repo.SecretKey,
	})
}

func (svc *Service) newMongoClient() (*mongo.Client, error) {
	mongoClient, err := mongo.Connect(
		context.Background(),
		&mongoOptions.ClientOptions{
			Hosts: svc.conf.MongoDB.Hosts,
			Auth: &mongoOptions.Credential{
				Username:      svc.conf.MongoDB.Username,
				Password:      svc.conf.MongoDB.Password,
				AuthSource:    svc.conf.MongoDB.AuthSource,
				AuthMechanism: svc.conf.MongoDB.AuthMechanism,
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongo client: %w", err)
	}

	return mongoClient, nil
}

// nolint: funlen
func (svc *Service) initialStorages() error {
	var err error

	svc.Cap.StorageUpload, err = storageUpload.NewStorage(
		svc.Cap.MongoClient,
		svc.conf.MongoDB.Database)
	if err != nil {
		return fmt.Errorf("failed to create upload storage: %w", err)
	}

	svc.Cap.StorageRelease, err = storageRelease.NewStorage(
		svc.Cap.MongoClient,
		svc.conf.MongoDB.Database)
	if err != nil {
		return fmt.Errorf("failed to create release storage: %w", err)
	}

	svc.Cap.StorageTopo, err = storageTopo.NewStorage(
		svc.Cap.MongoClient,
		svc.conf.MongoDB.Database)
	if err != nil {
		return fmt.Errorf("failed to create topo storage: %w", err)
	}

	return nil
}

// nolint: funlen
func (svc *Service) initialManager() error {
	// init upstream origin file groups from bkrepo.
	upstreamOriginAgentFG, err := svc.Cap.BKRepo.EnsureFileGroup(contextx.New(context.Background()), "origin/agent")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream origin agent file group: %w", err)
	}
	upstreamOriginServerFG, err := svc.Cap.BKRepo.EnsureFileGroup(contextx.New(context.Background()), "origin/server")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream origin server file group: %w", err)
	}
	upstreamOriginCertFG, err := svc.Cap.BKRepo.EnsureFileGroup(contextx.New(context.Background()), "origin/cert")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream origin cert file group: %w", err)
	}
	upstreamOriginBinToolFG, err := svc.Cap.BKRepo.EnsureFileGroup(contextx.New(context.Background()), "origin/bintool")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream origin bin tool file group: %w", err)
	}
	upstreamOriginPluginV2, err := svc.Cap.BKRepo.EnsureFileGroup(contextx.New(context.Background()), "origin/v2/plugin")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream origin plugin file group: %w", err)
	}
	upstreamOriginExternalPluginV2, err := svc.Cap.BKRepo.EnsureFileGroup(contextx.New(context.Background()), "origin/v2/external_plugin")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream origin external plugin file group: %w", err)
	}

	// init upstream release file groups from bkrepo.
	upstreamReleaseAgentFG, err := svc.Cap.BKRepo.EnsureFileGroup(contextx.New(context.Background()), "release/agent")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream release agent file group: %w", err)
	}
	upstreamReleaseProxyFg, err := svc.Cap.BKRepo.EnsureFileGroup(contextx.New(context.Background()), "release/proxy")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream release proxy file group: %w", err)
	}
	upstreamRealseCertFG, err := svc.Cap.BKRepo.EnsureFileGroup(contextx.New(context.Background()), "release/cert")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream release cert file group: %w", err)
	}
	upstreamReleaseBintoolFG, err := svc.Cap.BKRepo.EnsureFileGroup(contextx.New(context.Background()), "release/bintool")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream release bin tool file group: %w", err)
	}
	upstreamOriginPluginBinToolV2FG, err := svc.Cap.BKRepo.EnsureFileGroup(contextx.New(context.Background()), "origin/v2/plugin_bintool")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream origin plugin bin tool file group: %w", err)
	}
	upstreamReleasePluginBinToolV2FG, err := svc.Cap.BKRepo.EnsureFileGroup(contextx.New(context.Background()), "release/v2/plugin_bintool")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream release bin tool file group: %w", err)
	}
	upstreamReleasePlugin, err := svc.Cap.BKRepo.EnsureFileGroup(contextx.New(context.Background()), "release/plugin")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream release plugin file group: %w", err)
	}

	// init local temp file group.
	tempFG, err := local.NewLocalDir(filepath.Join(svc.conf.WorkspaceFileGroup.FullPath, "temp"))
	if err != nil {
		return fmt.Errorf("failed to init temp file group: %w", err)
	}
	installerFG, err := local.NewLocalDir(filepath.Join(svc.conf.WorkspaceFileGroup.FullPath, "installer"))
	if err != nil {
		return fmt.Errorf("failed to init installer file group: %w", err)
	}
	cacheFG, err := local.NewLocalDir(filepath.Join(svc.conf.WorkspaceFileGroup.FullPath, "cache"))
	if err != nil {
		return fmt.Errorf("failed to init cache file group: %w", err)
	}

	svc.Cap.Manager = manager.New(
		manager.WithUpstreamOriginAgentFileGroup(upstreamOriginAgentFG),
		manager.WithUpstreamOriginServerFileGroup(upstreamOriginServerFG),
		manager.WithUpstreamOriginCertFileGroup(upstreamOriginCertFG),
		manager.WithUpstreamOriginBinToolFileGroup(upstreamOriginBinToolFG),
		manager.WithUpstreamOriginPluginBinToolV2FileGroup(upstreamOriginPluginBinToolV2FG),
		manager.WithUpstreamReleaseAgentFileGroup(upstreamReleaseAgentFG),
		manager.WithUpstreamReleaseProxyFileGroup(upstreamReleaseProxyFg),
		manager.WithUpstreamReleaseCertFileGroup(upstreamRealseCertFG),
		manager.WithUpstreamReleaseBinToolFileGroup(upstreamReleaseBintoolFG),
		manager.WithUpstreamReleasePluginBinToolV2FileGroup(upstreamReleasePluginBinToolV2FG),
		manager.WithTempFileGroup(tempFG),
		manager.WithInstallerFileGroup(installerFG),
		manager.WithCacheFileGroup(cacheFG),
		manager.WithStorageUpload(svc.Cap.StorageUpload),
		manager.WithStorageRelease(svc.Cap.StorageRelease),
		manager.WithStorageTopo(svc.Cap.StorageTopo),
		manager.WithAdvertiseIPV4(svc.conf.BasicServer.AdvertiseIPV4),
		manager.WithAdvertiseIPV6(svc.conf.BasicServer.AdvertiseIPV6),
		manager.WithMount(svc.conf.MountHostDir, svc.conf.WorkspaceFileGroup.FullPath),
		manager.WithGSEHandler(svc.Cap.GSEHandler),
		manager.WithUpstreamOriginPluginV2FileGroup(upstreamOriginPluginV2),
		manager.WithUpstreamOriginExternalPluginV2FileGroup(upstreamOriginExternalPluginV2),
		manager.WithUpstreamReleasePluginFileGroup(upstreamReleasePlugin),
	)

	return nil
}

func (svc *Service) registerRestServer() error {
	if err := svc.registerInfoServer(); err != nil {
		return fmt.Errorf("failed to register info server: %w", err)
	}

	if err := svc.registerAdminServer(); err != nil {
		return fmt.Errorf("failed to register admin server: %w", err)
	}

	if err := svc.registerBasicServer(); err != nil {
		return fmt.Errorf("failed to register basic server: %w", err)
	}

	if err := svc.registerDownloadServer(); err != nil {
		return fmt.Errorf("failed to register download server: %w", err)
	}

	return nil
}

// nolint: unparam
func (svc *Service) registerInfoServer() error {
	server := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:            string(discover.EndpointNameFileInfo),
			IP:              svc.conf.InfoServer.BindIP,
			Port:            svc.conf.InfoServer.Port,
			RequestIDSetter: restserver.NewRequestIDSetter(),
			TenantIDSetter:  restserver.NewTenantIDSetter(),
		},
		restserver.WithPing(),
		withHealthz(svc.Cap),
		withMetrics(svc.Cap),
	)

	svc.servers = append(svc.servers, server)
	svc.instance.Update(discover.EndpointNameFileInfo, discover.Endpoint{
		IPV4: svc.conf.InfoServer.AdvertiseIPV4,
		IPV6: svc.conf.InfoServer.AdvertiseIPV6,
		Port: svc.conf.InfoServer.Port,
	})

	return nil
}

func (svc *Service) registerAdminServer() error {
	authIdentity := svc.authIdentityMap[svc.conf.AdminServer.AuthIdentity]
	if authIdentity == nil {
		return fmt.Errorf("no support this auth identity, auth-identity(%s), use-one-of(%v)",
			svc.conf.AdminServer.AuthIdentity, conv.MapKeyToSlice(svc.authIdentityMap))
	}

	server := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:            string(discover.EndpointNameFileAdmin),
			IP:              svc.conf.AdminServer.BindIP,
			Port:            svc.conf.AdminServer.Port,
			RequestIDSetter: restserver.NewRequestIDSetter(),
			TenantIDSetter:  restserver.NewTenantIDSetter(),
		},
		restserver.WithPing(),
	)

	svc.servers = append(svc.servers, server)
	svc.instance.Update(discover.EndpointNameFileAdmin, discover.Endpoint{
		IPV4: svc.conf.AdminServer.AdvertiseIPV4,
		IPV6: svc.conf.AdminServer.AdvertiseIPV6,
		Port: svc.conf.AdminServer.Port,
	})

	return nil
}

func (svc *Service) registerBasicServer() error {
	authIdentity := svc.authIdentityMap[svc.conf.BasicServer.AuthIdentity]
	if authIdentity == nil {
		return fmt.Errorf("no support this auth identity, auth-identity(%s), use-one-of(%v)",
			svc.conf.BasicServer.AuthIdentity, conv.MapKeyToSlice(svc.authIdentityMap))
	}

	server := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:            string(discover.EndpointNameFileBasic),
			IP:              svc.conf.BasicServer.BindIP,
			Port:            svc.conf.BasicServer.Port,
			RequestIDSetter: restserver.NewRequestIDSetter(),
			TenantIDSetter:  restserver.NewTenantIDSetter(),
		},
		restserver.WithPing(),
		withUpload(svc.Cap, authIdentity),
		withPublish(svc.Cap, authIdentity),
		withTransfer(svc.Cap, authIdentity),
		withDownload(svc.Cap, authIdentity),
	)

	svc.servers = append(svc.servers, server)
	svc.instance.Update(discover.EndpointNameFileBasic, discover.Endpoint{
		IPV4: svc.conf.BasicServer.AdvertiseIPV4,
		IPV6: svc.conf.BasicServer.AdvertiseIPV6,
		Port: svc.conf.BasicServer.Port,
	})

	return nil
}

func (svc *Service) registerDownloadServer() error {
	authIdentity := svc.authIdentityMap[svc.conf.DownloadServer.AuthIdentity]
	if authIdentity == nil {
		return fmt.Errorf("no support this auth identity, auth-identity(%s), use-one-of(%v)",
			svc.conf.DownloadServer.AuthIdentity, conv.MapKeyToSlice(svc.authIdentityMap))
	}

	server := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:            string(discover.EndpointNameFileDownload),
			IP:              svc.conf.DownloadServer.BindIP,
			Port:            svc.conf.DownloadServer.Port,
			RequestIDSetter: restserver.NewRequestIDSetter(),
			TenantIDSetter:  restserver.NewTenantIDSetter(),
		},
		restserver.WithPing(),
		withDownload(svc.Cap, authIdentity),
	)

	svc.servers = append(svc.servers, server)
	svc.instance.Update(discover.EndpointNameFileDownload, discover.Endpoint{
		IPV4: svc.conf.DownloadServer.AdvertiseIPV4,
		IPV6: svc.conf.DownloadServer.AdvertiseIPV6,
		Port: svc.conf.DownloadServer.Port,
	})

	return nil
}

// newAPIGwClientCapability creates a new api-gateway client capability.
func newAPIGwClientCapability(conf *config.APIGatewayClient) (*restclient.Capability, error) {
	httpClient, err := restclient.NewHTTPClient(&ssl.TLSConfig{
		InsecureSkipVerify: conf.TLS.InsecureSkipVerify,
		CertFile:           conf.TLS.CertFile,
		KeyFile:            conf.TLS.KeyFile,
		CAFile:             conf.TLS.CAFile,
		Password:           conf.TLS.Password,
	})
	if err != nil {
		return nil, err
	}

	clientCap := &restclient.Capability{
		HTTPClient:           httpClient,
		Discover:             restdiscovery.NewDiscovery(DiscoveryNameApigw, conf.Endpoints),
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
		MetricOpts:           restclient.MetricOption{},
	}

	return clientCap, nil
}

// newAPIGWUserConfig creates a new api-gateway client config.
func newAPIGWUserConfig(conf *config.APIGatewayClient) apigwclient.UserConfig {
	return apigwclient.UserConfig{
		AppConfig:   apigwclient.NewAppConfig(conf.Endpoints, conf.AppCode, conf.AppSecret),
		AuthMode:    apigwclient.AuthMode(conf.AuthMode),
		BKUsername:  conf.User,
		AccessToken: conf.AccessToken,
	}
}

// withHealthz load healthz.
func withHealthz(capability *options.Capability) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		healthz.Load(rg, capability)
	}
}

// withMetrics load metrics.
func withMetrics(_ *options.Capability) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		rg.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}
}

// withDownload load download.
func withDownload(capability *options.Capability, authIdentity restserver.IAuthIdentity) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		download.Load(rg, capability, authIdentity)
	}
}

// withUpload load upload.
func withUpload(capability *options.Capability, authIdentity restserver.IAuthIdentity) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		upload.Load(rg, capability, authIdentity)
	}
}

// withPublish load publish.
func withPublish(capability *options.Capability, authIdentity restserver.IAuthIdentity) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		publish.Load(rg, capability, authIdentity)
	}
}

// withTransfer load transfer.
func withTransfer(capability *options.Capability, authIdentity restserver.IAuthIdentity) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		transfer.Load(rg, capability, authIdentity)
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
			logger.G.Sys().With("name", server.Name(), "ip", server.IP(), "port", server.Port()).Info("started server")

			if err := server.Start(); err != nil {
				return err
			}

			return nil
		}
		gp.Go(fn)
	}

	// after all servers brings up, register the instance into discover provider.
	if err := svc.Cap.DiscoverProvider.Register(discover.ServiceNameFile, svc.instance); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to register instance")

		return err
	}

	// wait until all servers stopped or application error.
	if err := gp.Wait(); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to start servers")

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

	return nil
}
