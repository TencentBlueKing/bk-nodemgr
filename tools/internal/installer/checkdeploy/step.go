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
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/gopool"
	"os"
	"strconv"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
)

// Step this step is used to check this gse node is deploy or not.
type Step struct {
	runDir string
}

// StepArgs ...
type StepArgs struct {
	RunDir string
}

// NewStep new step to check this gse node is deploy or not.
func NewStep(args StepArgs) *Step {
	step := &Step{
		runDir: args.RunDir,
	}

	return step
}

// Run the step to check this gse node is deploy or not.
func (step *Step) Run(_ context.Context) error {
	logger.Infof(constant.StepCheckDeploy, constant.StateStart, "start check deploy result")

	pidFiles, err := utils.ListFiles(step.runDir)
	if err != nil {
		logger.Errorf(constant.StepInstallNode, constant.StateRunning,
			"list files failed, dir-path(%s), err(%s)", step.runDir, err)

		return err
	}

	gp := gopool.NewPool()
	for idx := range pidFiles {
		pidFilePath := pidFiles[idx]
		gp.Go(func() error {
			logger.Infof(constant.StepCheckDeploy, constant.StateRunning, "pid-file-path(%s)", pidFilePath)

			if err := step.checkPidFile(pidFilePath); err != nil {
				logger.Errorf(constant.StepCheckDeploy, constant.StateFailed,
					"check pid file failed, pid-file-path(%s), err: %v", pidFilePath, err)

				return err
			}
			logger.Infof(constant.StepCheckDeploy, constant.StateRunning,
				"successfully check pid file, pid-file-path(%s)", pidFilePath)

			return nil
		})
	}
	if err := gp.Wait(); err != nil {
		logger.Errorf(constant.StepCheckDeploy, constant.StateFailed, "check pid file failed, err: %v", err)

		return err
	}

	logger.Infof(constant.StepCheckDeploy, constant.StateDone, "successfully check deploy result")

	return nil
}

// checkPidFile check pid file exist or not and use utils.CheckPIDExist to check pid exist or not.
func (step *Step) checkPidFile(pidFilePath string) error {
	// nolint: gosec
	pidStr, err := os.ReadFile(pidFilePath)
	if err != nil {
		return fmt.Errorf("read pid file failed, pid-file-path(%s), err: %w", pidFilePath, err)
	}

	pid, err := strconv.Atoi(string(pidStr))
	if err != nil {
		return fmt.Errorf("convert pid(%s) to int failed, pid-file-path(%s)", pidStr, pidFilePath)
	}

	result, err := utils.CheckPIDExist(pid)
	if err != nil {
		return fmt.Errorf("check pid(%d) exist failed, err: %w", pid, err)
	}

	if !result {
		return fmt.Errorf("check pid exist failed, pid(%d), err: %w", pid, err)
	}

	return nil
}
