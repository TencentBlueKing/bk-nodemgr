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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/checkdeploy"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/datareporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/filedownloader"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logreporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/nodeinstaller"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/precheck"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/startnode"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/statusreporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/uninstallnode"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/gopool"
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
	rootCmd.AddCommand(NewStepReinstall())
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

		if err := SetToken(token); err != nil {
			return fmt.Errorf("set token failed, err: %v", err)
		}

		return nil
	}

	rootCmd.RunE = func(cmd *cobra.Command, _ []string) (runErr error) {
		defer func() {
			state := constant.StateSuccess
			if runErr != nil {
				state = constant.StateFailed
			}
			reportStatusStep := statusreporter.NewStep(statusreporter.StepArgs{
				Token:            GetToken(),
				Status:           state,
				CallbackEndpoint: callBackEndPoint,
			})
			if err := reportStatusStep.Run(cmd.Context()); err != nil {
				logger.Errorf(constant.StepReportStatus, "status report failed: %v", err)
			}
		}()

		logFile, err := os.OpenFile(GetLogFilePath(), os.O_RDONLY, 0600) // nolint: mnd
		if err != nil {
			return fmt.Errorf("open log file failed, err: %v", err)
		}

		defer func() {
			_ = logFile.Close()
		}()

		reporter := logreporter.NewReporter(logreporter.ReportLogsArgs{
			Token:            GetToken(),
			Reader:           logFile,
			LogRptCnt:        0,
			BulkSize:         defaultLogReportBulkSize,
			CallbackEndpoint: GetCallBackEndpoint(),
		})
		errCh, stop := reporter.Watch(1 * time.Second)
		go func() {
			for err := range errCh {
				fmt.Printf("log report failed, err: %v", err)
			}
		}()
		defer func() {
			if err := stop(); err != nil {
				fmt.Printf("stop log reporter failed, err: %v", err)
			}
		}()

		gp := gopool.NewPool()
		gp.Go(func() error {
			if !reinstall {
				// don't reinstall

				return nil
			}

			uninstallStep := uninstallnode.NewStep(uninstallnode.StepArgs{
				SetupDirPath: GetSetupDir(),
				BinDirPath:   GetBinDir(),
				GseCtlPath:   GetGseCtlPath(),
				DeployEnv:    GetDeployEnv(),
			})
			if err := uninstallStep.Run(cmd.Context()); err != nil {
				return fmt.Errorf("uninstall step failed, err: %v", err)
			}

			return nil
		})

		gp.Go(func() error {
			downloadFilesStep := filedownloader.NewStep(filedownloader.StepArgs{
				DownloadPoint:        GetDownloadEndPoint(),
				CallbackEndpoint:     GetCallBackEndpoint(),
				PkgGeneration:        GetNodePkgGeneration(),
				PkgPath:              GetGsePkgPath(),
				PkgVersion:           GetNodePkgVersion(),
				NodeRole:             GetNodeRole(),
				Token:                GetToken(),
				TmpAgentConfPath:     GetTmpAgentConfPath(),
				TmpFileProxyConfPath: GetTmpFileProxyConfPath(),
				TmpDataProxyConfPath: GetTmpDataProxyConfPath(),
				CheckListPath:        GetPreCheckFilePath(),
			})
			if err := downloadFilesStep.Run(cmd.Context()); err != nil {
				return fmt.Errorf("download files failed, err: %v", err)
			}

			return nil
		})

		if err := gp.Wait(); err != nil {
			return fmt.Errorf("concurrent tasks failed: %w", err)
		}

		preCheckStep := precheck.NewStep(precheck.StepArgs{
			PreCheckListPath: GetPreCheckFilePath(),
			SetupDirPath:     GetSetupDir(),
		})

		if err := preCheckStep.Run(cmd.Context()); err != nil {
			return fmt.Errorf("precheck failed: %w", err)
		}

		installAgentStep := nodeinstaller.NewStep(nodeinstaller.StepArgs{
			AgentID:           GetNodeAgentID(),
			ReRegisterAgentID: reRegisterAgentID && reinstall,
			SetupDirPath:      GetSetupDir(),
			PkgPath:           GetGsePkgPath(),
			SrcConfigDir:      GetTmpConfigDir(),
			Overwrite:         false,
		})

		if agentID, err = installAgentStep.Run(cmd.Context()); err != nil {
			return fmt.Errorf("install agent failed: %w", err)
		}

		if err = SetNodeAgentID(agentID); err != nil {
			return fmt.Errorf("set node agent id failed: %w", err)
		}

		startNodeStep := startnode.NewStep(startnode.StepArgs{
			GseCtlPath: GetGseCtlPath(),
		})
		if err = startNodeStep.Run(cmd.Context()); err != nil {
			return fmt.Errorf("start node failed: %w", err)
		}

		fmt.Println(agentID)

		checkDeployStep := checkdeploy.NewStep(checkdeploy.StepArgs{
			RunDir:    GetRunDir(),
			NodeRole:  GetNodeRole(),
			DeployEnv: GetDeployEnv(),
		})
		if err = checkDeployStep.Run(cmd.Context()); err != nil {
			return fmt.Errorf("check deploy failed: %w", err)
		}

		reportDataStep := datareporter.NewStep(datareporter.StepArgs{
			Token:            GetToken(),
			AgentID:          GetNodeAgentID(),
			CallbackEndpoint: GetCallBackEndpoint(),
		})
		if err = reportDataStep.Run(cmd.Context()); err != nil {
			return fmt.Errorf("report data failed: %w", err)
		}

		return nil
	}

	// this is the root command's private variable.
	rootCmd.Flags().StringVar(&downloadEndpoint, CmdFlagDownloadEndpoint, "", "download endpoint")
	rootCmd.Flags().StringVar(&callBackEndPoint, CmdFlagCallbackEndpoint, "", "callback endpoint")
	rootCmd.Flags().IntVar(&pkgGeneration, CmdFlagPkgGeneration, CmdDefaultPkgGeneration,
		"this is the gse pkg generation which will be installed")
	rootCmd.Flags().StringVar(&token, CmdFlagToken, "", "token")
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
}
