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
	"fmt"
	"mime/multipart"
	"net/http"

	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	"github.com/gin-gonic/gin"
)

// IRequest rest server request interface.
type IRequest interface {
	// GContext get rest context.
	GContext() *gin.Context

	// BindJSON bind json parameters from request body.
	BindJSON(body RequestBody) error

	// ParseFileForm bind file form from request body.
	ParseFileForm(body RequestBody) (*multipart.FileHeader, error)

	// GetRequestHeader get a http request from rest-context.
	GetRequestHeader(headerKey string) string

	// GetRequest get a http request from rest-context.
	GetRequest() *http.Request

	// GetCookie get a cookie from rest-context.
	GetCookie(name string) (string, error)

	// AbortWithJSONError provides handler process failing response.
	AbortWithJSONError(code resterrf.Code, errs []error)

	// AbortWithJSONPermDenied provides handler process permission denied response.
	AbortWithJSONPermDenied(code resterrf.Code, errs []error)

	// APIResponse provides handler process successfully and make a normal response.
	APIResponse(data interface{})

	// Data provides request data.
	Data() IRequestData
}

type IRequestData interface {
	// GetLoginName get login name.
	GetLoginName() string

	// SetLoginName get login name.
	SetLoginName(loginName string)

	// GetBKUsername get bk username.
	GetBKUsername() string

	// SetBKUsername get bk username.
	SetBKUsername(bkUsername string)

	// GetTenantID get tenant id.
	GetTenantID() string

	// SetTenantID get tenant id.
	SetTenantID(tenantID string)

	// GetRequestID get request id.
	GetRequestID() string

	// SetRequestID set request id.
	SetRequestID(requestID string)
}

// NewRequest creates a new rest request.
func NewRequest(gCtx *gin.Context) *Request {
	return &Request{
		gCtx: gCtx,
		data: &RequestData{},
	}
}

// Request implements the rest request.
type Request struct {
	gCtx *gin.Context

	data *RequestData
}

// GContext get rest context.
func (r *Request) GContext() *gin.Context {
	return r.gCtx
}

// BindJSON bind json.
func (r *Request) BindJSON(body RequestBody) error {
	if err := r.gCtx.BindJSON(body); err != nil {
		return err
	}

	// auto convert some fields in request body.
	body.AutoConvert()

	return body.Validate()
}

// ParseFileForm bind file form.
func (r *Request) ParseFileForm(body RequestBody) (*multipart.FileHeader, error) {
	metaData := r.gCtx.PostForm("metadata")
	if err := json.Unmarshal([]byte(metaData), body); err != nil {
		return nil, fmt.Errorf("failed to parse metadata(%s), err(%v)", metaData, err)
	}

	file, err := r.gCtx.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("failed to parse file, err(%v)", err)
	}

	return file, nil
}

// Param parse the param from url.
func (r *Request) Param(key string) string {
	return r.gCtx.Param(key)
}

// GetRequestHeader get a http request from rest-context.
func (r *Request) GetRequestHeader(headerKey string) string {
	return r.gCtx.Request.Header.Get(headerKey)
}

// GetRequest get a http request from rest-context.
func (r *Request) GetRequest() *http.Request {
	if r.gCtx.Request == nil {
		return nil
	}
	req := *r.gCtx.Request

	return &req
}

// GetCookie get a cookie from rest-context.
func (r *Request) GetCookie(name string) (string, error) {
	cookieValue, err := r.gCtx.Cookie(name)
	if err != nil {
		return "", err
	}

	return cookieValue, nil
}

// AbortWithJSONError provides handler process failing response.
func (r *Request) AbortWithJSONError(code resterrf.Code, errs []error) {
	result := Response{
		Code: code,
		Error: &Error{
			System:  system.Code,
			Message: resterrf.CodeErrMap(code).Error(),
			Details: nil,
		},
		RequestID: r.data.requestID,
	}

	for _, unwrapErr := range errs {
		result.Error.Details = append(result.Error.Details, Detail{
			Message: unwrapErr.Error(),
		})
	}

	r.gCtx.AbortWithStatusJSON(code.HttpStatusCode(), result)
}

// AbortWithJSONPermDenied provides handler process permission denied response.
func (r *Request) AbortWithJSONPermDenied(code resterrf.Code, _ []error) {
	result := Response{
		Code: code,
		Permission: &Permission{
			System:     system.Code,
			SystemName: system.Name,
			Actions:    nil,
		},
		RequestID: r.data.requestID,
	}

	// TODO: 参考 errf.ErrUnwrap 的写法实现 permission 的解析。

	r.gCtx.AbortWithStatusJSON(code.HttpStatusCode(), result)
}

// APIResponse provides handler process successfully and make a normal response.
func (r *Request) APIResponse(data interface{}) {
	result := Response{
		Code:       0,
		Message:    "OK",
		RequestID:  r.data.requestID,
		Data:       data,
		Error:      nil,
		Permission: nil,
	}

	r.gCtx.JSON(http.StatusOK, result)
}

// Data get request data.
func (r *Request) Data() IRequestData {
	return r.data
}

type RequestData struct {
	loginName  string
	bkUsername string
	tenantID   string
	requestID  string
}

// GetLoginName get login name.
func (r *RequestData) GetLoginName() string {
	return r.loginName
}

// SetLoginName get login name.
func (r *RequestData) SetLoginName(loginName string) {
	r.loginName = loginName
}

// GetBKUsername get bk username.
func (r *RequestData) GetBKUsername() string {
	return r.bkUsername
}

// SetBKUsername get bk username.
func (r *RequestData) SetBKUsername(bkUsername string) {
	r.bkUsername = bkUsername
}

// GetTenantID get tenant id.
func (r *RequestData) GetTenantID() string {
	return r.tenantID
}

// SetTenantID get tenant id.
func (r *RequestData) SetTenantID(tenantID string) {
	r.tenantID = tenantID
}

// GetRequestID get request id.
func (r *RequestData) GetRequestID() string {
	return r.requestID
}

// SetRequestID set request id.
func (r *RequestData) SetRequestID(requestID string) {
	r.requestID = requestID
}
