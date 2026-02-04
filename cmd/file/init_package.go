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
	"errors"
	"fmt"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/internal/file/manager"
	"github.com/TencentBlueKing/bk-nodemgr/internal/file/service"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/spf13/cobra"
)

const (
	localCertDir          = "cert"
	localBintoolDir       = "bintool"
	localPluginBintoolDir = "plugin_bintool"
	localAgentDir         = "agent"
	localProxyDir         = "proxy"
	localServerDir        = "server"
	localPluginV2Dir      = "plugin_v2"
	localPluginV3Dir      = "plugin_v3"
)

// InitPackageOptions defines the options for init package command.
type InitPackageOptions struct {
	Dir               string
	Operator          string
	InitCert          bool
	InitBintool       bool
	InitPluginBintool bool
	InitAgent         bool
	InitProxy         bool
	InitServer        bool
	InitPluginV2      bool
	InitPluginV3      bool
	InitAll           bool
}

// NewInitPackageCMD generates a new init package command.
func NewInitPackageCMD() *cobra.Command {
	var (
		configPath string
		verbose    bool
		options    InitPackageOptions
	)

	initPackageCMD := &cobra.Command{
		Use:   "init-package",
		Short: "Start the init package process.",
		RunE: func(_ *cobra.Command, _ []string) error {
			conf := config.NewFileService()
			if err := conf.LoadFromFile(configPath); err != nil {
				fmt.Printf("failed to load config(%s): %v\n", configPath, err)
				return fmt.Errorf("failed to load config file(%s): %w", configPath, err)
			}

			if err := conf.Validate(); err != nil {
				fmt.Printf("failed to validate config: %v\n", err)
				return fmt.Errorf("failed to validate config: %w", err)
			}

			// init log.
			logger.Init(logger.Config{
				Level: func(verbose bool) logger.Level {
					if verbose {
						return logger.LevelDebug
					}

					return logger.LevelInfo
				}(verbose),
				ToStdErr: true,
			})

			svc, err := service.NewCMDService(conf)
			if err != nil {
				fmt.Printf("failed to create cmd service: %v\n", err)
				return fmt.Errorf("failed to create cmd service: %w", err)
			}

			go watchShutdown(svc)

			if err := svc.StartCapability(); err != nil {
				fmt.Printf("failed to start capability: %v\n", err)
				return fmt.Errorf("failed to start capability: %w", err)
			}

			if err := initPackage(svc.Cap.Manager, options); err != nil {
				fmt.Printf("failed to init package: %v\n", err)
				return fmt.Errorf("failed to init package: %w", err)
			}

			fmt.Println("init package success")

			return nil
		},
	}

	initPackageCMD.PersistentFlags().StringVarP(
		&configPath, "file", "f", "", "path of service config file",
	)

	initPackageCMD.PersistentFlags().StringVarP(
		&options.Dir, "dir", "d", "", "directory for the package",
	)

	initPackageCMD.PersistentFlags().StringVarP(
		&options.Operator, "operator", "o", "admin", "operator of the init package",
	)

	initPackageCMD.PersistentFlags().BoolVar(
		&options.InitCert, "init-cert", false, "whether to init certs",
	)

	initPackageCMD.PersistentFlags().BoolVar(
		&options.InitBintool, "init-bintool", false, "whether to init bintool",
	)

	initPackageCMD.PersistentFlags().BoolVar(
		&options.InitPluginBintool, "init-plugin-bintool", false, "whether to init plugin bintool",
	)

	initPackageCMD.PersistentFlags().BoolVar(
		&options.InitAgent, "init-agent", false, "whether to init agent",
	)

	initPackageCMD.PersistentFlags().BoolVar(
		&options.InitProxy, "init-proxy", false, "whether to init proxy",
	)

	initPackageCMD.PersistentFlags().BoolVar(
		&options.InitServer, "init-server", false, "whether to init server",
	)

	initPackageCMD.PersistentFlags().BoolVar(
		&options.InitPluginV2, "init-plugin-v2", false, "whether to init plugin v2",
	)

	initPackageCMD.PersistentFlags().BoolVar(
		&options.InitPluginV3, "init-plugin-v3", false, "whether to init plugin v3",
	)

	initPackageCMD.PersistentFlags().BoolVar(
		&options.InitAll, "init-all", false, "whether to init all packages",
	)

	initPackageCMD.PersistentFlags().BoolVarP(
		&verbose, "verbose", "v", false, "enable verbose output",
	)

	return initPackageCMD
}

// nolint: gocognit, gocyclo, cyclop
func initPackage(mgr manager.IManager, options InitPackageOptions) error {
	if options.Dir == "" {
		return errors.New("dir is required")
	}

	nCtx := contextx.New(contextx.Background(), contextx.WithBKUsername(options.Operator))

	logger.G.Biz(nCtx).With("dir", options.Dir).Info("start to init package from local dir")

	if options.InitCert || options.InitAll {
		if err := mgr.PublishReleaseCertFromLocalDir(nCtx, filepath.Join(options.Dir, localCertDir)); err != nil {
			return err
		}
	}

	if options.InitBintool || options.InitAll {
		if err := mgr.PublishReleaseBinToolFromLocalDir(nCtx, filepath.Join(options.Dir, localBintoolDir)); err != nil {
			return err
		}
	}

	if options.InitPluginBintool || options.InitAll {
		if err := mgr.PublishReleasePluginBinToolFromLocalDir(nCtx, filepath.Join(options.Dir, localPluginBintoolDir)); err != nil {
			return err
		}
	}

	if options.InitAgent || options.InitAll {
		if err := mgr.PublishReleaseAgentFromLocalDir(nCtx, filepath.Join(options.Dir, localAgentDir)); err != nil {
			return err
		}
	}

	if options.InitProxy || options.InitAll {
		if err := mgr.PublishReleaseProxyFromLocalDir(nCtx, filepath.Join(options.Dir, localProxyDir)); err != nil {
			return err
		}
	}

	if options.InitServer || options.InitAll {
		if err := mgr.PublishReleaseServerFromLocalDir(nCtx, filepath.Join(options.Dir, localServerDir)); err != nil {
			return err
		}
	}

	if options.InitPluginV2 || options.InitAll {
		if err := mgr.PublishReleasePluginV2FromLocalDir(nCtx, filepath.Join(options.Dir, localPluginV2Dir)); err != nil {
			return err
		}
	}

	if options.InitPluginV3 || options.InitAll {
		if err := mgr.PublishReleasePluginV3FromLocalDir(nCtx, filepath.Join(options.Dir, localPluginV3Dir)); err != nil {
			return err
		}
	}

	logger.G.Biz(nCtx).With("dir", options.Dir).Info("finished to init package from local dir")

	return nil
}
