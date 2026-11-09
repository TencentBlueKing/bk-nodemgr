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

package plugin

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/compatibility"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/globalsettings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Install defines the handler to install plugin.
func (h *handler) Install(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginInstallReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to install plugin, failed to decode request body.")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	pluginName := conv.SliceToSlice(req.GetPlugin(), func(item *protoBackend.PluginInstallReq_InstallInfo) string {
		return item.GetPluginName()
	})
	if authErr := h.authorizedPluginOperate(rCtx, pluginName...); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).
			With("plugin-name", pluginName).
			Error("failed to install plugin, permission denied.")

		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	hostTopoMapping, err := h.daoHost.GetHostTopoRelationMapping(rCtx, req.GetHostIDs()...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to install plugin, failed to get host topo mapping.")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}
	policy := h.resolvePluginCompatibilityModePolicy(rCtx)
	pluginDeploymentParam := req.ConvertParamToTypesWithHostTopoMapping(hostTopoMapping)
	applyPluginCompatibilityModePolicy(policy, rCtx.TenantID(), pluginDeploymentParam...)

	workflowID, err := h.installPlugin(rCtx, pluginDeploymentParam)
	if err != nil {
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	if err := h.ensurePluginV2(rCtx, true, pluginDeploymentParam); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to ensure plugin v2")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	respData := &protoBackend.PluginInstallResp_Data{
		WorkflowId: workflowID,
	}

	logger.G.Biz(rCtx).With("workflow-id", workflowID).Info("launched install plugin workflow")

	return respData, nil
}

func (h *handler) installPlugin(rCtx restserver.IContext, pluginDeploymentParam []*types.PluginDeploymentParam) (string, error) {
	if err := h.preparePluginDeploymentParams(rCtx, pluginDeploymentParam...); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to prepare plugin installation")
		return "", err
	}

	pluginDeployments, hostIDs, bizIDs, err := types.NewPluginDeploymentsByParams(
		rCtx.TenantID(), types.PluginDeploymentTransferOptionsAll(), pluginDeploymentParam...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to install plugin, failed to generate plugin deployments.")
		return "", err
	}

	workflowID, err := h.pluginMgrIface.LaunchInstallPlugin(rCtx, types.InstallPluginParam{
		Type:              types.PluginWorkflowTypeInstall,
		HostIDs:           hostIDs,
		BizIDs:            bizIDs,
		Operator:          rCtx.BKUsername(),
		PluginDeployments: pluginDeployments,
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to install plugin.")
		return "", err
	}

	return workflowID, nil
}

