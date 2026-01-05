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
	pluginflag "github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/plugin/flag"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/plugin/handler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/plugin/persistent"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/plugin/datareporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/plugin/pluginuninstaller"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/plugin/statusreporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/logreporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/pluginhandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/types"
	"github.com/spf13/cobra"
	"path/filepath"
)

// NewFullUninstall creates a new full uninstall command.
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
		pluginHandler  pluginhandler.IPluginHandler
	)

	fullCmd := &cobra.Command{
		Use:   "full-uninstall",
		Short: "Full uninstall plugin",
		Long:  "Full uninstall plugin",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
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

			// uninstall plugin.
			if err := pluginuninstaller.NewStep(pluginuninstaller.StepArgs{
				PluginHandler: pluginHandler,
			}).Run(cmd.Context()); err != nil {
				return err
			}

			// report data.
			if err := datareporter.NewStep(datareporter.StepArgs{
				CallbackSvrAddr: callbackSvrAddr,
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
	fullCmd.Flags().StringVar(&callbackSvrAddr, pluginflag.CallbackSvrAddr, "", "callback server address, for downloading config files")
	_ = fullCmd.MarkFlagRequired(pluginflag.CallbackSvrAddr)

	fullCmd.Flags().StringVar(&deployToken, pluginflag.DeployToken, "", "deploy token, contains the details of files")
	_ = fullCmd.MarkFlagRequired(pluginflag.DeployToken)

	fullCmd.Flags().StringVar(&operInstID, pluginflag.OperInstID, "", "operation instance id")
	_ = fullCmd.MarkFlagRequired(pluginflag.OperInstID)

	/*
	 * optional flags.
	 */
	fullCmd.Flags().StringVar(&logDir, pluginflag.LogDir, "", "directory to save log files")
	fullCmd.Flags().BoolVar(&logToStd, pluginflag.LogToStd, false, "also output log to stdout")

	return fullCmd
}
