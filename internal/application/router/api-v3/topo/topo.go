/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package topo provides the topo API handler.
package topo

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	proto "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backend"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

type handler struct {
	rg             *gin.RouterGroup
	backendHandler backend.Handler
	logger         logger.Logger
}

func newHandler(rg *gin.RouterGroup, capability *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg:             rg.Group("/topo"),
		backendHandler: capability.BackendHandler,
		logger:         capability.Logger,
	}
}

// Load loads topo handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	// business apis.
	h.rg.POST("/business/list", rest.RestHandlerFunc(h.ListBusiness))

	// host apis.
	h.rg.POST("/host/list", rest.RestHandlerFunc(h.ListHost))

	// networkarea apis.
	h.rg.POST("/networkarea/list", rest.RestHandlerFunc(h.ListNetworkArea))
	h.rg.POST("/networkarea/get", rest.RestHandlerFunc(h.GetNetworkArea))

	// networkunit apis.
	h.rg.POST("/networkunit/list", rest.RestHandlerFunc(h.ListNetworkUnit))
	h.rg.POST("/networkunit/get", rest.RestHandlerFunc(h.GetNetworkUnit))
}

func validateTopoPage(reqPage *proto.Page) error {
	if reqPage == nil {
		return nil
	}

	if reqPage.GetOffset() < 0 {
		return errors.New("\"page.offset\" field must be >= 0")
	}

	if reqPage.GetLimit() < 0 {
		return errors.New("\"page.limit\" field must be >= 0")
	}

	return nil
}

// generatePage generates list page.
func generatePage(reqPage *proto.Page, maxLimit int) types.Page {
	page := types.Page{}
	if reqPage != nil {
		page.Offset = int(reqPage.GetOffset())
		page.Limit = int(reqPage.GetLimit())
	}

	if page.Offset < 0 {
		page.Offset = 0
	}

	if page.Limit < 0 {
		page.Limit = 0
	}

	if page.Limit > maxLimit {
		page.Limit = maxLimit
	}

	return page
}
