/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package service provides mock-server service.
package service

import (
	"fmt"
	"runtime"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/test/mock-server/router"
	"github.com/TencentBlueKing/bk-nodemgr/test/mock-server/router/bkrepo"
	"github.com/TencentBlueKing/bk-nodemgr/test/mock-server/router/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/test/mock-server/router/healthz"

	"github.com/gin-gonic/gin"
)

const (
	mockServerBasicSvcName = "mock-server-basic"
)

// Config holds the configuration for mock-server service.
type Config struct {
	// basic server config of mock-server.
	BasicServer config.HTTPServer

	// cmdb config of mock-server.
	CMDBConfig *cmdb.Config

	// bk-repo config of mock-server.
	BKRepoConfig *bkrepo.Config

	// mock data holds the optional preset mock data.
	MockData *router.MockData
}

// Service defines a server that provides mock services for testing.
// It manages the configuration, lifecycle, and mock API routes.
type Service struct {
	// conf holds the configuration for the service.
	conf Config

	// ctx is used to control the service lifecycle (cancellation and timeouts).
	ctx contextx.IContext

	// servers is the HTTP REST server instances.
	servers []*restserver.Server
}

// NewService creates a new mock-server service.
func NewService(conf Config) (*Service, error) {
	svc := &Service{
		conf: conf,
	}

	svc.ctx = contextx.New(contextx.Background())

	if err := svc.registerRestServer(); err != nil {
		return nil, fmt.Errorf("failed to register rest server: %w", err)
	}

	return svc, nil
}

// registerRestServer registers the REST servers with routes and middlewares.
func (svc *Service) registerRestServer() error {
	if err := svc.registerBasicServer(); err != nil {
		return fmt.Errorf("failed to register basic server: %w", err)
	}

	return nil
}

// registerBasicServer registers the basic server with mock APIs.
func (svc *Service) registerBasicServer() error {
	server, err := restserver.NewServer(
		svc.ctx,
		restserver.Options{
			Name:             mockServerBasicSvcName,
			IP:               svc.conf.BasicServer.BindIP,
			Port:             svc.conf.BasicServer.Port,
			TLSConfig:        svc.conf.BasicServer.TLSConfig,
			RequestIDSetter:  restserver.NewRequestIDSetter(),
			TraceServiceName: mockServerBasicSvcName,
		},
		restserver.WithPing(),
		withHealthz(),
		svc.withMockAPIs(),
	)
	if err != nil {
		return fmt.Errorf("failed to create basic server: %w", err)
	}

	svc.servers = append(svc.servers, server)

	return nil
}

// withHealthz load healthz.
func withHealthz() restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		healthz.Load(rg)
	}
}

// withMockAPIs registers mock APIs.
func (svc *Service) withMockAPIs() restserver.OptionFunc {
	return func(rg *gin.RouterGroup) {
		router.Load(rg,
			svc.conf.CMDBConfig,
			svc.conf.BKRepoConfig,
			svc.conf.MockData,
		)
	}
}

// Start starts the mock-server service.
func (svc *Service) Start() error {
	logger.G.Sys().Info("try to start mock-server service")

	runtime.GOMAXPROCS(runtime.NumCPU())

	// start servers
	gp := gopool.NewPool()
	for idx := range svc.servers {
		server := svc.servers[idx]

		// http server start will block until http server stop, so we need to run it in a goroutine.
		fn := func() error {
			logger.G.Sys().With("name", server.Name(), "ip", server.IP(), "port", server.Port()).Info("started HTTP Server")
			if err := server.Start(); err != nil {
				return err
			}

			return nil
		}
		gp.Go(fn)
	}

	// wait until all servers stopped or error.
	if err := gp.Wait(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
