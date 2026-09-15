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

// OperDefNameScheduledEnsureDefaultPlugin ensures each plugin package has a default plugin.
const OperDefNameScheduledEnsureDefaultPlugin = "scheduled_ensure_default_plugin"

// NewOperEnsureDefaultPlugin creates an operation for default plugin reconciliation.
func NewOperEnsureDefaultPlugin(param OperParamEnsureDefaultPlugin) operation.Definition {
	return &operScheduledEnsureDefaultPlugin{param: param}
}

type operScheduledEnsureDefaultPlugin struct {
	param OperParamEnsureDefaultPlugin
}

// OperParamEnsureDefaultPlugin defines the parameters for operScheduledEnsureDefaultPlugin.
type OperParamEnsureDefaultPlugin struct {
	utils.ScheduleActionStandardParam
}

// Name returns the name.
func (oper *operScheduledEnsureDefaultPlugin) Name() string {
	return OperDefNameScheduledEnsureDefaultPlugin
}

// ActionDefNames returns the action definition names.
func (oper *operScheduledEnsureDefaultPlugin) ActionDefNames() []string {
	return []string{syncdata.ActionNameEnsureDefaultPlugin}
}

// DefaultParameters returns the default parameters.
func (oper *operScheduledEnsureDefaultPlugin) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint:mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operScheduledEnsureDefaultPlugin) ExtraExecutionName() string {
	return OperExtraExecutionName
}
