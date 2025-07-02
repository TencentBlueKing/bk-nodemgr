/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package main ...
package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewStepUninstallAgent ...
func NewStepUninstallAgent() *cobra.Command {
	var (
		setupDirPath string
	)

	stepCmd := &cobra.Command{
		Use:   "step_uninstall",
		Short: "uninstall node",
		Long:  "the step is used to uninstall node",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if err := SetSetupDir(setupDirPath); err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := stepUninstallNode(cmd.Context()); err != nil {
				return err
			}

			fmt.Println("successfully uninstall node")

			return nil
		},
	}

	stepCmd.Flags().StringVar(&setupDirPath, CmdFlagSetupDirPath, "", "setup dir path")
	_ = stepCmd.MarkFlagRequired(CmdFlagSetupDirPath)

	return stepCmd
}
