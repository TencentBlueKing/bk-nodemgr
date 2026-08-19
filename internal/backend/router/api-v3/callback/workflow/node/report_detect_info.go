/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package node

import (
	"net/http"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoCallback "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/callback"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

// ReportDetectInfo report detect info.
func (h *handler) ReportDetectInfo(gCtx *gin.Context) {
	nCtx := contextx.New(gCtx)

	req := new(protoCallback.ReportDetectInfoReq)
	if err := gCtx.BindJSON(req); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to report detect info, failed to decode request")
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	if err := req.Validate(); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to report detect info, failed to validate request")
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	logger.G.Biz(nCtx).With("oper-inst-id", req.GetOperInstId(), "action", req.GetActionName()).Info("report detect info")

	info, err := conv.StructToMap(types.PDDetectInfo{
		OsType:  req.GetOsType(),
		CPUArch: req.GetCpuArch(),
		RunDir:  req.GetConnectionDir(),
		ErrMsg:  req.GetErrMsg(),
	})
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to report detect info, failed to convert struct to map")
		gCtx.JSON(http.StatusInternalServerError, err)

		return
	}

	if err := h.UpsertActionInstancePrivateData(nCtx, req.GetOperInstId(), req.GetActionName(), map[string]any{
		types.PDKeyReportDetectInfo: info,
	}); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to report detect info, failed to update action private data")
		gCtx.JSON(http.StatusInternalServerError, err)
	}

	gCtx.JSON(http.StatusOK, nil)
}
