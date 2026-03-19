/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/credit"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/creditvault"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/sshx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameWaitInstallerComplete defines the action name.
	ActionNameWaitInstallerComplete = "wait_node_installer_complete"

	waitInstallerCompleteInterval = 1 * time.Second

	sshPollingInterval    = 5 * time.Second  // nolint: mnd
	sshPollingMaxBackoff  = 30 * time.Second // nolint: mnd
	sshPollingBaseBackoff = 2 * time.Second  // nolint: mnd
)

// NewActionWaitInstallerComplete get a new action.
func NewActionWaitInstallerComplete(capability *Capability) action.Definition {
	return &actionWaitInstallerComplete{
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		storageActionInstance: capability.StorageWorkflow,
		storageHostCredit:     capability.StorageHostCredit,
		passwordVault:         capability.HostPasswordVault,
	}
}

// ActionWaitInstallerComplete defines the action param.
type ActionWaitInstallerComplete struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
	EnsureAgentID                     bool `json:"ensure_agent_id"`
	UseSSHPolling                     bool `json:"use_ssh_polling"`
}

type actionWaitInstallerComplete struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	storageActionInstance workflow.IStorageActionInstance
	storageHostCredit     credit.IStorageHostCredit
	passwordVault         creditvault.IHostPasswordVault
}

// Name returns the name of the action.
func (act *actionWaitInstallerComplete) Name() string {
	return ActionNameWaitInstallerComplete
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionWaitInstallerComplete) DisplayNameZh() string {
	return "等待安装器完成"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionWaitInstallerComplete) DisplayNameEn() string {
	return "Wait for Installer Complete"
}

// Version returns the version of the action.
func (act *actionWaitInstallerComplete) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionWaitInstallerComplete) Description() string {
	return "Wait for the installer command to complete"
}

// Timeout returns the timeout of the action.
func (act *actionWaitInstallerComplete) Timeout() time.Duration {
	return 30 * time.Minute // nolint:mnd
}

// Tags returns the tags of the action.
func (act *actionWaitInstallerComplete) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount this func define how many times this action will retry.
func (act *actionWaitInstallerComplete) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionWaitInstallerComplete) DelayFn() func() {
	return func() {}
}

// Do this func define what the action will do.
// nolint: gocognit,funlen
func (act *actionWaitInstallerComplete) Do(ctx *action.InstanceContext) error {
	param := new(ActionWaitInstallerComplete)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment, act.storageHost)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	if param.UseSSHPolling {
		return act.doSSHPolling(std, param)
	}

	return act.doCallbackPolling(std, param)
}

// doCallbackPolling is the original callback-based polling logic.
func (act *actionWaitInstallerComplete) doCallbackPolling(std *nodeUtils.NodeActionStandarder, param *ActionWaitInstallerComplete) error {
	instanceID := std.InstanceData().OperationInstanceID

	rawInstallerResult, err := act.waitInstallerField(std, types.PDKeyInstallerReportStatus)
	if err != nil {
		return fmt.Errorf("failed to wait for installer result: %w", err)
	}

	installerResult := installer.ProcessState(rawInstallerResult)

	if param.EnsureAgentID && installerResult == installer.ProcessStateSuccess {
		agentID, err := act.waitInstallerField(std, types.PDKeyInstallerReportAgentID)
		if err != nil {
			return fmt.Errorf("failed to wait for agent id: %w", err)
		}

		std.DeployInfo().Host.Dynamic.AgentID = agentID
	}

	return act.handleInstallerResult(std, instanceID, installerResult)
}

// sshStatusFile represents the installer.status.json on the target machine.
// SYNC: must stay in sync with statusFileContent in tools/internal/installer/node/statusreporter/step.go.
type sshStatusFile struct {
	OperInstID string `json:"oper_inst_id"`
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
}

// sshDataFile represents the installer.data.json on the target machine.
// SYNC: must stay in sync with dataFileContent in tools/internal/installer/node/datareporter/step.go.
type sshDataFile struct {
	AgentID    string `json:"agent_id"`
	Token      string `json:"token"`
	OperInstID string `json:"oper_inst_id"`
}

