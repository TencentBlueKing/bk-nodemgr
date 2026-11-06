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

package processv2

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

// IHandler process v2 handler interface.
type IHandler interface {
	// Create create a new v2 process.
	Create(nCtx contextx.IContext, process *types.Process) error

	// List lists v2 processes by page and conditions.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.Process, int64, error)

	// Get gets a v2 process by conditions.
	Get(nCtx contextx.IContext, opts ...OptFn) (*types.Process, error)

	// Exist check a v2 process exist by host id and plugin name.
	Exist(nCtx contextx.IContext, hostID int64, pluginName string) (bool, error)

	// Delete delete a v2 process by host id and plugin name.
	Delete(nCtx contextx.IContext, hostID int64, pluginName string) error

	// Update update a v2 process by host id and plugin name.
	Update(nCtx contextx.IContext, hostID int64, pluginName string, process *types.Process) error

	// UpdateInfo update the process info of a v2 process by host id and plugin name.
	UpdateInfo(nCtx contextx.IContext, hostID int64, pluginName string, processInfo *types.ProcessInfo) error

	// UpdateManyInfo batch update the process info of v2 processes.
	UpdateManyInfo(nCtx contextx.IContext, processInfoDeltas []*types.ProcessInfoDelta) error
}

var _ IHandler = &Handler{}

// Handler this is a Handler to operate process v2 table.
type Handler struct {
	client *mongo.Database
	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

// New create a new process v2 handler.
func New(client *mongo.Database) *Handler {
	return &Handler{
		client: client,
		daoMap: sync.Map{},
	}
}

func (h *Handler) tenantDao(tenantID string) *dao {
	tableName := TableName(tenantID)

	if d, ok := h.daoMap.Load(tableName); ok {
		return d.(*dao) // nolint: forcetypeassert
	}

	newDaoClient := newDao(h.client, tableName)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).With("table-name", tableName).Warn("failed to ensure process v2 indexes")
	}

	d, _ := h.daoMap.LoadOrStore(tableName, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao) // nolint: forcetypeassert
}

// Create create a new v2 process.
func (h *Handler) Create(nCtx contextx.IContext, process *types.Process) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	data := convProcessFromTypes(process)

	if err := h.tenantDao(nCtx.TenantID()).Create(nCtx, data); err != nil {
		return fmt.Errorf("failed to create process v2, err: %w", err)
	}

	return nil
}

// List lists v2 processes by page and conditions.
func (h *Handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.Process, int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.tenantDao(nCtx.TenantID()).Count(nCtx, filter)
	if err != nil {
		return nil, 0, err
	}

	data, err := h.tenantDao(nCtx.TenantID()).List(nCtx, filter, base.ParsePage(page))
	if err != nil {
		return nil, 0, err
	}

	processes := make([]*types.Process, len(data))
	for idx, process := range data {
		processes[idx] = convertProcessToTypes(process)
	}

	return processes, num, nil
}

// Get gets a v2 process by conditions.
func (h *Handler) Get(nCtx contextx.IContext, opts ...OptFn) (*types.Process, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	data, err := h.tenantDao(nCtx.TenantID()).Get(nCtx, filter)
	if err != nil {
		return nil, err
	}

	return convertProcessToTypes(data), nil
}

// Exist check a v2 process exist by host id and plugin name.
func (h *Handler) Exist(nCtx contextx.IContext, hostID int64, pluginName string) (bool, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return false, fmt.Errorf("failed to check tenant id: %v", err)
	}

	filter := base.AliveFilter()
	for _, opt := range []OptFn{WithHostID(hostID), WithPluginName(pluginName)} {
		filter = opt(filter)
	}

	exist, err := h.tenantDao(nCtx.TenantID()).Exist(nCtx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check process v2 exist: %v", err)
	}

	return exist, nil
}

// Delete delete a v2 process by host id and plugin name.
func (h *Handler) Delete(nCtx contextx.IContext, hostID int64, pluginName string) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return fmt.Errorf("failed to check tenant id: %v", err)
	}

	filter := base.AliveFilter()
	for _, opt := range []OptFn{WithHostID(hostID), WithPluginName(pluginName)} {
		filter = opt(filter)
	}

	if err := h.tenantDao(nCtx.TenantID()).HardDelete(nCtx, filter); err != nil {
		logger.G.Sys().With("host-id", hostID).With("plugin-name", pluginName).WithErr(err).
			Error("failed to delete process v2")

		return fmt.Errorf("failed to delete process v2: %v", err)
	}

	return nil
}

