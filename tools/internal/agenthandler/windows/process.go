//go:build windows

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

package windows

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/winapi"
)

var (
	agentIDFromCMDRegex = regexp.MustCompile(`agent-id:\s+([A-Za-z0-9]+)`)
)

// GetProcess get current node process status.
func (handler *AgentHandler) GetProcess(_ context.Context) (*agenthandler.NodeProcess, error) {
	status, err := winapi.GetServiceStatus(handler.agentDaemonServiceName)
	if err != nil {
		return nil, fmt.Errorf("failed to check daemon service status: %w", err)
	}

	if status != winapi.WinSvcStatusRunning {
		return &agenthandler.NodeProcess{
			Dead: []string{gseAgentDaemonName, gseAgentBinName},
		}, nil
	}

	// TODO: double check if the gse_agent process is really running.
	return &agenthandler.NodeProcess{
		Running: []string{gseAgentDaemonName, gseAgentBinName},
	}, nil
}

// DiagnoseVersion runs the agent binary version diagnostic command.
func (handler *AgentHandler) DiagnoseVersion(ctx context.Context) (*agenthandler.AgentVersionDiagnostic, error) {
	workDir := handler.getAbsPath(handler.binDir)
	executable := handler.getAbsPath(handler.agentBinFilePath)
	args := []string{"-v"}
	diagnostic := &agenthandler.AgentVersionDiagnostic{
		WorkDir:    workDir,
		Executable: executable,
		Args:       args,
	}

	var stdout, stderr bytes.Buffer

	// nolint: gosec
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	diagnostic.Stdout = stdout.String()
	diagnostic.Stderr = stderr.String()
	if err != nil {
		return diagnostic, fmt.Errorf(
			"failed to diagnose agent version, %s: %w",
			diagnostic.LogString(),
			err,
		)
	}

	return diagnostic, nil
}

// ForceKill force kill the agent process.
// including file and data when it is proxy.
func (handler *AgentHandler) ForceKill(_ context.Context) error {
	return handler.forceKillProcess()
}

// GetAgentID get the current running agent-id.
func (handler *AgentHandler) GetAgentID(ctx context.Context) (string, error) {
	var stdout, stderr bytes.Buffer

	// nolint: gosec
	cmd := exec.CommandContext(ctx, handler.getAbsPath(handler.agentBinFilePath), "--agent-id")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = handler.getAbsPath(handler.binDir)

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get agent-id. stdout(%s), stderr(%s): %w",
			stdout.String(), stderr.String(), err)
	}

	return handler.extractAgentIDFromCMDOutput(stdout.String())
}

// RegisterAgentID register agent-id.
func (handler *AgentHandler) RegisterAgentID(ctx context.Context, existingAgentID string) (string, error) {
	var stdout, stderr bytes.Buffer

	// nolint: gosec
	cmd := exec.CommandContext(ctx,
		handler.getAbsPath(handler.agentBinFilePath),
		"-f", handler.getAbsPath(handler.agentConfigFilePath),
		"--register", existingAgentID)
	cmd.Dir = handler.getAbsPath(handler.binDir)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if !isAccessViolationErr(err) {
			return "", fmt.Errorf("failed to register agent-id. stdout(%s), stderr(%s): %w",
				stdout.String(), stderr.String(), err)
		}
	}

	return handler.extractAgentIDFromCMDOutput(stdout.String())
}

// UnregisterAgentID unregister agent-id.
func (handler *AgentHandler) UnregisterAgentID(ctx context.Context) error {
	var stdout, stderr bytes.Buffer

	// nolint: gosec
	cmd := exec.CommandContext(ctx,
		handler.getAbsPath(handler.agentBinFilePath),
		"-f", handler.getAbsPath(handler.agentConfigFilePath),
		"--unregister")
	cmd.Dir = handler.getAbsPath(handler.binDir)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to unregister agent-id. stdout(%s), stderr(%s): %w",
			stdout.String(), stderr.String(), err)
	}

	return nil
}

