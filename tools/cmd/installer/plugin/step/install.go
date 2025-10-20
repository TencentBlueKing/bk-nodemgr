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

	pluginflag "github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/plugin/flag"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/plugin/handler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/plugin/persistent"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/plugin/plugininstaller"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/pluginhandler"
	"github.com/spf13/cobra"
)

// NewInstall creates a new install step command.
// nolint: lll
func NewInstall() *cobra.Command {
	var (
		// required flags.
		pkgPath string

		// pre-run.
		pluginHandler  pluginhandler.IPluginHandler
		persistentVars *persistent.Variables
	)

	stepCmd := &cobra.Command{
		Use:   "install",
		Short: "Install plugin",
		Long:  "Install plugin",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			vars, err := persistent.GetVariables(cmd)
			if err != nil {
				return err
			}
			persistentVars = vars

			pluginHandler, err = handler.NewPluginHandler(vars.DeployDir, "", vars.PluginName)
			if err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			step := plugininstaller.NewStep(plugininstaller.StepArgs{
				PluginHandler: pluginHandler,
				PkgPath:       pkgPath,
				SrcConfigDir:  persistentVars.ConfigDir,
			})

			stepResult, err := step.Run(cmd.Context())
			if err != nil {
				return err
			}

			fmt.Printf("successfully installed, step-result(%s).\n", stepResult)

			return nil
		},
	}

	/*
	 * required flags.
	 */
	stepCmd.Flags().StringVar(&pkgPath, pluginflag.PkgFile, "", "path to release package file to install")
	_ = stepCmd.MarkFlagRequired(pluginflag.PkgFile)

	return stepCmd
}
