/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package main for installer
package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/flag"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/persistent"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/step"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/types"
	"github.com/spf13/cobra"
)

func main() {
	if err := NewRootCommand().Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// NewRootCommand creates the root command of installer.
// nolint: lll
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "installer",
		Short:        "installer",
		Long:         "nodemgr node installer",
		SilenceUsage: true,
		Version:      "1.0.0",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			vars, err := persistent.GetVariables(cmd)
			if err != nil {
				return err
			}

			if err = vars.EnsureDirs(); err != nil {
				return err
			}

			return nil
		},
	}

	// sub commands.
	rootCmd.AddCommand(step.NewStepComand())
	rootCmd.AddCommand(NewFullInstall())
	rootCmd.AddCommand(NewFullUpgrade())
	rootCmd.AddCommand(NewFullReconfig())

	/*
	 * persistent required flags.
	 */
	rootCmd.PersistentFlags().StringP(flag.DeployEnv, flag.DeployEnvS, "", "the deploy environment which to operate at")
	_ = rootCmd.MarkFlagRequired(flag.DeployEnv)

	/*
	 * persistent optional flags.
	 */
	rootCmd.PersistentFlags().IntP(flag.Generation, flag.GenerationS, int(types.Generation2), "the generation of this node, 1 or 2")
	rootCmd.PersistentFlags().StringP(flag.NodeRole, flag.NodeRoleS, string(types.NodeRoleAgent), "node role, agent or proxy")
	rootCmd.PersistentFlags().String(flag.BaseDeployDir, defaultBaseDeployDir(), "base deployed directory of this node, the deploy dir will be created under this directory with deploy-env")
	rootCmd.PersistentFlags().String(flag.BaseWorkDir, defaultBaseWorkDir(), "base work directory of this node, the work dir will be created under this directory with deploy-env")

	return rootCmd
}

func defaultBaseDeployDir() string {
	if runtime.GOOS == "windows" {
		return `C:\`
	}

	return "/usr/local/"
}

func defaultBaseWorkDir() string {
	if runtime.GOOS == "windows" {
		return `C:\tmp\bknm\`
	}

	return "/tmp/bknm/"
}
