/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pkg

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
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

	page, err := req.ConvertPageToTypes(maxReleaseLimit)
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

// EnableReleasePlugin enable plugin.
func (h *handler) EnableReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleasePluginEnableReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to enable plugin, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	pluginPkgName, gen, plat, version := req.GetIdentifier()
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

	if err := h.initDefaultPluginForAllTenants(rCtx, key); err != nil {
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

func (h *handler) initDefaultPluginForAllTenants(rCtx restserver.IContext, key types.ReleasePluginKey) error {
	tenants, err := h.daoTenant.ListAllEnabledTenants(rCtx)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list all tenants")
		return err
	}

	gp := gopool.NewPool()
	for idx := range tenants {
		tenant := tenants[idx]
		nCtx := contextx.New(rCtx, contextx.WithTenantID(tenant.ID))

		gp.Go(func() error {
			return h.createDefaultPluginForTenant(nCtx, tenant.ID, key)
		})
	}

	if err := gp.Wait(); err != nil {
		return fmt.Errorf("failed to create default plugin for all tenants: %w", err)
	}

	return nil
}

// createDefaultPluginForTenant creates default plugin for a single tenant.
func (h *handler) createDefaultPluginForTenant(nCtx contextx.IContext, tenantID string, key types.ReleasePluginKey) error {
	exist, err := h.daoPlugin.ExistDefaultPluginByPluginPkgName(nCtx, key.Name)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("tenant_id", tenantID, "plugin_pkg_name", key.Name).
			Error("failed to check plugin exist for tenant")

		return fmt.Errorf("failed to check plugin exist, tenant-id(%s) plugin(%s): %w", tenantID, key.Name, err)
	}

	if exist {
		return nil
	}

	plugin, err := h.daoReleasePlugin.GetReleasePlugin(nCtx, key)
	if err != nil {
		logger.G.Biz(nCtx).
			WithErr(err).
			With("tenant_id", tenantID, "key", key).
			Error("failed to get release plugin")

		return fmt.Errorf("failed to get release plugin, key(%v): %w", key, err)
	}

	defaultPlugin := &types.Plugin{
		TenantID: tenantID,
		Name:     key.Name,
		PkgName:  key.Name,
		Group:    types.PluginGroupDefault,
		Memo:     buildPluginMemo(plugin),
	}

	if err := h.daoPlugin.CreatePlugin(nCtx, defaultPlugin); err != nil {
		logger.G.Biz(nCtx).WithErr(err).With("tenant_id", tenantID, "key", key).
			Error("failed to create default plugin for tenant")

		return fmt.Errorf("failed to create default plugin for tenant(%s) by plugin-pkg-name(%s): %w", tenantID, key.Name, err)
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

func (h *handler) SetAsDefaultReleasePlugin(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PackageReleasePluginSetAsDefaultReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to set default plugin, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	name, gen, plat, version := req.GetIdentifier()
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

	operator := rCtx.Data().GetLoginName()
	go func() {
		if err := h.daoPackageEvent.CreateManyPackageEvent(contextx.Background(),
			&types.PackageEvent{
				Name:        name,
				EventType:   eventType,
				ReleaseType: types.ReleaseTypePlugin,
				Generation:  gen,
				Version:     version,
				OSType:      plat.OS,
				CPUArch:     plat.Arch,
				OperateTime: time.Now(),
				Operator:    operator,
			}); err != nil {
			logger.G.Sys().WithErr(err).With("event-type", eventType).Error("failed to record package event")
		}
	}()
}
