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

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/nodeinstaller"
	"github.com/spf13/cobra"
)

// NewStepInstallAgent ...
func NewStepInstallAgent() *cobra.Command {
	var (
		gseRoot string
		pkgName string
	)
	stepCmd := &cobra.Command{
		Use:   "step_install",
		Short: "install",
		Long:  "install",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if err := SetGseRoot(gseRoot); err != nil {
				return err
			}

			if err := SetPkgName(pkgName); err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			step := nodeinstaller.NewStep(nodeinstaller.StepArgs{
				ReRegisterAgentID: false,
				SetupDirPath:      GetSetupDir(),
				PkgPath:           GetGsePkgPath(),
				SrcConfigDir:      GetTmpConfigDir(),
				Overwrite:         false,
			})
			agentID, err := step.Run(cmd.Context())
			if err != nil {
				return err
			}

			fmt.Print(agentID)

			return nil
		},
	}
	stepCmd.Flags().StringVar(&pkgName, CmdFlagPkgName, "", "gse pkg name")
	stepCmd.Flags().StringVar(&gseRoot, CmdFlagGseRoot, CmdDefaultGseRoot(), "this is the gse root path")

	_ = stepCmd.MarkFlagRequired(CmdFlagGseRoot)
	_ = stepCmd.MarkFlagRequired(CmdFlagPkgName)

	return stepCmd
}
