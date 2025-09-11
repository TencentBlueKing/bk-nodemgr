/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package node ...
package node

import (
	"fmt"
	"net/http"

	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/callback"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

// GetAgentConfig ...
func (h *handler) GetAgentConfig(gCtx *gin.Context) {
	req := new(protoBackend.GetAgentConfReq)
	if err := gCtx.BindJSON(req); err != nil {
		h.logger.Errorf("get gse agent config failed, err: %v", err)
		gCtx.IndentedJSON(http.StatusBadRequest, err)

		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Errorf("get gse agent config failed, err: %v", err)
		gCtx.IndentedJSON(http.StatusBadRequest, err)

		return
	}

	nodeConf, err := h.GetNodeDeploymentNodeConf(gCtx, req.GetToken())
	if err != nil {
		h.logger.Errorf("get gse agent setting failed, err: %v", err)
		gCtx.IndentedJSON(http.StatusInternalServerError, err)

		return
	}

	configTemplate, ok := nodeConf.ConfigTemplate[types.ConfigKeyAgent]
	if !ok {
		err := fmt.Errorf("config template not found: %s", types.ConfigKeyAgent)

		h.logger.Errorf("get gse agent config failed, err: %v", err)
		gCtx.IndentedJSON(http.StatusInternalServerError, err)

		return
	}

	conf, err := RenderConfig(Template{
		UniqueKey: types.ConfigKeyAgent,
		Content:   configTemplate,
	}, nodeConf)
	if err != nil {
		h.logger.Errorf("render gse agent config failed, err: %v", err)
		gCtx.IndentedJSON(http.StatusInternalServerError, err)

		return
	}

	gCtx.IndentedJSON(http.StatusOK, conf)

	return
}
