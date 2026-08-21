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
	"strings"
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	workflowStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/common"
)

const (
	// ActionNameRunDebugPlugin defines the action name.
	ActionNameRunDebugPlugin = "run_debug_plugin"
)

// NewActionRunDebugPlugin creates a plugin debug action.
func NewActionRunDebugPlugin(capability *Capability) action.Definition {
	return &actRunDebugPlugin{
		daoPluginDeployment:   capability.StoragePlugin,
		storageActionInstance: capability.StorageWorkflow,
		gseHandler:            capability.GSEHandler,
	}
}

// ActionParamRunDebugPlugin defines parameters for running plugin debug.
type ActionParamRunDebugPlugin struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

type actRunDebugPlugin struct {
	daoPluginDeployment   pluginStg.IDaoPluginDeployment
	storageActionInstance workflowStg.IStorage
	gseHandler            gse.IHandler
}

// Name returns the action name.
func (act *actRunDebugPlugin) Name() string {
	return ActionNameRunDebugPlugin
}

// Version returns the action version.
func (act *actRunDebugPlugin) Version() string {
	return "1.0.0"
}

// Description returns the action description.
func (act *actRunDebugPlugin) Description() string {
	return "run plugin debug task and stream output"
}

// Timeout returns the action timeout.
func (act *actRunDebugPlugin) Timeout() time.Duration {
	return 5 * time.Minute // nolint: mnd
}

// Tags returns action tags.
func (act *actRunDebugPlugin) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count.
func (act *actRunDebugPlugin) MaxRetryCount() uint {
	return 0
}

// DelayFn returns the retry delay function.
func (act *actRunDebugPlugin) DelayFn(_ int) func() {
	return func() {}
}

// Do runs the plugin debug task and streams output until completion or stop.
func (act *actRunDebugPlugin) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActionParamRunDebugPlugin)
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

	nCtx := std.Context()
	process := std.DeployInfo().Process
	debugTimeout := time.Duration(process.MonitorPolicy.OpTimeoutSecs) * time.Second

	scriptType := types.ScriptTypeBash
	if process.Platform.OS == criteria.OSWindows {
		scriptType = types.ScriptTypePowershell
	}
	endpoint := types.Endpoint{AgentID: process.Info.AgentID}
	taskID, err := act.gseHandler.ExecuteScript(nCtx, scriptType, process.Controller.StartCmd, debugTimeout, &types.EndpointWithAuth{
		Endpoint: endpoint,
	})
	if err != nil {
		return fmt.Errorf("failed to start debug task: %w", err)
	}

	var (
		pendingOutput string
		taskEnded     bool
	)
	defer func() {
		flushErr := act.storeDebugOutput(std, "", true, &pendingOutput)

		var terminateErr error
		if !taskEnded {
			terminateErr = act.terminateDebugTask(std, taskID, endpoint)
		}
		err = errors.Join(err, flushErr, terminateErr)
	}()

	std.InstanceData().Log().
		Zh("成功启动插件调试任务, task-id(%s), debug-cmd(%s)", taskID, process.Controller.StartCmd).
		En("succeeded to start plugin debug task, task-id(%s), debug-cmd(%s)", taskID, process.Controller.StartCmd).
		Info()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var outputOffset uint
	for {
		stop, stopErr := act.stopDebugRequested(std)
		if stopErr != nil {
			return stopErr
		}
		if stop {
			std.InstanceData().Log().
				Zh("收到调试停止信号").
				En("received debug stop signal").
				Info()

			return nil
		}

		scriptResult, pollErr := act.pollDebugTaskOnce(std, taskID, endpoint, &outputOffset, &pendingOutput)
		if pollErr != nil {
			return pollErr
		}
		if isTerminalDebugTaskStatus(scriptResult.Status) {
			taskEnded = true
			logTerminalDebugTaskStatus(std, scriptResult)

			return nil
		}

		select {
		case <-nCtx.Done():
			return nCtx.Err()
		case <-ticker.C:
		}
	}
}

func logTerminalDebugTaskStatus(std *pluginUtils.PluginActionStandarder, result *types.ScriptResult) {
	log := std.InstanceData().Log().
		Zh("插件调试任务已结束, 调试任务状态(%s), 错误码(%d), 错误信息(%s)", result.Status, result.ErrorCode, result.ErrorMessage).
		En("plugin debug task ended, status(%s)， error-code(%d), error-message(%s)", result.Status, result.ErrorCode, result.ErrorMessage)

	switch result.Status {
	case types.ScriptStatusFinished, types.ScriptStatusStopped:
		log.Info()
	case types.ScriptStatusFailed, types.ScriptStatusTimeout, types.ScriptStatusAgentRestarted:
		log.Error()
	default:
		log.Warn()
	}
}