// InstallAutoStartup install auto startup after host boot.
func (handler *AgentHandler) InstallAutoStartup(ctx context.Context, opts *agenthandler.AsyncOutputOptions) error {
	status, err := winapi.GetServiceStatus(handler.agentDaemonServiceName)
	if err != nil {
		return fmt.Errorf("failed to get service status: %w", err)
	}

	if status != winapi.WinSvcStatusNotInstalled {
		return fmt.Errorf("service already registered. service(%s)", handler.agentDaemonServiceName)
	}

	if err := handler.installWinSvc(ctx, opts); err != nil {
		return err
	}

	return nil
}

// UninstallAutoStartup uninstall auto startup after host boot.
func (handler *AgentHandler) UninstallAutoStartup(ctx context.Context, opts *agenthandler.AsyncOutputOptions) error {
	status, err := winapi.GetServiceStatus(handler.agentDaemonServiceName)
	if err != nil {
		return fmt.Errorf("failed to get service status: %w", err)
	}

	if status == winapi.WinSvcStatusNotInstalled {
		return nil
	}

	if err = handler.uninstallWinSvc(ctx, opts); err != nil {
		return err
	}

	return nil
}

// Start start the agent process(including file and data in proxy mode).
func (handler *AgentHandler) Start(ctx context.Context, opts *agenthandler.AsyncOutputOptions) error {
	if err := handler.executeGSECtl(ctx, opts, "start"); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	return nil
}

// Stop stop the agent process.
func (handler *AgentHandler) Stop(ctx context.Context, _ bool, opts *agenthandler.AsyncOutputOptions) error {
	if err := handler.executeGSECtl(ctx, opts, "stop"); err != nil {
		return fmt.Errorf("failed to stop agent: %w", err)
	}

	return nil
}

// Restart restart the agent process.
func (handler *AgentHandler) Restart(
	ctx context.Context, _ bool, opts *agenthandler.AsyncOutputOptions) error {

	if err := handler.executeGSECtl(ctx, opts, "restart"); err != nil {
		return fmt.Errorf("failed to restart agent: %w", err)
	}

	return nil
}

// Reload reload the agent process.
func (handler *AgentHandler) Reload(ctx context.Context, opts *agenthandler.AsyncOutputOptions) error {
	if err := handler.executeGSECtl(ctx, opts, "restart"); err != nil {
		return fmt.Errorf("failed to reload agent: %w", err)
	}

	return nil
}

func (handler *AgentHandler) extractAgentIDFromCMDOutput(output string) (string, error) {
	result := agentIDFromCMDRegex.FindStringSubmatch(output)

	if len(result) != 2 { // nolint: mnd
		return "", fmt.Errorf("failed to get agent-id. output: %s", output)
	}

	return result[1], nil
}

// forceKillProcess force kill gse process.
// nolint: gocognit
func (handler *AgentHandler) forceKillProcess() error {
	// kill daemon.
	if err := winapi.KillProcessByNameAndPath(
		gseAgentDaemonName, handler.getAbsPath(handler.agentDaemonFilePath)); err != nil {
		return fmt.Errorf("failed to kill gse_agent_daemon: %w", err)
	}

	// kill agent.
	if err := winapi.KillProcessByNameAndPath(
		gseAgentBinName, handler.getAbsPath(handler.agentBinFilePath)); err != nil {
		return fmt.Errorf("failed to kill gse_agent: %w", err)
	}

	return nil
}

// installWinSvc install gse agent daemon service through daemon.
func (handler *AgentHandler) installWinSvc(ctx context.Context, opts *agenthandler.AsyncOutputOptions) error {
	if err := handler.executeDaemon(ctx, opts, "--install", "-f", handler.getAbsPath(handler.agentConfigFilePath), "--name", handler.agentDaemonServiceName); err != nil {
		return fmt.Errorf("failed to uninstall gse agent daemon service: %w", err)
	}

	return nil
}

