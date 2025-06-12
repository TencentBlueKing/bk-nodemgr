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
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/statusreporter"
	"github.com/spf13/cobra"
)

// NewStepReportStatus new a step for report status.
func NewStepReportStatus() *cobra.Command {
	var (
		callBackEndPoint string
		token            string
		status           string
	)

	stepCmd := &cobra.Command{
		Use:   "step_reportstatus",
		Short: "report status",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if err := SetToken(token); err != nil {
				return fmt.Errorf("set token failed, err: %v", err)
			}

			if err := SetCallbackEndPoint(callBackEndPoint); err != nil {
				return fmt.Errorf("set callback endpoint failed, err: %v", err)
			}

			if status == "" {
				return errors.New("status cannot be empty")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			step := statusreporter.NewStep(statusreporter.StepArgs{
				Token:            GetToken(),
				Status:           constant.State(status),
				CallbackEndpoint: GetCallBackEndpoint(),
			})

			err := step.Run(cmd.Context())
			if err != nil {
				return err
			}

			return nil
		},
	}

	stepCmd.Flags().StringVar(&callBackEndPoint, CmdFlagCallbackEndpoint, "", "callback endpoint")
	stepCmd.Flags().StringVar(&token, CmdFlagToken, "", "token")
	stepCmd.Flags().StringVar(&status, CmdFlagAgentID, "", "status")

	return stepCmd
}
