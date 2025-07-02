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
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := NewRootCommand()
	err := rootCmd.Execute()

	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

// logFileMode is the permission bits for the log file.
// 0600 means read-write for owner only.
const logFileMode = 0600

// NewRootCommand creates the root command of installer.
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "installer",
		Short:        "installer",
		Long:         "nodemgr node installer",
		SilenceUsage: true,
		Version:      "1.0.0",
	}

	registerRootVars(rootCmd)
	registerRootPersistentVars(rootCmd)

	// this is a sub command.
	rootCmd.AddCommand(NewStepInstallAgent())
	rootCmd.AddCommand(NewStepUninstallAgent())
	rootCmd.AddCommand(NewStepDownloadFiles())
	rootCmd.AddCommand(NewStepReportLog())
	rootCmd.AddCommand(NewStepPreCheck())
	rootCmd.AddCommand(NewStepStartNode())
	rootCmd.AddCommand(NewCheckDeploy())
	rootCmd.AddCommand(NewStepReportData())
	rootCmd.AddCommand(NewStepReportStatus())

	return rootCmd
}

// nolint: gocognit
func registerRootPersistentVars(rootCmd *cobra.Command) {
	// this var will be set in PersistentPreRunE, they can act on the current command and its subcommands.
	var (
		logFilePath   string
		workspacePath string
		nodeRole      string
		debug         bool
	)

	// this is a global variable.
	rootCmd.PersistentFlags().
		StringVar(&logFilePath, CmdFlagLogFilePath, CmdDefaultLogFilePath(), "log file path")
	rootCmd.PersistentFlags().
		StringVar(&workspacePath, CmdFlagWorkspace, CmdDefaultWorkspace,
			"the dir which used to store agent pkg, configs and install log")
	rootCmd.PersistentFlags().
		StringVar(&nodeRole, CmdFlagNodeRole, "agent", "this is the node role")
	rootCmd.PersistentFlags().
		BoolVar(&debug, CmdFlagDebug, false, "debug mode, default is false")

	_ = rootCmd.MarkPersistentFlagRequired(CmdFlagNodeRole)

	logFile := new(os.File)
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		if err := SetNodeRole(constant.NodeRole(nodeRole)); err != nil {
			return fmt.Errorf("set node role failed, err: %v", err)
		}

		if logFilePath != "" {
			if err := SetLogFilePath(logFilePath); err != nil {
				return fmt.Errorf("set log file path failed, err: %v", err)
			}
		}

		if workspacePath != "" {
			if err := SetWorkspaceDir(workspacePath); err != nil {
				return fmt.Errorf("set workspace dir failed, err: %v", err)
			}
		}

		if err := utils.CheckWritePermission(workspacePath); err != nil {
			return fmt.Errorf("check write permission failed, err: %v", err)
		}

		if err := utils.TryCreateDir(filepath.Dir(GetLogFilePath())); err != nil {
			return fmt.Errorf("try create dir failed, err: %v", err)
		}

		file, err := os.OpenFile(GetLogFilePath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, logFileMode)
		if err != nil {
			return err
		}
		logFile = file

		if debug {
			err = logger.SetLevel(logger.LevelDebug)
			if err != nil {
				return err
			}

			combinedOutput := io.MultiWriter(os.Stdout, file)
			log.SetOutput(combinedOutput)
		} else {
			log.SetOutput(file)
		}

		return nil
	}

	rootCmd.PersistentPostRunE = func(_ *cobra.Command, _ []string) error {
		err := logFile.Close()
		if err != nil {
			return fmt.Errorf("close log file failed, err: %v", err)
		}

		return nil
	}
}

const defaultLogReportBulkSize = 10

