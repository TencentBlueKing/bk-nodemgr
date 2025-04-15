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
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/retrier"
)

// StopNode stop agent.
// nolint: gocognit
func StopNode(ctx context.Context, gseCtlPath string) error {
	defer func() {
		if r := recover(); r != nil {
			logger.Error(constant.StepUninstallNode, constant.StateFailed,
				fmt.Sprintf("stop agent panic: %s", r))
		}
	}()

	backoff := retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault())
	err := backoff.Do(ctx, func(attempt int) error {
		var stdOut, stdErr bytes.Buffer

		// nolint: gosec
		cmd := exec.CommandContext(ctx,
			gseCtlPath,
			"stop",
		)
		cmd.Stdout = &stdOut
		cmd.Stderr = &stdErr

		logger.Infof(constant.StepUninstallNode, constant.StateRunning,
			"run agent stop cmd: %s", cmd.String())

		err := cmd.Run()
		if err != nil {
			logger.Warn(constant.StepUninstallNode, constant.StateRunning,
				"stop agent failed, attempt: %d, stderr: %s, err: %w",
				attempt, stdErr.String(), err)

			return err
		}

		if stdErr.String() != "" {
			logger.Warn(constant.StepUninstallNode, constant.StateRunning,
				"stop agent failed, attempt: %d, stderr: %s",
				attempt, stdErr.String())
		}

		stdOutStr := stdOut.String()
		if stdOutStr != "" {
			outputLines := strings.Split(strings.TrimSpace(stdOutStr), "\n")
			for _, line := range outputLines {
				if line != "" {
					logger.Infof(constant.StepUninstallNode, constant.StateRunning, "agent output: %s", line)
				}

				if strings.Contains(line, "gse agent stop failed") {
					return errors.New("stop agent failed")
				}
			}
		} else {
			logger.Infof(constant.StepUninstallNode, constant.StateRunning, "stop agent success (no output)")
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
