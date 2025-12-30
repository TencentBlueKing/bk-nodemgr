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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoCallback "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/callback"
	"github.com/gin-gonic/gin"
)

// RelayReportDetectResult report relay detect info.
func (h *handler) RelayReportDetectResult(gCtx *gin.Context) {
	nCtx := contextx.New(gCtx)

	req := new(protoCallback.ReportDetectResultReq)
	if err := gCtx.BindJSON(req); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to report detect result, failed to decode request")
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	if err := req.Validate(); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to report detect result, failed to validate request")
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	logger.G.Biz(nCtx).With("oper-inst-id", req.GetOperInstId(), "action", req.GetActionName()).Info("report detect info")

	detectInfo := map[string]string{
		relayconstant.DetectResultCPUArchKey:       req.GetCpuArch(),
		relayconstant.DetectResultOsTypeKey:        req.GetOsType(),
		relayconstant.DetectResultConnectionDirKey: req.GetConnectionDir(),
		relayconstant.DetectResultErrMsgKey:        req.GetErrMsg(),
	}

	dataMap := map[string]any{
		relayconstant.DetectResultKey: detectInfo,
	}

	if err := h.UpsertActionInstancePrivateData(nCtx,
		req.GetOperInstId(), req.GetActionName(), dataMap); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to report detect result, failed to update action private data")
		gCtx.JSON(http.StatusInternalServerError, err)
	}

	gCtx.JSON(http.StatusOK, nil)
}
