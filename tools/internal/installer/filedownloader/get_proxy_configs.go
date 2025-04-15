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

// GetFileProxyConf get file proxy conf.
func GetFileProxyConf(ctx context.Context, tmpFileProxyConfPath, nodeRole, token, callbackEndpoint string) error {
	requestBody := GetFileProxyConfReq{
		OSType:   runtime.GOOS,
		CPUArch:  runtime.GOARCH,
		NodeRole: nodeRole,
		Token:    token,
	}

	downloadURL, err := url.JoinPath(callbackEndpoint, "/callback/workflow/node_install/get_file_proxy_config")
	if err != nil {
		logger.Errorf(constant.StepDownloadFiles, constant.StateFailed,
			"get file proxy config failed, err: %v", err)

		return fmt.Errorf("get file proxy config failed, err: %v", err)
	}

	if err := DownloadFile(ctx, requestBody, downloadURL, tmpFileProxyConfPath); err != nil {
		logger.Errorf(constant.StepDownloadFiles, constant.StateFailed, "get file proxy conf failed, err: %v", err)

		return fmt.Errorf("get file proxy conf failed, err: %v", err)
	}

	return nil
}

// GetFileProxyConfReq get file proxy config request.
type GetFileProxyConfReq struct {
	OSType   string `json:"os_type"`
	CPUArch  string `json:"cpu_arch"`
	NodeRole string `json:"node_role"`
	Token    string `json:"token"`
}

// GetDataProxyConf get data proxy conf.
func GetDataProxyConf(ctx context.Context, tmpDataProxyConfPath, nodeRole, token, callbackEndpoint string) error {
	requestBody := GetDataProxyConfReq{
		OSType:   runtime.GOOS,
		CPUArch:  runtime.GOARCH,
		NodeRole: nodeRole,
		Token:    token,
	}

	downloadURL, err := url.JoinPath(callbackEndpoint, "/callback/workflow/node_install/get_data_proxy_config")
	if err != nil {
		logger.Errorf(constant.StepDownloadFiles, constant.StateFailed, "get data proxy config failed, err: %v", err)

		return fmt.Errorf("get data proxy config failed, err: %v", err)
	}

	if err := DownloadFile(ctx, requestBody, downloadURL, tmpDataProxyConfPath); err != nil {
		logger.Errorf(constant.StepDownloadFiles, constant.StateFailed, "get data proxy conf failed, err: %v", err)

		return fmt.Errorf("get data proxy conf failed, err: %v", err)
	}

	return nil
}

// GetDataProxyConfReq get file proxy config request.
type GetDataProxyConfReq struct {
	OSType   string `json:"os_type"`
	CPUArch  string `json:"cpu_arch"`
	NodeRole string `json:"node_role"`
	Token    string `json:"token"`
}
