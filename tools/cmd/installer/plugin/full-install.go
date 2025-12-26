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

	pluginflag "github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/plugin/flag"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/plugin/handler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/plugin/persistent"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/plugin/step"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/plugin/datareporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/plugin/filedownloader"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/plugin/plugininstaller"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/plugin/pluginuninstaller"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/plugin/statusreporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/logreporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/pluginhandler"
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
		pluginVersion   string

		// optional flags.
		logDir   string
		logToStd bool

		// pre-run.
		persistentVars *persistent.Variables
		pkgPath        string
		pluginHandler  pluginhandler.IPluginHandler
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

			pkgPath = filepath.Join(persistentVars.DataDir, step.GenReleasePkgName(persistentVars.PluginName, pluginVersion))

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
			lHandler := logreporter.NewHandler(logDir, logToStd, deployToken, operInstID, reportLogUrl(callbackSvrAddr))
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
				DownloadSvrAddr:              downloadSvrAddr,
				CallbackSvrAddr:              callbackSvrAddr,
				PluginGroup:                  persistentVars.PluginGroup,
				PluginName:                   persistentVars.PluginName,
				PluginPkgName:         persistentVars.PluginPkgName,
				DeployToken:                  deployToken,
				PkgVersion:                   pluginVersion,
				PkgSavedPath:                 pkgPath,
				ConfigSavedDir:               persistentVars.ConfigDir,
				SelectDownloads:              false,
				EnableDownloadConfig:         false,
				EnableDownloadReleasePackage: false,
			}).Run(cmd.Context()); err != nil {
				return err
			}

			// uninstall plugin.
			if err := pluginuninstaller.NewStep(pluginuninstaller.StepArgs{
				PluginHandler: pluginHandler,
			}).Run(cmd.Context()); err != nil {
				return err
			}

			// install plugin.
			_, err := plugininstaller.NewStep(plugininstaller.StepArgs{
				PluginHandler: pluginHandler,
				PkgPath:       pkgPath,
				SrcConfigDir:  persistentVars.ConfigDir,
			}).Run(cmd.Context())
			if err != nil {
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
	fullCmd.Flags().StringVar(&downloadSvrAddr, pluginflag.DownloadSvrAddr, "", "download server address, for downloading release files and reporting status")
	_ = fullCmd.MarkFlagRequired(pluginflag.DownloadSvrAddr)

	fullCmd.Flags().StringVar(&callbackSvrAddr, pluginflag.CallbackSvrAddr, "", "callback server address, for downloading config files")
	_ = fullCmd.MarkFlagRequired(pluginflag.CallbackSvrAddr)

	fullCmd.Flags().StringVar(&deployToken, pluginflag.DeployToken, "", "deploy token, contains the details of files")
	_ = fullCmd.MarkFlagRequired(pluginflag.DeployToken)

	fullCmd.Flags().StringVar(&operInstID, pluginflag.OperInstID, "", "operation instance id")
	_ = fullCmd.MarkFlagRequired(pluginflag.OperInstID)

	fullCmd.Flags().StringVar(&pluginVersion, pluginflag.PluginVersion, "", "plugin version, for downloading package version")
	_ = fullCmd.MarkFlagRequired(pluginflag.PluginVersion)

	/*
	 * optional flags.
	 */
	fullCmd.Flags().StringVar(&logDir, pluginflag.LogDir, "", "directory to save log files")
	fullCmd.Flags().BoolVar(&logToStd, pluginflag.LogToStd, false, "also output log to stdout")

	return fullCmd
}
