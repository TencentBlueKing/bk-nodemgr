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

package uninstallnode

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/winapi"
)

const (
	gseAgentDaemonName     = "gse_agent_daemon"
	gseAgentName           = "gse_agent"
	gseAgentDaemonFileName = "gse_agent_daemon.exe"
	gseAgentFileName       = "gse_agent.exe"
)

// forceUninstall uninstall agent by force way.
func (step *Step) forceUninstall(ctx context.Context) error {
	// make sure the service is uninstalled.
	err := step.forceUninstallGseAgentDaemonSvc(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (step *Step) forceUninstallGseAgentDaemonSvc(ctx context.Context) error {
	gseAgentDaemonSvcName := gseAgentDaemonName
	if step.deployEnv != "gse" {
		gseAgentDaemonSvcName = gseAgentDaemonSvcName + "_" + step.deployEnv
	}

	// delete registry
	err := winapi.DelRegistryCurrentUserStartRun(gseAgentName)
	if err != nil {
		return fmt.Errorf("delete registry failed, err: %v", err)
	}

	// stop service
	status, err := winapi.GetServiceStatus(gseAgentDaemonSvcName)
	if err != nil {
		return fmt.Errorf("get service status failed, err: %v", err)
	}

	if status == winapi.WinSvcStatusNotInstalled {
		logger.Infof(constant.StepUninstallNode, "service not installed, svcName(%s)", gseAgentDaemonSvcName)

		return nil
	}

	gseAgentDaemonFilePath := filepath.Join(step.binDirPath, gseAgentDaemonFileName)
	if status != winapi.WinSvcStatusStopped {
		err = step.stopGseAgentDaemonSvc(ctx, gseAgentDaemonFilePath, gseAgentDaemonSvcName)
		if err != nil {
			logger.Infof(constant.StepUninstallNode, "stop by service failed, err: %v", err)

			logger.Infof(constant.StepUninstallNode, "start to stop by force kill")

			gseAgentFilePath := filepath.Join(step.binDirPath, gseAgentFileName)

			err = winapi.KillProcessByNameAndPath(gseAgentDaemonFileName, gseAgentDaemonFilePath)
			if err != nil {
				return fmt.Errorf("kill gse_agent_daemon failed, err: %v", err)
			}

			err = winapi.KillProcessByNameAndPath(gseAgentFileName, gseAgentFilePath)
			if err != nil {
				return fmt.Errorf("kill gse_agent failed, err: %v", err)
			}
		}
	}

	cmd := exec.CommandContext(ctx,
		gseAgentDaemonFilePath,
		"--uninstall",
		"--name",
		gseAgentDaemonSvcName,
	)
	cmd.Dir = step.binDirPath
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr

	logger.Infof(constant.StepUninstallNode, "run command: %s", cmd.String())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("uninstall service failed, err: %v", err)
	}

	if stdErr.String() != "" {
		return fmt.Errorf("uninstall service failed, stdOut(%s), stdErr(%s)",
			strings.ReplaceAll(stdOut.String(), "\n", ""),
			strings.ReplaceAll(stdErr.String(), "\n", ""))
	}

	logger.Infof(constant.StepUninstallNode, "uninstall service success, stdOut(%s)",
		strings.ReplaceAll(stdOut.String(), "\n", ""))

	return nil
}

func (step *Step) stopGseAgentDaemonSvc(ctx context.Context, gseAgentDaemonPath string, gseAgentDaemonSvcName string) error {
	cmd := exec.CommandContext(ctx,
		gseAgentDaemonPath,
		"--quit",
		"--name",
		gseAgentDaemonSvcName,
	)
	cmd.Dir = step.binDirPath
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr

	logger.Infof(constant.StepUninstallNode, "run command: %s", cmd.String())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("stop service failed, err: %v", err)
	}

	if stdErr.String() != "" {
		return fmt.Errorf("stop service failed, stdOut(%s), stdErr(%s)",
			strings.ReplaceAll(stdOut.String(), "\n", ""),
			strings.ReplaceAll(stdErr.String(), "\n", ""))
	}

	logger.Infof(constant.StepUninstallNode, "stop service success, stdOut(%s)",
		strings.ReplaceAll(stdOut.String(), "\n", ""))

	return nil
}
