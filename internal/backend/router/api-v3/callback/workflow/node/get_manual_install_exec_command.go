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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

// GetManualInstallExecCommand get manual install exec command.
func (h *handler) GetManualInstallExecCommand(gCtx *gin.Context) {
	nCtx := contextx.New(gCtx)

	req := new(protoCallback.GetManualInstallExecCommandReq)
	if err := gCtx.BindJSON(req); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to get manual install exec command, failed to decode request")
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	if err := req.Validate(); err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to get manual install exec command, failed to validate request")
		gCtx.JSON(http.StatusBadRequest, err)

		return
	}

	logger.G.Biz(nCtx).With("oper-inst-id", req.GetOperInstId(), "action", req.GetActionName()).Info("get manual install exec command")

	data, err := h.GetActionInstancePrivateData(nCtx, req.GetOperInstId(), req.GetActionName())
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to get manual install exec command, failed to get action instance private data")
		gCtx.JSON(http.StatusInternalServerError, err)

		return
	}

	// installer path.
	installerPathRaw, ok := data[types.PDKeyManualInstallInstallerPath]
	if !ok {
		gCtx.JSON(http.StatusOK, "")
		return
	}
	installerPath, ok := installerPathRaw.(string)
	if !ok {
		gCtx.JSON(http.StatusInternalServerError, "manual install installer path is invalid")
		return
	}

	// command.
	commandRaw, ok := data[types.PDKeyManualInstallExecCommand]
	if !ok {
		gCtx.JSON(http.StatusOK, "")
		return
	}
	command, ok := commandRaw.(string)
	if !ok {
		gCtx.JSON(http.StatusInternalServerError, "manual install exec command is invalid")
		return
	}

	gCtx.String(http.StatusOK, installerPath+"\n"+command)
}
