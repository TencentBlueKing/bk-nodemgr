/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package statusreporter this package provides the ability to report status
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

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/retrier"
)

// Step report status step.
type Step struct {
	token            string
	operInstID       string
	status           string
	callbackEndpoint string
}

// StepArgs define args for step.
type StepArgs struct {
	Token            string
	OperInstID       string
	Status           constant.State
	CallbackEndpoint string
}

// NewStep new a step to report data.
func NewStep(args StepArgs) *Step {
	step := &Step{
		token:            args.Token,
		operInstID:       args.OperInstID,
		status:           string(args.Status),
		callbackEndpoint: args.CallbackEndpoint,
	}

	return step
}

// Run run the step to report data.
func (step *Step) Run(ctx context.Context) error {
	logger.Infof(constant.StepReportStatus, "start report status. (status: %s)", step.status)
	req := &ReportStatusReq{
		Token:      step.token,
		OperInstID: step.operInstID,
		Status:     step.status,
	}

	backoff := retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault())
	if err := backoff.Do(ctx, func(attempt int) error {
		if err := ReportStatus(ctx, step.callbackEndpoint, req); err != nil {
			logger.Infof(constant.StepReportStatus,
				"retry report status, attempt: %d, err: %v", attempt, err)

			return err
		}

		return nil
	}); err != nil {
		return fmt.Errorf("report status failed, err: %v", err)
	}

	logger.Infof(constant.StepReportStatus, "report status success")

	return nil
}

// bulkReportStatusTimeout.
const bulkReportStatusTimeout = 10 * time.Second

// ReportStatusReq ...
type ReportStatusReq struct {
	Token      string `json:"token"`
	OperInstID string `json:"oper_inst_id"`
	Status     string `json:"status"`
}

// ReportStatus report status.
func ReportStatus(ctx context.Context, callbackEndpoint string, req *ReportStatusReq) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal status request failed: %v", err)
	}

	reportURL, err := url.JoinPath(callbackEndpoint, "/callback/workflow/node_install/report_status")
	if err != nil {
		return fmt.Errorf("format status URL failed: %v", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, reportURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("create status request failed: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{
		Timeout: bulkReportStatusTimeout,
	}
	retrier := retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault())

	return HTTPRequestWithRetr(ctx, request, httpClient, retrier)
}

// HTTPRequestWithRetr makes an HTTP request with retry logic.
func HTTPRequestWithRetr(ctx context.Context, request *http.Request,
	client *http.Client, retrier retrier.Retrier) error {

	return retrier.Do(ctx, func(attempt int) error {
		resp, err := client.Do(request)
		if err != nil {
			return fmt.Errorf("retry failed, attempt: %d, error: %v", attempt, err)
		}
		defer safeCloseBody(resp.Body)

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("send request failed, status: %d, body: %s\n", resp.StatusCode, body)

			return fmt.Errorf("send request failed, status: %d", resp.StatusCode)
		}

		return nil
	})
}

func safeCloseBody(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, body)
	_ = body.Close()
}
