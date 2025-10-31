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

	ctx.Data.LogI(fmt.Sprintf("install script: \n%s\n", installScriptContext))

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

	ctx.Data.LogI("install plugin task-id: " + taskID)

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

	deployConstant, err := deployconstant.GetPluginDeployConf(targetHost.Dynamic.NodeGeneration, targetHost.Dynamic.NodeOsType)
	if err != nil {
		return nil, fmt.Errorf("failed to get deploy constant, err: %w", err)
	}

	randSelector := discover.NewRandomSelector()
	downloadSvrEndpoint, err := act.provider.GetEndpoint(
		discover.ServiceNameFile,
		discover.EndpointNameFileDownload,
		randSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to get file endpoint: %w", err)
	}

	callbackSvrEndpoint, err := act.provider.GetEndpoint(
		discover.ServiceNameBackend,
		discover.EndpointNameBackendCallback,
		randSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to get backend callback endpoint: %w", err)
	}

	params := &pluginInstallParams{
		PluginInstallParams: installer.PluginInstallParams{
			PluginCommonParams: installer.PluginCommonParams{
				InstallWorkDir:    std.DeployInfo().InstallerWorkDir,
				InstallerFileName: toolName,
				BaseDeployDir:     deployConstant.BaseDeployDir,
				BaseWorkDir:       deployConstant.BaseWorkDir,
				DeployEnv:         system.GetEnv(),
			},
			PluginGroup:     targetPlugin.Group,
			PluginName:      targetPlugin.Name,
			PluginVersion:   std.DeployInfo().Process.Info.Version,
			CallbackSvrAddr: "http://" + callbackSvrEndpoint.GetIPV4Address(),
			DownloadSvrAddr: "http://" + downloadSvrEndpoint.GetIPV4Address(),
			DeployToken:     std.Token(),
			OperInstID:      std.InstanceData().OperationInstanceID,
		},
		InstallerWorkDir: std.DeployInfo().InstallerWorkDir,
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
