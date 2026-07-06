/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package node

import (
	"fmt"
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/relayhandler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameCheckSelfRelayPluginAlive defines the action name.
	ActionNameCheckSelfRelayPluginAlive = "check_self_relay_plugin_alive"
)

// NewActionCheckSelfRelayPluginAlive get a new action.
func NewActionCheckSelfRelayPluginAlive(capability *Capability) action.Definition {
	return &actionCheckSelfRelayPluginAlive{
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		daoProcess:            capability.StoragePlugin,
		gseHandlerProc:        capability.GSEHandler.NewHandlerProc(),
	}
}

// ActionParamCheckSelfRelayPluginAlive defines the action param.
type ActionParamCheckSelfRelayPluginAlive struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionCheckSelfRelayPluginAlive struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	daoProcess            pluginStg.IDaoProcess
	gseHandlerProc        gse.IHandlerProc
}

// Name returns the name of the action.
func (act *actionCheckSelfRelayPluginAlive) Name() string {
	return ActionNameCheckSelfRelayPluginAlive
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionCheckSelfRelayPluginAlive) DisplayNameZh() string {
	return "检查自身 Relay 插件存活"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionCheckSelfRelayPluginAlive) DisplayNameEn() string {
	return "Check Self Relay Plugin Alive"
}

// Version returns the version of the action.
func (act *actionCheckSelfRelayPluginAlive) Version() string {
	return "1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionCheckSelfRelayPluginAlive) Description() string {
	return "check self relay plugin alive"
}

// Timeout returns the timeout of the action.
func (act *actionCheckSelfRelayPluginAlive) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionCheckSelfRelayPluginAlive) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionCheckSelfRelayPluginAlive) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionCheckSelfRelayPluginAlive) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionCheckSelfRelayPluginAlive) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamCheckSelfRelayPluginAlive)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment, act.storageHost)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}

	deployInfo := std.DeployInfo()
	hostID := deployInfo.Host.HostID
	process, err := act.daoProcess.GetProcess(std.Context(), hostID, relayhandler.PluginName)
	if err != nil {
		return fmt.Errorf("failed to get self relay process from database, host-id(%d), plugin-name(%s): %w",
			hostID, relayhandler.PluginName, err)
	}
	if process == nil {
		return fmt.Errorf("self relay process not found in database, host-id(%d), plugin-name(%s)",
			hostID, relayhandler.PluginName)
	}
	if process.Identity.Name == "" {
		return fmt.Errorf("self relay process identity name is empty, host-id(%d), plugin-name(%s)",
			hostID, relayhandler.PluginName)
	}
	if deployInfo.Host.Dynamic.AgentID == "" {
		return fmt.Errorf("host dynamic agent id is empty, host-id(%d), plugin-name(%s)", hostID, relayhandler.PluginName)
	}

	processInfo, err := act.gseHandlerProc.QueryProcessInfo(
		std.Context(), relayhandler.PluginName, process.Identity.Name, deployInfo.Host.Dynamic.AgentID)
	if err != nil {
		return fmt.Errorf("failed to query self relay process info from gse, host-id(%d), plugin-name(%s), process-name(%s), agent-id(%s): %w",
			hostID, relayhandler.PluginName, process.Identity.Name, deployInfo.Host.Dynamic.AgentID, err)
	}
	if processInfo.Status != types.ProcessStatusRunning {
		return fmt.Errorf("self relay process status is not running, host-id(%d), plugin-name(%s), process-name(%s), agent-id(%s), status(%s)",
			hostID, relayhandler.PluginName, process.Identity.Name, deployInfo.Host.Dynamic.AgentID, processInfo.Status)
	}

	std.InstanceData().Log().
		Zh("检查自身 Relay 插件存活成功, plugin-name(%s), host-id(%d), status(%s), version(%s)",
			relayhandler.PluginName, hostID, processInfo.Status, processInfo.Version).
		En("check self relay plugin alive succeed, plugin-name(%s), host-id(%d), status(%s), version(%s)",
			relayhandler.PluginName, hostID, processInfo.Status, processInfo.Version).
		Info()

	return nil
}
