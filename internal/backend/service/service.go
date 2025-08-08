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
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"runtime"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/admin"
	backendapiv3 "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/basic"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/callback"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/healthz"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/proxy"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/configpolicy"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	globalsettingsStorage "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/globalsettings"
	nodedeployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-deployment"
	nodeworkflow "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node-workflow"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/operation"
	operinstdataStorage "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/operinstdata"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/scheduleworkflow"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/trigger"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/watcher"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/blog"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/etcddiscover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/globalsettings"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rediscache"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/redsync"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/crypter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/system"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	apigwserver "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/iegtjj"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// DiscoveryNameApigw defines the name of apigateway discovery.
	DiscoveryNameApigw = "apigateway"
)

// Service defines a apigwserver that provides backend services.
// It manages the configuration, lifecycle, and various capabilities (e.g., cmdb, topo storage).
// The service's capabilities are accessed through its 'cap' field, while 'router' is used to route requests.
type Service struct {
	// conf holds the configuration for the service.
	conf *config.BackendService

	// ctx is used to control the service lifecycle (cancellation and timeouts).
	ctx context.Context

	// cancelFunc is used to cancel the service and all associated operations.
	cancelFunc context.CancelFunc

	// router is the entry point of the service, routing requests to different capabilities.
	servers []*restserver.Server

	// Note: Cap is initialized in the Start() and could not be used in other package.
	// Cap is the capability of the service.
	Cap *options.Capability

	// watcher maintains all watcher in the service.
	watcher *watcher.Watcher

	// instance is the discover instance of the service.
	instance discover.Instance
}

