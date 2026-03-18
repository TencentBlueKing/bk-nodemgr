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
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"
)

const (
	// logFileMode is the permission bits for the log file.
	// 0600 means read-write for owner only.
	logFileMode = 0600

	// defaultLogReportBulkSize is the default log report bulk size.
	defaultLogReportBulkSize = 10

	// maxRetainedLogFiles is the maximum number of log files to retain during rotation.
	maxRetainedLogFiles = 5
)

// NewHandler creates a new logger handler.
func NewHandler(logDir string, logToStd bool, deployToken, operInstID string, reportLogURLs []string) *Handler {
	return &Handler{
		logDir:      logDir,
		logToStd:    logToStd,
		logFilePath: filepath.Join(logDir, fmt.Sprintf("installer_%s.log", time.Now().Format("2006-01-02T15-04-05"))),
		args: ReportLogsArgs{
			Token:         deployToken,
			OperInstID:    operInstID,
			LogRptCnt:     0,
			BulkSize:      defaultLogReportBulkSize,
			ReportLogURLs: reportLogURLs,
		},
	}
}

// Handler logs handler.
type Handler struct {
	logDir      string
	logToStd    bool
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

	cleanOldLogFiles(lh.logDir, maxRetainedLogFiles)

	lh.logWFile, err = os.OpenFile(lh.logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, logFileMode)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	if lh.logToStd {
		log.SetOutput(io.MultiWriter(lh.logWFile, os.Stdout))
	} else {
		log.SetOutput(lh.logWFile)
	}

	if len(lh.args.ReportLogURLs) == 0 {
		logger.Infof(node.StepGeneral, "no report log URLs configured, local-only log mode")
		lh.watchStopper = func() error { return nil }
		return nil
	}

	lh.logRFile, err = os.OpenFile(lh.logFilePath, os.O_RDONLY, logFileMode)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	lh.args.Reader = lh.logRFile
	lh.reporter = NewReporter(lh.args)

	errCh, stopper := lh.reporter.Watch(1 * time.Second)
	go func() {
		for err := range errCh {
			logger.Errorf(node.StepGeneral, "log reporter error: %v", err)
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

// cleanOldLogFiles removes old installer log files, keeping only the most recent ones.
func cleanOldLogFiles(logDir string, keep int) {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return
	}

	type logFileEntry struct {
		name    string
		modTime time.Time
	}

	var logFiles []logFileEntry
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "installer_") || !strings.HasSuffix(name, ".log") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		logFiles = append(logFiles, logFileEntry{name: name, modTime: info.ModTime()})
	}

	if len(logFiles) <= keep {
		return
	}

	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i].modTime.After(logFiles[j].modTime)
	})

	for _, f := range logFiles[keep:] {
		_ = os.Remove(filepath.Join(logDir, f.name))
	}
}
