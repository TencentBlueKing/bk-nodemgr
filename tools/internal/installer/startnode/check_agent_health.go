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
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/retrier"
)

// AgentHealthState agent health state.
type AgentHealthState struct {
	OK   bool `json:"ok"`
	Data struct {
		Base    string `json:"base"`
		Cluster string `json:"cluster"`
		Data    string `json:"data"`
		File    string `json:"file"`
	} `json:"data"`
}

// CheckAgentHealth check agent health.
func CheckAgentHealth(ctx context.Context, agentPath, configPath string) (*AgentHealthState, error) {
	state := &AgentHealthState{}

	backoff := retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault())
	err := backoff.Do(ctx, func(attempt int) error {
		var stdout, stderr bytes.Buffer

		// nolint: gosec
		cmd := exec.CommandContext(ctx, agentPath,
			"-f", configPath,
			"--healthz",
		)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			logger.Warnf(constant.StepInstallAgent, constant.StateRunning,
				"check agent health failed, attempt(%d), stderr(%s), err: %v", attempt, stderr.String(), err)
		}

		healthzStr := stdout.String()
		healthzStr = strings.TrimPrefix(healthzStr, "healthz: ")

		if err := json.Unmarshal([]byte(healthzStr), state); err != nil {
			logger.Warnf(constant.StepInstallAgent, constant.StateRunning,
				"check agent health failed, attempt(%d), stderr(%s), err: %v", attempt, stderr.String(), err)

			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return state, nil
}
