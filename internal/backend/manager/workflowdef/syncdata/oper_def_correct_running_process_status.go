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

package syncdata

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// OperDefNameCorrectRunningProcessStatus defines the V3 running-process expiry operation name.
const OperDefNameCorrectRunningProcessStatus = "correct_running_process_status"

// NewOperCorrectRunningProcessStatus creates a V3 running-process expiry operation.
func NewOperCorrectRunningProcessStatus(param OperParamCorrectRunningProcessStatus) operation.Definition {
	return &operCorrectRunningProcessStatus{param: param}
}

type operCorrectRunningProcessStatus struct {
	param OperParamCorrectRunningProcessStatus
}

// OperParamCorrectRunningProcessStatus defines the V3 expiry check scope.
type OperParamCorrectRunningProcessStatus struct {
	TenantID string  `json:"tenant_id"`
	Operator string  `json:"operator"`
	HostIDs  []int64 `json:"host_ids"`
}

// Name returns the operation name.
func (oper *operCorrectRunningProcessStatus) Name() string {
	return OperDefNameCorrectRunningProcessStatus
}

// ActionDefNames returns the V3 expiry action name.
func (oper *operCorrectRunningProcessStatus) ActionDefNames() []string {
	return []string{ActionNameCorrectRunningProcessStatus}
}

// DefaultParameters returns the operation parameters.
func (oper *operCorrectRunningProcessStatus) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operCorrectRunningProcessStatus) ExtraExecutionName() string {
	return ""
}
