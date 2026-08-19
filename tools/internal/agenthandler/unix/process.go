//go:build linux || darwin || freebsd || aix

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

package unix

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
)

var (
	agentIDFromCMDRegex = regexp.MustCompile(`agent-id:\s+([A-Za-z0-9]+)`)
)

// GetProcess get current node process status.
// including file and data when it is proxy.
func (handler *AgentHandler) GetProcess(_ context.Context) (*agenthandler.NodeProcess, error) {
	pidFiles := map[string]string{"gse_agent": filepath.Join(handler.binDir, "run", "agent.pid")}

	if handler.role == types.NodeRoleProxy {
		pidFiles["gse_file"] = filepath.Join(handler.binDir, "run", "file.pid")
		pidFiles["gse_data"] = filepath.Join(handler.binDir, "run", "data.pid")
	}

	result := agenthandler.NewNodeProcess()
	for processName, pidFile := range pidFiles {
		exist, err := handler.checkProcessWithPidFile(pidFile)
		if err != nil {
			// if check pid file failed, then use process name to check.
			if exist, err = handler.checkProcessWithProcessName(processName); err != nil {
				return nil, err
			}
		}

		if exist {
			result.Running = append(result.Running, processName)

			continue
		}
		result.Dead = append(result.Dead, processName)
	}

	return result, nil
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
	switch handler.role {
	case types.NodeRoleAgent:
		return handler.forceKillProcess(gseAgentBinName)
	case types.NodeRoleProxy:
		return handler.forceKillProcess(gseAgentBinName, gseFileBinName, gseDataBinName)
	default:
		return fmt.Errorf("invalid role: %s", handler.role)
	}
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
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to register agent-id. stdout(%s), stderr(%s): %w",
			stdout.String(), stderr.String(), err)
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
func (handler *AgentHandler) InstallAutoStartup(_ context.Context, opts *agenthandler.AsyncOutputOptions) error {
	if opts != nil {
		opts.Stdout("unix auto starup install will be ensure in start command")
	}

	return nil
}

// UninstallAutoStartup uninstall auto startup after host boot.
func (handler *AgentHandler) UninstallAutoStartup(_ context.Context, opts *agenthandler.AsyncOutputOptions) error {
	if opts != nil {
		opts.Stdout("unix auto starup uninstall will be ensure in stop command")
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
	if err := handler.executeGSECtl(ctx, opts, "reload"); err != nil {
		return fmt.Errorf("failed to reload agent: %w", err)
	}

	return nil
}

func (handler *AgentHandler) executeGSECtl(
	ctx context.Context, opts *agenthandler.AsyncOutputOptions, args ...string) error {

	// nolint: gosec
	cmd := exec.CommandContext(ctx, handler.getAbsPath(handler.gseCtlFilePath), args...)

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

func (handler *AgentHandler) extractAgentIDFromCMDOutput(output string) (string, error) {
	result := agentIDFromCMDRegex.FindStringSubmatch(output)

	if len(result) != 2 { // nolint: mnd
		return "", fmt.Errorf("failed to get agent-id. output: %s", output)
	}

	return result[1], nil
}

func (handler *AgentHandler) checkProcessWithPidFile(pidFileRelativePath string) (bool, error) {
	pidFile, err := handler.openFileForRead(pidFileRelativePath)
	if err != nil {
		return false, fmt.Errorf("failed to open pid file(%s): %w", pidFileRelativePath, err)
	}
	defer func() {
		_ = pidFile.Close()
	}()

	content, err := io.ReadAll(pidFile)
	if err != nil {
		return false, fmt.Errorf("failed to read pid file(%s): %w", pidFileRelativePath, err)
	}

	pid, err := strconv.Atoi(strings.Trim(string(content), " \n\r\t"))
	if err != nil {
		return false, fmt.Errorf("failed to convert pid to int. pid(%s), file(%s)", string(content), pidFileRelativePath)
	}

	process, err := utils.GetSameSpaceProcesses()
	if err != nil {
		return false, fmt.Errorf("failed to get same space processes: %w", err)
	}

	for _, p := range process {
		if p.PID == pid {
			if !isGseBin(p.Name) {
				return false, nil
			}

			break
		}
	}

	exist, err := utils.CheckPIDExist(pid)
	if err != nil {
		return false, fmt.Errorf("failed to check pid exist. pid(%d): %w", pid, err)
	}

	return exist, nil
}

func (handler *AgentHandler) checkProcessWithProcessName(processName string) (bool, error) {
	process, err := utils.GetSameSpaceProcesses()
	if err != nil {
		return false, fmt.Errorf("failed to get same space processes: %w", err)
	}

	binDir := handler.binDir
	if err := utils.CheckDirPathSafe(binDir); err != nil {
		return false, fmt.Errorf("will not force kill process within unsafe bin-dir(%s)", binDir)
	}

	for _, p := range process {
		if p.Name != processName {
			continue
		}

		if !isGseBin(p.Name) {
			continue
		}

		if !strings.HasPrefix(p.FullPath, binDir) {
			continue
		}

		exist, err := utils.CheckPIDExist(p.PID)
		if err != nil {
			return false, fmt.Errorf("failed to check pid exist. pid(%d): %w", p.PID, err)
		}

		if exist {
			return true, nil
		}
	}

	return false, nil
}

// forceKillProcess force kill gse process.
// nolint: gocognit
func (handler *AgentHandler) forceKillProcess(processName ...string) error {
	process, err := utils.GetSameSpaceProcesses()
	if err != nil {
		return fmt.Errorf("failed to get same space processes: %w", err)
	}

	binDir := handler.getAbsPath(handler.binDir)
	if err := utils.CheckDirPathSafe(binDir); err != nil {
		return fmt.Errorf("will not force kill process within unsafe bin-dir(%s)", binDir)
	}

	for _, p := range process {
		matched := false
		for _, pn := range processName {
			if pn == p.Name {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}

		if !isGseBin(p.Name) {
			continue
		}

		if !strings.HasPrefix(p.FullPath, binDir) {
			continue
		}

		exist, err := utils.CheckPIDExist(p.PID)
		if err != nil {
			return fmt.Errorf("failed to force kill, pid: %d: %w", p.PID, err)
		}

		if !exist {
			continue
		}

		// check if the pid again, make sure the pid is not a system pid.
		if p.PID <= 1 {
			return fmt.Errorf("process pid is too dangerous. pid(%d)", p.PID)
		}

		err = syscall.Kill(p.PID, syscall.SIGKILL)
		if err != nil {
			return fmt.Errorf("failed to force kill. pid(%d): %w", p.PID, err)
		}
	}

	return nil
}
