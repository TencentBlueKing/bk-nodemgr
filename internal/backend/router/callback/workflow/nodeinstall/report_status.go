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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/gin-gonic/gin"
)

func (h *handler) ReportStatus(gCtx *gin.Context) {
	req := new(proto.ReportStatusReq)
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
	h.logger.Infof("operation instance:%s, action:%s ,report status: %s",
		req.GetOperInstId(), info.BlockingActionName, req.GetStatus())

	if err := h.UpdateOperInstActionStatus(gCtx, req.GetOperInstId(), info.BlockingActionName,
		action.State(req.GetStatus())); err != nil {
		h.logger.Errorf("update action status failed: %s", err)
		gCtx.JSON(http.StatusInternalServerError, err)
	}
}

// InstallStatus this is the report status.
type InstallStatus struct {
	Status string `json:"status"`
}
