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

	pluginFlag "github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/plugin/flag"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/plugin/datareporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
	"github.com/spf13/cobra"
)

// NewReportData creates a new report data step command.
func NewReportData() *cobra.Command {
	var (
		// required flags.
		callbackSvrAddr string
		deployToken     string
	)

	stepCmd := &cobra.Command{
		Use:   "report-data",
		Short: "Report data",
		RunE: func(cmd *cobra.Command, _ []string) error {
			callbackSvrAddrs := utils.SplitServerAddrs(callbackSvrAddr)
			if len(callbackSvrAddrs) == 0 {
				return fmt.Errorf("callback server address is empty or invalid")
			}

			step := datareporter.NewStep(datareporter.StepArgs{
				CallbackSvrAddr: callbackSvrAddrs,
				Token:           deployToken,
			})

			if err := step.Run(cmd.Context()); err != nil {
				return err
			}

			fmt.Println("successfully reported data")

			return nil
		},
	}

	/*
	 * required flags.
	 */
	stepCmd.Flags().StringVar(&callbackSvrAddr, pluginFlag.CallbackSvrAddr, "", "callback server address, for reporting data")
	_ = stepCmd.MarkFlagRequired(pluginFlag.CallbackSvrAddr)

	stepCmd.Flags().StringVar(&deployToken, pluginFlag.DeployToken, "", "deploy token, contains the details of this process")
	_ = stepCmd.MarkFlagRequired(pluginFlag.DeployToken)

	return stepCmd
}
