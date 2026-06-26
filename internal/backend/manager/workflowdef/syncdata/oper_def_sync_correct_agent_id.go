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
	// OperDefNameSyncCorrectAgentID defines the operation def name.
	OperDefNameSyncCorrectAgentID = "sync_correct_agent_id"

	// OperDefNameSyncCorrectAgentIDTimeout defines the timeout for each sync correct agent id operation.
	OperDefNameSyncCorrectAgentIDTimeout = 10 * time.Minute
)

// NewOperSyncCorrectAgentID new an operation.
func NewOperSyncCorrectAgentID(param OperParamSyncCorrectAgentID) operation.Definition {
	return &operSyncCorrectAgentID{
		param: param,
	}
}

type operSyncCorrectAgentID struct {
	param OperParamSyncCorrectAgentID
}

// OperParamSyncCorrectAgentID defines the parameters for operSyncCorrectAgentID.
type OperParamSyncCorrectAgentID struct {
	TenantID string  `json:"tenant_id"`
	Operator string  `json:"operator"`
	HostIDs  []int64 `json:"host_ids"`
}

// Name returns the name.
func (oper *operSyncCorrectAgentID) Name() string {
	return OperDefNameSyncCorrectAgentID
}

// ActionDefNames returns the action def names.
func (oper *operSyncCorrectAgentID) ActionDefNames() []string {
	return []string{
		ActionNameSyncCorrectAgentID,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operSyncCorrectAgentID) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     OperDefNameSyncCorrectAgentIDTimeout,
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operSyncCorrectAgentID) ExtraExecutionName() string {
	return ""
}
