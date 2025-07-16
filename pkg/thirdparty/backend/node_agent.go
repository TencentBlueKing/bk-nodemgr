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

	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// InstallAgent node agent.
func (h *handler) InstallAgent(ctx context.Context, installParam *types.NodeAgentInstallParam) (string, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return "", err
	}

	req := &protoBackend.NodeAgentInstallReq{}

	req.ConvertHostParamFromTypes(installParam)

	resp, err := h.cli.installNodeAgent(ctx, tenantID, req)
	if err != nil {
		return "", err
	}

	result := resp.ConvertResultToComm()

	return result, nil
}

// UninstallAgent node agent.
func (h *handler) UninstallAgent(ctx context.Context) (string, error) {
	return "", nil
}

// UpgradeAgent node agent.
func (h *handler) UpgradeAgent(ctx context.Context) (string, error) {
	return "", nil
}

// RestartAgent node agent.
func (h *handler) RestartAgent(ctx context.Context) (string, error) {
	return "", nil
}

// ReconfigAgent node agent.
func (h *handler) ReconfigAgent(ctx context.Context) (string, error) {
	return "", nil
}

// ReloadAgent node agent.
func (h *handler) ReloadAgent(ctx context.Context) (string, error) {
	return "", nil
}
