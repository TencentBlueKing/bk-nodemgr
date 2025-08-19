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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/wmix"
)

func (h *handler) DetectInfoByWMI(ctx context.Context, payload []byte) {
	h.logger.Infof("handler detect info by ssh event.")

	var event protoRelay.DetechInfoByWMI
	if err := json.Unmarshal(payload, &event); err != nil {
		h.logger.Errorf("failed to unmarshal detch info by ssh event: %v", err)
		return
	}

	client, err := generateWMIClient(ctx,
		event.ip, event.port, event.user,
		event.loginMode, event.hostCredit, event.passwordVault,
		h.logger)
	if err != nil {
		h.ReportHostInfo(ctx, event.actionName, event.operInstID, "", "", "", err.Error())
		return
	}

	osType, cpuArch, err := detectInfoByWMI(ctx, client)
	if err != nil {
		h.logger.Errorf("failed to detect info: %v", err)
		h.ReportHostInfo(ctx, event.actionName, event.operInstID, osType, cpuArch, "", err.Error())
		return
	}

	err = h.ReportHostInfo(ctx, event.actionName, event.operInstID, osType, cpuArch, "", "")
	if err != nil {
		h.logger.Errorf("failed to report host info: %v", err)
		return
	}

	h.logger.Infof("detect info by ssh success. ip(%s), port(%d), user(%s)",
		event.ip, event.port, event.user)
}

// nolint: nonamedreturns,perfsprint
func detectInfoByWMI(ctx context.Context, client *wmix.Client) (
	osType criteria.OSType, cpuArch criteria.CPUArch, err error) {

	// 1. detect target system
	osTypeStr, _, err := client.RunCommand(ctx, "ver")
	if err != nil {
		err = fmt.Errorf("failed to run ver, err: %w", err)

		return "", "", err
	}
	osTypeStr = strings.TrimFunc(strings.ToLower(osTypeStr), func(r rune) bool {
		return r == '\n'
	})
	osType, err = platform.NormalizeOS(osTypeStr)
	if err != nil {
		return "", "", fmt.Errorf("failed to detect info, err: %w", err)
	}

	switch osType {
	case criteria.OSWindows:
	default:
		err = fmt.Errorf("unsupported os type, os-type(%s)", osType)

		return "", "", err
	}

	// 2. detect target cpu arch
	cpuArchStr, _, err := client.RunCommand(ctx, "echo %PROCESSOR_ARCHITECTURE%")
	if err != nil {
		return "", "", fmt.Errorf("failed to run uname -m, err: %w", err)
	}
	cpuArchStr = strings.TrimFunc(strings.ToLower(cpuArchStr), func(r rune) bool {
		return r == '\n'
	})
	cpuArch, err = platform.NormalizeArch(cpuArchStr)
	if err != nil {
		return "", "", fmt.Errorf("failed to detect info, err: %w", err)
	}

	return osType, cpuArch, nil
}
