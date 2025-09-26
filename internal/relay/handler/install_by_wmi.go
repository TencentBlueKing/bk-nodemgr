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
	"io"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tmp"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// InstallPagentBywmi installs pagent by wmi.
// nolint:funlen, errcheck
func (h *handler) InstallPagentByWMI(nCtx contextx.IContext, payload []byte) {
	logger.G.Biz(nCtx).Info("handler install pagent by wmi event")

	var (
		event  protoRelay.InstallPagentByWMIReq
		outStr string
		errMsg string
	)

	defer func() {
		if err := h.reportInstallResult(nCtx, event.ActionName, event.OperInstID, outStr, errMsg); err != nil {
			logger.G.Biz(nCtx).WithErr(err).Error("failed to report install result")

			return
		}

		logger.G.Biz(nCtx).With("stdout", outStr, "ip", event.IP, "port", event.Port, "user", event.User).Info("done report install result by wmi")
	}()

	if err := json.Unmarshal(payload, &event); err != nil {
		logger.G.Biz(nCtx).AssignWhenLogging(&errMsg).WithErr(err).Error("failed to unmarshal install pagent by wmi event")

		return
	}

	// connect to host.
	client, err := generateWMIClient(nCtx, event.IP, int(event.Port), event.User, event.Password,
		types.LoginMode(event.LoginMode))
	if err != nil {
		logger.G.Biz(nCtx).AssignWhenLogging(&errMsg).WithErr(err).Error("failed to generate wmi client")

		return
	}
	logger.G.Biz(nCtx).With("ip", event.IP, "port", event.Port, "user", event.User).Info("connect to host successfully")

	// make sure the installer workspace exists
	stdOut, stdErr, err := client.RunCommand(nCtx, "mkdir "+event.InstallerWorkDir)
	if err != nil {
		logger.G.Biz(nCtx).AssignWhenLogging(&errMsg).WithErr(err).With("dir", event.InstallerWorkDir).Error("failed to make dir")

		return
	}
	outStr += buildLogOutput("mkdir", event.InstallerWorkDir, stdOut, stdErr)
	logger.G.Biz(nCtx).With("dir", event.InstallerWorkDir, "stdout", stdOut, "stderr", stdErr).Info("mkdir successfully")

	// transfer tools
	toolFile, err := h.fileManager.GetFile(nCtx, event.ToolsName)
	if err != nil {
		logger.G.Biz(nCtx).AssignWhenLogging(&errMsg).WithErr(err).With("filename", event.ToolsName).Error("failed to get file")

		return
	}

	installerPath := local.GetLocalFileAbsFilePath(toolFile)
	stdOut, stdErr, err = client.UploadFile(nCtx, installerPath, event.InstallerWorkDir)
	if err != nil {
		logger.G.Biz(nCtx).
			AssignWhenLogging(&errMsg).
			WithErr(err).
			With("filename", event.ToolsName, "dest-dir", installerPath).
			Error("failed to tranfser file")

		return
	}
	outStr += buildLogOutput("upload", event.ToolsName, stdOut, stdErr)
	logger.G.Biz(nCtx).
		With("filename", event.ToolsName, "dest-dir", installerPath, "stdout", stdOut, "stderr", stdErr).
		Info("transfer file successfully")

	tmpInstallBat, err := tmp.NewTempFileWithSpecialName(io.NopCloser(strings.NewReader(event.InstallerCmd)), event.InstallerBatName)
	if err != nil {
		errMsg = fmt.Sprintf("failed to create temp file: %v", err)
		return
	}
	defer func() {
		if err := tmp.Clean(); err != nil {
			logger.G.Biz(nCtx).AssignWhenLogging(&errMsg).WithErr(err).Error("failed to clean temp file")
		}
	}()

	logger.G.Biz(nCtx).With("cmd", event.InstallerCmd).Info("try to run install command")

	// transfer install bat file
	stdOut, stdErr, err = client.UploadFile(nCtx, tmpInstallBat.Path(), event.InstallerWorkDir)
	if err != nil {
		errMsg = fmt.Sprintf("failed to transfer file: %v", err)
		return
	}
	outStr += buildLogOutput("upload", event.InstallerBatName, stdOut, stdErr)
	logger.G.Biz(nCtx).
		With("filename", event.InstallerBatName, "dest-dir", installerPath, "stdout", stdOut, "stderr", stdErr).
		Info("transfer file successfully")

	// execute install bat
	installCMD := winpath.Clean(winpath.Join(event.InstallerWorkDir, event.InstallerBatName))
	stdOut, stdErr, err = client.RunSilentCommand(nCtx, installCMD)
	if err != nil {
		errMsg = fmt.Sprintf("failed to run install node: %v", err)

		return
	}
	outStr += buildLogOutput("install", installCMD, stdOut, stdErr)

	logger.G.Biz(nCtx).With("stdout", stdOut, "ip", event.IP, "port", event.Port, "user", event.User).Info("install pagent by wmi successfully")
}
