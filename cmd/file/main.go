/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package main of file.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	machineryLog "github.com/RichardKnop/machinery/v2/log"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/service"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/blog"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/version"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

func init() {
	gin.DebugPrintFunc = func(format string, args ...interface{}) {
		_, err := fmt.Fprintf(blog.WriterDebug{}, format, args...)
		if err != nil {
			fmt.Printf("failed to write gin debug log, err: %v", err)
			os.Exit(1)
		}
	}

	machineryLog.Set(logger{})
}

type logger struct{}

// Print ...
func (l logger) Print(args ...interface{}) {
	blog.Debug(args...)
}

// Printf ...
func (l logger) Printf(s string, args ...interface{}) {
	blog.Debugf(s, args...)
}

// Println ...
func (l logger) Println(args ...interface{}) {
	blog.Debug(args...)
}

// Fatal ...
func (l logger) Fatal(args ...interface{}) {
	blog.Error(args...)
}

// Fatalf ...
func (l logger) Fatalf(s string, args ...interface{}) {
	blog.Errorf(s, args...)
}

// Fatalln ...
func (l logger) Fatalln(args ...interface{}) {
	blog.Error(args...)
}

// Panic ...
func (l logger) Panic(args ...interface{}) {
	blog.Error(args...)
}

// Panicf ...
func (l logger) Panicf(s string, args ...interface{}) {
	blog.Errorf(s, args...)
}

// Panicln ...
func (l logger) Panicln(args ...interface{}) {
	blog.Error(args...)
}

// file service entrypoint.
// nolint: funlen
// NOCC: golint/fnsize.
func main() {
	// configPath of file service.
	var configPath string

	serverCmd := &cobra.Command{
		Use:     "bk_nodeman_file",
		Short:   "bk-nodeman file server",
		Long:    "bk-nodeman file server",
		Version: version.FormatVersion(),
		PreRun: func(_ *cobra.Command, _ []string) {
			fmt.Println(version.GetStartInfo())
		},
		Run: func(_ *cobra.Command, _ []string) {
			conf := config.NewFileService()
			if err := conf.LoadFromFile(configPath); err != nil {
				fmt.Printf("failed to load config file(%s): %v\n", configPath, err)
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

			switch conf.TenantMode {
			case tenant.ModeSingle:
				tenant.SetMode(tenant.ModeSingle)
			case tenant.ModeMultiple:
				tenant.SetMode(tenant.ModeMultiple)
			default:
				fmt.Printf("invalid tenant mode: %s\n", conf.TenantMode)
				os.Exit(1)
			}

			fmt.Printf("tenant mode: %s\n", conf.TenantMode)

			// init log.
			logConfig := blog.NewLogConfig()
			logConfig.LogDir = conf.Log.Dir
			logConfig.LogMaxSizeMB = conf.Log.MaxSizeMB
			logConfig.LogMaxNum = conf.Log.MaxNum
			logConfig.Level = string(conf.Log.Level)
			logConfig.ToStdErr = conf.Log.ToStdErr
			logConfig.AlsoToStdErr = conf.Log.AlsoToStdErr
			blog.InitLogs(logConfig)

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

	serverCmd.PersistentFlags().StringVarP(
		&configPath, "file", "f", "", "path of service config file",
	)

	err := serverCmd.MarkPersistentFlagRequired("file")
	if err != nil {
		fmt.Printf("failed to mark flag required, err: %v\n", err)
		os.Exit(1)
	}

	err = serverCmd.Execute()
	if err != nil {
		fmt.Printf("failed to execute cmd, err: %v\n", err)
		os.Exit(1)
	}
}

// watchShutdown listens signal and calls service.GracefulShutdown.
func watchShutdown(svc *service.Service) {
	// listening signal
	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, syscall.SIGINT, syscall.SIGTERM)
	receivedSignal := <-signalC

	if err := svc.GracefulShutdown(); err != nil {
		fmt.Printf("failed to graceful shutdown service: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("received signal(%s), going to exit server\n", receivedSignal.String())

	os.Exit(1)
}
