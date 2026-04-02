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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameInjectPluginCustomDeployConfig defines the action name.
	ActionNameInjectPluginCustomDeployConfig = "inject_plugin_custom_deploy_config"
)

// NewActionInjectPluginCustomDeployConfig get a new action.
func NewActionInjectPluginCustomDeployConfig(capability *Capability) action.Definition {
	return &actionInjectPluginCustomDeployConfig{
		daoPluginDeployment: capability.StoragePlugin,
		daoHost:             capability.StorageTopo,
		daoDomainGse:        capability.StorageTopo,
	}
}

// ActParamInjectPluginCustomDeployConfig this is the param for render deployment.
type ActParamInjectPluginCustomDeployConfig struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

type actionInjectPluginCustomDeployConfig struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoHost             topoStg.IStorageHost
	daoDomainGse        topoStg.IStorageDomainGse
}

// Name returns the name of the action.
func (act *actionInjectPluginCustomDeployConfig) Name() string {
	return ActionNameInjectPluginCustomDeployConfig
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionInjectPluginCustomDeployConfig) DisplayNameZh() string {
	return "注入插件自定义部署配置"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionInjectPluginCustomDeployConfig) DisplayNameEn() string {
	return "Inject Plugin Custom Deploy Config"
}

// Version returns the version of the action.
func (act *actionInjectPluginCustomDeployConfig) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionInjectPluginCustomDeployConfig) Description() string {
	return "inject plugin custom deploy config"
}

// Timeout returns the timeout of the action.
func (act *actionInjectPluginCustomDeployConfig) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionInjectPluginCustomDeployConfig) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionInjectPluginCustomDeployConfig) MaxRetryCount() uint {
	return 1 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionInjectPluginCustomDeployConfig) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: lll
func (act *actionInjectPluginCustomDeployConfig) Do(ctx *action.InstanceContext) error {
	param := new(ActParamInjectPluginCustomDeployConfig)
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

	host, err := act.daoHost.GetHostByID(std.Context(), std.DeployInfo().Process.HostID)
	if err != nil {
		return fmt.Errorf("failed to get host by ID: %w", err)
	}

	nodeConstant, err := deployconstant.GetNodeDeployConf(host.Dynamic.NodeGeneration, host.Dynamic.NodeOsType)
	if err != nil {
		return fmt.Errorf("failed to get node deploy conf: %w", err)
	}

	pluginConstant, err := deployconstant.GetPluginDeployConf(host.Dynamic.NodeGeneration, host.Dynamic.NodeOsType)
	if err != nil {
		return fmt.Errorf("failed to get plugin deploy conf: %w", err)
	}

	customDeployConfig, err := act.daoDomainGse.GetNetworkUnitCustomDeployConfig(std.Context(), host.Dynamic.NetworkUnitID, host.Dynamic.NodeOsType)
	if err != nil {
		return fmt.Errorf("failed to get network unit custom deploy config: %w", err)
	}

	nodeRole := host.Dynamic.NodeRole
	// update deploy constant with custom deploy config if custom deploy config exists
	if customDeployConfig != nil {
		// override BaseDeployDir and generate corresponding configuration
		if customDeployConfig.PluginRuntime.BaseDeployDir != "" {
			pluginConstant.BaseDeployDir = customDeployConfig.PluginRuntime.BaseDeployDir
		}

		if customDeployConfig.GSERuntime.BaseDeployDir != "" {
			nodeConstant.BaseDeployDir = customDeployConfig.GSERuntime.BaseDeployDir
		}

		// override BaseWorkDir and generate corresponding configuration
		if customDeployConfig.InstallerRuntime.BaseWorkDir != "" {
			pluginConstant.BaseWorkDir = customDeployConfig.InstallerRuntime.BaseWorkDir
		}
	}

	// installer workdir priority: user specified in info > network unit custom deploy config > deploy constant default.
	if std.DeployInfo().InstallerRuntime.BaseWorkDir != "" {
		pluginConstant.BaseWorkDir = std.DeployInfo().InstallerRuntime.BaseWorkDir
	}

	// installer workdir priority: user specified in info > network unit custom deploy config > deploy constant default.
	std.DeployInfo().InstallerRuntime.BaseWorkDir = pluginConstant.BaseWorkDir
	std.DeployInfo().InstallerRuntime.WorkDir = pluginConstant.GenerateWorkDir()

	std.DeployInfo().BaseRuntime = types.PluginDeploymentBaseRuntime{
		BaseDeployDir:         pluginConstant.BaseDeployDir,
		DeployDir:             pluginConstant.GeneratePluginDeployDir(),
		GSEHomeDir:            nodeConstant.GenerateNodeHomeDir(nodeRole),
		PluginHomeDir:         pluginConstant.GeneratePluginHomeDir(std.DeployInfo().Process.PluginGroup, std.DeployInfo().Process.PluginName),
		DataIPC:               conv.NonEmptyOr(customDeployConfig.GSERuntime.DataIPC, nodeConstant.GenerateDataIPCPath(nodeRole)),
		PluginIPC:             conv.NonEmptyOr(customDeployConfig.GSERuntime.PluginIPC, nodeConstant.GeneratePluginIPCPath(nodeRole)),
		HostIDPath:            pluginConstant.HostIDPath,
		LogDir:                conv.NonEmptyOr(customDeployConfig.PluginRuntime.LogDir, pluginConstant.LogDir),
		DataDir:               conv.NonEmptyOr(customDeployConfig.PluginRuntime.DataDir, pluginConstant.GeneratePluginDataDir(std.DeployInfo().Process.PluginGroup, std.DeployInfo().Process.PluginName)),
		RunDir:                conv.NonEmptyOr(customDeployConfig.PluginRuntime.RunDir, pluginConstant.GeneratePluginRunDir(std.DeployInfo().Process.PluginGroup, std.DeployInfo().Process.PluginName)),
		ConfigDir:             pluginConstant.GeneratePluginConfigDir(std.DeployInfo().Process.PluginGroup, std.DeployInfo().Process.PluginName),
		SubConfigDir:          pluginConstant.GeneratePluginSubConfigDir(std.DeployInfo().Process.PluginGroup, std.DeployInfo().Process.PluginName),
		PluginCommonConstants: pluginConstant.GetPluginCommonConstants(std.DeployInfo().Process.PluginPkgName),
		GlobalCommonConstants: pluginConstant.GetGlobalCommonConstants(),
	}

	std.InstanceData().Log().
		Zh("注入插件自定义部署配置成功, 当前部署配置为(%+v), Installer部署配置为(%+v)", std.DeployInfo().BaseRuntime, std.DeployInfo().InstallerRuntime).
		En("inject plugin custom deploy config successfully, current deploy config(%+v), Installer deploy config(%+v)", std.DeployInfo().BaseRuntime, std.DeployInfo().InstallerRuntime).
		Info()

	return nil
}
