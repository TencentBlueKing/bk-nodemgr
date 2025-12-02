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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	flag2 "github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/flag"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/handler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/persistent"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/step"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/checkdeploy"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/datareporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/filedownloader"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/nodeinstaller"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/nodestarter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/nodestopper"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/nodeuninstaller"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/precheck"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/statusreporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/logreporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/types"
	"github.com/spf13/cobra"
)

// NewFullInstall creates a new full install command.
// nolint: lll, funlen, gocognit
func NewFullInstall() *cobra.Command {
	var (
		// required flags.
		downloadSvrAddr string
		callbackSvrAddr string
		deployToken     string
		operInstID      string
		nodeVersion     string

		// optional flags.
		logDir  string
		agentID string

		// pre-run.
		persistentVars   *persistent.Variables
		preCheckListConf string
		pkgPath          string
		agentHandler     agenthandler.IAgentHandler
	)

	fullCmd := &cobra.Command{
		Use:   "full-install",
		Short: "Full install process",
		Long:  "Full install process",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			vars, err := persistent.GetVariables(cmd)
			if err != nil {
				return err
			}
			persistentVars = vars

			preCheckListConf = filepath.Join(persistentVars.DataDir, "precheck.json")
			pkgPath = filepath.Join(persistentVars.DataDir, step.GenReleasePkgName(persistentVars.NodeRole, persistentVars.Generation, nodeVersion))

			if logDir == "" {
				logDir = filepath.Join(persistentVars.DataDir, "logs")
			}

			agentHandler = handler.NewAgentHandler(vars.NodeRole, vars.DeployDir, vars.DeployEnv)

			return nil
		},
		// nolint: nonamedreturns
		RunE: func(cmd *cobra.Command, _ []string) (runErr error) {
			// init log settings.
			lHandler := logreporter.NewHandler(logDir, deployToken, operInstID, reportLogUrl(callbackSvrAddr))
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

			// download files.
			if err := filedownloader.NewStep(filedownloader.StepArgs{
				DownloadSvrAddr:    downloadSvrAddr,
				CallbackSvrAddr:    callbackSvrAddr,
				NodeRole:           persistentVars.NodeRole,
				Generation:         persistentVars.Generation,
				DeployToken:        deployToken,
				PkgVersion:         nodeVersion,
				PkgSavedPath:       pkgPath,
				ConfigSavedDir:     persistentVars.ConfigDir,
				CheckListSavedPath: preCheckListConf,
			}).Run(cmd.Context()); err != nil {
				return err
			}

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

			// do precheck.
			// nolint: gosec
			content, err := os.ReadFile(preCheckListConf)
			if err != nil {
				return err
			}
			var checkListParam precheck.CheckList
			if err := json.Unmarshal(content, &checkListParam); err != nil {
				return err
			}
			if err := precheck.NewStep(precheck.StepArgs{
				AgentHandler: agentHandler,
				CheckList:    checkListParam,
			}).Run(cmd.Context()); err != nil {
				return err
			}

			// install node.
			installResult, err := nodeinstaller.NewStep(nodeinstaller.StepArgs{
				AgentHandler:    agentHandler,
				AgentID:         agentID,
				ReRegisterAgent: agentID == "",
				PkgPath:         pkgPath,
				SrcConfigDir:    persistentVars.ConfigDir,
			}).Run(cmd.Context())
			if err != nil {
				return err
			}

			// Start node.
			if err := nodestarter.NewStep(nodestarter.StepArgs{
				AgentHandler: agentHandler,
			}).Run(cmd.Context()); err != nil {
				return err
			}

			// check deploy.
			if err := checkdeploy.NewStep(checkdeploy.StepArgs{
				AgentHandler: agentHandler,
			}).Run(cmd.Context()); err != nil {
				return err
			}

			// report data.
			if err := datareporter.NewStep(datareporter.StepArgs{
				CallbackSvrAddr: callbackSvrAddr,
				Token:           deployToken,
				OperInstID:      operInstID,
				AgentID:         installResult.AgentID,
			}).Run(cmd.Context()); err != nil {
				return err
			}

			return nil
		},
	}

	/*
	 * required flags.
	 */
	fullCmd.Flags().StringVar(&downloadSvrAddr, flag2.DownloadSvrAddr, "", "download server address, for downloading release files and reporting status")
	_ = fullCmd.MarkFlagRequired(flag2.DownloadSvrAddr)

	fullCmd.Flags().StringVar(&callbackSvrAddr, flag2.CallbackSvrAddr, "", "callback server address, for downloading config files")
	_ = fullCmd.MarkFlagRequired(flag2.CallbackSvrAddr)

	fullCmd.Flags().StringVar(&deployToken, flag2.DeployToken, "", "deploy token, contains the details of files")
	_ = fullCmd.MarkFlagRequired(flag2.DeployToken)

	fullCmd.Flags().StringVar(&nodeVersion, flag2.NodeVersion, "", "node version, for downloading package version")
	_ = fullCmd.MarkFlagRequired(flag2.NodeVersion)

	fullCmd.Flags().StringVar(&operInstID, flag2.OperInstID, "", "operation instance id")
	_ = fullCmd.MarkFlagRequired(flag2.OperInstID)

	/*
	 * optional flags.
	 */
	fullCmd.Flags().StringVar(&logDir, flag2.LogDir, "", "directory to save log files")
	fullCmd.Flags().StringVar(&agentID, flag2.AgentID, "", "existing agent-id to install with, if not given, will register a new one")

	return fullCmd
}
