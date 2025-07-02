/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package main ...
package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewStepReportData new a step for report data.
func NewStepReportData() *cobra.Command {
	var (
		callBackEndPoint string
		token            string
		operInstID       string
		agentID          string
	)

	stepCmd := &cobra.Command{
		Use:   "step_reportdata",
		Short: "report data",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if err := SetServerConf(token, operInstID); err != nil {
				return fmt.Errorf("set server conf failed, err: %v", err)
			}

			if err := SetCallbackEndPoint(callBackEndPoint); err != nil {
				return fmt.Errorf("set callback endpoint failed, err: %v", err)
			}

			if err := SetNodeAgentID(agentID); err != nil {
				return fmt.Errorf("set node agent id failed, err: %v", err)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := stepReportData(cmd.Context()); err != nil {
				return err
			}

			fmt.Println("successfully report data")

			return nil
		},
	}

	stepCmd.Flags().StringVar(&callBackEndPoint, CmdFlagCallbackEndpoint, "", "callback endpoint")
	stepCmd.Flags().StringVar(&token, CmdFlagToken, "", "token")
	stepCmd.Flags().StringVar(&token, CmdFlagOperInstID, "", "operation instance id")
	stepCmd.Flags().StringVar(&agentID, CmdFlagAgentID, "", "agent id")

	return stepCmd
}
