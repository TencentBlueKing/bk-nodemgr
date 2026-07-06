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
	"strings"
	"time"

	pluginV2Utils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/pluginv2/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameInjectPluginBaseRuntimeV2 defines the action name.
	ActionNameInjectPluginBaseRuntimeV2 = "inject_plugin_base_runtime_v2"
)

// NewActionInjectPluginBaseRuntimeV2 get a new action.
func NewActionInjectPluginBaseRuntimeV2(capability *Capability) action.Definition {
	return &actionInjectPluginBaseRuntimeV2{
		daoPluginDeployment: capability.StoragePlugin,
		daoHost:             capability.StorageTopo,
		daoDomainGse:        capability.StorageTopo,
	}
}

// ActParamInjectPluginBaseRuntimeV2 this is the param for render deployment.
type ActParamInjectPluginBaseRuntimeV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

type actionInjectPluginBaseRuntimeV2 struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoHost             topoStg.IStorageHost
	daoDomainGse        topoStg.IStorageDomainGse
}

// Name returns the name of the action.
func (act *actionInjectPluginBaseRuntimeV2) Name() string {
	return ActionNameInjectPluginBaseRuntimeV2
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionInjectPluginBaseRuntimeV2) DisplayNameZh() string {
	return "注入 V2 插件基础运行时"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionInjectPluginBaseRuntimeV2) DisplayNameEn() string {
	return "Inject V2 Plugin Base Runtime"
}

// Version returns the version of the action.
func (act *actionInjectPluginBaseRuntimeV2) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionInjectPluginBaseRuntimeV2) Description() string {
	return "inject plugin base runtime v2"
}

// Timeout returns the timeout of the action.
func (act *actionInjectPluginBaseRuntimeV2) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionInjectPluginBaseRuntimeV2) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionInjectPluginBaseRuntimeV2) MaxRetryCount() uint {
	return 1 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionInjectPluginBaseRuntimeV2) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: lll
func (act *actionInjectPluginBaseRuntimeV2) Do(ctx *action.InstanceContext) error {
	param := new(ActParamInjectPluginBaseRuntimeV2)
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

	pluginV2Constant := pluginV2DeployConf{
		PluginDeployConf: pluginConstant,
	}
	nodeRole := host.Dynamic.NodeRole
	// update deploy constant with custom deploy config if custom deploy config exists
	if customDeployConfig != nil {
		// override BaseDeployDir and generate corresponding configuration
		if customDeployConfig.PluginRuntime.BaseDeployDir != "" {
			pluginV2Constant.BaseDeployDir = customDeployConfig.PluginRuntime.BaseDeployDir
		}

		if customDeployConfig.NodeRuntime.BaseDeployDir != "" {
			nodeConstant.BaseDeployDir = customDeployConfig.NodeRuntime.BaseDeployDir
		}

		// override BaseWorkDir and generate corresponding configuration
		if customDeployConfig.InstallerRuntime.BaseWorkDir != "" {
			pluginV2Constant.BaseWorkDir = customDeployConfig.InstallerRuntime.BaseWorkDir
		}
	}

	// installer workdir priority: user specified in info > network unit custom deploy config > deploy constant default.
	if std.DeployInfo().InstallerRuntime.BaseWorkDir != "" {
		pluginV2Constant.BaseWorkDir = std.DeployInfo().InstallerRuntime.BaseWorkDir
	}

	// installer workdir priority: user specified in info > network unit custom deploy config > deploy constant default.
	std.DeployInfo().InstallerRuntime.BaseWorkDir = pluginV2Constant.BaseWorkDir
	std.DeployInfo().InstallerRuntime.WorkDir = pluginV2Constant.GenerateWorkDir()

	act.injectBaseRuntime(std, pluginV2Constant, nodeConstant, nodeRole, customDeployConfig)

	std.InstanceData().Log().
		Zh("注入插件 V2 基础运行时成功, 当前部署配置为(%+v), Installer部署配置为(%+v)",
			std.DeployInfo().BaseRuntime, std.DeployInfo().InstallerRuntime).
		En("inject plugin base runtime successfully, current deploy config(%+v), Installer deploy config(%+v)",
			std.DeployInfo().BaseRuntime, std.DeployInfo().InstallerRuntime).
		Info()

	return nil
}

