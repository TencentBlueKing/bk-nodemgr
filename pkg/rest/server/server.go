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

// Package server is the restful API server.
package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	restmetrics "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/metrics"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tracing"
	"github.com/gin-gonic/gin"
)

const (
	defaultShutdownTimeout = 60 * time.Second
	maxListenAddressCount  = 2
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

	serversMu       sync.Mutex
	servers         []*http.Server
	shutdownStarted bool

	shutdownOnce sync.Once
	shutdownDone chan struct{}
	shutdownErr  error
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
	IPV6             string
	Port             int
	RequestIDSetter  IRequestIDSetter
	StaticOptions    *StaticOptions
	TLSConfig        config.TLSConfig
	ShutdownTimeout  time.Duration
	TraceServiceName string
	TraceSampleRate  float64
}

// NewServer creates a new restful API server.
func NewServer(ctx context.Context, opts Options, apiOptFns ...OptionFunc) (*Server, error) {
	if opts.ShutdownTimeout <= 0 {
		opts.ShutdownTimeout = defaultShutdownTimeout
	}

	svr := &Server{
		ctx:          ctx,
		opts:         opts,
		shutdownDone: make(chan struct{}),
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
		ServiceName:     svr.opts.TraceServiceName,
		ServiceCategory: tracing.ServiceCategoryHTTP,
		SampleRate:      svr.opts.TraceSampleRate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create tracer service: %w", err)
	}

	// Recover from panic
	svr.engine.Use(gin.RecoveryWithWriter(logger.G.Biz(nil).ErrorWriter()))

	// Set authentication middleware.
	svr.engine.Use(MiddlewareContext())

	// nolint: contextcheck
	svr.engine.Use(MiddlewareTracing(svr.tracerSvc)...)

	// Set request id middleware.
	svr.engine.Use(MiddlewareSetRequestID(opts.RequestIDSetter))

	// Set received log middleware.
	svr.engine.Use(middlewareReceivedLog([]logSkipConfig{
		{Path: "/ping", Methods: allMethods()},
		{Path: "/healthz", Methods: allMethods()},
		{Path: "/metrics", Methods: allMethods()},
		{Path: "/", Methods: []string{http.MethodHead}},
	}))

	// Set done log middleware.
	svr.engine.Use(middlewareReturnedLog([]logSkipConfig{
		{Path: "/ping", Methods: allMethods()},
		{Path: "/healthz", Methods: allMethods()},
		{Path: "/metrics", Methods: allMethods()},
		{Path: "/", Methods: []string{http.MethodHead}},
	}))

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
	if svr.opts.IP == "" && svr.opts.IPV6 == "" {
		return fmt.Errorf("IP and IPV6 cannot be empty at the same time")
	}

	gp := gopool.NewPool()
	gp.SetLimit(maxListenAddressCount)

	if svr.IP() != "" {
		gp.Go(func() error {
			logger.G.Sys().With("name", svr.Name(), "ip", svr.IP(), "port", svr.Port()).Info("started HTTP Server")

			return svr.startToListenIPV4()
		})
	}

	if svr.IPV6() != "" {
		gp.Go(func() error {
			logger.G.Sys().With("name", svr.Name(), "ipv6", svr.IPV6(), "port", svr.Port()).Info("started HTTP Server")

			return svr.startToListenIPV6()
		})
	}

	return gp.Wait()
}

func (svr *Server) startToListenIPV4() error {
	addr := fmt.Sprintf("%s:%d", svr.opts.IP, svr.opts.Port)

	// tls server.
	if svr.opts.TLSConfig.CAFile != "" && svr.opts.TLSConfig.CertFile != "" && svr.opts.TLSConfig.KeyFile != "" {
		return svr.startWithTLS(criteria.NetTypeTCP4, addr)
	}

	return svr.startWithoutTLS(criteria.NetTypeTCP4, addr)
}

