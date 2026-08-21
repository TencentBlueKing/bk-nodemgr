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

package plugin

import (
	"errors"
	"fmt"
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/tool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/system"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNamePrepareDebugProcess defines the action name.
	ActionNamePrepareDebugProcess = "prepare_debug_process"

	// defaultDebugTimeoutSec defines the default debug timeout in seconds.
	defaultDebugTimeoutSec = 250
)

// NewActionPrepareDebugProcess creates a debug process preparation action.
func NewActionPrepareDebugProcess(capability *Capability) action.Definition {
	return &actPrepareDebugProcess{
		daoPluginDeployment: capability.StoragePlugin,
		daoProcess:          capability.StoragePlugin,
	}
}

// ActionParamPrepareDebugProcess defines parameters for debug process preparation.
type ActionParamPrepareDebugProcess struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

type actPrepareDebugProcess struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoProcess          pluginStg.IDaoProcess
}

// Name returns the action name.
func (act *actPrepareDebugProcess) Name() string {
	return ActionNamePrepareDebugProcess
}

// Version returns the action version.
func (act *actPrepareDebugProcess) Version() string {
	return "1.0.0"
}

// Description returns the action description.
func (act *actPrepareDebugProcess) Description() string {
	return "prepare debug process"
}

// Timeout returns the action timeout.
func (act *actPrepareDebugProcess) Timeout() time.Duration {
	return time.Minute
}

// Tags returns action tags.
func (act *actPrepareDebugProcess) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count.
func (act *actPrepareDebugProcess) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn returns the retry delay function.
func (act *actPrepareDebugProcess) DelayFn(_ int) func() {
	return func() {
		time.Sleep(time.Second)
	}
}

// Do prepares the debug process.
func (act *actPrepareDebugProcess) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActionParamPrepareDebugProcess)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	std := pluginUtils.NewPluginActionStandarder(act.daoPluginDeployment)
	if err = std.Initialize(ctx, param.PluginActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	process := std.DeployInfo().Process
	process.Controller.StartCmd, err = buildDebugCommand(std, "start", process.Controller.DebugCmd)
	if err != nil {
		return fmt.Errorf("failed to build debug start command: %w", err)
	}
	process.MonitorPolicy.RestartType = types.ProcessRestartTypeManual
	process.MonitorPolicy.OpTimeoutSecs = defaultDebugTimeoutSec

	std.DeployInfo().Process = process
	std.InstanceData().Log().
		Zh("成功准备插件调试进程, plugin-name(%s), host-id(%d)", process.PluginName, process.HostID).
		En("succeeded to prepare plugin debug process, plugin-name(%s), host-id(%d)", process.PluginName, process.HostID).
		Info()

	return nil
}

// DisplayNameZh returns the Chinese display name.
func (act *actPrepareDebugProcess) DisplayNameZh() string {
	return "准备调试进程"
}

// DisplayNameEn returns the English display name.
func (act *actPrepareDebugProcess) DisplayNameEn() string {
	return "Prepare Debug Process"
}

func buildDebugCommand(std *pluginUtils.PluginActionStandarder, debugAction string, runCmd string) (string, error) {
	process := std.DeployInfo().Process
	installerFileName, err := tool.FormatInstallerName(process.Platform.OS, process.Platform.Arch)
	if err != nil {
		return "", fmt.Errorf("failed to format installer name: %w", err)
	}

	params := &installer.PluginDebugParams{
		PluginCommonParams: installer.PluginCommonParams{
			InstallWorkDir:    std.DeployInfo().InstallerRuntime.WorkDir,
			InstallerFileName: installerFileName,
			BaseDeployDir:     std.DeployInfo().BaseRuntime.BaseDeployDir,
			BaseWorkDir:       std.DeployInfo().InstallerRuntime.BaseWorkDir,
			DeployEnv:         system.GetEnv(),
		},
		PluginGroup: process.PluginGroup,
		PluginName:  process.PluginName,
		DeployToken: std.Token(),
		OperInstID:  std.InstanceData().OperationInstanceID,
		DebugAction: debugAction,
		RunCmd:      runCmd,
		PidDir:      tool.JoinPath(process.Platform.OS, std.DeployInfo().BaseRuntime.RunDir, "debug", std.Token()),
	}

	if process.Platform.OS == criteria.OSWindows {
		_, command, buildErr := params.ToWindowsScript()
		return command, buildErr
	}

	_, command, buildErr := params.ToUnixScript()

	return command, buildErr
}
