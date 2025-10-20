/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package plugin

import (
	"fmt"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler plugin handler interface.
type IHandler interface {
	// Create create a new plugin.
	Create(nCtx contextx.IContext, pluginType types.PluginType, plugin *types.Plugin) error

	// Count count plugins by conditions.
	Count(nCtx contextx.IContext, pluginType types.PluginType, opts ...OptFn) (int64, error)

	// List lists plugins by page and conditions.
	List(nCtx contextx.IContext, pluginType types.PluginType, page types.Page, opts ...OptFn) ([]*types.Plugin, int64, error)

	// Get gets a plugin by conditions.
	Get(nCtx contextx.IContext, pluginType types.PluginType, opts ...OptFn) (*types.Plugin, error)

	// Update update a plugin by conditions.
	Update(nCtx contextx.IContext, pluginType types.PluginType, plugin *types.Plugin) error
}

// Handler this is a Handler to operate plugin table.
type Handler struct {
	client *mongo.Database
	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

func (h *Handler) tenantDao(tenantID string, pluginType types.PluginType) *dao {
	tableName := TableName(tenantID, string(pluginType))

	if d, ok := h.daoMap.Load(tableName); ok {
		return d.(*dao) // nolint: forcetypeassert
	}

	newDaoClient := newDao(h.client, tableName)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).With("table-name", tableName).Warn("failed to ensure plugin indexes")
	}

	d, _ := h.daoMap.LoadOrStore(tableName, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao) // nolint: forcetypeassert
}

// New create a new plugin handler.
func New(client *mongo.Database) *Handler {
	return &Handler{
		client: client,
		daoMap: sync.Map{},
	}
}

// Create create a new plugin.
func (h *Handler) Create(nCtx contextx.IContext, pluginType types.PluginType, plugin *types.Plugin) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	data := convPluginFromTypes(plugin)

	if err := h.tenantDao(nCtx.TenantID(), pluginType).Create(nCtx, data); err != nil {
		return fmt.Errorf("failed to create plugin, err: %w", err)
	}

	return nil
}

func convPluginFromTypes(plugin *types.Plugin) *Plugin {
	data := &Plugin{
		TenantID: plugin.TenantID,
		PluginID: plugin.PluginID,
		Static: pluginStatic{
			Info: ProcessInfo{
				Pid:         plugin.Static.Info.Pid,
				Version:     plugin.Static.Info.Version,
				AgentID:     plugin.Static.Info.AgentID,
				Trusteeship: plugin.Static.Info.Trusteeship,
				Status:      string(plugin.Static.Info.Status),
			},
			Identity: processIdentity{
				Name:       plugin.Static.Identity.Name,
				SetupPath:  plugin.Static.Identity.SetupPath,
				PidPath:    plugin.Static.Identity.PidPath,
				ConfigPath: plugin.Static.Identity.ConfigPath,
				LogPath:    plugin.Static.Identity.LogPath,
				User:       plugin.Static.Identity.User,
			},
			Controller: processController{
				StartCmd:   plugin.Static.Controller.StartCmd,
				StopCmd:    plugin.Static.Controller.StopCmd,
				RestartCmd: plugin.Static.Controller.RestartCmd,
				ReloadCmd:  plugin.Static.Controller.ReloadCmd,
				KillCmd:    plugin.Static.Controller.KillCmd,
				VersionCmd: plugin.Static.Controller.VersionCmd,
				HealthCmd:  plugin.Static.Controller.HealthCmd,
			},
			Resource: processResource{
				CPU: plugin.Static.Resource.CPULimitPercent,
				Mem: plugin.Static.Resource.MemLimitPercent,
			},
			MonitorPolicy: processMonitorPolicy{
				AutoType:       string(plugin.Static.MonitorPolicy.AutoType),
				StartCheckSecs: plugin.Static.MonitorPolicy.StartCheckSecs,
				StopCheckSecs:  plugin.Static.MonitorPolicy.StopCheckSecs,
				OpTimeoutSecs:  plugin.Static.MonitorPolicy.OpTimeoutSecs,
			},
		},
		Dynamic: pluginDynamic{
			Name:       plugin.Dynamic.Name,
			Type:       string(plugin.Dynamic.Type),
			Generation: int64(plugin.Dynamic.Generation),
			Platform:   convPlatformFromTypes(plugin.Dynamic.Platform),
			Version:    plugin.Dynamic.Version,
			HostID:     plugin.Dynamic.HostID,
			Status:     string(plugin.Dynamic.Status),
		},
	}

	return data
}

