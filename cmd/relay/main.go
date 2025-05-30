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

	"github.com/TencentBlueKing/bk-nodemgr/internal/relay/service"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/blog"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
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
}

func main() {
	// configPath of backend service.
	var configPath string

	serverCmd := &cobra.Command{
		Use:     "bk_nodeman_relay",
		Short:   "bk-nodeman relay server",
		Long:    "bk-nodeman relay server",
		Version: version.FormatVersion(),
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

			if err := os.WriteFile(
				conf.Plugin.PidFile,
				[]byte(fmt.Sprintf("%d", os.Getpid())),
				os.FileMode(0644),
			); err != nil {
				fmt.Printf("failed to write pid file(%s): %v\n", conf.Plugin.PidFile, err)
				os.Exit(1)
			}

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
