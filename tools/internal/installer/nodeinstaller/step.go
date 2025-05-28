/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package nodeinstaller this package is used to install gse agent/proxy.
package nodeinstaller

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/retrier"
)

// Step install agent.
type Step struct {
	agentID           string
	reRegisterAgentID bool
	setupDirPath      string
	pkgPath           string
	srcConfigDir      string
	dstConfigDir      string
	agentPath         string
	agentConfigPath   string
	agentCtlPath      string
	overwrite         bool
}

// StepArgs args for step.
type StepArgs struct {
	AgentID           string
	ReRegisterAgentID bool
	SetupDirPath      string
	PkgPath           string
	SrcConfigDir      string
	Overwrite         bool
}

// NewStep new a step.
func NewStep(args StepArgs) *Step {
	step := &Step{
		agentID:           args.AgentID,
		reRegisterAgentID: args.ReRegisterAgentID,
		setupDirPath:      args.SetupDirPath,
		pkgPath:           args.PkgPath,
		srcConfigDir:      args.SrcConfigDir,
		dstConfigDir:      filepath.Join(args.SetupDirPath, "etc"),
		overwrite:         args.Overwrite,
	}

	step.agentPath = filepath.Join(step.setupDirPath, "bin", GetGseAgentName())
	step.agentCtlPath = filepath.Join(step.setupDirPath, "bin", GetGseAgentCtlName())
	step.agentConfigPath = filepath.Join(step.dstConfigDir, "gse_agent.conf")

	return step
}

// GetGseAgentName get gse agent name.
const baseNameAgent = "gse_agent"

// GetGseAgentName get gse agent name.
func GetGseAgentName() string {
	osType := runtime.GOOS
	switch osType {
	case "windows":
		return baseNameAgent + ".exe"
	default:
		return baseNameAgent
	}
}

const baseNameAgentCtl = "gsectl"

// GetGseAgentCtlName gse agent ctl name.
func GetGseAgentCtlName() string {
	osType := runtime.GOOS
	switch osType {
	case "windows":
		return baseNameAgentCtl + ".bat"
	default:
		return baseNameAgentCtl
	}
}

// Run run the step to install node.
// 1. prepare
// 1.1 try to create install dir.
// 1.2 clean config dir.
// 2. cp files to install dir
// 2.1 unzip gse pkg to setup dir
// 2.2 copy config files to setup dir
// 3. get agent id
// 3.1 unregister agent id [optional]
// 3.2 register agent id
// 4. check gse node health
// 5. start gse node.
func (step *Step) Run(ctx context.Context) (string, error) {
	logger.Info(constant.StepInstallNode, constant.StateStart, "start install node")

	logger.Infof(constant.StepInstallNode, constant.StateRunning, "setup-dir-path(%s)", step.setupDirPath)

	// 1. prepare
	// 1.1 try to create install dir.
	err := TryCreateInstallDir(ctx, step.setupDirPath, step.overwrite)
	if err != nil {
		logger.Error(constant.StepInstallNode, constant.StateFailed,
			fmt.Sprintf("create install dir failed: %v", err))

		return "", err
	}
	logger.Info(constant.StepInstallNode, constant.StateRunning, "successfully create install dir")

	// 2. cp files to install dir
	// 2.1 copy gse pkg files to setup dir

	err = UnzipPkgToSetupDir(ctx, step.pkgPath, step.setupDirPath)
	if err != nil {
		return "", err
	}

	// 2.2 make gse node and some other files executable.
	err = MakeGseBinFileExecutable(ctx, filepath.Join(step.setupDirPath, "bin"))
	if err != nil {
		return "", err
	}

	// 2.3 copy config files to setup dir
	err = CopyConfigFilesToSetupDir(step.srcConfigDir, step.dstConfigDir)
	if err != nil {
		return "", err
	}

	// 3. get agent id
	backoff := retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault())
	// 3.1 unregister agent id [optional]
	if step.reRegisterAgentID {
		err = UnregisterAgentID(ctx, backoff, step.agentPath, step.agentConfigPath)
		if err != nil {
			logger.Error(constant.StepInstallNode, constant.StateFailed,
				fmt.Sprintf("unregister agent failed: %s", err))

			return "", err
		}
		logger.Info(constant.StepInstallNode, constant.StateRunning, "successfully unregister agent")
	}

	// 3.2 register agent id
	agentID, err := RegisterAgentID(ctx, backoff, step.agentPath, step.agentConfigPath, step.agentID)
	if err != nil {
		logger.Error(constant.StepInstallNode, constant.StateFailed,
			fmt.Sprintf("register agent failed: %v", err))

		return "", err
	}
	logger.Info(constant.StepInstallNode, constant.StateRunning, "successfully register agent")

	logger.Info(constant.StepInstallNode, constant.StateDone, "success install node")

	return agentID, nil
}