// NewService creates a new backend service.
// nolint: funlen,gocognit,gocyclo,cyclop,maintidx
// NOCC: golint/fnsize(func design is not suitable for splitting).
func NewService(conf *config.BackendService) (*Service, error) {
	if err := loadSystemInfo(conf); err != nil {
		return nil, err
	}

	svc := &Service{
		conf: conf,
		Cap: &options.Capability{
			Logger: blog.GlobalLogger{},
		},
		instance: discover.NewInstance("backend", nil),
	}

	svc.ctx, svc.cancelFunc = context.WithCancel(context.Background())

	var err error

	svc.Cap.Crypter, err = crypter.NewAESCrypter([]byte(conf.EncryptKey))
	if err != nil {
		return nil, err
	}

	for idx := range svc.conf.GSEDeployConfs {
		deployConf := deployconstant.DeployConf{
			Generation:    types.Generation(svc.conf.GSEDeployConfs[idx].Generation),
			OsType:        criteria.OSType(svc.conf.GSEDeployConfs[idx].OsType),
			BaseWorkDir:   svc.conf.GSEDeployConfs[idx].BaseWorkDir,
			BaseDeployDir: svc.conf.GSEDeployConfs[idx].BaseDeployDir,

			LogDir:             svc.conf.GSEDeployConfs[idx].Custom.LogDir,
			HostIDPath:         svc.conf.GSEDeployConfs[idx].Custom.HostIDPath,
			AgentDataIPCPath:   svc.conf.GSEDeployConfs[idx].Custom.AgentDataIPCPath,
			AgentPluginIPCPath: svc.conf.GSEDeployConfs[idx].Custom.AgentPluginIPCPath,
			EnvironDir:         svc.conf.GSEDeployConfs[idx].Custom.EnvironDir,
		}
		if err := deployconstant.SetDeployConf(deployConf); err != nil {
			return nil, fmt.Errorf("failed to set deploy conf: %w", err)
		}
	}

	svc.Cap.DiscoverProvider = etcddiscover.NewProviderEtcd(&conf.Etcd,
		etcddiscover.WithLogger(svc.Cap.Logger),
		etcddiscover.WithWatch(discover.ServiceNameBackend, discover.ServiceNameFile),
	)

	svc.Cap.CmdbHandler, err = newCMDBHandler(conf.CMDB, svc.Cap.Logger)
	if err != nil {
		return nil, err
	}

	svc.Cap.GSEHandler, err = newGSEHandler(conf.GSE)
	if err != nil {
		return nil, err
	}

	svc.Cap.FileHandler, err = newFileHandler(svc.Cap.DiscoverProvider)
	if err != nil {
		return nil, err
	}

	redisClient, err := initRedis(&conf.Redis)
	if err != nil {
		return nil, err
	}

	svc.Cap.LockerFactory = redsync.New(redisClient)

	mongoClient, err := initMongoDB(&conf.MongoDB)
	if err != nil {
		return nil, err
	}

	svc.Cap.StorageTopo, err = topo.NewStorage(mongoClient, conf.MongoDB.Database, svc.Cap.Logger)
	if err != nil {
		return nil, err
	}

	svc.Cap.StorageTrigger, err = trigger.NewStorage(mongoClient, conf.MongoDB.Database, svc.Cap.Logger)
	if err != nil {
		return nil, err
	}

	svc.Cap.StorageOperInst, err = operinstdataStorage.NewStorage(mongoClient, conf.MongoDB.Database, svc.Cap.Logger)
	if err != nil {
		return nil, err
	}

	svc.Cap.StorageOperation, err = operation.NewStorage(mongoClient, conf.MongoDB.Database, svc.Cap.Logger)
	if err != nil {
		return nil, err
	}

	svc.Cap.StorageNodeDeployment, err = nodedeployment.NewStorage(mongoClient, conf.MongoDB.Database, svc.Cap.Logger)
	if err != nil {
		return nil, err
	}

	svc.Cap.StorageNodeWorkflow, err = nodeworkflow.NewStorage(mongoClient, conf.MongoDB.Database, svc.Cap.Logger)
	if err != nil {
		return nil, err
	}

	svc.Cap.StorageScheduleWorkflow, err = scheduleworkflow.NewStorage(mongoClient, conf.MongoDB.Database, svc.Cap.Logger)
	if err != nil {
		return nil, err
	}

	svc.Cap.StorageRelease, err = release.NewStorage(mongoClient, conf.MongoDB.Database, svc.Cap.Logger)
	if err != nil {
		return nil, err
	}

	svc.Cap.StorageCredit, err = credit.NewStorage(mongoClient, conf.MongoDB.Database, svc.Cap.Logger, svc.Cap.Crypter)
	if err != nil {
		return nil, err
	}

	svc.Cap.StorageConfigPolicy, err = configpolicy.NewStorage(mongoClient, conf.MongoDB.Database, svc.Cap.Logger)
	if err != nil {
		return nil, err
	}

	svc.Cap.StorageGlobalSettings, err = globalsettingsStorage.NewStorage(
		mongoClient,
		conf.MongoDB.Database,
		svc.Cap.Logger,
	)
	if err != nil {
		return nil, err
	}

	svc.Cap.InstallerFileGroup, err = local.NewLocalDir(conf.InstallerFileGroup.FullPath, svc.Cap.Logger)
	if err != nil {
		return nil, err
	}

	svc.Cap.CreditVault, err = newCreditVault(conf.CreditVault, svc.Cap.Logger)
	if err != nil {
		return nil, err
	}

	svc.Cap.AuthIdentity, err = newBKJWTAuthIdentity(conf.APIGateWayServer)
	if err != nil {
		return nil, err
	}

	svc.Cap.ProxyMessager = relayhandler.NewServerMessager(relayhandler.ServerMessagerConfig{
		SlotID:        conf.GSE.PluginSlotID,
		Token:         conf.GSE.PluginSlotToken,
		AppCode:       conf.GSE.AppCode,
		AppSecret:     conf.GSE.AppSecret,
		GSEBaseURL:    conf.GSE.Endpoints[0],
		SkipTLSVerify: conf.GSE.TLS.InsecureSkipVerify,
		RedisClient:   redisClient,
		Logger:        svc.Cap.Logger,
	})

	svc.Cap.Manager, err = manager.NewManager(manager.Config{
		CmdbHandler:           svc.Cap.CmdbHandler,
		GSEHandler:            svc.Cap.GSEHandler,
		Provider:              svc.Cap.DiscoverProvider,
		InstallerFileGroup:    svc.Cap.InstallerFileGroup,
		LockerFactory:         svc.Cap.LockerFactory,
		StorageTopo:           svc.Cap.StorageTopo,
		StorageRelease:        svc.Cap.StorageRelease,
		StorageNodeDeployment: svc.Cap.StorageNodeDeployment,
		StorageNodeWorkflow:   svc.Cap.StorageNodeWorkflow,
		StorageTrigger:        svc.Cap.StorageTrigger,
		StorageOperation:      svc.Cap.StorageOperation,
		StorageOperInst:       svc.Cap.StorageOperInst,
		StorageSchedule:       svc.Cap.StorageScheduleWorkflow,
		StorageHostCredit:     svc.Cap.StorageCredit,
		HostPasswordVault:     svc.Cap.CreditVault,
		FileHandler:           svc.Cap.FileHandler,
		WorkflowConfig: manager.WorkflowConfig{
			WorkNodeNum: conf.Workflow.WorkerNum,
			Redis: manager.RedisConfig{
				Addr:     fmt.Sprintf("%s:%d", conf.Redis.Host, conf.Redis.Port),
				Password: conf.Redis.Password,
				DB:       conf.Redis.DB,
			},
		},
	}, blog.GlobalLogger{})
	if err != nil {
		return nil, err
	}

	globalsettings.InitGlobalSettings(svc.Cap.StorageGlobalSettings)

	svc.watcher, err = watcher.NewWatcher(
		watcher.Config{
			CmdbHandler: svc.Cap.CmdbHandler,
			StorageTopo: svc.Cap.StorageTopo,
			Logger:      svc.Cap.Logger,
			Cache:       rediscache.NewRedisCache(redisClient, rediscache.DefaultTimeout),
		},
	)
	if err != nil {
		return nil, err
	}

	svc.registerRestServer(conf)

	return svc, nil
}

