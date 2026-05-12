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

	pluginV2Flag "github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/pluginv2/flag"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/pluginv2/handler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/pluginv2/persistent"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/pluginv2/plugininstaller"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/pluginv2handler"
	"github.com/spf13/cobra"
)

// NewInstall creates a new install step command.
// nolint: lll
func NewInstall() *cobra.Command {
	var (
		// required flags.
		pkgPath string

		// pre-run.
		pluginHandler  pluginv2handler.IPluginHandler
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

			if err := step.Run(cmd.Context()); err != nil {
				return err
			}

			fmt.Printf("successfully installed.\n")

			return nil
		},
	}

	/*
	 * required flags.
	 */
	stepCmd.Flags().StringVar(&pkgPath, pluginV2Flag.PkgFile, "", "path to release package file to install")
	_ = stepCmd.MarkFlagRequired(pluginV2Flag.PkgFile)

	return stepCmd
}
