/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package datareporter ...
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

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/retrier"
)

// Step report data step.
type Step struct {
	token            string
	agentID          string
	callbackEndpoint string
}

// StepArgs define args for step.
type StepArgs struct {
	Token            string
	AgentID          string
	CallbackEndpoint string
}

// NewStep new a step to report data.
func NewStep(args StepArgs) *Step {
	step := &Step{
		token:            args.Token,
		agentID:          args.AgentID,
		callbackEndpoint: args.CallbackEndpoint,
	}

	return step
}

// Run run the step to report data.
func (step *Step) Run(ctx context.Context) error {
	logger.Infof(constant.StepReportData, "start report data")
	req := &ReportDataReq{
		Token:   step.token,
		AgentID: step.agentID,
	}

	logger.Infof(constant.StepReportData, "report data, req: %v", req)

	backoff := retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault())
	if err := backoff.Do(ctx, func(attempt int) error {
		if err := ReportData(ctx, step.callbackEndpoint, req); err != nil {
			logger.Infof(constant.StepReportData,
				"retry report data, attempt: %d, err: %v", attempt, err)

			return err
		}

		return nil
	}); err != nil {
		return fmt.Errorf("report data failed, err: %v", err)
	}

	logger.Infof(constant.StepReportData, "report data success")

	return nil
}

const reportDataTimeout = 3 * time.Second

// ReportDataReq report log req.
type ReportDataReq struct {
	Token   string `json:"token"`
	AgentID string `json:"agent_id"`
}

const reportDataURLPath = "/callback/workflow/node_install/report_data"

// ReportData report data.
func ReportData(ctx context.Context, callbackEndpoint string, req *ReportDataReq) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal report data request body failed, err: %v", err)
	}

	reportURL, err := url.JoinPath(callbackEndpoint, reportDataURLPath)
	if err != nil {
		return fmt.Errorf("format report data url failed, err: %v", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, reportURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("create report data request failed: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: reportDataTimeout}
	resp, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("do report data request failed, err: %v", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("send report data request failed, status: %d, body: %s", resp.StatusCode, body)
	}

	return nil
}
