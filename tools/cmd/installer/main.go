/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package main for installer
package main

import (
	"fmt"
	"os"

	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node"
	"github.com/spf13/cobra"
)

func main() {
	if err := NewRootCommand().Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// NewRootCommand creates the root command of installer.
// nolint: lll
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "installer",
		Short:        "installer",
		Long:         "nodemgr installer",
		SilenceUsage: true,
		Version:      "1.0.0",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// sub commands.
	rootCmd.AddCommand(node.NewNodeCommand())

	return rootCmd
}
