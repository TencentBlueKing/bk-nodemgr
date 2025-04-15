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
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
)

// MakeGseBinFileExecutable makes gse bin file executable.
func MakeGseBinFileExecutable(_ context.Context, binDir string) error {
	binFiles, err := utils.ListFiles(binDir)
	if err != nil {
		logger.Errorf(constant.StepInstallNode, constant.StateRunning,
			"list files failed, dir-path(%s), err(%s)", binDir, err)

		return err
	}

	for _, binFile := range binFiles {
		if err := utils.MakeExecutable(binFile); err != nil {
			logger.Errorf(constant.StepInstallNode, constant.StateRunning,
				"make file executable failed, file-path(%s), err(%s)", binFile, err)

			return err
		}
		logger.Infof(constant.StepInstallNode, constant.StateRunning,
			"successfully make file executable, file-path(%s)", binFile)
	}

	return nil
}
