/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package constant defines the constant apis.
package constant

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/options"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

// handler ...
type handler struct {
	rg *gin.RouterGroup
}

// newHandler ...
func newHandler(rg *gin.RouterGroup, _ *options.Capability) *handler {
	return &handler{
		// this is a sub router, so we can use some special middleware in it and not affect the father router.
		rg: rg.Group("/constant"),
	}
}

// Load load agent handler.
func Load(rg *gin.RouterGroup, capability *options.Capability) {
	h := newHandler(rg, capability)

	h.rg.POST("/deploy/get", restserver.Handler(h.GetDeployConstant))
}

// GetDeployConstant get default deploy constant values.
func (h *handler) GetDeployConstant(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeConstantDeployGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get default deploy constant, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	generation := types.Generation(req.GetGeneration())
	osType := criteria.OSType(req.GetOsType())

	nodeConf, err := deployconstant.GetNodeDeployConf(generation, osType)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get default deploy constant, node deploy conf not found")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	pluginConf, err := deployconstant.GetPluginDeployConf(generation, osType)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get default deploy constant, plugin deploy conf not found")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	resp := new(protoBackend.NodeConstantDeployGetResp)
	resp.ConvertConstantFromTypes(nodeConf, pluginConf)

	return resp.GetData(), nil
}