func loadSystemInfo(conf *config.BackendService) error {
	system.SetEnv(conf.System.Env)
	if err := system.SetEdition(system.Edition(conf.System.Edition)); err != nil {
		return fmt.Errorf("failed to set edition: %w", err)
	}

	return nil
}

// nolint: funlen
func (svc *Service) registerRestServer(conf *config.BackendService) {
	apigwRequestIDSetter := apigwserver.NewBKAPIRequestIDSetter()
	tenantIDSetter := restserver.NewTenantIDSetter()

	httpServer := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:            string(discover.EndpointNameBackendBasic),
			IP:              conf.HTTPServer.BindIP,
			Port:            conf.HTTPServer.Port,
			LogWriter:       loggerWriterAdaptor{},
			RequestIDSetter: apigwRequestIDSetter,
			TenantIDSetter:  tenantIDSetter,
		},
		restserver.WithPing(),
		withHealthz(svc.Cap),
		withMetrics(svc.Cap),
		withAPIV3(svc.Cap),
		withBasic(svc.Cap),
	)
	svc.servers = append(svc.servers, httpServer)
	svc.instance.Update(discover.EndpointNameBackendBasic, discover.Endpoint{
		IPV4: conf.HTTPServer.AdvertiseIPV4,
		IPV6: conf.HTTPServer.AdvertiseIPV6,
		Port: conf.HTTPServer.Port,
	})

	requestIDSetter := restserver.NewRequestIDSetter()
	callbackServer := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:            string(discover.EndpointNameBackendCallback),
			IP:              conf.CallbackServer.BindIP,
			Port:            conf.CallbackServer.Port,
			LogWriter:       loggerWriterAdaptor{},
			RequestIDSetter: requestIDSetter,
			TenantIDSetter:  tenantIDSetter,
		},
		restserver.WithPing(),
		withCallback(svc.Cap),
	)
	svc.servers = append(svc.servers, callbackServer)
	svc.instance.Update(discover.EndpointNameBackendCallback, discover.Endpoint{
		IPV4: conf.CallbackServer.AdvertiseIPV4,
		IPV6: conf.CallbackServer.AdvertiseIPV6,
		Port: conf.CallbackServer.Port,
	})

	proxyServer := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:            string(discover.EndpointNameBackendPorxy),
			IP:              conf.ProxyServer.BindIP,
			Port:            conf.ProxyServer.Port,
			RequestIDSetter: requestIDSetter,
			TenantIDSetter:  tenantIDSetter,
			LogWriter:       loggerWriterAdaptor{},
		},
		restserver.WithPing(),
		withProxy(svc.Cap),
	)
	svc.servers = append(svc.servers, proxyServer)
	svc.instance.Update(discover.EndpointNameBackendPorxy, discover.Endpoint{
		IPV4: conf.ProxyServer.AdvertiseIPV4,
		IPV6: conf.ProxyServer.AdvertiseIPV6,
		Port: conf.ProxyServer.Port,
	})

	adminServer := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:            string(discover.EndpointNameBackendAdmin),
			IP:              conf.AdminServer.BindIP,
			Port:            conf.AdminServer.Port,
			RequestIDSetter: requestIDSetter,
			TenantIDSetter:  tenantIDSetter,
			LogWriter:       loggerWriterAdaptor{},
		},
		restserver.WithPing(),
		withHealthz(svc.Cap),
		withMetrics(svc.Cap),
		withAdmin(svc.Cap),
	)
	svc.servers = append(svc.servers, adminServer)
}