// ensurePluginV2 launches the plugin v2 workflow for the compatibility mode plugin deployments.
func (h *handler) ensurePluginV2(rCtx restserver.IContext, renewSubConfig bool, params []*types.PluginDeploymentParam) error {
	// The params are copied, so neither the selection nor the sub config policy leaks to the caller.
	selected := compatibilityModeParams(params)
	if err := h.preparePluginDeploymentParams(rCtx, selected...); err != nil {
		return err
	}
	// The prepare step may drop the compatibility mode, so the params are selected again.
	selected = compatibilityModeParams(selected)
	if len(selected) == 0 {
		logger.G.Biz(rCtx).Info("no plugin deployment need to ensure plugin v2.")
		return nil
	}

	for _, param := range selected {
		param.RenewSubConfig = renewSubConfig
	}

	tenantID := rCtx.TenantID()
	pluginDeployments, hostIDs, bizIDs, err := types.NewPluginDeploymentsByParams(tenantID, types.PluginDeploymentTransferOptionsAll(), selected...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to ensure plugin, failed to generate plugin deployments.")
		return err
	}

	_, err = h.pluginMgrIface.LaunchPluginEnsurePluginV2(rCtx, types.InstallPluginParam{
		Type:              types.PluginWorkflowTypePluginEnsureV2,
		HostIDs:           hostIDs,
		BizIDs:            bizIDs,
		Operator:          rCtx.BKUsername(),
		PluginDeployments: pluginDeployments,
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to ensure plugin v2.")
		return err
	}

	return nil
}

// compatibilityModeParams returns the copied params that enable the plugin compatibility mode.
func compatibilityModeParams(params []*types.PluginDeploymentParam) []*types.PluginDeploymentParam {
	selected := make([]*types.PluginDeploymentParam, 0, len(params))
	for _, param := range params {
		if param == nil || !param.EnableCompatibilityMode {
			continue
		}

		copied := *param
		selected = append(selected, &copied)
	}

	return selected
}

func (h *handler) preparePluginDeploymentParams(rCtx restserver.IContext, params ...*types.PluginDeploymentParam) error {
	if len(params) == 0 {
		return nil
	}

	hostIDs := conv.SliceToSlice(params, func(param *types.PluginDeploymentParam) int64 {
		return param.HostID
	})
	hosts, _, err := h.daoHost.ListHost(rCtx, types.UnlimitedPage(), &types.HostCondition{
		StaticExactInclude: &types.HostStaticExactFields{HostID: hostIDs},
	})
	if err != nil {
		return fmt.Errorf("failed to list plugin hosts: %w", err)
	}
	hostMap, err := conv.SliceToMap(hosts, func(host *types.Host) int64 {
		return host.HostID
	})
	if err != nil {
		return fmt.Errorf("failed to convert plugin hosts to map: %w", err)
	}

	pluginNames := conv.SliceToSlice(params, func(param *types.PluginDeploymentParam) string {
		return param.PluginName
	})
	plugins, _, err := h.daoPlugin.ListPlugins(rCtx, types.UnlimitedPage(), &types.PluginCondition{
		ExactInclude: &types.PluginExactFields{Name: conv.SliceUnique(pluginNames)},
	})
	if err != nil {
		return fmt.Errorf("failed to list plugins: %w", err)
	}
	pluginMap, err := conv.SliceToMap(plugins, func(plugin *types.Plugin) string {
		return plugin.Name
	})
	if err != nil {
		return fmt.Errorf("failed to convert plugins to map: %w", err)
	}

	releaseKeys := make([]types.ReleasePluginKey, len(params))
	for idx, param := range params {
		host, ok := hostMap[param.HostID]
		if !ok {
			return fmt.Errorf("plugin host not found, host-id(%d)", param.HostID)
		}
		if host.Dynamic == nil {
			return fmt.Errorf("plugin host dynamic info is empty, host-id(%d)", param.HostID)
		}
		plugin, ok := pluginMap[param.PluginName]
		if !ok {
			return fmt.Errorf("plugin not found, plugin-name(%s)", param.PluginName)
		}
		pluginPlatform := platform.NewPlatform(host.Dynamic.NodeOsType, host.Dynamic.NodeCPUArch)
		releaseKeys[idx] = types.ReleasePluginKey{
			Generation: host.Dynamic.NodeGeneration,
			Platform:   pluginPlatform,
			Name:       plugin.PkgName,
			Version:    param.Version,
		}
	}

	if err := h.resolvePluginReleaseVersions(rCtx, releaseKeys); err != nil {
		return err
	}
	releases, err := h.fetchPluginReleases(rCtx, conv.SliceUnique(releaseKeys))
	if err != nil {
		return err
	}

	for idx, param := range params {
		releaseKey := releaseKeys[idx]
		param.Version = releaseKey.Version
		param.EnableCompatibilityMode = param.EnableCompatibilityMode && !releases[releaseKey].LegacyConflict
	}

	return nil
}

func (h *handler) resolvePluginReleaseVersions(rCtx restserver.IContext, keys []types.ReleasePluginKey) error {
	defaultKeys := make([]types.ReleasePluginKey, 0)
	for _, key := range keys {
		if key.Version != "" {
			continue
		}
		defaultKeys = append(defaultKeys, key)
	}

	defaultVersions := make(map[types.ReleasePluginKey]string)
	for _, key := range conv.SliceUnique(defaultKeys) {
		version, err := h.fileHandler.GetReleasePluginDefaultVersion(rCtx, key.Name, key.Generation, key.Platform)
		if err != nil {
			return fmt.Errorf("failed to get default plugin version, plugin-pkg-name(%s): %w", key.Name, err)
		}
		defaultVersions[key] = version
	}

	for idx, key := range keys {
		if key.Version != "" {
			continue
		}
		keys[idx].Version = defaultVersions[key]
	}

	return nil
}

func (h *handler) fetchPluginReleases(rCtx restserver.IContext, keys []types.ReleasePluginKey) (map[types.ReleasePluginKey]*types.ReleasePlugin, error) {
	releases := make(map[types.ReleasePluginKey]*types.ReleasePlugin)
	for _, key := range keys {
		release, err := h.fileHandler.GetReleasePlugin(rCtx, key)
		if err != nil {
			return nil, fmt.Errorf("failed to get plugin release, plugin-pkg-name(%s), version(%s): %w", key.Name, key.Version, err)
		}
		releases[key] = release
	}

	return releases, nil
}

func (h *handler) resolvePluginCompatibilityModePolicy(rCtx restserver.IContext) compatibility.Policy {
	if h.storageGlobalSettings == nil {
		return compatibility.ParseStoredPolicy("", rCtx)
	}

	value, err := h.storageGlobalSettings.GetGlobalSetting(rCtx, globalsettings.PluginCompatibilityModePolicy)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).
			With("setting-name", globalsettings.PluginCompatibilityModePolicy).
			Warn("failed to get plugin compatibility mode policy, use default policy.")

		return compatibility.ParseStoredPolicy("", rCtx)
	}

	return compatibility.ParseStoredPolicy(value, rCtx)
}

func applyPluginCompatibilityModePolicy(policy compatibility.Policy, tenantID string, params ...*types.PluginDeploymentParam) {
	for _, param := range params {
		if param == nil {
			continue
		}

		param.EnableCompatibilityMode = compatibility.DecideCompatibilityMode(policy, tenantID, param.BizID, param.PluginName)
	}
}
