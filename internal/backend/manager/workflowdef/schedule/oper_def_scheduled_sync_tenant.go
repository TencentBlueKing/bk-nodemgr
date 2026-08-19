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

package schedule

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/schedule/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// OperDefNameScheduledSyncTenant scheduled sync all tenant.
const OperDefNameScheduledSyncTenant = "scheduled_sync_tenant"

// NewOperSyncTenant new an operation.
func NewOperSyncTenant(param OperParamSyncTenant) operation.Definition {
	return &operScheduledSyncTenant{
		param: param,
	}
}

type operScheduledSyncTenant struct {
	param OperParamSyncTenant
}

// OperParamSyncTenant defines the parameters for operScheduledSyncTenant.
type OperParamSyncTenant struct {
	utils.ScheduleActionStandardParam
}

// Name returns the name.
func (oper *operScheduledSyncTenant) Name() string {
	return OperDefNameScheduledSyncTenant
}

// ActionDefNames returns the action def names.
func (oper *operScheduledSyncTenant) ActionDefNames() []string {
	return []string{
		syncdata.ActionNameSyncTenant,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operScheduledSyncTenant) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint:mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operScheduledSyncTenant) ExtraExecutionName() string {
	return OperExtraExecutionName
}
