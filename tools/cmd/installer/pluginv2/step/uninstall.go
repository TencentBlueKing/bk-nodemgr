/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package step

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/pluginv2/handler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/pluginv2/persistent"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/pluginv2/pluginuninstaller"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/pluginv2handler"

	"github.com/spf13/cobra"
)

// NewUninstall creates a new uninstall step command.
func NewUninstall() *cobra.Command {
	var (
		// pre-run.
		pluginHandler pluginv2handler.IPluginHandler
	)

	stepCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall plugin",
		Long:  "Uninstall plugin",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			vars, err := persistent.GetVariables(cmd)
			if err != nil {
				return err
			}

			pluginHandler, err = handler.NewPluginHandler(vars.DeployDir, vars.PluginGroup, vars.PluginName)
			if err != nil {
				return err
			}

			return nil
		},

		RunE: func(cmd *cobra.Command, _ []string) error {
			step := pluginuninstaller.NewStep(pluginuninstaller.StepArgs{
				PluginHandler: pluginHandler,
			})

			if err := step.Run(cmd.Context()); err != nil {
				return err
			}

			fmt.Println("successfully uninstalled")

			return nil
		},
	}

	return stepCmd
}
