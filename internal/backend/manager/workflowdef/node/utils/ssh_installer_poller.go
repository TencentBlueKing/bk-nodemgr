/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/sshx"
)

const (
	sshPollingInterval    = 5 * time.Second  // nolint: mnd
	sshPollingMaxBackoff  = 30 * time.Second // nolint: mnd
	sshPollingBaseBackoff = 2 * time.Second  // nolint: mnd

	// sshCommandTimeout bounds how long a single SSH RunCommand may block.
	// If the SSH connection enters a half-open state (TCP alive but remote
	// unresponsive), session.Run() blocks forever and prevents both context
	// cancellation and the polling loop from progressing. This timeout
	// ensures we detect the stall and trigger a reconnect.
	sshCommandTimeout = 30 * time.Second // nolint: mnd
)

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

// SSHInstallerPollerConfig holds all inputs required for SSH-based installer polling.
type SSHInstallerPollerConfig struct {
	// SSHConfig is the SSH connection configuration for the target host.
	SSHConfig *sshx.Config

	// StatusFile is the absolute path of the installer status JSON file on the remote host.
	StatusFile string

	// DataFile is the absolute path of the installer data JSON file on the remote host.
	DataFile string

	// LogGlobPath is the glob pattern used to locate installer log files on the remote host.
	LogGlobPath string

	// InstanceID is the current operation instance ID used to guard against stale status files.
	InstanceID string

	// EnsureAgentID when true causes Wait to return an error if the installer reports success
	// but the data file contains no agent_id.
	EnsureAgentID bool
}

// SSHInstallerPollerResult holds the outcome returned by SSHInstallerPoller.Wait.
type SSHInstallerPollerResult struct {
	// State is the final installer process state.
	State installer.ProcessState

	// AgentID is the agent ID reported by the installer data file, if any.
	AgentID string
}

// SSHInstallerPoller polls installer status and log files over an SSH connection.
// It handles connection establishment, transparent reconnection with exponential backoff,
// log tailing, and status file parsing, returning a clean result to the caller.
type SSHInstallerPoller struct{}

// NewSSHInstallerPoller creates a new SSHInstallerPoller.
func NewSSHInstallerPoller() *SSHInstallerPoller {
	return &SSHInstallerPoller{}
}

