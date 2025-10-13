/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package backend

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// InstallAgent node agent.
func (h *Handler) InstallAgent(ctx contextx.IContext, installParam *types.NodeAgentInstallParam) (string, error) {
	req := &protoBackend.NodeAgentInstallReq{}

	req.ConvertHostParamFromTypes(installParam)

	resp, err := h.cli.installNodeAgent(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to install agent: %w", err)
	}

	result := resp.ConvertResultToComm()

	return result, nil
}

// CheckAgentInstall check agent install.
func (h *Handler) CheckAgentInstall(ctx contextx.IContext,
	checkParam []*types.NodeAgentInstallCheckInfo) ([]*types.NodeAgentInstallCheckResult, error) {

	req := &protoBackend.NodeAgentInstallCheckReq{}

	req.ConvertHostParamFromTypes(checkParam)

	resp, err := h.cli.checkAgentInstall(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to check agent install: %w", err)
	}

	result := resp.ConvertResultToTypes()

	return result, nil
}

// UninstallAgent node agent.
func (h *Handler) UninstallAgent(_ contextx.IContext) (string, error) {
	return "", nil
}

// UpgradeAgent node agent.
func (h *Handler) UpgradeAgent(ctx contextx.IContext, upgradeParam *types.NodeAgentUpgradeParam) (string, error) {
	req := new(protoBackend.NodeAgentUpgradeReq)

	req.ConvertParamFromTypes(upgradeParam)

	resp, err := h.cli.upgradeNodeAgent(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to upgrade agent: %w", err)
	}

	return resp.GetWorkflowID(), nil
}

// RestartAgent node agent.
func (h *Handler) RestartAgent(ctx contextx.IContext, restartParam *types.NodeAgentRestartParam) (string, error) {
	req := new(protoBackend.NodeAgentRestartReq)
	req.ConvertParamFromTypes(restartParam)

	resp, err := h.cli.restartNodeAgent(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.GetWorkflowID(), nil
}

// ReconfigAgent node agent.
func (h *Handler) ReconfigAgent(ctx contextx.IContext, reconfigParam *types.NodeAgentReconfigParam) (string, error) {
	req := new(protoBackend.NodeAgentReconfigReq)
	req.ConvertParamFromTypes(reconfigParam)

	resp, err := h.cli.reconfigNodeAgent(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.GetWorkflowID(), nil
}

// ReloadAgent node agent.
func (h *Handler) ReloadAgent(_ contextx.IContext) (string, error) {
	return "", nil
}
