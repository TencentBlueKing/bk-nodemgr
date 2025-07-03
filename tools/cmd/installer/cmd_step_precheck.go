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
	"encoding/json"
	"fmt"
	"os"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/precheck"
	"github.com/spf13/cobra"
)

// NewStepPreCheck ...
func NewStepPreCheck() *cobra.Command {
	var (
		preCheckListPath string
	)

	stepCmd := &cobra.Command{
		Use:   "step_precheck",
		Short: "precheck",
		Long:  "precheck",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if err := SetPreCheckFilePath(preCheckListPath); err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := stepPreCheck(cmd.Context()); err != nil {
				return err
			}

			fmt.Println("successfully precheck")

			return nil
		},
	}
	stepCmd.Flags().StringVar(&preCheckListPath, CmdFlagPreCheckListPath, "", "precheck list path")
	_ = stepCmd.MarkFlagRequired(CmdFlagPreCheckListPath)

	stepCmd.AddCommand(newPreCheckList())

	return stepCmd
}

func newPreCheckList() *cobra.Command {
	var (
		preCheckListPath string
	)
	stepCmd := &cobra.Command{
		Use:   "new",
		Short: "new precheck list",
		Long:  "new precheck list",
		RunE: func(_ *cobra.Command, _ []string) error {
			perCheckList := precheck.DefaultCheckList()
			bytes, err := json.Marshal(perCheckList)
			if err != nil {
				return err
			}

			checkListPath := preCheckListPath

			// nolint: mnd
			if err := os.WriteFile(checkListPath, bytes, 0600); err != nil {
				return err
			}

			fmt.Printf("create precheck list success, path: %s\n", checkListPath)

			return nil
		},
	}

	stepCmd.Flags().StringVar(&preCheckListPath, CmdFlagPreCheckListPath, "", "precheck list path")
	_ = stepCmd.MarkFlagRequired(CmdFlagPreCheckListPath)

	return stepCmd
}
