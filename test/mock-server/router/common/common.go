/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package common provides common helper functions for mock server components.
package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BindJSON binds JSON request body to the given struct.
func BindJSON[T any](gCtx *gin.Context, req *T) error {
	return gCtx.ShouldBindJSON(req)
}

// RespondJSON sends a JSON response with HTTP 200 status.
func RespondJSON[T any](gCtx *gin.Context, data T) {
	gCtx.JSON(http.StatusOK, data)
}
