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

package server

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tracing"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/trace"
)

// MiddlewareContext verify auth info.
func MiddlewareContext() gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		_ = initRestRequest(gCtx)

		gCtx.Next()
	}
}

// IAuthIdentity verify auth info.
type IAuthIdentity interface {
	Verify(r IRequest) error
}

// AuthMiddlewareOption customizes auth middleware behavior.
type AuthMiddlewareOption func(opt *authMiddlewareConfig)

type authMiddlewareConfig struct {
	skipPathPrefixes []string
}

// WithSkipPathPrefixes configures path prefixes to bypass identity verification.
func WithSkipPathPrefixes(skipPathPrefixes ...string) AuthMiddlewareOption {
	return func(opt *authMiddlewareConfig) {
		opt.skipPathPrefixes = append(opt.skipPathPrefixes, skipPathPrefixes...)
	}
}

// MiddlewareAuth verify auth info.
func MiddlewareAuth(identity IAuthIdentity, options ...AuthMiddlewareOption) gin.HandlerFunc {
	config := &authMiddlewareConfig{}
	for _, option := range options {
		option(config)
	}

	normalizedSkipPathPrefixes := normalizeSkipPathPrefixes(config.skipPathPrefixes)

	return func(gCtx *gin.Context) {
		r := loadRestRequest(gCtx)

		if !shouldSkipAuthByPath(gCtx.Request.URL.Path, normalizedSkipPathPrefixes) {
			if err := identity.Verify(r); err != nil {
				r.AbortWithJSONError(resterrf.Unauthorized, []error{err})

				return
			}
		}

		gCtx.Next()
	}
}

func normalizeSkipPathPrefixes(skipPathPrefixes []string) []string {
	normalizedSkipPathPrefixes := make([]string, 0, len(skipPathPrefixes))
	for _, skipPathPrefix := range skipPathPrefixes {
		normalizedSkipPathPrefix := strings.TrimSuffix(skipPathPrefix, "/")
		if normalizedSkipPathPrefix == "" {
			continue
		}

		normalizedSkipPathPrefixes = append(normalizedSkipPathPrefixes, normalizedSkipPathPrefix)
	}

	return normalizedSkipPathPrefixes
}

func shouldSkipAuthByPath(path string, normalizedSkipPathPrefixes []string) bool {
	for _, skipPathPrefix := range normalizedSkipPathPrefixes {
		if path == skipPathPrefix || strings.HasPrefix(path, skipPathPrefix+"/") {
			return true
		}
	}

	return false
}

var _ IAuthIdentity = &RestServerAuthIdentity{}

// RestServerAuthIdentity verify auth info.
type RestServerAuthIdentity struct {
	jwtParser restheader.IBKNodeMgrAuthorizationParser
}

// TenantIDRegexp tenant id regexp.
const TenantIDRegexp = `^[a-z][a-z0-9-]{1,30}[a-z0-9]$`

// Verify verify auth info.
func (identity *RestServerAuthIdentity) Verify(r IRequest) error {
	tenantID := restheader.BKTenantIDGetter(r.GetRequest())

	mustCompile := regexp.MustCompile(TenantIDRegexp)
	if !mustCompile.MatchString(tenantID) {
		return errors.New("failed to set tenant id: invalid tenant id")
	}

	r.Data().SetTenantID(tenantID)

	authorization := restheader.BKNodeMgrAuthorizationGetter(r.GetRequest())

	nodeMgrAuthorization, err := identity.jwtParser.Parse(authorization)
	if err != nil {
		return fmt.Errorf("failed to parse bk node mgr authorization: %v", err)
	}

	r.Data().SetLoginName(nodeMgrAuthorization.LoginName)
	r.Data().SetBKUsername(nodeMgrAuthorization.BkUserName)

	return nil
}

// NewRestServerAuthIdentity ...
func NewRestServerAuthIdentity(jwtSecretStr string) *RestServerAuthIdentity {
	return &RestServerAuthIdentity{
		jwtParser: restheader.NewNodeMgrAuthorizationManager(jwtSecretStr),
	}
}

var _ IAuthIdentity = &NoneAuthIdentity{}

// NoneAuthIdentity verify auth info.
type NoneAuthIdentity struct {
}

// Verify verify auth info.
func (identity *NoneAuthIdentity) Verify(r IRequest) error {
	r.Data().SetTenantID("default")
	r.Data().SetLoginName("unknown")
	r.Data().SetBKUsername("unknown")

	return nil
}

// NewNoneAuthIdentity ...
func NewNoneAuthIdentity() *NoneAuthIdentity {
	return &NoneAuthIdentity{}
}

// IRequestIDSetter set request id.
type IRequestIDSetter interface {
	SetRequestID(r IRequest) error
}

