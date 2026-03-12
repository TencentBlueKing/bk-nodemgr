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
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/wmix"
)

func (h *handler) DetectInfoByWMI(nCtx contextx.IContext, payload []byte) {
	logger.G.Biz(nCtx).Info("handler detect info by wmi event")

	var (
		event   protoRelay.DetectInfoByWMIReq
		osType  criteria.OSType
		cpuArch criteria.CPUArch
		errMsg  string
	)

	defer func() {
		if err := h.reportHostInfo(nCtx, event.ActionName, event.OperInstID, osType, cpuArch, "", errMsg); err != nil {
			logger.G.Biz(nCtx).WithErr(err).Error("failed to report host info")

			return
		}

		logger.G.Biz(nCtx).
			With("action", event.ActionName, "oper-inst-id", event.OperInstID).
			With("os-type", osType, "cpu-arch", cpuArch).
			Info("done report host info")
	}()

	if err := json.Unmarshal(payload, &event); err != nil {
		logger.G.Biz(nCtx).AssignWhenLogging(&errMsg).WithErr(err).Error("failed to unmarshal detect info by wmi event")

		return
	}

	client, err := generateWMIClient(event.IP, int(event.Port), event.User, event.Password, types.LoginMode(event.LoginMode))
	if err != nil {
		logger.G.Biz(nCtx).AssignWhenLogging(&errMsg).WithErr(err).Error("failed to generate wmi client")

		return
	}
	logger.G.Biz(nCtx).With("ip", event.IP, "port", event.Port, "user", event.User).Info("try to connect to host")

	osType, cpuArch, err = detectInfoByWMI(nCtx, client)
	if err != nil {
		logger.G.Biz(nCtx).AssignWhenLogging(&errMsg).WithErr(err).Error("failed to detect info")

		return
	}

	logger.G.Biz(nCtx).
		With("ip", event.IP, "port", event.Port, "user", event.User).
		With("os-type", osType, "cpu-arch", cpuArch).
		Info("detect info by wmi successfully")
}

// nolint: nonamedreturns,perfsprint
func detectInfoByWMI(nCtx contextx.IContext, client *wmix.Client) (
	osType criteria.OSType, cpuArch criteria.CPUArch, err error) {

	// 1. detect target system
	osTypeStr, _, err := client.RunCommand(nCtx, "ver")
	if err != nil {
		return "", "", fmt.Errorf("failed to run ver: %w", err)
	}
	osTypeStr = strings.TrimFunc(strings.ToLower(osTypeStr), func(r rune) bool {
		return r == '\n'
	})
	osType, err = platfmt.NormalizeOS(osTypeStr)
	if err != nil {
		return "", "", fmt.Errorf("failed to detect info: %w", err)
	}

	switch osType {
	case criteria.OSWindows:
	default:
		return "", "", fmt.Errorf("unsupported os type, os-type(%s)", osType)
	}

	// 2. detect target cpu arch
	cpuArchStr, _, err := client.RunCommand(nCtx, "echo %PROCESSOR_ARCHITECTURE%")
	if err != nil {
		return "", "", fmt.Errorf("failed to run uname -m: %w", err)
	}
	cpuArchStr = strings.TrimFunc(strings.ToLower(cpuArchStr), func(r rune) bool {
		return r == '\n'
	})
	cpuArch, err = platfmt.NormalizeArch(cpuArchStr)
	if err != nil {
		return "", "", fmt.Errorf("failed to detect info: %w", err)
	}

	return osType, cpuArch, nil
}
