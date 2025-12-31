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

// IHandlerNodeProxy defines the node proxy Handler.
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

	// UninstallProxy node proxy.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param uninstallParam the uninstall param.
	// @return the restarting workflow-ids and error.
	UninstallProxy(nCtx contextx.IContext, uninstallParm *types.NodeProxyUninstallParam) (string, error)
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
