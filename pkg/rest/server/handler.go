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
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/gin-gonic/gin"
)

// 32KB
// bufferSize defines the buffer size of stream response.
const bufferSize = 32 * 1024

// HandlerFunc defines the router handler.
type HandlerFunc func(rCtx IContext) (interface{}, error)

// Handler rest handler.
func Handler(handler HandlerFunc) gin.HandlerFunc { // nolint
	return func(gCtx *gin.Context) {
		rCtx, err := GenRestContext(gCtx)
		if err != nil {
			loadRestRequest(gCtx).AbortWithJSONError(resterrf.Unauthorized, nil)

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

// StdHandlerFunc defines the std handler.
type StdHandlerFunc func(IContext) (interface{}, error)

// StdHandler std handler.
func StdHandler(handler HandlerFunc) gin.HandlerFunc { // nolint
	return func(gCtx *gin.Context) {
		rCtx, err := GenRestContext(gCtx)
		if err != nil {
			loadRestRequest(gCtx).AbortWithJSONError(resterrf.Unauthorized, nil)

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
type StreamHandlerFunc func(IContext) (*StreamResponse, error)

// StreamHandler stream handler.
func StreamHandler(handler StreamHandlerFunc) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		rCtx, err := GenRestContext(gCtx)
		if err != nil {
			loadRestRequest(gCtx).AbortWithJSONError(resterrf.Unauthorized, nil)

			return
		}
		streamResp, err := handler(rCtx)
		unwrapCode, unwrapErrs := resterrf.ErrUnwrap(err)

		switch unwrapCode {
		case resterrf.OK:
			if streamResp == nil || streamResp.Data == nil {
				rCtx.AbortWithJSONError(resterrf.InvalidFileResource, nil)
				return
			}

			// handle stream
			if err := handleStream(gCtx, streamResp); err != nil {
				gCtx.Status(http.StatusInternalServerError)
				return
			}

		case resterrf.PermissionDenied:
			rCtx.AbortWithJSONPermDenied(unwrapCode, unwrapErrs)
		default:
			rCtx.AbortWithJSONError(unwrapCode, unwrapErrs)
		}
	}
}

func handleStream(gCtx *gin.Context, streamResp *StreamResponse) error {
	// copy headers
	for key, values := range streamResp.Headers {
		for _, value := range values {
			gCtx.Header(key, value)
		}
	}

	// copy status
	gCtx.Status(streamResp.StatusCode)

	defer func(data io.ReadCloser) {
		_ = data.Close()
	}(streamResp.Data)

	// copy data
	buffer := make([]byte, bufferSize)
	_, err := io.CopyBuffer(gCtx.Writer, streamResp.Data, buffer)
	if err != nil {
		return fmt.Errorf("failed to copy stream data: %w", err)
	}

	return nil
}

// FileHandlerFunc define file handler.
type FileHandlerFunc func(IContext) (*FileResponse, error)

// FileHandler file handler.
func FileHandler(handler FileHandlerFunc) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		rCtx, err := GenRestContext(gCtx)
		if err != nil {
			loadRestRequest(gCtx).AbortWithJSONError(resterrf.Unauthorized, nil)

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

			defer func(data io.ReadCloser) {
				_ = data.Close()
			}(fileResp.Data)

			gCtx.DataFromReader(http.StatusOK, fileResp.Size, fileResp.ContentType.String(), fileResp.Data, fileResp.Headers)
		case resterrf.PermissionDenied:
			rCtx.AbortWithJSONPermDenied(code, unwrapErrs)
		default:
			rCtx.AbortWithJSONError(code, unwrapErrs)
		}
	}
}

// nolint: perfsprint
func setFileHeaders(gCtx *gin.Context, resp *FileResponse) {
	gCtx.Header("Content-Description", "File Transfer")
	gCtx.Header("Content-Transfer-Encoding", "binary")
	gCtx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", resp.FileName))
	gCtx.Header("Content-Type", resp.ContentType.String())

	for key, value := range resp.Headers {
		gCtx.Header(key, value)
	}
}

func getDefaultContentType(filePath string) MIMEType {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".pdf":
		return MIMETypePdf
	case ".doc", ".docx":
		return MIMETypeDoc
	case ".xls", ".xlsx":
		return MIMETypeXls
	case ".zip":
		return MIMETypeZip
	case ".png":
		return MIMETypePng
	case ".jpg", ".jpeg":
		return MIMETypeJpg
	default:
		return MIMETypeBin
	}
}
