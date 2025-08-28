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
	"io"
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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/blog"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/etcddiscover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
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
	ctx context.Context

	// cancelFunc is used to cancel the service and all associated operations.
	cancelFunc context.CancelFunc

	// router is the entry point of the service, routing requests to different capabilities.
	servers []*restserver.Server

	// Note: Cap is initialized in the Start() and could not be used in other package.
	// Cap is the capability of the service.
	Cap *options.Capability

	// instance is the discover instance of the service.
	instance discover.Instance
}

// NewService creates a new file service.
// nolint:funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func NewService(conf *config.FileService) (*Service, error) {
	svc := &Service{
		conf: conf,
		Cap: &options.Capability{
			Logger: blog.GlobalLogger{},
		},
		instance: discover.NewInstance("file", nil),
	}

	svc.ctx, svc.cancelFunc = context.WithCancel(context.Background())

	svc.Cap.DiscoverProvider = etcddiscover.NewProviderEtcd(&conf.Etcd,
		etcddiscover.WithLogger(svc.Cap.Logger),
		etcddiscover.WithWatch(discover.ServiceNameBackend, discover.ServiceNameFile),
	)

	var err error

	// init mongoclient.
	mongoClient, err := initMongoDB(&conf.MongoDB)
	if err != nil {
		return nil, err
	}

	svc.Cap.StorageUpload, err = storageUpload.NewStorage(mongoClient, conf.MongoDB.Database, svc.Cap.Logger)
	if err != nil {
		return nil, err
	}

	svc.Cap.StorageRelease, err = storageRelease.NewStorage(mongoClient, conf.MongoDB.Database, svc.Cap.Logger)
	if err != nil {
		return nil, err
	}

	svc.Cap.StorageTopo, err = storageTopo.NewStorage(mongoClient, conf.MongoDB.Database, svc.Cap.Logger)
	if err != nil {
		return nil, err
	}

	// init gse handler.
	svc.Cap.GSEHandler, err = newGSEHandler(conf.GSE)
	if err != nil {
		return nil, err
	}

	// init bkrepo.
	svc.Cap.BKRepo, err = initBKRepo(conf, svc.Cap.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init bkrepo: %w", err)
	}

	// init manager.
	svc.Cap.Manager, err = initManager(conf,
		svc.Cap.BKRepo,
		svc.Cap.StorageUpload,
		svc.Cap.StorageRelease,
		svc.Cap.StorageTopo,
		svc.Cap.GSEHandler,
		svc.Cap.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init manager: %w", err)
	}

	requestIDSetter := restserver.NewRequestIDSetter()
	tenantIDSetter := restserver.NewTenantIDSetter()

	httpServer := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:            string(discover.EndpointNameFileBasic),
			IP:              conf.HTTPServer.BindIP,
			Port:            conf.HTTPServer.Port,
			LogWriter:       loggerWriterAdaptor{},
			RequestIDSetter: requestIDSetter,
			TenantIDSetter:  tenantIDSetter,
		},
		restserver.WithPing(),
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

	adminServer := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:            string(discover.EndpointNameFileAdmin),
			IP:              conf.AdminServer.BindIP,
			RequestIDSetter: requestIDSetter,
			TenantIDSetter:  tenantIDSetter,
			Port:            conf.AdminServer.Port,
			LogWriter:       loggerWriterAdaptor{},
		},
		restserver.WithPing(),
		withHealthz(svc.Cap),
		withMetrics(svc.Cap),
		withUpload(svc.Cap),
		withPublish(svc.Cap),
		withTransfer(svc.Cap),
	)
	svc.servers = append(svc.servers, adminServer)
	svc.instance.Update(discover.EndpointNameFileAdmin, discover.Endpoint{
		IPV4: conf.AdminServer.AdvertiseIPV4,
		IPV6: conf.AdminServer.AdvertiseIPV6,
		Port: conf.AdminServer.Port,
	})

	return svc, nil
}

