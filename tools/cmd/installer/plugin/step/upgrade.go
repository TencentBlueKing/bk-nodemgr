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
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/plugin/handler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/plugin/persistent"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/plugin/pluginupgrader"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/pluginhandler"
	"github.com/spf13/cobra"
)

// NewUpgrade creates a new upgrade step command.
func NewUpgrade() *cobra.Command {
	var (
		// pre-run.
		pluginHandler pluginhandler.IPluginHandler
	)

	stepCmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade plugin",
		Long:  "Upgrade plugin",
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
			step := pluginupgrader.NewStep(pluginupgrader.StepArgs{
				PluginHandler: pluginHandler,
			})

			if err := step.Run(cmd.Context()); err != nil {
				return err
			}

			fmt.Println("successfully upgraded")

			return nil
		},
	}

	return stepCmd
}
