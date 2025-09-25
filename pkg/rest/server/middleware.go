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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/gin-gonic/gin"
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
}

// Verify verify auth info.
func (identity *RestServerAuthIdentity) Verify(r IRequest) error {
	nodeMgrAuthorization := restheader.BKNodeMgrAuthorizationGetter(r.GetRequest())

	authInfo := make(map[string]string, 0)
	if err := json.Unmarshal([]byte(nodeMgrAuthorization), &authInfo); err != nil {
		return fmt.Errorf("failed to verify rest auth indentity, auth info(%s): %w", nodeMgrAuthorization, err)
	}

	r.Data().SetLoginName(authInfo["login_name"])
	r.Data().SetBKUsername(authInfo["bk_username"])

	return nil
}

// NewRestServerAuthIdentity ...
func NewRestServerAuthIdentity() *RestServerAuthIdentity {
	return &RestServerAuthIdentity{}
}

var _ IAuthIdentity = &NodeAuthIdentity{}

// NodeAuthIdentity verify auth info.
type NodeAuthIdentity struct {
}

// Verify verify auth info.
func (identity *NodeAuthIdentity) Verify(r IRequest) error {
	r.Data().SetLoginName("unknown")
	r.Data().SetBKUsername("unknown")

	return nil
}

// NewNodeAuthIdentity ...
func NewNodeAuthIdentity() *NodeAuthIdentity {
	return &NodeAuthIdentity{}
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

// ITenantIDSetter set request id.
type ITenantIDSetter interface {
	SetTenantID(r IRequest) error
}

// MiddlewareSetTenantID ...
func MiddlewareSetTenantID(tenantIDSetter ITenantIDSetter) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		r := loadRestRequest(gCtx)

		if tenantIDSetter == nil {
			r.AbortWithJSONError(resterrf.Aborted, []error{errors.New("this server doesn't load tenant id setter")})

			return
		}

		if err := tenantIDSetter.SetTenantID(r); err != nil {
			r.AbortWithJSONError(resterrf.Aborted, []error{err})

			return
		}

		if r.data.GetTenantID() == "" {
			r.AbortWithJSONError(resterrf.Aborted, []error{errors.New("failed to set tenant id")})
		}

		gCtx.Next()
	}
}

var _ ITenantIDSetter = &TenantIDSetter{}

// TenantIDSetter this is a midleware for setting request id.
type TenantIDSetter struct {
}

// TenantIDRegexp tenant id regexp.
const TenantIDRegexp = `^[a-z][a-z0-9-]{1,30}[a-z0-9]$`

// SetTenantID ...
func (setter *TenantIDSetter) SetTenantID(r IRequest) error {
	// note: for thread safety you need to reset it here.
	r.Data().SetTenantID(restheader.BKTenantIDGetter(r.GetRequest()))

	if r.Data().GetTenantID() == "" {
		tenantID, err := tenant.GetID(r.GContext())
		if err != nil {
			return fmt.Errorf("failed to set tenant id: %w", err)
		}

		r.Data().SetTenantID(tenantID)
	}

	mustCompile := regexp.MustCompile(TenantIDRegexp)
	if !mustCompile.MatchString(r.Data().GetTenantID()) {
		return errors.New("failed to set tenant id: invalid tenant id")
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
