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

	nodeFlag "github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/flag"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/handler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/persistent"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	nodeInstaller "github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/nodestopper"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/nodeuninstaller"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/statusreporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/systeminfo"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/logreporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
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
			callbackSvrAddrs := utils.SplitServerAddrs(callbackSvrAddr)
			if len(callbackSvrAddrs) == 0 {
				return fmt.Errorf("callback server address is empty or invalid")
			}
			logURLs, err := reportLogURLs(callbackSvrAddrs)
			if err != nil {
				return fmt.Errorf("failed to build log report URLs: %w", err)
			}
			lHandler := logreporter.NewHandler(logDir, logToStd, deployToken, operInstID, logURLs)
			if err := lHandler.Start(); err != nil {
				return fmt.Errorf("failed to init logger: %w", err)
			}
			defer lHandler.Stop()
			systeminfo.LogInitialTargetInfo(nodeInstaller.StepGeneral)

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
					CallbackSvrAddr: callbackSvrAddrs,
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
				// Intentionally use zero-value PreserveOptions: full-uninstall preserves existing GSE runtime files by default.
				PreserveOptions: agenthandler.PreserveOptions{},
			}).Run(cmd.Context()); err != nil {
				return err
			}

			return nil
		},
	}

	/*
	 * required flags.
	 */

	fullCmd.Flags().StringVar(&callbackSvrAddr, nodeFlag.CallbackSvrAddr, "", "callback server address")
	_ = fullCmd.MarkFlagRequired(nodeFlag.CallbackSvrAddr)

	fullCmd.Flags().StringVar(&deployToken, nodeFlag.DeployToken, "", "deploy token, contains the details of files")
	_ = fullCmd.MarkFlagRequired(nodeFlag.DeployToken)

	fullCmd.Flags().StringVar(&operInstID, nodeFlag.OperInstID, "", "operation instance id")
	_ = fullCmd.MarkFlagRequired(nodeFlag.OperInstID)

	/*
	 * optional flags.
	 */
	fullCmd.Flags().StringVar(&logDir, nodeFlag.LogDir, "", "directory to save log files")
	fullCmd.Flags().BoolVar(&logToStd, nodeFlag.LogToStd, false, "also output log to stdout")

	return fullCmd
}
