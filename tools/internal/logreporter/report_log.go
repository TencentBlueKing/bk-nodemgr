/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package logreporter this package provides the ability to report logs and data.
package logreporter

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/retrier"
)

// maxBulkLogSize define the maximum number of logs that can be sent in one bulk request.
const maxBulkLogSize = 100

// ReportLogsArgs report logs args.
type ReportLogsArgs struct {
	// Token the token for report log.
	Token string

	// OperInstID the operation instance ID for report log.
	OperInstID string

	// Reader the Reader for log lines.
	Reader io.ReadCloser

	// LogRptCnt the number of logs that have been reported.
	LogRptCnt uint

	// BulkSize the number of logs that can be sent in one bulk request.
	BulkSize int

	// ReportLogURLs the URLs for report log (multiple server addresses).
	ReportLogURLs []string
}

// Reporter report logs.
type Reporter struct {
	token         string
	operInstID    string
	reader        io.ReadCloser
	mu            sync.Mutex
	logRptCnt     uint
	bulkSize      int
	reportLogURLs []string
}

// Validate ReportLogsArgs.
func (reporter *Reporter) Validate() error {
	if reporter.token == "" {
		return errors.New("token is empty")
	}

	if reporter.operInstID == "" {
		return errors.New("operation instance id is empty")
	}

	if reporter.reader == nil {
		return errors.New("reader is nil")
	}

	return nil
}

// NewReporter new reporter.
func NewReporter(args ReportLogsArgs) *Reporter {
	reporter := &Reporter{
		token:         args.Token,
		operInstID:    args.OperInstID,
		reader:        args.Reader,
		logRptCnt:     args.LogRptCnt,
		bulkSize:      args.BulkSize,
		reportLogURLs: args.ReportLogURLs,
	}

	return reporter
}

// ReportLogs bulk report log entries.
func (reporter *Reporter) ReportLogs(ctx context.Context) (uint, error) {
	if ctx == nil {
		return 0, errors.New("ctx is nil")
	}

	if err := reporter.Validate(); err != nil {
		return 0, err
	}

	reporter.mu.Lock()
	defer reporter.mu.Unlock()

	scanner := bufio.NewScanner(reporter.reader)
	bulkLog := make([]string, 0)
	lineNo := uint(0)
	for scanner.Scan() {
		lineNo++
		if lineNo <= reporter.logRptCnt {
			// ignore already reported lines.
			continue
		}

		line := scanner.Text()
		if line != "" {
			bulkLog = append(bulkLog, line)
		}
	}

	// if no log entries, return.
	if len(bulkLog) == 0 {
		return lineNo, nil
	}

	// build log entries.
	logEntries := make([]*LogEntry, 0, len(bulkLog))
	for _, line := range bulkLog {
		logEntry, err := convLogLineToLogEntry(line)
		if err != nil {
			logger.Warnf(node.StepGeneral, "failed to convert log line: %v", err)
			continue
		}

		logEntries = append(logEntries, logEntry)
	}

	backoff := retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault())
	splitLog := splitIntoN(logEntries, len(logEntries)/maxBulkLogSize-1)
	for _, entries := range splitLog {
		req := &ReportLogReq{
			Token:      reporter.token,
			OperInstID: reporter.operInstID,
			Logs:       entries,
		}

		if err := bulkReportLogsMultiEndpoint(ctx, backoff, reporter.reportLogURLs, req); err != nil {
			return 0, fmt.Errorf("bulk report logs failed: %v", err)
		}
	}

	// ignore empty logs.
	if len(logEntries) == 0 {
		return 0, nil
	}

	return lineNo, nil
}

// splitIntoN split items into n parts.
func splitIntoN(items []*LogEntry, num int) [][]*LogEntry {
	if num <= 0 {
		return [][]*LogEntry{items}
	}

	if num > len(items) {
		num = len(items)
	}

	result := make([][]*LogEntry, num)
	for i := 0; i < num; i++ {
		start := i * len(items) / num
		end := (i + 1) * len(items) / num
		result[i] = items[start:end]
	}

	return result
}

const bulkReportLogsTimeout = 10 * time.Second

