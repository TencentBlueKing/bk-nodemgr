/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

// Package workflow ...
package workflow

import "errors"

// RetryOperationReq ...
type RetryOperationReq struct {
	OperationID string `json:"operation_id" binding:"required"`
}

// Validate RetryOperationReq.
func (req *RetryOperationReq) Validate() error {
	if len(req.OperationID) == 0 {
		return errors.New("operation_id is required")
	}

	return nil
}

// AutoConvert auto convert.
func (req *RetryOperationReq) AutoConvert() {
}

// RetryOperationResp ...
type RetryOperationResp struct {
}
