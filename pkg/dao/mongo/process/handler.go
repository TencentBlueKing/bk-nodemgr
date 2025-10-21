/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package process

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

// IHandler process handler interface.
type IHandler interface {
	// Create create a new process.
	Create(nCtx contextx.IContext, process *types.Process) error

	// Count count process by conditions.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// List lists process by page and conditions.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.Process, int64, error)

	// Get gets a process by conditions.
	Get(nCtx contextx.IContext, opts ...OptFn) (*types.Process, error)

	// Update update a process by conditions.
	Update(nCtx contextx.IContext, hostID int64, pluginName string, process *types.Process) error

	// UpdateInfo update a process info by conditions.
	UpdateInfo(nCtx contextx.IContext, hostID int64, pluginName string, processInfo *types.ProcessInfo) error

	// UpdateManyInfo batch update process info by process ID.
	UpdateManyInfo(nCtx contextx.IContext, processInfosDeltas []*types.ProcessInfoDelta) error

	// Delete delete a process by conditions.
	Delete(nCtx contextx.IContext, hostID int64, pluginName string) error

	// Exist check a process exist by conditions.
	Exist(nCtx contextx.IContext, hostID int64, pluginName string) (bool, error)
}

var _ IHandler = &Handler{}

// Handler this is a Handler to operate process table.
type Handler struct {
	client *mongo.Database
	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

func (h *Handler) tenantDao(tenantID string) *dao {
	tableName := TableName(tenantID)

	if d, ok := h.daoMap.Load(tableName); ok {
		return d.(*dao) // nolint: forcetypeassert
	}

	newDaoClient := newDao(h.client, tableName)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).With("table-name", tableName).Warn("failed to ensure process indexes")
	}

	d, _ := h.daoMap.LoadOrStore(tableName, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao) // nolint: forcetypeassert
}

// New create a new process handler.
func New(client *mongo.Database) *Handler {
	return &Handler{
		client: client,
		daoMap: sync.Map{},
	}
}

// Create create a new process.
func (h *Handler) Create(nCtx contextx.IContext, process *types.Process) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	data := convProcessFromTypes(process)

	if err := h.tenantDao(nCtx.TenantID()).Create(nCtx, data); err != nil {
		return fmt.Errorf("failed to create process, err: %w", err)
	}

	return nil
}

func convProcessFromTypes(process *types.Process) *Process {
	data := &Process{
		TenantID:      process.TenantID,
		HostID:        process.HostID,
		Name:          process.PluginName,
		Group:         process.PluginGroup,
		PluginPkgName: process.PluginPkgName,
		Generation:    int64(process.Generation),
		Platform:      convPlatformFromTypes(process.Platform),
		Info:          convProcessInfoFromTypes(process.Info),
		Identity:      convProcessIdentityFromTypes(process.Identity),
		Controller: processController{
			StartCmd:   process.Controller.StartCmd,
			StopCmd:    process.Controller.StopCmd,
			RestartCmd: process.Controller.RestartCmd,
			ReloadCmd:  process.Controller.ReloadCmd,
			KillCmd:    process.Controller.KillCmd,
			VersionCmd: process.Controller.VersionCmd,
			HealthCmd:  process.Controller.HealthCmd,
		},
		Resource: processResource{
			CPULimitPercent: process.Resource.CPULimitPercent,
			MemLimitPercent: process.Resource.MemLimitPercent,
		},
		MonitorPolicy: processMonitorPolicy{
			AutoType:       string(process.MonitorPolicy.AutoType),
			StartCheckSecs: process.MonitorPolicy.StartCheckSecs,
			StopCheckSecs:  process.MonitorPolicy.StopCheckSecs,
			OpTimeoutSecs:  process.MonitorPolicy.OpTimeoutSecs,
		},
	}

	return data
}

func convProcessInfoFromTypes(info types.ProcessInfo) processInfo {
	return processInfo{
		Pid:         info.Pid,
		Version:     info.Version,
		AgentID:     info.AgentID,
		Trusteeship: info.Trusteeship,
		Status:      string(info.Status),
	}
}

func convProcessIdentityFromTypes(identity types.ProcessIdentity) processIdentity {
	return processIdentity{
		Name:       identity.Name,
		SetupPath:  identity.SetupPath,
		PidPath:    identity.PidPath,
		ConfigPath: identity.ConfigPath,
		LogPath:    identity.LogPath,
		User:       identity.User,
	}
}

func convPlatformFromTypes(p platfmt.Platform) platform {
	return platform{
		OS:   string(p.OS),
		Arch: string(p.Arch),
	}
}

// Count count host by conditions.
func (h *Handler) Count(nCtx contextx.IContext, opts ...OptFn) (int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.tenantDao(nCtx.TenantID()).Count(nCtx, filter)
}

// List list process by page and conditions.
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

	findOpt := base.ParsePage(page)

	data, err := h.tenantDao(nCtx.TenantID()).List(nCtx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	process := make([]*types.Process, len(data))
	for idx, host := range data {
		process[idx] = convertProcessToTypes(host)
	}

	return process, num, nil
}

// Get get a process.
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

