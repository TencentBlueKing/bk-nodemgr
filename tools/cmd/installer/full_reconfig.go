/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package main

import (
	"fmt"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/flag"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/handler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/persistent"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/checkdeploy"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/datareporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/filedownloader"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/noderestarter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/nodeupgrader"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/statusreporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/types"
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
		logDir  string
		restart bool
		force   bool

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

			// init log settings.
			lHandler := newLoggerHandler(logDir, deployToken, operInstID, callbackSvrAddr)
			if err := lHandler.start(); err != nil {
				return fmt.Errorf("failed to init logger: %w", err)
			}
			defer lHandler.stop()

			// download config files.
			if err := filedownloader.NewStep(filedownloader.StepArgs{
				CallbackSvrAddr:      callbackSvrAddr,
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
				logger.Errorf(installer.StepGeneral, "failed to get agent-id: %v", err)

				return fmt.Errorf("failed to get agent-id: %w", err)
			}

			// report data.
			if err := datareporter.NewStep(datareporter.StepArgs{
				CallbackSvrAddr: callbackSvrAddr,
				Token:           deployToken,
				AgentID:         agentID,
			}).Run(cmd.Context()); err != nil {
				return err
			}

			return nil
		},
	}

	/*
	 * required flags.
	 */
	fullCmd.Flags().StringVar(&callbackSvrAddr, flag.CallbackSvrAddr, "", "callback server address, for downloading config files and reporting status")
	_ = fullCmd.MarkFlagRequired(flag.CallbackSvrAddr)

	fullCmd.Flags().StringVar(&deployToken, flag.DeployToken, "", "deploy token, contains the details of files")
	_ = fullCmd.MarkFlagRequired(flag.DeployToken)

	fullCmd.Flags().StringVar(&operInstID, flag.OperInstID, "", "operation instance id")
	_ = fullCmd.MarkFlagRequired(flag.OperInstID)

	/*
	 * optional flags.
	 */
	fullCmd.Flags().StringVar(&logDir, flag.LogDir, "", "directory to save log files")
	fullCmd.Flags().BoolVar(&restart, flag.Restart, false, "whether to restart node after reload config")
	fullCmd.Flags().BoolVar(&force, flag.Force, false, "whether to force restart when --restart is set")

	return fullCmd
}
