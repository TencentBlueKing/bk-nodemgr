/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package step

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/flag"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/handler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/persistent"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/nodeupgrader"
	"github.com/spf13/cobra"
)

// NewUpgrade creates a new upgrade step command.
// nolint: lll
func NewUpgrade() *cobra.Command {
	var (
		// required flags.
		pkgPath string

		// pre-run.
		agentHandler   agenthandler.IAgentHandler
		persistentVars *persistent.Variables
	)

	stepCmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade node without restart",
		Long:  "Upgrade node without restart, run restart after upgrade",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			vars, err := persistent.GetVariables(cmd)
			if err != nil {
				return err
			}
			persistentVars = vars

			agentHandler = handler.NewAgentHandler(vars.NodeRole, vars.DeployDir, vars.DeployEnv)

			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			step := nodeupgrader.NewStep(nodeupgrader.StepArgs{
				AgentHandler: agentHandler,
				PkgPath:      pkgPath,
				SrcConfigDir: persistentVars.ConfigDir,
			})

			if err := step.Run(cmd.Context()); err != nil {
				return err
			}

			fmt.Println("successfully upgraded, run `restart` to restart node")

			return nil
		},
	}

	/*
	 * required flags.
	 */
	stepCmd.Flags().StringVar(&pkgPath, flag.PkgFile, "", "path to release package file to install")
	_ = stepCmd.MarkFlagRequired(flag.PkgFile)

	return stepCmd
}
