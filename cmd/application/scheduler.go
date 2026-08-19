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

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/spf13/cobra"
)

// NewSchedulerCMD generates a new scheduler command.
func NewSchedulerCMD() *cobra.Command {
	var configPath string

	schedulerCMD := &cobra.Command{
		Use:   "scheduler",
		Short: "Execute tasks based on cron expressions, please ensure only one running scheduler.",
		Run: func(_ *cobra.Command, _ []string) {
			config := &config.ApplicationService{}
			if err := config.LoadFromFile(configPath); err != nil {
				fmt.Printf("failed to load config file(%s): %v\n", configPath, err)
				os.Exit(1)
			}

			if err := config.Validate(); err != nil {
				fmt.Printf("failed to validate config: %v\n", err)
				os.Exit(1)
			}

			// listening signal
			signalC := make(chan os.Signal, 1)
			signal.Notify(signalC, syscall.SIGINT, syscall.SIGTERM)
			receivedSignal := <-signalC

			fmt.Printf("received signal(%s), going to exit server\n", receivedSignal.String())
		},
	}

	schedulerCMD.Flags().StringVar(&configPath, "config", "", "config file")

	return schedulerCMD
}
