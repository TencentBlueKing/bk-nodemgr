/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package filedownloader

import (
	"context"
	"fmt"
	"net/url"
	"runtime"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
)

// GetAgentConfig get agent config file.
func GetAgentConfig(ctx context.Context, callbackEndpoint, tmpAgentConfPath, nodeType, token string) error {
	requestBody := GetAgentConfigReq{
		OSType:   runtime.GOOS,
		CPUArch:  runtime.GOARCH,
		NodeType: nodeType,
		Token:    token,
	}

	downloadURL, err := url.JoinPath(callbackEndpoint, "/callback/workflow/node_install/get_agent_config")
	if err != nil {
		logger.Errorf(constant.StepDownloadFiles, constant.StateFailed, "get agent config failed, err: %v", err)
		return fmt.Errorf("get agent config failed, err: %v", err)
	}

	if err := DownloadFile(ctx, requestBody, downloadURL, tmpAgentConfPath); err != nil {
		logger.Errorf(constant.StepDownloadFiles, constant.StateFailed, "get agent conf failed, err: %v", err)

		return fmt.Errorf("get agent conf failed, err: %v", err)
	}

	return nil
}

// GetAgentConfigReq get agent config request.
type GetAgentConfigReq struct {
	OSType   string `json:"os_type"`
	CPUArch  string `json:"cpu_arch"`
	NodeType string `json:"node_type"`
	Token    string `json:"token"`
}
