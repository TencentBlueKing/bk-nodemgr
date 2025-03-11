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
	"net/http"
	"net/url"
	"runtime"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils/downloader"
)

// GetCheckList ...
func GetCheckList(ctx context.Context, nodeType, checkListPath, token, callbackEndpoint string) error {
	requestBody := GetCheckListReq{
		OSType:   runtime.GOOS,
		CPUArch:  runtime.GOARCH,
		NodeType: nodeType,
		Token:    token,
	}

	downloadURL, err := url.JoinPath(callbackEndpoint, "/callback/workflow/node_install/get_check_list")
	if err != nil {
		logger.Errorf(constant.StepDownloadFiles, constant.StateFailed, "get check list failed: %v", err)
		return fmt.Errorf("get check list failed: %v", err)
	}

	downloadConfig := downloader.Config{
		URL:         downloadURL,
		Method:      http.MethodPost,
		RequestBody: requestBody,
		DestPath:    checkListPath,
		Timeout:     maxTime,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
	logger.Infof(constant.StepDownloadFiles, constant.StateRunning,
		"download check list, url(%s), file-path(%s)", downloadConfig.URL, downloadConfig.DestPath)

	lastProgress := int64(0)

	// start progress report.
	downloadConfig.ProgressFunc = func(current, total int64) {
		if current == total {
			logger.Infof(constant.StepDownloadFiles, constant.StateRunning,
				"get check list complete")
		}

		if current-lastProgress < 1024*1024 {
			return
		}

		logger.Infof(constant.StepDownloadFiles, constant.StateRunning,
			"get check list progress: %d/%d", current, total)
		lastProgress = current
	}

	d := new(downloader.HTTPDownloader)
	err = d.Download(ctx, downloadConfig)

	if err != nil {
		logger.Errorf(constant.StepDownloadFiles, constant.StateFailed, "download check list failed: %v", err)

		return fmt.Errorf("download check list failed: %v", err)
	}

	return nil
}

// GetCheckListReq get check list request.
type GetCheckListReq struct {
	OSType   string `json:"os_type"`
	CPUArch  string `json:"cpu_arch"`
	NodeType string `json:"node_type"`
	Token    string `json:"token"`
}
