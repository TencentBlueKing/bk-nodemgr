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

	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
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
		h.logger.Infof("done report install result by ssh. stdout(%s). ip(%s), port(%d), user(%s),",
			outStr, event.IP, event.Port, event.User)
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
	if err != nil {
		errMsg = fmt.Sprintf("failed to mkdir -p %s: %v", event.InstallerWorkDir, err)
		h.logger.Errorf(errMsg)

		return
	}
	outStr += buildLogOutput("mkdir", event.InstallerWorkDir, result, "")
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
	if err != nil {
		errMsg = fmt.Sprintf("failed to chmod +x %s: %v", installerPath, err)
		h.logger.Errorf(errMsg)

		return
	}
	outStr += buildLogOutput("chmod", installerPath, result, "")
	h.logger.Infof("run command chmod success. chmod +x %s", installerPath)

	// execute install command
	installCmd := h.buildCMD(installerPath, event.InstallerCmd)
	h.logger.Infof("install node cmd: %s", installCmd)

	result, err = client.RunCommand(fmt.Sprintf(
		`mkdir -p %s && cd %s && echo "%s" > install.sh && sh install.sh`,
		event.InstallerWorkDir, event.InstallerWorkDir, installCmd))
	if err != nil {
		errMsg = fmt.Sprintf("failed to exec cmd: %v", err)
		h.logger.Errorf(errMsg)

		return
	}
	outStr += buildLogOutput("install", "install.sh", result, "")

	h.logger.Infof("install pagent by ssh success. stdout(%s). ip(%s), port(%d), user(%s),",
		outStr, event.IP, event.Port, event.User)
}

// add cmd backend svc and file svc.
// nolint: perfsprint
func (h *handler) buildCMD(installerPath string, args []string) string {
	args = append(args, "--dlsvr_addr "+getIPV4Address(h.downloadSvcIP, h.downloadSvcPort))
	args = append(args, "--cbsvr_addr "+getIPV4Address(h.callbackSvcIP, h.callbackSvcPort))

	cmd := fmt.Sprintf("%s node full-install %s", installerPath, strings.Join(args, " "))

	installLogPath := path.Clean(fmt.Sprintf("%s.stdout", installerPath))

	return fmt.Sprintf("%s >%s 2>&1 &", cmd, installLogPath)
}
func getIPV4Address(ip string, port int) string {
	return "http://" + net.JoinHostPort(ip, strconv.Itoa(port))
}

func buildLogOutput(action, target, stdout, stderr string) string {
	var builder strings.Builder

	if stdout != "" {
		builder.WriteString(fmt.Sprintf("%s %s, stdout:\n%s\n", action, target, stdout))
	}

	if stderr != "" {
		builder.WriteString(fmt.Sprintf("%s %s, stderr:\n%s\n", action, target, stderr))
	}

	return builder.String()
}
