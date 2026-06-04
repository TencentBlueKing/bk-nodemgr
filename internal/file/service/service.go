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
	"crypto/tls"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/file/manager"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/router/admin"
	fileapiv3 "github.com/TencentBlueKing/bk-nodemgr/internal/file/router/api-v3"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/router/healthz"
	packageEventStg "github.com/TencentBlueKing/bk-nodemgr/internal/file/storage/packageevent"
	storageRelease "github.com/TencentBlueKing/bk-nodemgr/internal/file/storage/release"
	storageTopo "github.com/TencentBlueKing/bk-nodemgr/internal/file/storage/topo"
	storageUpload "github.com/TencentBlueKing/bk-nodemgr/internal/file/storage/upload"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover/etcddiscover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filecache"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/bkrepo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tracing"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/version"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	clientNameRepo = "bkrepo"
	clientNameGse  = "gse"

	mongoMaxPoolSize     = uint64(500)
	mongoMinPoolSize     = uint64(5)
	mongoMaxConnIdleTime = 3 * time.Minute

	serverName = "file"
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

	// authIdentityValidMap defines the mapping between auth identity and auth identity handler.
	authIdentityValidMap map[config.AuthIdentity]struct{}
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

	svc.ctx, svc.cancelFunc = contextx.WithCancel(contextx.New(contextx.Background()))

	if err := svc.initialStaticsConfigs(); err != nil {
		return nil, fmt.Errorf("failed to initialize static configs: %w", err)
	}

	if err := svc.initTracing(); err != nil {
		return nil, fmt.Errorf("failed to init tracing: %w", err)
	}

	if err := svc.initialCapability(svc.ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize capability: %w", err)
	}

	if err := svc.registerRestServer(); err != nil {
		return nil, fmt.Errorf("failed to register http rest server: %w", err)
	}

	return svc, nil
}

// nolint: unparam
func (svc *Service) initialStaticsConfigs() error {
	svc.authIdentityValidMap = map[config.AuthIdentity]struct{}{
		config.AuthIdentityNone:       {},
		config.AuthIdentityRestServer: {},
	}

	return nil
}

// nolint: funlen
func (svc *Service) initialCapability(nCtx contextx.IContext) error {
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
	if err = svc.initialManager(nCtx); err != nil {
		return fmt.Errorf("failed to initial manager: %w", err)
	}

	return nil
}

