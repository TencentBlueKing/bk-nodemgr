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
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/plugin/statusreporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
	"github.com/spf13/cobra"
)

// NewReportStatus creates a new report status step command.
func NewReportStatus() *cobra.Command {
	var (
		// required flags.
		callbackSvrAddr string
		deployToken     string
		status          string
		operInstID      string
	)

	stepCmd := &cobra.Command{
		Use:   "report-status",
		Short: "Report status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			callbackSvrAddrs := utils.SplitServerAddrs(callbackSvrAddr)
			if len(callbackSvrAddrs) == 0 {
				return fmt.Errorf("callback server address is empty or invalid")
			}

			step := statusreporter.NewStep(statusreporter.StepArgs{
				CallbackSvrAddr: callbackSvrAddrs,
				Token:           deployToken,
				Status:          types.ProcessState(status),
				OperInstID:      operInstID,
			})

			if err := step.Run(cmd.Context()); err != nil {
				return err
			}

			fmt.Println("successfully reported status")

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

	stepCmd.Flags().StringVar(&status, pluginFlag.Status, "", "status to report, start, running, success, failed, timeout, skip")
	_ = stepCmd.MarkFlagRequired(pluginFlag.Status)

	stepCmd.Flags().StringVar(&operInstID, pluginFlag.OperInstID, "", "operation instance id")
	_ = stepCmd.MarkFlagRequired(pluginFlag.OperInstID)

	return stepCmd
}
