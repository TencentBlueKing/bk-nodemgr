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
	"net"
	"path"
	"strconv"
	"strings"
	"time"

	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/pkg/errors"
)

const reportRelayInstallResultURL = "/relay/report_install_result"

// InstallPagentBySSH installs pagent by SSH.
// nolint:funlen
func (h *handler) InstallPagentBySSH(ctx context.Context, payload []byte) {
	h.logger.Infof("handler install pagent by ssh event.")

	var (
		event  protoRelay.InstallPagentBySSHReq
		outStr string
		errMsg string
	)

	defer func() {
		if err := h.reportInstallResult(ctx, event.ActionName, event.OperInstID, outStr, errMsg); err != nil {
			h.logger.Errorf("failed to report install result: %v", err)
		}
	}()

	if err := json.Unmarshal(payload, &event); err != nil {
		errMsg = fmt.Sprintf("failed to unmarshal install pagent by ssh event: %v", err)
		h.logger.Errorf(errMsg)

		return
	}

	client, err := generateSSHClient(ctx, event.IP, int(event.Port), event.User, event.Password,
		types.LoginMode(event.LoginMode), h.logger)
	if err != nil {
		errMsg = fmt.Sprintf("failed to generate ssh client: %v", err)
		h.logger.Errorf(errMsg)

		return
	}
	h.logger.Infof("connect to host success. ip(%s), port(%d), user(%s)", event.IP, event.Port, event.User)

	// ensure the workspace dir
	result, err := client.RunCommand("mkdir -p " + event.InstallerWorkDir)
	outStr += buildLogOutput("mkdir", event.InstallerWorkDir, result)
	if err != nil {
		errMsg = fmt.Sprintf("failed to mkdir -p %s, error: %v", event.InstallerWorkDir, err)
		h.logger.Errorf(errMsg)

		return
	}
	h.logger.Infof("run command mkdir success. mkdir -p %s", event.InstallerWorkDir)

	// get the tool file
	toolFile, err := h.fileManager.GetFile(ctx, event.ToolsName)
	if err != nil {
		errMsg = fmt.Sprintf("failed to get file: %v", err)
		h.logger.Errorf(errMsg)

		return
	}

	reader, err := toolFile.Content(ctx)
	if err != nil {
		errMsg = fmt.Sprintf("failed to get file content: %v", err)
		h.logger.Errorf(errMsg)

		return
	}

	installerPath := path.Clean(path.Join(event.InstallerWorkDir, event.ToolsName))
	err = client.TransferFile(reader, installerPath)
	if err != nil {
		errMsg = fmt.Sprintf("failed to transfer file: %v", err)
		h.logger.Errorf(errMsg)

		return
	}
	h.logger.Infof("transfer file success. tools-name(%s).  dest-dir(%s)", event.ToolsName, installerPath)

	// ensure tool is executable
	result, err = client.RunCommand("chmod +x " + installerPath)
	outStr += buildLogOutput("chmod", installerPath, result)
	if err != nil {
		errMsg = fmt.Sprintf("failed to chmod +x %s, error: %v", installerPath, err)
		h.logger.Errorf(errMsg)

		return
	}
	h.logger.Infof("install pagent by ssh success. chmod +x %s", installerPath)

	// execute install command
	installCmd := h.buildCMD(installerPath, event.InstallerCmd)
	h.logger.Infof("install node cmd: %s", installCmd)

	result, err = client.RunCommand(fmt.Sprintf(
		`mkdir -p %s && cd %s && echo "%s" > install.sh && sh install.sh`,
		event.InstallerWorkDir, event.InstallerWorkDir, installCmd))
	outStr += buildLogOutput("install", "install.sh", result)
	if err != nil {
		errMsg = fmt.Sprintf("failed to exec cmd: %v", err)
		h.logger.Errorf(errMsg)

		return
	}

	h.logger.Infof("install pagent by ssh success. stdout(%s). ip(%s), port(%d), user(%s),",
		outStr, event.IP, event.Port, event.User)
}

func (h *handler) reportInstallResult(ctx context.Context,
	actionName, operInstID, outStr, errMsg string) error {

	h.logger.Infof("report install info. action-name(%s), instance-id(%s)",
		actionName, operInstID)

	req := &reportInstallResult{
		ActionName: actionName,
		OperInstID: operInstID,
		StdOut:     outStr,
		ErrMsg:     errMsg,
	}
	jsonData, err := json.Marshal(req)
	if err != nil {
		h.logger.Errorf("failed to marshal install result request: %v", err)
		return fmt.Errorf("failed to marshal install result request: %w", err)
	}

	errCh := h.client.ClientPushReq(ctx, reportRelayInstallResultURL, jsonData)

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

	StdOut string `json:"std_out"`

	ErrMsg string `json:"err_msg"`
}

// add cmd backend svc and file svc.
// nolint: perfsprint
func (h *handler) buildCMD(installerPath string, args []string) string {
	args = append(args, "--filesvr_addr "+getIPV4Address(h.fileSvcIP, h.fileSvcPort))
	args = append(args, "--cbsvr_addr "+getIPV4Address(h.callbackSvcIP, h.callbackSvcPort))

	cmd := fmt.Sprintf("%s full-install %s", installerPath, strings.Join(args, " "))

	installLogPath := path.Clean(fmt.Sprintf("%s.stdout", installerPath))

	return fmt.Sprintf("%s >%s 2>&1 &", cmd, installLogPath)
}

func getIPV4Address(ip string, port int) string {
	return "http://" + net.JoinHostPort(ip, strconv.Itoa(port))
}

func buildLogOutput(action, target, result string) string {
	return fmt.Sprintf("action: %s, target: %s, result: %s\n", action, target, result)
}