func initRedis(conf *config.Redis) (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", conf.Host, conf.Port),
		Password: conf.Password,
		DB:       conf.DB,
	})

	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return redisClient, nil
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

// loggerWriterAdaptor implements rest.LoggerWriter.
type loggerWriterAdaptor struct{}

func (l loggerWriterAdaptor) InfoWriter() io.Writer {
	return blog.WriterInfo{}
}

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

// withApiV3 load api v3.
func withAPIV3(capability *options.Capability) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		backendapiv3.Load(rg, capability)
	}
}

// withBasic load basic.
func withBasic(capability *options.Capability) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		basic.Load(rg, capability)
	}
}

// withAdmin load admin.
func withAdmin(capability *options.Capability) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		admin.Load(rg, capability)
	}
}

// withCallback load callback.
func withCallback(capability *options.Capability) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		callback.Load(rg, capability)
	}
}

// withProxy load proxy.
func withProxy(capability *options.Capability) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		proxy.Load(rg, capability)
	}
}

// newCMDBHandler.
func newCMDBHandler(conf config.CMDB, logger logger.Logger) (cmdb.IHandler, error) {
	apiGwClientConfig := newAPIGwClientConfig(&conf.APIGatewayClient)
	apiGwClientCapability, err := newAPIGwClientCapability(&conf.APIGatewayClient)
	if err != nil {
		return nil, err
	}

	apiGwClientCapability.Name = "cmdb"
	cmdbHandler, err := cmdb.New(apiGwClientCapability, &cmdb.Config{
		SupplierAccount:   conf.SupplierAccount,
		APIGWClientConfig: apiGwClientConfig,
	}, cmdb.WithLogger(logger))
	if err != nil {
		return nil, err
	}

	return cmdbHandler, nil
}

func newCreditVault(conf config.CreditVault, logger logger.Logger) (creditvault.ICreditVault, error) {
	hostPasswordVault, err := newHostPasswordVault(conf.HostCreditVault, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to new credit password vault: %w", err)
	}

	vault := creditvault.New(creditvault.WithHostPasswordVault(hostPasswordVault))

	return vault, nil
}

func newBKJWTAuthIdentity(conf config.APIGateWayServer) (*apigwserver.BKGWJWTAuthIdentity, error) {
	publickeyPem, err := base64.StdEncoding.DecodeString(conf.PublickeyPem)
	if err != nil {
		return nil, fmt.Errorf("failed to new bk jwt auth identity: %w", err)
	}

	authIdentity := apigwserver.NewBKGWJWTAuthIdentity(publickeyPem)

	return authIdentity, nil
}

