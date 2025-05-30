/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package callback handles the callback request.
package callback

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/options"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/gin-gonic/gin"
)

const (
	defaultRequestTimeout = 5 * time.Second
)

type handler struct {
	rg     *gin.RouterGroup
	client relayhandler.CallbackClient
	logger logger.Logger
}

func (h *handler) request(gCtx *gin.Context) {
	content, err := io.ReadAll(gCtx.Request.Body)
	if err != nil {
		gCtx.JSON(http.StatusInternalServerError, err)

		return
	}

	// create a new context with timeout
	rctx, rcancel := context.WithTimeout(gCtx.Request.Context(), defaultRequestTimeout)
	defer rcancel()

	h.logger.InfoCtxf(rctx, "request to callback with url(%s), content(%s)", gCtx.Request.URL.Path, string(content))
	resp, statusCode, err := h.client.RequestCallback(rctx, gCtx.Request.URL.Path, content)
	if err != nil {
		h.logger.ErrorCtxf(rctx, "failed to request to callback, err: %v", err)
		gCtx.JSON(statusCode, err)

		return
	}

	h.logger.InfoCtxf(rctx, "response from callback with url(%s), content(%s)", gCtx.Request.URL.Path, string(resp))
	gCtx.Data(statusCode, "application/json; charset=utf-8", resp)
}

func newHandler(rg *gin.RouterGroup, cap *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:     rg.Group("/callback"),
		client: cap.Messager,
		logger: cap.Logger,
	}
}

// Load register the callback router.
func Load(rg *gin.RouterGroup, cap *options.Capability) {
	h := newHandler(rg, cap)

	h.rg.POST("/*path", h.request)
}
