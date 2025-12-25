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
	"errors"
	"fmt"
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameUpsertProcess defines the action name.
	ActionNameUpsertProcess = "upsert_process"
)

// NewActionUpsertProcess ...
func NewActionUpsertProcess(capability *Capability) action.Definition {
	return &actionUpsertProcess{
		daoHost:             capability.StorageTopo,
		daoPlugin:           capability.StoragePlugin,
		daoProcess:          capability.StoragePlugin,
		daoPluginDeployment: capability.StoragePlugin,
		provider:            capability.DiscoverProvider,
		gseHandler:          capability.GSEHandler,
	}
}

// ActParamUpsertProcess ...
type ActParamUpsertProcess struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// actionUpsertProcess ...
type actionUpsertProcess struct {
	daoHost             topoStg.IStorageHost
	daoPlugin           pluginStg.IDaoPlugin
	daoProcess          pluginStg.IDaoProcess
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	provider            discover.Discover
	gseHandler          gse.IHandler
}

// Name returns the name of the action.
func (act *actionUpsertProcess) Name() string {
	return ActionNameUpsertProcess
}

// Version returns the version of the action.
func (act *actionUpsertProcess) Version() string {
	return "1.0.0" // nolint: mnd,goconst
}

// Description returns the description of the action.
func (act *actionUpsertProcess) Description() string {
	return ""
}

// Timeout returns the timeout of the action.
func (act *actionUpsertProcess) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionUpsertProcess) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionUpsertProcess) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionUpsertProcess) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionUpsertProcess) Do(ctx *action.InstanceContext) error {
	param := new(ActParamUpsertProcess)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := pluginUtils.NewPluginActionStandarder(act.daoPluginDeployment)
	if err = std.Initialize(ctx, param.PluginActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	nCtx := std.Context()
	process := types.Process{
		TenantID:   nCtx.TenantID(),
		HostID:     std.DeployInfo().Process.HostID,
		PluginName: std.DeployInfo().Process.PluginName,
		Platform:   platfmt.UnknownPlatform(),
		Info: types.ProcessInfo{
			Version: std.DeployInfo().InstallOptions.Version,
			Status:  types.ProcessStatusInit,
		},
	}

	exist, err := act.daoProcess.ExistProcess(nCtx, process.HostID, process.PluginName)
	if err != nil {
		return fmt.Errorf("failed to check process exist: %w", err)
	}

	if !exist {
		std.InstanceData().LogI(fmt.Sprintf("the specified process does not exist, host-id(%d), plugin-name(%s)",
			process.HostID, process.PluginName))

		err = act.daoProcess.CreateProcess(nCtx, &process)
		if err != nil {
			return fmt.Errorf("failed to create process: %w", err)
		}

		std.InstanceData().LogI(fmt.Sprintf("successfully create process, host-id(%d), plugin-name(%s)", process.HostID, process.PluginName))
	} else {
		std.InstanceData().LogI(fmt.Sprintf("the specified process exist, host-id(%d), plugin-name(%s)", process.HostID, process.PluginName))

		err = act.daoProcess.UpdateProcessInfo(nCtx, process.HostID, process.PluginName, &process.Info)
		if err != nil {
			return fmt.Errorf("failed to update process info: %w", err)
		}

		std.InstanceData().LogI(fmt.Sprintf("successfully update process, host-id(%d), plugin-name(%s)", process.HostID, process.PluginName))
	}

	std.DeployInfo().Process = process

	return nil
}
