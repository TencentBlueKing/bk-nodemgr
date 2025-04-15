/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package startnode ...
package startnode

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/gopool"
	"os/exec"
	"strings"
)

// StartProxy ...
func StartProxy(ctx context.Context,
	dataProxyPath string,
	dataProxyConfigPath string,
	fileProxyPath string,
	fileProxyConfigPath string,
) error {
	defer func() {
		if r := recover(); r != nil {
			logger.Error(constant.StepStartNode, constant.StateFailed,
				fmt.Sprintf("start proxy panic: %s", r))
		}
	}()

	gp := gopool.NewPool()
	gp.Go(func() error {
		backoff := retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault())
		err := backoff.Do(ctx, func(attempt int) error {
			var stdOut, stdErr bytes.Buffer

			// nolint: gosec
			cmd := exec.CommandContext(ctx,
				dataProxyPath,
				"-f",
				dataProxyConfigPath,
			)
			cmd.Dir = filepath.Dir(dataProxyPath)
			cmd.Stdout = &stdOut
			cmd.Stderr = &stdErr

			logger.Debugf(constant.StepStartNode, constant.StateRunning,
				"run data proxy start cmd: %s", cmd.String())

			err := cmd.Run()
			if err != nil {
				logger.Warn(constant.StepStartNode, constant.StateRunning,
					"start data proxy failed, attempt: %d, stderr: %s, err: %w",
					attempt, stdErr.String(), err)

				return err
			}

			if stdErr.String() != "" {
				logger.Warn(constant.StepStartNode, constant.StateRunning,
					"start data proxy failed, attempt: %d, stderr: %s",
					attempt, stdErr.String())
			}

			stdOutStr := stdOut.String()
			if stdOutStr != "" {
				outputLines := strings.Split(strings.TrimSpace(stdOutStr), "\n")
				for _, line := range outputLines {
					if line != "" {
						logger.Infof(constant.StepStartNode, constant.StateRunning, "dat proxy output: %s", line)
					}
				}
			} else {
				logger.Infof(constant.StepStartNode, constant.StateRunning, "start data proxy success (no output)")
			}

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})

	gp.Go(func() error {
		backoff := retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault())
		err := backoff.Do(ctx, func(attempt int) error {
			var stdOut, stdErr bytes.Buffer

			// nolint: gosec
			cmd := exec.CommandContext(ctx,
				fileProxyPath,
				"-f",
				fileProxyConfigPath,
			)
			cmd.Dir = filepath.Dir(fileProxyPath)
			cmd.Stdout = &stdOut
			cmd.Stderr = &stdErr

			logger.Debugf(constant.StepStartNode, constant.StateRunning,
				"run file proxy start cmd: %s", cmd.String())

			err := cmd.Run()
			if err != nil {
				logger.Warn(constant.StepStartNode, constant.StateRunning,
					"start file proxy failed, attempt: %d, stderr: %s, err: %w",
					attempt, stdErr.String(), err)

				return err
			}

			if stdErr.String() != "" {
				logger.Warn(constant.StepStartNode, constant.StateRunning,
					"start file proxy failed, attempt: %d, stderr: %s",
					attempt, stdErr.String())
			}

			stdOutStr := stdOut.String()
			if stdOutStr != "" {
				outputLines := strings.Split(strings.TrimSpace(stdOutStr), "\n")
				for _, line := range outputLines {
					if line != "" {
						logger.Infof(constant.StepStartNode, constant.StateRunning, "file proxy output: %s", line)
					}
				}
			} else {
				logger.Infof(constant.StepStartNode, constant.StateRunning, "start file proxy success (no output)")
			}

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})

	if err := gp.Wait(); err != nil {
		return fmt.Errorf("start proxy failed, err: %w", err)
	}

	return nil
}
