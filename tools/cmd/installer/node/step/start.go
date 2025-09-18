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

	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/handler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/persistent"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/nodestarter"
	"github.com/spf13/cobra"
)

// NewStart creates a new start step command.
func NewStart() *cobra.Command {
	var (
		// pre-run.
		agentHandler agenthandler.IAgentHandler
	)

	stepCmd := &cobra.Command{
		Use:   "start",
		Short: "Start node",
		Long:  "Start node brings up all the services on this node",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			vars, err := persistent.GetVariables(cmd)
			if err != nil {
				return err
			}

			agentHandler = handler.NewAgentHandler(vars.NodeRole, vars.DeployDir, vars.DeployEnv)

			return nil
		},

		RunE: func(cmd *cobra.Command, _ []string) error {
			step := nodestarter.NewStep(nodestarter.StepArgs{
				AgentHandler: agentHandler,
			})

			if err := step.Run(cmd.Context()); err != nil {
				return err
			}

			fmt.Println("successfully started")

			return nil
		},
	}

	return stepCmd
}
