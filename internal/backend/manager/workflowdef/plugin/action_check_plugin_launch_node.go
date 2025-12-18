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
	releaseStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameCheckPluginLaunchNode the name of action check plugin launch node.
	ActionNameCheckPluginLaunchNode = "check_plugin_launch_node"
)

// NewActionCheckPluginLaunchNode new an action to check plugin launch node.
func NewActionCheckPluginLaunchNode(capability *Capability) action.Definition {
	return &actionCheckPluginLaunchNode{
		daoPluginDeployment: capability.StoragePlugin,
		daoReleasePlugin:    capability.StorageRelease,
		daoHost:             capability.StorageTopo,
	}
}

// ActParamCheckPluginLaunchNode defines the parameters for actionCheckPluginLaunchNode.
type ActParamCheckPluginLaunchNode struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

type actionCheckPluginLaunchNode struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoReleasePlugin    releaseStg.IPlugin
	daoHost             topoStg.IStorageHost
}

// Name returns the name of the action.
func (act *actionCheckPluginLaunchNode) Name() string {
	return ActionNameCheckPluginLaunchNode
}

// Version returns the version of the action.
func (act *actionCheckPluginLaunchNode) Version() string {
	return "1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionCheckPluginLaunchNode) Description() string {
	return "check plugin launch node"
}

// Timeout returns the timeout of the action.
func (act *actionCheckPluginLaunchNode) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionCheckPluginLaunchNode) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionCheckPluginLaunchNode) MaxRetryCount() uint {
	return 1 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionCheckPluginLaunchNode) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionCheckPluginLaunchNode) Do(ctx *action.InstanceContext) error {
	param := new(ActParamCheckPluginLaunchNode)
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

	host, err := act.daoHost.GetHostByID(nCtx, deployInfo.Process.HostID)
	if err != nil {
		return err
	}

	plugin, err := act.daoReleasePlugin.GetReleasePlugin(
		nCtx,
		deployInfo.Process.PluginName,
		host.Dynamic.NodeGeneration,
		platform.NewPlatform(host.Dynamic.NodeOsType, host.Dynamic.NodeCPUArch),
		deployInfo.InstallOptions.Version)
	if err != nil {
		return err
	}

	if !plugin.ReleaseAdditionInfoPlugin.LaunchNodeType.IsLaunchNode(host.Dynamic.NodeRole) {
		std.InstanceData().LogE(
			fmt.Sprintf("plugin launch node type not match host node role, plugin-name(%s), host-id(%d), launch-node-type(%s), host-node-role(%s)",
				deployInfo.Process.PluginName, deployInfo.Process.HostID,
				plugin.ReleaseAdditionInfoPlugin.LaunchNodeType, host.Dynamic.NodeRole))

		return errors.New("plugin launch node type not match host node role")
	}

	std.InstanceData().LogI(
		fmt.Sprintf("plugin launch node type match host node role, plugin-name(%s), host-id(%d), launch-node-type(%s), host-node-role(%s)",
			deployInfo.Process.PluginName, deployInfo.Process.HostID,
			plugin.ReleaseAdditionInfoPlugin.LaunchNodeType, host.Dynamic.NodeRole))

	return nil
}
