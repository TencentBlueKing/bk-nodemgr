/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package server

import (
	"errors"
	"fmt"
	"regexp"
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

// MiddlewareAuth verify auth info.
func MiddlewareAuth(identity IAuthIdentity) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		r := loadRestRequest(gCtx)

		if err := identity.Verify(r); err != nil {
			r.AbortWithJSONError(resterrf.Unauthorized, []error{err})

			return
		}

		gCtx.Next()
	}
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

// MiddlewareReceivedLog print log when received request.
// nolint: contextcheck
func MiddlewareReceivedLog(skipPaths ...string) gin.HandlerFunc {
	skip := make(map[string]struct{})
	for _, skipPath := range skipPaths {
		skip[skipPath] = struct{}{}
	}

	return func(gCtx *gin.Context) {
		if _, ok := skip[gCtx.Request.URL.Path]; !ok {
			rCtx, _ := GenRestContext(gCtx)

			path := gCtx.Request.URL.Path
			raw := gCtx.Request.URL.RawQuery

			if raw != "" {
				path = path + "?" + raw
			}

			logger.G.Biz(rCtx).With("client-ip", gCtx.ClientIP()).Info("[request recv] %s", path)
		}

		gCtx.Next()
	}
}

// MiddlewareReturnedLog print log when returned response.
// nolint: contextcheck
func MiddlewareReturnedLog(skipPaths ...string) gin.HandlerFunc {
	skip := make(map[string]struct{})
	for _, skipPath := range skipPaths {
		skip[skipPath] = struct{}{}
	}

	return func(gCtx *gin.Context) {
		start := time.Now()

		gCtx.Next()

		if _, ok := skip[gCtx.Request.URL.Path]; !ok {
			rCtx, _ := GenRestContext(gCtx)

			path := gCtx.Request.URL.Path
			raw := gCtx.Request.URL.RawQuery

			if raw != "" {
				path = path + "?" + raw
			}

			logger.G.Biz(rCtx).
				WithDuration(time.Since(start)).
				With("client-ip", gCtx.ClientIP()).
				With("code", gCtx.Writer.Status()).
				Info("[request done] %s", path)
		}
	}
}

// MiddlewareTracing tracing.
func MiddlewareTracing(tracerSvc tracing.IService) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		fn := otelgin.Middleware(tracerSvc.ServiceName(),
			otelgin.WithTracerProvider(tracerSvc.TracerProvider()),
			otelgin.WithPropagators(tracerSvc.TracerPropagator()),
		)
		fn(gCtx)

		span := trace.SpanFromContext(gCtx.Request.Context())
		if !span.SpanContext().IsValid() {
			traceID := identifier.GenTraceID()
			spanID := identifier.GenSpanID()

			span.SpanContext().WithTraceID(traceID)
			span.SpanContext().WithSpanID(spanID)
		}

		gCtx.Next()
	}
}
