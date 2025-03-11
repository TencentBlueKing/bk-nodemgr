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

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/retrier"
)

// UnregisterAgentID run the unregister agent ID command.
func UnregisterAgentID(ctx context.Context, retrier retrier.Retrier, agentPath, configPath string) error {
	err := retrier.Do(ctx, func(attempt int) error {
		var stderr bytes.Buffer

		cmd := exec.CommandContext(ctx, agentPath,
			"-f", configPath,
			"--unregister",
		)
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			logger.Warnf(constant.StepInstallAgent, constant.StateRunning,
				"unregister agent id failed, attempt(%d), err: %v", attempt, err)

			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	logger.Infof(constant.StepInstallAgent, constant.StateRunning, "unregister agent id success")

	return nil
}
