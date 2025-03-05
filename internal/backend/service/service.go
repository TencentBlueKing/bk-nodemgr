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
	"runtime"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/admin"
	apiv3 "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/basic"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/callback"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/healthz"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/operation"
	operinstdataStorage "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/operinstdata"
	topoStorage "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/trigengine"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/blog"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/redsync"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/ssl"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/crypter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
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

// Service defines a server that provides backend services.
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
	servers []*rest.Server

	// Note: Cap is initialized in the Start() and could not be used in other package.
	// Cap is the capability of the service.
	Cap *options.Capability
}

const (
	// RouterNameHTTPServer defines the name of http server router.
	RouterNameHTTPServer = "http-server"

	// RouterNameAdminServer defines the name of admin server router.
	RouterNameAdminServer = "admin-server"
)

// NewService creates a new backend service.
func NewService(conf *config.BackendService) (*Service, error) {
	svc := &Service{
		conf: conf,
		Cap: &options.Capability{
			Logger: blog.GlobalLogger{},
		},
	}

	svc.ctx, svc.cancelFunc = context.WithCancel(context.Background())

	var err error

	svc.Cap.Crypter, err = crypter.NewAESCrypter([]byte(conf.EncryptKey))
	if err != nil {
		return nil, err
	}

	svc.Cap.CmdbHandler, err = newCMDBHandler(conf.CMDB)
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

	svc.Cap.TopoStorage, err = topoStorage.NewStorage(mongoClient, conf.MongoDB.Database, svc.Cap.Logger)
	if err != nil {
		return nil, err
	}

	svc.Cap.TrigEngineStorage, err = trigengine.NewStorage(mongoClient, conf.MongoDB.Database, svc.Cap.Logger)
	if err != nil {
		return nil, err
	}

	svc.Cap.OperInstStorage, err = operinstdataStorage.NewStorage(mongoClient, conf.MongoDB.Database, svc.Cap.Logger)
	if err != nil {
		return nil, err
	}

	svc.Cap.OperStorage, err = operation.NewStorage(mongoClient, conf.MongoDB.Database, svc.Cap.Logger)
	if err != nil {
		return nil, err
	}

	svc.Cap.Manager, err = manager.NewManager(manager.Config{
		CmdbHandler:     svc.Cap.CmdbHandler,
		TopoStorage:     svc.Cap.TopoStorage,
		LockerFactory:   svc.Cap.LockerFactory,
		OperStorage:     svc.Cap.OperStorage,
		OperInstStorage: svc.Cap.OperInstStorage,
		WorkflowConfig: manager.WorkflowConfig{
			WorkNodeNum: conf.Workflow.WorkerNum,
			Redis: manager.RedisConfig{
				Addr:     fmt.Sprintf("%s:%d", conf.Redis.Host, conf.Redis.Port),
				Password: conf.Redis.Password,
				DB:       conf.Redis.DB,
			},
		},
		Crypter: svc.Cap.Crypter,
	}, blog.GlobalLogger{})
	if err != nil {
		return nil, err
	}

	svc.registerRestServer(conf)

	return svc, nil
}