// nolint: funlen,gocognit,cyclop,gocyclo
func registerRootVars(rootCmd *cobra.Command) {
	// this var will be set in PreRunE, their scope is limited to this command.
	var (
		downloadEndpoint  string
		pkgGeneration     int
		pkgVersion        string
		callBackEndPoint  string
		gseRoot           string
		token             string
		operInstID        string
		deployEnv         string
		reinstall         bool
		reRegisterAgentID bool
		agentID           string
	)
	rootCmd.PreRunE = func(_ *cobra.Command, _ []string) error {
		if gseRoot != "" {
			if err := SetGseRoot(gseRoot); err != nil {
				return fmt.Errorf("set gse root failed, err: %v", err)
			}
		}

		if agentID != "" {
			if err := SetNodeAgentID(agentID); err != nil {
				return fmt.Errorf("set node agent id failed, err: %v", err)
			}
		}

		if err := SetDownloadEndPoint(downloadEndpoint); err != nil {
			return fmt.Errorf("set download endpoint failed, err: %v", err)
		}

		if err := SetCallbackEndPoint(callBackEndPoint); err != nil {
			return fmt.Errorf("set callback endpoint failed, err: %v", err)
		}

		if err := SetNodeGeneration(pkgGeneration); err != nil {
			return fmt.Errorf("set node pkgGeneration failed, err: %v", err)
		}

		if err := SetNodeVersion(pkgVersion); err != nil {
			return fmt.Errorf("set node pkgVersion failed, err: %v", err)
		}

		if err := SetServerConf(token, operInstID); err != nil {
			return fmt.Errorf("set server conf failed, err: %v", err)
		}

		if err := SetDeployEnv(deployEnv); err != nil {
			return fmt.Errorf("set deploy env failed, err: %v", err)
		}

		if reinstall {
			SetReinstall(reinstall)
		}

		if reRegisterAgentID {
			SetReRegisterAgentID(reRegisterAgentID)
		}

		return nil
	}

	rootCmd.RunE = install

	// this is the root command's private variable.
	rootCmd.Flags().StringVar(&downloadEndpoint, CmdFlagDownloadEndpoint, "", "download endpoint")
	rootCmd.Flags().StringVar(&callBackEndPoint, CmdFlagCallbackEndpoint, "", "callback endpoint")
	rootCmd.Flags().IntVar(&pkgGeneration, CmdFlagPkgGeneration, CmdDefaultPkgGeneration,
		"this is the gse pkg generation which will be installed")
	rootCmd.Flags().StringVar(&token, CmdFlagToken, "", "token")
	rootCmd.Flags().StringVar(&operInstID, CmdFlagOperInstID, "", "operation instance id")
	rootCmd.Flags().StringVar(&deployEnv, CmdFlagDeployEnv, "", "deploy env")
	rootCmd.Flags().StringVar(&pkgVersion, CmdFlagPkgVersion, "", "this gse node pkg version which will be installed")
	rootCmd.Flags().StringVar(&gseRoot, CmdFlagGseRoot, CmdDefaultGseRoot(), "gse root")
	rootCmd.Flags().StringVar(&agentID, CmdFlagAgentID, "", "gse agent id")
	rootCmd.Flags().BoolVar(&reinstall, CmdFlagReinstall, false, "reinstall")
	rootCmd.Flags().BoolVar(&reRegisterAgentID, CmdFlagReRegisterAgentID, false, "re register agent id")
	_ = rootCmd.MarkFlagRequired(CmdFlagDownloadEndpoint)
	_ = rootCmd.MarkFlagRequired(CmdFlagCallbackEndpoint)
	_ = rootCmd.MarkFlagRequired(CmdFlagPkgGeneration)
	_ = rootCmd.MarkFlagRequired(CmdFlagPkgVersion)
	_ = rootCmd.MarkFlagRequired(CmdFlagGseRoot)
	_ = rootCmd.MarkFlagRequired(CmdFlagToken)
	_ = rootCmd.MarkFlagRequired(CmdFlagOperInstID)
	_ = rootCmd.MarkFlagRequired(CmdFlagDeployEnv)
}