// stopDebugRequested reports whether a debug stop signal was stored for the run action instance.
func (act *actRunDebugPlugin) stopDebugRequested(std *pluginUtils.PluginActionStandarder) (bool, error) {
	privateData, err := act.storageActionInstance.GetActionInstancePrivateData(
		std.Context(), std.InstanceData().OperationInstanceID, ActionNameRunDebugPlugin)
	if err != nil {
		return false, fmt.Errorf("failed to get debug wait private data: %w", err)
	}

	stopSignal, ok := privateData[types.PDKeyDebugStopSignal].(bool)

	return ok && stopSignal, nil
}

// pollDebugTaskOnce queries the debug task once, streams any new output, and returns the latest task status.
// An empty status means the task has no result yet.
func (act *actRunDebugPlugin) pollDebugTaskOnce(
	std *pluginUtils.PluginActionStandarder, taskID string, endpoint types.Endpoint, outputOffset *uint, pendingOutput *string) (
	*types.ScriptResult, error) {

	results, err := act.gseHandler.QueryScriptExecutionResult(
		std.Context(), taskID, &types.EndpointWithRestrict{
			Endpoint: endpoint,
			Offset:   *outputOffset,
			Limit:    0,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to query debug task result: %w", err)
	}
	if len(results) != 1 {
		return nil, fmt.Errorf("debug task result count is not 1, count(%d)", len(results))
	}

	result := results[0]
	*outputOffset += uint(len(result.ScreenLog))
	if err := act.storeDebugOutput(std, result.ScreenLog, isTerminalDebugTaskStatus(result.Status), pendingOutput); err != nil {
		return nil, err
	}

	return result, nil
}

func (act *actRunDebugPlugin) storeDebugOutput(std *pluginUtils.PluginActionStandarder, output string, flush bool, pending *string) error {
	bufferedOutput := *pending + output
	lines, remainder := splitDebugOutput(bufferedOutput, flush)

	messages := make([]common.Message, 0, len(lines))
	now := time.Now()
	for _, line := range lines {
		if line == "" {
			continue
		}
		messages = append(messages, common.Message{Time: now, TextZh: line, TextEn: line, Level: "INFO"})
	}
	if len(messages) == 0 {
		*pending = remainder
		return nil
	}

	if err := act.storageActionInstance.PushActionInstanceMessage(
		std.Context(), std.InstanceData().OperationInstanceID, ActionNameRunDebugPlugin, messages...); err != nil {
		*pending = bufferedOutput
		return err
	}

	*pending = remainder

	return nil
}

func splitDebugOutput(output string, flush bool) ([]string, string) {
	splitAt := strings.LastIndexByte(output, '\n')
	if flush && splitAt < len(output)-1 {
		splitAt = len(output)
	}
	if splitAt < 0 {
		return nil, output
	}

	lines := strings.Split(output[:splitAt], "\n")
	for index := range lines {
		lines[index] = strings.TrimSuffix(lines[index], "\r")
	}
	if flush {
		return lines, ""
	}

	return lines, output[splitAt+1:]
}

func isTerminalDebugTaskStatus(status types.ScriptStatus) bool {
	switch status {
	case types.ScriptStatusFinished, types.ScriptStatusFailed, types.ScriptStatusTimeout, types.ScriptStatusStopped, types.ScriptStatusAgentRestarted:
		return true
	default:
		return false
	}
}

func (act *actRunDebugPlugin) terminateDebugTask(std *pluginUtils.PluginActionStandarder, taskID string, endpoint types.Endpoint) error {
	terminateTaskID, err := act.gseHandler.TerminateScriptExecution(std.Context(), taskID, &endpoint)
	if err != nil {
		return fmt.Errorf("failed to terminate debug task: %w", err)
	}

	std.InstanceData().Log().
		Zh("成功提交终止插件调试任务, task-id(%s), terminate-task-id(%s)", taskID, terminateTaskID).
		En("succeeded to submit plugin debug task termination, task-id(%s), terminate-task-id(%s)", taskID, terminateTaskID).
		Info()

	return nil
}

// DisplayNameZh returns the Chinese display name.
func (act *actRunDebugPlugin) DisplayNameZh() string { return "运行插件调试" }

// DisplayNameEn returns the English display name.
func (act *actRunDebugPlugin) DisplayNameEn() string { return "Run Plugin Debug" }
