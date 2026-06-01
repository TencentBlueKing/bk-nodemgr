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
	// OperDefNameSyncHostTopoRelation defines the operation def name.
	OperDefNameSyncHostTopoRelation = "sync_host_topo_relation"

	// OperDefNameSyncHostTopoRelationTimeout defines the timeout for each sync host topo relation operation.
	OperDefNameSyncHostTopoRelationTimeout = 30 * time.Minute
)

// NewOperSyncHostTopoRelation new an operation.
func NewOperSyncHostTopoRelation(param OperParamSyncHostTopoRelation) operation.Definition {
	return &operSyncHostTopoRelation{
		param: param,
	}
}

type operSyncHostTopoRelation struct {
	param OperParamSyncHostTopoRelation
}

// OperParamSyncHostTopoRelation defines the parameters for operSyncHostTopoRelation.
type OperParamSyncHostTopoRelation struct {
	BizID    int64  `json:"biz_id"`
	TenantID string `json:"tenant_id"`
	Operator string `json:"operator"`
}

// Name returns the name.
func (oper *operSyncHostTopoRelation) Name() string {
	return OperDefNameSyncHostTopoRelation
}

// ActionDefNames returns the action def names.
func (oper *operSyncHostTopoRelation) ActionDefNames() []string {
	return []string{
		ActionNameSyncHostTopoRelation,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operSyncHostTopoRelation) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     OperDefNameSyncHostTopoRelationTimeout,
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operSyncHostTopoRelation) ExtraExecutionName() string {
	return ""
}
