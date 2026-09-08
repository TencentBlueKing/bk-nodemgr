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

package pkg

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authProvider "github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth/provider"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	// maxMemoFields is the maximum number of memo fields (Description, Scenario, DescriptionEn, ScenarioEn).
	maxMemoFields = 4
)

// ListReleasePlugin list plugin.
func (h *handler) ListReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleasePluginListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())
	exactIncludeCond := req.ConvertExactIncludeConditionsToTypes()
	exactIncludeCond.Generation = append(exactIncludeCond.Generation, gen)
	cond := &types.ReleaseCondition{
		ExactInclude: exactIncludeCond,
	}
	// Check permission and narrow by authorized plugin names.
	requestedNames := exactIncludeCond.Name
	narrowedNames, scopeIsAny, authErr := h.narrowAuthorizedPackageNames(rCtx, requestedNames, types.ReleaseTypePlugin)
	if authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to list plugin, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}
	cond = narrowReleaseCondition(cond, narrowedNames, scopeIsAny, types.ReleaseTypePlugin)

	// only count.
	if req.GetOnlyCount() {
		num, err := h.daoReleasePlugin.CountReleasePlugin(rCtx, cond)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin. failed to count host")
			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.PackageReleasePluginListResp)
		resp.ConvertReleasePluginsFromTypes(num, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	hosts, num, err := h.daoReleasePlugin.ListReleasePlugin(rCtx, page, cond)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PackageReleasePluginListResp)
	resp.ConvertReleasePluginsFromTypes(num, hosts)

	return resp.GetData(), nil
}

// ListReleasePluginBrief list plugin brief.
func (h *handler) ListReleasePluginBrief(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleasePluginListBriefReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin brief, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())
	exactIncludeCond := req.ConvertExactIncludeConditionsToTypes()
	exactIncludeCond.Generation = append(exactIncludeCond.Generation, gen)
	cond := &types.ReleaseCondition{
		ExactInclude: exactIncludeCond,
	}
	// Check permission and narrow by authorized plugin names.
	requestedNames := exactIncludeCond.Name
	narrowedNames, scopeIsAny, authErr := h.narrowAuthorizedPackageNames(rCtx, requestedNames, types.ReleaseTypePlugin)
	if authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to list plugin brief, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}
	cond = narrowReleaseCondition(cond, narrowedNames, scopeIsAny, types.ReleaseTypePlugin)

	// only count.
	if req.GetOnlyCount() {
		num, err := h.daoReleasePlugin.CountReleasePlugin(rCtx, cond)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin brief. failed to count host")
			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.PackageReleasePluginListBriefResp)
		resp.ConvertReleasePluginsFromTypes(num, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin brief, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	hosts, num, err := h.daoReleasePlugin.ListReleasePlugin(rCtx, page, cond)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin brief")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PackageReleasePluginListBriefResp)
	resp.ConvertReleasePluginsFromTypes(num, hosts)

	return resp.GetData(), nil
}

// DistinctReleasePlugin distinct plugin releases.
func (h *handler) DistinctReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleasePluginDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct plugin release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	gen := types.Generation(req.GetGeneration())
	exactIncludeCond := req.ConvertExactIncludeConditionsToTypes()
	exactIncludeCond.Generation = append(exactIncludeCond.Generation, gen)
	condition := &types.ReleaseCondition{
		ExactInclude: exactIncludeCond,
	}
	distinctField := types.ReleaseDistinctField{
		OSType:  req.GetDistinctField().GetOsType(),
		CPUArch: req.GetDistinctField().GetCpuArch(),
		Name:    req.GetDistinctField().GetName(),
		Version: req.GetDistinctField().GetVersion(),
	}

	result, err := h.daoReleasePlugin.DistinctReleasePlugin(rCtx, distinctField, condition)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct plugin release")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PackageReleasePluginDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}

// EnableReleasePlugin enable plugin.
func (h *handler) EnableReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleasePluginEnableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to enable plugin, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	pluginPkgName, gen, plat, version := req.GetIdentifier()

	// Check permission.
	resources := authProvider.BuildPackageResources(pluginPkgName)
	if err := h.authorizer.Check(rCtx, auth.ActionPackageManage, resources); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to enable plugin, permission denied")
		return nil, err
	}

	key := types.ReleasePluginKey{
		Name:       pluginPkgName,
		Generation: gen,
		Platform:   plat,
		Version:    version,
	}

	exist, err := h.daoReleasePlugin.ExistReleasePlugin(rCtx, key)
	if err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "platform", plat, "version", version).
			Error("failed to enable plugin")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	if !exist {
		logger.G.Biz(rCtx).
			With("gen", gen, "platform", plat, "version", version).
			Error("failed to enable plugin. plugin not exist")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, errors.New("plugin not exist"))
	}

	if err := h.daoReleasePlugin.EnableReleasePlugin(rCtx, key); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "platform", plat, "version", version).
			Error("failed to enable plugin")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	if err := h.createDefaultPlugin(rCtx, key); err != nil {
		logger.G.Biz(rCtx).WithErr(err).With("gen", gen, "platform", plat, "version", version).
			Error("failed to check and create default plugin for all tenants")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record release plugin event.
	h.recordPluginEvent(rCtx, gen, pluginPkgName, version, plat, types.PackageEventTypeEnable)

	logger.G.Biz(rCtx).
		With("gen", gen, "platform", plat, "version", version).
		Info("enabled plugin")

	resp := new(protoBackend.PackageReleasePluginEnableResp)

	return resp.GetData(), nil
}

