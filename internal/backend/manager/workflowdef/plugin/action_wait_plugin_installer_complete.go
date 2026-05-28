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
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameWaitPluginInstallerComplete defines the action name.
	ActionNameWaitPluginInstallerComplete = "wait_plugin_installer_complete"

	waitInstallerCompleteInterval = 1 * time.Second
)

// NewActionWaitPluginInstallerComplete get a new action.
func NewActionWaitPluginInstallerComplete(capability *Capability) action.Definition {
	return &actionWaitPluginInstallerComplete{
		daoPluginDeployment:   capability.StoragePlugin,
		gseHandler:            capability.GSEHandler,
		storageActionInstance: capability.StorageWorkflow,
	}
}

// ActionWaitPluginInstallerComplete defines the action param.
type ActionWaitPluginInstallerComplete struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

type actionWaitPluginInstallerComplete struct {
	daoPluginDeployment   pluginStg.IDaoPluginDeployment
	gseHandler            gse.IHandler
	storageActionInstance workflow.IStorageActionInstance
}

type offlinePluginInstallerStatus struct {
	OperInstID string                 `json:"oper_inst_id"`
	Status     installer.ProcessState `json:"status"`
	Error      string                 `json:"error"`
}

// Name returns the name of the action.
func (act *actionWaitPluginInstallerComplete) Name() string {
	return ActionNameWaitPluginInstallerComplete
}

// Version returns the version of the action.
func (act *actionWaitPluginInstallerComplete) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionWaitPluginInstallerComplete) Description() string {
	return "Wait for the installer command to complete"
}

// Timeout returns the timeout of the action.
func (act *actionWaitPluginInstallerComplete) Timeout() time.Duration {
	return 30 * time.Minute // nolint:mnd
}

// Tags returns the tags of the action.
func (act *actionWaitPluginInstallerComplete) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount this func define how many times this action will retry.
func (act *actionWaitPluginInstallerComplete) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionWaitPluginInstallerComplete) DelayFn() func() {
	return func() {}
}

// Do this func define what the action will do.
func (act *actionWaitPluginInstallerComplete) Do(ctx *action.InstanceContext) error {
	param := new(ActionWaitPluginInstallerComplete)
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

	pollingSwitch := conv.ToBoolDefault(
		ctx.Data.PrivateData[types.PDKeyActionWaitInstallerCompletePollingSwitch], false)

	// Check if this is offline mode by reading from deployment info.
	if pollingSwitch {
		std.InstanceData().Log().
			Zh("采用离线轮询模式, 等待插件安装完成").
			En("using offline polling mode to wait for plugin installation to complete").
			Info()

		return act.doOfflinePolling(std)
	}

	std.InstanceData().Log().
		Zh("采用回调模式，等待插件安装完成").
		En("using callback mode to wait for plugin installation to complete").
		Info()

	return act.doCallbackPolling(std)
}

func (act *actionWaitPluginInstallerComplete) doCallbackPolling(std *pluginUtils.PluginActionStandarder) error {
	instanceID := std.InstanceData().OperationInstanceID

	rawInstallerResult, err := act.waitInstallerField(std, types.PDKeyInstallerReportStatus)
	if err != nil {
		return fmt.Errorf("failed to wait for installer result: %w", err)
	}

	installerResult := installer.ProcessState(rawInstallerResult)

	return act.handleInstallerResult(std, instanceID, installerResult, "")
}

// doOfflinePolling waits for the offline installer to complete by polling the status file via GSE.
// Uses a single GSE ExecuteScript call with internal polling logic to minimize API calls.
// The script polls every second and returns when a terminal state is reached (success/failed/timeout).
func (act *actionWaitPluginInstallerComplete) doOfflinePolling(std *pluginUtils.PluginActionStandarder) error {
	instanceID := std.InstanceData().OperationInstanceID

	// Use one-time GSE script execution with internal polling loop
	status, err := act.waitOfflineInstallerStatus(std)
	if err != nil {
		return fmt.Errorf("failed to wait for offline installer result: %w", err)
	}

	return act.handleInstallerResult(std, instanceID, status.Status, status.Error)
}

