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
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameUpgradePluginV2 defines the action name.
	ActionNameUpgradePluginV2 = "upgrade_plugin_v2"

	pluginUpgradeScriptTimeout = 10 * time.Minute
)

// NewActionUpgradePluginV2 ...
func NewActionUpgradePluginV2(capability *Capability) action.Definition {
	return &actionUpgradePluginV2{
		daoHost:               capability.StorageTopo,
		daoPlugin:             capability.StoragePlugin,
		daoPluginDeployment:   capability.StoragePlugin,
		provider:              capability.DiscoverProvider,
		gseHandler:            capability.GSEHandler,
		storageActionInstance: capability.StorageWorkflow,
	}
}

// ActParamUpgradePluginV2 ...
type ActParamUpgradePluginV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

// actionUpgradePluginV2 ...
type actionUpgradePluginV2 struct {
	daoHost               topoStg.IStorageHost
	daoPlugin             pluginStg.IDaoPlugin
	daoPluginDeployment   pluginStg.IDaoPluginDeployment
	provider              discover.Discover
	gseHandler            gse.IHandler
	storageActionInstance workflow.IStorageActionInstance
}

// Name returns the name of the action.
func (act *actionUpgradePluginV2) Name() string {
	return ActionNameUpgradePluginV2
}

// Version returns the version of the action.
func (act *actionUpgradePluginV2) Version() string {
	return "1.0.0" // nolint: mnd,goconst
}

// Description returns the description of the action.
func (act *actionUpgradePluginV2) Description() string {
	return ""
}

// Timeout returns the timeout of the action.
func (act *actionUpgradePluginV2) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionUpgradePluginV2) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionUpgradePluginV2) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionUpgradePluginV2) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionUpgradePluginV2) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamUpgradePluginV2)
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

	upgradeParams, err := act.buildUpgradeParams(std, targetHost, targetPlugin)
	if err != nil {
		return fmt.Errorf("build upgrade params failed: %w", err)
	}

	var (
		upgradeScriptType    types.ScriptType
		upgradeScriptContext string
	)

	if targetHost.Dynamic.NodeOsType == criteria.OSWindows {
		upgradeScriptType, upgradeScriptContext, err = act.buildWindowsUpgradeScript(upgradeParams)
	} else {
		upgradeScriptType, upgradeScriptContext, err = act.buildUnixUpgradeScript(upgradeParams)
	}
	if err != nil {
		return fmt.Errorf("build script failed: %w", err)
	}

	std.InstanceData().Log().
		Zh("升级脚本: \n%s\n", upgradeScriptContext).
		En("upgrade script: \n%s\n", upgradeScriptContext).
		Info()

	taskID, err := act.gseHandler.ExecuteScript(nCtx,
		upgradeScriptType,
		upgradeScriptContext,
		pluginUpgradeScriptTimeout,
		&types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: targetHost.Dynamic.AgentID,
			},
		})
	if err != nil {
		return err
	}

	std.InstanceData().Log().
		Zh("升级插件任务ID: %s", taskID).
		En("upgrade plugin task-id: %s", taskID).
		Info()

	if err = act.storageActionInstance.UpsertActionInstancePrivateData(std.Context(),
		std.InstanceData().OperationInstanceID,
		ActionNameWaitPluginInstallerCompleteV2,
		map[string]any{
			types.PDKeyActionWaitInstallerCompletePollingSwitch: upgradeParams.SkipCallback,
		}); err != nil {
		return fmt.Errorf("failed to save wait plugin installer private data: %w", err)
	}

	return nil
}

type pluginUpgradeParams struct {
	installer.PluginUpgradeParams

	InstallerWorkDir string
}

func (act *actionUpgradePluginV2) buildUpgradeParams(
	std *pluginV2Utils.PluginActionStandarder,
	targetHost *types.Host,
	targetPlugin *types.Plugin,
) (*pluginUpgradeParams, error) {
	// select matching tools.
	toolName, err := tool.FormatInstallerName(targetHost.Dynamic.NodeOsType, targetHost.Dynamic.NodeCPUArch)
	if err != nil {
		err = fmt.Errorf("failed to format tools name: %w", err)

		return nil, err
	}

	downloadEndpoints, err := act.provider.SelectEndpoints(
		discover.ServiceNameFile,
		discover.EndpointNameFileDownload,
		pluginV2Utils.DefaultEndpointSelectionCount,
		discover.NewRoundRobinSelector())
	if err != nil {
		return nil, fmt.Errorf("failed to select file endpoints: %w", err)
	}

	callbackEndpoints, err := act.provider.SelectEndpoints(
		discover.ServiceNameBackend,
		discover.EndpointNameBackendCallback,
		pluginV2Utils.DefaultEndpointSelectionCount,
		discover.NewRoundRobinSelector())
	if err != nil {
		return nil, fmt.Errorf("failed to select backend callback endpoints: %w", err)
	}

	params := &pluginUpgradeParams{
		PluginUpgradeParams: installer.PluginUpgradeParams{
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
			CallbackSvrAddr: pluginV2Utils.BuildServerURLs(callbackEndpoints...),
			DownloadSvrAddr: pluginV2Utils.BuildServerURLs(downloadEndpoints...),
			DeployToken:     std.Token(),
			OperInstID:      std.InstanceData().OperationInstanceID,
			SkipCallback:    std.DeployInfo().InstallOptions.IsOffline || len(callbackEndpoints) == 0,
			SkipDownload:    std.DeployInfo().InstallOptions.IsOffline || len(downloadEndpoints) == 0,
		},
		InstallerWorkDir: std.DeployInfo().InstallerRuntime.WorkDir,
	}

	return params, nil
}

func (act *actionUpgradePluginV2) buildUnixUpgradeScript(param *pluginUpgradeParams) (types.ScriptType, string, error) {
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

func (act *actionUpgradePluginV2) buildWindowsUpgradeScript(param *pluginUpgradeParams) (types.ScriptType, string, error) {
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
func (act *actionUpgradePluginV2) DisplayNameZh() string {
	return "升级 V2 插件"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionUpgradePluginV2) DisplayNameEn() string {
	return "Upgrade V2 Plugin"
}