// uninstallWinSvc uninstall gse agent daemon service through daemon.
func (handler *AgentHandler) uninstallWinSvc(ctx context.Context, opts *agenthandler.AsyncOutputOptions) error {
	if err := handler.executeDaemon(ctx, opts, "--uninstall", "--name", handler.agentDaemonServiceName); err != nil {
		return fmt.Errorf("failed to uninstall gse agent daemon service: %w", err)
	}

	return nil
}

func (handler *AgentHandler) executeDaemon(ctx context.Context, opts *agenthandler.AsyncOutputOptions, args ...string) error {
	// nolint: gosec
	cmd := exec.CommandContext(ctx, handler.getAbsPath(handler.agentDaemonFilePath), args...)
	cmd.Dir = handler.getAbsPath(handler.binDir)

	// sync execute.
	if opts == nil {
		var stdout, stderr bytes.Buffer

		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to execute gse daemon command. stdout(%s), stderr(%s): %w",
				stdout.String(), stderr.String(), err)
		}

		if stderr.String() != "" {
			return fmt.Errorf("failed to execute gse daemon command, stdout(%s), stderr(%s)",
				strings.ReplaceAll(stdout.String(), "\n", ""),
				strings.ReplaceAll(stderr.String(), "\n", ""))
		}

		return nil
	}

	// async execute with logging.
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	defer func() {
		_ = stdout.Close()
	}()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}
	defer func() {
		_ = stderr.Close()
	}()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start gsectl: %w", err)
	}

	gp := gopool.NewPool()
	gp.Go(func() error {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()

			if line != "" {
				opts.Stdout(line)
			}
		}

		return nil
	})
	gp.Go(func() error {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()

			if line != "" {
				opts.Stderr(line)
			}
		}

		return nil
	})

	err = cmd.Wait()
	_ = gp.Wait()

	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr

	if err != nil {
		return fmt.Errorf("failed to execute gse daemon command: %w", err)
	}

	if stdErr.String() != "" {
		return fmt.Errorf("failed to execute gse daemon command, stdout(%s), stderr(%s)",
			strings.ReplaceAll(stdOut.String(), "\n", ""),
			strings.ReplaceAll(stdErr.String(), "\n", ""))
	}

	return nil
}

func (handler *AgentHandler) executeGSECtl(
	ctx context.Context, opts *agenthandler.AsyncOutputOptions, args ...string) error {

	// nolint: gosec
	cmd := exec.CommandContext(ctx, handler.getAbsPath(handler.gseCtlFilePath), args...)
	cmd.Dir = handler.getAbsPath(handler.binDir)

	// sync execute.
	if opts == nil {
		var stdout, stderr bytes.Buffer

		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to execute gsectl. stdout(%s), stderr(%s): %w",
				stdout.String(), stderr.String(), err)
		}

		return nil
	}

	// async execute with logging.
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	defer func() {
		_ = stdout.Close()
	}()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}
	defer func() {
		_ = stderr.Close()
	}()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start gsectl: %w", err)
	}

	gp := gopool.NewPool()
	gp.Go(func() error {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()

			if line != "" {
				opts.Stdout(line)
			}
		}

		return nil
	})
	gp.Go(func() error {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()

			if line != "" {
				opts.Stderr(line)
			}
		}

		return nil
	})

	err = cmd.Wait()
	_ = gp.Wait()

	if err != nil {
		return fmt.Errorf("failed to wait gsectl: %w", err)
	}

	return nil
}

const accessViolationStatusCode = uint64(0xC0000005)

// isAccessViolationErr This is a compatibility issue, it will appear randomly in the Windows server 2019 version;
// since this error does not affect the main functionality at the moment, it is handled through retry;
// suspected: https://github.com/golang/go/issues/45153
func isAccessViolationErr(err error) bool {
	var exiterr *exec.ExitError
	if errors.As(err, &exiterr) {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return uint64(status.ExitStatus()) == accessViolationStatusCode
		}
	}

	return false
}
