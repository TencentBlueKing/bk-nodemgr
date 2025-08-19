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
	"path"
	"time"

	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/pkg/errors"
)

// need workdir.
func (h *handler) TransferAndExecBySSH(ctx context.Context, payload []byte) {
	h.logger.Infof("handler detect info by ssh event.")

	var event protoRelay.TransferAndExecBySSH
	if err := json.Unmarshal(payload, &event); err != nil {
		h.logger.Errorf("failed to unmarshal transfer and exec by ssh event: %v", err)
		h.reportInstallResult(ctx, event.actionName, event.operInstID, "", err.Error())
		return
	}

	client, err := generateSSHClient(ctx,
		event.ip, event.port, event.user,
		event.loginMode, event.hostCredit, event.passwordVault,
		h.logger)
	if err != nil {
		h.logger.Errorf("failed to connect host: %v", err)
		h.reportInstallResult(ctx, event.actionName, event.operInstID, "", err.Error())
		return
	}

	// ensure the workspace dir.
	if result, err := client.RunCommand("mkdir -p " + event.InstallerWorkDir); err != nil {
		h.logger.Errorf("failed to mkdir -p %s , result(%s):%v", event.InstallerWorkDir, result, err)
		h.reportInstallResult(ctx, event.actionName, event.operInstID, result, err.Error())
		return
	}

	toolFile, err := h.fileManager.GetFile(ctx, event.toolName)
	if err != nil {
		h.logger.Errorf("failed to get file: %v", err)
		h.reportInstallResult(ctx, event.actionName, event.operInstID, "", err.Error())
		return
	}

	reader, err := toolFile.Content(ctx)
	if err != nil {
		h.logger.Errorf("failed to get file content: %v", err)
		h.reportInstallResult(ctx, event.actionName, event.operInstID, "", err.Error())
		return
	}

	installerPath := path.Clean(path.Join(event.InstallerWorkDir, event.toolName))
	if err := client.TransferFile(reader, installerPath); err != nil {
		h.logger.Errorf("failed to transfer file: %v", err)
		h.reportInstallResult(ctx, event.actionName, event.operInstID, "", err.Error())
		return
	}

	// make sure tool is executable
	if result, err := client.RunCommand("chmod +x " + installerPath); err != nil {
		h.logger.Errorf("failed to chmod +x %s , result(%s):%v", installerPath, result, err)
		h.reportInstallResult(ctx, event.actionName, event.operInstID, result, err.Error())
		return
	}

	outStr, err := client.RunCommand(fmt.Sprintf(
		`mkdir -p %s && cd %s && echo "%s" > install.sh && sh install.sh`,
		event.InstallerWorkDir,
		event.InstallerWorkDir,
		event.installParams),
	)
	if err != nil {
		h.logger.Errorf("failed to exec cmd: %v", err)
		h.reportInstallResult(ctx, event.actionName, event.operInstID, outStr, err.Error())
		return
	}

	h.reportInstallResult(ctx, event.actionName, event.operInstID, outStr, "")
	h.logger.Infof("transfer and exec cmd success. ip(%s), port(%d), user(%s)",
		event.ip, event.port, event.user)
}

func (h *handler) reportInstallResult(ctx context.Context,
	actionName, operInstID, outStr, errMsg string) error {

	h.logger.Infof("report host info. action-name(%s), instance-id(%s)",
		actionName, operInstID)

	req := &reportInstallResult{
		ActionName: actionName,
		OperInstID: operInstID,
		OutStr:     outStr,
		ErrMsg:     errMsg,
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

type reportInstallResult struct {
	ActionName string `json:"action_name"`
	OperInstID string `json:"oper_inst_id"`

	OutStr string `json:"out_str"`
	ErrMsg string `json:"err_msg"`
}
