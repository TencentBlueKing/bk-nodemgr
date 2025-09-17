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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoCallback "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/callback"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/gin-gonic/gin"
)

func (h *handler) ReportStatus(gCtx *gin.Context) {
	req := new(protoCallback.ReportStatusReq)
	if err := gCtx.BindJSON(req); err != nil {
		h.logger.Errorf("report status failed: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Errorf("report status failed: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	info, err := h.GetNodeDeploymentInfo(gCtx, req.GetToken())
	if err != nil {
		h.logger.Errorf("token is invalid: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}
	h.logger.Infof("operation instance:%s, action:%s ,report status: %s",
		req.GetOperInstId(), info.BlockingActionName, req.GetStatus())

	ctx := contextx.NewContext(gCtx, map[string]any{})
	if err = h.UpdateOperInstActionStatus(ctx, req.GetOperInstId(), info.BlockingActionName,
		action.State(req.GetStatus())); err != nil {
		h.logger.Errorf("update action status failed: %s", err)
		gCtx.JSON(http.StatusInternalServerError, err)
	}

	gCtx.JSON(http.StatusOK, nil)
}
