/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package report_log

import "fmt"

// ReportInstallLogReq AgentInstall req.
type ReportInstallLogReq struct {
	OperInstID string `json:"oper_inst_id"`
	Token      string `json:"token"`
	Logs       []struct {
		Timestamp string `json:"timestamp"`
		Level     string `json:"level"`
		Step      string `json:"step"`
		Log       string `json:"log"`
		Status    string `json:"status"`
	} `json:"logs"`
}

// Validate ReportInstallLogReq.
func (req ReportInstallLogReq) Validate() error {
	if req.OperInstID == "" {
		return fmt.Errorf("oper_inst_id is empty")
	}

	if req.Token == "" {
		return fmt.Errorf("token is empty")
	}

	return nil
}

// ReportLogResp AgentInstall resp.
type ReportLogResp struct {
}
