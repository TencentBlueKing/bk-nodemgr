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

package schedule

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/schedule/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// OperDefNameScheduledCorrectAllUnknownProcessStatus defines the scheduled operation name.
const OperDefNameScheduledCorrectAllUnknownProcessStatus = "scheduled_correct_all_unknown_process_status"

// NewOperCorrectAllUnknownProcessStatus creates the scheduled unknown-process correction operation.
func NewOperCorrectAllUnknownProcessStatus(param OperParamCorrectAllUnknownProcessStatus) operation.Definition {
	return &operCorrectAllUnknownProcessStatus{param: param}
}

type operCorrectAllUnknownProcessStatus struct {
	param OperParamCorrectAllUnknownProcessStatus
}

// OperParamCorrectAllUnknownProcessStatus defines scheduled operation parameters.
type OperParamCorrectAllUnknownProcessStatus struct {
	utils.ScheduleActionStandardParam
}

// Name returns the scheduled operation name.
func (oper *operCorrectAllUnknownProcessStatus) Name() string {
	return OperDefNameScheduledCorrectAllUnknownProcessStatus
}

// ActionDefNames returns the process expiry operation generation action.
func (oper *operCorrectAllUnknownProcessStatus) ActionDefNames() []string {
	return []string{syncdata.ActionNameGenOperCorrectUnknownProcessStatus}
}

// DefaultParameters returns the scheduled operation parameters.
func (oper *operCorrectAllUnknownProcessStatus) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the schedule extra execution definition name.
func (oper *operCorrectAllUnknownProcessStatus) ExtraExecutionName() string {
	return OperExtraExecutionName
}
