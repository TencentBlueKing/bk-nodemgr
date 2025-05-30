//go:build linux || darwin || freebsd || aix

/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package precheck

import (
	"context"
	"fmt"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
)

// CheckRemnantProcessInSetupDir check for any process remnants in SetupDir.
func CheckRemnantProcessInSetupDir(_ context.Context, setupDirPath string) error {
	process, err := utils.GetSameSpaceProcesses()
	if err != nil {
		return fmt.Errorf("get same space processes failed, err: %v", err)
	}

	for _, p := range process {
		if strings.HasPrefix(p.FullPath, setupDirPath) {
			return fmt.Errorf("remnant process found, pid(%d), file path(%s)", p.PID, p.FullPath)
		}
	}

	return nil
}
