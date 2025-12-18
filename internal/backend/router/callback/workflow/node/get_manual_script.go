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
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
)

func (h *handler) GetManualScript(gCtx *gin.Context) {
	nCtx := contextx.New(gCtx)

	// get os-type.
	osTypeStr := gCtx.Param(pathGetKeyOsType)
	osType, err := platform.NormalizeOS(osTypeStr)
	if err != nil {
		logger.G.Biz(nCtx).With("os-type-str", osTypeStr).Error("failed to get manual script, os-type is invalid")
		gCtx.JSON(http.StatusBadRequest, fmt.Errorf("get manual script failed: os_type in path parameters is invalid"))

		return
	}

	// get operation-instance-id.
	operInstID := gCtx.Param(pathGetKeyOperationInstanceID)
	if operInstID == "" {
		logger.G.Biz(nCtx).With("os-type", osType).Error("failed to get manual script, operation-instance-id is empty")
		gCtx.JSON(http.StatusBadRequest, fmt.Errorf("get manual script failed: operation_instance_id in path parameters is required"))

		return
	}

	// get action name.
	actionName := gCtx.Query(queryGetKeyActionName)
	if actionName == "" {
		logger.G.Biz(nCtx).With("os-type", osType).Error("failed to get manual script, action-name is empty")
		gCtx.JSON(http.StatusBadRequest, fmt.Errorf("get manual script failed: action_name in query parameters is required"))

		return
	}

	privateData, err := h.GetActionInstancePrivateData(nCtx, operInstID, actionName)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to get manual script, failed to get action instance private data")
		gCtx.JSON(http.StatusInternalServerError, err)

		return
	}

	deployConstant, err := deployconstant.GetNodeDeployConf(types.Generation2, osType)
	if err != nil {
		logger.G.Biz(nCtx).With("os-type", osType).Error("failed to get manual script, get node deploy conf failed", err)
		gCtx.JSON(http.StatusBadRequest, fmt.Errorf("get manual script failed: get node deploy conf failed"))

		return
	}

	// TODO: get file from cache, instead of get file from file system every time.
	file, err := local.NewLocalFile(deployConstant.ManualScriptPath)
	if err != nil {
		logger.G.Biz(nCtx).With("os-type", osType, "path", deployConstant.ManualScriptPath).
			Error("failed to get manual script, get file failed", err)
		gCtx.JSON(http.StatusBadRequest, fmt.Errorf("get manual script failed: get file failed"))

		return
	}

	reader, err := file.Content(nCtx)
	if err != nil {
		logger.G.Biz(nCtx).With("os-type", osType, "path", deployConstant.ManualScriptPath).
			Error("failed to get manual script, get script content failed", err)
		gCtx.JSON(http.StatusBadRequest, fmt.Errorf("get manual script failed: get script content failed"))

		return
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		logger.G.Biz(nCtx).With("os-type", osType, "path", deployConstant.ManualScriptPath).
			Error("failed to get manual script, read script content failed", err)
		gCtx.JSON(http.StatusBadRequest, fmt.Errorf("get manual script failed: read script content failed"))

		return
	}

	result, err := manualScriptRender(string(content), operInstID, privateData)
	if err != nil {
		logger.G.Biz(nCtx).With("os-type", osType, "path", deployConstant.ManualScriptPath, "private-data", privateData).
			Error("failed to get manual script, render script content failed", err)
		gCtx.JSON(http.StatusBadRequest, fmt.Errorf("get manual script failed: render script content failed"))

		return
	}

	gCtx.String(http.StatusOK, result)
}

func manualScriptRender(content string, operInstID string, privateData map[string]any) (string, error) {
	// callback address.
	raw, ok := privateData[types.PDKeyManualInstallCallbackAddress]
	if !ok {
		return "", fmt.Errorf("private data does not contain callback address")
	}
	callbackAddress, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("private data does not contain valid callback address")
	}

	// download address.
	raw, ok = privateData[types.PDKeyManualInstallDownloadAddress]
	if !ok {
		return "", fmt.Errorf("private data does not contain file address")
	}
	downloadAddress, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("private data does not contain valid file address")
	}

	// action name report detect info.
	raw, ok = privateData[types.PDKeyManualInstallActionNameReportDetectInfo]
	if !ok {
		return "", fmt.Errorf("private data does not contain action name report detect info")
	}
	actionNameReportDetectInfo, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("private data does not contain valid action name report detect info")
	}

	// action name get exec command.
	raw, ok = privateData[types.PDKeyManualInstallActionNameGetExecCommand]
	if !ok {
		return "", fmt.Errorf("private data does not contain action name get exec command")
	}
	actionNameGetExecCommand, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("private data does not contain valid action name get exec command")
	}

	content = strings.ReplaceAll(content, manualScriptTemplateKeyOperInstID, operInstID)
	content = strings.ReplaceAll(content, manualScriptTemplateKeyCallbackAddress, callbackAddress)
	content = strings.ReplaceAll(content, manualScriptTemplateKeyDownloadAddress, downloadAddress)
	content = strings.ReplaceAll(content, manualScriptTemplateKeyActionNameReportDetectInfo, actionNameReportDetectInfo)
	content = strings.ReplaceAll(content, manualScriptTemplateKeyActionNameGetExecCommand, actionNameGetExecCommand)

	return content, nil
}

const (
	manualScriptTemplateKeyOperInstID                 = "__BK_NODEMGR_OPERATION_INSTANCE_ID__"
	manualScriptTemplateKeyActionNameReportDetectInfo = "__BK_NODEMGR_ACTION_NAME_REPORT_DETECT_INFO__"
	manualScriptTemplateKeyActionNameGetExecCommand   = "__BK_NODEMGR_ACTION_NAME_GET_EXEC_COMMAND__"
	manualScriptTemplateKeyCallbackAddress            = "__BK_NODEMGR_CALLBACK_ADDRESS__"
	manualScriptTemplateKeyDownloadAddress            = "__BK_NODEMGR_DOWNLOAD_ADDRESS__"
)
