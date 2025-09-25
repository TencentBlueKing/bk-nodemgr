/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package node

import (
	"net/http"

	protoCallback "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/callback"
	"github.com/gin-gonic/gin"
)

// ReportData ...
func (h *handler) ReportData(gCtx *gin.Context) {
	req := new(protoCallback.ReportDataReq)
	if err := gCtx.BindJSON(req); err != nil {
		h.logger.Errorf("report data failed: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Errorf("report data failed: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	info, err := h.GetNodeDeploymentInfo(gCtx, req.GetToken())
	if err != nil {
		h.logger.Errorf("token is invalid: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	info.Host.Dynamic.AgentID = req.GetAgentId()

	err = h.UpdateNodeDeploymentInfo(gCtx, req.GetToken(), info)
	if err != nil {
		h.logger.Errorf("update info failed: %s", err)
		gCtx.JSON(http.StatusInternalServerError, err)

		return
	}
	gCtx.JSON(http.StatusOK, nil)

	return
}
