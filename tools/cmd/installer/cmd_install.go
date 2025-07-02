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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logreporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/statusreporter"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/gopool"
	"github.com/spf13/cobra"
)

// nolint: nonamedreturns
func install(cmd *cobra.Command, _ []string) (runErr error) {
	defer func() {
		state := constant.StateSuccess
		if runErr != nil {
			state = constant.StateFailed
		}

		reportStatusStep := statusreporter.NewStep(statusreporter.StepArgs{
			Token:            GetToken(),
			Status:           state,
			CallbackEndpoint: GetCallBackEndpoint(),
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
		return stepUninstallNode(cmd.Context())
	})

	gp.Go(func() error {
		return stepDownloadFiles(cmd.Context())
	})

	if err := gp.Wait(); err != nil {
		return fmt.Errorf("concurrent tasks failed, err: %w", err)
	}

	if err := stepPreCheck(cmd.Context()); err != nil {
		return err
	}

	if err := stepInstallNode(cmd.Context()); err != nil {
		return err
	}

	fmt.Println(GetNodeAgentID())

	if err := stepStartNode(cmd.Context()); err != nil {
		return err
	}

	if err := stepCheckDeploy(cmd.Context()); err != nil {
		return err
	}

	if err := stepReportData(cmd.Context()); err != nil {
		return err
	}

	return nil
}
