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

	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/callback"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/gin-gonic/gin"
)

func (h *handler) ReportStatus(gCtx *gin.Context) {
	req := new(protoBackend.ReportStatusReq)
	if err := gCtx.BindJSON(req); err != nil {
		h.logger.Errorf("report status failed, err: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Errorf("report status failed, err: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	info, err := h.GetInfo(gCtx, req.GetToken())
	if err != nil {
		h.logger.Errorf("token is invalid, err: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	// `TODO`: 目前tools上报的状态只有两个: `success` 和 `failed`, 但是后续可能会有更多状态, 所以这里先不做校验
	if err := h.UpdateOperInstActionStatus(gCtx, info.OperInstID, info.BlockingActionName,
		action.StateFailed); err != nil {
		h.logger.Errorf("update action status failed: %s", err)
		gCtx.JSON(http.StatusInternalServerError, err)

		return
	}
}

// InstallStatus this is the report status.
type InstallStatus struct {
	Status string `json:"status"`
}
