/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package syncdata

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

const (
	// OperDefNameSyncAgentState defines the operation def name.
	OperDefNameSyncAgentState = "sync_agent_state"

	// OperDefNameSyncAgentStateTimeout defines the timeout for each sync agent state operation.
	OperDefNameSyncAgentStateTimeout = 10 * time.Minute
)

// NewOperSyncAgentState new an operation.
func NewOperSyncAgentState(param OperParamSyncAgentState) operation.Definition {
	return &operSyncAgentState{
		param: param,
	}
}

type operSyncAgentState struct {
	param OperParamSyncAgentState
}

// OperParamSyncAgentState defines the parameters for operSyncAgentState.
type OperParamSyncAgentState struct {
	TenantID            string           `json:"tenant_id"`
	Hosts               []*HostIDAgentID `json:"hosts"`
	Operator            string           `json:"operator"`
	CompareCurrentState bool             `json:"compare_current_state,omitempty"`
}

// Name returns the name.
func (oper *operSyncAgentState) Name() string {
	return OperDefNameSyncAgentState
}

// ActionDefNames returns the action def names.
func (oper *operSyncAgentState) ActionDefNames() []string {
	return []string{
		ActionNameSyncAgentState,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operSyncAgentState) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     OperDefNameSyncAgentStateTimeout,
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operSyncAgentState) ExtraExecutionName() string {
	return ""
}