// bulkReportLogsMultiEndpoint tries multiple server addresses in order until one succeeds.
func bulkReportLogsMultiEndpoint(ctx context.Context, retrier retrier.Retrier, reportURLs []string, req *ReportLogReq) error {
	if len(reportURLs) == 0 {
		return fmt.Errorf("no report log URLs provided")
	}

	var lastErr error
	for i, reportURL := range reportURLs {
		logger.Debugf(node.StepGeneral, "attempting to report logs to server, index(%d/%d), url(%s)", i+1, len(reportURLs), reportURL)
		err := bulkReportLogs(ctx, retrier, reportURL, req)
		if err == nil {
			return nil
		}
		logger.Warnf(node.StepGeneral, "failed to report logs: %v", err)
		lastErr = err
	}

	// All servers failed, return error with server URLs for debugging
	return fmt.Errorf("failed to report logs to all %d server(s) %v: %w", len(reportURLs), reportURLs, lastErr)
}

// bulkReportLogs bulk report logs to a single server.
func bulkReportLogs(ctx context.Context, retrier retrier.Retrier, reportURL string, req *ReportLogReq) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request body failed: %v", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, reportURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("create report request failed: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{
		Timeout: bulkReportLogsTimeout,
	}

	err = retrier.Do(ctx, func(attempt int) error {
		resp, err := httpClient.Do(request)
		if err != nil {
			return fmt.Errorf("retry failed, attempt: %d, error: %v", attempt, err)
		}

		// make sure to read and close the body to prevent leaks.
		defer func() {
			// inorder to avoid leaking the body, we read it and discard it.
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			logger.Errorf(node.StepGeneral, "failed to send report request, status: %d, body: %s", resp.StatusCode, body)

			return fmt.Errorf("send report request failed, status: %d", resp.StatusCode)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("send report request failed: %v", err)
	}

	return nil
}

// ReportLogReq report log request.
type ReportLogReq struct {
	Token      string      `json:"token"`
	OperInstID string      `json:"oper_inst_id"`
	Logs       []*LogEntry `json:"logs"`
}

// LogEntry log entry.
type LogEntry struct {
	Timestamp int64  `json:"timestamp"`
	Level     string `json:"level"`
	Step      string `json:"step"`
	Log       string `json:"log"`
}

// convLogLineToLogEntry converts a log line to a log entry.
func convLogLineToLogEntry(logLine string) (*LogEntry, error) {
	fields := strings.Split(logLine, "|")

	if len(fields) < 4 { // nolint: mnd
		return nil, fmt.Errorf("invalid log line: %s", logLine)
	}

	datetime := strings.TrimSpace(fields[0])
	logLevel := strings.TrimSpace(fields[1])
	step := strings.TrimSpace(fields[2])

	// merge remaining fields into a single log message
	message := strings.Join(fields[3:], " ")

	t, err := time.Parse("2006/01/02 15:04:05", datetime)
	if err != nil {
		return nil, fmt.Errorf("parse time failed: %v", err)
	}

	logEntry := &LogEntry{
		Timestamp: t.Unix(),
		Level:     logLevel,
		Step:      step,
		Log:       message,
	}

	return logEntry, nil
}

const defaultReportDataTimeOut = 30 * time.Second

// Watch watch logs and report.
func (reporter *Reporter) Watch(interval time.Duration) (<-chan error, func() error) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	errorCh := make(chan error, 10) // nolint: mnd

	go func() {
		defer close(done)
		defer close(errorCh)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				reportCtx, reportCancel := context.WithTimeout(context.Background(), defaultReportDataTimeOut)
				// nolint: contextcheck
				_, err := reporter.ReportLogs(reportCtx)
				reportCancel()

				if err != nil {
					errorCh <- fmt.Errorf("final report failed: %w", err)
				}

				return
			case <-ticker.C:
				reportCtx, reportCancel := context.WithTimeout(context.Background(), 2*interval) // nolint: mnd
				// nolint: contextcheck
				_, err := reporter.ReportLogs(reportCtx)
				reportCancel()

				if err != nil {
					errorCh <- fmt.Errorf("failed to report: %w", err)
				}
			}
		}
	}()

	return errorCh, func() error {
		cancel()

		select {
		case <-done:
			return nil
		case <-time.After(1 * time.Minute):
			return errors.New("timeout waiting for reporter to stop")
		}
	}
}