func (svc *Service) newGSEHandler() (gse.IHandler, error) {
	apiGWUserConfig := newAPIGWUserConfig(&svc.conf.GSE.APIGatewayClient)
	apiGwClientCapability, err := newAPIGwClientCapability(clientNameGse, &svc.conf.GSE.APIGatewayClient)
	if err != nil {
		return nil, fmt.Errorf("failed to new apigw client for gse: %w", err)
	}

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

	traceSvc, err := tracing.G().NewService(tracing.ServiceConfig{
		ServiceName:     svc.conf.Repo.TraceServiceName,
		ServiceCategory: tracing.ServiceCategoryHTTP,
		SampleRate:      svc.conf.Repo.TraceSampleRate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to new trace service: %w", err)
	}

	clientCap := &restclient.Capability{
		Name:                 clientNameRepo,
		HTTPClient:           httpClient,
		Discover:             restdiscovery.NewDiscovery(clientNameRepo, []string{svc.conf.Repo.Endpoint}),
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
		MetricOpts:           restclient.MetricOption{},
		TraceSvc:             traceSvc,
	}

	return bkrepo.New(clientCap, &bkrepo.Config{
		ProjectID: svc.conf.Repo.ProjectID,
		RepoName:  svc.conf.Repo.RepoName,
		Username:  svc.conf.Repo.AccessKey,
		Password:  svc.conf.Repo.SecretKey,
	})
}

func (svc *Service) newMongoClient() (*mongo.Client, error) {
	var tlsConfig *tls.Config
	if svc.conf.MongoDB.TLS.CAFile != "" && svc.conf.MongoDB.TLS.CertFile != "" && svc.conf.MongoDB.TLS.KeyFile != "" {
		sslConf := &ssl.TLSConfig{
			CAFile:   svc.conf.MongoDB.TLS.CAFile,
			CertFile: svc.conf.MongoDB.TLS.CertFile,
			KeyFile:  svc.conf.MongoDB.TLS.KeyFile,
			Password: svc.conf.MongoDB.TLS.Password,
		}

		var err error
		tlsConfig, err = sslConf.NewClientTLSConf()
		if err != nil {
			return nil, fmt.Errorf("failed to create tls config: %w", err)
		}
	}

	maxConnIdleTime := mongoMaxConnIdleTime
	maxPoolSize := mongoMaxPoolSize
	minPoolSize := mongoMinPoolSize
	var replicaSet *string
	if svc.conf.MongoDB.ReplicaSet != "" {
		replicaSet = &svc.conf.MongoDB.ReplicaSet
	}
	mongoClient, err := mongo.Connect(
		contextx.Background(),
		&mongoOptions.ClientOptions{
			AppName: &svc.conf.MongoDB.AppName,
			Auth: &mongoOptions.Credential{
				AuthMechanism: svc.conf.MongoDB.AuthMechanism,
				AuthSource:    svc.conf.MongoDB.AuthSource,
				Username:      svc.conf.MongoDB.Username,
				Password:      svc.conf.MongoDB.Password,
				PasswordSet:   true,
			},
			ReplicaSet:      replicaSet,
			Hosts:           svc.conf.MongoDB.Hosts,
			ReadPreference:  readpref.Primary(),
			TLSConfig:       tlsConfig,
			MaxConnIdleTime: &maxConnIdleTime,
			MaxPoolSize:     &maxPoolSize,
			MinPoolSize:     &minPoolSize,
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

	svc.Cap.StorageEvent, err = packageEventStg.NewStorage(
		svc.Cap.MongoClient,
		svc.conf.MongoDB.Database)
	if err != nil {
		return fmt.Errorf("failed to create event storage: %w", err)
	}

	return nil
}

// nolint: funlen
func (svc *Service) initialManager(nCtx contextx.IContext) error {
	// init upstream origin file groups from bkrepo.
	upstreamOriginAgentFG, err := svc.Cap.BKRepo.EnsureFileGroup(nCtx, "origin/agent")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream origin agent file group: %w", err)
	}
	upstreamOriginServerFG, err := svc.Cap.BKRepo.EnsureFileGroup(nCtx, "origin/server")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream origin server file group: %w", err)
	}
	upstreamOriginProxyFG, err := svc.Cap.BKRepo.EnsureFileGroup(nCtx, "origin/proxy")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream origin proxy file group: %w", err)
	}
	upstreamOriginCertFG, err := svc.Cap.BKRepo.EnsureFileGroup(nCtx, "origin/cert")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream origin cert file group: %w", err)
	}
	upstreamOriginBinToolFG, err := svc.Cap.BKRepo.EnsureFileGroup(nCtx, "origin/bintool")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream origin bin tool file group: %w", err)
	}
	upstreamOriginPluginV2, err := svc.Cap.BKRepo.EnsureFileGroup(nCtx, "origin/v2/plugin")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream origin plugin file group: %w", err)
	}
	upstreamOriginExternalPluginV2, err := svc.Cap.BKRepo.EnsureFileGroup(nCtx, "origin/v2/external_plugin")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream origin external plugin file group: %w", err)
	}
	upstreamOriginPluginV3, err := svc.Cap.BKRepo.EnsureFileGroup(nCtx, "origin/v3/plugin")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream origin plugin file group: %w", err)
	}

	// init upstream release file groups from bkrepo.
	upstreamReleaseAgentFG, err := svc.Cap.BKRepo.EnsureFileGroup(nCtx, "release/agent")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream release agent file group: %w", err)
	}
	upstreamReleaseProxyFg, err := svc.Cap.BKRepo.EnsureFileGroup(nCtx, "release/proxy")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream release proxy file group: %w", err)
	}
	upstreamRealseCertFG, err := svc.Cap.BKRepo.EnsureFileGroup(nCtx, "release/cert")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream release cert file group: %w", err)
	}
	upstreamReleaseBintoolFG, err := svc.Cap.BKRepo.EnsureFileGroup(nCtx, "release/bintool")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream release bin tool file group: %w", err)
	}
	upstreamOriginPluginBinToolV2FG, err := svc.Cap.BKRepo.EnsureFileGroup(nCtx, "origin/plugin_bintool")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream origin plugin bin tool file group: %w", err)
	}
	upstreamReleasePluginBinToolFG, err := svc.Cap.BKRepo.EnsureFileGroup(nCtx, "release/plugin_bintool")
	if err != nil {
		return fmt.Errorf("failed to ensure upstream release bin tool file group: %w", err)
	}
	upstreamReleasePlugin, err := svc.Cap.BKRepo.EnsureFileGroup(nCtx, "release/plugin")
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

	// filecache.New creates the cache directory if it does not exist.
	cacheDir := filepath.Join(svc.conf.WorkspaceFileGroup.FullPath, "cache")

	fc, err := filecache.New(nCtx, cacheDir, filecache.Options{
		ExpirationTime: time.Duration(svc.conf.FileCache.ExpirationHours) * time.Hour,
		GCInterval:     time.Duration(svc.conf.FileCache.GCIntervalHours) * time.Hour,
		RestoreOnStart: svc.conf.FileCache.RestoreOnStart,
	})
	if err != nil {
		return fmt.Errorf("failed to init file cache: %w", err)
	}

	svc.Cap.Manager = manager.New(
		manager.WithUpstreamOriginAgentFileGroup(upstreamOriginAgentFG),
		manager.WithUpstreamOriginServerFileGroup(upstreamOriginServerFG),
		manager.WithUpstreamOriginProxyFileGroup(upstreamOriginProxyFG),
		manager.WithUpstreamOriginCertFileGroup(upstreamOriginCertFG),
		manager.WithUpstreamOriginBinToolFileGroup(upstreamOriginBinToolFG),
		manager.WithUpstreamOriginPluginBinToolV2FileGroup(upstreamOriginPluginBinToolV2FG),
		manager.WithUpstreamReleaseAgentFileGroup(upstreamReleaseAgentFG),
		manager.WithUpstreamReleaseProxyFileGroup(upstreamReleaseProxyFg),
		manager.WithUpstreamReleaseCertFileGroup(upstreamRealseCertFG),
		manager.WithUpstreamReleaseBinToolFileGroup(upstreamReleaseBintoolFG),
		manager.WithUpstreamReleasePluginBinToolFileGroup(upstreamReleasePluginBinToolFG),
		manager.WithTempFileGroup(tempFG),
		manager.WithInstallerFileGroup(installerFG),
		manager.WithFileCache(fc),
		manager.WithStorageUpload(svc.Cap.StorageUpload),
		manager.WithStorageRelease(svc.Cap.StorageRelease),
		manager.WithStorageTopo(svc.Cap.StorageTopo),
		manager.WithStorageEvent(svc.Cap.StorageEvent),
		manager.WithAdvertiseIPV4(svc.conf.BasicServer.AdvertiseIPV4),
		manager.WithAdvertiseIPV6(svc.conf.BasicServer.AdvertiseIPV6),
		manager.WithMount(svc.conf.MountHostDir, svc.conf.WorkspaceFileGroup.FullPath),
		manager.WithGSEHandler(svc.Cap.GSEHandler),
		manager.WithUpstreamOriginPluginV2FileGroup(upstreamOriginPluginV2),
		manager.WithUpstreamOriginExternalPluginV2FileGroup(upstreamOriginExternalPluginV2),
		manager.WithUpstreamOriginPluginV3FileGroup(upstreamOriginPluginV3),
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

func newAuthIdentity(conf config.HTTPServer) (restserver.IAuthIdentity, error) {
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
			Name:             string(discover.EndpointNameFileInfo),
			IP:               svc.conf.InfoServer.BindIP,
			IPV6:             svc.conf.InfoServer.BindIPV6,
			Port:             svc.conf.InfoServer.Port,
			TLSConfig:        svc.conf.InfoServer.TLSConfig,
			RequestIDSetter:  restserver.NewRequestIDSetter(),
			TraceServiceName: svc.conf.InfoServer.TraceServiceName,
			TraceSampleRate:  svc.conf.InfoServer.TraceSampleRate,
		},
		restserver.WithPing(),
		withHealthz(svc.Cap),
		withMetrics(svc.Cap),
	)
	if err != nil {
		return fmt.Errorf("failed to register info server: %w", err)
	}

	svc.servers = append(svc.servers, server)
	svc.instance.Update(discover.EndpointNameFileInfo, discover.Endpoint{
		IPV4: svc.conf.InfoServer.AdvertiseIPV4,
		IPV6: svc.conf.InfoServer.AdvertiseIPV6,
		Port: svc.conf.InfoServer.Port,
	})

	return nil
}

