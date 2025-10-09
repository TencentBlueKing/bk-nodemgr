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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
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
)

// NewActionInstallPlugin ...
func NewActionInstallPlugin(
	daoHost topoStg.IStorageHost,
	daoPluginDeployment pluginStg.IDaoPluginDeployment,
	provider discover.Discover,
	gseHandler gse.IHandler,
) action.Definition {

	return &actionInstallPlugin{
		daoHost:             daoHost,
		daoPluginDeployment: daoPluginDeployment,
		provider:            provider,
		gseHandler:          gseHandler,
	}
}

// ActParamInstallPlugin ...
type ActParamInstallPlugin struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// actionInstallPlugin ...
type actionInstallPlugin struct {
	daoHost             topoStg.IStorageHost
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
	return "1.0.0"
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

	nCtx := contextx.New(ctx.Ctx, contextx.WithTenantID(param.TenantID), contextx.WithBKUsername(param.Operator))

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

	targetHost, err := act.daoHost.GetHostByID(nCtx, std.DeployInfo().Plugin.HostID)
	if err != nil {
		return fmt.Errorf("failed to get host by id. host-id(%d): %w", std.DeployInfo().Plugin.HostID, err)
	}

	// select matching tools.
	toolName, err := tool.FormatInstallerName(targetHost.Dynamic.NodeOsType, targetHost.Dynamic.NodeCPUArch)
	if err != nil {
		err = fmt.Errorf("failed to format tools name: %w", err)

		return err
	}

	deployConstant, err := deployconstant.GetPluginDeployConf(std.DeployInfo().Plugin.Generation, std.DeployInfo().Plugin.Platform.OS)
	if err != nil {
		return fmt.Errorf("failed to get deploy constant, err: %w", err)
	}

	randSelector := discover.NewRandomSelector()
	downloadSvrEndpoint, err := act.provider.GetEndpoint(
		discover.ServiceNameFile,
		discover.EndpointNameFileDownload,
		randSelector)
	if err != nil {
		return fmt.Errorf("failed to get file endpoint: %w", err)
	}

	callbackSvrEndpoint, err := act.provider.GetEndpoint(
		discover.ServiceNameBackend,
		discover.EndpointNameBackendCallback,
		randSelector)
	if err != nil {
		return fmt.Errorf("failed to get backend callback endpoint: %w", err)
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
			PluginType:      string(std.DeployInfo().Plugin.Type),
			PluginName:      std.DeployInfo().Plugin.Name,
			PluginVersion:   std.DeployInfo().Plugin.Version,
			CallbackSvrAddr: "http://" + callbackSvrEndpoint.GetIPV4Address(),
			DownloadSvrAddr: "http://" + downloadSvrEndpoint.GetIPV4Address(),
			DeployToken:     std.Token(),
			OperInstID:      ctx.Data.OperationInstanceID,
		},
		InstallerWorkDir: std.DeployInfo().InstallerWorkDir,
		AgentID:          targetHost.Dynamic.AgentID,
	}

	if targetHost.Dynamic.NodeOsType == criteria.OSWindows {
		return act.doInstallWindows(ctx, params)
	}

	return act.doInstallUnix(ctx, params)
}

type pluginInstallParams struct {
	installer.PluginInstallParams

	InstallerWorkDir string

	AgentID string
}

const pluginInstallScriptTimeout = 10 * time.Minute

func (act *actionInstallPlugin) doInstallUnix(ctx *action.InstanceContext, param *pluginInstallParams) error {
	scriptName, scriptContent, err := param.ToUnixScript()
	if err != nil {
		return err
	}

	ctx.Data.LogI("install plugin script: " + scriptContent)

	taskID, err := act.gseHandler.ExecuteScript(ctx.Ctx,
		types.ScriptTypeBash,
		fmt.Sprintf(
			`mkdir -p %s && cd %s && echo "%s" > %s && sh %s`,
			param.InstallerWorkDir,
			param.InstallerWorkDir,
			scriptContent,
			scriptName,
			scriptName,
		),
		pluginInstallScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: param.AgentID,
			},
		})

	if err != nil {
		return fmt.Errorf("failed to execute plugin install script: %w", err)
	}
	ctx.Data.LogI("install plugin task-id: " + taskID)

	return nil
}

func (act *actionInstallPlugin) doInstallWindows(ctx *action.InstanceContext, param *pluginInstallParams) error {
	scriptName, scriptContent, err := param.ToWindowsScript()
	if err != nil {
		return err
	}

	ctx.Data.LogI("install plugin script: " + scriptContent)

	taskID, err := act.gseHandler.ExecuteScript(ctx.Ctx,
		types.ScriptTypeBat,
		fmt.Sprintf(
			`mkdir "%s" 2>nul & cd "%s" && echo %s > "%s" && cmd /c "%s"`,
			param.InstallerWorkDir,
			param.InstallerWorkDir,
			scriptContent,
			scriptName,
			scriptName,
		),
		pluginInstallScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: param.AgentID,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to execute plugin install script: %w", err)
	}
	ctx.Data.LogI("install plugin task-id: " + taskID)

	return nil
}
