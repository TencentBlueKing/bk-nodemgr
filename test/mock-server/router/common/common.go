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

// Package common provides common helper functions for mock server components.
package common

import (
	"net/http"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/gin-gonic/gin"
)

// BindJSON binds JSON request body to the given struct.
func BindJSON[T any](gCtx *gin.Context, req *T) error {
	return gCtx.ShouldBindJSON(req)
}

// BindURI binds URI parameters to the given struct.
func BindURI[T any](gCtx *gin.Context, params *T) error {
	return gCtx.ShouldBindUri(params)
}

// BindQuery binds query parameters to the given struct.
func BindQuery[T any](gCtx *gin.Context, params *T) error {
	return gCtx.ShouldBindQuery(params)
}

// GetHeaderBool gets a header value and converts it to bool.
// Returns the bool value and an error if conversion fails.
// If the header is missing, it returns false with no error.
func GetHeaderBool(gCtx *gin.Context, headerName string) (bool, error) {
	headerValue := gCtx.GetHeader(headerName)
	if headerValue == "" {
		return false, nil
	}

	return conv.StringToBool(headerValue)
}

// RespondJSON sends a JSON response with HTTP 200 status.
func RespondJSON[T any](gCtx *gin.Context, data T) {
	gCtx.JSON(http.StatusOK, data)
}
