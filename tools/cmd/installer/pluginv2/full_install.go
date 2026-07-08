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
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/pluginv2/handler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/pluginv2/persistent"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/pluginv2/step"
	pluginv2Installer "github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/pluginv2"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/pluginv2/configfetcher"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/pluginv2/datareporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/pluginv2/filedownloader"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/pluginv2/plugininstaller"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/pluginv2/pluginuninstaller"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/pluginv2/statusreporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/systeminfo"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/logreporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/pluginv2handler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
	"github.com/spf13/cobra"
)

// NewFullInstall creates a new full install command.
// nolint: lll, funlen, gocognit, gocyclo
func NewFullInstall() *cobra.Command {
	var (
		// required flags.
		downloadSvrAddr string
		callbackSvrAddr string
		deployToken     string
		operInstID      string
		pluginVersion   string

		// optional flags.
		logDir       string
		logToStd     bool
		skipCallback bool
		skipDownload bool

		// pre-run.
		persistentVars *persistent.Variables
		pkgPath        string
		pluginHandler  pluginv2handler.IPluginHandler
	)

	fullCmd := &cobra.Command{
		Use:   "full-install",
		Short: "Full install V2 plugin",
		Long:  "Full install V2 plugin",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			if downloadSvrAddr == "" && !skipDownload {
				return fmt.Errorf("%s is required when %s is not set", pluginFlag.DownloadSvrAddr, pluginFlag.SkipDownload)
			}

			if callbackSvrAddr == "" && !skipCallback {
				return fmt.Errorf("%s is required when %s is not set", pluginFlag.CallbackSvrAddr, pluginFlag.SkipCallback)
			}

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

			// download files.
			if !skipDownload {
				downloadSvrAddrs := utils.SplitServerAddrs(downloadSvrAddr)
				if len(downloadSvrAddrs) == 0 {
					return fmt.Errorf("download server address is empty or invalid")
				}
				if err := filedownloader.NewStep(filedownloader.StepArgs{
					DownloadSvrAddr: downloadSvrAddrs,
					PluginGroup:     persistentVars.PluginGroup,
					PluginName:      persistentVars.PluginName,
					PluginPkgName:   persistentVars.PluginPkgName,
					DeployToken:     deployToken,
					PkgVersion:      pluginVersion,
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
					CallbackSvrAddr: callbackSvrAddrs,
					PluginGroup:     persistentVars.PluginGroup,
					PluginName:      persistentVars.PluginName,
					PluginPkgName:   persistentVars.PluginPkgName,
					DeployToken:     deployToken,
					ConfigSavedDir:  persistentVars.ConfigDir,
				}).Run(cmd.Context()); err != nil {
					return err
				}
			}

			// uninstall plugin.
			if err := pluginuninstaller.NewStep(pluginuninstaller.StepArgs{
				PluginHandler: pluginHandler,
			}).Run(cmd.Context()); err != nil {
				return err
			}

			// install plugin.
			if err := plugininstaller.NewStep(plugininstaller.StepArgs{
				PluginHandler: pluginHandler,
				PkgPath:       pkgPath,
				SrcConfigDir:  persistentVars.ConfigDir,
			}).Run(cmd.Context()); err != nil {
				return err
			}

			dataFilePath := filepath.Join(persistentVars.DataDir, "installer.data.json")
			if err := datareporter.NewStep(datareporter.StepArgs{
				CallbackSvrAddr: callbackSvrAddrs,
				Token:           deployToken,
				OperInstID:      operInstID,
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
	fullCmd.Flags().StringVar(&downloadSvrAddr, pluginFlag.DownloadSvrAddr, "", "download server address. if skip_download is set, this can be empty")

	fullCmd.Flags().StringVar(&callbackSvrAddr, pluginFlag.CallbackSvrAddr, "", "callback server address. if skip_callback is set, this can be empty")

	fullCmd.Flags().StringVar(&deployToken, pluginFlag.DeployToken, "", "deploy token, contains the details of files")
	_ = fullCmd.MarkFlagRequired(pluginFlag.DeployToken)

	fullCmd.Flags().StringVar(&operInstID, pluginFlag.OperInstID, "", "operation instance id")
	_ = fullCmd.MarkFlagRequired(pluginFlag.OperInstID)

	fullCmd.Flags().StringVar(&pluginVersion, pluginFlag.PluginVersion, "", "plugin version, for downloading package version")
	_ = fullCmd.MarkFlagRequired(pluginFlag.PluginVersion)

	/*
	 * optional flags.
	 */
	fullCmd.Flags().StringVar(&logDir, pluginFlag.LogDir, "", "directory to save log files")
	fullCmd.Flags().BoolVar(&logToStd, pluginFlag.LogToStd, false, "also output log to stdout")
	fullCmd.Flags().BoolVar(&skipDownload, pluginFlag.SkipDownload, false, "whether to skip downloading files")
	fullCmd.Flags().BoolVar(&skipCallback, pluginFlag.SkipCallback, false, "whether to skip callback reporting (write results to local files instead)")

	return fullCmd
}

func errString(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}