func (act *actionInjectPluginBaseRuntimeV2) injectBaseRuntime(std *pluginV2Utils.PluginActionStandarder,
	pluginConstant pluginV2DeployConf, nodeConstant deployconstant.NodeDeployConf, nodeRole types.NodeRole,
	customDeployConfig *types.CustomDeployConfig) {

	pluginHomeDir := pluginConstant.GeneratePluginV2HomeDir()
	dataIPC := conv.NonEmptyOr(customDeployConfig.NodeRuntime.DataIPC, nodeConstant.GenerateDataIPC(nodeRole))
	pluginIPC := conv.NonEmptyOr(customDeployConfig.NodeRuntime.PluginIPC, nodeConstant.GeneratePluginIPC(nodeRole))
	logDir := tool.JoinPath(nodeConstant.OsType, nodeConstant.LogDir, pluginV2BaseDirName)
	std.DeployInfo().BaseRuntime = types.PluginDeploymentBaseRuntime{
		BaseDeployDir:         pluginConstant.BaseDeployDir,
		DeployDir:             pluginConstant.GeneratePluginDeployDir(),
		GSEHomeDir:            nodeConstant.GenerateNodeHomeDir(nodeRole),
		PluginHomeDir:         pluginHomeDir,
		DataIPC:               dataIPC,
		PluginIPC:             pluginIPC,
		HostIDPath:            pluginConstant.HostIDPath,
		LogDir:                logDir,
		DataDir:               pluginConstant.GeneratePluginV2DataDir(),
		RunDir:                pluginConstant.GeneratePluginV2RunDir(),
		ConfigDir:             pluginConstant.GeneratePluginV2ConfigDir(),
		SubConfigDir:          pluginConstant.GeneratePluginV2SubConfigDir(std.DeployInfo().Process.PluginName),
		PluginCommonConstants: pluginConstant.GetPluginCommonConstants(std.DeployInfo().Process.PluginPkgName),
		GlobalCommonConstants: pluginConstant.GetGlobalCommonConstants(),
	}
}

// pluginV2DeployConf defines the deployment configuration for agent.
type pluginV2DeployConf struct {
	deployconstant.PluginDeployConf
}

const (
	pluginV2BaseDirName   = "plugins"
	pluginV2ConfigDirName = "etc"
	pluginV2RunDirName    = "run"
	pluginV2DataDirName   = "data"
)

// GeneratePluginV2HomeDir generates the plugin v2 home directory based on plugin group and plugin name.
func (conf pluginV2DeployConf) GeneratePluginV2HomeDir() string {
	pluginHomeDir := tool.JoinPath(conf.OsType, conf.GeneratePluginDeployDir())
	// Replace the plugins folder at the end of the path with the plugins folder
	pluginV2HomeDir := strings.Replace(pluginHomeDir, deployconstant.PluginBaseDirName, pluginV2BaseDirName, 1)

	return pluginV2HomeDir
}

// GeneratePluginV2DataDir generates the run directory.
func (conf pluginV2DeployConf) GeneratePluginV2DataDir() string {
	return tool.JoinPath(conf.OsType, conf.GeneratePluginV2HomeDir(), pluginV2DataDirName)
}

// GeneratePluginV2RunDir generates the run directory.
func (conf pluginV2DeployConf) GeneratePluginV2RunDir() string {
	return tool.JoinPath(conf.OsType, conf.GeneratePluginV2HomeDir(), pluginV2RunDirName)
}

// GeneratePluginV2ConfigDir generates the configuration directory.
func (conf pluginV2DeployConf) GeneratePluginV2ConfigDir() string {
	return tool.JoinPath(conf.OsType, conf.GeneratePluginV2HomeDir(), pluginV2ConfigDirName)
}

// GeneratePluginV2SubConfigDir generates the sub-configuration directory.
func (conf pluginV2DeployConf) GeneratePluginV2SubConfigDir(pluginName string) string {
	return tool.JoinPath(conf.OsType, conf.GeneratePluginV2ConfigDir(), pluginName)
}
