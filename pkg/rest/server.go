/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package rest is the restful API router.
package rest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/metrics"
	"github.com/gin-gonic/gin"
)

// LogWriter defines the log writer.
type LogWriter interface {
	InfoWriter() io.Writer
	ErrorWriter() io.Writer
}

// Server defines the restful API server.
type Server struct {
	engine *gin.Engine
	// rg it contains the rest context, please use to realize some business logic.
	rg  *gin.RouterGroup
	ctx context.Context

	opts ServerOptions

	metrics *metrics.Monitor
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

// ServerOptions describes the server options.
type ServerOptions struct {
	Name          string
	IP            string
	Port          int
	LogWriter     LogWriter
	StaticOptions *StaticOptions
}

// NewServer creates a new restful API server.
func NewServer(ctx context.Context,
	opts ServerOptions,
	apiOptFns ...OptionFunc) *Server {

	svr := &Server{
		ctx:  ctx,
		opts: opts,
		engine: gin.New(func(engine *gin.Engine) {
			engine.RedirectTrailingSlash = false
			engine.RedirectFixedPath = false
		}),
	}

	// Recover from panic
	svr.engine.Use(gin.RecoveryWithWriter(opts.LogWriter.ErrorWriter()))

	// Set authentication middleware.
	svr.engine.Use(MiddlewareContext())

	// Set received log middleware.
	svr.engine.Use(MiddlewareReceivedLog(recvLoggerConfig{
		Output:    opts.LogWriter.InfoWriter(),
		Formatter: customLogRecvFormatter,
		SkipPaths: []string{"/ping", "/healthz", "/metrics"},
	}))

	// Set done log middleware.
	svr.engine.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output:    opts.LogWriter.InfoWriter(),
		Formatter: customLogDoneFormatter,
		SkipPaths: []string{"/ping", "/healthz", "/metrics"},
	}))

	// Set metrics monitor.
	svr.metrics = metrics.NewMonitor(opts.Name).
		WithSlowTime(1 * time.Second).
		WithExcludePaths([]string{"/ping", "/healthz", "/metrics"}).
		RegisterMiddleware(svr.engine).
		Enable()

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

	return svr
}

// customLogRecvFormatter is a custom log recv formatter.
func customLogRecvFormatter(gCtx *gin.Context) string {
	path := gCtx.Request.URL.Path
	raw := gCtx.Request.URL.RawQuery

	if raw != "" {
		path = path + "?" + raw
	}

	return fmt.Sprintf("%s[request recv] %s | %s",
		logWithCtxKeys(gCtx.Keys), path, gCtx.ClientIP())
}

// customLogDoneFormatter is a custom log done formatter.
func customLogDoneFormatter(param gin.LogFormatterParams) string {
	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}

	return fmt.Sprintf("%s[request done] %s | %s | code(%3d) cost(%dms) %s",
		logWithCtxKeys(param.Keys), param.Path, param.ClientIP,
		param.StatusCode, param.Latency.Milliseconds(), param.ErrorMessage,
	)
}

func logWithCtxKeys(keys map[string]any) string {
	if v, ok := keys[restContextKey]; ok {
		if ctx, ok := v.(*Context); ok {
			return fmt.Sprintf("[%s][tenant:%s][user:%s]",
				ctx.RequestID, ctx.TenantID, ctx.Username)
		}
	}

	return ""
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
