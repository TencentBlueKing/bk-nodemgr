/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

// Package service provides backend service.
package service

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authProvider "github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth/provider"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/periodictask"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/admin"
	backendapiv3 "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/healthz"
	cipherStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/cipher"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/configpolicy"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/deploypolicy"
	globalsettingsStorage "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/globalsettings"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	pkgStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/pkg"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	tenantStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/tenant"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/access"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover/etcddiscover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filecache"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/globalsettings"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rediscache"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/redsync"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/crypter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	apigwserver "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/iamv3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/iegtjj"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/usermanager"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tracing"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/version"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
)

const (
	clientNameCMDB        = "cmdb"
	clientNameGSE         = "gse"
	clientNameUserManager = "usermanager"
	clientNameIEGTJJ      = "iegtjj"
	clientNameFile        = "file"
	clientNameIAM         = "iam-v3"

	mongoMaxPoolSize     = uint64(500)
	mongoMinPoolSize     = uint64(5)
	mongoMaxConnIdleTime = 3 * time.Minute

	serverName = "backend"
)

// Service defines a apigwserver that provides backend services.
// It manages the configuration, lifecycle, and various capabilities (e.g., cmdb, topo storage).
// The service's capabilities are accessed through its 'cap' field, while 'router' is used to route requests.
type Service struct {
	// conf holds the configuration for the service.
	conf *config.BackendService

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

// NewService creates a new backend service.
// nolint: funlen,gocognit,gocyclo,cyclop,maintidx
// NOCC: golint/fnsize(func design is not suitable for splitting).
func NewService(conf *config.BackendService) (*Service, error) {
	svc := &Service{
		conf:     conf,
		Cap:      &options.Capability{},
		instance: discover.NewInstance(string(discover.ServiceNameBackend), nil),
	}

	svc.ctx, svc.cancelFunc = contextx.WithCancel(contextx.New(contextx.Background()))

	if err := svc.initialStaticsConfigs(); err != nil {
		return nil, fmt.Errorf("failed to initialize static configs: %w", err)
	}

	if err := svc.initTracing(); err != nil {
		return nil, fmt.Errorf("failed to init tracing: %w", err)
	}

	if err := svc.initialCapability(); err != nil {
		return nil, fmt.Errorf("failed to initialize capability: %w", err)
	}

	if err := svc.registerRestServer(); err != nil {
		return nil, fmt.Errorf("failed to register http rest server: %w", err)
	}

	return svc, nil
}

func (svc *Service) initialStaticsConfigs() error {
	// initial system edition.
	system.SetEnv(svc.conf.System.Env)
	if err := system.SetEdition(system.Edition(svc.conf.System.Edition)); err != nil {
		return fmt.Errorf("failed to set system edition, edition(%s): %w", svc.conf.System.Edition, err)
	}

	// initial access virtual user.
	access.SetVirtualUser(svc.conf.Access.VirtualUser)

	// initial gse deploy conf.
	for idx := range svc.conf.GSEDeployConfs {
		deployConf := deployconstant.DeployConf{
			Generation:       types.Generation(svc.conf.GSEDeployConfs[idx].Generation),
			OsType:           criteria.OSType(svc.conf.GSEDeployConfs[idx].OsType),
			BaseDeployDir:    svc.conf.GSEDeployConfs[idx].BaseDeployDir,
			BaseWorkDir:      svc.conf.GSEDeployConfs[idx].BaseWorkDir,
			ManualScriptPath: svc.conf.GSEDeployConfs[idx].ManualScriptPath,
		}
		if err := deployconstant.SetDeployConf(deployConf); err != nil {
			return fmt.Errorf("failed to set deploy conf: %w", err)
		}

		nodeDeployConf := deployconstant.NodeDeployConf{
			DeployConf:        deployConf,
			LogDir:            svc.conf.GSEDeployConfs[idx].Custom.LogDir,
			ExtraConfigDir:    svc.conf.GSEDeployConfs[idx].Custom.ExtraConfigDir,
			DataIPC:           svc.conf.GSEDeployConfs[idx].Custom.DataIPC,
			PluginIPC:         svc.conf.GSEDeployConfs[idx].Custom.PluginIPC,
			ProxyFileCacheDir: svc.conf.GSEDeployConfs[idx].Custom.ProxyFileCacheDir,
			ZoneID:            svc.conf.GSEDeployConfs[idx].Custom.ZoneID,
			CityID:            svc.conf.GSEDeployConfs[idx].Custom.CityID,
			EventDataIDConfs:  buildNodeEventDataIDConfs(svc.conf.GSEDeployConfs[idx].Custom.EventDataIDs),
		}
		if err := deployconstant.SetNodeDeployConf(nodeDeployConf); err != nil {
			return fmt.Errorf("failed to set node deploy conf: %w", err)
		}

		pluginDeployConf := deployconstant.PluginDeployConf{
			DeployConf:      deployConf,
			LogDir:          svc.conf.GSEDeployConfs[idx].PluginCustom.LogDir,
			HostIDPath:      svc.conf.GSEDeployConfs[idx].PluginCustom.HostIDPath,
			CommonConstants: svc.conf.GSEDeployConfs[idx].PluginCustom.CommonConstants,
		}
		if err := deployconstant.SetPluginDeployConf(pluginDeployConf); err != nil {
			return fmt.Errorf("failed to set pluginStg deploy conf: %w", err)
		}
	}

	svc.authIdentityValidMap = map[config.AuthIdentity]struct{}{
		config.AuthIdentityNone:       {},
		config.AuthIdentityAPIGW:      {},
		config.AuthIdentityRestServer: {},
	}

	return nil
}

func buildNodeEventDataIDConfs(eventDataIDs []config.GSEDeployEventDataID) map[string]deployconstant.NodeEventDataIDConf {
	if len(eventDataIDs) == 0 {
		return nil
	}

	eventDataIDConfs := make(map[string]deployconstant.NodeEventDataIDConf, len(eventDataIDs))
	for idx := range eventDataIDs {
		eventDataIDConfs[eventDataIDs[idx].TenantID] = deployconstant.NodeEventDataIDConf{
			AgentBaseAlarmEventDataID: eventDataIDs[idx].AgentBaseAlarmEventDataID,
			TaskProcEventDataID:       eventDataIDs[idx].TaskProcEventDataID,
		}
	}

	return eventDataIDConfs
}

// nolint: funlen
func (svc *Service) initialCapability() error {
	var err error

	// discover provider watch backend and file service.
	svc.Cap.DiscoverProvider = etcddiscover.NewProviderEtcd(&svc.conf.Etcd,
		etcddiscover.WithWatch(discover.ServiceNameBackend, discover.ServiceNameFile),
	)

	// initial local installer file group.
	svc.Cap.InstallerFileGroup, err = local.NewLocalDir(svc.conf.InstallerFileGroup.FullPath)
	if err != nil {
		return fmt.Errorf("failed to create installer file group: %w", err)
	}

	// initial local file cache.
	svc.Cap.FileCache, err = filecache.New(svc.ctx, svc.conf.FileCache.Dir, filecache.Options{
		ExpirationTime: time.Duration(svc.conf.FileCache.ExpirationHours) * time.Hour,
		GCInterval:     time.Duration(svc.conf.FileCache.GCIntervalHours) * time.Hour,
		RestoreOnStart: svc.conf.FileCache.RestoreOnStart,
	})
	if err != nil {
		return fmt.Errorf("failed to create file cache: %w", err)
	}

	// initial AES crypter with given key.
	svc.Cap.Crypter, err = crypter.NewAESCrypter([]byte(svc.conf.EncryptKey))
	if err != nil {
		return fmt.Errorf("failed to create AES crypter: %w", err)
	}

	// initial cmdb handler.
	svc.Cap.CmdbHandler, err = svc.newCMDBHandler()
	if err != nil {
		return fmt.Errorf("failed to create cmdb handler: %w", err)
	}

	// initial gse handler.
	svc.Cap.GSEHandler, err = svc.newGSEHandler()
	if err != nil {
		return fmt.Errorf("failed to create gse handler: %w", err)
	}

	// initial file handler.
	svc.Cap.FileHandler, err = svc.newFileHandler()
	if err != nil {
		return fmt.Errorf("failed to create file handler: %w", err)
	}

	// initial user manager handler.
	svc.Cap.UserManagerHandler, err = svc.newUserManagerHandler()
	if err != nil {
		return fmt.Errorf("failed to create user manager handler: %w", err)
	}

	// initial IAM v3 handler.
	svc.Cap.IAMV3Handler, err = svc.newIAMV3Handler()
	if err != nil {
		return fmt.Errorf("failed to create IAM v3 handler: %w", err)
	}

	// initial credit vault.
	svc.Cap.CreditVault, err = svc.newCreditVault()
	if err != nil {
		return fmt.Errorf("failed to create credit vault: %w", err)
	}

	// initial redis client.
	svc.Cap.RedisClient, err = svc.newRedisClient()
	if err != nil {
		return fmt.Errorf("failed to create redis client: %w", err)
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

	// register global setting.
	if err = svc.registerGlobalSetting(); err != nil {
		return fmt.Errorf("failed to register global setting: %w", err)
	}

	// initial locker factory.
	svc.Cap.LockerFactory = redsync.New(svc.Cap.RedisClient)

	// initial manager.
	if err = svc.initialManager(); err != nil {
		return fmt.Errorf("failed to initial manager: %w", err)
	}

	// initial period task.
	svc.Cap.PeriodicTask = periodictask.NewPeriodicTask(periodictask.Config{
		Locker:      svc.Cap.LockerFactory,
		StgWorkflow: svc.Cap.StorageWorkflow,
	})

	// initial IAM callback handler.
	svc.Cap.AuthProviderHandler = svc.newAuthProviderHandler()

	// initial authorizer.
	svc.Cap.Authorizer = svc.newAuthorizer()

	return nil
}

func (svc *Service) newCMDBHandler() (cmdb.IHandler, error) {
	apiGWAPPConfig := newAPIGWAppConfig(&svc.conf.CMDB.APIGatewayClient)
	apiGwClientCapability, err := newAPIGwClientCapability(clientNameCMDB, &svc.conf.CMDB.APIGatewayClient)
	if err != nil {
		return nil, fmt.Errorf("failed to new apigw client for cmdb: %w", err)
	}

	cmdbHandler, err := cmdb.New(
		apiGwClientCapability,
		&cmdb.Config{
			SupplierAccount: svc.conf.CMDB.SupplierAccount,
			VirtualUser:     access.GetVirtualUser(),
			APIGWAppConfig:  apiGWAPPConfig,
		},
	)
	if err != nil {
		return nil, err
	}

	return cmdbHandler, nil
}

func (svc *Service) newGSEHandler() (gse.IHandler, error) {
	apiGWUserConfig := newAPIGWUserConfig(&svc.conf.GSE.APIGatewayClient)
	apiGwClientCapability, err := newAPIGwClientCapability(clientNameGSE, &svc.conf.GSE.APIGatewayClient)
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

func (svc *Service) newFileHandler() (file.IHandler, error) {
	httpClient, err := restclient.NewHTTPClient(&ssl.TLSConfig{InsecureSkipVerify: true})
	if err != nil {
		return nil, fmt.Errorf("failed to create http client for file service: %w", err)
	}

	traceSvc, err := tracing.G().NewService(tracing.ServiceConfig{
		ServiceName:     svc.conf.File.TraceServiceName,
		ServiceCategory: tracing.ServiceCategoryHTTP,
		SampleRate:      svc.conf.File.TraceSampleRate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to new trace service: %w", err)
	}

	clientCap := &restclient.Capability{
		Name:       clientNameFile,
		TraceSvc:   traceSvc,
		HTTPClient: httpClient,
		Discover: restdiscovery.NewServiceDiscovery(
			svc.Cap.DiscoverProvider,
			discover.ServiceNameFile,
			discover.EndpointNameFileBasic,
		),
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
		MetricOpts:           restclient.MetricOption{},
	}

	return file.New(clientCap, &file.Config{
		RestJwtSecret:          svc.conf.File.JWTClientConfig.SymmetricKey,
		RestJwtTokenExpiration: time.Duration(svc.conf.File.JWTClientConfig.TokenExpirationHour) * time.Hour,
	})
}

func (svc *Service) newUserManagerHandler() (usermanager.IHandler, error) {
	apiGWUserConfig := newAPIGWUserConfig(&svc.conf.UserManager.APIGatewayClient)
	apiGwClientCapability, err := newAPIGwClientCapability(clientNameUserManager, &svc.conf.UserManager.APIGatewayClient)
	if err != nil {
		return nil, fmt.Errorf("failed to new apigw client for gse: %w", err)
	}

	var (
		usermgrHandler usermanager.IHandler
	)

	if tenant.GetMode() == tenant.ModeSingle {
		usermgrHandler, err = usermanager.NewHandlerSingle(
			apiGwClientCapability,
			&usermanager.Config{
				APIGWUserConfig: apiGWUserConfig,
			},
		)
		if err != nil {
			return nil, err
		}
	} else {
		usermgrHandler, err = usermanager.NewHandlerMultiTenant(
			apiGwClientCapability,
			&usermanager.Config{
				APIGWUserConfig: apiGWUserConfig,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	if err := tenant.SetTenantIDProvider(usermgrHandler); err != nil {
		return nil, fmt.Errorf("failed to set tenant id provider: %w", err)
	}

	tenant.SetTenantUserResolver(usermgrHandler)

	return usermgrHandler, nil
}

func (svc *Service) newAuthorizer() auth.IAuthorizer {
	if !svc.conf.IAMV3.Enable {
		return auth.NewNoOpAuthorizer()
	}

	return auth.NewIAMV3Authorizer(
		svc.conf.IAMV3.SystemID,
		svc.Cap.IAMV3Handler,
		svc.Cap.AuthProviderHandler, // IAttributeEnricher
		svc.Cap.AuthProviderHandler, // IResolver
	)
}

// newAuthProviderHandler creates a unified IAM callback handler with all providers registered.
func (svc *Service) newAuthProviderHandler() *authProvider.Handler {
	handler := authProvider.NewHandler()

	// Register NetworkArea provider
	networkAreaProvider := authProvider.NewNetworkAreaProvider(svc.Cap.StorageTopo)
	handler.RegisterProvider(authProvider.ResourceTypeNetworkArea, networkAreaProvider)

	// Register NetworkUnit provider
	networkUnitProvider := authProvider.NewNetworkUnitProvider(svc.Cap.StorageTopo)
	handler.RegisterProvider(authProvider.ResourceTypeNetworkUnit, networkUnitProvider)

	// Register PackageType provider
	packageTypeProvider := authProvider.NewPackageTypeProvider()
	handler.RegisterProvider(authProvider.ResourceTypePackageType, packageTypeProvider)

	// Register Package provider
	packageProvider := authProvider.NewPackageProvider(svc.Cap.StorageRelease)
	handler.RegisterProvider(authProvider.ResourceTypePackage, packageProvider)

	return handler
}

// newIAMV3Handler creates a new IAM v3 handler.
func (svc *Service) newIAMV3Handler() (iamv3.IHandler, error) {
	// Return no-op handler if IAM v3 is disabled
	if !svc.conf.IAMV3.Enable {
		return iamv3.NewNoOpHandler(), nil
	}

	apiGwAppConfig := newAPIGWAppConfig(&svc.conf.IAMV3.APIGatewayClient)
	apiGwUserConfig := apigwclient.UserConfig{
		AppConfig:   apiGwAppConfig,
		AuthMode:    apigwclient.AuthMode(svc.conf.IAMV3.AuthMode),
		BKUsername:  svc.conf.IAMV3.User,
		AccessToken: svc.conf.IAMV3.AccessToken,
	}

	apiGwClientCapability, err := newAPIGwClientCapability(
		clientNameIAM,
		&svc.conf.IAMV3.APIGatewayClient,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to new apigw client for IAM v3: %w", err)
	}

	iamHandler, err := iamv3.New(apiGwClientCapability, &iamv3.Config{
		APIGWUserConfig: apiGwUserConfig,
		SystemID:        svc.conf.IAMV3.SystemID,
		CallbackPath:    svc.conf.IAMV3.CallbackPath,
	})
	if err != nil {
		return nil, err
	}

	return iamHandler, nil
}

func (svc *Service) newCreditVault() (creditvault.ICreditVault, error) {
	if !svc.conf.CreditVault.HostCreditVault.Enable {
		return creditvault.New(creditvault.WithHostPasswordVault(&creditvault.DisabledHostPasswordVault{})), nil
	}

	switch svc.conf.CreditVault.HostCreditVault.Type {
	case "iegtjj":
		iegtjjHandler, err := newIEGTJJHandler(svc.conf.CreditVault.HostCreditVault.IEGTJJ)
		if err != nil {
			return nil, fmt.Errorf("failed to create IEG TJJ vault: %w", err)
		}

		return creditvault.New(creditvault.WithHostPasswordVault(iegtjjHandler)), nil

	default:
		return nil, fmt.Errorf("unsupported host password vault type: %s", svc.conf.CreditVault.HostCreditVault.Type)
	}
}

func (svc *Service) newRedisClient() (redis.UniversalClient, error) {
	var tlsConfig *tls.Config
	if svc.conf.Redis.TLS.CAFile != "" && svc.conf.Redis.TLS.CertFile != "" && svc.conf.Redis.TLS.KeyFile != "" {
		sslConf := &ssl.TLSConfig{
			CAFile:   svc.conf.Redis.TLS.CAFile,
			CertFile: svc.conf.Redis.TLS.CertFile,
			KeyFile:  svc.conf.Redis.TLS.KeyFile,
			Password: svc.conf.Redis.TLS.Password,
		}

		var err error
		tlsConfig, err = sslConf.NewClientTLSConf()
		if err != nil {
			return nil, fmt.Errorf("failed to create tls config: %w", err)
		}
	}

	logger.G.Sys().With("addrs", svc.conf.Redis.Addrs, "master-name", svc.conf.Redis.MasterName).
		Info("initializing redis universal client")

	client := redis.NewUniversalClient(&redis.UniversalOptions{
		IsClusterMode: svc.conf.Redis.Type == config.RedisTypeCluster,
		Addrs:         svc.conf.Redis.Addrs,
		MasterName:    svc.conf.Redis.MasterName,
		Password:      svc.conf.Redis.Password,
		DB:            svc.conf.Redis.DB,
		TLSConfig:     tlsConfig,
	})

	traceSvc, err := tracing.G().NewService(tracing.ServiceConfig{
		ServiceName:     svc.conf.Redis.TraceServiceName,
		ServiceCategory: tracing.ServiceCategoryCache,
		SampleRate:      svc.conf.Redis.TraceSampleRate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create tracing service: %w", err)
	}

	err = redisotel.InstrumentTracing(client,
		redisotel.WithTracerProvider(traceSvc.TracerProvider()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to instrument tracing for redis client: %w", err)
	}

	_, err = client.Ping(contextx.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to ping redis after creating redis client: %w", err)
	}

	return client, nil
}

func (svc *Service) newMongoClient() (*mongo.Client, error) {
	mongoSvc, err := tracing.G().NewService(tracing.ServiceConfig{
		ServiceName:     svc.conf.MongoDB.TraceServiceName,
		ServiceCategory: tracing.ServiceCategoryDB,
		SampleRate:      svc.conf.MongoDB.TraceSampleRate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create mongo service: %w", err)
	}

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
			ReplicaSet:     replicaSet,
			Hosts:          svc.conf.MongoDB.Hosts,
			TLSConfig:      tlsConfig,
			ReadPreference: readpref.Primary(),
			// All reads go to the primary (ReadPreference=primary), so w:1 writes are
			// immediately readable. The majority replication ack (~190ms/command) is
			// unnecessary for retried sync operations that are reconciled by the next cycle.
			WriteConcern:    writeconcern.W1(),
			Monitor:         otelmongo.NewMonitor(otelmongo.WithTracerProvider(mongoSvc.TracerProvider())),
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

	svc.Cap.StorageTopo, err = topoStg.NewStorage(
		svc.Cap.MongoClient,
		svc.conf.MongoDB.Database)
	if err != nil {
		return fmt.Errorf("failed to create topo storage: %w", err)
	}

	svc.Cap.StorageNode, err = nodeStg.NewStorage(
		svc.Cap.MongoClient,
		svc.conf.MongoDB.Database)
	if err != nil {
		return fmt.Errorf("failed to create node storage: %w", err)
	}

	svc.Cap.StoragePlugin, err = pluginStg.NewStorage(
		svc.Cap.MongoClient,
		svc.conf.MongoDB.Database)
	if err != nil {
		return fmt.Errorf("failed to create pluginStg storage: %w", err)
	}

	svc.Cap.StorageWorkflow, err = workflow.NewStorage(
		svc.Cap.MongoClient,
		svc.conf.MongoDB.Database)
	if err != nil {
		return fmt.Errorf("failed to create workflow storage: %w", err)
	}

	svc.Cap.StoragePackage, err = pkgStg.NewStorage(
		svc.Cap.MongoClient,
		svc.conf.MongoDB.Database)
	if err != nil {
		return fmt.Errorf("failed to create pkg storage: %w", err)
	}

	svc.Cap.StorageRelease, err = release.NewStorage(
		svc.Cap.MongoClient,
		svc.conf.MongoDB.Database)
	if err != nil {
		return fmt.Errorf("failed to create release storage: %w", err)
	}

	svc.Cap.StorageCredit, err = credit.NewStorage(
		svc.Cap.MongoClient,
		svc.conf.MongoDB.Database,
		svc.Cap.Crypter)
	if err != nil {
		return fmt.Errorf("failed to create credit storage: %w", err)
	}

	svc.Cap.StorageConfigPolicy, err = configpolicy.NewStorage(
		svc.Cap.MongoClient,
		svc.conf.MongoDB.Database)
	if err != nil {
		return fmt.Errorf("failed to create config policy storage: %w", err)
	}

	svc.Cap.StorageDeployPolicy, err = deploypolicy.NewStorage(
		svc.Cap.MongoClient,
		svc.conf.MongoDB.Database)
	if err != nil {
		return fmt.Errorf("failed to create deploy policy storage: %w", err)
	}

	svc.Cap.StorageGlobalSettings, err = globalsettingsStorage.NewStorage(
		svc.Cap.MongoClient,
		svc.conf.MongoDB.Database)
	if err != nil {
		return fmt.Errorf("failed to create global settings storage: %w", err)
	}

	svc.Cap.StorageTenant, err = tenantStg.NewStorage(
		svc.Cap.MongoClient,
		svc.conf.MongoDB.Database)
	if err != nil {
		return fmt.Errorf("failed to create tenant storage: %w", err)
	}

	svc.Cap.StorageCipher, err = cipherStg.NewStorage(
		svc.Cap.MongoClient,
		svc.conf.MongoDB.Database,
	)
	if err != nil {
		return fmt.Errorf("failed to create asymmetric encryption storage: %w", err)
	}

	return nil
}

func (svc *Service) registerGlobalSetting() error {
	storageGlobalSettings, err := globalsettingsStorage.NewStorage(
		svc.Cap.MongoClient,
		svc.conf.MongoDB.Database)
	if err != nil {
		return fmt.Errorf("failed to create global settings storage: %w", err)
	}

	if err := storageGlobalSettings.Start(svc.ctx); err != nil {
		return fmt.Errorf("failed to start the storage of globalsetting: %w", err)
	}

	// initial global settings.
	if err := globalsettings.Register(svc.ctx, storageGlobalSettings); err != nil {
		return fmt.Errorf("failed to initial global settings: %w", err)
	}

	return nil
}

func (svc *Service) initialManager() error {
	svc.Cap.ProxyMessager = relayhandler.NewServerMessager(relayhandler.ServerMessagerConfig{
		SlotID:        svc.conf.GSE.PluginSlotID,
		Token:         svc.conf.GSE.PluginSlotToken,
		AppCode:       svc.conf.GSE.AppCode,
		AppSecret:     svc.conf.GSE.AppSecret,
		GSEBaseURL:    svc.conf.GSE.Endpoints[0],
		SkipTLSVerify: svc.conf.GSE.TLS.InsecureSkipVerify,
		RedisClient:   svc.Cap.RedisClient,
	})

	traceSvc, err := tracing.G().NewService(tracing.ServiceConfig{
		ServiceName:     svc.conf.Workflow.TraceServiceName,
		ServiceCategory: tracing.ServiceCategoryAsyncBackend,
		SampleRate:      svc.conf.Workflow.TraceSampleRate,
	})
	if err != nil {
		return fmt.Errorf("failed to create tracing service: %w", err)
	}

	svc.Cap.Manager, err = manager.NewManager(manager.Config{
		CmdbHandler:         svc.Cap.CmdbHandler,
		GSEHandler:          svc.Cap.GSEHandler,
		FileHandler:         svc.Cap.FileHandler,
		UserManagerHandler:  svc.Cap.UserManagerHandler,
		Provider:            svc.Cap.DiscoverProvider,
		InstallerFileGroup:  svc.Cap.InstallerFileGroup,
		FileCache:           svc.Cap.FileCache,
		LockerFactory:       svc.Cap.LockerFactory,
		StorageTopo:         svc.Cap.StorageTopo,
		StorageRelease:      svc.Cap.StorageRelease,
		StorageNode:         svc.Cap.StorageNode,
		StorageWorkflow:     svc.Cap.StorageWorkflow,
		StoragePackage:      svc.Cap.StoragePackage,
		StoragePlugin:       svc.Cap.StoragePlugin,
		StorageHostCredit:   svc.Cap.StorageCredit,
		StorageConfigPolicy: svc.Cap.StorageConfigPolicy,
		StorageTenant:       svc.Cap.StorageTenant,
		StorageDeployPolicy: svc.Cap.StorageDeployPolicy,
		HostPasswordVault:   svc.Cap.CreditVault,
		ProxyMessager:       svc.Cap.ProxyMessager,
		Cache:               rediscache.NewRedisCache(svc.Cap.RedisClient, rediscache.DefaultTimeout),
		WorkflowConfig: manager.WorkflowConfig{
			WorkNodeNum:             svc.conf.Workflow.WorkerNum,
			GracefulShutdownTimeout: time.Duration(svc.conf.Workflow.GracefulShutdownTimeoutSeconds) * time.Second,
			Redis:                   svc.conf.Redis,
		},
		TraceService: traceSvc,
	})
	if err != nil {
		return fmt.Errorf("failed to create manager: %w", err)
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

	if err := svc.registerBasicServer(); err != nil {
		return fmt.Errorf("failed to register basic server: %w", err)
	}

	if err := svc.registerCallbackServer(); err != nil {
		return fmt.Errorf("failed to register callback server: %w", err)
	}

	if err := svc.registerProxyServer(); err != nil {
		return fmt.Errorf("failed to register proxy server: %w", err)
	}

	return nil
}

// nolint: unparam
func (svc *Service) registerInfoServer() error {
	server, err := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:             string(discover.EndpointNameBackendInfo),
			IP:               svc.conf.InfoServer.BindIP,
			IPV6:             svc.conf.InfoServer.BindIPV6,
			Port:             svc.conf.InfoServer.Port,
			TLSConfig:        svc.conf.InfoServer.TLSConfig,
			ShutdownTimeout:  time.Duration(svc.conf.InfoServer.GracefulShutdownTimeoutSec) * time.Second,
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
	svc.instance.Update(discover.EndpointNameBackendInfo, discover.Endpoint{
		IPV4: svc.conf.InfoServer.AdvertiseIPV4,
		IPV6: svc.conf.InfoServer.AdvertiseIPV6,
		Port: svc.conf.InfoServer.Port,
	})

	return nil
}

func newAuthIdentity(conf config.HTTPServer) (restserver.IAuthIdentity, error) {
	switch conf.AuthIdentity {
	case config.AuthIdentityNone:
		return restserver.NewNoneAuthIdentity(), nil

	case config.AuthIdentityAPIGW:
		publickeyPem, err := base64.StdEncoding.DecodeString(conf.JWTServerConfig.PublicKeyPem)
		if err != nil {
			return nil, fmt.Errorf("failed to decode publickey: %w", err)
		}

		return apigwserver.NewBKGWJWTAuthIdentity(publickeyPem), nil

	case config.AuthIdentityRestServer:
		return restserver.NewRestServerAuthIdentity(conf.JWTServerConfig.SymmetricKey), nil

	default:
		return nil, fmt.Errorf("no support this auth identity, auth-identity(%s)", conf.AuthIdentity)
	}
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
			Name:             string(discover.EndpointNameBackendAdmin),
			IP:               svc.conf.AdminServer.BindIP,
			IPV6:             svc.conf.AdminServer.BindIPV6,
			Port:             svc.conf.AdminServer.Port,
			TLSConfig:        svc.conf.AdminServer.TLSConfig,
			ShutdownTimeout:  time.Duration(svc.conf.AdminServer.GracefulShutdownTimeoutSec) * time.Second,
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
	svc.instance.Update(discover.EndpointNameBackendAdmin, discover.Endpoint{
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
		return fmt.Errorf("failed to new auth identity: %w", err)
	}

	server, err := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:             string(discover.EndpointNameBackendBasic),
			IP:               svc.conf.BasicServer.BindIP,
			IPV6:             svc.conf.BasicServer.BindIPV6,
			Port:             svc.conf.BasicServer.Port,
			TLSConfig:        svc.conf.BasicServer.TLSConfig,
			ShutdownTimeout:  time.Duration(svc.conf.BasicServer.GracefulShutdownTimeoutSec) * time.Second,
			RequestIDSetter:  apigwserver.NewBKAPIRequestIDSetter(),
			TraceServiceName: svc.conf.BasicServer.TraceServiceName,
			TraceSampleRate:  svc.conf.BasicServer.TraceSampleRate,
		},
		restserver.WithPing(),
		withAPIV3Basic(svc.Cap,
			restserver.MiddlewareAuth(authIdentity, restserver.WithSkipPathPrefixes("/api/v3/iam"))),
	)
	if err != nil {
		return fmt.Errorf("failed to register basic server: %w", err)
	}

	svc.servers = append(svc.servers, server)
	svc.instance.Update(discover.EndpointNameBackendBasic, discover.Endpoint{
		IPV4: svc.conf.BasicServer.AdvertiseIPV4,
		IPV6: svc.conf.BasicServer.AdvertiseIPV6,
		Port: svc.conf.BasicServer.Port,
	})

	return nil
}

// nolint: unparam
func (svc *Service) registerCallbackServer() error {
	server, err := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:             string(discover.EndpointNameBackendCallback),
			IP:               svc.conf.CallbackServer.BindIP,
			IPV6:             svc.conf.CallbackServer.BindIPV6,
			Port:             svc.conf.CallbackServer.Port,
			TLSConfig:        svc.conf.CallbackServer.TLSConfig,
			ShutdownTimeout:  time.Duration(svc.conf.CallbackServer.GracefulShutdownTimeoutSec) * time.Second,
			RequestIDSetter:  restserver.NewRequestIDSetter(),
			TraceServiceName: svc.conf.CallbackServer.TraceServiceName,
			TraceSampleRate:  svc.conf.CallbackServer.TraceSampleRate,
		},
		restserver.WithPing(),
		withAPIV3Callback(svc.Cap),
	)
	if err != nil {
		return fmt.Errorf("failed to register callback server: %w", err)
	}

	svc.servers = append(svc.servers, server)
	svc.instance.Update(discover.EndpointNameBackendCallback, discover.Endpoint{
		IPV4: svc.conf.CallbackServer.AdvertiseIPV4,
		IPV6: svc.conf.CallbackServer.AdvertiseIPV6,
		Port: svc.conf.CallbackServer.Port,
	})

	return nil
}

// nolint: unparam
func (svc *Service) registerProxyServer() error {
	server, err := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:             string(discover.EndpointNameBackendPorxy),
			IP:               svc.conf.ProxyServer.BindIP,
			IPV6:             svc.conf.ProxyServer.BindIPV6,
			Port:             svc.conf.ProxyServer.Port,
			TLSConfig:        svc.conf.ProxyServer.TLSConfig,
			ShutdownTimeout:  time.Duration(svc.conf.ProxyServer.GracefulShutdownTimeoutSec) * time.Second,
			TraceServiceName: svc.conf.ProxyServer.TraceServiceName,
			TraceSampleRate:  svc.conf.ProxyServer.TraceSampleRate,
			RequestIDSetter:  restserver.NewRequestIDSetter(),
		},
		restserver.WithPing(),
		withAPIV3Proxy(svc.Cap),
	)
	if err != nil {
		return fmt.Errorf("failed to register proxy server: %w", err)
	}

	svc.servers = append(svc.servers, server)
	svc.instance.Update(discover.EndpointNameBackendPorxy, discover.Endpoint{
		IPV4: svc.conf.ProxyServer.AdvertiseIPV4,
		IPV6: svc.conf.ProxyServer.AdvertiseIPV6,
		Port: svc.conf.ProxyServer.Port,
	})

	return nil
}

// withHealthz load healthz.
func withHealthz(capability *options.Capability, middleware ...gin.HandlerFunc) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		healthz.Load(rg, capability, middleware...)
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
		backendapiv3.LoadBasicAPIs(rg, capability, middleware...)
	}
}

// withAPIV3Callback load api v3 callback.
func withAPIV3Callback(capability *options.Capability, middleware ...gin.HandlerFunc) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		backendapiv3.LoadCallbackAPIs(rg, capability, middleware...)
	}
}

// withAPIV3Proxy load api v3 basic.
func withAPIV3Proxy(capability *options.Capability, middleware ...gin.HandlerFunc) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		backendapiv3.LoadProxyAPIs(rg, capability, middleware...)
	}
}

