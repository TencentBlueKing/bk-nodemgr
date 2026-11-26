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

// OperDefNameCorrectRunningProcessV2Status defines the V2 running-process expiry operation name.
const OperDefNameCorrectRunningProcessV2Status = "correct_running_process_v2_status"

// NewOperCorrectRunningProcessV2Status creates a V2 running-process expiry operation.
func NewOperCorrectRunningProcessV2Status(param OperParamCorrectRunningProcessV2Status) operation.Definition {
	return &operCorrectRunningProcessV2Status{param: param}
}

type operCorrectRunningProcessV2Status struct {
	param OperParamCorrectRunningProcessV2Status
}

// OperParamCorrectRunningProcessV2Status defines the V2 expiry check scope.
type OperParamCorrectRunningProcessV2Status struct {
	TenantID string  `json:"tenant_id"`
	Operator string  `json:"operator"`
	HostIDs  []int64 `json:"host_ids"`
}

// Name returns the operation name.
func (oper *operCorrectRunningProcessV2Status) Name() string {
	return OperDefNameCorrectRunningProcessV2Status
}

// ActionDefNames returns the V2 expiry action name.
func (oper *operCorrectRunningProcessV2Status) ActionDefNames() []string {
	return []string{ActionNameCorrectRunningProcessV2Status}
}

// DefaultParameters returns the operation parameters.
func (oper *operCorrectRunningProcessV2Status) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operCorrectRunningProcessV2Status) ExtraExecutionName() string {
	return ""
}
