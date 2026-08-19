/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package main of relay.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/service"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/profiling"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/version"
	"github.com/spf13/cobra"
)

const (
	fileMode644 = 0644
	dirMode600  = 0600
)

func main() {
	// configPath of relay service.
	var configPath string

	serverCmd := &cobra.Command{
		Use:     "bk_nodemgr_relay",
		Short:   "bk-nodemgr relay server",
		Long:    "bk-nodemgr relay server",
		Version: version.VERSION,
		PreRun: func(_ *cobra.Command, _ []string) {
			fmt.Println(version.GetStartInfo())
		},
		Run: func(_ *cobra.Command, _ []string) {
			conf := config.NewRelayService()
			if err := conf.LoadFromFile(configPath); err != nil {
				fmt.Printf("failed to load config file(%s): %v\n", configPath, err)
				os.Exit(1)
			}

			if err := conf.Validate(); err != nil {
				fmt.Printf("failed to validate config: %v\n", err)
				os.Exit(1)
			}

			ensureAndCreatePidFile(conf.Plugin.PidFile)

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

			stopProfiler, err := profiling.Start(conf.Profiling, profiling.Defaults{
				ApplicationName: "bk-nodemgr-relay",
				Tags:            map[string]string{"service": "relay"},
			})
			if err != nil {
				fmt.Printf("failed to start profiling: %v\n", err)
				os.Exit(1)
			}

			svc, err := service.NewService(conf)
			if err != nil {
				if err := stopProfiler(); err != nil {
					fmt.Printf("failed to stop profiling: %v\n", err)
				}
				fmt.Printf("failed to create service: %v\n", err)
				os.Exit(1)
			}

			if err := svc.Start(); err != nil {
				if err := stopProfiler(); err != nil {
					fmt.Printf("failed to stop profiling: %v\n", err)
				}
				fmt.Printf("failed to start service: %v\n", err)
				os.Exit(1)
			}

			if err := stopProfiler(); err != nil {
				fmt.Printf("failed to stop profiling: %v\n", err)
			}
		},
	}

	serverCmd.PersistentFlags().StringVarP(
		&configPath, "config", "c", "", "path of service config file",
	)
	serverCmd.SetVersionTemplate("{{.Version}}\n")

	err := serverCmd.MarkPersistentFlagRequired("config")
	if err != nil {
		fmt.Printf("failed to mark flag required: %v\n", err)
		os.Exit(1)
	}

	err = serverCmd.Execute()
	if err != nil {
		fmt.Printf("failed to execute cmd: %v\n", err)
		os.Exit(1)
	}
}

func ensureAndCreatePidFile(pidFilePath string) {
	pidDir := filepath.Dir(pidFilePath)
	info, err := os.Stat(pidDir)
	if err == nil {
		if !info.IsDir() {
			fmt.Printf("failed to create pid dir(%s): file exists and not a dir\n", pidDir)
			os.Exit(1)
		}
	}

	if os.IsNotExist(err) {
		if err := os.MkdirAll(pidDir, os.FileMode(dirMode600)); err != nil {
			fmt.Printf("failed to create pid dir(%s): %v\n", pidDir, err)
			os.Exit(1)
		}
	}

	if err := os.WriteFile(
		pidFilePath,
		fmt.Appendf(nil, "%d", os.Getpid()),
		os.FileMode(fileMode644),
	); err != nil {
		fmt.Printf("failed to write pid file(%s): %v\n", pidFilePath, err)
		os.Exit(1)
	}
}
