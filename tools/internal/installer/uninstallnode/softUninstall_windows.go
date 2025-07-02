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

// softUninstall uninstall agent by soft way.
func (step *Step) softUninstall(ctx context.Context) error {
	logger.Infof(constant.StepUninstallNode, "stop agent, gse-ctl(%s)", step.gseCtlPath)
	if err := StopNode(ctx, step.gseCtlPath); err != nil {
		logger.Infof(constant.StepUninstallNode, "stop agent failed: %v", err)

		return fmt.Errorf("stop agent failed, err: %w", err)
	}
	logger.Infof(constant.StepUninstallNode, "successfully stop agent")

	logger.Infof(constant.StepUninstallNode, "remove gse agent self start task")
	if err := step.delGseAgentSelfStartTask(); err != nil {
		logger.Infof(constant.StepUninstallNode, "remove gse agent self start task failed: %v", err)

		return fmt.Errorf("remove gse agent self start task failed, err: %w", err)
	}
	logger.Infof(constant.StepUninstallNode, "successfully remove gse agent self start task")

	logger.Infof(constant.StepUninstallNode, "try to uninstall gse agent daemon service")
	if err := step.uninstallWinSvc(ctx, step.gseAgentDaemonSvcName); err != nil {
		logger.Infof(constant.StepUninstallNode, "uninstall gse agent daemon service failed: %v", err)

		return fmt.Errorf("uninstall gse agent daemon service failed, err: %w", err)
	}
	logger.Infof(constant.StepUninstallNode, "successfully uninstall gse agent daemon service")

	logger.Infof(constant.StepUninstallNode, "remove setup dir(%s)", step.setupDirPath)
	if err := RemoveSetupDir(ctx, step.setupDirPath); err != nil {

		logger.Infof(constant.StepUninstallNode, "remove setup dir failed: %v", err)

		return err
	}

	return nil
}

func (step *Step) uninstallWinSvc(
	ctx context.Context,
	gseAgentDaemonSvcName string,
) error {

	gseAgentDaemonFilePath := filepath.Join(step.binDirPath, gseAgentDaemonFileName)

	// stop service
	status, err := winapi.GetServiceStatus(gseAgentDaemonSvcName)
	if err != nil {
		return fmt.Errorf("get service status failed, err: %v", err)
	}

	if status == winapi.WinSvcStatusNotInstalled {
		logger.Infof(constant.StepUninstallNode, "service not installed, svcName(%s)", gseAgentDaemonSvcName)

		return nil
	}

	// Regardless of whether the service is stopped or not, we should stop it first
	err = step.stopGseAgentDaemonSvc(ctx, gseAgentDaemonFilePath, gseAgentDaemonSvcName)
	if err != nil {
		return fmt.Errorf("stop service failed, status(%s), svcName(%s)", status, gseAgentDaemonSvcName)
	}

	err = step.uninstallGseAgentDaemonSvc(ctx, gseAgentDaemonFilePath, gseAgentDaemonSvcName)
	if err != nil {
		return err
	}

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
		return fmt.Errorf("stop gse agent daemon service failed, err: %v", err)
	}

	if stdErr.String() != "" {
		return fmt.Errorf("stop gse agent daemon service failed, stdOut(%s), stdErr(%s)",
			strings.ReplaceAll(stdOut.String(), "\n", ""),
			strings.ReplaceAll(stdErr.String(), "\n", ""))
	}

	logger.Infof(constant.StepUninstallNode, "successfully stop gse agent daemon service, stdOut(%s)",
		strings.ReplaceAll(stdOut.String(), "\n", ""))

	return nil
}

// RegistryNameGseSelfStart is the name of the self-start task for GSE agent in the registry.
const RegistryNameGseSelfStart = "gse_agent"

// delGseAgentSelfStartTask deletes the self-start task of GSE agent from the registry.
func (step *Step) delGseAgentSelfStartTask() error {
	// delete registry
	err := winapi.DelRegistryCurrentUserStartRun(RegistryNameGseSelfStart)
	if err != nil {
		return fmt.Errorf("delete registry failed, err: %v", err)
	}

	return nil
}

func (step *Step) uninstallGseAgentDaemonSvc(ctx context.Context,
	gseAgentDaemonFilePath string,
	gseAgentDaemonSvcName string,
) error {

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
