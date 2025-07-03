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

package main

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/uninstallnode"
)

func stepUninstallNode(ctx context.Context) error {
	if !GetReinstall() {
		// don't reinstall

		return nil
	}

	step := uninstallnode.NewStep(uninstallnode.StepArgs{
		SetupDirPath: GetSetupDir(),
		BinDirPath:   GetBinDir(),
		GseCtlPath:   GetGseCtlPath(),
		DeployEnv:    GetDeployEnv(),
	})
	if err := step.Run(ctx); err != nil {
		return fmt.Errorf("uninstall step failed, err: %v", err)
	}

	return nil
}
