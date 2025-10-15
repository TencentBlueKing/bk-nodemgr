/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package node

import (
	"runtime"

	nodeflag "github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/flag"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/persistent"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/step"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/types"
	"github.com/spf13/cobra"
)

// NewNodeCommand creates a new node sub command.
// nolint: lll
func NewNodeCommand() *cobra.Command {
	nodeCommand := &cobra.Command{
		Use:   "node",
		Short: "Node command",
		Long:  "Node command",
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	nodeCommand.AddCommand(step.NewStepComand())
	nodeCommand.AddCommand(NewFullInstall())
	nodeCommand.AddCommand(NewFullUpgrade())
	nodeCommand.AddCommand(NewFullReconfig())
	nodeCommand.AddCommand(NewFullUninstall())

	/*
	 * persistent required flags.
	 */
	nodeCommand.PersistentFlags().StringP(nodeflag.DeployEnv, nodeflag.DeployEnvS, "", "the deploy environment which to operate at")
	_ = nodeCommand.MarkFlagRequired(nodeflag.DeployEnv)

	/*
	 * persistent optional flags.
	 */
	nodeCommand.PersistentFlags().StringP(nodeflag.NodeRole, nodeflag.NodeRoleS, string(types.NodeRoleAgent), "node role, agent or proxy")
	nodeCommand.PersistentFlags().IntP(nodeflag.Generation, nodeflag.GenerationS, int(types.Generation2), "the generation of this node, 1 or 2")
	nodeCommand.PersistentFlags().String(nodeflag.BaseDeployDir, defaultBaseDeployDir(), "base deployed directory of this node, the deploy dir will be created under this directory with deploy-env")
	nodeCommand.PersistentFlags().String(nodeflag.BaseWorkDir, defaultBaseWorkDir(), "base work directory of this node, the work dir will be created under this directory with deploy-env")

	return nodeCommand
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
