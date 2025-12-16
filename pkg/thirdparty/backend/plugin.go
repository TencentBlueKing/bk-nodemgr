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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// InstallPlugin install plugin.
func (h *Handler) InstallPlugin(ctx contextx.IContext, installParam ...*types.PluginInstallParam) (string, error) {
	req := new(protoBackend.PluginInstallReq)
	req.ConvertParamFromTypes(installParam...)

	resp, err := h.cli.installPlugin(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.GetData().GetWorkflowId(), nil
}

// CountPlugins count plugins.
func (h *Handler) CountPlugins(ctx contextx.IContext, condition *types.PluginCondition) (int64, error) {
	req := new(protoBackend.PluginListReq)
	req.ConvertConditionFromTypes(condition)
	req.OnlyCount = true

	resp, err := h.cli.listPlugins(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// ListPlugins list plugins.
func (h *Handler) ListPlugins(ctx contextx.IContext, page types.Page, condition *types.PluginCondition) ([]*types.Plugin, int64, error) {
	req := new(protoBackend.PluginListReq)
	req.ConvertConditionFromTypes(condition)
	req.Page = convertPage(page)
	req.OnlyCount = false

	resp, err := h.cli.listPlugins(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	plugins, total := resp.ConvertPluginToTypes()

	return plugins, total, nil
}

// ApplyPluginSubConfig apply plugin sub config.
func (h *Handler) ApplyPluginSubConfig(ctx contextx.IContext, applyParam ...*types.PluginApplySubConfigParam) (string, error) {
	req := new(protoBackend.PluginApplySubConfigReq)
	if err := req.ConvertParamFromTypes(applyParam...); err != nil {
		return "", err
	}

	resp, err := h.cli.applyPluginSubConfig(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.GetData().GetWorkflowId(), nil
}

// SetPluginMemo set plugin memo.
func (h *Handler) SetPluginMemo(ctx contextx.IContext, pluginName string, memo string) error {
	req := new(protoBackend.PluginSetMemoReq)
	req.PluginName = pluginName
	req.Memo = memo

	if err := h.cli.setPluginMemo(ctx, req); err != nil {
		return err
	}

	return nil
}
