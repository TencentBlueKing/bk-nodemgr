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
	// ActionNameRenderPluginDeploymentV2 defines the action name.
	ActionNameRenderPluginDeploymentV2 = "render_plugin_deployment_v2"
)

// NewActionRenderPluginDeploymentV2 ...
func NewActionRenderPluginDeploymentV2(capability *Capability) action.Definition {
	return &actionRenderPluginDeploymentV2{
		daoHost:             capability.StorageTopo,
		daoPluginPkg:        capability.StorageRelease,
		daoPluginDeployment: capability.StoragePlugin,
	}
}

// ActParamRenderPluginDeploymentV2 ...
type ActParamRenderPluginDeploymentV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

// actionRenderPluginDeploymentV2 ...
type actionRenderPluginDeploymentV2 struct {
	daoHost             topoStg.IStorageHost
	daoPluginPkg        releaseStg.IPlugin
	daoPluginDeployment pluginStg.IDaoPluginDeployment
}

// Name returns the name of the action.
func (act *actionRenderPluginDeploymentV2) Name() string {
	return ActionNameRenderPluginDeploymentV2
}

// Version returns the version of the action.
func (act *actionRenderPluginDeploymentV2) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actionRenderPluginDeploymentV2) Description() string {
	return "render plugin deployment v2"
}

// Timeout returns the timeout of the action.
func (act *actionRenderPluginDeploymentV2) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionRenderPluginDeploymentV2) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionRenderPluginDeploymentV2) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionRenderPluginDeploymentV2) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen,gocognit
func (act *actionRenderPluginDeploymentV2) Do(ctx *action.InstanceContext) error {
	param := new(ActParamRenderPluginDeploymentV2)
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
	if err = pluginV2Utils.EnsureHostLoginUser(std, host); err != nil {
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

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionRenderPluginDeploymentV2) DisplayNameZh() string {
	return "渲染 V2 插件部署"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionRenderPluginDeploymentV2) DisplayNameEn() string {
	return "Render V2 Plugin Deployment"
}