func (svc *Service) registerRestServer(conf *config.BackendService) {
	httpServer := rest.NewServer(svc.ctx, RouterNameHTTPServer, conf.HTTPServer.BindIP, conf.HTTPServer.Port,
		loggerWriterAdaptor{},
		nil,
		rest.WithPing(),
		withHealthz(svc.Cap),
		withMetrics(svc.Cap),
		withAPIV3(svc.Cap),
		withBasic(svc.Cap),
	)
	svc.servers = append(svc.servers, httpServer)

	adminServer := rest.NewServer(svc.ctx, RouterNameAdminServer, conf.AdminServer.BindIP, conf.AdminServer.Port,
		loggerWriterAdaptor{},
		nil,
		rest.WithPing(),
		withHealthz(svc.Cap),
		withMetrics(svc.Cap),
		withAdmin(svc.Cap),
	)

	svc.servers = append(svc.servers, adminServer)

	callbackServer := rest.NewServer(svc.ctx, "callback", conf.CallbackServer.BindIP, conf.CallbackServer.Port,
		loggerWriterAdaptor{},
		nil,
		rest.WithPing(),
		withCallback(svc.Cap),
	)

	svc.servers = append(svc.servers, callbackServer)
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

// withApiV3 load api v3.
func withAPIV3(capability *options.Capability) rest.OptionFunc {
	return func(rg *gin.RouterGroup) {
		apiv3.Load(rg, capability)
	}
}

// withBasic load basic.
func withBasic(capability *options.Capability) rest.OptionFunc {
	return func(rg *gin.RouterGroup) {
		basic.Load(rg, capability)
	}
}

// withAdmin load admin.
func withAdmin(capability *options.Capability) rest.OptionFunc {
	return func(rg *gin.RouterGroup) {
		admin.Load(rg, capability)
	}
}

// withCallback load callback.
func withCallback(capability *options.Capability) rest.OptionFunc {
	return func(rg *gin.RouterGroup) {
		callback.Load(rg, capability)
	}
}

// newCMDBHandler.
func newCMDBHandler(conf config.CMDB) (cmdb.Handler, error) {
	apiGwHeaderSetter := newAPIGwHeaderSetter(&conf.APIGateway)
	apiGwClientCapability, err := newAPIGwClientCapability(&conf.APIGateway)
	if err != nil {
		return nil, err
	}

	apiGwClientCapability.Name = "cmdb"
	cmdbHandler, err := cmdb.New(apiGwClientCapability, &cmdb.Config{
		SupplierAccount: conf.SupplierAccount,
		HeaderSetter:    apiGwHeaderSetter,
	})
	if err != nil {
		return nil, err
	}

	return cmdbHandler, nil
}

// newAPIGwClientCapability creates a new api-gateway client capability.
func newAPIGwClientCapability(conf *config.APIGateway) (*client.Capability, error) {
	httpClient, err := client.NewClient(&ssl.TLSConfig{
		InsecureSkipVerify: conf.TLS.InsecureSkipVerify,
		CertFile:           conf.TLS.CertFile,
		KeyFile:            conf.TLS.KeyFile,
		CAFile:             conf.TLS.CAFile,
		Password:           conf.TLS.Password,
	})
	if err != nil {
		return nil, err
	}

	clientCap := &client.Capability{
		Client:               httpClient,
		Discover:             discovery.NewDiscovery(DiscoveryNameApigw, conf.Endpoints),
		ToleranceLatencyTime: client.ToleranceLatencyTimeDefault,
		MetricOpts:           client.MetricOption{},
		Logger:               blog.GlobalLogger{},
	}

	return clientCap, nil
}

// newAPIGwHeaderSetter creates a new api-gateway header setter.
func newAPIGwHeaderSetter(conf *config.APIGateway) apigw.HeaderSetter {
	return &apigw.Config{
		Endpoints:   conf.Endpoints,
		AppCode:     conf.AppCode,
		AppSecret:   conf.AppSecret,
		User:        conf.User,
		AuthMode:    apigw.AuthMode(conf.AuthMode),
		BkTicket:    conf.BkTicket,
		BkToken:     conf.BkToken,
		AccessToken: conf.AccessToken,
	}
}

// Start starts the backend service.
func (svc *Service) Start() error {
	runtime.GOMAXPROCS(runtime.NumCPU())

	logConfig := blog.NewLogConfig()
	logConfig.LogDir = svc.conf.Log.Dir
	logConfig.LogMaxSizeMB = svc.conf.Log.MaxSizeMB
	logConfig.LogMaxNum = svc.conf.Log.MaxNum
	logConfig.Level = svc.conf.Log.Level
	logConfig.ToStdErr = svc.conf.Log.ToStdErr
	logConfig.AlsoToStdErr = svc.conf.Log.AlsoToStdErr
	blog.InitLogs(logConfig)

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

// GracefulShutdown ...
func (svc *Service) GracefulShutdown() error {
	if svc.ctx == nil || svc.cancelFunc == nil {
		return errors.New("service is not running")
	}

	defer svc.cancelFunc()

	err := svc.Cap.GracefulShutdown()
	if err != nil {
		blog.Errorf("failed to shutdown capability, err: %v", err)
		return err
	}

	blog.CloseLogs()

	return nil
}