func (svc *Service) registerAdminServer() error {
	if svc.conf.AdminServer.AuthIdentity != config.AuthIdentityNone &&
		svc.conf.AdminServer.AuthIdentity != config.AuthIdentityRestServer {

		return fmt.Errorf("no support this auth identity, auth-identity(%s), support auth-identity(%v, %v)",
			svc.conf.AdminServer.AuthIdentity, config.AuthIdentityNone, config.AuthIdentityRestServer)
	}

	authIdentity, err := newAuthIdentity(svc.conf.AdminServer)
	if err != nil {
		return fmt.Errorf("failed to new auth identity: %w", err)
	}

	server, err := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:             string(discover.EndpointNameFileAdmin),
			IP:               svc.conf.AdminServer.BindIP,
			IPV6:             svc.conf.AdminServer.BindIPV6,
			Port:             svc.conf.AdminServer.Port,
			TLSConfig:        svc.conf.AdminServer.TLSConfig,
			RequestIDSetter:  restserver.NewRequestIDSetter(),
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
	svc.instance.Update(discover.EndpointNameFileAdmin, discover.Endpoint{
		IPV4: svc.conf.AdminServer.AdvertiseIPV4,
		IPV6: svc.conf.AdminServer.AdvertiseIPV6,
		Port: svc.conf.AdminServer.Port,
	})

	return nil
}

