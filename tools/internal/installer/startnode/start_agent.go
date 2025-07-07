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
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/gopool"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/retrier"
)

// StartNode ...
func StartNode(ctx context.Context, gseCtlPath string) error {
	defer func() {
		if r := recover(); r != nil {
			logger.Error(constant.StepStartNode, fmt.Sprintf("start agent panic, recovery: %s", r))
		}
	}()
	backoff := retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault())
	err := backoff.Do(ctx, func(attempt int) error {
		// nolint: gosec
		cmd := exec.CommandContext(ctx,
			gseCtlPath,
			"start",
		)
		cmd.Dir = filepath.Dir(gseCtlPath)
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("failed to create stdout pipe, err: %w", err)
		}

		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			return fmt.Errorf("failed to create stderr pipe, err: %w", err)
		}

		logger.Infof(constant.StepStartNode, "run agent start cmd, cmd(%s)", cmd.String())

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start command, err: %w", err)
		}

		gp := gopool.NewPool()
		gp.Go(func() error {
			scanner := bufio.NewScanner(stderrPipe)
			for scanner.Scan() {
				line := scanner.Text()
				logger.Warnf(constant.StepStartNode,
					"agent stderr: %s", line)
			}

			return nil
		})

		var hasOutput bool
		gp.Go(func() error {
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				line := scanner.Text()
				if line != "" {
					hasOutput = true
					logger.Infof(constant.StepStartNode,
						"agent output: %s", line)
				}
			}

			return nil
		})

		err = cmd.Wait()

		_ = gp.Wait()

		if err != nil {
			logger.Warnf(constant.StepStartNode, "start agent failed, attempt: %d, err: %v",
				attempt, err)

			return err
		}

		if !hasOutput {
			logger.Infof(constant.StepStartNode, "successfully start agent (no output)")
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