// createDefaultPlugin creates default plugin.
func (h *handler) createDefaultPlugin(nCtx contextx.IContext, key types.ReleasePluginKey) error {
	exist, err := h.daoPlugin.ExistDefaultPluginByPluginPkgName(nCtx, key.Name)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("plugin_pkg_name", key.Name).
			Error("failed to check plugin exist for tenant")

		return fmt.Errorf("failed to check plugin exist, plugin(%s): %w", key.Name, err)
	}

	if exist {
		return nil
	}

	plugin, err := h.daoReleasePlugin.GetReleasePlugin(nCtx, key)
	if err != nil {
		logger.G.Biz(nCtx).
			WithErr(err).
			With("key", key).
			Error("failed to get release plugin")

		return fmt.Errorf("failed to get release plugin, key(%v): %w", key, err)
	}

	defaultPlugin := &types.Plugin{
		Name:    key.Name,
		PkgName: key.Name,
		Group:   types.PluginGroupDefault,
		Memo:    buildPluginMemo(plugin),
	}

	if err := h.daoPlugin.CreatePlugin(nCtx, defaultPlugin); err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("key", key).
			Error("failed to create default plugin for tenant")

		return fmt.Errorf("failed to create default plugin by plugin-pkg-name(%s): %w", key.Name, err)
	}

	return nil
}

// buildPluginMemo builds memo string from plugin description and scenario fields, only including non-empty fields.
func buildPluginMemo(plugin *types.ReleasePlugin) string {
	memoParts := make([]string, 0, maxMemoFields)
	if plugin.Description != "" {
		memoParts = append(memoParts, fmt.Sprintf("描述: %s", plugin.Description))
	}
	if plugin.Scenario != "" {
		memoParts = append(memoParts, fmt.Sprintf("场景: %s", plugin.Scenario))
	}
	if plugin.DescriptionEn != "" {
		memoParts = append(memoParts, fmt.Sprintf("Description: %s", plugin.DescriptionEn))
	}
	if plugin.ScenarioEn != "" {
		memoParts = append(memoParts, fmt.Sprintf("Scene: %s", plugin.ScenarioEn))
	}

	return strings.Join(memoParts, "\n")
}

// DisableReleasePlugin disable plugin.
func (h *handler) DisableReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleasePluginDisableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to disable plugin, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	name, gen, plat, version := req.GetIdentifier()

	// Check permission.
	resources := authProvider.BuildPackageResources(name)
	if err := h.authorizer.Check(rCtx, auth.ActionPackageManage, resources); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to disable plugin, permission denied")
		return nil, err
	}

	key := types.ReleasePluginKey{
		Name:       name,
		Generation: gen,
		Platform:   plat,
		Version:    version,
	}
	if err := h.daoReleasePlugin.DisableReleasePlugin(rCtx, key); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "platform", plat, "version", version).
			Error("failed to disable plugin")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record release plugin event.
	h.recordPluginEvent(rCtx, gen, name, version, plat, types.PackageEventTypeDisable)

	logger.G.Biz(rCtx).
		With("gen", gen, "platform", plat, "version", version).
		Info("disable plugin")

	resp := new(protoBackend.PackageReleasePluginDisableResp)

	return resp.GetData(), nil
}

func (h *handler) SetHiddenReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleasePluginSetHiddenReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set hidden plugin, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	name, gen, plat, version := req.GetIdentifier()

	// Check permission.
	resources := authProvider.BuildPackageResources(name)
	if err := h.authorizer.Check(rCtx, auth.ActionPackageManage, resources); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set hidden plugin, permission denied")
		return nil, err
	}

	key := types.ReleasePluginKey{
		Name:       name,
		Generation: gen,
		Platform:   plat,
		Version:    version,
	}
	if err := h.daoReleasePlugin.SetHiddenReleasePlugin(rCtx, key, req.GetIsHidden()); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "platform", plat, "version", version, "is_hidden", req.GetIsHidden()).
			Error("failed to set hidden plugin")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	logger.G.Biz(rCtx).
		With("gen", gen, "platform", plat, "version", version, "is_hidden", req.GetIsHidden()).
		Info("set hidden plugin")

	resp := new(protoBackend.PackageReleasePluginSetHiddenResp)

	return resp.GetData(), nil
}

