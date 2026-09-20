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

// OperDefNameCorrectAllUnknownProcessStatus defines the tenant-wide process expiry generation operation name.
const OperDefNameCorrectAllUnknownProcessStatus = "correct_all_unknown_process_status"

// NewOperCorrectAllUnknownProcessStatus creates a tenant-wide process expiry generation operation.
func NewOperCorrectAllUnknownProcessStatus(param OperParamCorrectAllUnknownProcessStatus) operation.Definition {
	return &operCorrectAllUnknownProcessStatus{param: param}
}

type operCorrectAllUnknownProcessStatus struct {
	param OperParamCorrectAllUnknownProcessStatus
}

// OperParamCorrectAllUnknownProcessStatus defines the tenant-wide process expiry correction parameters.
type OperParamCorrectAllUnknownProcessStatus struct {
	TenantID string `json:"tenant_id"`
	Operator string `json:"operator"`
}

// Name returns the all-host operation definition name.
func (oper *operCorrectAllUnknownProcessStatus) Name() string {
	return OperDefNameCorrectAllUnknownProcessStatus
}

// ActionDefNames returns the correction action definition.
func (oper *operCorrectAllUnknownProcessStatus) ActionDefNames() []string {
	return []string{
		ActionNameGenOperCorrectUnknownProcessStatus,
	}
}

// DefaultParameters returns the all-host operation parameters.
func (oper *operCorrectAllUnknownProcessStatus) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operCorrectAllUnknownProcessStatus) ExtraExecutionName() string {
	return ""
}
