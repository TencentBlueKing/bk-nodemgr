/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package checkdeploy ...
package checkdeploy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
)

// Step this step is used to check this gse node is deploy or not.
type Step struct {
	setupDir string
}

// StepArgs ...
type StepArgs struct {
	SetupDir string
}

// NewStep new step to check this gse node is deploy or not.
func NewStep(args StepArgs) *Step {
	step := &Step{
		setupDir: args.SetupDir,
	}

	return step
}

// Run the step to check this gse node is deploy or not.
func (step *Step) Run(_ context.Context) error {
	logger.Infof(constant.StepCheckDeploy, constant.StateStart, "start check deploy result")

	pidFilePath := filepath.Join(step.setupDir, "bin", "run", "agent.pid")
	logger.Infof(constant.StepCheckDeploy, constant.StateRunning, "pid-file-path(%s)", pidFilePath)

	// nolint: gosec
	pidStr, err := os.ReadFile(pidFilePath)
	if err != nil {
		logger.Errorf(constant.StepCheckDeploy, constant.StateFailed,
			"read pid file failed, pid-file-path(%s), err: %v", pidFilePath, err)

		return fmt.Errorf("read pid file failed, pid-file-path(%s), err: %w", pidFilePath, err)
	}

	pid, err := strconv.Atoi(string(pidStr))
	if err != nil {
		logger.Errorf(constant.StepCheckDeploy, constant.StateFailed, "convert pid(%s) to int failed, pid-file-path(%s)",
			pidStr, pidFilePath)

		return fmt.Errorf("convert pid(%s) to int failed, pid-file-path(%s)", pidStr, pidFilePath)
	}

	result, err := utils.CheckPIDExist(pid)
	if err != nil {
		logger.Errorf(constant.StepCheckDeploy, constant.StateFailed, "check pid(%d) exist failed, err: %v", pid, err)

		return fmt.Errorf("check pid(%d) exist failed, err: %w", pid, err)
	}

	if !result {
		logger.Errorf(constant.StepCheckDeploy, constant.StateFailed, "check pid exist failed, pid(%d), err: %v", pid, err)

		return fmt.Errorf("check pid exist failed, pid(%d), err: %w", pid, err)
	}

	logger.Infof(constant.StepCheckDeploy, constant.StateDone, "successfully check deploy result")

	return nil
}
