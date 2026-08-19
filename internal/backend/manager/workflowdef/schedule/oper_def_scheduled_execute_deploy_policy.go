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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/deploypolicy"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/schedule/utils"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// OperDefNameScheduledExecuteDeployPolicy scheduled execute deploy policy.
const OperDefNameScheduledExecuteDeployPolicy = "scheduled_execute_deploy_policy"

// NewOperExecuteDeployPolicy new an operation.
func NewOperExecuteDeployPolicy(param OperParamExecuteDeployPolicy) operation.Definition {
	return &operScheduledExecuteDeployPolicy{
		param: param,
	}
}

type operScheduledExecuteDeployPolicy struct {
	param OperParamExecuteDeployPolicy
}

// OperParamExecuteDeployPolicy defines the parameters for operScheduledExecuteDeployPolicy.
type OperParamExecuteDeployPolicy struct {
	utils.ScheduleActionStandardParam
}

// Name returns the name.
func (oper *operScheduledExecuteDeployPolicy) Name() string {
	return OperDefNameScheduledExecuteDeployPolicy
}

// ActionDefNames returns the action def names.
func (oper *operScheduledExecuteDeployPolicy) ActionDefNames() []string {
	return []string{
		deploypolicy.ActionNameGenOperExecuteDeployPolicy,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operScheduledExecuteDeployPolicy) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint:mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operScheduledExecuteDeployPolicy) ExtraExecutionName() string {
	return OperExtraExecutionName
}
