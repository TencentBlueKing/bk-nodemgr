/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.

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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameCleanDebugPlugin defines the action name.
	ActionNameCleanDebugPlugin = "clean_debug_plugin"

	cleanDebugPluginScriptTimeout   = time.Minute
	cleanDebugPluginPollingInterval = 5 * time.Second
)

// NewActionCleanDebugPlugin creates a debug plugin cleanup action.
func NewActionCleanDebugPlugin(capability *Capability) action.Definition {
	return &actCleanDebugPlugin{
		daoPluginDeployment: capability.StoragePlugin,
		gseHandler:          capability.GSEHandler,
	}
}

// ActionParamCleanDebugPlugin defines parameters for debug plugin cleanup.
type ActionParamCleanDebugPlugin struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

type actCleanDebugPlugin struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	gseHandler          gse.IHandler
}

// Name returns the action name.
func (act *actCleanDebugPlugin) Name() string { return ActionNameCleanDebugPlugin }

// Version returns the action version.
func (act *actCleanDebugPlugin) Version() string { return "1.0.0" }

// Description returns the action description.
func (act *actCleanDebugPlugin) Description() string { return "clean debug plugin directory" }

// Timeout returns the action timeout.
func (act *actCleanDebugPlugin) Timeout() time.Duration {
	return 2 * time.Minute // nolint: mnd
}

// Tags returns action tags.
func (act *actCleanDebugPlugin) Tags() []action.Tag { return []action.Tag{} }

// MaxRetryCount returns the max retry count.
func (act *actCleanDebugPlugin) MaxRetryCount() uint { return 3 } // nolint: mnd

// DelayFn returns the retry delay function.
func (act *actCleanDebugPlugin) DelayFn(_ int) func() {
	return func() { time.Sleep(time.Second) }
}

// Do removes the debug plugin directory through the installer.
func (act *actCleanDebugPlugin) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamCleanDebugPlugin)
	err := conv.MapToStruct(ctx.Data.Content, param)
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

	command, err := buildDebugCommand(std, "clean", "")
	if err != nil {
		return fmt.Errorf("failed to build debug clean command: %w", err)
	}

	process := std.DeployInfo().Process
	scriptType := types.ScriptTypeBash
	if process.Platform.OS == criteria.OSWindows {
		scriptType = types.ScriptTypePowershell
	}
	endpoint := types.Endpoint{AgentID: process.Info.AgentID}
	taskID, err := act.gseHandler.ExecuteScript(
		std.Context(), scriptType, command, cleanDebugPluginScriptTimeout,
		&types.EndpointWithAuth{Endpoint: endpoint},
	)
	if err != nil {
		return fmt.Errorf("failed to submit debug plugin cleanup: %w", err)
	}

	std.InstanceData().Log().
		Zh("执行清理插件调试目录操作, plugin-name(%s), host-id(%d)", process.PluginName, process.HostID).
		En("execute clean plugin debug directory operation, plugin-name(%s), host-id(%d)", process.PluginName, process.HostID).
		Info()

	if err = act.waitCleanupResult(std, taskID, endpoint); err != nil {
		return err
	}

	std.InstanceData().Log().
		Zh("成功清理插件调试目录, plugin-name(%s), host-id(%d)", process.PluginName, process.HostID).
		En("succeeded to clean plugin debug directory, plugin-name(%s), host-id(%d)", process.PluginName, process.HostID).
		Info()

	return nil
}

func (act *actCleanDebugPlugin) waitCleanupResult(std *pluginUtils.PluginActionStandarder, taskID string, endpoint types.Endpoint) error {
	ticker := time.NewTicker(cleanDebugPluginPollingInterval)
	defer ticker.Stop()

	for {
		results, err := act.gseHandler.QueryScriptExecutionResult(
			std.Context(), taskID, &types.EndpointWithRestrict{Endpoint: endpoint})
		if err != nil {
			return fmt.Errorf("failed to query debug plugin cleanup result: %w", err)
		}
		if len(results) > 1 {
			return fmt.Errorf("debug plugin cleanup result count is not 1, count(%d)", len(results))
		}
		if len(results) == 1 {
			result := results[0]
			switch result.Status {
			case types.ScriptStatusFinished:
				return nil
			case types.ScriptStatusFailed, types.ScriptStatusTimeout, types.ScriptStatusStopped,
				types.ScriptStatusAgentRestarted:
				return fmt.Errorf(
					"debug plugin cleanup failed, status(%s), exit-code(%d), err-code(%d), err-msg(%s), screen(%s)",
					result.Status, result.ExitCode, result.ErrorCode, result.ErrorMessage, result.ScreenLog)
			case types.ScriptStatusUnknown, types.ScriptStatusReceived, types.ScriptStatusRunning:
				// Cleanup task is still in progress, keep polling.
			}
		}

		select {
		case <-std.Context().Done():
			return fmt.Errorf("debug plugin cleanup interrupted: %w", std.Context().Err())
		case <-ticker.C:
		}
	}
}

// DisplayNameZh returns the Chinese display name.
func (act *actCleanDebugPlugin) DisplayNameZh() string { return "清理调试插件目录" }

// DisplayNameEn returns the English display name.
func (act *actCleanDebugPlugin) DisplayNameEn() string { return "Clean Debug Plugin Directory" }