// newGSEHandler.
func newGSEHandler(conf config.GSE) (gse.IHandler, error) {
	apiGwClientConfig := newAPIGwClientConfig(&conf.APIGatewayClient)
	apiGwClientCapability, err := newAPIGwClientCapability(&conf.APIGatewayClient)
	if err != nil {
		return nil, err
	}

	apiGwClientCapability.Name = "gse"
	gseHandler, err := gse.New(apiGwClientCapability, &gse.Config{
		APIGWUserConfig: apiGwClientConfig,
	})
	if err != nil {
		return nil, err
	}

	return gseHandler, nil
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
		Logger:               blog.GlobalLogger{},
	}

	return clientCap, nil
}

// newAPIGwClientConfig creates a new api-gateway client config.
func newAPIGwClientConfig(conf *config.APIGatewayClient) apigwclient.UserConfig {
	return apigwclient.UserConfig{
		AppConfig: apigwclient.NewAppConfig(
			conf.Endpoints,
			conf.AppCode,
			conf.AppSecret),
		AuthMode:    apigwclient.AuthMode(conf.AuthMode),
		BKUsername:  conf.User,
		AccessToken: conf.AccessToken,
	}
}

var _ restserver.ILogWriter = &loggerWriterAdaptor{}

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
func withDownload(capability *options.Capability) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		download.Load(rg, capability)
	}
}

// withUpload load upload.
func withUpload(capability *options.Capability) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		upload.Load(rg, capability)
	}
}

// withPublish load publish.
func withPublish(capability *options.Capability) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		publish.Load(rg, capability)
	}
}

// withTransfer load transfer.
func withTransfer(capability *options.Capability) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		transfer.Load(rg, capability)
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

