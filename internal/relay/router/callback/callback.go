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
	"io"
	"net/http"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/options"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/gin-gonic/gin"
)

const (
	defaultRequestTimeout = 5 * time.Second
)

type handler struct {
	rg     *gin.RouterGroup
	client relayhandler.ICallbackClient
}

func (h *handler) request(gCtx *gin.Context) {
	content, err := io.ReadAll(gCtx.Request.Body)
	if err != nil {
		gCtx.JSON(http.StatusInternalServerError, err)

		return
	}

	// create a new context with timeout
	nCtx, cancel := contextx.WithTimeout(contextx.New(gCtx), defaultRequestTimeout)
	defer cancel()

	logger.G.Biz(nCtx).With("url", gCtx.Request.URL.Path, "content", string(content)).Info("request to callback")
	resp, statusCode, err := h.client.RequestCallback(nCtx, gCtx.Request.URL.Path, content)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("url", gCtx.Request.URL.Path).Error("failed to request to callback")
		gCtx.JSON(statusCode, err)

		return
	}

	logger.G.Biz(nCtx).With("url", gCtx.Request.URL.Path, "content", string(resp)).Info("response from callback")
	gCtx.Data(statusCode, "application/json; charset=utf-8", resp)
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:     rg.Group("/callback"),
		client: capability.Messager,
	}
}

// Load register the callback router.
func Load(rg *gin.RouterGroup, capability *options.Capability) restserver.IMiddlewareChain {
	h := newHandler(rg, capability)

	h.rg.POST("/*path", h.request)

	return restserver.NewMiddlewareChain(h.rg)
}