func newHostPasswordVault(conf config.HostCreditVault, logger logger.Logger) (creditvault.IHostPasswordVault, error) {
	if !conf.Enable {
		return &creditvault.DisabledHostPasswordVault{}, nil
	}

	switch conf.Type {
	case "iegtjj":
		iegtjjHandler, err := newIEGTJJHandler(conf.IEGTJJ, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to new host password vault: %w", err)
		}

		return iegtjjHandler, nil

	default:
		return nil, fmt.Errorf("unknown host password vault type: %s", conf.Type)
	}
}

func newIEGTJJHandler(conf config.IEGTJJ, logger logger.Logger) (iegtjj.IHandler, error) {
	// apiGwClientConfig := newAPIGwClientConfig(&conf.APIGatewayClient)
	// TODO: 等待 iegtjj 迁移到 apigw, 将此处替换为 apigwclient.Config
	apiGwClientCapability, err := newAPIGwClientCapability(&conf.APIGatewayClient)
	if err != nil {
		return nil, err
	}

	apiGwClientCapability.Name = "iegtjj"
	iegtjjHandler, err := iegtjj.New(apiGwClientCapability, &iegtjj.Config{
		Key:       conf.Key,
		SecretKey: conf.SecretKey,
	}, iegtjj.WithLogger(logger))
	if err != nil {
		return nil, fmt.Errorf("failed to new iegtjj handler: %w", err)
	}

	return iegtjjHandler, nil
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
		APIGWClientConfig: apiGwClientConfig,
	})
	if err != nil {
		return nil, err
	}

	return gseHandler, nil
}

// newFileHandler creates a new file handler.
func newFileHandler(discov discover.Discover) (file.IHandler, error) {
	httpClient, err := restclient.NewHTTPClient(&ssl.TLSConfig{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, err
	}

	clientCap := &restclient.Capability{
		HTTPClient: httpClient,
		Discover: restdiscovery.NewServiceDiscovery(
			discov,
			discover.ServiceNameFile,
			discover.EndpointNameFileAdmin),
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
		MetricOpts:           restclient.MetricOption{},
		Logger:               logger.LoggerDefault{},
	}

	return file.New(clientCap, &file.Config{})
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
func newAPIGwClientConfig(conf *config.APIGatewayClient) apigwclient.Config {
	return apigwclient.Config{
		Endpoints:   conf.Endpoints,
		AppCode:     conf.AppCode,
		AppSecret:   conf.AppSecret,
		User:        conf.User,
		AuthMode:    apigwclient.AuthMode(conf.AuthMode),
		BkTicket:    conf.BkTicket,
		BkToken:     conf.BkToken,
		AccessToken: conf.AccessToken,
	}
}

// Start starts the backend service.
func (svc *Service) Start() error {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if err := svc.Cap.Start(svc.ctx); err != nil {
		return err
	}

	// start watcher right after all capabilities started.
	if err := svc.watcher.Start(svc.ctx); err != nil {
		return err
	}

	// start servers
	gp := gopool.NewPool()
	for idx := range svc.servers {
		server := svc.servers[idx]

		// apigwserver start will block until apigwserver stop, so we need to run it in a goroutine.
		fn := func() error {
			blog.Infof("started apigwserver. name(%s), ip(%s), port(%d)", server.Name(), server.IP(), server.Port())

			if err := server.Start(); err != nil {
				return err
			}

			return nil
		}
		gp.Go(fn)
	}

	// after all servers brings up, register the instance into discover provider.
	if err := svc.Cap.DiscoverProvider.Register(discover.ServiceNameBackend, svc.instance); err != nil {
		blog.Errorf("failed to register instance: %v", err)
		return err
	}

	// wait until all servers stopped or application error.
	if err := gp.Wait(); err != nil {
		blog.Errorf("failed to start servers: %v", err)
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
		blog.Errorf("failed to shutdown capability: %v", err)
		return err
	}

	blog.CloseLogs()

	return nil
}
