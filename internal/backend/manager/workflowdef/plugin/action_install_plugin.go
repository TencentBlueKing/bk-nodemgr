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
	// ActionNameInstallPlugin defines the action name.
	ActionNameInstallPlugin = "install_plugin"

	pluginInstallScriptTimeout = 10 * time.Minute
)

// NewActionInstallPlugin ...
func NewActionInstallPlugin(capability *Capability) action.Definition {
	return &actionInstallPlugin{
		daoHost:             capability.StorageTopo,
		daoPlugin:           capability.StoragePlugin,
		daoPluginDeployment: capability.StoragePlugin,
		provider:            capability.DiscoverProvider,
		gseHandler:          capability.GSEHandler,
	}
}

// ActParamInstallPlugin ...
type ActParamInstallPlugin struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// actionInstallPlugin ...
type actionInstallPlugin struct {
	daoHost             topoStg.IStorageHost
	daoPlugin           pluginStg.IDaoPlugin
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	provider            discover.Discover
	gseHandler          gse.IHandler
}

// Name returns the name of the action.
func (act *actionInstallPlugin) Name() string {
	return ActionNameInstallPlugin
}

// Version returns the version of the action.
func (act *actionInstallPlugin) Version() string {
	return "1.0.0" // nolint: mnd,goconst
}

// Description returns the description of the action.
func (act *actionInstallPlugin) Description() string {
	return ""
}

// Timeout returns the timeout of the action.
func (act *actionInstallPlugin) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionInstallPlugin) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionInstallPlugin) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionInstallPlugin) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionInstallPlugin) Do(ctx *action.InstanceContext) error {
	param := new(ActParamInstallPlugin)
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

	// let the callback server known which action to mark and log.
	if err := std.SaveBlockingActionName(ActionNameWaitPluginInstallerComplete); err != nil {
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

	installParams, err := act.buildInstallParams(std, targetHost, targetPlugin)
	if err != nil {
		return fmt.Errorf("build install params failed: %w", err)
	}

	var (
		installScriptType    types.ScriptType
		installScriptContext string
	)

	if targetHost.Dynamic.NodeOsType == criteria.OSWindows {
		installScriptType, installScriptContext, err = act.buildWindowsInstallScript(installParams)
	} else {
		installScriptType, installScriptContext, err = act.buildUnixInstallScript(installParams)
	}
	if err != nil {
		return fmt.Errorf("build script failed: %w", err)
	}

	std.InstanceData().Log().
		Zh("安装脚本: \n%s\n", installScriptContext).
		En("install script: \n%s\n", installScriptContext).
		Info()

	taskID, err := act.gseHandler.ExecuteScript(nCtx,
		installScriptType,
		installScriptContext,
		pluginInstallScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: targetHost.Dynamic.AgentID,
			},
		})
	if err != nil {
		return err
	}

	std.InstanceData().Log().
		Zh("安装插件任务ID: %s", taskID).
		En("install plugin task-id: %s", taskID).
		Info()

	return nil
}

type pluginInstallParams struct {
	installer.PluginInstallParams

	InstallerWorkDir string
}

func (act *actionInstallPlugin) buildInstallParams(
	std *pluginUtils.PluginActionStandarder,
	targetHost *types.Host,
	targetPlugin *types.Plugin,
) (*pluginInstallParams, error) {
	// select matching tools.
	toolName, err := tool.FormatInstallerName(targetHost.Dynamic.NodeOsType, targetHost.Dynamic.NodeCPUArch)
	if err != nil {
		err = fmt.Errorf("failed to format tools name: %w", err)

		return nil, err
	}

	downloadEndpoints, err := act.provider.SelectEndpoints(
		discover.ServiceNameFile,
		discover.EndpointNameFileDownload,
		pluginUtils.DefaultEndpointSelectionCount,
		discover.NewRoundRobinSelector())
	if err != nil {
		return nil, fmt.Errorf("failed to select file endpoints: %w", err)
	}

	callbackEndpoints, err := act.provider.SelectEndpoints(
		discover.ServiceNameBackend,
		discover.EndpointNameBackendCallback,
		pluginUtils.DefaultEndpointSelectionCount,
		discover.NewRoundRobinSelector())
	if err != nil {
		return nil, fmt.Errorf("failed to select backend callback endpoints: %w", err)
	}

	params := &pluginInstallParams{
		PluginInstallParams: installer.PluginInstallParams{
			PluginCommonParams: installer.PluginCommonParams{
				InstallWorkDir:    std.DeployInfo().InstallerRuntime.WorkDir,
				InstallerFileName: toolName,
				BaseDeployDir:     std.DeployInfo().BaseRuntime.BaseDeployDir,
				BaseWorkDir:       std.DeployInfo().InstallerRuntime.BaseWorkDir,
				DeployEnv:         system.GetEnv(),
			},
			PluginGroup:     targetPlugin.Group,
			PluginName:      targetPlugin.Name,
			PluginVersion:   std.DeployInfo().Process.Info.Version,
			PluginPkgName:   targetPlugin.PkgName,
			CallbackSvrAddr: pluginUtils.BuildServerURLs(callbackEndpoints...),
			DownloadSvrAddr: pluginUtils.BuildServerURLs(downloadEndpoints...),
			DeployToken:     std.Token(),
			OperInstID:      std.InstanceData().OperationInstanceID,
		},
		InstallerWorkDir: std.DeployInfo().InstallerRuntime.WorkDir,
	}

	return params, nil
}

func (act *actionInstallPlugin) buildUnixInstallScript(param *pluginInstallParams) (types.ScriptType, string, error) {
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

func (act *actionInstallPlugin) buildWindowsInstallScript(param *pluginInstallParams) (types.ScriptType, string, error) {
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
func (act *actionInstallPlugin) DisplayNameZh() string {
	return "安装插件"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionInstallPlugin) DisplayNameEn() string {
	return "Install Plugin"
}
