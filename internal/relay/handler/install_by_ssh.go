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
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const reportRelayInstallResultURL = "/relay/report_install_result"

// InstallPagentBySSH installs pagent by SSH.
// nolint:funlen
func (h *handler) InstallPagentBySSH(nCtx contextx.IContext, payload []byte) {
	logger.G.Biz(nCtx).Info("handler install pagent by ssh event")

	var (
		event  protoRelay.InstallPagentBySSHReq
		outStr string
		errMsg string
	)

	defer func() {
		if err := h.reportInstallResult(nCtx, event.ActionName, event.OperInstID, outStr, errMsg); err != nil {
			logger.G.Biz(nCtx).WithErr(err).Error("failed to report install result")

			return
		}

		logger.G.Biz(nCtx).With("stdout", outStr, "ip", event.IP, "port", event.Port, "user", event.User).Info("done report install result by ssh")
	}()

	if err := json.Unmarshal(payload, &event); err != nil {
		logger.G.Biz(nCtx).AssignWhenLogging(&errMsg).WithErr(err).Error("failed to unmarshal install pagent by ssh event")

		return
	}

	// connect to host.
	client, err := generateSSHClient(nCtx, event.IP, int(event.Port), event.User, event.Password,
		types.LoginMode(event.LoginMode))
	if err != nil {
		logger.G.Biz(nCtx).AssignWhenLogging(&errMsg).WithErr(err).Error("failed to generate ssh client")

		return
	}
	logger.G.Biz(nCtx).With("ip", event.IP, "port", event.Port, "user", event.User).Info("connect to host successfully")

	// ensure the workspace dir
	stdoutResult, stderrResult, err := client.RunCommand("mkdir -p " + event.InstallerWorkDir)
	if err != nil {
		logger.G.Biz(nCtx).AssignWhenLogging(&errMsg).WithErr(err).With("dir", event.InstallerWorkDir).Error("failed to make dir")

		return
	}
	outStr += buildLogOutput("mkdir", event.InstallerWorkDir, stdoutResult, stderrResult)
	logger.G.Biz(nCtx).With("dir", event.InstallerWorkDir).Info("mkdir successfully")

	// get the tool file
	toolFile, err := h.fileManager.GetFile(nCtx, event.ToolsName)
	if err != nil {
		logger.G.Biz(nCtx).AssignWhenLogging(&errMsg).WithErr(err).With("filename", event.ToolsName).Error("failed to get file")

		return
	}

	reader, err := toolFile.Content(nCtx)
	if err != nil {
		logger.G.Biz(nCtx).AssignWhenLogging(&errMsg).WithErr(err).With("filename", event.ToolsName).Error("failed to get file content")

		return
	}

	// transfer tools
	installerPath := path.Clean(path.Join(event.InstallerWorkDir, event.ToolsName))
	err = client.TransferFile(reader, installerPath)
	if err != nil {
		logger.G.Biz(nCtx).
			AssignWhenLogging(&errMsg).
			WithErr(err).
			With("filename", event.ToolsName, "dest-dir", installerPath).
			Error("failed to tranfser file")

		return
	}
	logger.G.Biz(nCtx).With("filename", event.ToolsName, "dest-dir", installerPath).Info("transfer file successfully")

	// ensure tool is executable
	stdoutResult, stderrResult, err = client.RunCommand("chmod +x " + installerPath)
	if err != nil {
		logger.G.Biz(nCtx).AssignWhenLogging(&errMsg).WithErr(err).With("filename", installerPath).Error("failed to chmod file")

		return
	}
	outStr += buildLogOutput("chmod", installerPath, stdoutResult, stderrResult)
	logger.G.Biz(nCtx).With("filename", installerPath).Info("run command chmod successfully")

	// execute install command
	logger.G.Biz(nCtx).With("cmd", event.InstallerCmd).Info("try to run install command")

	stdoutResult, stderrResult, err = client.RunCommand(event.InstallerCmd)
	if err != nil {
		logger.G.Biz(nCtx).AssignWhenLogging(&errMsg).WithErr(err).With("cmd", event.InstallerCmd).Error("failed to run install cmd")

		return
	}
	outStr += buildLogOutput("install", "install.sh", stdoutResult, stderrResult)

	logger.G.Biz(nCtx).With("stdout", outStr, "ip", event.IP, "port", event.Port, "user", event.User).Info("install pagent by ssh successfully")
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
