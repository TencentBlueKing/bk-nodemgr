/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package callback ...
package callback

import (
	"errors"
	"fmt"
)

// Validate check request body.
func (x *GetAgentConfReq) Validate() error {
	if x.GetToken() == "" {
		return errors.New("token is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *GetAgentConfReq) AutoConvert() {
}

// Validate check request body.
func (x *GetDataProxyConfReq) Validate() error {
	if x.GetToken() == "" {
		return errors.New("token is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *GetDataProxyConfReq) AutoConvert() {
}

// Validate check request body.
func (x *GetFileProxyConfReq) Validate() error {
	if x.GetToken() == "" {
		return errors.New("token is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *GetFileProxyConfReq) AutoConvert() {
}

// Validate check request body.
func (x *GetCheckListReq) Validate() error {
	if x.GetToken() == "" {
		return errors.New("token is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *GetCheckListReq) AutoConvert() {
}

// Validate check request body.
func (x *ReportLogReq) Validate() error {
	if x.GetToken() == "" {
		return errors.New("token is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *ReportLogReq) AutoConvert() {
}

// AutoConvert auto convert.
func (x *ReportStatusReq) AutoConvert() {
}

// Validate check request body.
func (x *ReportStatusReq) Validate() error {
	if x.GetToken() == "" {
		return errors.New("token is required")
	}

	return nil
}

// Validate check request body.
func (x *ReportDataReq) Validate() error {
	if x.GetToken() == "" {
		return errors.New("token is required")
	}

	if x.GetAgentId() == "" {
		return errors.New("agent_id is required")
	}

	return nil
}

// AutoConvert auto convert.
func (x *ReportDataReq) AutoConvert() {
}

// Validate check request body.
func (x *ReportFileStateReq) Validate() error {
	if x.GetActionName() == "" {
		return errors.New("action_name is required")
	}
	if x.GetOperInstId() == "" {
		return errors.New("oper_inst_id is required")
	}
	if len(x.GetFileState()) == 0 {
		return errors.New("file_name is required")
	}

	if err := x.checkStatus(); err != nil {
		return fmt.Errorf("check status failed, err: %w", err)
	}

	return nil
}

// TODO: check with relayProto.
func (x *ReportFileStateReq) checkStatus() error {
	return errors.New("implement me")
}

// AutoConvert auto convert.
func (x *ReportFileStateReq) AutoConvert() {
}