// nolint: funlen
func initManager(conf *config.FileService,
	repo bkrepo.IHandler,
	storageUpload storageUpload.IStorage,
	storageRelease storageRelease.IStorage,
	storageTopo storageTopo.IStorage,
	gseHandler gse.IHandler,
	logger logger.ILogger) (manager.IManager, error) {

	// init upstream origin file groups from bkrepo.
	upstreamOriginAgentFG, err := repo.EnsureFileGroup(context.Background(), "origin/agent")
	if err != nil {
		return nil, fmt.Errorf("failed to ensure upstream origin agent file group: %w", err)
	}
	upstreamOriginServerFG, err := repo.EnsureFileGroup(context.Background(), "origin/server")
	if err != nil {
		return nil, fmt.Errorf("failed to ensure upstream origin server file group: %w", err)
	}
	upstreamOriginCertFG, err := repo.EnsureFileGroup(context.Background(), "origin/cert")
	if err != nil {
		return nil, fmt.Errorf("failed to ensure upstream origin cert file group: %w", err)
	}
	upstreamOriginBinToolFG, err := repo.EnsureFileGroup(context.Background(), "origin/bintool")
	if err != nil {
		return nil, fmt.Errorf("failed to ensure upstream origin bin tool file group: %w", err)
	}
	upstreamOriginOfficialPlugin, err := repo.EnsureFileGroup(context.Background(), "origin/official_plugin")
	if err != nil {
		return nil, fmt.Errorf("failed to ensure upstream origin official plugin file group: %w", err)
	}
	upstreamOriginExternalPlugin, err := repo.EnsureFileGroup(context.Background(), "origin/external_plugin")
	if err != nil {
		return nil, fmt.Errorf("failed to ensure upstream origin external plugin file group: %w", err)
	}

	// init upstream release file groups from bkrepo.
	upstreamReleaseAgentFG, err := repo.EnsureFileGroup(context.Background(), "release/agent")
	if err != nil {
		return nil, fmt.Errorf("failed to ensure upstream release agent file group: %w", err)
	}
	upstreamReleaseProxyFg, err := repo.EnsureFileGroup(context.Background(), "release/proxy")
	if err != nil {
		return nil, fmt.Errorf("failed to ensure upstream release proxy file group: %w", err)
	}
	upstreamRealseCertFG, err := repo.EnsureFileGroup(context.Background(), "release/cert")
	if err != nil {
		return nil, fmt.Errorf("failed to ensure upstream release cert file group: %w", err)
	}
	upstreamReleaseBintoolFG, err := repo.EnsureFileGroup(context.Background(), "release/bintool")
	if err != nil {
		return nil, fmt.Errorf("failed to ensure upstream release bin tool file group: %w", err)
	}
	upstreamReleaseOfficialPlugin, err := repo.EnsureFileGroup(context.Background(), "release/official_plugin")
	if err != nil {
		return nil, fmt.Errorf("failed to ensure upstream release official plugin file group: %w", err)
	}
	upstreamReleaseExternalPlugin, err := repo.EnsureFileGroup(context.Background(), "release/external_plugin")
	if err != nil {
		return nil, fmt.Errorf("failed to ensure upstream release external plugin file group: %w", err)
	}

	// init local temp file group.
	tempFG, err := local.NewLocalDir(filepath.Join(conf.WorkspaceFileGroup.FullPath, "temp"), logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init temp file group: %w", err)
	}
	installerFG, err := local.NewLocalDir(filepath.Join(conf.WorkspaceFileGroup.FullPath, "installer"), logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init installer file group: %w", err)
	}
	cacheFG, err := local.NewLocalDir(filepath.Join(conf.WorkspaceFileGroup.FullPath, "cache"), logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init cache file group: %w", err)
	}

	return manager.New(
		manager.WithLogger(logger),
		manager.WithUpstreamOriginAgentFileGroup(upstreamOriginAgentFG),
		manager.WithUpstreamOriginServerFileGroup(upstreamOriginServerFG),
		manager.WithUpstreamOriginCertFileGroup(upstreamOriginCertFG),
		manager.WithUpstreamOriginBinToolFileGroup(upstreamOriginBinToolFG),
		manager.WithUpstreamReleaseAgentFileGroup(upstreamReleaseAgentFG),
		manager.WithUpstreamReleaseProxyFileGroup(upstreamReleaseProxyFg),
		manager.WithUpstreamReleaseCertFileGroup(upstreamRealseCertFG),
		manager.WithUpstreamReleaseBinToolFileGroup(upstreamReleaseBintoolFG),
		manager.WithTempFileGroup(tempFG),
		manager.WithInstallerFileGroup(installerFG),
		manager.WithCacheFileGroup(cacheFG),
		manager.WithStorageUpload(storageUpload),
		manager.WithStorageRelease(storageRelease),
		manager.WithStorageTopo(storageTopo),
		manager.WithAdvertiseIPV4(conf.HTTPServer.AdvertiseIPV4),
		manager.WithAdvertiseIPV6(conf.HTTPServer.AdvertiseIPV6),
		manager.WithMount(conf.MountHostDir, conf.WorkspaceFileGroup.FullPath),
		manager.WithGSEHandler(gseHandler),
		manager.WithUpstreamOriginOfficialPluginFileGroup(upstreamOriginOfficialPlugin),
		manager.WithUpstreamReleaseOfficialPluginFileGroup(upstreamReleaseOfficialPlugin),
		manager.WithUpstreamOriginExternalPluginFileGroup(upstreamOriginExternalPlugin),
		manager.WithUpstreamReleaseExternalPluginFileGroup(upstreamReleaseExternalPlugin),
	), nil
}

func initBKRepo(conf *config.FileService, logger logger.ILogger) (bkrepo.IHandler, error) {
	// init repo
	httpClient, err := restclient.NewHTTPClient(&ssl.TLSConfig{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init rest client: %w", err)
	}

	clientCap := &restclient.Capability{
		HTTPClient:           httpClient,
		Discover:             restdiscovery.NewDiscovery("bkrepo", []string{conf.Repo.Endpoint}),
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
		MetricOpts:           restclient.MetricOption{},
		Logger:               logger,
	}

	return bkrepo.New(clientCap, &bkrepo.Config{
		ProjectID: conf.Repo.ProjectID,
		RepoName:  conf.Repo.RepoName,
		Username:  conf.Repo.AccessKey,
		Password:  conf.Repo.SecretKey,
	})
}

func initMongoDB(conf *config.MongoDB) (*mongo.Client, error) {
	mongoClient, err := mongo.Connect(
		context.Background(),
		&mongoOptions.ClientOptions{
			Hosts: conf.Hosts,
			Auth: &mongoOptions.Credential{
				Username:      conf.Username,
				Password:      conf.Password,
				AuthSource:    conf.AuthSource,
				AuthMechanism: conf.AuthMechanism,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return mongoClient, nil
}
