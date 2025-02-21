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
	"github.com/gin-gonic/gin"
)

// Response standard response.
type Response struct {
	Result    bool        `json:"result"`
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
	Data      interface{} `json:"data"`
}

// HandlerFunc defines the router handler.
type HandlerFunc func(*Context) (interface{}, error)

// StreamHandlerFunc defines the stream handler.
type StreamHandlerFunc func(*Context)

// AbortWithBadRequestError provides handler process failing response.
func (c *Context) AbortWithBadRequestError(err error) {
	result := Response{Code: errf.InvalidParameter, Message: err.Error(), RequestID: c.RequestID}
	c.gCtx.AbortWithStatusJSON(http.StatusBadRequest, result)
}

// AbortWithUnauthorizedError provides auth check failing response.
func (c *Context) AbortWithUnauthorizedError(err error) {
	result := Response{Code: errf.DoAuthorizeFailed, Message: err.Error(), RequestID: c.RequestID}
	c.gCtx.AbortWithStatusJSON(http.StatusUnauthorized, result)
}

// AbortWithWithForbiddenError provides permission denied response.
func (c *Context) AbortWithWithForbiddenError(err error) {
	result := Response{Code: errf.PermissionDenied, Message: err.Error(), RequestID: c.RequestID}
	c.gCtx.AbortWithStatusJSON(http.StatusForbidden, result)
}

// AbortWithJSONError provides handler process failing response.
func (c *Context) AbortWithJSONError(err error) {
	// TODO: support error code
	result := Response{Code: errf.Aborted, Result: false, Message: err.Error(), RequestID: c.RequestID}
	c.gCtx.AbortWithStatusJSON(http.StatusOK, result)
}

// APIResponse provides handler process successfully and make a normal response.
func (c *Context) APIResponse(data interface{}) {
	result := Response{Code: 0, Result: true, Message: "OK", RequestID: c.RequestID, Data: data}
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
		return nil, errf.ErrorUnauthorized
	}

	restContext, ok := ctxObj.(*Context)
	if !ok {
		return nil, errf.ErrorUnauthorized
	}

	return restContext, nil
}

// RestHandlerFunc rest handler.
func RestHandlerFunc(handler HandlerFunc) gin.HandlerFunc { // nolint
	return func(gCtx *gin.Context) {
		rCtx, err := GetRestContext(gCtx)
		if err != nil {
			InitRestContext(gCtx).AbortWithUnauthorizedError(err)

			return
		}
		result, err := handler(rCtx)
		if err != nil {
			rCtx.AbortWithJSONError(err)

			return
		}

		rCtx.APIResponse(result)
	}
}

// STDRestHandlerFunc std rest handler.
func STDRestHandlerFunc(handler HandlerFunc) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		rCtx, err := GetRestContext(gCtx)
		if err != nil {
			InitRestContext(gCtx).AbortWithUnauthorizedError(err)
			return
		}
		result, err := handler(rCtx)
		if err != nil {
			rCtx.AbortWithBadRequestError(err)
			return
		}

		rCtx.APIResponse(result)
	}
}

// StreamHandler stream handler.
func StreamHandler(handler StreamHandlerFunc) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		rCtx, err := GetRestContext(gCtx)
		if err != nil {
			InitRestContext(gCtx).AbortWithUnauthorizedError(err)

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
			InitRestContext(gCtx).AbortWithUnauthorizedError(err)

			return
		}

		fileResp, err := handler(rCtx)
		if err != nil {
			rCtx.AbortWithJSONError(err)
			return
		}

		if fileResp == nil || fileResp.Data == nil {
			rCtx.AbortWithJSONError(fmt.Errorf("invalid file response"))
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
