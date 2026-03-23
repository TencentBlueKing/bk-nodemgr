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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameFetchPluginProcess the name of action fetch plugin process.
	ActionNameFetchPluginProcess = "fetch_plugin_process"
)

// NewActionFetchPluginProcess new an action to fetch plugin process.
func NewActionFetchPluginProcess(capability *Capability) action.Definition {
	return &actionFetchPluginProcess{
		daoPluginDeployment: capability.StoragePlugin,
		daoProcess:          capability.StoragePlugin,
		daoHost:             capability.StorageTopo,
	}
}

// ActParamFetchPluginProcess defines the parameters for actionFetchPluginProcess.
type ActParamFetchPluginProcess struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

type actionFetchPluginProcess struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoProcess          pluginStg.IDaoProcess
	daoHost             topoStg.IStorageHost
}

// Name returns the name of the action.
func (act *actionFetchPluginProcess) Name() string {
	return ActionNameFetchPluginProcess
}

// Version returns the version of the action.
func (act *actionFetchPluginProcess) Version() string {
	return "1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionFetchPluginProcess) Description() string {
	return "fetch plugin process"
}

// Timeout returns the timeout of the action.
func (act *actionFetchPluginProcess) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionFetchPluginProcess) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionFetchPluginProcess) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionFetchPluginProcess) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionFetchPluginProcess) Do(ctx *action.InstanceContext) error {
	param := new(ActParamFetchPluginProcess)
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
	deployInfo := std.DeployInfo()
	process, err := act.daoProcess.GetProcess(nCtx, deployInfo.Process.HostID, deployInfo.Process.PluginName)
	if err != nil {
		std.InstanceData().Log().
			Zh("获取进程失败，process-name(%s), host-id(%d): %v",
				deployInfo.Process.PluginName, deployInfo.Process.HostID, err).
			En("failed to get process, process-name(%s), host-id(%d): %v",
				deployInfo.Process.PluginName, deployInfo.Process.HostID, err).
			Error()

		return fmt.Errorf("failed to get plugin process, process-name(%s), host-id(%d): %w",
			deployInfo.Process.PluginName, deployInfo.Process.HostID, err)
	}

	std.InstanceData().Log().
		Zh("获取插件进程成功，plugin-name(%s), host-id(%d)",
			deployInfo.Process.PluginName, deployInfo.Process.HostID).
		En("fetch plugin process succeed, plugin-name(%s), host-id(%d)",
			deployInfo.Process.PluginName, deployInfo.Process.HostID).
		Info()

	deployInfo.Process = *process

	host, err := act.daoHost.GetHostByID(nCtx, deployInfo.Process.HostID)
	if err != nil {
		return fmt.Errorf("failed to get host by id, host-id(%d): %w", deployInfo.Process.HostID, err)
	}
	if host.Dynamic.AgentID == "" {
		return fmt.Errorf("host dynamic agent id is empty, host-id(%d)", deployInfo.Process.HostID)
	}

	// notice: overwrite the potentially stale process AgentID with the latest host value.
	deployInfo.Process.Info.AgentID = host.Dynamic.AgentID

	std.InstanceData().Log().
		Zh("获取插件AgentID成功，agent-id(%s)",
			deployInfo.Process.Info.AgentID).
		En("fetch plugin agent id succeed, agent-id(%s)",
			deployInfo.Process.Info.AgentID).
		Info()

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionFetchPluginProcess) DisplayNameZh() string {
	return "获取插件进程"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionFetchPluginProcess) DisplayNameEn() string {
	return "Fetch Plugin Process"
}
