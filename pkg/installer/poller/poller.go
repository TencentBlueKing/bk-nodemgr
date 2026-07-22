/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package poller provides installer file polling over Unix SSH.
package poller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
)

const (
	reconnectBaseBackoff   = 2 * time.Second
	reconnectMaxBackoff    = 30 * time.Second
	reconnectBackoffFactor = 2
)

// FileClient is the remote-file capability required by the installer poller.
type FileClient interface {
	ReadFile(ctx context.Context, filePath string) (string, error)
	ReadLatestFile(ctx context.Context, pattern string) (string, error)
	Close() error
}

// FileClientFactory creates remote-file clients for initial connection and reconnection.
type FileClientFactory func(context.Context) (FileClient, error)

// LogEntry is a parsed installer log line.
type LogEntry struct {
	// Timestamp is the log timestamp as Unix seconds.
	Timestamp int64

	// Level is the installer log level.
	Level string

	// Step is the installer step name.
	Step string

	// Message is the installer log message.
	Message string
}

// Config contains the installer files and polling behavior.
type Config struct {
	// NewClient creates the initial and replacement remote-file clients.
	NewClient FileClientFactory

	// StatusFile is the remote installer status file path.
	StatusFile string

	// DataFile is the remote installer data file path.
	DataFile string

	// LogGlobPath locates the newest remote installer log file.
	LogGlobPath string

	// InstanceID identifies the operation whose status and data are accepted.
	InstanceID string

	// EnsureAgentID requires a non-empty agent ID on success.
	EnsureAgentID bool

	// Interval is the delay between installer file reads.
	Interval time.Duration

	// Timeout is the maximum duration of installer polling.
	Timeout time.Duration

	// ReadTimeout bounds one remote file read.
	ReadTimeout time.Duration

	// OnLogs receives newly parsed installer log entries. Returning an error keeps
	// the current log offset so the same entries can be retried on the next poll.
	OnLogs func(context.Context, []LogEntry) error
}

// Result is the final installer polling result.
type Result struct {
	// State is the final installer process state.
	State installer.ProcessState

	// AgentID is the agent ID reported by a successful installer.
	AgentID string

	// Error is the installer-provided terminal error detail.
	Error string
}

// Wait polls installer status, data, and logs until the installer reaches a terminal state.
func Wait(ctx context.Context, config Config) (Result, error) {
	if ctx == nil {
		return Result{}, errors.New("context is nil")
	}
	if err := config.validate(); err != nil {
		return Result{}, fmt.Errorf("invalid installer polling config: %w", err)
	}

	pollCtx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	client, err := config.NewClient(pollCtx)
	if err != nil {
		return Result{}, fmt.Errorf("create installer polling file client: %w", err)
	}
	if client == nil {
		return Result{}, errors.New("create installer polling file client: client is nil")
	}

	poller := filePoller{
		config:  config,
		client:  client,
		backoff: reconnectBaseBackoff,
	}
	defer poller.close()

	return poller.wait(pollCtx)
}

func (config Config) validate() error {
	if config.NewClient == nil {
		return errors.New("new client factory is nil")
	}
	if config.StatusFile == "" {
		return errors.New("status file is empty")
	}
	if config.DataFile == "" {
		return errors.New("data file is empty")
	}
	if config.LogGlobPath == "" {
		return errors.New("log glob path is empty")
	}
	if config.InstanceID == "" {
		return errors.New("instance id is empty")
	}
	if config.Interval <= 0 {
		return errors.New("polling interval must be positive")
	}
	if config.Timeout <= 0 {
		return errors.New("polling timeout must be positive")
	}
	if config.ReadTimeout <= 0 {
		return errors.New("read timeout must be positive")
	}

	return nil
}

type filePoller struct {
	config   Config
	client   FileClient
	logState logPollingState
	backoff  time.Duration
}

func (poller *filePoller) wait(ctx context.Context) (Result, error) {
	ticker := time.NewTicker(poller.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return Result{}, ctx.Err()
		case <-ticker.C:
			result, complete, err := poller.poll(ctx)
			if err != nil {
				return Result{}, err
			}
			if complete {
				return result, nil
			}
		}
	}
}

func (poller *filePoller) poll(ctx context.Context) (Result, bool, error) {
	if poller.client == nil {
		if err := poller.reconnect(ctx); err != nil {
			if ctx.Err() != nil {
				return Result{}, false, ctx.Err()
			}

			return Result{}, false, nil
		}
	}

	statusContent, dataContent, logContent, err := poller.readFiles(ctx)
	if err != nil {
		poller.close()

		return Result{}, false, nil
	}
	poller.backoff = reconnectBaseBackoff
	poller.handleLogs(ctx, logContent)

	result := Result{}
	complete, err := pollInstallerFiles(statusContent, dataContent, poller.config, &result)
	if err != nil {
		return Result{}, false, err
	}

	return result, complete, nil
}

