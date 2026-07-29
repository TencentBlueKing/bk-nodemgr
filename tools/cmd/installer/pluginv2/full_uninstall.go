/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pluginv2

import (
	"fmt"
	"path/filepath"

	pluginFlag "github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/pluginv2/flag"
	pluginV2Flag "github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/pluginv2/flag"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/pluginv2/handler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/pluginv2/persistent"
	pluginv2Installer "github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/pluginv2"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/pluginv2/datareporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/pluginv2/pluginuninstaller"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/pluginv2/statusreporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/systeminfo"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/logreporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/pluginv2handler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
	"github.com/spf13/cobra"
)

// NewFullUninstall creates a new full uninstall V2 command.
func NewFullUninstall() *cobra.Command {
	var (
		// required flags.
		callbackSvrAddr string
		deployToken     string
		operInstID      string

		// optional flags.
		logDir       string
		logToStd     bool
		skipCallback bool

		// pre-run.
		persistentVars *persistent.Variables
		pluginHandler  pluginv2handler.IPluginHandler
	)

	fullCmd := &cobra.Command{
		Use:   "full-uninstall",
		Short: "Full uninstall V2 plugin",
		Long:  "Full uninstall V2 plugin",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			if callbackSvrAddr == "" && !skipCallback {
				return fmt.Errorf("%s is required when %s is not set", pluginFlag.CallbackSvrAddr, pluginFlag.SkipCallback)
			}

			vars, err := persistent.GetVariables(cmd)
			if err != nil {
				return err
			}
			persistentVars = vars

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

			systeminfo.LogInitialTargetInfo(pluginv2Installer.StepGeneral)

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

			if err := pluginuninstaller.NewStep(pluginuninstaller.StepArgs{
				PluginHandler: pluginHandler,
			}).Run(cmd.Context()); err != nil {
				return err
			}

			// report data.
			if err := datareporter.NewStep(datareporter.StepArgs{
				CallbackSvrAddr: callbackSvrAddrs,
				Token:           deployToken,
			}).Run(cmd.Context()); err != nil {
				return err
			}

			return nil
		},
	}

	/*
	 * required flags.
	 */
	fullCmd.Flags().StringVar(&callbackSvrAddr, pluginV2Flag.CallbackSvrAddr, "", "callback server address. if skip_callback is set, this can be empty")

	fullCmd.Flags().StringVar(&deployToken, pluginV2Flag.DeployToken, "", "deploy token, contains the details of files")
	_ = fullCmd.MarkFlagRequired(pluginV2Flag.DeployToken)

	fullCmd.Flags().StringVar(&operInstID, pluginV2Flag.OperInstID, "", "operation instance id")
	_ = fullCmd.MarkFlagRequired(pluginV2Flag.OperInstID)

	/*
	 * optional flags.
	 */
	fullCmd.Flags().StringVar(&logDir, pluginV2Flag.LogDir, "", "directory to save log files")
	fullCmd.Flags().BoolVar(&logToStd, pluginV2Flag.LogToStd, false, "also output log to stdout")
	fullCmd.Flags().BoolVar(&skipCallback, pluginV2Flag.SkipCallback, false, "whether to skip callback reporting (write results to local files instead)")

	return fullCmd
}
