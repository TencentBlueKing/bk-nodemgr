/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package main of application.
package main

import (
	"fmt"
	"os"

	"github.com/TencentBlueKing/bk-nodemgr/internal/version"
	"github.com/spf13/cobra"
)

// application service entrypoint.
func main() {
	var rootCMD = &cobra.Command{
		Use:   "bk_nodeman_application",
		Short: "bk-nodeman application server",
		PreRun: func(_ *cobra.Command, _ []string) {
			fmt.Println(version.GetStartInfo())
		},
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println("welcome to use bk-nodeman-application, use `bk-nodeman-application -h` for help")
		},
	}

	// add sub commands.
	rootCMD.AddCommand(
		NewInitDataCMD(),
		NewMakeMigrationCMD(),
		NewMigrateCMD(),
		NewSchedulerCMD(),
		NewVersionCMD(),
		NewWebServerCMD(),
	)

	if err := rootCMD.Execute(); err != nil {
		fmt.Printf("failed to execute cmd, err: %v\n", err)
		os.Exit(1)
	}
}
