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
	"runtime"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/periodictask"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/admin"
	backendapiv3 "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/callback"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/healthz"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/proxy"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/configpolicy"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	globalsettingsStorage "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/globalsettings"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/access"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover/etcddiscover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"
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

// NewService creates a new backend service.
// nolint: funlen,gocognit,gocyclo,cyclop,maintidx
// NOCC: golint/fnsize(func design is not suitable for splitting).
func NewService(conf *config.BackendService) (*Service, error) {
	svc := &Service{
		conf:     conf,
		Cap:      &options.Capability{},
		instance: discover.NewInstance(string(discover.ServiceNameBackend), nil),
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
			Generation:    types.Generation(svc.conf.GSEDeployConfs[idx].Generation),
			OsType:        criteria.OSType(svc.conf.GSEDeployConfs[idx].OsType),
			BaseDeployDir: svc.conf.GSEDeployConfs[idx].BaseDeployDir,
			BaseWorkDir:   svc.conf.GSEDeployConfs[idx].BaseWorkDir,
		}
		if err := deployconstant.SetDeployConf(deployConf); err != nil {
			return fmt.Errorf("failed to set deploy conf: %w", err)
		}

		nodeDeployConf := deployconstant.NodeDeployConf{
			DeployConf:         deployConf,
			LogDir:             svc.conf.GSEDeployConfs[idx].Custom.LogDir,
			HostIDPath:         svc.conf.GSEDeployConfs[idx].Custom.HostIDPath,
			AgentDataIPCPath:   svc.conf.GSEDeployConfs[idx].Custom.AgentDataIPCPath,
			AgentPluginIPCPath: svc.conf.GSEDeployConfs[idx].Custom.AgentPluginIPCPath,
			EnvironDir:         svc.conf.GSEDeployConfs[idx].Custom.EnvironDir,
		}
		if err := deployconstant.SetNodeDeployConf(nodeDeployConf); err != nil {
			return fmt.Errorf("failed to set node deploy conf: %w", err)
		}

		pluginDeployConf := deployconstant.PluginDeployConf{
			DeployConf:         deployConf,
			LogDir:             svc.conf.GSEDeployConfs[idx].PluginCustom.LogDir,
			DataDir:            svc.conf.GSEDeployConfs[idx].PluginCustom.DataDir,
			RunDir:             svc.conf.GSEDeployConfs[idx].PluginCustom.RunDir,
			HostIDPath:         svc.conf.GSEDeployConfs[idx].PluginCustom.HostIDPath,
			AgentDataIPCPath:   svc.conf.GSEDeployConfs[idx].PluginCustom.AgentDataIPCPath,
			AgentPluginIPCPath: svc.conf.GSEDeployConfs[idx].PluginCustom.AgentPluginIPCPath,
		}
		if err := deployconstant.SetPluginDeployConf(pluginDeployConf); err != nil {
			return fmt.Errorf("failed to set pluginStg deploy conf: %w", err)
		}
	}

	// initial idenity map.
	publickeyPem, err := base64.StdEncoding.DecodeString(svc.conf.APIGateWayServer.PublickeyPem)
	if err != nil {
		return fmt.Errorf("failed to decode publickey: %w", err)
	}
	svc.authIdentityMap = map[config.AuthIdentity]restserver.IAuthIdentity{
		config.AuthIdentityNone:       restserver.NewNodeAuthIdentity(),
		config.AuthIdentityAPIGW:      apigwserver.NewBKGWJWTAuthIdentity(publickeyPem),
		config.AuthIdentityRestServer: restserver.NewRestServerAuthIdentity(),
	}

	return nil
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

	// initial locker factory.
	svc.Cap.LockerFactory = redsync.New(svc.Cap.RedisClient)

	// initial manager.
	if err = svc.initialManager(); err != nil {
		return fmt.Errorf("failed to initial manager: %w", err)
	}

	// initial period task.
	svc.Cap.PeriodicTask = periodictask.NewPeriodicTask(periodictask.Config{
		Locker:           svc.Cap.LockerFactory,
		StgGlobalSetting: svc.Cap.StorageGlobalSettings,
		StgWorkflow:      svc.Cap.StorageWorkflow,
	})

	return nil
}

