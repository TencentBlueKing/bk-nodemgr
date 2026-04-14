//go:build windows

/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package windows provides the plugin process handler implementation for windows system.
package windows

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/pluginhandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
)

var _ pluginhandler.IPluginProcessHandler = &PluginHandler{}

// RunCmd run the command.
func (handler *PluginHandler) RunCmd(ctx context.Context, cmd string, pidDir string) error {
	args := strings.Fields(strings.TrimSpace(cmd))
	if len(args) == 0 {
		return fmt.Errorf("run command is empty")
	}

	setupAbsDir := handler.getAbsPath(handler.setupDir)
	if err := utils.CheckDirPathSafe(setupAbsDir); err != nil {
		return fmt.Errorf("invalid plugin setup dir(%s): %w", setupAbsDir, err)
	}

	execAbsPath, err := resolveExecutablePath(setupAbsDir, args[0])
	if err != nil {
		return err
	}

	outputCh, errCh, err := utils.RunCommandStreaming(ctx, setupAbsDir, execAbsPath, args[1:])
	if err != nil {
		return fmt.Errorf("failed to start command(%s): %w", cmd, err)
	}

	logger.Infof(plugin.StepDebugPlugin, "plugin command started, dir(%s), command(%s)", setupAbsDir, cmd)

	for line := range outputCh {
		if strings.TrimSpace(line) == "" {
			continue
		}
		logger.Infof(plugin.StepDebugPlugin, "[run-command-output] %s", line)
	}

	waitErr := <-errCh

	cleanupErr := handler.Stop(ctx, pidDir)
	if cleanupErr != nil {
		logger.Errorf(plugin.StepDebugPlugin, "failed to cleanup debug process by pid-dir(%s): %v", pidDir, cleanupErr)
	}

	if waitErr != nil {
		if cleanupErr != nil {
			return errors.Join(
				fmt.Errorf("command execution failed: %w", waitErr),
				fmt.Errorf("cleanup failed: %w", cleanupErr),
			)
		}

		return fmt.Errorf("command execution failed: %w", waitErr)
	}

	if cleanupErr != nil {
		return cleanupErr
	}

	return nil
}

// Stop stop plugin process handler.
func (handler *PluginHandler) Stop(_ context.Context, pidDir string) error {
	pidFiles, err := utils.CollectPIDFiles(handler.getAbsPath(""), pidDir)
	if err != nil {
		return err
	}

	if len(pidFiles) == 0 {
		logger.Infof(plugin.StepDebugPlugin, "no pid files found in pid-dir(%s), skip process cleanup", pidDir)
		return nil
	}

	setupAbsDir := handler.getAbsPath(handler.setupDir)
	if err := utils.CheckDirPathSafe(setupAbsDir); err != nil {
		return fmt.Errorf("invalid plugin setup dir(%s): %w", setupAbsDir, err)
	}

	var stopErr error

	for _, pidFile := range pidFiles {
		pid, err := utils.ReadPIDFromFile(pidFile)
		if err != nil {
			logger.Warnf(plugin.StepDebugPlugin, "skip invalid pid file(%s): %v", pidFile, err)

			continue
		}

		if err := utils.TerminateProcess(pid, setupAbsDir); err != nil {
			if errors.Is(err, utils.ErrProcessNotInDir) {
				logger.Infof(plugin.StepDebugPlugin, "skip pid(%d), process is not owned by plugin dir(%s)", pid, setupAbsDir)

				continue
			}

			stopErr = errors.Join(stopErr, fmt.Errorf("failed to stop pid(%d) from file(%s): %w", pid, pidFile, err))

			continue
		}

		logger.Infof(plugin.StepDebugPlugin, "stopped plugin process pid(%d), pid-file(%s)", pid, pidFile)
	}

	return stopErr
}

func resolveExecutablePath(pluginSetupAbsDir string, executableArg string) (string, error) {
	var executableAbsPath string
	if filepath.IsAbs(executableArg) {
		executableAbsPath = filepath.Clean(executableArg)
	} else {
		executableAbsPath = filepath.Clean(filepath.Join(pluginSetupAbsDir, executableArg))
	}

	if !utils.IsPathInDirectory(executableAbsPath, pluginSetupAbsDir) {
		return "", fmt.Errorf("executable is out of plugin setup dir. executable(%s), setup-dir(%s)", executableAbsPath, pluginSetupAbsDir)
	}

	fileInfo, err := os.Stat(executableAbsPath)
	if err != nil {
		return "", fmt.Errorf("failed to stat executable(%s): %w", executableAbsPath, err)
	}
	if fileInfo.IsDir() {
		return "", fmt.Errorf("executable is a directory(%s)", executableAbsPath)
	}

	return executableAbsPath, nil
}
