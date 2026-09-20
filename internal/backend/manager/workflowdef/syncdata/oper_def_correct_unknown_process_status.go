/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package syncdata

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// OperDefNameCorrectUnknownProcessStatus defines the process expiry correction operation name.
const OperDefNameCorrectUnknownProcessStatus = "correct_unknown_process_status"

// NewOperCorrectUnknownProcessStatus creates a process expiry correction operation.
func NewOperCorrectUnknownProcessStatus(param OperParamCorrectUnknownProcessStatus) operation.Definition {
	return &operCorrectUnknownProcessStatus{param: param}
}

type operCorrectUnknownProcessStatus struct {
	param OperParamCorrectUnknownProcessStatus
}

// OperParamCorrectUnknownProcessStatus defines the process expiry correction scope.
type OperParamCorrectUnknownProcessStatus struct {
	TenantID string  `json:"tenant_id"`
	Operator string  `json:"operator"`
	HostIDs  []int64 `json:"host_ids"`
}

// Name returns the per-host-batch operation definition name.
func (oper *operCorrectUnknownProcessStatus) Name() string {
	return OperDefNameCorrectUnknownProcessStatus
}

// ActionDefNames returns the correction action definition.
func (oper *operCorrectUnknownProcessStatus) ActionDefNames() []string {
	return []string{
		ActionNameCorrectUnknownProcessStatus,
	}
}

// DefaultParameters returns the per-host-batch operation parameters.
func (oper *operCorrectUnknownProcessStatus) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operCorrectUnknownProcessStatus) ExtraExecutionName() string {
	return ""
}
