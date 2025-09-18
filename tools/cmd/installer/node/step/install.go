/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package step

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/flag"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/handler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/persistent"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/nodeinstaller"
	"github.com/spf13/cobra"
)

// NewInstall creates a new install step command.
// nolint: lll
func NewInstall() *cobra.Command {
	var (
		// required flags.
		pkgPath string

		// optional flags.
		agentID string

		// pre-run.
		agentHandler   agenthandler.IAgentHandler
		persistentVars *persistent.Variables
	)

	stepCmd := &cobra.Command{
		Use:   "install",
		Short: "Install node without brings-up",
		Long:  "Install node without brings-up, run start after install",
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
			step := nodeinstaller.NewStep(nodeinstaller.StepArgs{
				AgentHandler:    agentHandler,
				AgentID:         agentID,
				ReRegisterAgent: agentID == "",
				PkgPath:         pkgPath,
				SrcConfigDir:    persistentVars.ConfigDir,
			})

			newAgentID, err := step.Run(cmd.Context())
			if err != nil {
				return err
			}

			fmt.Printf("successfully installed, agent-id(%s). run `start` to brings-up node\n", newAgentID)

			return nil
		},
	}

	/*
	 * required flags.
	 */
	stepCmd.Flags().StringVar(&pkgPath, flag.PkgFile, "", "path to release package file to install")
	_ = stepCmd.MarkFlagRequired(flag.PkgFile)

	/*
	 * optional flags.
	 */
	stepCmd.Flags().StringVar(&agentID, flag.AgentID, "", "existing agent-id to install with. if not given, will register a new one")

	return stepCmd
}
