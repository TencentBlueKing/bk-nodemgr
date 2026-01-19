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
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/checkdeploy"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/datareporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/filedownloader"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/noderestarter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/nodeupgrader"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/statusreporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/logreporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
	"github.com/spf13/cobra"
)

// NewFullReconfig creates a new full reload config command.
// nolint: lll, funlen, gocognit
func NewFullReconfig() *cobra.Command {
	var (
		// required flags.
		callbackSvrAddr string
		deployToken     string
		operInstID      string

		// optional flags.
		logDir   string
		logToStd bool
		restart  bool
		force    bool

		// pre-run.
		persistentVars *persistent.Variables
		agentHandler   agenthandler.IAgentHandler
	)
	fullCmd := &cobra.Command{
		Use:   "full-reconfig",
		Short: "Full reconfig",
		Long:  "Full reconfig",
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

			// download config files.
			if err := filedownloader.NewStep(filedownloader.StepArgs{
				CallbackSvrAddr:      callbackSvrAddrs,
				NodeRole:             persistentVars.NodeRole,
				Generation:           persistentVars.Generation,
				DeployToken:          deployToken,
				ConfigSavedDir:       persistentVars.ConfigDir,
				SelectDownloads:      true,
				EnableDownloadConfig: true,
			}).Run(cmd.Context()); err != nil {
				return err
			}

			// upgrade node with only config files.
			if err := nodeupgrader.NewStep(nodeupgrader.StepArgs{
				AgentHandler:        agentHandler,
				SrcConfigDir:        persistentVars.ConfigDir,
				Backup:              true,
				SelectUpgrades:      true,
				EnableUpgradeConfig: true,
			}).Run(cmd.Context()); err != nil {
				return err
			}

			// restart node if need.
			if restart {
				if err := noderestarter.NewStep(noderestarter.StepArgs{
					AgentHandler: agentHandler,
					Force:        force,
				}).Run(cmd.Context()); err != nil {
					return err
				}

				// check deploy.
				if err := checkdeploy.NewStep(checkdeploy.StepArgs{
					AgentHandler: agentHandler,
				}).Run(cmd.Context()); err != nil {
					return err
				}
			}

			// get agent-id.
			agentID, err := agentHandler.Process().GetAgentID(cmd.Context())
			if err != nil {
				logger.Errorf(node.StepGeneral, "failed to get agent-id: %v", err)

				return fmt.Errorf("failed to get agent-id: %w", err)
			}

			// report data.
			if err := datareporter.NewStep(datareporter.StepArgs{
				CallbackSvrAddr: callbackSvrAddrs,
				Token:           deployToken,
				AgentID:         agentID,
				OperInstID:      operInstID,
			}).Run(cmd.Context()); err != nil {
				return err
			}

			return nil
		},
	}

	/*
	 * required flags.
	 */
	fullCmd.Flags().StringVar(&callbackSvrAddr, flag2.CallbackSvrAddr, "", "callback server address, for downloading config files and reporting status")
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
	fullCmd.Flags().BoolVar(&restart, flag2.Restart, false, "whether to restart node after reload config")
	fullCmd.Flags().BoolVar(&force, flag2.Force, false, "whether to force restart when --restart is set")

	return fullCmd
}
