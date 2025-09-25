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

	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/relayconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoCallback "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/callback"
	"github.com/gin-gonic/gin"
)

func (h *handler) RelayReportInstallResult(gCtx *gin.Context) {
	req := new(protoCallback.ReportInstallResultReq)
	if err := gCtx.BindJSON(req); err != nil {
		h.logger.Errorf("report install by ssh info failed: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Errorf("report install by ssh info failed: %s", err)
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	h.logger.Infof("report detect info. instance(%s), action(%s) ",
		req.GetOperInstId(), req.GetActionName())

	dataMap := map[string]any{
		relayconstant.InstallResultKey: map[string]string{
			relayconstant.InstallResultOutStrKey: req.GetOutStr(),
			relayconstant.InstallResultErrMsgKey: req.GetErrMsg(),
		},
	}

	nCtx := contextx.New(gCtx)
	if err := h.UpsertActionInstancePrivateData(nCtx,
		req.GetOperInstId(), req.GetActionName(), dataMap); err != nil {
		h.logger.Errorf("update action private failed: %s", err)
		gCtx.JSON(http.StatusInternalServerError, err)
	}

	gCtx.JSON(http.StatusOK, nil)
}
