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
	"errors"
	"fmt"
	"net/url"
	"runtime"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
)

// DownloadPkg ...
func DownloadPkg(ctx context.Context, pkgGeneration int, pkgPath, pkgVersion, downloadPoint, nodeRole string) error {
	if pkgPath == "" {
		logger.Errorf(constant.StepDownloadFiles, "pkg path is empty")
		return errors.New("pkg path is empty")
	}

	downloadURL, err := url.JoinPath(downloadPoint, "/download", nodeRole)
	if err != nil {
		logger.Errorf(constant.StepDownloadFiles, "download agent pkg failed: %v", err)
		return fmt.Errorf("download agent pkg failed: %v", err)
	}

	logger.Infof(constant.StepDownloadFiles,
		"download node pkg, download-url(%s), pkg-path(%s)", downloadURL, pkgPath)

	requestBody := struct {
		OSType     string `json:"os_type"`
		CPUArch    string `json:"cpu_arch"`
		Generation int    `json:"generation"`
		Version    string `json:"version"`
	}{
		OSType:     runtime.GOOS,
		CPUArch:    runtime.GOARCH,
		Generation: pkgGeneration,
		Version:    pkgVersion,
	}

	logger.Infof(constant.StepDownloadFiles,
		"download agent pkg request body: %v", requestBody)

	if err := DownloadFile(ctx, requestBody, downloadURL, pkgPath); err != nil {
		logger.Errorf(constant.StepDownloadFiles, "download node pkg failed, err: %v", err)

		return fmt.Errorf("download node pkg failed, err: %v", err)
	}

	return nil
}
