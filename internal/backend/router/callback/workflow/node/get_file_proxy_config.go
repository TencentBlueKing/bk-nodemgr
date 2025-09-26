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

	protoCallback "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/callback"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

// GetFileProxyConfig get file proxy config.
func (h *handler) GetFileProxyConfig(gCtx *gin.Context) {
	req := new(protoCallback.GetFileProxyConfReq)
	if err := gCtx.BindJSON(req); err != nil {
		h.logger.Errorf("get gse file proxy config failed: %s", err)
		gCtx.JSON(http.StatusBadRequest, fmt.Errorf("get gse file proxy config failed: %w", err))

		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Errorf("get gse file proxy config failed: %s", err)
		gCtx.JSON(http.StatusBadRequest, fmt.Errorf("get gse file proxy config failed: %w", err))

		return
	}

	nodeConf, err := h.GetNodeDeploymentNodeConf(gCtx, req.GetToken())
	if err != nil {
		h.logger.Errorf("get gse file proxy setting failed: %s", err)
		gCtx.JSON(http.StatusInternalServerError, fmt.Errorf("get gse file proxy setting failed: %w", err))

		return
	}

	conf, err := RenderConfig(Template{
		UniqueKey: types.ConfigKeyFile,
		Content:   nodeConf.ConfigTemplate[types.ConfigKeyFile],
	}, nodeConf)
	if err != nil {
		h.logger.Errorf("render gse file proxy config failed: %s", err)
		gCtx.JSON(http.StatusInternalServerError, fmt.Errorf("render gse file proxy config failed: %w", err))

		return
	}

	gCtx.IndentedJSON(http.StatusOK, conf)

	return
}
