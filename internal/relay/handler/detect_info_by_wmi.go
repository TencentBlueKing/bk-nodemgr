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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/wmix"
)

func (h *handler) DetectInfoByWMI(ctx context.Context, payload []byte) {
	h.logger.Infof("handler detect info by wmi event.")

	var (
		event   protoRelay.DetectInfoByWMIReq
		osType  criteria.OSType
		cpuArch criteria.CPUArch
		errMsg  string
	)

	defer func() {
		if err := h.reportHostInfo(ctx, event.ActionName, event.OperInstID, osType, cpuArch, "", errMsg); err != nil {
			h.logger.Errorf("failed to report host info: %v", err)
			return
		}
		h.logger.Infof("report host info success. action-name(%s), instance-id(%s),os-type(%s), cpu-arch(%s)",
			event.ActionName, event.OperInstID, osType, cpuArch)
	}()

	if err := json.Unmarshal(payload, &event); err != nil {
		errMsg = fmt.Sprintf("failed to unmarshal detect info by wmi event: %v", err)
		h.logger.Errorf(errMsg)

		return
	}

	client, err := generateWMIClient(ctx, event.IP, int(event.Port), event.User, event.Password, types.LoginMode(event.LoginMode), h.logger)
	if err != nil {
		errMsg = fmt.Sprintf("failed to generate wmi client: %v", err)
		h.logger.Errorf(errMsg)

		return
	}

	h.logger.Infof("start to detect info. ip(%s), port(%d), user(%s)", event.IP, event.Port, event.User)

	osType, cpuArch, err = detectInfoByWMI(ctx, client)
	if err != nil {
		errMsg = fmt.Sprintf("failed to detect info: %v", err)
		h.logger.Errorf(errMsg)

		return
	}
	h.logger.Infof("detect info success. os-type(%s), cpu-arch(%s)", osType, cpuArch)

	h.logger.Infof("detect info by wmi success. ip(%s), port(%d), user(%s)", event.IP, event.Port, event.User)
}

// nolint: nonamedreturns,perfsprint
func detectInfoByWMI(ctx context.Context, client *wmix.Client) (
	osType criteria.OSType, cpuArch criteria.CPUArch, err error) {

	// 1. detect target system
	osTypeStr, _, err := client.RunCommand(ctx, "ver")
	if err != nil {
		return "", "", fmt.Errorf("failed to run ver: %w", err)
	}
	osTypeStr = strings.TrimFunc(strings.ToLower(osTypeStr), func(r rune) bool {
		return r == '\n'
	})
	osType, err = platform.NormalizeOS(osTypeStr)
	if err != nil {
		return "", "", fmt.Errorf("failed to detect info: %w", err)
	}

	switch osType {
	case criteria.OSWindows:
	default:
		return "", "", fmt.Errorf("unsupported os type, os-type(%s)", osType)
	}

	// 2. detect target cpu arch
	cpuArchStr, _, err := client.RunCommand(ctx, "echo %PROCESSOR_ARCHITECTURE%")
	if err != nil {
		return "", "", fmt.Errorf("failed to run uname -m: %w", err)
	}
	cpuArchStr = strings.TrimFunc(strings.ToLower(cpuArchStr), func(r rune) bool {
		return r == '\n'
	})
	cpuArch, err = platform.NormalizeArch(cpuArchStr)
	if err != nil {
		return "", "", fmt.Errorf("failed to detect info: %w", err)
	}

	return osType, cpuArch, nil
}
