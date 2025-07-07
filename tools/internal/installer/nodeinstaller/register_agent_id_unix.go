//go:build linux || darwin || freebsd || aix

/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodeinstaller

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/retrier"
)

// RegisterAgentID run the register agent ID command.
func RegisterAgentID(ctx context.Context, retrier retrier.Retrier, agentPath, configPath, agentID string) (
	string, error) {

	logger.Infof(constant.StepInstallNode,
		"register agent id, agent-path(%s), config-path(%s)", agentPath, configPath)

	var agentIDStr string
	err := retrier.Do(ctx, func(attempt int) error {
		var stdout, stdErr bytes.Buffer

		cmd := exec.CommandContext(ctx, agentPath,
			"-f", configPath,
			"--register", agentID,
		)
		cmd.Dir = filepath.Dir(agentPath)
		cmd.Stdout = &stdout
		cmd.Stderr = &stdErr

		logger.Infof(constant.StepInstallNode, "register agent id, cmd(%s)", cmd.String())

		err := cmd.Run()
		if err != nil {
			logger.Warnf(constant.StepInstallNode,
				"register agent id failed, attempt(%d), stdOut(%s), stdErr(%s), err: %v",
				attempt,
				strings.ReplaceAll(stdout.String(), "\n", ""),
				strings.ReplaceAll(stdErr.String(), "\n", ""),
				err)

			return err
		}

		regex := `agent-id:\s+[A-Za-z0-9]+`
		agentIDStr = regexp.MustCompile(regex).FindString(stdout.String())
		agentIDStr = strings.TrimPrefix(agentIDStr, "agent-id: ")

		return nil
	})
	if err != nil {
		return "", err
	}

	return agentIDStr, nil
}
