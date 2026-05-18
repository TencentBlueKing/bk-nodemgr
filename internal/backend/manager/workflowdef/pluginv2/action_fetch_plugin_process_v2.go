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
	"time"

	pluginV2Utils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/pluginv2/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameFetchPluginProcessV2 the name of action fetch plugin process.
	ActionNameFetchPluginProcessV2 = "fetch_plugin_process_v2"
)

// NewActionFetchPluginProcessV2 new an action to fetch plugin process.
func NewActionFetchPluginProcessV2(capability *Capability) action.Definition {
	return &actionFetchPluginProcessV2{
		daoPluginDeployment: capability.StoragePlugin,
		daoProcessV2:        capability.StoragePlugin,
		daoHost:             capability.StorageTopo,
	}
}

// ActParamFetchPluginProcessV2 defines the parameters for actionFetchPluginProcessV2.
type ActParamFetchPluginProcessV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

type actionFetchPluginProcessV2 struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoProcessV2        pluginStg.IDaoProcessV2
	daoHost             topoStg.IStorageHost
}

// Name returns the name of the action.
func (act *actionFetchPluginProcessV2) Name() string {
	return ActionNameFetchPluginProcessV2
}

// Version returns the version of the action.
func (act *actionFetchPluginProcessV2) Version() string {
	return "1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionFetchPluginProcessV2) Description() string {
	return "fetch plugin process v2"
}

// Timeout returns the timeout of the action.
func (act *actionFetchPluginProcessV2) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionFetchPluginProcessV2) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionFetchPluginProcessV2) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionFetchPluginProcessV2) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionFetchPluginProcessV2) Do(ctx *action.InstanceContext) error {
	param := new(ActParamFetchPluginProcessV2)
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
	deployInfo := std.DeployInfo()
	process, err := pluginV2Utils.GetActualExistingProcess(
		nCtx,
		act.daoProcessV2,
		act.daoHost,
		deployInfo.Process.HostID,
		deployInfo.Process.PluginName,
	)
	if err != nil {
		std.InstanceData().Log().
			Zh("获取 V2 进程失败, process-name(%s), host-id(%d): %v",
				deployInfo.Process.PluginName, deployInfo.Process.HostID, err).
			En("failed to get V2 process, process-name(%s), host-id(%d): %v",
				deployInfo.Process.PluginName, deployInfo.Process.HostID, err).
			Error()

		return err
	}

	std.InstanceData().Log().
		Zh("获取 V2 插件进程成功, plugin-name(%s), host-id(%d)",
			deployInfo.Process.PluginName, deployInfo.Process.HostID).
		En("fetch V2 plugin process succeed, plugin-name(%s), host-id(%d)",
			deployInfo.Process.PluginName, deployInfo.Process.HostID).
		Info()

	deployInfo.Process = *process

	std.InstanceData().Log().
		Zh("获取 V2 插件AgentID成功, agent-id(%s)",
			deployInfo.Process.Info.AgentID).
		En("fetch V2 plugin agent id succeed, agent-id(%s)",
			deployInfo.Process.Info.AgentID).
		Info()

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionFetchPluginProcessV2) DisplayNameZh() string {
	return "获取 V2 插件进程"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionFetchPluginProcessV2) DisplayNameEn() string {
	return "Fetch V2 Plugin Process"
}
