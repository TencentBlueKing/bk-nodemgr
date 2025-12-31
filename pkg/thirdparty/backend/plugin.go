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

// IHandlerPlugin defines the backend Handler for plugin.
type IHandlerPlugin interface {
	// ListPlugins lists plugins by page and conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return plugin list with page and the total count with filter and error.
	ListPlugins(nCtx contextx.IContext, page types.Page, condition *types.PluginCondition) ([]*types.Plugin, int64, error)

	// CountPlugins counts plugins by conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the plugin count with filter and error.
	CountPlugins(nCtx contextx.IContext, condition *types.PluginCondition) (int64, error)

	// InstallPlugin install plugin.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param installParam the install param.
	// @return the installing workflow-ids and error.
	InstallPlugin(nCtx contextx.IContext, installParam ...*types.PluginDeploymentParam) (string, error)

	// ApplyPluginSubConfig apply plugin sub config.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param applyParam the apply param.
	// @return the applying workflow-ids and error.
	ApplyPluginSubConfig(nCtx contextx.IContext, applyParam ...*types.PluginDeploymentParam) (string, error)

	// SetPluginMemo set plugin memo.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param pluginName the plugin name.
	// @param memo the plugin memo.
	// @return the error.
	SetPluginMemo(nCtx contextx.IContext, pluginName string, memo string) error
}

// InstallPlugin install plugin.
func (h *Handler) InstallPlugin(nCtx contextx.IContext, installParam ...*types.PluginDeploymentParam) (string, error) {
	req := new(protoBackend.PluginInstallReq)
	if err := req.ConvertParamFromTypes(installParam...); err != nil {
		return "", err
	}

	resp, err := h.cli.installPlugin(nCtx, req)
	if err != nil {
		return "", err
	}

	return resp.GetData().GetWorkflowId(), nil
}

// CountPlugins count plugins.
func (h *Handler) CountPlugins(nCtx contextx.IContext, condition *types.PluginCondition) (int64, error) {
	req := new(protoBackend.PluginListReq)
	req.ConvertConditionFromTypes(condition)
	req.OnlyCount = true

	resp, err := h.cli.listPlugins(nCtx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// ListPlugins list plugins.
func (h *Handler) ListPlugins(nCtx contextx.IContext, page types.Page, condition *types.PluginCondition) ([]*types.Plugin, int64, error) {
	req := new(protoBackend.PluginListReq)
	req.ConvertConditionFromTypes(condition)
	req.Page = convertPage(page)
	req.OnlyCount = false

	resp, err := h.cli.listPlugins(nCtx, req)
	if err != nil {
		return nil, 0, err
	}

	plugins, total := resp.ConvertPluginToTypes()

	return plugins, total, nil
}

// ApplyPluginSubConfig apply plugin sub config.
func (h *Handler) ApplyPluginSubConfig(nCtx contextx.IContext, applyParam ...*types.PluginDeploymentParam) (string, error) {
	req := new(protoBackend.PluginApplySubConfigReq)
	if err := req.ConvertParamFromTypes(applyParam...); err != nil {
		return "", err
	}

	resp, err := h.cli.applyPluginSubConfig(nCtx, req)
	if err != nil {
		return "", err
	}

	return resp.GetData().GetWorkflowId(), nil
}

// SetPluginMemo set plugin memo.
func (h *Handler) SetPluginMemo(nCtx contextx.IContext, pluginName string, memo string) error {
	req := new(protoBackend.PluginSetMemoReq)
	req.PluginName = pluginName
	req.Memo = memo

	if err := h.cli.setPluginMemo(nCtx, req); err != nil {
		return err
	}

	return nil
}
