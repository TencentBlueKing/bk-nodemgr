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

	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/flag"

	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/handler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/persistent"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/nodeuninstaller"
	"github.com/spf13/cobra"
)

// NewUninstall creates a new uninstall step command.
func NewUninstall() *cobra.Command {
	var (
		// optional flags.
		renewGSEProc bool
		renewGSETask bool

		// pre-run.
		agentHandler agenthandler.IAgentHandler
	)

	stepCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall node",
		Long:  "Uninstall node",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			vars, err := persistent.GetVariables(cmd)
			if err != nil {
				return err
			}

			agentHandler = handler.NewAgentHandler(vars.NodeRole, vars.DeployDir, vars.DeployEnv)

			return nil
		},

		RunE: func(cmd *cobra.Command, _ []string) error {
			step := nodeuninstaller.NewStep(nodeuninstaller.StepArgs{
				AgentHandler: agentHandler,
				PreserveOptions: agenthandler.PreserveOptions{
					RenewGSEProc: renewGSEProc,
					RenewGSETask: renewGSETask,
				},
			})

			if err := step.Run(cmd.Context()); err != nil {
				return err
			}

			fmt.Println("successfully uninstalled")

			return nil
		},
	}

	stepCmd.Flags().BoolVar(&renewGSEProc, flag.RenewGSEProc, false, "renew GSE .proc runtime file instead of preserving it")
	stepCmd.Flags().BoolVar(&renewGSETask, flag.RenewGSETask, false, "renew GSE .task runtime file instead of preserving it")

	return stepCmd
}
