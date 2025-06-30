/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package checkdeploy

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/winapi"
)

const gseAgentDaemonName = "gse_agent_daemon"

// TODO: make this configurable
const gseProxyPort = 28668

func (step *Step) checkDeploy(ctx context.Context) error {
	gseAgentDaemonSvcName := gseAgentDaemonName
	if step.deployEnv != "gse" {
		gseAgentDaemonSvcName = gseAgentDaemonSvcName + "_" + step.deployEnv
	}

	r := retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault())
	err := r.Do(ctx, func(attempt int) error {
		status, err := winapi.GetServiceStatus(gseAgentDaemonSvcName)
		if err != nil {
			return fmt.Errorf("check service status failed, err: %w", err)
		}

		if status != winapi.WinSvcStatusRunning {
			return fmt.Errorf("check service status failed, status: %s", status)
		}

		logger.Infof(constant.StepCheckDeploy, "windows-service(%s), status: %s", gseAgentDaemonSvcName, status)

		return nil
	})
	if err != nil {
		return err
	}

	err = r.Do(ctx, func(attempt int) error {
		cmd := exec.Command("netstat", "-an")
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("check netstat status failed, err: %w", err)
		}

		portStr := fmt.Sprintf(":%d", gseProxyPort)
		linkStrs := []string{}
		for _, line := range strings.Split(string(output), "\n") {
			if strings.Contains(line, portStr) && strings.Contains(line, "ESTABLISHED") {
				linkStrs = append(linkStrs, line)
			}
		}

		if len(linkStrs) == 0 {
			return fmt.Errorf("no found established connections on port %d", gseProxyPort)
		}

		for _, linkStr := range linkStrs {
			logger.Infof(constant.StepCheckDeploy, "established link:%s", linkStr)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
