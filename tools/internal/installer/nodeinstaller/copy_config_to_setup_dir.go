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
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
)

// CopyConfigFilesToSetupDir copy config files to setup dir.
func CopyConfigFilesToSetupDir(srcConfigDir, dstConfigDir string) error {
	if srcConfigDir == "" {
		logger.Errorf(constant.StepInstallNode, constant.StateFailed, "src config dir is empty")

		return errors.New("src config dir is empty")
	}

	if dstConfigDir == "" {
		logger.Errorf(constant.StepInstallNode, constant.StateFailed, "dst config dir is empty")

		return errors.New("dst config dir is empty")
	}

	err := utils.CopyDir(srcConfigDir, dstConfigDir)
	if err != nil {
		logger.Errorf(constant.StepInstallNode, constant.StateFailed,
			"copy config files to setup dir failed, err: %v", err)

		return fmt.Errorf("copy config files to setup dir failed, err: %v", err)
	}

	logger.Info(constant.StepInstallNode, constant.StateRunning, "successfully copy config files to setup dir")

	return nil
}