// doSSHPolling implements the SSH polling branch for cross-unit proxy installations.
// nolint: gocognit,funlen,cyclop
func (act *actionWaitInstallerComplete) doSSHPolling(std *nodeUtils.NodeActionStandarder, param *ActionWaitInstallerComplete) error {
	instanceID := std.InstanceData().OperationInstanceID

	std.InstanceData().Log().
		Zh("使用 SSH 轮询模式等待安装器完成").
		En("using SSH polling mode to wait for installer completion").
		Info()

	creditHandler := nodeUtils.NewCreditHandler(act.storageHostCredit, act.passwordVault)
	cMethod, cKey, err := creditHandler.GetSSHCredit(std)
	if err != nil {
		return fmt.Errorf("failed to get ssh credit for polling: %w", err)
	}

	sshConfig := &sshx.Config{
		Network:    sshx.NetworkTCP,
		IP:         std.DeployInfo().Host.Dynamic.LoginIP,
		Port:       int(std.DeployInfo().Host.Dynamic.LoginPort),
		User:       std.DeployInfo().Host.Dynamic.LoginUser,
		AuthMethod: cMethod,
		Password: func() string {
			if cMethod == sshx.AuthMethodPassword {
				return cKey
			}

			return ""
		}(),
		PrivateKey: func() []byte {
			if cMethod == sshx.AuthMethodPrivateKey {
				return []byte(cKey)
			}

			return nil
		}(),
	}

	dataDir := path.Join(std.DeployInfo().InstallerWorkDir, "data")
	statusFilePath := path.Join(dataDir, installer.StatusFileName)
	dataFilePath := path.Join(dataDir, installer.DataFileName)
	logGlobPath := path.Join(dataDir, "logs", "installer_*.log")

	client, err := sshx.NewClient(std.Context(), sshConfig, sshx.DefaultTimeout)
	if err != nil {
		return fmt.Errorf("failed to create SSH client for polling: %w", err)
	}
	defer func() { _ = client.Close() }()

	ticker := time.NewTicker(sshPollingInterval)
	defer ticker.Stop()

	// NOTE on backoff strategy: We use a simple inline exponential backoff for SSH reconnect
	// delays rather than pkg/runtime/retrier.ExpoBackoff because the behavior differs:
	// ExpoBackoff.Do() retries a function N times then fails, whereas here we need an
	// indefinite polling loop where reconnect delays grow on consecutive failures but
	// reset on success. The reconnect backoff is only a wait-before-retry within the
	// outer ticker loop, not a standalone retry-until-exhaustion pattern.
	backoff := sshPollingBaseBackoff
	logLineOffset := 1

	for {
		select {
		case <-std.Context().Done():
			return std.Context().Err()

		case <-ticker.C:
			statusContent, _, connectErr := client.RunCommand(fmt.Sprintf("cat %s 2>/dev/null", statusFilePath))
			if connectErr != nil {
				logger.G.Sys().With("oper-inst-id", instanceID).WithErr(connectErr).
					Warn("SSH connection error during polling, attempting reconnect")

				_ = client.Close()
				newClient, reconnErr := act.reconnectSSH(std, sshConfig, backoff)
				if reconnErr != nil {
					logger.G.Sys().With("oper-inst-id", instanceID).WithErr(reconnErr).
						Warn("SSH reconnect failed, will retry on next tick")
					backoff = min(backoff*2, sshPollingMaxBackoff) // nolint: mnd

					continue
				}

				client = newClient
				backoff = sshPollingBaseBackoff

				continue
			}

			backoff = sshPollingBaseBackoff

			logLineOffset = act.tailInstallerLogs(client, std, logGlobPath, logLineOffset)

			statusContent = strings.TrimSpace(statusContent)
			if statusContent == "" {
				logger.G.Sys().With("oper-inst-id", instanceID).Debug("status file not yet available")

				continue
			}

			var status sshStatusFile
			if jsonErr := json.Unmarshal([]byte(statusContent), &status); jsonErr != nil {
				logger.G.Sys().With("oper-inst-id", instanceID).
					Warn("status file JSON parse failed (partial write), treating as in-progress")

				continue
			}

			// Guard against stale status files left by a previous installation run.
			// Only accept a status whose oper_inst_id matches the current operation.
			if status.OperInstID != instanceID {
				logger.G.Sys().With("oper-inst-id", instanceID, "file-oper-inst-id", status.OperInstID).
					Debug("status file belongs to a different operation, skipping")

				continue
			}

			installerResult := installer.ProcessState(status.Status)

			// Log error details before any data file reads — if the installer failed,
			// we want the error recorded regardless of EnsureAgentID or data file availability.
			if installerResult == installer.ProcessStateFailed {
				std.InstanceData().Log().
					Zh("安装器错误详情: %s", status.Error).
					En("installer error detail: %s", status.Error).
					Info()
			}

			// Always attempt to read the data file on success so that results
			// are captured. Only enforce AgentID presence when EnsureAgentID is set.
			if installerResult == installer.ProcessStateSuccess {
				dataContent, _, dataErr := client.RunCommand(fmt.Sprintf("cat %s 2>/dev/null", dataFilePath))
				if dataErr != nil {
					return fmt.Errorf("failed to read data file via SSH: %w", dataErr)
				}

				var data sshDataFile
				if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(dataContent)), &data); jsonErr != nil {
					return fmt.Errorf("failed to parse data file: %w", jsonErr)
				}

				if data.AgentID != "" {
					std.DeployInfo().Host.Dynamic.AgentID = data.AgentID
				} else if param.EnsureAgentID {
					return fmt.Errorf("installer succeeded but agent_id is empty in data file")
				}
			}

			return act.handleInstallerResult(std, instanceID, installerResult)
		}
	}
}

