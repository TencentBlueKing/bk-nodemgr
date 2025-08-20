/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package server defines rest server context.
package server

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"time"

	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime"
	"github.com/gin-gonic/gin"
)

// restContextKey was used to store the restContext in gin.Context.
const restContextKey = "rest_context"

// initRestContext initializes a new rest context.
// notice: please don't set Context's fields here, use Middleware to set fields.
func initRestContext(gCtx *gin.Context) *Context {
	restContext := &Context{
		gCtx: gCtx,
	}

	gCtx.Set(restContextKey, restContext)

	return restContext
}

// loadRestContext loads rest context from gin.Context.
func loadRestContext(gCtx *gin.Context) *Context {
	restContext, ok := gCtx.Value(restContextKey).(*Context)
	if !ok {
		return initRestContext(gCtx)
	}

	return restContext
}

// GetRestContext only when user has authenticated, otherwise, return ErrorUnauthorized.
func GetRestContext(c *gin.Context) (*Context, error) {
	ctxObj, ok := c.Get(restContextKey)
	if !ok {
		return nil, resterrf.CodeErrMap(resterrf.Unauthorized)
	}

	restContext, ok := ctxObj.(*Context)
	if !ok {
		return nil, resterrf.CodeErrMap(resterrf.Unauthorized)
	}

	return restContext, nil
}

// Context rest context.
type Context struct {
	gCtx *gin.Context

	// requestID is a unique id for request.
	requestID string

	// loginName is a readable name for user, and is unique in a tenant.
	loginName string

	// bkUsername is a unique name for user, and is unique in all tenants, but it is not readable
	bkUsername string

	// tenantID is a unique id for tenant.
	tenantID string
}

// Deadline implement context.Context.
// nolint: nonamedreturns
func (ctx *Context) Deadline() (deadline time.Time, ok bool) {
	return ctx.gCtx.Deadline()
}

// Done implement context.Context.
func (ctx *Context) Done() <-chan struct{} {
	return ctx.gCtx.Done()
}

// Err implement context.Context.
func (ctx *Context) Err() error {
	return ctx.gCtx.Err()
}

// Value implement context.Context.
func (ctx *Context) Value(key any) any {
	return ctx.gCtx.Value(key)
}

// BindJSON bind json.
func (ctx *Context) BindJSON(body RequestBody) error {
	if err := ctx.gCtx.BindJSON(body); err != nil {
		return err
	}

	// auto convert some fields in request body.
	body.AutoConvert()

	return body.Validate()
}

// ParseFileForm bind file form.
func (ctx *Context) ParseFileForm(body RequestBody) (*multipart.FileHeader, error) {
	metaData := ctx.gCtx.PostForm("metadata")
	if err := json.Unmarshal([]byte(metaData), body); err != nil {
		return nil, fmt.Errorf("failed to parse metadata(%s), err(%v)", metaData, err)
	}

	file, err := ctx.gCtx.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("failed to parse file, err(%v)", err)
	}

	return file, nil
}

// Param parse the param from url.
func (ctx *Context) Param(key string) string {
	return ctx.gCtx.Param(key)
}

// GetRequestHeader get a http request from rest-context.
func (ctx *Context) GetRequestHeader(headerKey string) string {
	return ctx.gCtx.Request.Header.Get(headerKey)
}

// Request get a http request from rest-context.
func (ctx *Context) Request() *http.Request {
	if ctx.gCtx.Request == nil {
		return nil
	}
	req := *ctx.gCtx.Request

	return &req
}

// GetCookie get a cookie from rest-context.
func (ctx *Context) GetCookie(name string) (string, error) {
	cookieValue, err := ctx.gCtx.Cookie(name)
	if err != nil {
		return "", err
	}

	return cookieValue, nil
}

// AbortWithJSONError provides handler process failing response.
func (ctx *Context) AbortWithJSONError(code resterrf.Code, errs []error) {
	result := Response{
		Code: code,
		Error: &Error{
			System:  runtime.System,
			Message: resterrf.CodeErrMap(code).Error(),
			Details: nil,
		},
		RequestID: ctx.RequestID(),
	}

	for _, unwrapErr := range errs {
		result.Error.Details = append(result.Error.Details, Detail{
			Message: unwrapErr.Error(),
		})
	}

	ctx.gCtx.AbortWithStatusJSON(code.HttpStatusCode(), result)
}

// AbortWithJSONPermDenied provides handler process permission denied response.
func (ctx *Context) AbortWithJSONPermDenied(code resterrf.Code, _ []error) {
	result := Response{
		Code: code,
		Permission: &Permission{
			System:     runtime.System,
			SystemName: runtime.SystemName,
			Actions:    nil,
		},
		RequestID: ctx.RequestID(),
	}

	// TODO: 参考 errf.ErrUnwrap 的写法实现 permission 的解析。

	ctx.gCtx.AbortWithStatusJSON(code.HttpStatusCode(), result)
}

// APIResponse provides handler process successfully and make a normal response.
func (ctx *Context) APIResponse(data interface{}) {
	result := Response{
		Code:       0,
		Message:    "OK",
		RequestID:  ctx.RequestID(),
		Data:       data,
		Error:      nil,
		Permission: nil,
	}

	ctx.gCtx.JSON(http.StatusOK, result)
}

// LoginName get login name.
func (ctx *Context) LoginName() string {
	return ctx.loginName
}

// SetLoginName get login name.
func (ctx *Context) SetLoginName(loginName string) {
	ctx.loginName = loginName
}

// BKUsername get bk username.
func (ctx *Context) BKUsername() string {
	return ctx.bkUsername
}

// SetBKUsername get bk username.
func (ctx *Context) SetBKUsername(bkUsername string) {
	ctx.bkUsername = bkUsername
}

// TenantID get tenant id.
func (ctx *Context) TenantID() string {
	return ctx.tenantID
}

// SetTenantID get tenant id.
func (ctx *Context) SetTenantID(tenantID string) {
	ctx.tenantID = tenantID
}

// RequestID get request id.
func (ctx *Context) RequestID() string {
	return ctx.requestID
}

// SetRequestID set request id.
func (ctx *Context) SetRequestID(requestID string) {
	ctx.requestID = requestID
}
