/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package service provides application service.
package service

import (
	"context"
	"errors"
	"fmt"
	"runtime"

	"github.com/TencentBlueKing/bk-nodemgr/internal/application/frontsetting"
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/router/admin"
	applicationapiv3 "github.com/TencentBlueKing/bk-nodemgr/internal/application/router/api-v3"
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/router/healthz"
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/router/web"
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/storage/cptemplate"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover/etcddiscover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	apigwserver "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backend"
	bksaasbklogin "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/bksaas/bklogin"
	bksaasheader "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/bksaas/header"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tracing"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	clientNameBackend = "backend"
	clientNameBKLogin = "bklogin"
	clientNameFile    = "file"
)

// Service defines a apigwserver that provides application services.
// It provides a website for user to operate with nodeman.
type Service struct {
	// conf holds the configuration for the service.
	conf *config.ApplicationService

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

	// bkloginHandler is the handler of bklogin.
	bkloginHandler bksaasbklogin.IHandler
}

// NewService creates a new application service.
func NewService(conf *config.ApplicationService) (*Service, error) {
	svc := &Service{
		conf:     conf,
		Cap:      &options.Capability{},
		instance: discover.NewInstance(string(discover.ServiceNameApplication), nil),
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

// nolint: unparam
func (svc *Service) initialStaticsConfigs() error {
	// initial idenity map.
	svc.authIdentityValidMap = map[config.AuthIdentity]struct{}{
		config.AuthIdentityNone:    {},
		config.AuthIdentityBKLogin: {},
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

	// initial bklogin handler.
	svc.bkloginHandler, err = newBKLoginHandler(svc.conf.BKSaas.BKLogin)
	if err != nil {
		return fmt.Errorf("failed to create bklogin handler: %w", err)
	}

	// initial backend handler.
	svc.Cap.BackendHandler, err = svc.newBackendHandler()
	if err != nil {
		return fmt.Errorf("failed to create backend handler: %w", err)
	}

	// initial file handler.
	svc.Cap.FileHandler, err = svc.newFileHandler()
	if err != nil {
		return fmt.Errorf("failed to create file handler: %w", err)
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

	// initial front setting.
	svc.Cap.FrontSetting, err = frontsetting.NewFrontSetting(
		frontsetting.Option{
			BKLoginURL:            svc.bkloginHandler.GetLoginURL(),
			BKRequestIDHeaderKEy:  bksaasheader.KeyBKRequestID,
			BKPassAnalyticsScript: svc.conf.BKPaas.AnalysisScript,
			PasswordVaultSwitch:   svc.conf.Front.PasswordVaultSwitch,
			PasswordVaultName:     svc.conf.Front.PasswordVaultName,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create front setting: %w", err)
	}

	return nil
}

func (svc *Service) newBackendHandler() (backend.IHandler, error) {
	apiGwClientConfig := newAPIGWAppConfig(&svc.conf.Backend.APIGatewayClient)

	apiGwClientCapability, err := newAPIGwClientCapability(clientNameBackend, &svc.conf.Backend.APIGatewayClient)
	if err != nil {
		return nil, fmt.Errorf("failed to new apigw client for backend: %w", err)
	}

	backendHandler, err := backend.New(apiGwClientCapability, backend.Config{
		APIGWAppConfig: apiGwClientConfig,
	})
	if err != nil {
		return nil, err
	}

	return backendHandler, nil
}

func (svc *Service) newFileHandler() (file.IHandler, error) {
	httpClient, err := restclient.NewHTTPClient(&ssl.TLSConfig{InsecureSkipVerify: true})
	if err != nil {
		return nil, fmt.Errorf("failed to create http client for file service: %w", err)
	}

	traceSvc, err := tracing.G().NewService(tracing.ServiceConfig{
		ServiceName: svc.conf.File.TraceServiceName,
		SampleRate:  svc.conf.File.TraceSampleRate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to new trace service: %w", err)
	}

	clientCap := &restclient.Capability{
		Name:       clientNameFile,
		HTTPClient: httpClient,
		Discover: restdiscovery.NewServiceDiscovery(
			svc.Cap.DiscoverProvider,
			discover.ServiceNameFile,
			discover.EndpointNameFileBasic,
		),
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
		MetricOpts:           restclient.MetricOption{},
		TraceSvc:             traceSvc,
	}

	return file.New(clientCap, &file.Config{
		RestJwtSecret: svc.conf.File.JWTClientConfig.SymmetricKey,
	})
}

func (svc *Service) newMongoClient() (*mongo.Client, error) {
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
			Hosts:          svc.conf.MongoDB.Hosts,
			ReadPreference: readpref.SecondaryPreferred(),
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

	svc.Cap.StorageConfigPolicyTemplate, err = cptemplate.NewStorage(
		svc.Cap.MongoClient,
		svc.conf.MongoDB.Database)
	if err != nil {
		return fmt.Errorf("failed to create config policy template storage: %w", err)
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

	return nil
}

func (svc *Service) newAuthIdentity(conf config.HTTPServer) (restserver.IAuthIdentity, error) {
	switch conf.AuthIdentity {
	case config.AuthIdentityNone:
		return restserver.NewNoneAuthIdentity(), nil

	case config.AuthIdentityBKLogin:
		return svc.bkloginHandler.GetAuthIdentity(), nil

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
			Name:             string(discover.EndpointNameApplicationInfo),
			IP:               svc.conf.InfoServer.BindIP,
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
	svc.instance.Update(discover.EndpointNameApplicationInfo, discover.Endpoint{
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

	authIdentity, err := svc.newAuthIdentity(svc.conf.AdminServer)
	if err != nil {
		return fmt.Errorf("failed to new auth identity: %w", err)
	}

	server, err := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:             string(discover.EndpointNameApplicationAdmin),
			IP:               svc.conf.AdminServer.BindIP,
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
	svc.instance.Update(discover.EndpointNameApplicationAdmin, discover.Endpoint{
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
			svc.conf.AdminServer.AuthIdentity, conv.MapKeyToSlice(svc.authIdentityValidMap))
	}

	authIdentity, err := svc.newAuthIdentity(svc.conf.BasicServer)
	if err != nil {
		return fmt.Errorf("failed to new auth identity: %w", err)
	}

	server, err := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:             string(discover.EndpointNameApplicationBasic),
			IP:               svc.conf.BasicServer.BindIP,
			Port:             svc.conf.BasicServer.Port,
			TLSConfig:        svc.conf.BasicServer.TLSConfig,
			TraceServiceName: svc.conf.BasicServer.TraceServiceName,
			TraceSampleRate:  svc.conf.BasicServer.TraceSampleRate,
			RequestIDSetter:  apigwserver.NewBKAPIRequestIDSetter(),
			StaticOptions: restserver.NewStaticOptions(svc.conf.BasicServer.StaticDir).
				WithHTMLs("index.html").
				WithDirs("assets").
				WithDirs("static").
				WithDirs("images").
				WithFiles("bk.svg", "favicon.png", "nodeman.png"),
		},
		restserver.WithPing(),
		withWeb(svc.Cap),
		withAPIV3Basic(svc.Cap,
			restserver.MiddlewareAuth(authIdentity),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to register basic server: %w", err)
	}

	svc.servers = append(svc.servers, server)
	svc.instance.Update(discover.EndpointNameApplicationBasic, discover.Endpoint{
		IPV4: svc.conf.BasicServer.AdvertiseIPV4,
		IPV6: svc.conf.BasicServer.AdvertiseIPV6,
		Port: svc.conf.BasicServer.Port,
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

// withWeb load web page handler.
func withWeb(capability *options.Capability, middleware ...gin.HandlerFunc) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		web.Load(rg, capability, middleware...)
	}
}

// withAPIV3Basic load api v3 basic.
func withAPIV3Basic(capability *options.Capability, middleware ...gin.HandlerFunc) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		applicationapiv3.LoadBasicAPIs(rg, capability, middleware...)
	}
}

// withAdmin load admin.
func withAdmin(capability *options.Capability, middleware ...gin.HandlerFunc) restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		admin.Load(rg, capability, middleware...)
	}
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
		ServiceName: conf.TraceServiceName,
		SampleRate:  conf.TraceSampleRate,
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
	apigwAppConf := apigwclient.NewAppConfig(conf.Endpoints, conf.AppCode, conf.AppSecret)

	return apigwAppConf
}

// newBKLoginHandler creates a new bklogin handler.
func newBKLoginHandler(conf config.BKLogin) (bksaasbklogin.IHandler, error) {
	httpClient, err := restclient.NewHTTPClient(&ssl.TLSConfig{
		InsecureSkipVerify: conf.TLS.InsecureSkipVerify,
		CertFile:           conf.TLS.CertFile,
		KeyFile:            conf.TLS.KeyFile,
		CAFile:             conf.TLS.CAFile,
		Password:           conf.TLS.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create http client for bklogin handler: %w", err)
	}

	traceSvc, err := tracing.G().NewService(tracing.ServiceConfig{
		ServiceName: conf.TraceServiceName,
		SampleRate:  conf.TraceSampleRate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to new trace service: %w", err)
	}

	clientCap := &restclient.Capability{
		Name:                 clientNameBKLogin,
		HTTPClient:           httpClient,
		Discover:             restdiscovery.NewDiscovery(clientNameBKLogin, []string{conf.LoginURL}),
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
		MetricOpts:           restclient.MetricOption{},
		TraceSvc:             traceSvc,
	}

	bkloginHandler, err := bksaasbklogin.New(
		clientCap,
		&bksaasbklogin.Config{LoginURL: conf.LoginURL, AuthType: conf.AuthType.String()},
	)
	if err != nil {
		return nil, err
	}

	return bkloginHandler, nil
}

// Start starts the application service.
func (svc *Service) Start() error {
	logger.G.Sys().Info("try to start application service")

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

	// wait until all servers stopped or application error.
	if err := gp.Wait(); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to start servers")

		return err
	}

	logger.G.Sys().Info("application service started")

	return nil
}

// GracefulShutdown gracefully shuts down the application service.
func (svc *Service) GracefulShutdown() error {
	logger.G.Sys().Info("try to gracefully shutdown application service")

	if svc.ctx == nil || svc.cancelFunc == nil {
		return errors.New("service is not running")
	}

	defer svc.cancelFunc()

	err := svc.Cap.GracefulShutdown()
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to gracefully shutdown capability")

		return err
	}

	logger.G.Sys().Info("application service gracefully shutdown")

	return nil
}

func (svc *Service) initTracing() error {
	tracingConf := tracing.Config{
		Exporter: tracing.ExporterConfig{
			ExporterType: tracing.ExporterType(svc.conf.Tracing.ExporterType),
		},
		Environment: system.GetEnv(),
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