func (svc *Service) newCMDBHandler() (cmdb.IHandler, error) {
	apiGWAPPConfig := newAPIGWAppConfig(&svc.conf.CMDB.APIGatewayClient)
	apiGwClientCapability, err := newAPIGwClientCapability(&svc.conf.CMDB.APIGatewayClient)
	if err != nil {
		return nil, fmt.Errorf("failed to new apigw client for cmdb: %w", err)
	}

	apiGwClientCapability.Name = "cmdb"
	cmdbHandler, err := cmdb.New(
		apiGwClientCapability,
		&cmdb.Config{
			SupplierAccount: svc.conf.CMDB.SupplierAccount,
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

func (svc *Service) newFileHandler() (file.IHandler, error) {
	httpClient, err := restclient.NewHTTPClient(&ssl.TLSConfig{InsecureSkipVerify: true})
	if err != nil {
		return nil, fmt.Errorf("failed to create http client for file service: %w", err)
	}

	clientCap := &restclient.Capability{
		Name:       "file",
		HTTPClient: httpClient,
		Discover: restdiscovery.NewServiceDiscovery(
			svc.Cap.DiscoverProvider,
			discover.ServiceNameFile,
			discover.EndpointNameFileBasic,
		),
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
		MetricOpts:           restclient.MetricOption{},
	}

	return file.New(clientCap, &file.Config{})
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

func (svc *Service) newRedisClient() (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", svc.conf.Redis.Host, svc.conf.Redis.Port),
		Password: svc.conf.Redis.Password,
		DB:       svc.conf.Redis.DB,
	})

	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to ping redis after creating redis client: %w", err)
	}

	return redisClient, nil
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

	svc.Cap.StorageGlobalSettings, err = globalsettingsStorage.NewStorage(
		svc.Cap.MongoClient,
		svc.conf.MongoDB.Database)
	if err != nil {
		return fmt.Errorf("failed to create global settings storage: %w", err)
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

	var err error
	svc.Cap.Manager, err = manager.NewManager(manager.Config{
		CmdbHandler:         svc.Cap.CmdbHandler,
		GSEHandler:          svc.Cap.GSEHandler,
		Provider:            svc.Cap.DiscoverProvider,
		InstallerFileGroup:  svc.Cap.InstallerFileGroup,
		LockerFactory:       svc.Cap.LockerFactory,
		StorageTopo:         svc.Cap.StorageTopo,
		StorageRelease:      svc.Cap.StorageRelease,
		StorageNode:         svc.Cap.StorageNode,
		StorageWorkflow:     svc.Cap.StorageWorkflow,
		StoragePlugin:       svc.Cap.StoragePlugin,
		StorageHostCredit:   svc.Cap.StorageCredit,
		StorageConfigPolicy: svc.Cap.StorageConfigPolicy,
		HostPasswordVault:   svc.Cap.CreditVault,
		FileHandler:         svc.Cap.FileHandler,
		ProxyMessager:       svc.Cap.ProxyMessager,
		Cache:               rediscache.NewRedisCache(svc.Cap.RedisClient, rediscache.DefaultTimeout),
		WorkflowConfig: manager.WorkflowConfig{
			WorkNodeNum: svc.conf.Workflow.WorkerNum,
			Redis: manager.RedisConfig{
				Addr:     fmt.Sprintf("%s:%d", svc.conf.Redis.Host, svc.conf.Redis.Port),
				Password: svc.conf.Redis.Password,
				DB:       svc.conf.Redis.DB,
			},
		},
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
	server := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:            string(discover.EndpointNameBackendInfo),
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
	svc.instance.Update(discover.EndpointNameBackendInfo, discover.Endpoint{
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
			Name:            string(discover.EndpointNameBackendAdmin),
			IP:              svc.conf.AdminServer.BindIP,
			Port:            svc.conf.AdminServer.Port,
			RequestIDSetter: restserver.NewRequestIDSetter(),
			TenantIDSetter:  restserver.NewTenantIDSetter(),
		},
		restserver.WithPing(),
		withAdmin(svc.Cap, authIdentity),
	)

	svc.servers = append(svc.servers, server)
	svc.instance.Update(discover.EndpointNameBackendAdmin, discover.Endpoint{
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
			Name:            string(discover.EndpointNameBackendBasic),
			IP:              svc.conf.BasicServer.BindIP,
			Port:            svc.conf.BasicServer.Port,
			RequestIDSetter: apigwserver.NewBKAPIRequestIDSetter(),
			TenantIDSetter:  restserver.NewTenantIDSetter(),
		},
		restserver.WithPing(),
		withAPIV3(svc.Cap, authIdentity),
	)

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
	server := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:            string(discover.EndpointNameBackendCallback),
			IP:              svc.conf.CallbackServer.BindIP,
			Port:            svc.conf.CallbackServer.Port,
			RequestIDSetter: restserver.NewRequestIDSetter(),
			TenantIDSetter:  restserver.NewTenantIDSetter(),
		},
		restserver.WithPing(),
		withCallback(svc.Cap),
	)

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
	server := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:            string(discover.EndpointNameBackendPorxy),
			IP:              svc.conf.ProxyServer.BindIP,
			Port:            svc.conf.ProxyServer.Port,
			RequestIDSetter: restserver.NewRequestIDSetter(),
			TenantIDSetter:  restserver.NewTenantIDSetter(),
		},
		restserver.WithPing(),
		withProxy(svc.Cap),
	)

	svc.servers = append(svc.servers, server)
	svc.instance.Update(discover.EndpointNameBackendPorxy, discover.Endpoint{
		IPV4: svc.conf.ProxyServer.AdvertiseIPV4,
		IPV6: svc.conf.ProxyServer.AdvertiseIPV6,
		Port: svc.conf.ProxyServer.Port,
	})

	return nil
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
func withAPIV3(capability *options.Capability, authIdentity restserver.IAuthIdentity) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		backendapiv3.Load(rg, capability, authIdentity)
	}
}

// withAdmin load admin.
func withAdmin(capability *options.Capability, authIdentity restserver.IAuthIdentity) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		admin.Load(rg, capability, authIdentity)
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

func newIEGTJJHandler(conf config.IEGTJJ) (iegtjj.IHandler, error) {
	// apiGwClientConfig := newAPIGwClientConfig(&conf.APIGatewayClient)
	// TODO: 等待 iegtjj 迁移到 apigw, 将此处替换为 apigwclient.UserConfig
	apiGwClientCapability, err := newAPIGwClientCapability(&conf.APIGatewayClient)
	if err != nil {
		return nil, err
	}

	apiGwClientCapability.Name = "iegtjj"
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

			logger.G.Sys().With("name", server.Name(), "ip", server.IP(), "port", server.Port()).Info("started HTTP Server")

			return nil
		}
		gp.Go(fn)
	}

	// after all servers brings up, register the instance into discover provider.
	if err := svc.Cap.DiscoverProvider.Register(discover.ServiceNameBackend, svc.instance); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to register backend server")
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

	err := svc.Cap.GracefulShutdown()
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to gracefully shutdown capability")

		return err
	}

	logger.G.Sys().Info("backend service gracefully shutdown")

	return nil
}
