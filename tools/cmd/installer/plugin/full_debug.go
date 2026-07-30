/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package plugin

import (
	"fmt"

	"path/filepath"

	pluginFlag "github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/plugin/flag"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/plugin/handler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/plugin/persistent"
	pluginInstaller "github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/plugin/pluginrunner"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/plugin/statusreporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/systeminfo"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/logreporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/pluginhandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
	"github.com/spf13/cobra"
)

// NewFullDebug creates a new full-debug command.
// nolint: lll
func NewFullDebug() *cobra.Command {
	var (
		// required flags.
		callbackSvrAddr string
		deployToken     string
		operInstID      string
		runCmd          string
		pidDir          string
		debugAction     string

		// optional flags.
		logDir       string
		logToStd     bool
		skipCallback bool

		// pre-run.
		persistentVars *persistent.Variables
		pluginHandler  pluginhandler.IPluginHandler
	)

	debugCmd := &cobra.Command{
		Use:   "full-debug",
		Short: "full debug plugin",
		Long:  "full debug plugin by executing a user-supplied command inside the plugin deploy directory",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			if callbackSvrAddr == "" && !skipCallback {
				return fmt.Errorf("%s is required when %s is not set", pluginFlag.CallbackSvrAddr, pluginFlag.SkipCallback)
			}

			vars, err := persistent.GetVariables(cmd)
			if err != nil {
				return err
			}
			persistentVars = vars

			if err := persistentVars.EnsureDirs(); err != nil {
				return err
			}

			if logDir == "" {
				logDir = filepath.Join(persistentVars.DataDir, "logs")
			}

			pluginHandler, err = handler.NewPluginHandler(vars.DeployDir, vars.PluginGroup, vars.PluginName)
			if err != nil {
				return err
			}

			return nil
		},
		// nolint: nonamedreturns
		RunE: func(cmd *cobra.Command, _ []string) (runErr error) {
			// init log settings.
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

			action := types.PluginDebugAction(debugAction)
			if err := action.Validate(); err != nil {
				return fmt.Errorf("failed to validate debug action: %w", err)
			}

			lHandler := logreporter.NewHandler(logDir, logToStd, deployToken, operInstID, logURLs)
			if err := lHandler.Start(); err != nil {
				return fmt.Errorf("failed to init logger: %w", err)
			}
			defer lHandler.Stop()

			systeminfo.LogInitialTargetInfo(pluginInstaller.StepGeneral)

			statusFilePath := filepath.Join(persistentVars.DataDir, "installer.status.json")

			// report status on exit.
			defer func() {
				state := types.ProcessStateSuccess
				if runErr != nil {
					state = types.ProcessStateFailed
				}

				reportErr := statusreporter.NewStep(statusreporter.StepArgs{
					Token:           deployToken,
					OperInstID:      operInstID,
					Status:          state,
					CallbackSvrAddr: callbackSvrAddrs,
					SkipCallback:    skipCallback,
					StatusFilePath:  statusFilePath,
					ErrorMessage:    errString(runErr),
				}).Run(cmd.Context())
				if reportErr != nil && runErr == nil && skipCallback {
					runErr = fmt.Errorf("failed to write debug status file: %w", reportErr)
				}
			}()

			// run the user-supplied command inside the deploy directory.
			return pluginrunner.NewStep(pluginrunner.StepArgs{
				PluginHandler: pluginHandler,
				DebugAction:   action,
				RunCmd:        runCmd,
				PidDir:        pidDir,
			}).Run(cmd.Context())
		},
	}

	/*
	 * required flags.
	 */
	debugCmd.Flags().StringVar(&deployToken, pluginFlag.DeployToken, "", "deploy token, contains the details of files")
	_ = debugCmd.MarkFlagRequired(pluginFlag.DeployToken)

	debugCmd.Flags().StringVar(&operInstID, pluginFlag.OperInstID, "", "operation instance id")
	_ = debugCmd.MarkFlagRequired(pluginFlag.OperInstID)

	debugCmd.Flags().StringVar(&debugAction, pluginFlag.DebugAction, "", "debug action, e.g. start or stop")
	_ = debugCmd.MarkFlagRequired(pluginFlag.DebugAction)

	/*
	 * optional flags.
	 */
	debugCmd.Flags().StringVar(&callbackSvrAddr, pluginFlag.CallbackSvrAddr, "",
		"callback server address, for reporting status and logs. if skip_callback is set, this can be empty")
	debugCmd.Flags().BoolVar(&skipCallback, pluginFlag.SkipCallback, false, "whether to skip callback reporting (write status to local file instead)")
	debugCmd.Flags().StringVar(&runCmd, pluginFlag.RunCmd, "", "command to execute inside the plugin deploy directory")
	debugCmd.Flags().StringVar(&pidDir, pluginFlag.PidDir, "", "directory to save debug process pid file")
	debugCmd.Flags().StringVar(&logDir, pluginFlag.LogDir, "", "directory to save log files")
	debugCmd.Flags().BoolVar(&logToStd, pluginFlag.LogToStd, false, "also output log to stdout")

	return debugCmd
}
