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
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameUninstallPluginV2 defines the action name.
	ActionNameUninstallPluginV2 = "uninstall_plugin_v2"

	pluginUninstallScriptTimeout = 10 * time.Minute
)

// NewActionUninstallPluginV2 ...
func NewActionUninstallPluginV2(capability *Capability) action.Definition {
	return &actionUninstallPluginV2{
		daoHost:             capability.StorageTopo,
		daoPlugin:           capability.StoragePlugin,
		daoPluginDeployment: capability.StoragePlugin,
		provider:            capability.DiscoverProvider,
		gseHandler:          capability.GSEHandler,
	}
}

// ActParamUninstallPluginV2 ...
type ActParamUninstallPluginV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

// actionUninstallPluginV2 ...
type actionUninstallPluginV2 struct {
	daoHost             topoStg.IStorageHost
	daoPlugin           pluginStg.IDaoPlugin
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	provider            discover.Discover
	gseHandler          gse.IHandler
}

// Name returns the name of the action.
func (act *actionUninstallPluginV2) Name() string {
	return ActionNameUninstallPluginV2
}

// Version returns the version of the action.
func (act *actionUninstallPluginV2) Version() string {
	return "1.0.0" // nolint: mnd,goconst
}

// Description returns the description of the action.
func (act *actionUninstallPluginV2) Description() string {
	return "uninstall plugin v2"
}

// Timeout returns the timeout of the action.
func (act *actionUninstallPluginV2) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionUninstallPluginV2) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionUninstallPluginV2) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionUninstallPluginV2) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionUninstallPluginV2) Do(ctx *action.InstanceContext) error {
	param := new(ActParamUninstallPluginV2)
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

	if param.SkipAction {
		std.InstanceData().Log().
			Zh("跳过卸载 V2 插件, 主机id(%d), 插件名(%s)", std.DeployInfo().Process.HostID, std.DeployInfo().Process.PluginName).
			En("skip uninstall V2 plugin, host-id(%d), plugin-name(%s)", std.DeployInfo().Process.HostID, std.DeployInfo().Process.PluginName).
			Info()

		return nil
	}

	// let the callback server known which action to mark and log.
	if err := std.SaveBlockingActionName(ActionNameWaitPluginInstallerCompleteV2); err != nil {
		return fmt.Errorf("failed to save blocking action name: %w", err)
	}

	nCtx := std.Context()
	targetHost, err := act.daoHost.GetHostByID(nCtx, std.DeployInfo().Process.HostID)
	if err != nil {
		return fmt.Errorf("failed to get host: %w", err)
	}

	targetPlugin, err := act.daoPlugin.GetPlugin(nCtx, std.DeployInfo().Process.PluginName)
	if err != nil {
		return fmt.Errorf("failed to get plugin: %w", err)
	}

	uninstallParams, err := act.buildUninstallParams(std, targetHost, targetPlugin)
	if err != nil {
		return fmt.Errorf("build uninstall params failed: %w", err)
	}

	var (
		uninstallScriptType    types.ScriptType
		uninstallScriptContext string
	)

	if targetHost.Dynamic.NodeOsType == criteria.OSWindows {
		uninstallScriptType, uninstallScriptContext, err = act.buildWindowsUninstallScript(uninstallParams)
	} else {
		uninstallScriptType, uninstallScriptContext, err = act.buildUnixUninstallScript(uninstallParams)
	}
	if err != nil {
		return fmt.Errorf("build script failed: %w", err)
	}

	std.InstanceData().Log().
		Zh("卸载脚本: \n%s\n", uninstallScriptContext).
		En("uninstall script: \n%s\n", uninstallScriptContext).
		Info()

	taskID, err := act.gseHandler.ExecuteScript(nCtx,
		uninstallScriptType,
		uninstallScriptContext,
		pluginUninstallScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: targetHost.Dynamic.AgentID,
			},
		})
	if err != nil {
		return err
	}

	std.InstanceData().Log().
		Zh("卸载插件任务ID: %s", taskID).
		En("uninstall plugin task-id: %s", taskID).
		Info()

	return nil
}

type pluginV2UninstallParams struct {
	installer.PluginV2UninstallParams

	InstallerWorkDir string
}

func (act *actionUninstallPluginV2) buildUninstallParams(
	std *pluginV2Utils.PluginActionStandarder,
	targetHost *types.Host,
	targetPlugin *types.Plugin,
) (*pluginV2UninstallParams, error) {
	// select matching tools.
	toolName, err := tool.FormatInstallerName(targetHost.Dynamic.NodeOsType, targetHost.Dynamic.NodeCPUArch)
	if err != nil {
		err = fmt.Errorf("failed to format tools name: %w", err)

		return nil, err
	}

	callbackEndpoints, err := act.provider.SelectEndpoints(
		discover.ServiceNameBackend,
		discover.EndpointNameBackendCallback,
		pluginV2Utils.DefaultEndpointSelectionCount,
		discover.NewRoundRobinSelector())
	if err != nil {
		return nil, fmt.Errorf("failed to select backend callback endpoints: %w", err)
	}

	params := &pluginV2UninstallParams{
		PluginV2UninstallParams: installer.PluginV2UninstallParams{
			PluginV2CommonParams: installer.PluginV2CommonParams{
				InstallWorkDir:    std.DeployInfo().InstallerRuntime.WorkDir,
				InstallerFileName: toolName,
				BaseDeployDir:     std.DeployInfo().BaseRuntime.BaseDeployDir,
				BaseWorkDir:       std.DeployInfo().InstallerRuntime.BaseWorkDir,
				DeployEnv:         system.GetEnv(),
			},
			PluginGroup:     targetPlugin.Group,
			PluginName:      targetPlugin.Name,
			CallbackSvrAddr: pluginV2Utils.BuildServerURLs(callbackEndpoints...),
			DeployToken:     std.Token(),
			OperInstID:      std.InstanceData().OperationInstanceID,
		},
		InstallerWorkDir: std.DeployInfo().InstallerRuntime.WorkDir,
	}

	return params, nil
}

func (act *actionUninstallPluginV2) buildUnixUninstallScript(param *pluginV2UninstallParams) (types.ScriptType, string, error) {
	scriptName, scriptContent, err := param.ToUnixScript()
	if err != nil {
		return "", "", err
	}

	scriptContent = fmt.Sprintf(
		`mkdir -p %s && cd %s && echo "%s" > %s && sh %s`,
		param.InstallerWorkDir,
		param.InstallerWorkDir,
		scriptContent,
		scriptName,
		scriptName,
	)

	return types.ScriptTypeBash, scriptContent, nil
}

func (act *actionUninstallPluginV2) buildWindowsUninstallScript(param *pluginV2UninstallParams) (types.ScriptType, string, error) {
	scriptName, scriptContent, err := param.ToWindowsScript()
	if err != nil {
		return "", "", err
	}

	scriptContent = fmt.Sprintf(
		`mkdir "%s" 2>nul & cd "%s" && echo %s > "%s" && cmd /c "%s"`,
		param.InstallerWorkDir,
		param.InstallerWorkDir,
		scriptContent,
		scriptName,
		scriptName,
	)

	return types.ScriptTypeBat, scriptContent, nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionUninstallPluginV2) DisplayNameZh() string {
	return "卸载 V2 插件"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionUninstallPluginV2) DisplayNameEn() string {
	return "Uninstall Plugin V2"
}