// Wait blocks until the installer completes, the context is cancelled, or a fatal error occurs.
// It logs installer output to the action instance log via std so that progress is visible in the frontend.
//
// NOTE on backoff strategy: We use a simple inline exponential backoff for SSH reconnect
// delays rather than pkg/runtime/retrier.ExpoBackoff because the behavior differs:
// ExpoBackoff.Do() retries a function N times then fails, whereas here we need an
// indefinite polling loop where reconnect delays grow on consecutive failures but
// reset on success. The reconnect backoff is only a wait-before-retry within the
// outer ticker loop, not a standalone retry-until-exhaustion pattern.
// nolint: gocognit,funlen,cyclop
func (p *SSHInstallerPoller) Wait(std *NodeActionStandarder, cfg SSHInstallerPollerConfig) (SSHInstallerPollerResult, error) {
	std.InstanceData().Log().
		Zh("使用 SSH 轮询模式等待安装器完成").
		En("using SSH polling mode to wait for installer completion").
		Info()

	client, err := sshx.NewClient(std.Context(), cfg.SSHConfig, sshx.DefaultTimeout)
	if err != nil {
		return SSHInstallerPollerResult{}, fmt.Errorf("failed to create SSH client for polling: %w", err)
	}
	defer func() { _ = client.Close() }()

	ticker := time.NewTicker(sshPollingInterval)
	defer ticker.Stop()

	backoff := sshPollingBaseBackoff
	logLineOffset := 1

	for {
		select {
		case <-std.Context().Done():
			return SSHInstallerPollerResult{}, std.Context().Err()

		case <-ticker.C:
			statusContent, connectErr := runCommandWithContext(
				std.Context(), client,
				fmt.Sprintf("cat %s 2>/dev/null", cfg.StatusFile),
				sshCommandTimeout,
			)
			if connectErr != nil {
				logger.G.Sys().Ctx(std.Context()).With("oper-inst-id", cfg.InstanceID).WithErr(connectErr).
					Warn("SSH connection error during polling, attempting reconnect")

				client, backoff = p.attemptSSHReconnect(std, cfg.InstanceID, client, cfg.SSHConfig, backoff)

				continue
			}

			backoff = sshPollingBaseBackoff

			// Tail logs before parsing status: if the tail fails (connection died mid-tick),
			// we still fall through to process the status we already read — discarding it
			// would risk missing a terminal result (success/failed) if the context is about
			// to expire. On tail failure we do reconnect after parsing status.
			var tailErr error
			logLineOffset, tailErr = p.tailInstallerLogs(std.Context(), client, std, cfg.LogGlobPath, logLineOffset)
			if tailErr != nil {
				logger.G.Sys().Ctx(std.Context()).With("oper-inst-id", cfg.InstanceID).WithErr(tailErr).
					Warn("SSH error while tailing logs, will reconnect after processing status")
			}

			statusContent = strings.TrimSpace(statusContent)
			if statusContent == "" {
				logger.G.Sys().Ctx(std.Context()).With("oper-inst-id", cfg.InstanceID).Debug("status file not yet available")

				if tailErr != nil {
					client, backoff = p.attemptSSHReconnect(std, cfg.InstanceID, client, cfg.SSHConfig, backoff)
				}

				continue
			}

			var status sshStatusFile
			if jsonErr := json.Unmarshal([]byte(statusContent), &status); jsonErr != nil {
				logger.G.Sys().Ctx(std.Context()).With("oper-inst-id", cfg.InstanceID).
					Warn("status file JSON parse failed (partial write), treating as in-progress")

				if tailErr != nil {
					client, backoff = p.attemptSSHReconnect(std, cfg.InstanceID, client, cfg.SSHConfig, backoff)
				}

				continue
			}

			// Guard against stale status files left by a previous installation run.
			// Only accept a status whose oper_inst_id matches the current operation.
			if status.OperInstID != cfg.InstanceID {
				logger.G.Sys().Ctx(std.Context()).With("oper-inst-id", cfg.InstanceID, "file-oper-inst-id", status.OperInstID).
					Debug("status file belongs to a different operation, skipping")

				if tailErr != nil {
					client, backoff = p.attemptSSHReconnect(std, cfg.InstanceID, client, cfg.SSHConfig, backoff)
				}

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
				// If the log tail failed, the client was closed by runCommandWithContext.
				// Reconnect before reading the data file so we don't misreport a successful
				// installation as a data-read failure.
				if tailErr != nil {
					client, _ = p.attemptSSHReconnect(std, cfg.InstanceID, client, cfg.SSHConfig, sshPollingBaseBackoff)
				}

				dataContent, dataErr := runCommandWithContext(
					std.Context(), client,
					fmt.Sprintf("cat %s 2>/dev/null", cfg.DataFile),
					sshCommandTimeout,
				)
				if dataErr != nil {
					return SSHInstallerPollerResult{}, fmt.Errorf("failed to read data file via SSH: %w", dataErr)
				}

				var data sshDataFile
				if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(dataContent)), &data); jsonErr != nil {
					return SSHInstallerPollerResult{}, fmt.Errorf("failed to parse data file: %w", jsonErr)
				}

				if data.AgentID == "" && cfg.EnsureAgentID {
					return SSHInstallerPollerResult{}, fmt.Errorf("installer succeeded but agent_id is empty in data file")
				}

				return SSHInstallerPollerResult{State: installerResult, AgentID: data.AgentID}, nil
			}

			// Non-success terminal state (failed/timeout): no further SSH reads needed.
			return SSHInstallerPollerResult{State: installerResult}, nil
		}
	}
}

