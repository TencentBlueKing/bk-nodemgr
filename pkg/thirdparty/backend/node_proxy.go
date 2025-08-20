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
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// InstallProxy install node proxy.
func (h *handler) InstallProxy(ctx contextx.ITenantUserContext, installParam *types.NodeProxyInstallParam) (string, error) {
	req := new(protoBackend.NodeProxyInstallReq)
	req.ConvertParamFromTypes(installParam)

	resp, err := h.cli.installNodeProxy(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.GetWorkflowID(), nil
}

// UpgradeProxy node proxy.
func (h *handler) UpgradeProxy(ctx contextx.ITenantUserContext, upgradeParam *types.NodeProxyUpgradeParam) (string, error) {
	req := new(protoBackend.NodeProxyUpgradeReq)
	req.ConvertParamFromTypes(upgradeParam)

	resp, err := h.cli.upgradeNodeProxy(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.GetWorkflowID(), nil
}

// UninstallProxy node proxy.
func (h *handler) UninstallProxy(ctx context.Context) (string, error) {
	return "", nil
}

// RestartProxy node proxy.
func (h *handler) RestartProxy(ctx context.Context) (string, error) {
	return "", nil
}

// ReconfigProxy node proxy.
func (h *handler) ReconfigProxy(ctx context.Context) (string, error) {
	return "", nil
}

// ReloadProxy node proxy.
func (h *handler) ReloadProxy(ctx context.Context) (string, error) {
	return "", nil
}
