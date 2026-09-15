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

// OperDefNameEnsureDefaultPlugin defines the operation def name.
const OperDefNameEnsureDefaultPlugin = "ensure_default_plugin"

// NewOperEnsureDefaultPlugin creates a one-time default plugin reconciliation operation.
func NewOperEnsureDefaultPlugin(param OperParamEnsureDefaultPlugin) operation.Definition {
	return &operEnsureDefaultPlugin{
		param: param,
	}
}

type operEnsureDefaultPlugin struct {
	param OperParamEnsureDefaultPlugin
}

// OperParamEnsureDefaultPlugin defines the parameters for operEnsureDefaultPlugin.
type OperParamEnsureDefaultPlugin struct {
	TenantID string `json:"tenant_id"`
	Operator string `json:"operator"`
}

// Name returns the name.
func (oper *operEnsureDefaultPlugin) Name() string {
	return OperDefNameEnsureDefaultPlugin
}

// ActionDefNames returns the action def names.
func (oper *operEnsureDefaultPlugin) ActionDefNames() []string {
	return []string{
		ActionNameEnsureDefaultPlugin,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operEnsureDefaultPlugin) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint:mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operEnsureDefaultPlugin) ExtraExecutionName() string {
	return ""
}
