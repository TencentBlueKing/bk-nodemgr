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
	"fmt"
	"path/filepath"

	flag2 "github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/flag"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/handler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/persistent"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/nodestopper"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/nodeuninstaller"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/statusreporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/logreporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/types"
	"github.com/spf13/cobra"
)

// NewFullUninstall creates a new full uninstall command.
// nolint: lll, funlen, gocognit
func NewFullUninstall() *cobra.Command {
	var (
		// required flags.
		callbackSvrAddr string
		deployToken     string
		operInstID      string

		// optional flags.
		logDir   string
		logToStd bool

		// pre-run.
		persistentVars *persistent.Variables
		agentHandler   agenthandler.IAgentHandler
	)

	fullCmd := &cobra.Command{
		Use:   "full-uninstall",
		Short: "Full uninstall process",
		Long:  "Full uninstall process",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			vars, err := persistent.GetVariables(cmd)
			if err != nil {
				return err
			}
			persistentVars = vars

			if logDir == "" {
				logDir = filepath.Join(persistentVars.DataDir, "logs")
			}

			agentHandler = handler.NewAgentHandler(vars.NodeRole, vars.DeployDir, vars.DeployEnv)

			return nil
		},
		// nolint: nonamedreturns
		RunE: func(cmd *cobra.Command, _ []string) (runErr error) {
			// init log settings.
			lHandler := logreporter.NewHandler(logDir, logToStd, deployToken, operInstID, reportLogURL(callbackSvrAddr))
			if err := lHandler.Start(); err != nil {
				return fmt.Errorf("failed to init logger: %w", err)
			}
			defer lHandler.Stop()

			// report status.
			defer func() {
				state := types.ProcessStateSuccess
				if runErr != nil {
					state = types.ProcessStateFailed
				}

				_ = statusreporter.NewStep(statusreporter.StepArgs{
					Token:           deployToken,
					OperInstID:      operInstID,
					Status:          state,
					CallbackSvrAddr: callbackSvrAddr,
				}).Run(cmd.Context())
			}()

			// Stop node.
			if err := nodestopper.NewStep(nodestopper.StepArgs{
				AgentHandler: agentHandler,
				Force:        true,
			}).Run(cmd.Context()); err != nil {
				return err
			}

			// uninstall node.
			if err := nodeuninstaller.NewStep(nodeuninstaller.StepArgs{
				AgentHandler: agentHandler,
				Backup:       true,
			}).Run(cmd.Context()); err != nil {
				return err
			}

			return nil
		},
	}

	/*
	 * required flags.
	 */

	fullCmd.Flags().StringVar(&callbackSvrAddr, flag2.CallbackSvrAddr, "", "callback server address")
	_ = fullCmd.MarkFlagRequired(flag2.CallbackSvrAddr)

	fullCmd.Flags().StringVar(&deployToken, flag2.DeployToken, "", "deploy token, contains the details of files")
	_ = fullCmd.MarkFlagRequired(flag2.DeployToken)

	fullCmd.Flags().StringVar(&operInstID, flag2.OperInstID, "", "operation instance id")
	_ = fullCmd.MarkFlagRequired(flag2.OperInstID)

	/*
	 * optional flags.
	 */
	fullCmd.Flags().StringVar(&logDir, flag2.LogDir, "", "directory to save log files")
	fullCmd.Flags().BoolVar(&logToStd, flag2.LogToStd, false, "also output log to stdout")

	return fullCmd
}
