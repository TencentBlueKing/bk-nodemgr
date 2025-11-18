/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package server is the restful API server.
package server

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	restmetrics "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/metrics"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tracing"
	"github.com/gin-gonic/gin"
)

// Server defines the restful API server.
type Server struct {
	engine *gin.Engine
	// rg it contains the rest context, please use to realize some business logic.
	rg  *gin.RouterGroup
	ctx context.Context

	opts Options

	metrics *restmetrics.Monitor

	// tracerSvc is the OpenTelemetry tracerSvc for distributed tracing
	tracerSvc tracing.IService
}

// OptionFunc defines a function that can be used to modify the router.
type OptionFunc func(rg *gin.RouterGroup)

// WithPing with ping pong api.
func WithPing() OptionFunc {
	return func(rg *gin.RouterGroup) {
		rg.Any("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})
	}
}

type staticResourcePair struct {
	relative string
	target   string
}

// NewStaticOptions generates a new static options.
func NewStaticOptions(baseStaticDir string) *StaticOptions {
	return &StaticOptions{
		baseStaticDir: baseStaticDir,
		dirs:          make([]*staticResourcePair, 0),
		files:         make([]*staticResourcePair, 0),
		htmls:         make([]string, 0),
	}
}

// StaticOptions describes the static file settings and routers.
type StaticOptions struct {
	// BaseStaticDir is the base dir of all static files.
	baseStaticDir string

	dirs []*staticResourcePair

	files []*staticResourcePair

	htmls []string
}

// WithDirs loads static dirs into options.
func (opt *StaticOptions) WithDirs(relatives ...string) *StaticOptions {
	for _, relative := range relatives {
		opt.dirs = append(opt.dirs, &staticResourcePair{
			relative: "/" + relative,
			target:   path.Join(opt.baseStaticDir, relative),
		})
	}

	return opt
}

// WithFiles loads static files into options.
func (opt *StaticOptions) WithFiles(relatives ...string) *StaticOptions {
	for _, relative := range relatives {
		opt.files = append(opt.files, &staticResourcePair{
			relative: "/" + relative,
			target:   path.Join(opt.baseStaticDir, relative),
		})
	}

	return opt
}

// WithHTMLs loads static htmls into options.
func (opt *StaticOptions) WithHTMLs(relatives ...string) *StaticOptions {
	for _, relative := range relatives {
		opt.htmls = append(opt.htmls, path.Join(opt.baseStaticDir, relative))
	}

	return opt
}

// Options describes the server options.
type Options struct {
	Name             string
	IP               string
	Port             int
	RequestIDSetter  IRequestIDSetter
	StaticOptions    *StaticOptions
	TraceServiceName string
	TraceSampleRate  float64
}

// NewServer creates a new restful API server.
func NewServer(ctx context.Context, opts Options, apiOptFns ...OptionFunc) (*Server, error) {
	svr := &Server{
		ctx:  ctx,
		opts: opts,
		engine: gin.New(func(engine *gin.Engine) {
			engine.RedirectTrailingSlash = false
			engine.RedirectFixedPath = false
			// link tracing must open this option.
			engine.ContextWithFallback = true
		}),
	}
	gin.DebugPrintFunc = logger.G.Sys().Debug

	var err error
	svr.tracerSvc, err = tracing.G().NewService(tracing.ServiceConfig{
		ServiceName: svr.opts.TraceServiceName,
		SampleRate:  svr.opts.TraceSampleRate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create tracer service: %w", err)
	}

	// Recover from panic
	svr.engine.Use(gin.RecoveryWithWriter(logger.G.Biz(nil).ErrorWriter()))

	// nolint: contextcheck
	svr.engine.Use(MiddlewareTracing(svr.tracerSvc))

	// Set authentication middleware.
	svr.engine.Use(MiddlewareContext())

	// Set request id middleware.
	svr.engine.Use(MiddlewareSetRequestID(opts.RequestIDSetter))

	// Set received log middleware.
	svr.engine.Use(MiddlewareReceivedLog("/ping", "/healthz", "/metrics"))

	// Set done log middleware.
	svr.engine.Use(MiddlewareReturnedLog("/ping", "/healthz", "/metrics"))

	// Set metrics monitor.
	svr.metrics = restmetrics.NewMonitor("server_"+opts.Name,
		restmetrics.WithSlowTime(1*time.Second),
		restmetrics.WithExcludePaths([]string{"/ping", "/healthz", "/metrics"}),
	).RegisterMiddleware(svr.engine).Enable()

	svr.rg = svr.engine.Group("/")

	// Set static settings.
	if opts.StaticOptions != nil {
		// load html templates.
		svr.engine.LoadHTMLFiles(opts.StaticOptions.htmls...)

		// load static dirs.
		for _, item := range opts.StaticOptions.dirs {
			svr.engine.Static(item.relative, item.target)
		}

		// load static files.
		for _, item := range opts.StaticOptions.files {
			svr.engine.StaticFile(item.relative, item.target)
		}
	}

	for _, fn := range apiOptFns {
		fn(svr.rg)
	}

	return svr, nil
}

// Start starts the router.
func (svr *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", svr.opts.IP, svr.opts.Port)
	if err := svr.engine.Run(addr); err != nil {
		return err
	}

	return nil
}

// Name returns the router name.
func (svr *Server) Name() string {
	return svr.opts.Name
}

// IP returns the router ip.
func (svr *Server) IP() string {
	return svr.opts.IP
}

// Port returns the router port.
func (svr *Server) Port() int {
	return svr.opts.Port
}