// tailInstallerLogs reads new lines from the installer log file on the target machine
// via SSH and pushes them as action instance messages for frontend visibility.
// Returns the updated line offset and an error if the SSH command failed (which
// means the client has been closed by runCommandWithContext and must be reconnected).
func (p *SSHInstallerPoller) tailInstallerLogs(
	ctx context.Context,
	client *sshx.Client,
	std *NodeActionStandarder,
	logGlobPath string,
	offset int,
) (int, error) {
	// Find the newest installer log file and read lines starting from offset.
	// The subshell resolves the glob; if no file exists, tail receives an empty
	// argument and silently produces no output.
	tailCmd := fmt.Sprintf(
		`tail -n +%d "$(ls -1t %s 2>/dev/null | head -1)" 2>/dev/null`,
		offset, logGlobPath,
	)

	output, err := runCommandWithContext(ctx, client, tailCmd, sshCommandTimeout)
	if err != nil {
		return offset, fmt.Errorf("failed to tail installer logs: %w", err)
	}

	if strings.TrimSpace(output) == "" {
		return offset, nil
	}

	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	for _, line := range lines {
		if msg := parseInstallerLogLine(line); msg != "" {
			std.InstanceData().Log().Zh(msg).En(msg).Info()
		}
	}

	return offset + len(lines), nil
}

// attemptSSHReconnect closes the dead client and re-establishes a fresh SSH connection.
// The client may already have been closed by runCommandWithContext; calling Close again is
// safe on *ssh.Client (double-close returns an ignorable error).
// On reconnect failure the original (closed) client is returned unchanged so the next polling
// tick will naturally fail again and the exponential backoff continues to grow.
func (p *SSHInstallerPoller) attemptSSHReconnect(
	std *NodeActionStandarder,
	instanceID string,
	client *sshx.Client,
	config *sshx.Config,
	backoff time.Duration,
) (*sshx.Client, time.Duration) {

	_ = client.Close()

	logger.G.Sys().Ctx(std.Context()).With("backoff", backoff.String()).Info("waiting before SSH reconnect")

	// Use NewTimer + Stop to avoid the timer-leak that time.After causes when ctx.Done()
	// fires before the backoff elapses (time.After timers are not GC'd until they fire).
	backoffTimer := time.NewTimer(backoff)
	defer backoffTimer.Stop()

	select {
	case <-std.Context().Done():
		// Context cancelled during backoff wait; outer loop will handle it on the next tick.
		return client, min(backoff*2, sshPollingMaxBackoff) // nolint: mnd
	case <-backoffTimer.C:
	}

	newClient, err := sshx.NewClient(std.Context(), config, sshx.DefaultTimeout)
	if err != nil {
		logger.G.Sys().Ctx(std.Context()).With("oper-inst-id", instanceID).WithErr(err).
			Warn("SSH reconnect failed, will retry on next tick")

		return client, min(backoff*2, sshPollingMaxBackoff) // nolint: mnd
	}

	logger.G.Sys().Ctx(std.Context()).Info("SSH reconnected successfully")

	return newClient, sshPollingBaseBackoff
}

// sshCommandResult holds the stdout of a single RunCommand call.
type sshCommandResult struct {
	stdout string
	err    error
}

// runCommandWithContext runs an SSH command with context awareness and a hard timeout.
//
// sshx.Client.RunCommand (session.Run) does not accept a context and will block
// indefinitely when the SSH connection enters a half-open state. This wrapper
// runs the command in a goroutine and returns early when the context is cancelled
// or the timeout fires, closing the underlying client to unblock session.Run.
//
// IMPORTANT: on timeout / context cancellation the caller MUST NOT reuse the
// client — it has been forcibly closed. The returned error signals the caller to
// enter the reconnect path.
func runCommandWithContext(ctx context.Context, client *sshx.Client, cmd string, timeout time.Duration) (string, error) {
	ch := make(chan sshCommandResult, 1)

	go func() {
		stdout, _, err := client.RunCommand(cmd)
		ch <- sshCommandResult{stdout, err}
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case res := <-ch:
		return res.stdout, res.err

	case <-ctx.Done():
		// Force-close the SSH connection so the blocked session.Run returns.
		_ = client.Close()

		return "", fmt.Errorf("context cancelled while running SSH command: %w", ctx.Err())

	case <-timer.C:
		// Force-close the SSH connection so the blocked session.Run returns.
		_ = client.Close()

		return "", fmt.Errorf("SSH command timed out after %v: %s", timeout, cmd)
	}
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