// withAdmin load admin.
func withAdmin(capability *options.Capability, middleware ...gin.HandlerFunc) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		admin.Load(rg, capability, middleware...)
	}
}

func newIEGTJJHandler(conf config.IEGTJJ) (iegtjj.IHandler, error) {
	// apiGwClientConfig := newAPIGwClientConfig(&conf.APIGatewayClient)
	// TODO: 等待 iegtjj 迁移到 apigw, 将此处替换为 apigwclient.UserConfig
	apiGwClientCapability, err := newAPIGwClientCapability(clientNameIEGTJJ, &conf.APIGatewayClient)
	if err != nil {
		return nil, err
	}

	iegtjjHandler, err := iegtjj.New(apiGwClientCapability, &iegtjj.Config{
		Key:       conf.Key,
		SecretKey: conf.SecretKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to new iegtjj handler: %w", err)
	}

	return iegtjjHandler, nil
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

// newAPIGWAppConfig creates a new api-gateway client config.
func newAPIGWAppConfig(conf *config.APIGatewayClient) apigwclient.AppConfig {
	return apigwclient.NewAppConfig(conf.Endpoints, conf.AppCode, conf.AppSecret)
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

// Start starts the backend service.
func (svc *Service) Start() error {
	logger.G.Sys().Info("try to start backend service")

	runtime.GOMAXPROCS(runtime.NumCPU())
	if err := svc.Cap.Start(svc.ctx); err != nil {
		return err
	}

	// start servers
	gp := gopool.NewPool()
	for idx := range svc.servers {
		server := svc.servers[idx]

		// http server start will block until http server stop, so we need to run it in a goroutine.
		fn := func() error {
			if err := server.Start(); err != nil {
				return err
			}

			return nil
		}
		gp.Go(fn)
	}

	// after all servers brings up, register the instance into discover provider.
	if err := svc.Cap.DiscoverProvider.Register(discover.ServiceNameBackend, svc.instance); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to register backend server")
		return err
	}

	// after all servers brings up, initial default cipher.
	if err := svc.initialCipher(); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to initial default cipher")
		return err
	}

	// wait until all servers stopped or backend error.
	if err := gp.Wait(); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to start servers")
		return err
	}

	logger.G.Sys().Info("backend service started")

	return nil
}

