/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package logreporter

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	// logFileMode is the permission bits for the log file.
	// 0600 means read-write for owner only.
	logFileMode = 0600

	// defaultLogReportBulkSize is the default log report bulk size.
	defaultLogReportBulkSize = 10
)

// NewHandler creates a new logger handler.
func NewHandler(logDir, deployToken, operInstID, reportLogURL string) *Handler {
	return &Handler{
		logDir:      logDir,
		logFilePath: filepath.Join(logDir, fmt.Sprintf("installer_%s.log", time.Now().Format("2006-01-02T15-04-05"))),
		args: ReportLogsArgs{
			Token:        deployToken,
			OperInstID:   operInstID,
			LogRptCnt:    0,
			BulkSize:     defaultLogReportBulkSize,
			ReportLogURL: reportLogURL,
		},
	}
}

// Handler logs handler.
type Handler struct {
	logDir      string
	logFilePath string
	args        ReportLogsArgs

	logWFile     *os.File
	logRFile     *os.File
	reporter     *Reporter
	watchStopper func() error
}

// Start starts the logger handler.
func (lh *Handler) Start() error {
	err := os.MkdirAll(lh.logDir, logFileMode)
	if err != nil {
		return fmt.Errorf("failed to mkdir log dir: %w", err)
	}

	lh.logWFile, err = os.OpenFile(lh.logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, logFileMode)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	log.SetOutput(lh.logWFile)

	lh.logRFile, err = os.OpenFile(lh.logFilePath, os.O_RDONLY, logFileMode)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	lh.args.Reader = lh.logRFile
	lh.reporter = NewReporter(lh.args)

	errCh, stopper := lh.reporter.Watch(1 * time.Second)
	go func() {
		for err := range errCh {
			fmt.Println("failed to report logs: ", err.Error())
		}
	}()
	lh.watchStopper = stopper

	return nil
}

// Stop stops the logger handler.
func (lh *Handler) Stop() {
	_ = lh.watchStopper()

	if lh.logWFile != nil {
		_ = lh.logWFile.Close()
	}

	if lh.logRFile != nil {
		_ = lh.logRFile.Close()
	}
}
