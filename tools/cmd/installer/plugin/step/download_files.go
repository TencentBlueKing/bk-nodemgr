/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package step

import (
	"fmt"
	"path/filepath"
	"runtime"

	pluginflag "github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/plugin/flag"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/plugin/persistent"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/plugin/filedownloader"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
	"github.com/spf13/cobra"
)

// NewDownloadFiles creates a new download files step command
// nolint: lll
func NewDownloadFiles() *cobra.Command {
	var (
		// required flags.
		downloadSvrAddr string
		callbackSvrAddr string
		deployToken     string
		pluginVersion   string

		// pre-run.
		persistentVars *persistent.Variables
	)

	stepCmd := &cobra.Command{
		Use:   "download-files",
		Short: "Download plugin files",
		Long:  "Download plugin files",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			vars, err := persistent.GetVariables(cmd)
			if err != nil {
				return err
			}
			persistentVars = vars

			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			downloadSvrAddrs := utils.SplitServerAddrs(downloadSvrAddr)
			if len(downloadSvrAddrs) == 0 {
				return fmt.Errorf("download server address is empty or invalid")
			}

			callbackSvrAddrs := utils.SplitServerAddrs(callbackSvrAddr)
			if len(callbackSvrAddrs) == 0 {
				return fmt.Errorf("callback server address is empty or invalid")
			}

			step := filedownloader.NewStep(filedownloader.StepArgs{
				DownloadSvrAddr:              downloadSvrAddrs,
				CallbackSvrAddr:              callbackSvrAddrs,
				PluginGroup:                  persistentVars.PluginGroup,
				PluginName:                   persistentVars.PluginName,
				PluginPkgName:                persistentVars.PluginPkgName,
				DeployToken:                  deployToken,
				PkgVersion:                   pluginVersion,
				PkgSavedPath:                 filepath.Join(persistentVars.DataDir, GenReleasePkgName(persistentVars.PluginName, pluginVersion)),
				ConfigSavedDir:               filepath.Join(persistentVars.DataDir, "configs"),
				SelectDownloads:              false,
				EnableDownloadConfig:         false,
				EnableDownloadReleasePackage: false,
			})

			if err := step.Run(cmd.Context()); err != nil {
				return err
			}

			fmt.Println("successfully downloaded files")

			return nil
		},
	}

	/*
	 * required flags.
	 */
	stepCmd.Flags().StringVar(&downloadSvrAddr, pluginflag.DownloadSvrAddr, "", "download server address, for downloading release files")
	_ = stepCmd.MarkFlagRequired(pluginflag.DownloadSvrAddr)

	stepCmd.Flags().StringVar(&callbackSvrAddr, pluginflag.CallbackSvrAddr, "", "callback server address, for downloading config files")
	_ = stepCmd.MarkFlagRequired(pluginflag.CallbackSvrAddr)

	stepCmd.Flags().StringVar(&deployToken, pluginflag.DeployToken, "", "deploy token, contains the details of files")
	_ = stepCmd.MarkFlagRequired(pluginflag.DeployToken)

	stepCmd.Flags().StringVar(&pluginVersion, pluginflag.PluginVersion, "", "plugin version")
	_ = stepCmd.MarkFlagRequired(pluginflag.PluginVersion)

	return stepCmd
}

const pluginReleasePkgExt = "tgz"

// GenReleasePkgName generates release package name.
// match the pattern with pkg/format/pluginpkg/plugin.go
// bk-nodemgr_plugin_{generation}_{pkg-name}-{version}-{os}-{arch}.{ext}
// TODO: generation should be a flag from command.
func GenReleasePkgName(pluginName, version string) string {
	return fmt.Sprintf(
		"bk-nodemgr_plugin_2_%s-%s-%s_%s.%s",
		pluginName,
		version,
		runtime.GOOS,
		runtime.GOARCH,
		pluginReleasePkgExt)
}