// GracefulShutdown gracefully shuts down the backend service.
func (svc *Service) GracefulShutdown() error {
	logger.G.Sys().Info("try to gracefully shutdown backend service")

	if svc.ctx == nil || svc.cancelFunc == nil {
		return errors.New("service is not running")
	}

	defer svc.cancelFunc()

	serverErr := svc.shutdownRestServers(context.Background())
	if serverErr != nil {
		logger.G.Sys().WithErr(serverErr).Error("failed to gracefully shutdown rest servers")
	}

	capErr := svc.Cap.GracefulShutdown()
	if capErr != nil {
		logger.G.Sys().WithErr(capErr).Error("failed to gracefully shutdown capability")
	}

	if err := errors.Join(serverErr, capErr); err != nil {
		return err
	}

	logger.G.Sys().Info("backend service gracefully shutdown")

	return nil
}

func (svc *Service) shutdownRestServers(ctx context.Context) error {
	errCh := make(chan error, len(svc.servers))
	var wg sync.WaitGroup
	for _, server := range svc.servers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := server.Shutdown(ctx); err != nil {
				errCh <- fmt.Errorf("failed to shutdown rest server %s: %w", server.Name(), err)
			}
		}()
	}
	wg.Wait()
	close(errCh)

	errs := make([]error, 0, len(svc.servers))
	for err := range errCh {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
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

func (svc *Service) initialCipher() error {
	tenantIDs := make([]string, 0)
	switch tenant.GetMode() {
	case tenant.ModeSingle:
		tenantIDs = append(tenantIDs, tenant.SingleModeTenantID)
	case tenant.ModeMultiple:
		allTenant, err := svc.Cap.UserManagerHandler.ListALLTenants(svc.ctx)
		if err != nil {
			logger.G.Sys().WithErr(err).Error("failed to get rsa public key, failed to list all tenants")
			return err
		}

		tenantIDs = conv.SliceToSlice(allTenant, func(tenantInfo *types.Tenant) string {
			return tenantInfo.ID
		})
	default:
		return fmt.Errorf("unsupported tenant mode: %s", tenant.GetMode())
	}

	for _, tenantID := range tenantIDs {
		nCtx := contextx.From(svc.ctx, contextx.WithTenantID(tenantID))
		if err := svc.Cap.StorageCipher.EnsureDefaultCipher(nCtx); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to ensure rsa cipher")
			return err
		}
	}

	return nil
}
