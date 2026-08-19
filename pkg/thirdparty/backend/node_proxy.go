/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package backend

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandlerNodeProxy defines the node proxy Handler.
// nolint: interfacebloat
type IHandlerNodeProxy interface {
	// InstallProxy node proxy.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param installParam the install param.
	// @return the installing workflow-ids and error.
	InstallProxy(nCtx contextx.IContext, installParam *types.NodeProxyInstallParam) (string, error)

	// UpgradeProxy node proxy.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param upgradeParam the upgrade param.
	// @return the upgrading workflow-ids and error.
	UpgradeProxy(nCtx contextx.IContext, upgradeParam *types.NodeProxyUpgradeParam) (string, error)

	// RestartProxy node proxy.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param restartParam the restart param.
	// @return the restarting workflow-ids and error.
	RestartProxy(nCtx contextx.IContext, restartParam *types.NodeProxyRestartParam) (string, error)

	// ReconfigProxy node proxy.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param reconfigParam the reconfig param.
	// @return the reconfig workflow-ids and error.
	ReconfigProxy(nCtx contextx.IContext, reconfigParam *types.NodeProxyReconfigParam) (string, error)

	// UpdateProxy update node proxy.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param updateParam the update param.
	// @return the error.
	UpdateProxy(nCtx contextx.IContext, updateParam *types.NodeProxyUpdateParam) error

	// UpdateProxyOpsFields update node proxy ops fields.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param hosts the host list with ops fields.
	// @return the error.
	UpdateProxyOpsFields(nCtx contextx.IContext, hosts []*types.Host) error

	// UninstallProxy node proxy.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param uninstallParam the uninstall param.
	// @return the restarting workflow-ids and error.
	UninstallProxy(nCtx contextx.IContext, uninstallParm *types.NodeProxyUninstallParam) (string, error)

	// CheckInstallProxy node proxy.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param params the check param.
	// @return the check result and error.
	CheckInstallProxy(ctx contextx.IContext, params []*types.NodeProxyInstallCheckParam) ([]*types.NodeProxyInstallCheckResult, error)

	// CheckUpgradeProxy checks whether node proxies can be upgraded.
	// @param ctx contextx.IContext, contains tenant-id and username.
	// @param params the upgrade check params.
	// @param targetVersions the target versions for os/arch matching.
	// @return the check results and error.
	CheckUpgradeProxy(
		ctx contextx.IContext,
		params []*types.NodeProxyUpgradeCheckParam,
		targetVersions []*types.TargetVersion,
	) ([]*types.NodeProxyUpgradeCheckResult, error)

	// AssignUnitProxy batch-assigns a network unit to proxy hosts.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param param the assign unit param.
	// @return the assign unit result and error.
	AssignUnitProxy(nCtx contextx.IContext, param *types.NodeProxyAssignUnitParam) (*types.NodeProxyAssignUnitResult, error)

	// AssignUnitProxyMulti batch-assigns multiple network units to proxy hosts.
	AssignUnitProxyMulti(nCtx contextx.IContext, param *types.NodeProxyAssignUnitMultiParam) (*types.NodeProxyAssignUnitResult, error)
}

// InstallProxy install node proxy.
func (h *Handler) InstallProxy(nCtx contextx.IContext, installParam *types.NodeProxyInstallParam) (string, error) {
	req := new(protoBackend.NodeProxyInstallReq)
	req.ConvertParamFromTypes(installParam)

	resp, err := h.cli.installNodeProxy(nCtx, req)
	if err != nil {
		return "", err
	}

	return resp.GetWorkflowID(), nil
}

// UpgradeProxy node proxy.
func (h *Handler) UpgradeProxy(nCtx contextx.IContext, upgradeParam *types.NodeProxyUpgradeParam) (string, error) {
	req := new(protoBackend.NodeProxyUpgradeReq)
	req.ConvertParamFromTypes(upgradeParam)

	resp, err := h.cli.upgradeNodeProxy(nCtx, req)
	if err != nil {
		return "", err
	}

	return resp.GetWorkflowID(), nil
}

// RestartProxy node proxy.
func (h *Handler) RestartProxy(nCtx contextx.IContext, restartParam *types.NodeProxyRestartParam) (string, error) {
	req := new(protoBackend.NodeProxyRestartReq)
	req.ConvertParamFromTypes(restartParam)

	resp, err := h.cli.restartNodeProxy(nCtx, req)
	if err != nil {
		return "", err
	}

	return resp.GetWorkflowID(), nil
}

