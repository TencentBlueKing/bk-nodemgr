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

// OperDefNameSyncTenant defines the operation def name.
const OperDefNameSyncTenant = "sync_tenant"

// NewOperSyncTenant new an operation.
func NewOperSyncTenant(param OperParamSyncTenant) operation.Definition {
	return &operSyncTenant{
		param: param,
	}
}

type operSyncTenant struct {
	param OperParamSyncTenant
}

// OperParamSyncTenant defines the parameters for operSyncTenant.
type OperParamSyncTenant struct {
	TenantID string `json:"tenant_id"`
	Operator string `json:"operator"`
}

// Name returns the name.
func (oper *operSyncTenant) Name() string {
	return OperDefNameSyncTenant
}

// ActionDefNames returns the action def names.
func (oper *operSyncTenant) ActionDefNames() []string {
	return []string{
		ActionNameSyncTenant,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operSyncTenant) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint:mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operSyncTenant) ExtraExecutionName() string {
	return ""
}
