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

	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/flag"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/persistent"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/configfetcher"
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
				CallbackSvrAddr:    callbackSvrAddrs,
				NodeRole:           persistentVars.NodeRole,
				Generation:         persistentVars.Generation,
				DeployToken:        deployToken,
				ConfigSavedDir:     persistentVars.ConfigDir,
				CheckListSavedPath: filepath.Join(persistentVars.DataDir, "precheck.json"),
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
	stepCmd.Flags().StringVar(&callbackSvrAddr, flag.CallbackSvrAddr, "", "callback server address, for fetch config files")
	_ = stepCmd.MarkFlagRequired(flag.CallbackSvrAddr)

	stepCmd.Flags().StringVar(&deployToken, flag.DeployToken, "", "deploy token, contains the details of files")
	_ = stepCmd.MarkFlagRequired(flag.DeployToken)

	return stepCmd
}
