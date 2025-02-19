/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package agent ...
package agent

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
)

// ReportLog report shell log.
func (h handler) ReportLog(ctx *rest.Context) (interface{}, error) {
	req := new(ReportLogReq)
	if err := ctx.BindJSON(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	token, err := h.parseToken(req.Token)
	if err != nil {
		return nil, fmt.Errorf("parse token failed, err: %v", err)
	}

	if token.TaskID != req.TaskID {
		return nil, fmt.Errorf("task_id not match")
	}

	for _, log := range req.Logs {
		h.logger.Debugf("report log: %s", log)
		// TODO: 推送数据到数据库
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

type Token struct {
	TaskID    string `json:"task_id"`
	BKHostID  string `json:"bk_host_id"`
	InnerIP   string `json:"inner_ip"`
	BKCloudID string `json:"bk_cloud_id"`
	Timestamp string `json:"timestamp"`
	InstID    string `json:"inst_id"`
	HostApID  string `json:"host_ap_id"`
}
