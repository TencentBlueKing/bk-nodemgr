/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package statusreporter this package provides the ability to report status.
package statusreporter

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
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/types"
)

// Step report status step.
type Step struct {
	args StepArgs
}

// StepArgs define args for step.
type StepArgs struct {
	Token           string
	OperInstID      string
	Status          types.ProcessState
	CallbackSvrAddr []string
}

// String step args string message.
func (args StepArgs) String() string {
	return fmt.Sprintf("token(%s), oper-inst-id(%s), status(%s), callback-svr-addrs(%v)",
		args.Token, args.OperInstID, args.Status, args.CallbackSvrAddr)
}

// NewStep new a step to report data.
func NewStep(args StepArgs) *Step {
	return &Step{args: args}
}

// Run run the step to report data.
func (step *Step) Run(ctx context.Context) error {
	logger.Infof(node.StepReportStatus, "start to report status: %s", step.args.String())

	backoff := retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault())
	if err := backoff.Do(ctx, func(attempt int) error {
		if err := step.reportStatusMultiEndpoint(ctx); err != nil {
			logger.Errorf(node.StepReportStatus,
				"failed to retry multi-endpoint report status. attempt(%d): %v", attempt, err)

			return err
		}

		return nil
	}); err != nil {
		logger.Errorf(node.StepReportStatus, "failed to report status: %v", err)
		return fmt.Errorf("failed to report status: %w", err)
	}

	logger.Infof(node.StepReportStatus, "reported status")

	return nil
}

const (
	reportStatusTimeout = 10 * time.Second
	reportStatusPath    = "/api/v3/callback/workflow/node_install/report_status"
)

// reportStatusMultiEndpoint tries multiple server addresses in order until one succeeds.
func (step *Step) reportStatusMultiEndpoint(ctx context.Context) error {
	if len(step.args.CallbackSvrAddr) == 0 {
		return fmt.Errorf("no callback server addresses provided")
	}

	var lastErr error
	for i, callbackSvrAddr := range step.args.CallbackSvrAddr {
		logger.Infof(node.StepReportStatus, "attempting to report status to server, index(%d/%d), url(%s)", i+1, len(step.args.CallbackSvrAddr), callbackSvrAddr)
		err := step.reportStatus(ctx, callbackSvrAddr)
		if err == nil {
			return nil
		}
		logger.Warnf(node.StepReportStatus, "failed to report status: %v", err)
		lastErr = err
	}

	// All servers failed, return error with server addresses for debugging
	return fmt.Errorf("failed to report status to all %d server(s) %v: %w", len(step.args.CallbackSvrAddr), step.args.CallbackSvrAddr, lastErr)
}

func (step *Step) reportStatus(ctx context.Context, baseURL string) error {
	type reportStatusReq struct {
		Token      string `json:"token"`
		OperInstID string `json:"oper_inst_id"`
		Status     string `json:"status"`
	}

	req := &reportStatusReq{
		Token:      step.args.Token,
		OperInstID: step.args.OperInstID,
		Status:     string(step.args.Status),
	}
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal status request: %w", err)
	}

	reportURL, err := url.JoinPath(baseURL, reportStatusPath)
	if err != nil {
		return fmt.Errorf("failed to format status URL: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, reportURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("create status request failed: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{
		Timeout: reportStatusTimeout,
	}
	resp, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send status request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Errorf(node.StepReportStatus, "failed to send request to report status. resp-code(%d), resp-body(%s)",
			resp.StatusCode, body)

		return fmt.Errorf("failed to send status request. resp-code(%d)", resp.StatusCode)
	}

	return nil
}
