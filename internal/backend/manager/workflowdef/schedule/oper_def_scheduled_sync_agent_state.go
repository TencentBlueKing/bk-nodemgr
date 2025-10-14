/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package schedule

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/schedule/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// OperDefNameScheduledSyncAgentState scheduled sync agent state.
const OperDefNameScheduledSyncAgentState = "scheduled_sync_agent_state"

// NewOperSyncAgentState new an operation.
func NewOperSyncAgentState(param OperParamSyncAgentState) operation.Definition {
	return &operScheduledSyncAgentState{
		param: param,
	}
}

type operScheduledSyncAgentState struct {
	param OperParamSyncAgentState
}

// OperParamSyncAgentState defines the parameters for operScheduledSyncAgentState.
type OperParamSyncAgentState struct {
	utils.ScheduleActionStandardParam
}

// Name returns the name.
func (oper *operScheduledSyncAgentState) Name() string {
	return OperDefNameScheduledSyncAgentState
}

// ActionDefNames returns the action def names.
func (oper *operScheduledSyncAgentState) ActionDefNames() []string {
	return []string{
		syncdata.ActionNameGenOperSyncAgentState,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operScheduledSyncAgentState) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint:mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operScheduledSyncAgentState) ExtraExecutionName() string {
	return OperExtraExecutionName
}