// waitOfflineInstallerStatus executes a single GSE script that polls internally until status file is ready.
// nolint: gocognit
func (act *actionWaitPluginInstallerComplete) waitOfflineInstallerStatus(
	std *pluginUtils.PluginActionStandarder,
) (*offlinePluginInstallerStatus, error) {

	osType := std.DeployInfo().Process.Platform.OS
	if osType != criteria.OSLinux {
		return nil, fmt.Errorf("offline wait status script only supports linux, os(%s)", osType)
	}

	authEndpoint, restrictEndpoint, err := act.buildOfflineInstallerReadEndpoints(std)

	if err != nil {
		return nil, fmt.Errorf("failed to build offline installer read endpoints: %w", err)
	}
	// Build script with internal polling loop
	script := act.buildWaitStatusScript(std)

	// Timeout = action timeout + 30s buffer for script execution overhead
	timeout := act.Timeout() + 30*time.Second // nolint: mnd

	std.InstanceData().Log().
		Zh("开始等待离线安装器状态文件 (超时=%v)", timeout).
		En("waiting for offline installer status file (timeout=%v)", timeout).
		Info()

	taskID, err := act.gseHandler.ExecuteScript(
		std.Context(),
		types.ScriptTypeBash,
		script,
		timeout,
		authEndpoint,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute wait status script: %w", err)
	}

	// Poll script execution result until finished or timeout
	pollingInterval := 5 * time.Second // nolint: mnd
	pollingDeadline := time.Now().Add(timeout)

	for {
		results, err := act.gseHandler.QueryScriptExecutionResult(std.Context(), taskID, restrictEndpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to query wait status result: %w", err)
		}
		if len(results) != 1 {
			return nil, fmt.Errorf("wait status result count is not 1, count(%d)", len(results))
		}

		result := results[0]

		// Check if script finished
		if result.Status == types.ScriptStatusFinished {
			// Script completed, process the result
			break
		}

		// Check if script failed
		if result.Status == types.ScriptStatusFailed {
			return nil, fmt.Errorf(
				"wait status script failed, exit-code(%d), err-code(%d), err-msg(%s), screen(%s)",
				result.ExitCode, result.ErrorCode, result.ErrorMessage, result.ScreenLog,
			)
		}

		// Check timeout
		if time.Now().After(pollingDeadline) {
			return nil, fmt.Errorf(
				"wait status script polling timeout, status(%s), elapsed(%v)",
				result.Status, timeout,
			)
		}

		// Script still running, wait before next poll
		time.Sleep(pollingInterval)
	}

	// Get final result
	results, err := act.gseHandler.QueryScriptExecutionResult(std.Context(), taskID, restrictEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to query final wait status result: %w", err)
	}
	if len(results) != 1 {
		return nil, fmt.Errorf("final wait status result count is not 1, count(%d)", len(results))
	}

	result := results[0]

	rawContent := strings.TrimSpace(result.ScreenLog)
	if rawContent == "" {
		return nil, errors.New("wait status script returned empty content")
	}

	var installerStatus offlinePluginInstallerStatus
	if err := json.Unmarshal([]byte(rawContent), &installerStatus); err != nil {
		return nil, fmt.Errorf("failed to unmarshal installer status: %w", err)
	}

	instanceID := std.InstanceData().OperationInstanceID
	if installerStatus.OperInstID == "" {
		return nil, errors.New("installer status oper-inst-id is empty")
	}
	if installerStatus.OperInstID != instanceID {
		return nil, fmt.Errorf("installer status oper-inst-id mismatch, expect(%s), got(%s)",
			instanceID, installerStatus.OperInstID)
	}

	switch installerStatus.Status {
	case installer.ProcessStateSuccess, installer.ProcessStateFailed, installer.ProcessStateTimeout:
		std.InstanceData().Log().
			Zh("收到离线安装器状态: %s", installerStatus.Status).
			En("received offline installer status: %s", installerStatus.Status).
			Info()

		return &installerStatus, nil
	default:
		return nil, fmt.Errorf("installer status is not terminal, status(%s)", installerStatus.Status)
	}
}

func (act *actionWaitPluginInstallerComplete) buildWaitStatusScript(
	std *pluginUtils.PluginActionStandarder,
) string {

	statusFile := path.Join(
		path.Join(std.DeployInfo().InstallerRuntime.WorkDir, "data", "plugin", std.DeployInfo().Process.PluginName),
		installer.StatusFileName)

	// Timeout in seconds (use action timeout)
	timeoutSeconds := int(act.Timeout().Seconds())

	instanceID := std.InstanceData().OperationInstanceID

	// nolint: dupword
	return fmt.Sprintf(`#!/bin/sh
STATUS_FILE="%s"
OPER_INST_ID="%s"
TIMEOUT=%d

i=0
while [ "$i" -lt "$TIMEOUT" ]; do
    if [ -f "$STATUS_FILE" ]; then
        STATUS_CONTENT=$(cat "$STATUS_FILE" 2>/dev/null)
        if [ -n "$STATUS_CONTENT" ]; then
            # Extract oper_inst_id and status from JSON
            FILE_OPER_INST_ID=$(printf '%%s' "$STATUS_CONTENT" | sed -n 's/.*"oper_inst_id":"\([^"]*\)".*/\1/p')
            STATUS_STATE=$(printf '%%s' "$STATUS_CONTENT" | sed -n 's/.*"status":"\([^"]*\)".*/\1/p')
            
            # Check if oper_inst_id matches current instance
            if [ "$FILE_OPER_INST_ID" = "$OPER_INST_ID" ]; then
                # Check if status is terminal (success/failed/timeout)
                if [ "$STATUS_STATE" = "success" ] || [ "$STATUS_STATE" = "failed" ] || [ "$STATUS_STATE" = "timeout" ]; then
                    echo "$STATUS_CONTENT"
                    exit 0
                fi
            fi
        fi
    fi
    i=$((i + 1))
    sleep 1
done

# Timeout reached
echo '{"oper_inst_id":"'"$OPER_INST_ID"'","status":"timeout","error":"wait status file timeout after '"$TIMEOUT"' seconds"}'
exit 1
`, statusFile, instanceID, timeoutSeconds)
}