func (poller *filePoller) readFiles(ctx context.Context) (string, string, string, error) {
	statusContent, err := poller.readFile(ctx, poller.config.StatusFile)
	if err != nil {
		return "", "", "", err
	}

	dataContent, err := poller.readFile(ctx, poller.config.DataFile)
	if err != nil {
		return "", "", "", err
	}

	logContent, err := poller.readLatestFile(ctx, poller.config.LogGlobPath)
	if err != nil {
		return "", "", "", err
	}

	return statusContent, dataContent, logContent, nil
}

func (poller *filePoller) readFile(ctx context.Context, filePath string) (string, error) {
	readCtx, cancel := context.WithTimeout(ctx, poller.config.ReadTimeout)
	defer cancel()

	return poller.client.ReadFile(readCtx, filePath)
}

func (poller *filePoller) readLatestFile(ctx context.Context, pattern string) (string, error) {
	readCtx, cancel := context.WithTimeout(ctx, poller.config.ReadTimeout)
	defer cancel()

	return poller.client.ReadLatestFile(readCtx, pattern)
}

func (poller *filePoller) handleLogs(ctx context.Context, content string) {
	if poller.config.OnLogs == nil {
		return
	}

	entries, nextOffset := poller.logState.pendingEntries(content)
	if len(entries) == 0 {
		return
	}
	if err := poller.config.OnLogs(ctx, entries); err == nil {
		poller.logState.lineOffset = nextOffset
	}
}

func (poller *filePoller) reconnect(ctx context.Context) error {
	timer := time.NewTimer(poller.backoff)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
	}

	client, err := poller.config.NewClient(ctx)
	if err != nil {
		poller.increaseBackoff()

		return fmt.Errorf("reconnect installer polling file client: %w", err)
	}
	if client == nil {
		poller.increaseBackoff()

		return errors.New("reconnect installer polling file client: client is nil")
	}

	poller.client = client
	poller.backoff = reconnectBaseBackoff

	return nil
}

func (poller *filePoller) increaseBackoff() {
	poller.backoff = min(poller.backoff*reconnectBackoffFactor, reconnectMaxBackoff)
}

func (poller *filePoller) close() {
	if poller.client == nil {
		return
	}

	_ = poller.client.Close()
	poller.client = nil
}

type statusFile struct {
	OperInstID string `json:"oper_inst_id"`
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
}

type dataFile struct {
	AgentID    string `json:"agent_id"`
	Token      string `json:"token"`
	OperInstID string `json:"oper_inst_id"`
}

type logPollingState struct {
	lineOffset int
}

func pollInstallerFiles(statusContent string, dataContent string, config Config, result *Result) (bool, error) {
	var status statusFile
	if json.Unmarshal([]byte(strings.TrimSpace(statusContent)), &status) != nil {
		// The installer replaces this file while reporting status; retry partial writes.
		// nolint: nilerr
		return false, nil
	}
	if status.OperInstID != config.InstanceID {
		return false, nil
	}

	result.State = installer.ProcessState(status.Status)
	result.Error = status.Error
	switch result.State {
	case installer.ProcessStateSuccess:
		var data dataFile
		if err := json.Unmarshal([]byte(strings.TrimSpace(dataContent)), &data); err != nil {
			return false, fmt.Errorf("parse installer data file: %w", err)
		}
		if data.OperInstID != "" && data.OperInstID != config.InstanceID {
			return false, fmt.Errorf("installer data belongs to another operation, oper-inst-id(%s)", data.OperInstID)
		}
		if config.EnsureAgentID && data.AgentID == "" {
			return false, errors.New("installer succeeded but agent_id is empty in data file")
		}
		result.AgentID = data.AgentID

		return true, nil

	case installer.ProcessStateFailed, installer.ProcessStateTimeout:
		return true, nil

	default:
		return false, nil
	}
}

func (state *logPollingState) pendingEntries(content string) ([]LogEntry, int) {
	content = strings.TrimRight(content, "\n")
	if content == "" {
		return nil, 0
	}

	lines := strings.Split(content, "\n")
	lineOffset := state.lineOffset
	if lineOffset > len(lines) {
		lineOffset = 0
	}
	entries := make([]LogEntry, 0, len(lines)-lineOffset)
	for _, line := range lines[lineOffset:] {
		if entry := parseLogLine(line); entry.Message != "" {
			entries = append(entries, entry)
		}
	}

	return entries, len(lines)
}

func parseLogLine(line string) LogEntry {
	line = strings.TrimSpace(line)
	if line == "" {
		return LogEntry{}
	}

	parts := strings.SplitN(line, installer.LogFieldSeparator, installer.LogFieldCount)
	if len(parts) < installer.LogFieldCount {
		return LogEntry{Timestamp: time.Now().Unix(), Level: "INFO", Message: line}
	}

	timestamp, err := time.Parse("2006/01/02 15:04:05", strings.TrimSpace(parts[0]))
	if err != nil {
		return LogEntry{Timestamp: time.Now().Unix(), Level: "INFO", Message: line}
	}

	return LogEntry{
		Timestamp: timestamp.Unix(),
		Level:     strings.TrimSpace(parts[1]),
		Step:      strings.TrimSpace(parts[2]),
		Message:   strings.TrimSpace(parts[3]),
	}
}
