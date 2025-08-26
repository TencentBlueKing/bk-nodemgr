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

// OperDefNameSyncAgentState defines the operation def name.
const OperDefNameSyncAgentState = "sync_agent_state_from_gse"

// NewOperSyncAgentStateFromGSE new an operation.
func NewOperSyncAgentStateFromGSE(param OperParamSyncAgentStateFromGSE) operation.Definition {
	return &operSyncAgentStateFromGSE{
		param: param,
	}
}

type operSyncAgentStateFromGSE struct {
	param OperParamSyncAgentStateFromGSE
}

// OperParamSyncAgentStateFromGSE defines the parameters for operSyncAgentStateFromGSE.
type OperParamSyncAgentStateFromGSE struct {
	TenantID string           `json:"tenant_id"`
	Hosts    []*HostIDAgentID `json:"hosts"`
	Operator string           `json:"operator"`
}

// Name returns the name.
func (oper *operSyncAgentStateFromGSE) Name() string {
	return OperDefNameSyncAgentState
}

// ActionDefNames returns the action def names.
func (oper *operSyncAgentStateFromGSE) ActionDefNames() []string {
	return []string{
		ActionNameSyncAgentState,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operSyncAgentStateFromGSE) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint:mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operSyncAgentStateFromGSE) ExtraExecutionName() string {
	return ""
}