func convPlatformFromTypes(p platfmt.Platform) Platform {
	return Platform{
		OS:   string(p.OS),
		Arch: string(p.Arch),
	}
}

// Count count host by conditions.
func (h *Handler) Count(nCtx contextx.IContext, pluginType types.PluginType, opts ...OptFn) (int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.tenantDao(nCtx.TenantID(), pluginType).Count(nCtx, filter)
}

// List list plugin by page and conditions.
func (h *Handler) List(nCtx contextx.IContext, pluginType types.PluginType, page types.Page, opts ...OptFn) (
	[]*types.Plugin, int64, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.tenantDao(nCtx.TenantID(), pluginType).Count(nCtx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	data, err := h.tenantDao(nCtx.TenantID(), pluginType).List(nCtx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	plugins := make([]*types.Plugin, len(data))
	for idx, host := range data {
		plugins[idx] = convertPluginToTypes(host)
	}

	return plugins, num, nil
}

// Get get a plugin.
func (h *Handler) Get(nCtx contextx.IContext, pluginType types.PluginType, opts ...OptFn) (*types.Plugin, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	data, err := h.tenantDao(nCtx.TenantID(), pluginType).Get(nCtx, filter)
	if err != nil {
		return nil, err
	}

	return convertPluginToTypes(data), nil
}

func convertPluginToTypes(data *Plugin) *types.Plugin {
	plugin := &types.Plugin{
		PluginID: data.PluginID,
		TenantID: data.TenantID,
		Static: types.PluginStatic{
			Info: types.ProcessInfo{
				Pid:         data.Static.Info.Pid,
				Version:     data.Static.Info.Version,
				AgentID:     data.Static.Info.AgentID,
				Trusteeship: data.Static.Info.Trusteeship,
				Status:      types.ProcessStatus(data.Static.Info.Status),
			},
			Identity: types.ProcessIdentity{
				Name:       data.Static.Identity.Name,
				SetupPath:  data.Static.Identity.SetupPath,
				PidPath:    data.Static.Identity.PidPath,
				ConfigPath: data.Static.Identity.ConfigPath,
				LogPath:    data.Static.Identity.LogPath,
				User:       data.Static.Identity.User,
			},
			Controller: types.ProcessController{
				StartCmd:   data.Static.Controller.StartCmd,
				StopCmd:    data.Static.Controller.StopCmd,
				RestartCmd: data.Static.Controller.RestartCmd,
				ReloadCmd:  data.Static.Controller.ReloadCmd,
				KillCmd:    data.Static.Controller.KillCmd,
				VersionCmd: data.Static.Controller.VersionCmd,
				HealthCmd:  data.Static.Controller.HealthCmd,
			},
			Resource: types.ProcessResource{
				CPULimitPercent: data.Static.Resource.CPU,
				MemLimitPercent: data.Static.Resource.Mem,
			},
			MonitorPolicy: types.ProcessMonitorPolicy{
				AutoType:       types.ProcessAutoType(data.Static.MonitorPolicy.AutoType),
				StartCheckSecs: data.Static.MonitorPolicy.StartCheckSecs,
				StopCheckSecs:  data.Static.MonitorPolicy.StopCheckSecs,
				OpTimeoutSecs:  data.Static.MonitorPolicy.OpTimeoutSecs,
			},
		},
		Dynamic: types.PluginDynamic{
			Name:       data.Dynamic.Name,
			Type:       types.PluginType(data.Dynamic.Type),
			Generation: types.Generation(data.Dynamic.Generation),
			Platform:   convPlatformToTypes(data.Dynamic.Platform),
			Version:    data.Dynamic.Version,
			HostID:     data.Dynamic.HostID,
			Status:     types.ProcessStatus(data.Dynamic.Status),
		},
	}

	return plugin
}

func convPlatformToTypes(data Platform) platfmt.Platform {
	return platfmt.Platform{
		OS:   criteria.OSType(data.OS),
		Arch: criteria.CPUArch(data.Arch),
	}
}
