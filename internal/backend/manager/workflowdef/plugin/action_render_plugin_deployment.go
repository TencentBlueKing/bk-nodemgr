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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameRenderPluginDeployment defines the action name.
	ActionNameRenderPluginDeployment = "render_plugin_deployment"
)

// NewActionRenderPluginDeployment ...
func NewActionRenderPluginDeployment(capability *Capability) action.Definition {
	return &actionRenderPluginDeployment{
		daoHost:             capability.StorageTopo,
		daoPluginPkg:        capability.StorageRelease,
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
func (act *actionRenderPluginDeployment) DelayFn(_ int) func() {
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

	version := std.DeployInfo().Process.Info.Version
	if version == "" {
		version = std.DeployInfo().InstallOptions.Version
	}

	// setting process by plugin.
	std.DeployInfo().Process.Info = types.ProcessInfo{
		Version: version,
		AgentID: host.Dynamic.AgentID,
	}

	// setting process by plugin pkg.

	pluginPkg, err := act.daoPluginPkg.GetReleasePlugin(nCtx, types.ReleasePluginKey{
		Generation: std.DeployInfo().Process.Generation,
		Platform:   std.DeployInfo().Process.Platform,
		Version:    version,
		Name:       std.DeployInfo().Process.PluginPkgName,
	})
	if err != nil {
		return fmt.Errorf("failed to get plugin pkg by name, plugin-pkg-name(%s): %w", std.DeployInfo().Process.PluginPkgName, err)
	}
	if !pluginPkg.Enabled {
		return fmt.Errorf("plugin pkg is not enabled, plugin-pkg-name(%s), version(%s)", std.DeployInfo().Process.PluginPkgName, version)
	}

	std.DeployInfo().Process.Controller = pluginPkg.PluginController

	programName := std.DeployInfo().Process.PluginPkgName
	if std.DeployInfo().Process.Platform.OS == criteria.OSWindows {
		programName += ".exe"
	}

	pidFileName := fmt.Sprintf("%s.pid", std.DeployInfo().Process.PluginPkgName)
	pidFilePath := tool.JoinPath(std.DeployInfo().Process.Platform.OS, std.DeployInfo().BaseRuntime.RunDir, pidFileName)

	var mainConfigPath string
	for _, configTemplate := range pluginPkg.ConfigTemplates {
		if !configTemplate.IsMainConfig {
			continue
		}

		mainConfigPath = tool.JoinPath(std.DeployInfo().Process.Platform.OS, std.DeployInfo().BaseRuntime.PluginHomeDir,
			configTemplate.FilePath, configTemplate.Name)

		break
	}

	// TODO: 接入配置管理
	if err = pluginUtils.EnsureHostLoginUser(std, host); err != nil {
		return err
	}

	std.DeployInfo().Process.Identity = types.ProcessIdentity{
		Name:       programName,
		SetupPath:  std.DeployInfo().BaseRuntime.PluginHomeDir,
		PidPath:    pidFilePath,
		ConfigPath: mainConfigPath,
		LogPath:    std.DeployInfo().BaseRuntime.LogDir,
		User:       host.Dynamic.LoginUser,
	}

	// windows use user direct need provide password, so we use system user to operate the process.
	if host.Dynamic.NodeOsType == criteria.OSWindows {
		std.DeployInfo().Process.Identity.User = gse.WindowsOperateUser
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

	if std.DeployInfo().InstallOptions.CustomSpec != nil {
		std.DeployInfo().Process.Resource = std.DeployInfo().InstallOptions.CustomSpec.Resource
		std.DeployInfo().Process.MonitorPolicy = std.DeployInfo().InstallOptions.CustomSpec.MonitorPolicy
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
