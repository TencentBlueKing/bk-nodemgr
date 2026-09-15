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

package deploypolicy

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// OperDefNameExecuteDeployPolicy defines the operation def name.
const OperDefNameExecuteDeployPolicy = "execute_deploy_policy"

// NewOperExecuteDeployPolicy new an operation.
func NewOperExecuteDeployPolicy(param OperParamExecuteDeployPolicy) operation.Definition {
	return &operExecuteDeployPolicy{
		param: param,
	}
}

type operExecuteDeployPolicy struct {
	param OperParamExecuteDeployPolicy
}

// OperParamExecuteDeployPolicy defines the parameters for operExecuteDeployPolicy.
type OperParamExecuteDeployPolicy struct {
	TenantID        string  `json:"tenant_id"`
	Operator        string  `json:"operator"`
	DeployPolicyIDs []int64 `json:"deploy_policy_ids"`
}

// Name returns the name.
func (oper *operExecuteDeployPolicy) Name() string {
	return OperDefNameExecuteDeployPolicy
}

// ActionDefNames returns the action def names.
func (oper *operExecuteDeployPolicy) ActionDefNames() []string {
	return []string{
		ActionNameExecuteDeployPolicy,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operExecuteDeployPolicy) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     30 * time.Minute, // nolint:mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operExecuteDeployPolicy) ExtraExecutionName() string {
	return ""
}