func (h *handler) SetAsDefaultReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleasePluginSetAsDefaultReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set default plugin, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	name, gen, plat, version := req.GetIdentifier()

	// Check permission.
	resources := authProvider.BuildPackageResources(name)
	if err := h.authorizer.Check(rCtx, auth.ActionPackageManage, resources); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set default plugin, permission denied")
		return nil, err
	}

	key := types.ReleasePluginKey{
		Name:       name,
		Generation: gen,
		Platform:   plat,
		Version:    version,
	}
	if err := h.daoReleasePlugin.SetAsDefaultReleasePlugin(rCtx, key); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "platform", plat, "version", version).
			Error("failed to set as default plugin")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record release plugin event.
	h.recordPluginEvent(rCtx, gen, name, version, plat, types.PackageEventTypeSetAsDefault)

	logger.G.Biz(rCtx).
		With("gen", gen, "platform", plat, "version", version).
		Info("set as default plugin")

	resp := new(protoBackend.PackageReleasePluginSetAsDefaultResp)

	return resp.GetData(), nil
}

func (h *handler) CancelAsDefaultReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleasePluginCancelAsDefaultReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to cancel default plugin, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	name, gen, plat, version := req.GetIdentifier()

	// Check permission.
	resources := authProvider.BuildPackageResources(name)
	if err := h.authorizer.Check(rCtx, auth.ActionPackageManage, resources); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to cancel default plugin, permission denied")
		return nil, err
	}

	key := types.ReleasePluginKey{
		Name:       name,
		Generation: gen,
		Platform:   plat,
		Version:    version,
	}
	if err := h.daoReleasePlugin.CancelAsDefaultReleasePlugin(rCtx, key); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "platform", plat, "version", version).
			Error("failed to cancel as default plugin")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record release plugin event.
	h.recordPluginEvent(rCtx, gen, name, version, plat, types.PackageEventTypeCancelAsDefault)

	logger.G.Biz(rCtx).
		With("gen", gen, "platform", plat, "version", version).
		Info("canceled as default plugin")

	resp := new(protoBackend.PackageReleasePluginCancelAsDefaultResp)

	return resp.GetData(), nil
}

func (h *handler) DeleteReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleasePluginDeleteReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete plugin, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	name, gen, plat, version := req.GetIdentifier()

	// Check permission.
	resources := authProvider.BuildPackageResources(name)
	if err := h.authorizer.Check(rCtx, auth.ActionPackageManage, resources); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to delete plugin, permission denied")
		return nil, err
	}

	key := types.ReleasePluginKey{
		Name:       name,
		Generation: gen,
		Platform:   plat,
		Version:    version,
	}
	if err := h.daoReleasePlugin.DeleteReleasePlugin(rCtx, key); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("gen", gen, "platform", plat, "version", version).
			Error("failed to delete plugin")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	// record release plugin event.
	h.recordPluginEvent(rCtx, gen, name, version, plat, types.PackageEventTypeDelete)

	logger.G.Biz(rCtx).
		With("gen", gen, "platform", plat, "version", version).
		Info("deleted plugin")

	resp := new(protoBackend.PackageReleasePluginDeleteResp)

	return resp.GetData(), nil
}

func (h *handler) recordPluginEvent(rCtx restserver.IContext, gen types.Generation, name, version string,
	plat platfmt.Platform, eventType types.PackageEventType) {

	h.recordPackageEvents(rCtx, &types.PackageEvent{
		Name:        name,
		EventType:   eventType,
		ReleaseType: types.ReleaseTypePlugin,
		Generation:  gen,
		Version:     version,
		OSType:      plat.OS,
		CPUArch:     plat.Arch,
		OperateTime: time.Now(),
		Operator:    rCtx.Data().GetLoginName(),
	})
}

// GetConfigVariablesReleasePlugin gets the config variables of a plugin release.
func (h *handler) GetConfigVariablesReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleasePluginGetConfigVariablesReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get config variables of plugin release, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	cond := req.ConvertConditionsToTypes()
	plugins, _, err := h.daoReleasePlugin.ListReleasePlugin(rCtx, types.UnlimitedPage(), cond)
	if err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("name", req.GetName(), "gen", req.GetGeneration(), "platform", req.GetPlatforms(), "version", req.GetVersion()).
			Error("failed to get release plugin")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PackageReleasePluginGetConfigVariablesResp)
	if err := resp.ConvertConfigVariablesFromTypes(plugins); err != nil {
		logger.G.Biz(rCtx).
			WithErr(err).
			With("name", req.GetName(), "gen", req.GetGeneration(), "platform", req.GetPlatforms(), "version", req.GetVersion()).
			Error("failed to convert config variables of release plugin")

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	return resp.GetData(), nil
}
