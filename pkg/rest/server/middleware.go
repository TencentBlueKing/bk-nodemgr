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
	"fmt"
	"io"
	"regexp"

	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/gin-gonic/gin"
)

// AuthIdentity verify auth info.
type AuthIdentity interface {
	Verify(rCtx *Context) error
}

// MiddlewareContext verify auth info.
func MiddlewareContext() gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		_ = initRestContext(gCtx)

		gCtx.Next()
	}
}

// MiddlewareAuth verify auth info.
func MiddlewareAuth(identity AuthIdentity) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		rCtx := loadRestContext(gCtx)

		if err := identity.Verify(rCtx); err != nil {
			rCtx.AbortWithJSONError(resterrf.Unauthorized, []error{err})

			return
		}

		gCtx.Next()
	}

}

// IRequestIDSetter set request id.
type IRequestIDSetter interface {
	SetRequestID(*Context) error
}

// MiddlewareSetRequestID ...
func MiddlewareSetRequestID(requestIDSetter IRequestIDSetter) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		rCtx := loadRestContext(gCtx)

		if requestIDSetter == nil {
			rCtx.AbortWithJSONError(resterrf.Aborted, []error{fmt.Errorf("this server doesn't load request id setter")})

			return
		}

		if err := requestIDSetter.SetRequestID(rCtx); err != nil {
			rCtx.AbortWithJSONError(resterrf.Aborted, []error{err})

			return
		}

		if rCtx.RequestID == "" {
			rCtx.AbortWithJSONError(resterrf.Aborted, []error{fmt.Errorf("failed to set request id")})
		}

		gCtx.Next()
	}
}

// RequestIDSetter this is a midleware for setting request id.
type RequestIDSetter struct {
}

// SetRequestID ...
func (setter *RequestIDSetter) SetRequestID(rCtx *Context) error {
	// note: for thread safety you need to reset it here.
	rCtx.RequestID = restheader.BKNodemgrRequestIDGetter(rCtx.Request())

	if rCtx.RequestID == "" {
		rCtx.RequestID = identifier.GenRequestID()
	}

	return nil
}

// NewRequestIDSetter ...
func NewRequestIDSetter() *RequestIDSetter {
	return &RequestIDSetter{}
}

// ITenantIDSetter set request id.
type ITenantIDSetter interface {
	SetTenantID(*Context) error
}

// MiddlewareSetTenantID ...
func MiddlewareSetTenantID(tenantIDSetter ITenantIDSetter) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		rCtx := loadRestContext(gCtx)

		if tenantIDSetter == nil {
			rCtx.AbortWithJSONError(resterrf.Aborted, []error{fmt.Errorf("this server doesn't load tenant id setter")})

			return
		}

		if err := tenantIDSetter.SetTenantID(rCtx); err != nil {
			rCtx.AbortWithJSONError(resterrf.Aborted, []error{err})

			return
		}

		if rCtx.TenantID == "" {
			rCtx.AbortWithJSONError(resterrf.Aborted, []error{fmt.Errorf("failed to set tenant id")})
		}

		gCtx.Next()
	}
}

var _ ITenantIDSetter = &TenantIDSetter{}

// TenantIDSetter this is a midleware for setting request id.
type TenantIDSetter struct {
}

// SetTenantID ...
func (setter *TenantIDSetter) SetTenantID(rCtx *Context) error {
	// note: for thread safety you need to reset it here.
	rCtx.TenantID = restheader.BKTenantIDGetter(rCtx.Request())

	if rCtx.TenantID == "" {
		tenantID, err := tenant.GetID(rCtx)
		if err != nil {
			return fmt.Errorf("failed to set tenant id: %w", err)
		}

		rCtx.TenantID = tenantID
	}

	mustCompile := regexp.MustCompile(`^[a-z][a-z0-9-]{1,30}[a-z0-9]$`)
	if !mustCompile.MatchString(rCtx.TenantID) {
		return fmt.Errorf("failed to set tenant id: invalid tenant id")
	}

	return nil
}

// NewTenantIDSetter ...
func NewTenantIDSetter() *TenantIDSetter {
	return &TenantIDSetter{}
}

type recvLoggerConfig struct {
	Output    io.Writer
	Formatter func(*gin.Context) string
	SkipPaths []string
}

// MiddlewareReceivedLog print log when received request.
func MiddlewareReceivedLog(conf recvLoggerConfig) gin.HandlerFunc {
	if conf.Output == nil || conf.Formatter == nil {
		return func(gCtx *gin.Context) {
			gCtx.Next()
		}
	}

	skip := make(map[string]struct{})
	for _, skipPath := range conf.SkipPaths {
		skip[skipPath] = struct{}{}
	}

	return func(gCtx *gin.Context) {
		if _, ok := skip[gCtx.Request.URL.Path]; !ok {
			_, _ = fmt.Fprint(conf.Output, conf.Formatter(gCtx))
		}

		gCtx.Next()
	}
}