// Update update a v2 process by host id and plugin name.
func (h *Handler) Update(nCtx contextx.IContext, hostID int64, pluginName string, process *types.Process) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return fmt.Errorf("failed to check tenant id: %v", err)
	}

	if process == nil {
		return fmt.Errorf("process is nil")
	}

	filter := base.AliveFilter()
	for _, opt := range []OptFn{WithHostID(hostID), WithPluginName(pluginName)} {
		filter = opt(filter)
	}

	data := convProcessFromTypes(process)

	updates := []*base.DocumentFieldUpdate{
		{
			Filter: filter,
			Fields: map[string]any{
				FieldKeyHostID:        data.HostID,
				FieldKeyBizID:         data.BizID,
				FieldKeyPluginName:    data.Name,
				FieldKeyGroup:         data.Group,
				FieldKeyPkgName:       data.PluginPkgName,
				FieldKeyGeneration:    data.Generation,
				FieldKeyPlatform:      data.Platform,
				FieldKeyBindIP:        data.BindIP,
				FieldKeyBindPort:      data.BindPort,
				FieldKeyInfo:          data.Info,
				FieldKeyIdentity:      data.Identity,
				FieldKeyController:    data.Controller,
				FieldKeyResource:      data.Resource,
				FieldKeyMonitorPolicy: data.MonitorPolicy,
			},
		},
	}

	if err := h.tenantDao(nCtx.TenantID()).UpdateOneFieldBulk(nCtx, updates); err != nil {
		return fmt.Errorf("failed to update process v2: %v", err)
	}

	return nil
}

// UpdateInfo update the process info of a v2 process by host id and plugin name.
func (h *Handler) UpdateInfo(nCtx contextx.IContext, hostID int64, pluginName string, processInfo *types.ProcessInfo) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return fmt.Errorf("failed to check tenant id: %v", err)
	}

	if processInfo == nil {
		return fmt.Errorf("process info is nil")
	}

	filter := base.AliveFilter()
	for _, opt := range []OptFn{WithHostID(hostID), WithPluginName(pluginName)} {
		filter = opt(filter)
	}

	err := h.tenantDao(nCtx.TenantID()).UpdateField(nCtx, filter, FieldKeyInfo, convProcessInfoFromTypes(*processInfo))
	if err != nil {
		logger.G.Sys().With("host-id", hostID).With("plugin-name", pluginName).WithErr(err).
			Error("failed to update process v2 info")

		return fmt.Errorf("failed to update process v2 info: %v", err)
	}

	return nil
}

// UpdateManyInfo batch update the process info of v2 processes.
func (h *Handler) UpdateManyInfo(nCtx contextx.IContext, processInfoDeltas []*types.ProcessInfoDelta) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return fmt.Errorf("failed to check tenant id: %v", err)
	}

	if len(processInfoDeltas) == 0 {
		return nil
	}

	processUpdates := make([]*base.DocumentFieldUpdate, len(processInfoDeltas))
	for idx, processInfoDelta := range processInfoDeltas {
		filter := base.AliveFilter()
		for _, opt := range []OptFn{WithHostID(processInfoDelta.HostID), WithPluginName(processInfoDelta.PluginName)} {
			filter = opt(filter)
		}

		processUpdates[idx] = &base.DocumentFieldUpdate{
			Filter: filter,
			Fields: map[string]any{
				FieldKeyInfo: convProcessInfoFromTypes(processInfoDelta.ProcessInfo),
			},
		}
	}

	if err := h.tenantDao(nCtx.TenantID()).UpdateOneFieldBulk(nCtx, processUpdates); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to batch update process v2 info")

		return fmt.Errorf("failed to batch update process v2 info: %v", err)
	}

	return nil
}

