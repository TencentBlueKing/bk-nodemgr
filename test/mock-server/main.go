/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package main of mock-server.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/version"
	"github.com/TencentBlueKing/bk-nodemgr/test/mock-server/service"

	"github.com/spf13/cobra"
)

func main() {
	var configPath string

	serverCmd := &cobra.Command{
		Use:     "bk_nodeman_mock_server",
		Short:   "bk-nodeman mock-server",
		Long:    "bk-nodeman mock-server for API testing supports CMDB",
		Version: version.FormatVersion(),
		PreRun: func(_ *cobra.Command, _ []string) {
			fmt.Println(version.GetStartInfo())
		},
		Run: func(_ *cobra.Command, _ []string) {
			conf := NewMockService()
			if err := conf.LoadFromFile(configPath); err != nil {
				fmt.Printf("failed to load config file(%s): %v\n", configPath, err)
				os.Exit(1)
			}

			if err := conf.Validate(); err != nil {
				fmt.Printf("failed to validate config: %v\n", err)
				os.Exit(1)
			}

			// init log.
			logger.Init(logger.Config{
				LogDir:       conf.Log.Dir,
				LogMaxSizeMB: conf.Log.MaxSizeMB,
				LogMaxNum:    conf.Log.MaxNum,
				Level: func(level config.LogLevel) logger.Level {
					switch level {
					case config.LogLevelDebug:
						return logger.LevelDebug
					case config.LogLevelInfo:
						return logger.LevelInfo
					case config.LogLevelWarn:
						return logger.LevelWarn
					case config.LogLevelError:
						return logger.LevelError
					default:
						return logger.LevelInfo
					}
				}(conf.Log.Level),
				ToStdErr:     conf.Log.ToStdErr,
				AlsoToStdErr: conf.Log.AlsoToStdErr,
			})

			svc, err := service.NewService(service.Config{
				BasicServer:  conf.BasicServer,
				CMDBMockData: conf.MockData.CMDB,
			})
			if err != nil {
				fmt.Printf("failed to create service: %v\n", err)
				os.Exit(1)
			}

			go watchShutdown()

			if err := svc.Start(); err != nil {
				fmt.Printf("failed to start service: %v\n", err)
				os.Exit(1)
			}
		},
	}

	// command line arguments.
	serverCmd.PersistentFlags().StringVarP(
		&configPath, "file", "f", "", "path of service config file",
	)

	err := serverCmd.MarkPersistentFlagRequired("file")
	if err != nil {
		fmt.Printf("failed to mark flag required: %v\n", err)
		os.Exit(1)
	}

	if err := serverCmd.Execute(); err != nil {
		fmt.Printf("failed to execute cmd: %v\n", err)
		os.Exit(1)
	}
}

// watchShutdown listens signal and exits the server.
func watchShutdown() {
	// listening signal
	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, syscall.SIGINT, syscall.SIGTERM)
	receivedSignal := <-signalC

	// flush logs before shutdown.
	logger.G.Flush()

	fmt.Printf("received signal(%s), going to exit server\n", receivedSignal.String())

	os.Exit(1)
}
