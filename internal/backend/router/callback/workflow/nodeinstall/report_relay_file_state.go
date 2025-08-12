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

const relayStateKey = "relay_state"

func (h *handler) RelayReportFileState(gCtx *gin.Context) {
	req := new(proto.ReportFileStateReq)
	if err := gCtx.BindJSON(req); err != nil {
		h.logger.Errorf("report relay file state failed: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Errorf("report relay file state failed: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	h.logger.Infof("report relay file state. instance(%s), action(%s) ",
		req.GetOperInstId(), req.GetActionName())

	fileStateMap := make(map[string]string)
	for _, fileState := range req.GetFileState() {
		fileStateMap[fileState.GetFileName()] = fileState.GetFileStatus()
	}

	dataMap := make(map[string]any)
	dataMap[relayStateKey] = fileStateMap

	if err := h.IDomainNodeInstall.UpsertActionInstancePrivateData(gCtx,
		req.GetOperInstId(), req.GetActionName(), dataMap); err != nil {
		h.logger.Errorf("update action private failed: %s", err)
		gCtx.JSON(http.StatusInternalServerError, err)
	}

	gCtx.JSON(http.StatusOK, nil)
}
