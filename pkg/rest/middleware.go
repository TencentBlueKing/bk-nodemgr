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
	"net/http"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/gin-gonic/gin"
)

// MiddlewareContext ...
func MiddlewareContext() gin.HandlerFunc {
	return func(gCtx *gin.Context) {
		rCtx := InitRestContext(gCtx)

		if gCtx.Request.Method == http.MethodOptions {
			gCtx.Next()
			return
		}

		switch {
		case initContextWithJWT(rCtx):
		default:
			rCtx.AbortWithJSONError(errf.Unauthorized, nil)
			return
		}

		gCtx.Next()
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

// initContextWithJWT init context with jwt
func initContextWithJWT(_ *Context) bool {
	// TODO: implement jwt
	return true
}
