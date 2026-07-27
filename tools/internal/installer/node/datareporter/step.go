/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package datareporter provides the data report step.
package datareporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/retrier"
)

// Step report data step.
type Step struct {
	args StepArgs
}

// StepArgs define args for step.
type StepArgs struct {
	Token           string
	AgentID         string
	OperInstID      string
	CallbackSvrAddr []string
	SkipCallback    bool
	DataFilePath    string
}

// String step args string message.
func (args StepArgs) String() string {
	return fmt.Sprintf("token(%s), oper-inst-id(%s),  agent-id(%s), callback-svr-addrs(%v)",
		args.Token, args.OperInstID, args.AgentID, args.CallbackSvrAddr)
}

// NewStep new a step to report data.
func NewStep(args StepArgs) *Step {
	return &Step{args: args}
}

// dataFileContent is the JSON structure written to the local data file in skip-callback mode.
// SYNC: must stay in sync with dataFile in pkg/installer/poller/poller.go.
type dataFileContent struct {
	AgentID    string `json:"agent_id"`
	Token      string `json:"token"`
	OperInstID string `json:"oper_inst_id"`
}

// Run run the step to report data.
func (step *Step) Run(ctx context.Context) error {
	if step.args.SkipCallback {
		return step.writeDataFile()
	}

	logger.Infof(node.StepReportData, "start to report data. %s", step.args.String())

	backoff := retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault())
	if err := backoff.Do(ctx, func(attempt int) error {
		if err := step.reportDataMultiEndpoint(ctx); err != nil {
			logger.Infof(node.StepReportData, "retry multi-endpoint report data. attempt(%d): %v", attempt, err)

			return err
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to report data. %s: %w", step.args.String(), err)
	}

	logger.Infof(node.StepReportData, "successfully reported data")

	return nil
}

func (step *Step) writeDataFile() error {
	logger.Infof(node.StepReportData, "skip-callback mode: writing data to %s", step.args.DataFilePath)

	content := dataFileContent{
		AgentID:    step.args.AgentID,
		Token:      step.args.Token,
		OperInstID: step.args.OperInstID,
	}

	data, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to marshal data file: %w", err)
	}

	// nolint: gosec
	if err := os.WriteFile(step.args.DataFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write data file: %w", err)
	}

	logger.Infof(node.StepReportData, "wrote data file: agent_id=%s", step.args.AgentID)

	return nil
}

// reportDataReq report log req.
type reportDataReq struct {
	Token      string `json:"token"`
	OperInstID string `json:"oper_inst_id"`
	AgentID    string `json:"agent_id"`
}

const (
	reportDataTimeout = 3 * time.Second

	reportDataURLPath = "/api/v3/callback/workflow/node_install/report_data"
)

// reportDataMultiEndpoint tries multiple server addresses in order until one succeeds.
func (step *Step) reportDataMultiEndpoint(ctx context.Context) error {
	if len(step.args.CallbackSvrAddr) == 0 {
		return fmt.Errorf("no callback server addresses provided")
	}

	var lastErr error
	for i, callbackSvrAddr := range step.args.CallbackSvrAddr {
		logger.Infof(node.StepReportData, "attempting to report data to server, index(%d/%d), url(%s)", i+1, len(step.args.CallbackSvrAddr), callbackSvrAddr)
		err := step.reportData(ctx, callbackSvrAddr)
		if err == nil {
			return nil
		}
		logger.Warnf(node.StepReportData, "failed to report data: %v", err)
		lastErr = err
	}

	// All servers failed, return error with server addresses for debugging
	return fmt.Errorf("failed to report data to all %d server(s) %v: %w", len(step.args.CallbackSvrAddr), step.args.CallbackSvrAddr, lastErr)
}

// reportData report data to a single server.
func (step *Step) reportData(ctx context.Context, baseURL string) error {
	req := &reportDataReq{
		Token:      step.args.Token,
		OperInstID: step.args.OperInstID,
		AgentID:    step.args.AgentID,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal report data request body: %w", err)
	}

	reportURL, err := url.JoinPath(baseURL, reportDataURLPath)
	if err != nil {
		return fmt.Errorf("failed to format report data url: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, reportURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create report data request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: reportDataTimeout}
	resp, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("do report data request failed: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send report data request. status(%d), body(%s)", resp.StatusCode, body)
	}

	return nil
}