// MiddlewareSetRequestID ...
func MiddlewareSetRequestID(requestIDSetter IRequestIDSetter) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		r := loadRestRequest(gCtx)

		if requestIDSetter == nil {
			r.AbortWithJSONError(resterrf.Aborted, []error{errors.New("this server doesn't load request id setter")})

			return
		}

		if err := requestIDSetter.SetRequestID(r); err != nil {
			r.AbortWithJSONError(resterrf.Aborted, []error{err})

			return
		}

		if r.Data().GetRequestID() == "" {
			r.AbortWithJSONError(resterrf.Aborted, []error{errors.New("failed to set request id")})
		}

		gCtx.Next()
	}
}

// RequestIDSetter this is a midleware for setting request id.
type RequestIDSetter struct {
}

// SetRequestID ...
func (setter *RequestIDSetter) SetRequestID(r IRequest) error {
	// note: for thread safety you need to reset it here.
	r.Data().SetRequestID(restheader.BKNodemgrRequestIDGetter(r.GetRequest()))

	if r.Data().GetRequestID() == "" {
		r.Data().SetRequestID(identifier.GenRequestID())
	}

	return nil
}

// NewRequestIDSetter ...
func NewRequestIDSetter() *RequestIDSetter {
	return &RequestIDSetter{}
}

// logSkipConfig defines log skip configuration.
type logSkipConfig struct {
	Path    string
	Methods []string // Methods to skip. Use allMethods() to skip all HTTP methods.
}

// allMethods returns all HTTP methods.
func allMethods() []string {
	return []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodHead,
		http.MethodOptions,
		http.MethodConnect,
		http.MethodTrace,
	}
}

// middlewareReceivedLog prints log when received request.
// nolint: contextcheck
func middlewareReceivedLog(skipConfigs []logSkipConfig) gin.HandlerFunc {
	pathMethodMap := make(map[string]map[string]struct{})
	for _, config := range skipConfigs {
		methodMap := make(map[string]struct{})
		for _, method := range config.Methods {
			methodMap[method] = struct{}{}
		}
		pathMethodMap[config.Path] = methodMap
	}

	return func(gCtx *gin.Context) {
		path := gCtx.Request.URL.Path
		method := gCtx.Request.Method

		methodMap, pathExists := pathMethodMap[path]
		if pathExists {
			if _, methodExists := methodMap[method]; methodExists {
				gCtx.Next()

				return
			}
		}

		rCtx, _ := GenRestContext(gCtx)
		fullPath := path
		raw := gCtx.Request.URL.RawQuery

		if raw != "" {
			fullPath = fullPath + "?" + raw
		}

		logger.G.Biz(rCtx).
			With("client-ip", gCtx.ClientIP()).
			Info("[request recv] %s %s", method, fullPath)

		gCtx.Next()
	}
}

// middlewareReturnedLog prints log when returned response.
// nolint: contextcheck
func middlewareReturnedLog(skipConfigs []logSkipConfig) gin.HandlerFunc {
	pathMethodMap := make(map[string]map[string]struct{})
	for _, config := range skipConfigs {
		methodMap := make(map[string]struct{})
		for _, method := range config.Methods {
			methodMap[method] = struct{}{}
		}
		pathMethodMap[config.Path] = methodMap
	}

	return func(gCtx *gin.Context) {
		start := time.Now()

		gCtx.Next()

		urlPath := gCtx.Request.URL.Path
		method := gCtx.Request.Method

		methodMap, pathExists := pathMethodMap[urlPath]
		if pathExists {
			if _, methodExists := methodMap[method]; methodExists {
				return
			}
		}

		rCtx, _ := GenRestContext(gCtx)
		fullPath := urlPath
		raw := gCtx.Request.URL.RawQuery

		if raw != "" {
			fullPath = fullPath + "?" + raw
		}

		logger.G.Biz(rCtx).
			WithDuration(time.Since(start)).
			With("client-ip", gCtx.ClientIP()).
			With("code", gCtx.Writer.Status()).
			Info("[request done] %s %s", method, fullPath)
	}
}

// MiddlewareTracing tracing.
func MiddlewareTracing(tracerSvc tracing.IService) []gin.HandlerFunc {
	middlewares := []gin.HandlerFunc{
		otelgin.Middleware(tracerSvc.ServiceName(),
			otelgin.WithTracerProvider(tracerSvc.TracerProvider()),
			otelgin.WithPropagators(tracerSvc.TracerPropagator()),
		),
		func(gCtx *gin.Context) {
			spanContext := trace.SpanContextFromContext(gCtx)

			gCtx.Writer.Header().Set(restheader.TraceID, spanContext.TraceID().String())

			gCtx.Next()
		},
	}

	return middlewares
}
