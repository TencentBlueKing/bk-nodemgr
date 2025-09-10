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
	"net/http"

	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/callback"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

// GetDataProxyConfig get gse data proxy config.
func (h *handler) GetDataProxyConfig(gCtx *gin.Context) {
	req := new(protoBackend.GetDataProxyConfReq)
	if err := gCtx.BindJSON(req); err != nil {
		h.logger.Errorf("get gse data proxy config failed, err: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Errorf("get gse data proxy config failed, err: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	nodeConf, err := h.GetNodeConf(gCtx, req.GetToken())
	if err != nil {
		h.logger.Errorf("get gse data proxy setting failed, err: %s", err)
		gCtx.JSON(http.StatusInternalServerError, err)

		return
	}

	conf, err := RenderConfig(Template{
		UniqueKey: types.ConfigKeyData,
		Content:   nodeConf.ConfigTemplate[types.ConfigKeyData],
	}, nodeConf)
	if err != nil {
		h.logger.Errorf("render gse data proxy config failed, err: %s", err)
		gCtx.JSON(http.StatusInternalServerError, err.Error())

		return
	}

	gCtx.IndentedJSON(http.StatusOK, conf)

	return
}
