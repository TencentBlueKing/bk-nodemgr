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
	"io"
	"runtime"

	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/router/healthz"
	"github.com/TencentBlueKing/bk-nodemgr/internal/application/router/web"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/blog"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Service defines a server that provides application services.
// It provides a website for user to operate with nodeman.
type Service struct {
	// conf holds the configuration for the service.
	conf *config.ApplicationService

	// ctx is used to control the service lifecycle (cancellation and timeouts).
	ctx context.Context
	// cancelFunc is used to cancel the service and all associated operations.
	cancelFunc context.CancelFunc

	// router is the entry point of the service, routing requests to different capabilities.
	servers []*rest.Server

	// Note: Capability is initialized in the Start() and could not be used in other package.
	// Capability is the capability of the service.
	Capability *options.Capability
}

const (
	// RouterNameHTTPServer defines the name of http server router.
	RouterNameHTTPServer = "http-server"
)

// NewService creates a new application service.
func NewService(conf *config.ApplicationService) *Service {
	svc := &Service{
		conf: conf,
	}
	svc.ctx, svc.cancelFunc = context.WithCancel(context.Background())

	httpServer := rest.NewServer(svc.ctx, RouterNameHTTPServer, conf.HTTPServer.BindIP, conf.HTTPServer.Port,
		loggerWriter{},
		rest.NewStaitcOptions(conf.HTTPServer.StaticDir).
			WithHTMLs("index.html").
			WithDirs("assets").
			WithFiles("bk.svg", "favicon.png", "nodeman.png"),
		rest.WithPing(),
		withHealthz(svc.Capability),
		withMetrics(svc.Capability),
		withWeb(svc.Capability),
	)

	svc.servers = append(svc.servers, httpServer)

	return svc
}

// loggerWriter implements rest.LoggerWriter.
type loggerWriter struct{}

func (l loggerWriter) InfoWriter() io.Writer {
	return blog.WriterInfo{}
}

func (l loggerWriter) ErrorWriter() io.Writer {
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

// withWeb load web page handler.
func withWeb(capability *options.Capability) rest.OptionFunc {
	return func(rg *gin.RouterGroup) {
		web.Load(rg, capability)
	}
}

// Start starts the application service.
func (svc *Service) Start(ctx context.Context) error {
	runtime.GOMAXPROCS(runtime.NumCPU())

	logConfig := blog.NewLogConfig()
	logConfig.LogDir = svc.conf.Log.Dir
	logConfig.LogMaxSizeMB = svc.conf.Log.MaxSizeMB
	logConfig.LogMaxNum = svc.conf.Log.MaxNum
	logConfig.Level = svc.conf.Log.Level
	logConfig.ToStdErr = svc.conf.Log.ToStdErr
	logConfig.AlsoToStdErr = svc.conf.Log.AlsoToStdErr
	blog.InitLogs(logConfig)

	svc.ctx, svc.cancelFunc = context.WithCancel(ctx)

	// start servers
	gp := gopool.NewPool()
	for idx, _ := range svc.servers {
		server := svc.servers[idx]

		// server start will block until router stop, so we need to run it in a goroutine.
		fn := func() error {
			if err := server.Start(); err != nil {
				return err
			}

			return nil
		}
		gp.Go(fn)
		blog.Infof("started server. name(%s), ip(%s), port(%d)", server.Name(), server.IP(), server.Port())
	}

	// wait until all routers stopped or application error.
	if err := gp.Wait(); err != nil {
		blog.Errorf("failed to start servers, err: %v", err)
		return err
	}

	return nil
}
