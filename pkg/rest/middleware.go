/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package rest ...
package rest

import (
	"fmt"
	"io"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/gin-gonic/gin"
)

// AuthIdentity verify auth info.
type AuthIdentity interface {
	Verify(rCtx *Context) error
}

// MiddlewareContext verify auth info.
func MiddlewareContext() gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		_ = initRestContext(gCtx)

		gCtx.Next()
	}
}

// MiddlewareAuth verify auth info.
func MiddlewareAuth(identity AuthIdentity) gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		rCtx := loadRestContext(gCtx)

		if err := identity.Verify(rCtx); err != nil {
			rCtx.AbortWithJSONError(errf.Unauthorized, []error{err})

			return
		}
	}

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
