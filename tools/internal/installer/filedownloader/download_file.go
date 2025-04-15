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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils/downloader"
)

const (
	maxTime = 300 * time.Second
)

// DownloadFile download file.
func DownloadFile(ctx context.Context, reqBody any, downloadURL, filePath string) error {
	downloadConfig := downloader.Config{
		URL:         downloadURL,
		Method:      http.MethodPost,
		RequestBody: reqBody,
		DestPath:    filePath,
		Timeout:     maxTime,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	lastProgress := int64(0)

	// start progress report.
	downloadConfig.ProgressFunc = func(current, total int64) {
		if current == total {
			logger.Infof(constant.StepDownloadFiles, constant.StateRunning,
				"download file complete, file-path(%s)", filePath)
		}

		if current-lastProgress < total/10 {
			return
		}

		logger.Infof(constant.StepDownloadFiles, constant.StateRunning,
			"download file running, file-path(%s),progress(%d/%d)", filePath, current, total)

		lastProgress = current
	}

	d := new(downloader.HTTPDownloader)
	err := d.Download(ctx, downloadConfig)

	if err != nil {
		return fmt.Errorf("download file failed, file-path(%s), err: %v", filePath, err)
	}

	return nil
}