// ReconfigProxy node proxy.
func (h *Handler) ReconfigProxy(nCtx contextx.IContext, reconfigParam *types.NodeProxyReconfigParam) (string, error) {
	req := new(protoBackend.NodeProxyReconfigReq)
	req.ConvertParamFromTypes(reconfigParam)

	resp, err := h.cli.reconfigNodeProxy(nCtx, req)
	if err != nil {
		return "", err
	}

	return resp.GetWorkflowID(), nil
}

// UpdateProxy update node proxy.
func (h *Handler) UpdateProxy(nCtx contextx.IContext, updateParam *types.NodeProxyUpdateParam) error {
	req := new(protoBackend.NodeProxyUpdateReq)
	req.ConvertParamFromTypes(updateParam)

	_, err := h.cli.updateNodeProxy(nCtx, req)
	if err != nil {
		return err
	}

	return nil
}

// UpdateProxyOpsFields update node proxy ops fields.
func (h *Handler) UpdateProxyOpsFields(nCtx contextx.IContext, hosts []*types.Host) error {
	req := new(protoBackend.NodeProxyUpdateOpsFieldsReq)
	req.ConvertParamFromTypes(hosts)

	_, err := h.cli.updateNodeProxyOpsFields(nCtx, req)
	if err != nil {
		return err
	}

	return nil
}

// UninstallProxy node proxy.
func (h *Handler) UninstallProxy(nCtx contextx.IContext, uninstallParam *types.NodeProxyUninstallParam) (string, error) {
	req := new(protoBackend.NodeProxyUninstallReq)
	req.ConvertParamFromTypes(uninstallParam)

	resp, err := h.cli.uninstallNodeProxy(nCtx, req)
	if err != nil {
		return "", fmt.Errorf("failed to uninstall proxy: %w", err)
	}

	return resp.GetWorkflowID(), nil
}

// CheckInstallProxy check proxy install.
func (h *Handler) CheckInstallProxy(ctx contextx.IContext, params []*types.NodeProxyInstallCheckParam) ([]*types.NodeProxyInstallCheckResult, error) {
	req := &protoBackend.NodeProxyInstallCheckReq{}

	req.ConvertParamFromTypes(params)

	resp, err := h.cli.checkInstallProxy(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to check install proxy: %w", err)
	}

	result := resp.ConvertResultToTypes()

	return result, nil
}

// CheckUpgradeProxy checks whether node proxies can be upgraded.
func (h *Handler) CheckUpgradeProxy(
	ctx contextx.IContext,
	params []*types.NodeProxyUpgradeCheckParam,
	targetVersions []*types.TargetVersion,
) ([]*types.NodeProxyUpgradeCheckResult, error) {

	req := &protoBackend.NodeProxyUpgradeCheckReq{}

	req.ConvertParamFromTypes(params, targetVersions)

	resp, err := h.cli.checkUpgradeProxy(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to check upgrade proxy: %w", err)
	}

	return resp.ConvertResultToTypes(), nil
}

// AssignUnitProxy batch-assigns a network unit to proxy hosts.
func (h *Handler) AssignUnitProxy(nCtx contextx.IContext, param *types.NodeProxyAssignUnitParam) (
	*types.NodeProxyAssignUnitResult, error) {

	req := new(protoBackend.NodeProxyAssignUnitReq)
	req.ConvertParamFromTypes(param)

	resp, err := h.cli.assignUnitNodeProxy(nCtx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to assign unit proxy: %w", err)
	}

	data := resp.GetData()

	return &types.NodeProxyAssignUnitResult{
		SuccessCount:  data.GetSuccessCount(),
		FailedCount:   data.GetFailedCount(),
		FailedReasons: data.GetFailedReasons(),
		WorkflowID:    data.GetWorkflowId(),
	}, nil
}

// AssignUnitProxyMulti batch-assigns multiple network units to proxy hosts.
func (h *Handler) AssignUnitProxyMulti(nCtx contextx.IContext, param *types.NodeProxyAssignUnitMultiParam) (
	*types.NodeProxyAssignUnitResult, error) {

	req := new(protoBackend.NodeProxyAssignUnitMultiReq)
	req.ConvertParamFromTypes(param)

	resp, err := h.cli.assignUnitNodeProxyMulti(nCtx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to assign multiple units to proxy: %w", err)
	}

	data := resp.GetData()

	return &types.NodeProxyAssignUnitResult{
		SuccessCount:  data.GetSuccessCount(),
		FailedCount:   data.GetFailedCount(),
		FailedReasons: data.GetFailedReasons(),
		WorkflowID:    data.GetWorkflowId(),
	}, nil
}