func (act *actionWaitPluginInstallerComplete) handleInstallerResult(
	std *pluginUtils.PluginActionStandarder,
	instanceID string,
	installerResult installer.ProcessState,
	installerError string,
) error {

	switch installerResult {
	case installer.ProcessStateSuccess:
		std.InstanceData().Log().
			Zh("收到安装器结果成功").
			En("received installer result is success").
			Info()

		return nil

	case installer.ProcessStateFailed, installer.ProcessStateTimeout:
		std.InstanceData().Log().
			Zh("收到安装器结果不成功, installer-result(%s)", installerResult).
			En("received installer result is not success. installer-result(%s)", installerResult).
			Info()

		if installerError != "" {
			return fmt.Errorf("installer failed. oper-inst-id(%s), action-name(%s), installer-result(%s), err(%s)",
				instanceID, ActionNameWaitPluginInstallerComplete, installerResult, installerError)
		}

		return fmt.Errorf("installer failed. oper-inst-id(%s), action-name(%s), installer-result(%s)",
			instanceID, ActionNameWaitPluginInstallerComplete, installerResult)

	default:
		logger.G.Sys().With("oper-inst-id", instanceID, "state", installerResult).Warn("installer state is not supported")

		return fmt.Errorf("unexpected installer state. state(%s)", installerResult)
	}
}

func (act *actionWaitPluginInstallerComplete) buildOfflineInstallerReadEndpoints(
	std *pluginUtils.PluginActionStandarder,
) (*types.EndpointWithAuth, *types.EndpointWithRestrict, error) {

	agentID := std.DeployInfo().Process.Info.AgentID
	if agentID == "" {
		return nil, nil, errors.New("plugin process agent id is empty")
	}

	return &types.EndpointWithAuth{
			Endpoint: types.Endpoint{AgentID: agentID},
			User:     "root",
		}, &types.EndpointWithRestrict{
			Endpoint: types.Endpoint{AgentID: agentID},
			Offset:   0,
			Limit:    0,
		}, nil
}

func (act *actionWaitPluginInstallerComplete) waitInstallerField(std *pluginUtils.PluginActionStandarder, key string) (string, error) {
	ticker := time.NewTicker(waitInstallerCompleteInterval)
	defer ticker.Stop()

	for {
		select {
		case <-std.Context().Done():

			return "", std.Context().Err()
		case <-ticker.C:
			value, err := act.tryFetchValue(std, key)
			if err != nil {
				return "", err
			}

			if value != "" {
				return value, nil
			}
		}
	}
}

func (act *actionWaitPluginInstallerComplete) tryFetchValue(std *pluginUtils.PluginActionStandarder, key string) (string, error) {
	privateData, err := act.storageActionInstance.GetActionInstancePrivateData(
		std.Context(),
		std.InstanceData().OperationInstanceID,
		ActionNameWaitPluginInstallerComplete)
	if err != nil {
		logger.G.Sys().Ctx(std.Context()).WithErr(err).Error("failed to get action private data")

		return "", err
	}

	rawValue, exists := privateData[key]
	if !exists {
		logger.G.Sys().Ctx(std.Context()).With("oper-inst-id", std.InstanceData().OperationInstanceID, "key", key).Debug("no receive data")

		return "", nil
	}

	value, err := conv.ToString(rawValue)
	if err != nil {
		logger.G.Sys().Ctx(std.Context()).With("oper-inst-id", std.InstanceData().OperationInstanceID, "key", key, "raw-value", rawValue).
			Error("unexpected type for fetched value")

		return "", fmt.Errorf("failed to get value for key. key(%s)", key)
	}

	return value, nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionWaitPluginInstallerComplete) DisplayNameZh() string {
	return "等待插件安装完成"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionWaitPluginInstallerComplete) DisplayNameEn() string {
	return "Wait Plugin Installer Complete"
}
