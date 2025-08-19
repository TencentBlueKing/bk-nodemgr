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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	protoRelay "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/relay"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/sshx"
	"github.com/pkg/errors"
)

func (h *handler) DetectInfoBySSH(ctx context.Context, payload []byte) {
	h.logger.Infof("handler detect info by ssh event.")

	var event protoRelay.DetechInfoBySSH
	if err := json.Unmarshal(payload, &event); err != nil {
		h.logger.Errorf("failed to unmarshal detch info by ssh event: %v", err)
		return
	}

	client, err := generateSSHClient(ctx,
		event.ip, event.port, event.user,
		event.loginMode, event.hostCredit, event.passwordVault,
		h.logger)
	if err != nil {
		h.ReportHostInfo(ctx, event.actionName, event.operInstID, "", "", "", err.Error())
		return
	}

	osType, cpuArch, connectedDir, err := detectInfo(client)
	if err != nil {
		h.logger.Errorf("failed to detect info: %v", err)
		h.ReportHostInfo(ctx, event.actionName, event.operInstID, osType, cpuArch, connectedDir, err.Error())
		return
	}

	err = h.ReportHostInfo(ctx, event.actionName, event.operInstID, osType, cpuArch, connectedDir, "")
	if err != nil {
		h.logger.Errorf("failed to report host info: %v", err)
		return
	}

	h.logger.Infof("detect info by ssh success. ip(%s), port(%d), user(%s)",
		event.ip, event.port, event.user)
}

// nolint: nonamedreturns,perfsprint
func detectInfo(client *sshx.Client) (
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

func (h *handler) ReportHostInfo(ctx context.Context,
	actionName, operInstID string,
	osType criteria.OSType, cpuArch criteria.CPUArch,
	connectedDir, msg string) error {

	h.logger.Infof("report host info. action-name(%s), instance-id(%s)",
		actionName, operInstID)

	req := &reportHostInfo{
		ActionName:   actionName,
		OperInstID:   operInstID,
		OsType:       string(osType),
		CpuArch:      string(cpuArch),
		ConnectedDir: connectedDir,
		ErrMsg:       msg,
	}
	jsonData, err := json.Marshal(req)
	if err != nil {
		h.logger.Errorf("failed to marshal status request: %v", err)
		return fmt.Errorf("failed to marshal status request: %w", err)
	}

	errCh := h.client.ClientPushReq(ctx, reportRelayFileStateURL, jsonData)

	select {
	case err := <-errCh:
		if err != nil {
			h.logger.Errorf("report host info failed. action-name(%s), instance-id(%s): %v",
				req.ActionName, req.OperInstID, err)

			return fmt.Errorf("report host info failed. action-name(%s), instance-id(%s): %w",
				req.ActionName, req.OperInstID, err)
		}
		h.logger.Infof("report host info success. action-name(%s), instance-id(%s)",
			req.ActionName, req.OperInstID)

		return nil
	case <-time.After(ReportPrivateDataTimeout):
		h.logger.Errorf("report host info timed out. action-name(%s), instance-id(%s)",
			req.ActionName, req.OperInstID)

		return errors.New("report host info timed out")
	}
}

type reportHostInfo struct {
	ActionName string `json:"action_name"`
	OperInstID string `json:"oper_inst_id"`

	OsType       string `json:"os_type"`
	CpuArch      string `json:"cpu_arch"`
	ConnectedDir string `json:"connected_dir"`

	ErrMsg string `json:"err_msg"`
}
