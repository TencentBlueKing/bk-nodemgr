/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package handler ...
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tmp"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
	"github.com/pkg/errors"
)

const installBatName = "install.bat"

// need workdir.
func (h *handler) TransferAndExecByWMI(ctx context.Context, payload []byte) {
	h.logger.Infof("handler detect info by wmi event.")

	var event protoRelay.TransferAndExecByWMI
	if err := json.Unmarshal(payload, &event); err != nil {
		h.logger.Errorf("failed to unmarshal transfer and exec by ssh event: %v", err)
		h.reportInstallResult(ctx, event.actionName, event.operInstID, "", err.Error())
		return
	}

	client, err := generateWMIClient(ctx,
		event.ip, event.port, event.user,
		event.loginMode, event.hostCredit, event.passwordVault,
		h.logger)
	if err != nil {
		h.logger.Errorf("failed to connect host: %v", err)
		h.reportInstallResult(ctx, event.actionName, event.operInstID, "", err.Error())
		return
	}

	stdOut, stdErr, err := client.RunCommand(ctx, "mkdir "+event.InstallerWorkDir)
	if err != nil {
		err = fmt.Errorf("failed to run mkdir %s: %w", event.InstallerWorkDir, err)

		return
	}

	toolFile, err := h.fileManager.GetFile(ctx, event.toolName)
	if err != nil {
		h.logger.Errorf("failed to get file: %v", err)
		h.reportInstallResultByWMI(ctx, event.actionName, event.operInstID, "", "", err.Error())
		return
	}

	tmpInstallFilePath := local.GetLocalFileAbsFilePath(toolFile)
	stdOut, stdErr, err = client.UploadFile(ctx, tmpInstallFilePath, event.InstallerWorkDir)
	if err != nil {
		h.logger.Errorf("failed to transfer file, stdOut(%s), stdErr(%s) : %w",
			stdOut, stdErr, err)
		h.reportInstallResultByWMI(ctx, event.actionName, event.operInstID, stdOut, stdErr, err.Error())
		return
	}

	installBat := event.installParams

	tmpInstallBat, err := tmp.NewTempFileWithSpecialName(io.NopCloser(strings.NewReader(installBat)), installBatName)
	if err != nil {
		h.logger.Errorf("failed to create temp file: %w", err)
		h.reportInstallResultByWMI(ctx, event.actionName, event.operInstID, "", "", err.Error())
		return
	}
	defer tmp.Clean()

	_, _, err = client.UploadFile(ctx, tmpInstallBat.Path(), event.InstallerWorkDir)
	if err != nil {
		h.logger.Errorf("failed to transfer file: %w", err)
		h.reportInstallResultByWMI(ctx, event.actionName, event.operInstID, "", "", err.Error())

		return
	}

	installCMD := winpath.Clean(winpath.Join(event.InstallerWorkDir, installBatName))
	stdOutStr, stdErrStr, err := client.RunSilentCommand(ctx, installCMD)
	if err != nil {
		h.logger.Errorf("failed to run cmd: %v", err)
		h.reportInstallResultByWMI(ctx, event.actionName, event.operInstID, "", "", err.Error())

		return
	}

	h.reportInstallResultByWMI(ctx, event.actionName, event.operInstID, stdOutStr, stdErrStr, "")
	h.logger.Infof("transfer and exec cmd success. ip(%s), port(%d), user(%s)",
		event.ip, event.port, event.user)
}

func (h *handler) reportInstallResultByWMI(ctx context.Context,
	actionName, operInstID, stdOutStr, stdErrStr, errMsg string) error {

	h.logger.Infof("report host info. action-name(%s), instance-id(%s)",
		actionName, operInstID)

	req := &reportInstallResultByWMI{
		ActionName: actionName,
		OperInstID: operInstID,

		StdOutStr: stdOutStr,
		StdErrStr: stdErrStr,
		ErrMsg:    errMsg,
	}
	jsonData, err := json.Marshal(req)
	if err != nil {
		h.logger.Errorf("failed to marshal install result request: %v", err)
		return fmt.Errorf("failed to marshal install result request: %w", err)
	}

	errCh := h.client.ClientPushReq(ctx, reportRelayFileStateURL, jsonData)

	select {
	case err := <-errCh:
		if err != nil {
			h.logger.Errorf("report install result failed. action-name(%s), instance-id(%s): %v",
				req.ActionName, req.OperInstID, err)

			return fmt.Errorf("report install result failed. action-name(%s), instance-id(%s): %w",
				req.ActionName, req.OperInstID, err)
		}
		h.logger.Infof("report install result success. action-name(%s), instance-id(%s)",
			req.ActionName, req.OperInstID)

		return nil
	case <-time.After(ReportPrivateDataTimeout):
		h.logger.Errorf("report install result timed out. action-name(%s), instance-id(%s)",
			req.ActionName, req.OperInstID)

		return errors.New("report install result imed out")
	}
}

type reportInstallResultByWMI struct {
	ActionName string `json:"action_name"`
	OperInstID string `json:"oper_inst_id"`

	StdOutStr string `json:"out_str"`
	StdErrStr string `json:"err_str"`
	ErrMsg    string `json:"err_msg"`
}
