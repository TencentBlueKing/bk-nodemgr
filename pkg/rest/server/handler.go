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
	"net/http"
	"path/filepath"

	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/gin-gonic/gin"
)

// HandlerFunc defines the router handler.
type HandlerFunc func(*Context) (interface{}, error)

// Handler rest handler.
func Handler(handler HandlerFunc) gin.HandlerFunc { // nolint
	return func(gCtx *gin.Context) {
		rCtx, err := GetRestContext(gCtx)
		if err != nil {
			loadRestContext(gCtx).AbortWithJSONError(resterrf.Unauthorized, nil)

			return
		}
		result, err := handler(rCtx)

		code, unwrapErrs := resterrf.ErrUnwrap(err)
		switch code {
		case resterrf.OK:
			rCtx.APIResponse(result)
		case resterrf.PermissionDenied:
			rCtx.AbortWithJSONPermDenied(code, unwrapErrs)
		default:
			rCtx.AbortWithJSONError(code, unwrapErrs)
		}
	}
}

// StreamHandlerFunc defines the stream handler.
type StreamHandlerFunc func(*Context)

// StreamHandler stream handler.
func StreamHandler(handler StreamHandlerFunc) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		rCtx, err := GetRestContext(gCtx)
		if err != nil {
			loadRestContext(gCtx).AbortWithJSONError(resterrf.Unauthorized, nil)

			return
		}
		handler(rCtx)
	}
}

// FileHandlerFunc define file handler.
type FileHandlerFunc func(*Context) (*FileResponse, error)

// FileHandler file handler.
func FileHandler(handler FileHandlerFunc) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		rCtx, err := GetRestContext(gCtx)
		if err != nil {
			loadRestContext(gCtx).AbortWithJSONError(resterrf.Unauthorized, nil)

			return
		}

		fileResp, err := handler(rCtx)
		code, unwrapErrs := resterrf.ErrUnwrap(err)
		switch code {
		case resterrf.OK:
			if fileResp == nil || fileResp.Data == nil {
				rCtx.AbortWithJSONError(resterrf.InvalidFileResource, nil)

				return
			}

			if fileResp.ContentType == "" {
				fileResp.ContentType = getDefaultContentType(fileResp.FilePath)
			}

			if fileResp.FileName == "" {
				fileResp.FileName = filepath.Base(fileResp.FilePath)
			}

			setFileHeaders(gCtx, fileResp)

			defer fileResp.Data.Close()
			gCtx.DataFromReader(http.StatusOK, fileResp.Size, fileResp.ContentType, fileResp.Data, fileResp.Headers)
		case resterrf.PermissionDenied:
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
