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

	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/handler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/persistent"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/noderestarter"
	"github.com/spf13/cobra"
)

// NewRestart creates a new restart step command.
// nolint: lll
func NewRestart() *cobra.Command {
	var (
		// optional flags.
		force bool

		// pre-run.
		agentHandler agenthandler.IAgentHandler
	)

	stepCmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart node",
		Long:  "Restart node shutdown and brings up all the services",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			vars, err := persistent.GetVariables(cmd)
			if err != nil {
				return err
			}

			agentHandler = handler.NewAgentHandler(vars.NodeRole, vars.DeployDir, vars.DeployEnv)

			return nil
		},

		RunE: func(cmd *cobra.Command, _ []string) error {
			step := noderestarter.NewStep(noderestarter.StepArgs{
				AgentHandler: agentHandler,
				Force:        force,
			})

			if err := step.Run(cmd.Context()); err != nil {
				return err
			}

			fmt.Println("successfully started")

			return nil
		},
	}

	/*
	 * optional flags.
	 */
	stepCmd.Flags().BoolVarP(&force, "force", "f", false, "force stop will shutdown and brings up the services immediately without idle checking")

	return stepCmd
}
