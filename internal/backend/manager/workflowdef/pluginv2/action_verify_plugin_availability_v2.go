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

package pluginv2

import (
	"errors"
	"fmt"
	"time"

	pluginV2Utils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/pluginv2/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	releaseStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameVerifyPluginAvailabilityV2 the name of action verify plugin availability.
	ActionNameVerifyPluginAvailabilityV2 = "verify_plugin_availability_v2"
)

// NewActionVerifyPluginAvailabilityV2 new an action to verify plugin availability.
func NewActionVerifyPluginAvailabilityV2(capability *Capability) action.Definition {
	return &actionVerifyPluginAvailabilityV2{
		daoPluginDeployment: capability.StoragePlugin,
		daoReleasePlugin:    capability.StorageRelease,
		daoHost:             capability.StorageTopo,
		daoPlugin:           capability.StoragePlugin,
	}
}

// ActParamVerifyPluginAvailabilityV2 defines the parameters for actionVerifyPluginAvailabilityV2.
type ActParamVerifyPluginAvailabilityV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

type actionVerifyPluginAvailabilityV2 struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoReleasePlugin    releaseStg.IPlugin
	daoHost             topoStg.IStorageHost
	daoPlugin           pluginStg.IDaoPlugin
}

// Name returns the name of the action.
func (act *actionVerifyPluginAvailabilityV2) Name() string {
	return ActionNameVerifyPluginAvailabilityV2
}

// Version returns the version of the action.
func (act *actionVerifyPluginAvailabilityV2) Version() string {
	return "1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionVerifyPluginAvailabilityV2) Description() string {
	return "verify plugin availability v2"
}

// Timeout returns the timeout of the action.
func (act *actionVerifyPluginAvailabilityV2) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionVerifyPluginAvailabilityV2) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionVerifyPluginAvailabilityV2) MaxRetryCount() uint {
	return 1 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionVerifyPluginAvailabilityV2) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionVerifyPluginAvailabilityV2) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamVerifyPluginAvailabilityV2)
	err = conv.MapToStruct(ctx.Data.Content, param)
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

	host, err := act.daoHost.GetHostByID(nCtx, deployInfo.Process.HostID)
	if err != nil {
		return err
	}

	plugin, err := act.daoPlugin.GetPlugin(nCtx, deployInfo.Process.PluginName)
	if err != nil {
		return err
	}

	version := deployInfo.InstallOptions.Version
	if version == "" {
		version = deployInfo.Process.Info.Version
	}

	pluginPkg, err := act.daoReleasePlugin.GetReleasePlugin(nCtx, types.ReleasePluginKey{
		Generation: host.Dynamic.NodeGeneration,
		Platform:   platform.NewPlatform(host.Dynamic.NodeOsType, host.Dynamic.NodeCPUArch),
		Version:    version,
		Name:       plugin.PkgName,
	})
	if err != nil {
		return err
	}

	if !pluginPkg.LaunchNodeType.IsLaunchNode(host.Dynamic.NodeRole) {
		std.InstanceData().Log().
			Zh("插件包启动节点类型与主机节点角色不匹配,"+
				"plugin-name(%s), plugin-pkg-name(%s), version(%s), host-id(%d), launch-node-type(%s), host-node-role(%s)",
				plugin.Name, pluginPkg.Name, version, deployInfo.Process.HostID,
				pluginPkg.LaunchNodeType, host.Dynamic.NodeRole).
			En("plugin pkg launch node type not match host node role, "+
				"plugin-name(%s), plugin-pkg-name(%s), version(%s), host-id(%d), launch-node-type(%s), host-node-role(%s)",
				plugin.Name, pluginPkg.Name, version, deployInfo.Process.HostID,
				pluginPkg.LaunchNodeType, host.Dynamic.NodeRole).
			Error()

		return fmt.Errorf("plugin pkg launch node type not match host node role, "+
			"plugin-name(%s), plugin-pkg-name(%s), version(%s), host-id(%d), launch-node-type(%s), host-node-role(%s)",
			plugin.Name, pluginPkg.Name, version, deployInfo.Process.HostID,
			pluginPkg.LaunchNodeType, host.Dynamic.NodeRole)
	}

	if !pluginPkg.Enabled {
		std.InstanceData().Log().
			Zh("插件包未启用, plugin-name(%s), plugin-pkg-name(%s), version(%s), host-id(%d)",
				plugin.Name, pluginPkg.Name, version, deployInfo.Process.HostID).
			En("plugin pkg is not enabled, plugin-name(%s), plugin-pkg-name(%s), version(%s), host-id(%d)",
				plugin.Name, pluginPkg.Name, version, deployInfo.Process.HostID).
			Error()

		return fmt.Errorf("plugin pkg is not enabled, plugin-name(%s), plugin-pkg-name(%s), version(%s), host-id(%d)",
			plugin.Name, pluginPkg.Name, version, deployInfo.Process.HostID)
	}

	std.InstanceData().Log().
		Zh("插件包验证成功,"+
			"plugin-name(%s), plugin-pkg-name(%s), version(%s), host-id(%d), launch-node-type(%s)",
			plugin.Name, pluginPkg.Name, version, deployInfo.Process.HostID,
			pluginPkg.LaunchNodeType).
		En("plugin pkg verify succeed, "+
			"plugin-name(%s), plugin-pkg-name(%s), version(%s), host-id(%d), launch-node-type(%s)",
			plugin.Name, pluginPkg.Name, version, deployInfo.Process.HostID,
			pluginPkg.LaunchNodeType).
		Info()

	std.DeployInfo().Process.PluginName = plugin.Name
	std.DeployInfo().Process.PluginGroup = plugin.Group
	std.DeployInfo().Process.PluginPkgName = plugin.PkgName
	std.DeployInfo().Process.Generation = pluginPkg.Generation
	std.DeployInfo().Process.Platform = platform.NewPlatform(host.Dynamic.NodeOsType, host.Dynamic.NodeCPUArch)

	std.InstanceData().Log().
		Zh("更新插件信息成功, plugin-name(%s), plugin-group(%s), plugin-pkg-name(%s), version(%s), generation(%d), platform(%s)",
			std.DeployInfo().Process.PluginName, std.DeployInfo().Process.PluginGroup, std.DeployInfo().Process.PluginPkgName,
			version, std.DeployInfo().Process.Generation, std.DeployInfo().Process.Platform).
		En("update plugin info succeed, plugin-name(%s), plugin-group(%s), plugin-pkg-name(%s), version(%s), generation(%d), platform(%s)",
			std.DeployInfo().Process.PluginName, std.DeployInfo().Process.PluginGroup, std.DeployInfo().Process.PluginPkgName,
			version, std.DeployInfo().Process.Generation, std.DeployInfo().Process.Platform).
		Info()

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionVerifyPluginAvailabilityV2) DisplayNameZh() string {
	return "验证 V2 插件可用性"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionVerifyPluginAvailabilityV2) DisplayNameEn() string {
	return "Verify V2 Plugin Availability"
}
