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
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/sshx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	reportRelayDetectResultURL = "/relay/report_detect_result"
)

func (h *handler) DetectInfoBySSH(ctx context.Context, payload []byte) {
	h.logger.Infof("handler detect info by ssh event.")

	var (
		event        protoRelay.DetectInfoBySSHReq
		osType       criteria.OSType
		cpuArch      criteria.CPUArch
		connectedDir string
		errMsg       string
	)

	defer func() {
		if err := h.reportHostInfo(ctx, event.ActionName, event.OperInstID, osType, cpuArch, connectedDir, errMsg); err != nil {
			h.logger.Errorf("failed to report host info: %v", err)
			return
		}
		h.logger.Infof("report host info success. action-name(%s), instance-id(%s),os-type(%s), cpu-arch(%s), connected-dir(%s), err-msg(%s)",
			event.ActionName, event.OperInstID, osType, cpuArch, connectedDir, errMsg)
	}()

	if err := json.Unmarshal(payload, &event); err != nil {
		errMsg = fmt.Sprintf("failed to unmarshal detect info by ssh event: %v", err)
		h.logger.Errorf(errMsg)

		return
	}

	client, err := generateSSHClient(ctx, event.IP, int(event.Port), event.User, event.Password, types.LoginMode(event.LoginMode), h.logger)
	if err != nil {
		errMsg = fmt.Sprintf("failed to generate SSH client: %v", err)
		h.logger.Errorf(errMsg)

		return
	}

	h.logger.Infof("start to detect info. ip(%s), port(%d), user(%s)", event.IP, event.Port, event.User)

	osType, cpuArch, connectedDir, err = detectInfoBySSH(client)
	if err != nil {
		errMsg = fmt.Sprintf("failed to detect info: %v", err)
		h.logger.Errorf(errMsg)

		return
	}
	h.logger.Infof("detect info success. os-type(%s), cpu-arch(%s), connected-dir(%s)", osType, cpuArch, connectedDir)

	h.logger.Infof("detect info by ssh success. ip(%s), port(%d), user(%s)", event.IP, event.Port, event.User)
}

// nolint: nonamedreturns,perfsprint
func detectInfoBySSH(client *sshx.Client) (
	osType criteria.OSType, cpuArch criteria.CPUArch, connectedDir string, err error) {

	// 1. detect target system
	osTypeStr, err := client.RunCommand("uname -s")
	if err != nil {
		err = fmt.Errorf("failed to run (uname -s): %w", err)

		return "", "", "", err
	}
	osTypeStr = strings.TrimFunc(strings.ToLower(osTypeStr), func(r rune) bool {
		return r == '\n'
	})
	osType, err = platform.NormalizeOS(osTypeStr)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to detect info: %w", err)
	}

	switch osType {
	case criteria.OSLinux, criteria.OSDarwin:
	default:
		err = fmt.Errorf("unsupported os type, os-type(%s)", osType)

		return "", "", "", err
	}

	// 2. detect target cpu arch
	cpuArchStr, err := client.RunCommand("uname -m")
	if err != nil {
		return "", "", "", fmt.Errorf("failed to run (uname -m): %w", err)
	}
	cpuArchStr = strings.TrimFunc(strings.ToLower(cpuArchStr), func(r rune) bool {
		return r == '\n'
	})
	cpuArch, err = platform.NormalizeArch(cpuArchStr)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to detect info: %w", err)
	}

	// 3. detect target dir
	connectedDir, err = client.RunCommand("pwd")
	if err != nil {
		err = fmt.Errorf("failed to run pwd: %w", err)

		return "", "", "", err
	}
	connectedDir = strings.TrimFunc(connectedDir, func(r rune) bool {
		return r == '\n'
	})

	return osType, cpuArch, connectedDir, nil
}
