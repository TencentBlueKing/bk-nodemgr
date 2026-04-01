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

// IHandlerNodeAgent defines the node agent Handler.
type IHandlerNodeAgent interface {
	// InstallAgent node agent.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param installParam the install param.
	// @return the installing workflow-ids and error.
	InstallAgent(nCtx contextx.IContext, installParam *types.NodeAgentInstallParam) (string, error)

	// UpgradeAgent node agent.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param upgradeParam the upgrade param.
	// @return the upgrading workflow-ids and error.
	UpgradeAgent(nCtx contextx.IContext, upgradeParam *types.NodeAgentUpgradeParam) (string, error)

	// ReconfigAgent node agent.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param reconfigParam the reconfig param.
	// @return the reconfig workflow-ids and error.
	ReconfigAgent(nCtx contextx.IContext, reconfigParam *types.NodeAgentReconfigParam) (string, error)

	// RestartProxy node agent.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param restartParam the restart param.
	// @return the restarting workflow-ids and error.
	RestartAgent(nCtx contextx.IContext, restartParam *types.NodeAgentRestartParam) (string, error)

	// CheckInstallAgent node agent.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param params the check param.
	// @return the check result and error.
	CheckInstallAgent(ctx contextx.IContext, params []*types.NodeAgentInstallCheckParam) ([]*types.NodeAgentInstallCheckResult, error)

	// CheckUpgradeAgent checks whether node agents can be upgraded.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param params the upgrade check params.
	// @param targetVersions the target versions for os/arch matching.
	// @return the check results and error.
	CheckUpgradeAgent(
		ctx contextx.IContext,
		params []*types.NodeAgentUpgradeCheckParam,
		targetVersions []*types.TargetVersion,
	) ([]*types.NodeAgentUpgradeCheckResult, error)

	// UninstallAgent node agent.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param uninstallParam the uninstall param.
	// @return the restarting workflow-ids and error.
	UninstallAgent(nCtx contextx.IContext, uninstallParam *types.NodeAgentUninstallParam) (string, error)

	// AssignUnitAgent batch-assigns a network unit to unassigned hosts.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param hostIDs the host IDs to assign.
	// @param networkUnitID the target network unit ID.
	// @return success count, failed count, failed reasons, and error.
	AssignUnitAgent(nCtx contextx.IContext, hostIDs []int64, networkUnitID int64) (int64, int64, []string, error)
}

// InstallAgent node agent.
func (h *Handler) InstallAgent(nCtx contextx.IContext, installParam *types.NodeAgentInstallParam) (string, error) {
	req := &protoBackend.NodeAgentInstallReq{}

	req.ConvertHostParamFromTypes(installParam)

	resp, err := h.cli.installNodeAgent(nCtx, req)
	if err != nil {
		return "", fmt.Errorf("failed to install agent: %w", err)
	}

	result := resp.ConvertResultToComm()

	return result, nil
}

// CheckInstallAgent check agent install.
func (h *Handler) CheckInstallAgent(ctx contextx.IContext, params []*types.NodeAgentInstallCheckParam) ([]*types.NodeAgentInstallCheckResult, error) {
	req := &protoBackend.NodeAgentInstallCheckReq{}

	req.ConvertParamFromTypes(params)

	resp, err := h.cli.checkInstallAgent(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to check install agent: %w", err)
	}

	result := resp.ConvertResultToTypes()

	return result, nil
}

// CheckUpgradeAgent checks whether node agents can be upgraded.
func (h *Handler) CheckUpgradeAgent(
	ctx contextx.IContext,
	params []*types.NodeAgentUpgradeCheckParam,
	targetVersions []*types.TargetVersion,
) ([]*types.NodeAgentUpgradeCheckResult, error) {

	req := &protoBackend.NodeAgentUpgradeCheckReq{}

	req.ConvertParamFromTypes(params, targetVersions)

	resp, err := h.cli.checkUpgradeAgent(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to check upgrade agent: %w", err)
	}

	return resp.ConvertResultToTypes(), nil
}

// UpgradeAgent node agent.
func (h *Handler) UpgradeAgent(nCtx contextx.IContext, upgradeParam *types.NodeAgentUpgradeParam) (string, error) {
	req := new(protoBackend.NodeAgentUpgradeReq)

	req.ConvertParamFromTypes(upgradeParam)

	resp, err := h.cli.upgradeNodeAgent(nCtx, req)
	if err != nil {
		return "", fmt.Errorf("failed to upgrade agent: %w", err)
	}

	return resp.GetWorkflowID(), nil
}

// RestartAgent node agent.
func (h *Handler) RestartAgent(nCtx contextx.IContext, restartParam *types.NodeAgentRestartParam) (string, error) {
	req := new(protoBackend.NodeAgentRestartReq)
	req.ConvertParamFromTypes(restartParam)

	resp, err := h.cli.restartNodeAgent(nCtx, req)
	if err != nil {
		return "", fmt.Errorf("failed to restart agent: %w", err)
	}

	return resp.GetWorkflowID(), nil
}

// ReconfigAgent node agent.
func (h *Handler) ReconfigAgent(nCtx contextx.IContext, reconfigParam *types.NodeAgentReconfigParam) (string, error) {
	req := new(protoBackend.NodeAgentReconfigReq)
	req.ConvertParamFromTypes(reconfigParam)

	resp, err := h.cli.reconfigNodeAgent(nCtx, req)
	if err != nil {
		return "", fmt.Errorf("failed to reconfig agent: %w", err)
	}

	return resp.GetWorkflowID(), nil
}

// UninstallAgent node agent.
func (h *Handler) UninstallAgent(nCtx contextx.IContext, reconfigParam *types.NodeAgentUninstallParam) (string, error) {
	req := new(protoBackend.NodeAgentUninstallReq)
	req.ConvertParamFromTypes(reconfigParam)

	resp, err := h.cli.uninstallNodeAgent(nCtx, req)
	if err != nil {
		return "", fmt.Errorf("failed to uninstall agent: %w", err)
	}

	return resp.GetWorkflowID(), nil
}

// AssignUnitAgent batch-assigns a network unit to unassigned hosts.
func (h *Handler) AssignUnitAgent(nCtx contextx.IContext, hostIDs []int64, networkUnitID int64) (
	int64, int64, []string, error) {

	req := &protoBackend.NodeAgentAssignUnitReq{
		BkHostId:        hostIDs,
		BkNetworkunitId: networkUnitID,
	}

	resp, err := h.cli.assignUnitNodeAgent(nCtx, req)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to assign unit agent: %w", err)
	}

	data := resp.GetData()

	return data.GetSuccessCount(), data.GetFailedCount(), data.GetFailedReasons(), nil
}
