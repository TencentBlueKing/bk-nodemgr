/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package nodeinstall ...
package nodeinstall

import (
	"net/http"

	proto "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/callback"
	"github.com/gin-gonic/gin"
)

// GetCheckList get gse file proxy config.
func (h *handler) GetCheckList(gCtx *gin.Context) {
	req := new(proto.GetCheckListReq)
	if err := gCtx.BindJSON(req); err != nil {
		h.logger.Errorf("get gse file proxy config failed, err: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Errorf("get gse file proxy config failed, err: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	nodeConf, err := h.nodeDeploymentStorage.GetNodeConf(gCtx, req.GetToken())
	if err != nil {
		h.logger.Errorf("get gse file proxy setting failed, err: %s", err)
		gCtx.JSON(http.StatusInternalServerError, err)

		return
	}

	conf, err := RenderConfig(DefaultTemplateNodeInstallerCheckList, nodeConf)
	if err != nil {
		h.logger.Errorf("render gse file proxy config failed, err: %s", err)
		gCtx.JSON(http.StatusInternalServerError, err)

		return
	}

	gCtx.String(http.StatusOK, conf)

	return
}

// DefaultTemplateNodeInstallerCheckList  gse file proxy.
const DefaultTemplateNodeInstallerCheckList = `{
    "disk_requires": __BK_NODE_INSTALLER_DISK_REQUIRES__,
	"port_policies": __BK_NODE_INSTALLER_PORT_POLICIES__,
	"network_policies": __BK_NODE_INSTALLER_NETWORK_POLICIES__,
}`
