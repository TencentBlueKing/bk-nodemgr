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
	CallbackSvrAddr string
}

// String step args string message.
func (args StepArgs) String() string {
	return fmt.Sprintf("token(%s), oper-inst-id(%s),  agent-id(%s), callback-svr-addr(%s)",
		args.Token, args.OperInstID, args.AgentID, args.CallbackSvrAddr)
}

// NewStep new a step to report data.
func NewStep(args StepArgs) *Step {
	return &Step{args: args}
}

// Run run the step to report data.
func (step *Step) Run(ctx context.Context) error {
	logger.Infof(node.StepReportData, "start to report data. %s", step.args.String())

	backoff := retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault())
	if err := backoff.Do(ctx, func(attempt int) error {
		if err := step.reportData(ctx); err != nil {
			logger.Infof(node.StepReportData, "retry report data. attempt(%d): %v", attempt, err)

			return err
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to report data. %s: %w", step.args.String(), err)
	}

	logger.Infof(node.StepReportData, "successfully reported data")

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

	reportDataURLPath = "/callback/workflow/node_install/report_data"
)

// ReportData report data.
func (step *Step) reportData(ctx context.Context) error {
	req := &reportDataReq{
		Token:      step.args.Token,
		OperInstID: step.args.OperInstID,
		AgentID:    step.args.AgentID,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal report data request body: %w", err)
	}

	reportURL, err := url.JoinPath(step.args.CallbackSvrAddr, reportDataURLPath)
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
