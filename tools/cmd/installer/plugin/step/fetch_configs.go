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

	pluginFlag "github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/plugin/flag"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/plugin/persistent"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/plugin/configfetcher"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
	"github.com/spf13/cobra"
)

// NewFetchConfigs creates a new fetch configs step command.
// nolint: lll
func NewFetchConfigs() *cobra.Command {
	var (
		// required flags.
		callbackSvrAddr string
		deployToken     string

		// pre-run.
		persistentVars *persistent.Variables
	)

	stepCmd := &cobra.Command{
		Use:   "fetch-configs",
		Short: "Fetch configs",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			vars, err := persistent.GetVariables(cmd)
			if err != nil {
				return err
			}
			persistentVars = vars

			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			callbackSvrAddrs := utils.SplitServerAddrs(callbackSvrAddr)
			if len(callbackSvrAddrs) == 0 {
				return fmt.Errorf("callback server address is empty or invalid")
			}

			step := configfetcher.NewStep(configfetcher.StepArgs{
				CallbackSvrAddr: callbackSvrAddrs,
				PluginGroup:     persistentVars.PluginGroup,
				PluginName:      persistentVars.PluginName,
				PluginPkgName:   persistentVars.PluginPkgName,
				DeployToken:     deployToken,
				ConfigSavedDir:  persistentVars.ConfigDir,
			})

			if err := step.Run(cmd.Context()); err != nil {
				return err
			}

			fmt.Println("successfully fetch configs")

			return nil
		},
	}

	/*
	 * required flags.
	 */
	stepCmd.Flags().StringVar(&callbackSvrAddr, pluginFlag.CallbackSvrAddr, "", "callback server address, for downloading config files")
	_ = stepCmd.MarkFlagRequired(pluginFlag.CallbackSvrAddr)

	stepCmd.Flags().StringVar(&deployToken, pluginFlag.DeployToken, "", "deploy token, contains the details of files")
	_ = stepCmd.MarkFlagRequired(pluginFlag.DeployToken)

	return stepCmd
}