func (svc *Service) registerBasicServer() error {
	_, valid := svc.authIdentityValidMap[svc.conf.BasicServer.AuthIdentity]
	if !valid {
		return fmt.Errorf("no support this auth identity, auth-identity(%s), use-one-of(%v)",
			svc.conf.BasicServer.AuthIdentity, conv.MapKeyToSlice(svc.authIdentityValidMap))
	}

	authIdentity, err := newAuthIdentity(svc.conf.BasicServer)
	if err != nil {
		return err
	}

	server, err := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:             string(discover.EndpointNameFileBasic),
			IP:               svc.conf.BasicServer.BindIP,
			IPV6:             svc.conf.BasicServer.BindIPV6,
			Port:             svc.conf.BasicServer.Port,
			TLSConfig:        svc.conf.BasicServer.TLSConfig,
			RequestIDSetter:  restserver.NewRequestIDSetter(),
			TraceServiceName: svc.conf.BasicServer.TraceServiceName,
			TraceSampleRate:  svc.conf.BasicServer.TraceSampleRate,
		},
		restserver.WithPing(),
		withAPIV3Basic(svc.Cap,
			restserver.MiddlewareAuth(authIdentity),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to register basic server: %w", err)
	}

	svc.servers = append(svc.servers, server)
	svc.instance.Update(discover.EndpointNameFileBasic, discover.Endpoint{
		IPV4: svc.conf.BasicServer.AdvertiseIPV4,
		IPV6: svc.conf.BasicServer.AdvertiseIPV6,
		Port: svc.conf.BasicServer.Port,
	})

	return nil
}

