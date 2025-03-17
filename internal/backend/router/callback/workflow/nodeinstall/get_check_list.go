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

// GetCheckList ...
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

	preSetting, customSetting, err := h.nodeDeploymentStorage.GetGseFileProxySetting(gCtx, req.GetToken())
	if err != nil {
		h.logger.Errorf("get gse file proxy setting failed, err: %s", err)
		gCtx.JSON(http.StatusInternalServerError, err)

		return
	}

	conf, err := RenderConfig(DefaultTemplateGseCheckList, preSetting, customSetting)
	if err != nil {
		h.logger.Errorf("render gse file proxy config failed, err: %s", err)
		gCtx.JSON(http.StatusInternalServerError, err)

		return
	}

	gCtx.String(http.StatusOK, conf)

	return
}

const DefaultTemplateGseCheckList = `{
    "disk_requires": "__BK_GSE_DISK_REQUIRES__",
	"port_policies": "__BK_GSE_PORT_POLICIES__",
	"network_policies": "__BK_GSE_NETWORK_POLICIES__",
}`

func mockCheckList() map[string]any {
	return map[string]any{
		"disk_requires": []map[string]any{
			{
				"demand_mb": 300,
				"dir_path":  "/usr/local/gse",
			},
		},
		"port_policies": []map[string]any{
			{
				"port":    8009,
				"network": "tcp",
			},
		},
		"network_policies": []map[string]any{
			{
				"network": "tcp",
				"port":    8006,
				"host":    "10.0.0.1",
			},
		},
	}
}