func convertProcessToTypes(data *Process) *types.Process {
	return &types.Process{
		TenantID:      data.TenantID,
		HostID:        data.HostID,
		BizID:         data.BizID,
		PluginName:    data.Name,
		PluginPkgName: data.PluginPkgName,
		PluginGroup:   data.Group,
		Platform: platfmt.Platform{
			OS:   criteria.OSType(data.Platform.OS),
			Arch: criteria.CPUArch(data.Platform.Arch),
		},
		Generation: types.Generation(data.Generation),
		BindIP:     data.BindIP,
		BindPort:   data.BindPort,
		Info: types.ProcessInfo{
			Pid:        data.Info.Pid,
			Version:    data.Info.Version,
			AgentID:    data.Info.AgentID,
			AutoStart:  data.Info.Trusteeship,
			Status:     types.ProcessStatus(data.Info.Status),
			LastSyncAt: data.Info.LastSyncAt,
		},
		Identity: types.ProcessIdentity{
			Name:       data.Identity.Name,
			SetupPath:  data.Identity.SetupPath,
			PidPath:    data.Identity.PidPath,
			ConfigPath: data.Identity.ConfigPath,
			LogPath:    data.Identity.LogPath,
			User:       data.Identity.User,
		},
		Controller: types.ProcessController{
			StartCmd:   data.Controller.StartCmd,
			StopCmd:    data.Controller.StopCmd,
			RestartCmd: data.Controller.RestartCmd,
			ReloadCmd:  data.Controller.ReloadCmd,
			DebugCmd:   data.Controller.DebugCmd,
			KillCmd:    data.Controller.KillCmd,
			VersionCmd: data.Controller.VersionCmd,
			HealthCmd:  data.Controller.HealthCmd,
		},
		Resource: types.ProcessResource{
			CPULimitPercent: data.Resource.CPULimitPercent,
			MemLimitPercent: data.Resource.MemLimitPercent,
		},
		MonitorPolicy: types.ProcessMonitorPolicy{
			RestartType:    types.ProcessRestartType(data.MonitorPolicy.RestartType),
			StartCheckSecs: data.MonitorPolicy.StartCheckSecs,
			StopCheckSecs:  data.MonitorPolicy.StopCheckSecs,
			OpTimeoutSecs:  data.MonitorPolicy.OpTimeoutSecs,
		},
	}
}

func convProcessFromTypes(process *types.Process) *Process {
	return &Process{
		TenantID:      process.TenantID,
		HostID:        process.HostID,
		BizID:         process.BizID,
		Name:          process.PluginName,
		Group:         process.PluginGroup,
		PluginPkgName: process.PluginPkgName,
		Generation:    int64(process.Generation),
		Platform: platform{
			OS:   string(process.Platform.OS),
			Arch: string(process.Platform.Arch),
		},
		BindIP:   process.BindIP,
		BindPort: process.BindPort,
		Info:     convProcessInfoFromTypes(process.Info),
		Identity: processIdentity{
			Name:       process.Identity.Name,
			SetupPath:  process.Identity.SetupPath,
			PidPath:    process.Identity.PidPath,
			ConfigPath: process.Identity.ConfigPath,
			LogPath:    process.Identity.LogPath,
			User:       process.Identity.User,
		},
		Controller: processController{
			StartCmd:   process.Controller.StartCmd,
			StopCmd:    process.Controller.StopCmd,
			RestartCmd: process.Controller.RestartCmd,
			ReloadCmd:  process.Controller.ReloadCmd,
			DebugCmd:   process.Controller.DebugCmd,
			KillCmd:    process.Controller.KillCmd,
			VersionCmd: process.Controller.VersionCmd,
			HealthCmd:  process.Controller.HealthCmd,
		},
		Resource: processResource{
			CPULimitPercent: process.Resource.CPULimitPercent,
			MemLimitPercent: process.Resource.MemLimitPercent,
		},
		MonitorPolicy: processMonitorPolicy{
			RestartType:    string(process.MonitorPolicy.RestartType),
			StartCheckSecs: process.MonitorPolicy.StartCheckSecs,
			StopCheckSecs:  process.MonitorPolicy.StopCheckSecs,
			OpTimeoutSecs:  process.MonitorPolicy.OpTimeoutSecs,
		},
	}
}

func convProcessInfoFromTypes(info types.ProcessInfo) processInfo {
	return processInfo{
		Pid:         info.Pid,
		Version:     info.Version,
		AgentID:     info.AgentID,
		Trusteeship: info.AutoStart,
		Status:      string(info.Status),
		LastSyncAt:  info.LastSyncAt,
	}
}