func (svc *Service) registerDownloadServer() error {
	if svc.conf.DownloadServer.AuthIdentity != config.AuthIdentityNone {
		return fmt.Errorf("no support this auth identity, auth-identity(%s), support auth-identity(%v)",
			svc.conf.DownloadServer.AuthIdentity, config.AuthIdentityNone)
	}

	authIdentity, err := newAuthIdentity(svc.conf.DownloadServer)
	if err != nil {
		return err
	}

	server, err := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:             string(discover.EndpointNameFileDownload),
			IP:               svc.conf.DownloadServer.BindIP,
			IPV6:             svc.conf.DownloadServer.BindIPV6,
			Port:             svc.conf.DownloadServer.Port,
			TLSConfig:        svc.conf.DownloadServer.TLSConfig,
			RequestIDSetter:  restserver.NewRequestIDSetter(),
			TraceServiceName: svc.conf.DownloadServer.TraceServiceName,
			TraceSampleRate:  svc.conf.DownloadServer.TraceSampleRate,
		},
		restserver.WithPing(),
		withAPIV3Download(svc.Cap,
			restserver.MiddlewareAuth(authIdentity),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to register download server: %w", err)
	}

	svc.servers = append(svc.servers, server)
	svc.instance.Update(discover.EndpointNameFileDownload, discover.Endpoint{
		IPV4: svc.conf.DownloadServer.AdvertiseIPV4,
		IPV6: svc.conf.DownloadServer.AdvertiseIPV6,
		Port: svc.conf.DownloadServer.Port,
	})

	return nil
}

// newAPIGwClientCapability creates a new api-gateway client capability.
func newAPIGwClientCapability(name string, conf *config.APIGatewayClient) (*restclient.Capability, error) {
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

	traceSvc, err := tracing.G().NewService(tracing.ServiceConfig{
		ServiceName:     conf.TraceServiceName,
		ServiceCategory: tracing.ServiceCategoryHTTP,
		SampleRate:      conf.TraceSampleRate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to new trace service: %w", err)
	}

	clientCap := &restclient.Capability{
		Name:                 name,
		HTTPClient:           httpClient,
		Discover:             restdiscovery.NewDiscovery(name, conf.Endpoints),
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
		MetricOpts:           restclient.MetricOption{},
		TraceSvc:             traceSvc,
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

// withAdmin load admin.
func withAdmin(capability *options.Capability, middleware ...gin.HandlerFunc) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		admin.Load(rg, capability, middleware...)
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

// withAPIV3Basic load api v3 basic.
func withAPIV3Basic(capability *options.Capability, middleware ...gin.HandlerFunc) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		fileapiv3.LoadBasicAPIs(rg, capability, middleware...)
	}
}

// withAPIV3Download load api v3 download.
func withAPIV3Download(capability *options.Capability, middleware ...gin.HandlerFunc) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		fileapiv3.LoadDownloadAPIs(rg, capability, middleware...)
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

	err := svc.Cap.GracefulShutdown()
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to gracefully shutdown capability")

		return err
	}

	logger.G.Sys().Info("file service gracefully shutdown")

	return nil
}

func (svc *Service) initTracing() error {
	tracingConf := tracing.Config{
		Exporter: tracing.ExporterConfig{
			ExporterType: tracing.ExporterType(svc.conf.Tracing.ExporterType),
		},
		Environment: system.GetEnv(),
		Namespace:   serverName,
		InstanceID:  svc.conf.Tracing.InstanceID,
		Version:     version.Version().Version,
	}

	if tracingConf.Exporter.ExporterType == tracing.ExporterTypeOTLP {
		tracingConf.Exporter.OTLPConfig = &tracing.OTLPConfig{
			Endpoint: svc.conf.Tracing.OTLPEndpoint,
			Insecure: svc.conf.Tracing.OTLPInsecure,
			Headers:  svc.conf.Tracing.OTLPHeaders,
		}
	}

	if err := tracing.Init(tracingConf); err != nil {
		return fmt.Errorf("failed to init tracing: %w", err)
	}

	return nil
}
