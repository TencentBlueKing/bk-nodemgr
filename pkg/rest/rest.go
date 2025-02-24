/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package rest xxx
package rest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime"
	"github.com/gin-gonic/gin"
)

// Response standard response.
type Response struct {
	Code       errf.Code   `json:"code"`
	Message    string      `json:"message,omitempty"`
	RequestID  string      `json:"request_id"`
	Data       interface{} `json:"data,omitempty"`
	Error      *Error      `json:"error,omitempty"`
	Permission *Permission `json:"permission,omitempty"`
}

// Error defines the error struct.
type Error struct {
	System  string   `json:"system"`
	Message string   `json:"message"`
	Details []Detail `json:"details,omitempty"`
}

// Detail defines the error detail.
type Detail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Permission defines the permission struct.
type Permission struct {
	System     string   `json:"system"`
	SystemName string   `json:"system_name"`
	Actions    []Action `json:"actions"`
}

// Action defines the action struct.
type Action struct {
	// TODO: 支持 IAM 权限
}

// HandlerFunc defines the router handler.
type HandlerFunc func(*Context) (interface{}, error)

// StreamHandlerFunc defines the stream handler.
type StreamHandlerFunc func(*Context)

// AbortWithJSONError provides handler process failing response.
func (c *Context) AbortWithJSONError(code errf.Code, errs []error) {
	result := Response{
		Code: code,
		Error: &Error{
			System:  runtime.System,
			Message: errf.CodeErrMap(code).Error(),
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
func (c *Context) AbortWithJSONPermDenied(code errf.Code, errs []error) {
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

// restContextKey was used to store the restContext in gin.Context.
const restContextKey = "rest_context"

// InitRestContext initializes a new rest context.
func InitRestContext(gCtx *gin.Context) *Context {
	restContext := &Context{
		gCtx:      gCtx,
		RequestID: header.RIDGetter(gCtx.Request, true),
		Username:  gCtx.GetHeader(header.UserKey),
		TenantID:  gCtx.GetHeader(header.TenantIDKey),
	}

	gCtx.Set(restContextKey, restContext)

	// note: for thread safety you need to reset it here.
	ctx := context.WithValue(gCtx.Request.Context(), header.RIDKey, restContext.RequestID)
	restContext.gCtx.Request = restContext.gCtx.Request.WithContext(ctx)

	return restContext
}

// GetRestContext only when user has authenticated, otherwise, return ErrorUnauthorized.
func GetRestContext(c *gin.Context) (*Context, error) {
	ctxObj, ok := c.Get(restContextKey)
	if !ok {
		return nil, errf.CodeErrMap(errf.Unauthorized)
	}

	restContext, ok := ctxObj.(*Context)
	if !ok {
		return nil, errf.CodeErrMap(errf.Unauthorized)
	}

	return restContext, nil
}

// RestHandlerFunc rest handler.
func RestHandlerFunc(handler HandlerFunc) gin.HandlerFunc { // nolint
	return func(gCtx *gin.Context) {
		rCtx, err := GetRestContext(gCtx)
		if err != nil {
			InitRestContext(gCtx).AbortWithJSONError(errf.Unauthorized, nil)

			return
		}
		result, err := handler(rCtx)

		code, unwrapErrs := errf.ErrUnwrap(err)
		switch code {
		case errf.OK:
			rCtx.APIResponse(result)
		case errf.PermissionDenied:
			rCtx.AbortWithJSONPermDenied(code, unwrapErrs)
		default:
			rCtx.AbortWithJSONError(code, unwrapErrs)
		}
	}
}

// StreamHandler stream handler.
func StreamHandler(handler StreamHandlerFunc) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		rCtx, err := GetRestContext(gCtx)
		if err != nil {
			InitRestContext(gCtx).AbortWithJSONError(errf.Unauthorized, nil)

			return
		}
		handler(rCtx)
	}
}

// FileHandlerFunc define file handler.
type FileHandlerFunc func(*Context) (*FileResponse, error)

// FileResponse file response.
type FileResponse struct {
	// Data defines file data.
	Data io.Reader

	// Size defines file size.
	Size int64

	// FilePath defines file path.
	FilePath string

	// FileName defines file name.
	FileName string

	// ContentType defines content type.
	ContentType string

	// Headers defines custom headers.
	Headers map[string]string
}

// FileHandler file handler.
func FileHandler(handler FileHandlerFunc) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		rCtx, err := GetRestContext(gCtx)
		if err != nil {
			InitRestContext(gCtx).AbortWithJSONError(errf.Unauthorized, nil)

			return
		}

		fileResp, err := handler(rCtx)
		code, unwrapErrs := errf.ErrUnwrap(err)
		switch code {
		case errf.OK:
			if fileResp == nil || fileResp.Data == nil {
				rCtx.AbortWithJSONError(errf.InvalidFileResource, nil)

				return
			}

			if fileResp.ContentType == "" {
				fileResp.ContentType = getDefaultContentType(fileResp.FilePath)
			}

			if fileResp.FileName == "" {
				fileResp.FileName = filepath.Base(fileResp.FilePath)
			}

			setFileHeaders(gCtx, fileResp)

			gCtx.DataFromReader(http.StatusOK, fileResp.Size, fileResp.ContentType, fileResp.Data, fileResp.Headers)
		case errf.PermissionDenied:
			rCtx.AbortWithJSONPermDenied(code, unwrapErrs)
		default:
			rCtx.AbortWithJSONError(code, unwrapErrs)
		}
	}
}

func setFileHeaders(c *gin.Context, resp *FileResponse) {
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", resp.FileName))
	c.Header("Content-Type", resp.ContentType)

	for key, value := range resp.Headers {
		c.Header(key, value)
	}
}

func getDefaultContentType(filePath string) string {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".doc", ".docx":
		return "application/msword"
	case ".xls", ".xlsx":
		return "application/vnd.ms-excel"
	case ".zip":
		return "application/zip"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	default:
		return "application/octet-stream"
	}
}