// tailInstallerLogs reads new lines from the installer log file on the target machine
// via SSH and pushes them as action instance messages for frontend visibility.
// Returns the updated line offset for the next call.
func (act *actionWaitInstallerComplete) tailInstallerLogs(
	client *sshx.Client,
	std *nodeUtils.NodeActionStandarder,
	logGlobPath string,
	offset int,
) int {
	// Find the newest installer log file and read lines starting from offset.
	// The subshell resolves the glob; if no file exists, tail receives an empty
	// argument and silently produces no output.
	tailCmd := fmt.Sprintf(
		`tail -n +%d "$(ls -1t %s 2>/dev/null | head -1)" 2>/dev/null`,
		offset, logGlobPath,
	)

	output, _, err := client.RunCommand(tailCmd)
	if err != nil || strings.TrimSpace(output) == "" {
		return offset
	}

	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	for _, line := range lines {
		if msg := parseInstallerLogLine(line); msg != "" {
			std.InstanceData().Log().Zh(msg).En(msg).Info()
		}
	}

	return offset + len(lines)
}

// parseInstallerLogLine extracts a human-readable message from a raw installer log line.
// Expected format: "YYYY/MM/DD HH:MM:SS | LEVEL | step_name | message".
func parseInstallerLogLine(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}

	parts := strings.SplitN(line, installer.LogFieldSeparator, installer.LogFieldCount)
	if len(parts) < installer.LogFieldCount {
		return line
	}

	step := strings.TrimSpace(parts[2])
	msg := strings.TrimSpace(parts[3])

	if step != "" {
		return fmt.Sprintf("[%s] %s", step, msg)
	}

	return msg
}

// reconnectSSH attempts to re-establish an SSH connection with exponential backoff delay.
func (act *actionWaitInstallerComplete) reconnectSSH(
	std *nodeUtils.NodeActionStandarder,
	config *sshx.Config,
	backoff time.Duration,
) (*sshx.Client, error) {

	logger.G.Sys().With("backoff", backoff.String()).Info("waiting before SSH reconnect")

	select {
	case <-std.Context().Done():
		return nil, std.Context().Err()
	case <-time.After(backoff):
	}

	client, err := sshx.NewClient(std.Context(), config, sshx.DefaultTimeout)
	if err != nil {
		return nil, fmt.Errorf("SSH reconnect failed: %w", err)
	}

	logger.G.Sys().Info("SSH reconnected successfully")

	return client, nil
}

// handleInstallerResult processes the final installer result for both polling modes.
func (act *actionWaitInstallerComplete) handleInstallerResult(
	std *nodeUtils.NodeActionStandarder,
	instanceID string,
	result installer.ProcessState,
) error {

	switch result {
	case installer.ProcessStateSuccess:
		std.ResetInstanceDataContext()
		std.InstanceData().Log().
			Zh("接收到安装器结果为成功").
			En("received installer result is success").
			Info()

		return nil

	case installer.ProcessStateFailed, installer.ProcessStateTimeout:
		std.InstanceData().Log().
			Zh("接收到安装器结果不成功。installer-result(%s)", result).
			En("received installer result is not success. installer-result(%s)", result).
			Info()

		return fmt.Errorf("installer failed. oper-inst-id(%s), action-name(%s), installer-result(%s)",
			instanceID, ActionNameWaitInstallerComplete, result)

	default:
		logger.G.Sys().With("oper-inst-id", instanceID, "state", result).Warn("installer state is not supported")

		return fmt.Errorf("unexpected installer state. state(%s)", result)
	}
}

func (act *actionWaitInstallerComplete) waitInstallerField(std *nodeUtils.NodeActionStandarder, key string) (string, error) {
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

func (act *actionWaitInstallerComplete) tryFetchValue(std *nodeUtils.NodeActionStandarder, key string) (string, error) {
	privateData, err := act.storageActionInstance.GetActionInstancePrivateData(
		std.Context(),
		std.InstanceData().OperationInstanceID,
		ActionNameWaitInstallerComplete)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to get action private data")

		return "", err
	}

	rawValue, exists := privateData[key]
	if !exists {
		logger.G.Sys().With("oper-inst-id", std.InstanceData().OperationInstanceID, "key", key).Debug("no receive data")

		return "", nil
	}

	value, err := conv.ToString(rawValue)
	if err != nil {
		logger.G.Sys().With("oper-inst-id", std.InstanceData().OperationInstanceID, "key", key, "raw-value", rawValue).
			Error("unexpected type for fetched value")

		return "", fmt.Errorf("failed to get value for key. key(%s)", key)
	}

	return value, nil
}

