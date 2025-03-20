/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodeinstall

import (
	"errors"
	"fmt"
)

// ReportLogReq the api of InstallLog's req.
type ReportLogReq struct {
	OperInstID string `json:"oper_inst_id"`
	Token      string `json:"token"`
	Logs       []struct {
		Timestamp int64  `json:"timestamp"`
		Level     string `json:"level"`
		Step      string `json:"step"`
		Log       string `json:"log"`
		Data      string `json:"data"`
		Status    string `json:"status"`
	} `json:"logs"`
}

// Validate ReportLogReq.
func (req ReportLogReq) Validate() error {
	if req.OperInstID == "" {
		return fmt.Errorf("oper_inst_id is empty")
	}

	if req.Token == "" {
		return errors.New("token is empty")
	}

	return nil
}

// AutoConvert auto convert.
func (req ReportLogReq) AutoConvert() {
}

// ReportLogResp the api of InstallLog's resp.
type ReportLogResp struct {
}

// ReportDataReq the api of ReportData's req.
type ReportDataReq struct {
	OSType  string `json:"os_type" validate:"required"`
	CPUArch string `json:"cpu_arch" validate:"required"`
	AgentID string `json:"agent_id" validate:"required"`
}

// Validate ReportDataReq.
func (req ReportDataReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (req ReportDataReq) AutoConvert() {
}

// ReportDataResp the api of ReportData's resp.
type ReportDataResp struct {
}
