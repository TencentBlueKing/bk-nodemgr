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

	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/flag"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/handler"
	logger2 "github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/step"
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

// NewFullUpgrade creates a new full upgrade command.
// nolint: lll, funlen, gocognit
func NewFullUpgrade() *cobra.Command {
	var (
		// required flags.
		callbackSvrAddr string
		deployToken     string
		operInstID      string
		nodeVersion     string

		// optional flags.
		downloadSvrAddr string
		logDir          string
		restart         bool
		force           bool
		skipDownload    bool

		// pre-run.
		persistentVars   *persistent.Variables
		preCheckListConf string
		pkgPath          string
		agentHandler     agenthandler.IAgentHandler
	)
	fullCmd := &cobra.Command{
		Use:   "full-upgrade",
		Short: "Full upgrade process",
		Long:  "Full upgrade process",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			if downloadSvrAddr == "" && !skipDownload {
				return fmt.Errorf("%s is required when %s is not set", flag.DownloadSvrAddr, flag.SkipDownload)
			}

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
			lHandler := logger2.NewHandler(logDir, deployToken, operInstID, callbackSvrAddr)
			if err := lHandler.Start(); err != nil {
				return fmt.Errorf("failed to init logger: %w", err)
			}
			defer lHandler.Stop()

			// download files.
			if !skipDownload {
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
			}

			// upgrade node.
			if err := nodeupgrader.NewStep(nodeupgrader.StepArgs{
				AgentHandler: agentHandler,
				PkgPath:      pkgPath,
				SrcConfigDir: persistentVars.ConfigDir,
				Backup:       true,
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

	fullCmd.Flags().StringVar(&nodeVersion, flag.NodeVersion, "", "node version, for downloading package version")
	_ = fullCmd.MarkFlagRequired(flag.NodeVersion)

	fullCmd.Flags().StringVar(&operInstID, flag.OperInstID, "", "operation instance id")
	_ = fullCmd.MarkFlagRequired(flag.OperInstID)

	/*
	 * optional flags.
	 */
	fullCmd.Flags().StringVar(&downloadSvrAddr, flag.DownloadSvrAddr, "", "download server address, for downloading release files. if skip_download is set, this can be empty")
	fullCmd.Flags().StringVar(&logDir, flag.LogDir, "", "directory to save log files")
	fullCmd.Flags().BoolVar(&restart, flag.Restart, false, "whether to restart node after upgrade")
	fullCmd.Flags().BoolVar(&force, flag.Force, false, "whether to force restart when --restart is set")
	fullCmd.Flags().BoolVar(&skipDownload, flag.SkipDownload, false, "whether to skip downloading files")

	return fullCmd
}
