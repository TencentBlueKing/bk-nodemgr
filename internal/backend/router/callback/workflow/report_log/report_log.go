/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package report_log ...
package report_log

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types/reportlog"
)

// AgentInstall report agent install shell script log.
func (h handler) AgentInstall(ctx *rest.Context) (interface{}, error) {
	req := new(ReportInstallLogReq)
	if err := ctx.BindJSON(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	h.logger.Infof("req: %+v", req)

	//token, err := h.parseToken(req.Token)
	//if err != nil {
	//	return nil, fmt.Errorf("parse token failed, err: %v", err)
	//}

	//if token.OperInstID != req.OperInstID {
	//	return nil, fmt.Errorf("task_id not match")
	//}

	for _, log := range req.Logs {
		h.logger.Debugf("report log: %s", log)
		// TODO: 推送数据到数据库

		timestamp, err := strconv.ParseInt(log.Timestamp, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid timestamp, err: %v", err)
		}

		rpLog := reportlog.ReportLog{
			Timestamp: time.Unix(timestamp, 0),
			Level:     log.Level,
			Step:      reportlog.Step(log.Step),
			Log:       log.Log,
			Status:    log.Status,
		}

		err = h.parseLog(rpLog)
		if err != nil {
			return nil, err
		}
	}

	resp := new(ReportLogResp)

	return resp, nil
}

// parseToken ...
func (h handler) parseToken(tokenStr string) (*Token, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(tokenStr)
	if err != nil {
		return nil, err
	}

	jsonBytes, err := h.crypter.Decrypt(ciphertext)
	if err != nil {
		return nil, err
	}

	token := new(Token)

	if err = json.Unmarshal(jsonBytes, token); err != nil {
		return nil, err
	}

	return token, nil
}

// Token the token for callback.
type Token struct {
	OperInstID string `json:"oper_inst_id"`
	BKHostID   string `json:"bk_host_id"`
	InnerIP    string `json:"inner_ip"`
	BKCloudID  string `json:"bk_cloud_id"`
	Timestamp  string `json:"timestamp"`
	InstID     string `json:"inst_id"`
	HostApID   string `json:"host_ap_id"`
}

// parseLog
func (h handler) parseLog(log reportlog.ReportLog) error {
	switch log.Step {
	case reportlog.StepCheckEnv:
		h.logger.Infof("check env log: %s", log)
	case reportlog.StepDownloadPkg:
		h.logger.Infof("download pkg log: %s", log)
	case reportlog.StepRemoveAgent:
		h.logger.Infof("remove agent log: %s", log)
	case reportlog.StepRemoveProxyIfExists:
		h.logger.Infof("remove proxy if exists log: %s", log)
	case reportlog.StepSetupAgent:
		h.logger.Infof("setup agent log: %s", log)
	case reportlog.StepCheckDeployResult:
		h.logger.Infof("check deploy result log: %s", log)
	default:
		return fmt.Errorf("invalid step in agent install shell script, step(%s)", log.Step)
	}

	return nil
}