func convertProcessToTypes(data *Process) *types.Process {
	process := &types.Process{
		TenantID:      data.TenantID,
		HostID:        data.HostID,
		PluginName:    data.Name,
		PluginPkgName: data.PluginPkgName,
		PluginGroup:   data.Group,
		Platform: platfmt.Platform{
			OS:   criteria.OSType(data.Platform.OS),
			Arch: criteria.CPUArch(data.Platform.Arch),
		},
		Generation: types.Generation(data.Generation),
		Info: types.ProcessInfo{
			Pid:         data.Info.Pid,
			Version:     data.Info.Version,
			AgentID:     data.Info.AgentID,
			Trusteeship: data.Info.Trusteeship,
			Status:      types.ProcessStatus(data.Info.Status),
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
			KillCmd:    data.Controller.KillCmd,
			VersionCmd: data.Controller.VersionCmd,
			HealthCmd:  data.Controller.HealthCmd,
		},
		Resource: types.ProcessResource{
			CPULimitPercent: data.Resource.CPULimitPercent,
			MemLimitPercent: data.Resource.MemLimitPercent,
		},
		MonitorPolicy: types.ProcessMonitorPolicy{
			AutoType:       types.ProcessAutoType(data.MonitorPolicy.AutoType),
			StartCheckSecs: data.MonitorPolicy.StartCheckSecs,
			StopCheckSecs:  data.MonitorPolicy.StopCheckSecs,
			OpTimeoutSecs:  data.MonitorPolicy.OpTimeoutSecs,
		},
	}

	return process
}

// UpdateInfo update process info.
func (h *Handler) UpdateInfo(nCtx contextx.IContext, hostID int64, pluginName string, processInfo *types.ProcessInfo) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return fmt.Errorf("failed to check tenant id: %v", err)
	}

	filter := base.AliveFilter()
	opts := []base.OptFn{
		WithHostID(hostID),
		WithPluginName(pluginName),
	}
	for _, opt := range opts {
		filter = opt(filter)
	}

	data := convProcessInfoFromTypes(*processInfo)
	err := h.tenantDao(nCtx.TenantID()).UpdateField(nCtx, filter, FieldKeyInfo, data)
	if err != nil {
		logger.G.Sys().With("host-id", hostID).With("plugin-name", pluginName).WithErr(err).Error("failed to update process info")

		return fmt.Errorf("failed to update process info: %v", err)
	}

	return nil
}

// UpdateManyInfo batch update process info by process ID.
func (h *Handler) UpdateManyInfo(nCtx contextx.IContext, processInfosDeltas []*types.ProcessInfoDelta) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return fmt.Errorf("failed to check tenant id: %v", err)
	}

	processUpdates := make([]*base.DocumentFieldUpdate, len(processInfosDeltas))
	for idx, processInfosDelta := range processInfosDeltas {
		filter := base.AliveFilter()
		opts := []base.OptFn{
			WithHostID(processInfosDelta.HostID),
			WithPluginName(processInfosDelta.PluginName),
		}
		for _, opt := range opts {
			filter = opt(filter)
		}

		processUpdates[idx] = &base.DocumentFieldUpdate{
			Filter: filter,
			Fields: map[string]any{
				FieldKeyInfo: convProcessInfoFromTypes(processInfosDelta.ProcessInfo),
			},
		}
	}

	err := h.tenantDao(nCtx.TenantID()).UpdateFieldsBulk(nCtx, processUpdates)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to batch update process info")

		return fmt.Errorf("failed to batch update process info: %v", err)
	}

	return nil
}

// Delete delete process.
func (h *Handler) Delete(nCtx contextx.IContext, hostID int64, pluginName string) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return fmt.Errorf("failed to check tenant id: %v", err)
	}

	filter := base.AliveFilter()
	opts := []base.OptFn{
		WithHostID(hostID),
		WithPluginName(pluginName),
	}

	for _, opt := range opts {
		filter = opt(filter)
	}

	err := h.tenantDao(nCtx.TenantID()).DeleteMany(nCtx, filter)

	if err != nil {
		logger.G.Sys().With("host-id", hostID).With("plugin-name", pluginName).WithErr(err).Error("failed to delete process")

		return fmt.Errorf("failed to delete process: %v", err)
	}

	return nil
}

// Exist check process exist.
func (h *Handler) Exist(nCtx contextx.IContext, hostID int64, pluginName string) (bool, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return false, fmt.Errorf("failed to check tenant id: %v", err)
	}

	filter := base.AliveFilter()
	opts := []base.OptFn{
		WithHostID(hostID),
		WithPluginName(pluginName),
	}
	for _, opt := range opts {
		filter = opt(filter)
	}

	exist, err := h.tenantDao(nCtx.TenantID()).Exist(nCtx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check process exist: %v", err)
	}

	return exist, nil
}

// Update update process.
func (h *Handler) Update(nCtx contextx.IContext, hostID int64, pluginName string, process *types.Process) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return fmt.Errorf("failed to check tenant id: %v", err)
	}

	if process == nil {
		return fmt.Errorf("process is nil")
	}

	filter := base.AliveFilter()
	opts := []base.OptFn{
		WithHostID(hostID),
		WithPluginName(pluginName),
	}

	for _, opt := range opts {
		filter = opt(filter)
	}

	data := convProcessFromTypes(process)

	updates := []*base.DocumentFieldUpdate{
		{
			Filter: filter,
			Fields: map[string]any{
				FieldKeyHostID:        data.HostID,
				FieldKeyPluginName:    data.Name,
				FieldKeyGroup:         data.Group,
				FieldKeyPkgName:       data.PluginPkgName,
				FieldKeyGeneration:    data.Generation,
				FieldKeyPlatform:      data.Platform,
				FieldKeyInfo:          data.Info,
				FieldKeyIdentity:      data.Identity,
				FieldKeyController:    data.Controller,
				FieldKeyResource:      data.Resource,
				FieldKeyMonitorPolicy: data.MonitorPolicy,
			},
		},
	}

	if err := h.tenantDao(nCtx.TenantID()).UpdateFieldsBulk(nCtx, updates); err != nil {
		return fmt.Errorf("failed to update process: %v", err)
	}

	return nil
}
