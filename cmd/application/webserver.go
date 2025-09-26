/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/TencentBlueKing/bk-nodemgr/internal/application/service"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

// NewWebServerCMD generates a new webserver command.
func NewWebServerCMD() *cobra.Command {
	// configPath of application service.
	var configPath string

	wsCMD := &cobra.Command{
		Use:   "webserver",
		Short: "Start the HTTP server.",
		Run: func(_ *cobra.Command, _ []string) {
			conf := config.NewApplicationService()
			if err := conf.Load(configPath); err != nil {
				fmt.Printf("failed to load config(%s): %v\n", configPath, err)
				os.Exit(1)
			}

			if err := conf.Validate(); err != nil {
				fmt.Printf("failed to validate config: %v\n", err)
				os.Exit(1)
			}

			switch conf.RunMode {
			case config.RunModeDebug:
				gin.SetMode(gin.DebugMode)
			case config.RunModeRelease:
				gin.SetMode(gin.ReleaseMode)
			default:
				fmt.Printf("invalid mode: %s\n", conf.RunMode)
				os.Exit(1)
			}

			fmt.Printf("run mode: %s\n", conf.RunMode)

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

			svc, err := service.NewService(conf)
			if err != nil {
				fmt.Printf("failed to create service: %v\n", err)
				os.Exit(1)
			}

			go watchShutdown(svc)

			if err := svc.Start(); err != nil {
				fmt.Printf("failed to start service: %v\n", err)
				os.Exit(1)
			}
		},
	}

	wsCMD.PersistentFlags().StringVarP(
		&configPath, "file", "f", "", "path of service config file",
	)

	return wsCMD
}

// watchShutdown listens signal and calls service.GracefulShutdown.
func watchShutdown(svc *service.Service) {
	// listening signal
	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, syscall.SIGINT, syscall.SIGTERM)
	receivedSignal := <-signalC

	// flush logs before shutdown.
	logger.G.Flush()

	if err := svc.GracefulShutdown(); err != nil {
		fmt.Printf("failed to graceful shutdown service: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("received signal(%s), going to exit server\n", receivedSignal.String())

	os.Exit(1)
}
