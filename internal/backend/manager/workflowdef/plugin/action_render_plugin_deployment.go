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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameRenderPluginDeployment defines the action name.
	ActionNameRenderPluginDeployment = "render_plugin_deployment"

	// systemUser in gse system user mean use current login user to operate process.
	systemUser = "system"
)

// NewActionRenderPluginDeployment ...
func NewActionRenderPluginDeployment(capability *Capability) action.Definition {
	return &actionRenderPluginDeployment{
		daoHost:             capability.StorageTopo,
		daoPluginPkg:        capability.StorageRelease,
		daoPlugin:           capability.StoragePlugin,
		daoPluginDeployment: capability.StoragePlugin,
	}
}

// ActParamRenderPluginDeployment ...
type ActParamRenderPluginDeployment struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// actionRenderPluginDeployment ...
type actionRenderPluginDeployment struct {
	daoHost             topoStg.IStorageHost
	daoPluginPkg        releaseStg.IPlugin
	daoPlugin           pluginStg.IDaoPlugin
	daoPluginDeployment pluginStg.IDaoPluginDeployment
}

// Name returns the name of the action.
func (act *actionRenderPluginDeployment) Name() string {
	return ActionNameRenderPluginDeployment
}

// Version returns the version of the action.
func (act *actionRenderPluginDeployment) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actionRenderPluginDeployment) Description() string {
	return "render plugin deployment"
}

// Timeout returns the timeout of the action.
func (act *actionRenderPluginDeployment) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionRenderPluginDeployment) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionRenderPluginDeployment) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionRenderPluginDeployment) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen,gocognit
func (act *actionRenderPluginDeployment) Do(ctx *action.InstanceContext) error {
	param := new(ActParamRenderPluginDeployment)
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
	host, err := act.daoHost.GetHostByID(nCtx, std.DeployInfo().Process.HostID)
	if err != nil {
		return fmt.Errorf("failed to get host by id, host-id(%d): %w", std.DeployInfo().Process.HostID, err)
	}

	plugin, err := act.daoPlugin.GetPlugin(nCtx, std.DeployInfo().Process.PluginName)
	if err != nil {
		return fmt.Errorf("failed to get plugin, host-id(%d), plugin-name(%s): %w",
			std.DeployInfo().Process.HostID, std.DeployInfo().Process.PluginName, err)
	}

	version := std.DeployInfo().Process.Info.Version
	if version == "" {
		version = std.DeployInfo().InstallOptions.Version
	}
	pluginPkgName := plugin.PkgName
	pluginGroup := plugin.Group
	pluginName := plugin.Name
	nodeGeneration := host.Dynamic.NodeGeneration
	nodePlatform := platfmt.Platform{
		OS:   host.Dynamic.NodeOsType,
		Arch: host.Dynamic.NodeCPUArch,
	}

	// setting process by host.
	std.DeployInfo().Process.Platform = nodePlatform

	std.DeployInfo().Process.Generation = nodeGeneration

	// setting process by plugin.
	std.DeployInfo().Process.PluginName = pluginName
	std.DeployInfo().Process.PluginPkgName = pluginPkgName
	std.DeployInfo().Process.PluginGroup = pluginGroup
	std.DeployInfo().Process.Info = types.ProcessInfo{
		Version: version,
		AgentID: host.Dynamic.AgentID,
	}

	// setting process by plugin pkg.

	pluginPkg, err := act.daoPluginPkg.GetReleasePlugin(nCtx, types.ReleasePluginKey{
		Generation: nodeGeneration,
		Platform:   nodePlatform,
		Version:    version,
		Name:       pluginPkgName,
	})
	if err != nil {
		return fmt.Errorf("failed to get plugin pkg by name, plugin-pkg-name(%s): %w", pluginPkgName, err)
	}
	if !pluginPkg.Enabled {
		return fmt.Errorf("plugin pkg is not enabled, plugin-pkg-name(%s), version(%s)", pluginPkgName, version)
	}

	std.DeployInfo().Process.Controller = pluginPkg.PluginController

	// setting process by plugin deploy conf.
	pluginDeployConf, err := deployconstant.GetPluginDeployConf(nodeGeneration, nodePlatform.OS)
	if err != nil {
		return fmt.Errorf("failed to get plugin deploy conf, node-generation(%d), node-os(%s), node-arch(%s): %w",
			nodeGeneration, nodePlatform.OS, nodePlatform.Arch, err)
	}

	programName := pluginPkgName
	if nodePlatform.OS == criteria.OSWindows {
		programName += ".exe"
	}

	pidFileName := fmt.Sprintf("%s.pid", pluginPkgName)
	pidFilePath := tool.JoinPath(
		nodePlatform.OS,
		pluginDeployConf.GenerateDefaultRunDir(std.DeployInfo().Process.PluginGroup, std.DeployInfo().Process.PluginName),
		pidFileName,
	)

	setupPath := pluginDeployConf.GenerateDefaultSetupPath(std.DeployInfo().Process.PluginGroup, std.DeployInfo().Process.PluginName)

	var mainConfigPath string
	for _, configTemplate := range pluginPkg.ConfigTemplates {
		if !configTemplate.IsMainConfig {
			continue
		}

		mainConfigPath = tool.JoinPath(
			nodePlatform.OS,
			pluginDeployConf.GenerateDefaultSetupPath(std.DeployInfo().Process.PluginGroup, std.DeployInfo().Process.PluginName),
			configTemplate.FilePath,
			configTemplate.Name,
		)

		break
	}

	logDirPath := pluginDeployConf.LogDir

	// TODO: 接入配置管理
	if host.Dynamic.LoginUser == "" {
		return fmt.Errorf("host login user is empty, host-id(%d)", host.HostID)
	}

	std.DeployInfo().Process.Identity = types.ProcessIdentity{
		Name:       programName,
		SetupPath:  setupPath,
		PidPath:    pidFilePath,
		ConfigPath: mainConfigPath,
		LogPath:    logDirPath,
		User:       host.Dynamic.LoginUser,
	}

	// windows use user direct need provide password, so we use system user to operate the process.
	if host.Dynamic.NodeOsType == criteria.OSWindows {
		std.DeployInfo().Process.Identity.User = systemUser
	}

	// TODO: 接入配置管理
	// nolint: mnd
	std.DeployInfo().Process.Resource = types.ProcessResource{
		CPULimitPercent: 10,
		MemLimitPercent: 10,
	}

	// TODO: 接入配置管理
	// nolint: mnd
	std.DeployInfo().Process.MonitorPolicy = types.ProcessMonitorPolicy{
		RestartType:    types.ProcessRestartTypeAuto,
		StartCheckSecs: 5,
		StopCheckSecs:  5,
		OpTimeoutSecs:  5,
	}

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionRenderPluginDeployment) DisplayNameZh() string {
	return "渲染插件部署"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionRenderPluginDeployment) DisplayNameEn() string {
	return "Render Plugin Deployment"
}