func (svr *Server) startToListenIPV6() error {
	addr := fmt.Sprintf("[%s]:%d", svr.opts.IPV6, svr.opts.Port)

	// tls server.
	if svr.opts.TLSConfig.CAFile != "" && svr.opts.TLSConfig.CertFile != "" && svr.opts.TLSConfig.KeyFile != "" {
		return svr.startWithTLS(criteria.NetTypeTCP6, addr)
	}

	return svr.startWithoutTLS(criteria.NetTypeTCP6, addr)
}

func (svr *Server) startWithoutTLS(network criteria.NetType, addr string) error {
	listener, err := net.Listen(string(network), addr)
	if err != nil {
		return fmt.Errorf("failed to listen %s %s: %w", network, addr, err)
	}

	server, err := svr.registerHTTPServer(addr, nil)
	if err != nil {
		_ = listener.Close()

		return ignoreServerClosed(err)
	}

	return ignoreServerClosed(server.Serve(listener))
}

func (svr *Server) startWithTLS(network criteria.NetType, addr string) error {
	conf := &ssl.TLSConfig{
		InsecureSkipVerify: svr.opts.TLSConfig.InsecureSkipVerify,
		VerifyClient:       svr.opts.TLSConfig.VerifyClient,
		CertFile:           svr.opts.TLSConfig.CertFile,
		KeyFile:            svr.opts.TLSConfig.KeyFile,
		CAFile:             svr.opts.TLSConfig.CAFile,
		Password:           svr.opts.TLSConfig.Password,
	}
	tlsConfig, err := conf.NewServerTLSConf()
	if err != nil {
		return fmt.Errorf("failed to create server tls config: %w", err)
	}

	listener, err := net.Listen(string(network), addr)
	if err != nil {
		return fmt.Errorf("failed to listen %s %s: %w", network, addr, err)
	}

	server, err := svr.registerHTTPServer(addr, tlsConfig)
	if err != nil {
		_ = listener.Close()

		return ignoreServerClosed(err)
	}

	return ignoreServerClosed(server.ServeTLS(listener, "", ""))
}

// Shutdown gracefully shuts down all listeners owned by the server.
func (svr *Server) Shutdown(ctx context.Context) error {
	if ctx == nil {
		return errors.New("shutdown context is nil")
	}

	svr.shutdownOnce.Do(func() {
		svr.shutdownErr = svr.shutdown(ctx)
		close(svr.shutdownDone)
	})
	<-svr.shutdownDone

	return svr.shutdownErr
}

func (svr *Server) registerHTTPServer(addr string, tlsConfig *tls.Config) (*http.Server, error) {
	server := &http.Server{
		Addr:      addr,
		Handler:   svr.engine.Handler(),
		TLSConfig: tlsConfig,
	}

	svr.serversMu.Lock()
	defer svr.serversMu.Unlock()
	if svr.shutdownStarted {
		return nil, http.ErrServerClosed
	}

	svr.servers = append(svr.servers, server)

	return server, nil
}

func (svr *Server) shutdown(ctx context.Context) error {
	svr.serversMu.Lock()
	svr.shutdownStarted = true
	servers := append([]*http.Server(nil), svr.servers...)
	svr.serversMu.Unlock()

	if len(servers) == 0 {
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, svr.opts.ShutdownTimeout)
	defer cancel()

	gp := gopool.NewPool()
	gp.SetLimit(len(servers))
	for idx := range servers {
		server := servers[idx]
		gp.Go(func() error {
			if err := server.Shutdown(shutdownCtx); err != nil {
				if shutdownCtx.Err() != nil {
					if closeErr := server.Close(); closeErr != nil {
						return errors.Join(err, closeErr)
					}

					return err
				}

				return err
			}

			return nil
		})
	}

	return gp.Wait()
}

func ignoreServerClosed(err error) error {
	if !errors.Is(err, http.ErrServerClosed) {
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

// IPV6 returns the router ipv6.
func (svr *Server) IPV6() string {
	return svr.opts.IPV6
}

// Port returns the router port.
func (svr *Server) Port() int {
	return svr.opts.Port
}
