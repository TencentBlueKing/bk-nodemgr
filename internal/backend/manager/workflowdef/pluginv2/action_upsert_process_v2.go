/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pluginv2

import (
	"errors"
	"fmt"
	"time"

	pluginV2Utils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/pluginv2/utils"
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
	// ActionNameUpsertProcessV2 defines the action name.
	ActionNameUpsertProcessV2 = "upsert_process_v2"
)

// NewActionUpsertProcessV2 ...
func NewActionUpsertProcessV2(capability *Capability) action.Definition {
	return &actionUpsertProcessV2{
		daoHost:             capability.StorageTopo,
		daoPlugin:           capability.StoragePlugin,
		daoProcessV2:        capability.StoragePlugin,
		daoPluginDeployment: capability.StoragePlugin,
		provider:            capability.DiscoverProvider,
		gseHandler:          capability.GSEHandler,
	}
}

// ActParamUpsertProcessV2 ...
type ActParamUpsertProcessV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

// actionUpsertProcessV2 ...
type actionUpsertProcessV2 struct {
	daoHost             topoStg.IStorageHost
	daoPlugin           pluginStg.IDaoPlugin
	daoProcessV2        pluginStg.IDaoProcessV2
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	provider            discover.Discover
	gseHandler          gse.IHandler
}

// Name returns the name of the action.
func (act *actionUpsertProcessV2) Name() string {
	return ActionNameUpsertProcessV2
}

// Version returns the version of the action.
func (act *actionUpsertProcessV2) Version() string {
	return "1.0.0" // nolint: mnd,goconst
}

// Description returns the description of the action.
func (act *actionUpsertProcessV2) Description() string {
	return ""
}

// Timeout returns the timeout of the action.
func (act *actionUpsertProcessV2) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionUpsertProcessV2) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionUpsertProcessV2) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionUpsertProcessV2) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionUpsertProcessV2) Do(ctx *action.InstanceContext) error {
	param := new(ActParamUpsertProcessV2)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := pluginV2Utils.NewPluginActionStandarder(act.daoPluginDeployment)
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
		BizID:      std.DeployInfo().Process.BizID,
		PluginName: std.DeployInfo().Process.PluginName,
		Platform:   platfmt.UnknownPlatform(),
		Info: types.ProcessInfo{
			Version: std.DeployInfo().InstallOptions.Version,
			Status:  types.ProcessStatusInit,
		},
	}

	exist, err := act.daoProcessV2.ExistProcessV2(nCtx, process.HostID, process.PluginName)
	if err != nil {
		return fmt.Errorf("failed to check process v2 existence: %w", err)
	}

	if !exist {
		std.InstanceData().Log().
			Zh("指定进程不存在, host-id(%d), plugin-name(%s)",
				process.HostID, process.PluginName).
			En("the specified process does not exist, host-id(%d), plugin-name(%s)",
				process.HostID, process.PluginName).
			Info()

		err = act.daoProcessV2.CreateProcessV2(nCtx, &process)
		if err != nil {
			return fmt.Errorf("failed to create process v2: %w", err)
		}

		std.InstanceData().Log().
			Zh("成功创建 V2 进程, host-id(%d), plugin-name(%s)", process.HostID, process.PluginName).
			En("successfully create V2 process, host-id(%d), plugin-name(%s)", process.HostID, process.PluginName).
			Info()
	} else {
		std.InstanceData().Log().
			Zh("指定 V2 进程已存在, host-id(%d), plugin-name(%s)", process.HostID, process.PluginName).
			En("the specified V2 process exist, host-id(%d), plugin-name(%s)", process.HostID, process.PluginName).
			Info()

		err = act.daoProcessV2.UpdateProcessV2Info(nCtx, process.HostID, process.PluginName, &process.Info)
		if err != nil {
			return fmt.Errorf("failed to update process v2 info: %w", err)
		}

		std.InstanceData().Log().
			Zh("成功更新 V2 进程, host-id(%d), plugin-name(%s)", process.HostID, process.PluginName).
			En("successfully update V2 process, host-id(%d), plugin-name(%s)", process.HostID, process.PluginName).
			Info()
	}

	std.DeployInfo().Process = process

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionUpsertProcessV2) DisplayNameZh() string {
	return "创建或更新 V2 进程"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionUpsertProcessV2) DisplayNameEn() string {
	return "Upsert V2 Process"
}
