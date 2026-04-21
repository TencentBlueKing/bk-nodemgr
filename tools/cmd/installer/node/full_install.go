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

	nodeFlag "github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/flag"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/handler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/persistent"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/step"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/checkdeploy"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/configfetcher"
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
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
	"github.com/spf13/cobra"
)

// NewFullInstall creates a new full install command.
// nolint: lll, funlen, gocognit, gocyclo, cyclop, maintidx
func NewFullInstall() *cobra.Command {
	var (
		// required flags.
		downloadSvrAddr string
		callbackSvrAddr string
		deployToken     string
		operInstID      string
		nodeVersion     string

		// optional flags.
		logDir       string
		logToStd     bool
		agentID      string
		skipDownload bool
		skipCallback bool

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
			if downloadSvrAddr == "" && !skipDownload {
				return fmt.Errorf("%s is required when %s is not set", nodeFlag.DownloadSvrAddr, nodeFlag.SkipDownload)
			}

			if callbackSvrAddr == "" && !skipCallback {
				return fmt.Errorf("%s is required when %s is not set", nodeFlag.CallbackSvrAddr, nodeFlag.SkipCallback)
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
			CleanOldReleasePackages(persistentVars.DataDir, pkgPath)

			var callbackSvrAddrs []string
			var logURLs []string
			if !skipCallback {
				callbackSvrAddrs = utils.SplitServerAddrs(callbackSvrAddr)
				if len(callbackSvrAddrs) == 0 {
					return fmt.Errorf("callback server address is empty or invalid")
				}
				var err error
				logURLs, err = reportLogURLs(callbackSvrAddrs)
				if err != nil {
					return fmt.Errorf("failed to build log report URLs: %w", err)
				}
			}

			lHandler := logreporter.NewHandler(logDir, logToStd, deployToken, operInstID, logURLs)
			if err := lHandler.Start(); err != nil {
				return fmt.Errorf("failed to init logger: %w", err)
			}
			defer lHandler.Stop()

			// SYNC: file name must match installer.StatusFileName in pkg/installer/constant.go.
			statusFilePath := filepath.Join(persistentVars.DataDir, "installer.status.json")
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
					SkipCallback:    skipCallback,
					StatusFilePath:  statusFilePath,
					ErrorMessage:    errString(runErr),
				}).Run(cmd.Context())
			}()

			// download release package.
			if !skipDownload {
				downloadSvrAddrs := utils.SplitServerAddrs(downloadSvrAddr)
				if len(downloadSvrAddrs) == 0 {
					return fmt.Errorf("download server address is empty or invalid")
				}

				if err := filedownloader.NewStep(filedownloader.StepArgs{
					DownloadSvrAddr: downloadSvrAddrs,
					NodeRole:        persistentVars.NodeRole,
					Generation:      persistentVars.Generation,
					DeployToken:     deployToken,
					PkgVersion:      nodeVersion,
					PkgSavedPath:    pkgPath,
				}).Run(cmd.Context()); err != nil {
					return err
				}
			}

			// fetch configs.
			if !skipCallback {
				callbackSvrAddrs = utils.SplitServerAddrs(callbackSvrAddr)
				if len(callbackSvrAddrs) == 0 {
					return fmt.Errorf("callback server address is empty or invalid")
				}

				if err := configfetcher.NewStep(configfetcher.StepArgs{
					CallbackSvrAddr:    callbackSvrAddrs,
					NodeRole:           persistentVars.NodeRole,
					Generation:         persistentVars.Generation,
					DeployToken:        deployToken,
					ConfigSavedDir:     persistentVars.ConfigDir,
					CheckListSavedPath: preCheckListConf,
				}).Run(cmd.Context()); err != nil {
					return err
				}
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

			// SYNC: file name must match installer.DataFileName in pkg/installer/constant.go.
			dataFilePath := filepath.Join(persistentVars.DataDir, "installer.data.json")
			if err := datareporter.NewStep(datareporter.StepArgs{
				CallbackSvrAddr: callbackSvrAddrs,
				Token:           deployToken,
				OperInstID:      operInstID,
				AgentID:         installResult.AgentID,
				SkipCallback:    skipCallback,
				DataFilePath:    dataFilePath,
			}).Run(cmd.Context()); err != nil {
				return err
			}

			return nil
		},
	}

	/*
	 * required flags.
	 */
	fullCmd.Flags().StringVar(&downloadSvrAddr, nodeFlag.DownloadSvrAddr, "", "download server address. if skip_download is set, this can be empty")
	fullCmd.Flags().StringVar(&callbackSvrAddr, nodeFlag.CallbackSvrAddr, "", "callback server address. if skip_callback is set, this can be empty")
	fullCmd.Flags().StringVar(&deployToken, nodeFlag.DeployToken, "", "deploy token, contains the details of files")
	_ = fullCmd.MarkFlagRequired(nodeFlag.DeployToken)

	fullCmd.Flags().StringVar(&nodeVersion, nodeFlag.NodeVersion, "", "node version, for downloading package version")
	_ = fullCmd.MarkFlagRequired(nodeFlag.NodeVersion)

	fullCmd.Flags().StringVar(&operInstID, nodeFlag.OperInstID, "", "operation instance id")
	_ = fullCmd.MarkFlagRequired(nodeFlag.OperInstID)

	/*
	 * optional flags.
	 */
	fullCmd.Flags().StringVar(&logDir, nodeFlag.LogDir, "", "directory to save log files")
	fullCmd.Flags().BoolVar(&logToStd, nodeFlag.LogToStd, false, "also output log to stdout")
	fullCmd.Flags().StringVar(&agentID, nodeFlag.AgentID, "", "existing agent-id to install with, if not given, will register a new one")
	fullCmd.Flags().BoolVar(&skipDownload, nodeFlag.SkipDownload, false, "whether to skip downloading files")
	fullCmd.Flags().BoolVar(&skipCallback, nodeFlag.SkipCallback, false, "whether to skip callback reporting (write results to local files instead)")

	return fullCmd
}

func errString(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}
