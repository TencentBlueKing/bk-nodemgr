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
	gCtx      *gin.Context
	RequestID string `json:"request_id"`

	// LoginName is a readable name for user, and is unique in a tenant.
	LoginName string `json:"login_name"`

	// BKUsername is a unique name for user, and is unique in all tenants, but it is not readable
	BKUsername string `json:"bk_username"`
	TenantID   string `json:"tenant_id"`
}

// Deadline implement context.Context.
// nolint: nonamedreturns
func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.gCtx.Deadline()
}

// Done implement context.Context.
func (c *Context) Done() <-chan struct{} {
	return c.gCtx.Done()
}

// Err implement context.Context.
func (c *Context) Err() error {
	return c.gCtx.Err()
}

// Value implement context.Context.
func (c *Context) Value(key any) any {
	return c.gCtx.Value(key)
}

// BindJSON bind json.
func (c *Context) BindJSON(body RequestBody) error {
	if err := c.gCtx.BindJSON(body); err != nil {
		return err
	}

	// auto convert some fields in request body.
	body.AutoConvert()

	return body.Validate()
}

// ParseFileForm bind file form.
func (c *Context) ParseFileForm(body RequestBody) (*multipart.FileHeader, error) {
	metaData := c.gCtx.PostForm("metadata")
	if err := json.Unmarshal([]byte(metaData), body); err != nil {
		return nil, fmt.Errorf("failed to parse metadata(%s), err(%v)", metaData, err)
	}

	file, err := c.gCtx.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("failed to parse file, err(%v)", err)
	}

	return file, nil
}

// Param parse the param from url.
func (c *Context) Param(key string) string {
	return c.gCtx.Param(key)
}

// GetRequestHeader get a http request from rest-context.
func (c *Context) GetRequestHeader(headerKey string) string {
	return c.gCtx.Request.Header.Get(headerKey)
}

// Request get a http request from rest-context.
func (c *Context) Request() *http.Request {
	if c.gCtx.Request == nil {
		return nil
	}
	req := *c.gCtx.Request

	return &req
}

// GetCookie get a cookie from rest-context.
func (c *Context) GetCookie(name string) (string, error) {
	cookieValue, err := c.gCtx.Cookie(name)
	if err != nil {
		return "", err
	}

	return cookieValue, nil
}

// AbortWithJSONError provides handler process failing response.
func (c *Context) AbortWithJSONError(code resterrf.Code, errs []error) {
	result := Response{
		Code: code,
		Error: &Error{
			System:  runtime.System,
			Message: resterrf.CodeErrMap(code).Error(),
			Details: nil,
		},
		RequestID: c.RequestID,
	}

	for _, unwrapErr := range errs {
		result.Error.Details = append(result.Error.Details, Detail{
			Message: unwrapErr.Error(),
		})
	}

	c.gCtx.AbortWithStatusJSON(code.HttpStatusCode(), result)
}

// AbortWithJSONPermDenied provides handler process permission denied response.
func (c *Context) AbortWithJSONPermDenied(code resterrf.Code, errs []error) {
	result := Response{
		Code: code,
		Permission: &Permission{
			System:     runtime.System,
			SystemName: runtime.SystemName,
			Actions:    nil,
		},
		RequestID: c.RequestID,
	}

	// TODO: 参考 errf.ErrUnwrap 的写法实现 permission 的解析。

	c.gCtx.AbortWithStatusJSON(code.HttpStatusCode(), result)
}

// APIResponse provides handler process successfully and make a normal response.
func (c *Context) APIResponse(data interface{}) {
	result := Response{
		Code:       0,
		Message:    "OK",
		RequestID:  c.RequestID,
		Data:       data,
		Error:      nil,
		Permission: nil,
	}

	c.gCtx.JSON(http.StatusOK, result)
}
