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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tmp"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/winpath"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// InstallPagentBywmi installs pagent by wmi.
// nolint:funlen, errcheck
func (h *handler) InstallPagentByWMI(ctx context.Context, payload []byte) {
	h.logger.Infof("handler install pagent by wmi event.")

	var (
		event  protoRelay.InstallPagentByWMIReq
		outStr string
		errMsg string
	)

	defer func() {
		if err := h.reportInstallResult(ctx, event.ActionName, event.OperInstID, outStr, errMsg); err != nil {
			h.logger.Errorf("failed to report install result: %v", err)
		}
		h.logger.Infof("done report install result by wmi. stdout(%s). ip(%s), port(%d), user(%s),",
			outStr, event.IP, event.Port, event.User)
	}()

	if err := json.Unmarshal(payload, &event); err != nil {
		errMsg = fmt.Sprintf("failed to unmarshal install pagent by wmi event: %v", err)
		h.logger.Errorf(errMsg)

		return
	}

	// connect to host.
	client, err := generateWMIClient(ctx, event.IP, int(event.Port), event.User, event.Password,
		types.LoginMode(event.LoginMode), h.logger)
	if err != nil {
		errMsg = fmt.Sprintf("failed to generate wmi client: %v", err)
		h.logger.Errorf(errMsg)

		return
	}
	h.logger.Infof("connect to host success. ip(%s), port(%d), user(%s)", event.IP, event.Port, event.User)

	// make sure the installer workspace exists
	stdOut, stdErr, err := client.RunCommand(ctx, "mkdir "+event.InstallerWorkDir)
	if err != nil {
		errMsg = fmt.Sprintf("failed to run mkdir %s : %v", event.InstallerWorkDir, err)

		return
	}
	outStr += buildLogOutput("mkdir", event.InstallerWorkDir, stdOut, stdErr)
	h.logger.Infof("make sure the installer workspace exists, stdout: %s, stderr: %s", stdOut, stdErr)

	// transfer tools
	toolFile, err := h.fileManager.GetFile(ctx, event.ToolsName)
	if err != nil {
		errMsg = fmt.Sprintf("failed to get file: %v", err)
		h.logger.Errorf(errMsg)

		return
	}

	InstallFilePath := local.GetLocalFileAbsFilePath(toolFile)
	stdOut, stdErr, err = client.UploadFile(ctx, InstallFilePath, event.InstallerWorkDir)
	if err != nil {
		errMsg = fmt.Sprintf("failed to transfer file: %v", err)

		return
	}
	outStr += buildLogOutput("upload", event.ToolsName, stdOut, stdErr)
	h.logger.Infof("transfer file success, stdout: %s, stderr: %s", stdOut, stdErr)

	tmpInstallBat, err := tmp.NewTempFileWithSpecialName(io.NopCloser(strings.NewReader(event.InstallerCmd)), event.InstallerBatName)
	if err != nil {
		errMsg = fmt.Sprintf("failed to create temp file: %v", err)
		return
	}
	defer func() {
		if err := tmp.Clean(); err != nil {
			h.logger.Errorf(fmt.Sprintf("failed to clean temp file: %v", err))
		}
	}()
	h.logger.Infof("install node cmd: %s", event.InstallerCmd)

	// transfer install bat file
	stdOut, stdErr, err = client.UploadFile(ctx, tmpInstallBat.Path(), event.InstallerWorkDir)
	if err != nil {
		errMsg = fmt.Sprintf("failed to transfer file: %v", err)
		return
	}
	outStr += buildLogOutput("upload", event.InstallerBatName, stdOut, stdErr)
	h.logger.Infof("transfer install bat file success,stdout: %s, stderr: %s", stdOut, stdErr)

	// execute install bat
	installCMD := winpath.Clean(winpath.Join(event.InstallerWorkDir, event.InstallerBatName))
	stdOut, stdErr, err = client.RunSilentCommand(ctx, installCMD)
	if err != nil {
		errMsg = fmt.Sprintf("failed to run install node: %v", err)

		return
	}
	stdOut += buildLogOutput("install", installCMD, stdOut, stdErr)
	h.logger.Infof("run install command success, stdout: %s, stderr: %s", stdOut, stdErr)

	h.logger.Infof("install pagent by wmi success. stdout(%s). ip(%s), port(%d), user(%s),",
		stdOut, event.IP, event.Port, event.User)
}
