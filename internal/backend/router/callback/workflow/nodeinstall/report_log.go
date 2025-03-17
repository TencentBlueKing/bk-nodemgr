/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package nodeinstall ...
package nodeinstall

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
)

// ReportLog report agent install shell script log.
func (h *handler) ReportLog(ctx *rest.Context) (any, error) {
	req := new(ReportLogReq)
	if err := ctx.BindJSON(req); err != nil {
		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	token, err := h.parseToken(req.Token)
	if err != nil {
		return nil, fmt.Errorf("parse token failed, err: %v", err)
	}

	if token.OperInstID != req.OperInstID {
		return nil, fmt.Errorf("oper inst id not match")
	}

	for _, log := range req.Logs {
		installLog := InstallLog{
			Timestamp: time.Unix(log.Timestamp, 0),
			Level:     log.Level,
			Step:      log.Step,
			Log:       log.Log,
			Status:    log.Status,
		}

		h.logger.Infof("report log: %+v", installLog)
	}

	resp := new(ReportLogResp)

	return resp, nil
}

// parseToken ...
func (h *handler) parseToken(tokenStr string) (*Token, error) {
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

// InstallLog this is the report log.
type InstallLog struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Step      string    `json:"step"`
	Log       string    `json:"log"`
	Status    string    `json:"status"`
}
