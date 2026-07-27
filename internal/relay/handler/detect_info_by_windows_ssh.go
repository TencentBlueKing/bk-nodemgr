/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package handler

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// DetectInfoByWindowsSSH detects host info by Windows SSH.
func (h *handler) DetectInfoByWindowsSSH(nCtx contextx.IContext, payload []byte) {
	logger.G.Biz(nCtx).Info("handler detect info by windows ssh event")

	var (
		event   protoRelay.DetectInfoByWindowsSSHReq
		osType  criteria.OSType
		cpuArch criteria.CPUArch
		errMsg  string
	)

	defer func() {
		if err := h.reportHostInfo(nCtx, event.ActionName, event.OperInstID, osType, cpuArch, "", errMsg); err != nil {
			logger.G.Biz(nCtx).WithErr(err).Error("failed to report windows ssh detect result")

			return
		}

		logger.G.Biz(nCtx).
			With("action", event.ActionName, "oper-inst-id", event.OperInstID).
			With("os-type", osType, "cpu-arch", cpuArch).
			Info("done report host info by windows ssh")
	}()

	if err := json.Unmarshal(payload, &event); err != nil {
		errMsg = "failed to decode windows ssh detect request"
		logger.G.Biz(nCtx).WithErr(err).Error(errMsg)

		return
	}

	client, err := generateWindowsSSHClient(
		nCtx,
		event.IP,
		int(event.Port),
		event.User,
		event.Password,
		types.LoginMode(event.LoginMode),
	)
	if err != nil {
		errMsg = "failed to create windows ssh client"
		logger.G.Biz(nCtx).WithErr(err).With("ip", event.IP, "port", event.Port, "user", event.User).Error(errMsg)

		return
	}
	defer func() { _ = client.Close() }()

	osType, cpuArch, err = detectInfoByWindowsSSH(client)
	if err != nil {
		errMsg = err.Error()
		logger.G.Biz(nCtx).WithErr(err).With("ip", event.IP, "port", event.Port, "user", event.User).Error(errMsg)

		return
	}

	logger.G.Biz(nCtx).
		With("ip", event.IP, "port", event.Port, "user", event.User).
		With("os-type", osType, "cpu-arch", cpuArch).
		Info("detect info by windows ssh successfully")
}

func detectInfoByWindowsSSH(runner windowsSSHCommandRunner) (criteria.OSType, criteria.CPUArch, error) {
	profile, _, err := detectWindowsSSHProfile(runner)
	if err != nil {
		return "", "", err
	}

	if profile == windowsSSHProfileCygwin {
		return detectWindowsSSHArch(runner, "uname -m")
	}

	return detectWindowsSSHArch(runner, windowsNativeArchCommand)
}

func detectWindowsSSHArch(runner windowsSSHCommandRunner, command string) (criteria.OSType, criteria.CPUArch, error) {
	cpuArchStr, _, err := runner.RunCommand(command)
	if err != nil {
		return "", "", fmt.Errorf("failed to run %s: %w", command, err)
	}

	cpuArch, err := platfmt.NormalizeArch(strings.TrimSpace(cpuArchStr))
	if err != nil {
		return "", "", fmt.Errorf("failed to detect windows ssh cpu arch: %w", err)
	}

	return criteria.OSWindows, cpuArch, nil
}
