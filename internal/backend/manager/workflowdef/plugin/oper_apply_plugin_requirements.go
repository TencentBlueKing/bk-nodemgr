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
 * either express or implied. See the License for the specific
 * language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package plugin

import (
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

const (
	// OperDefNameApplyPluginRequirements applies dependency files and restarts a process.
	OperDefNameApplyPluginRequirements = "apply_plugin_requirements"
)

// NewOperApplyPluginRequirements creates an operation to apply dependencies and restart a process.
func NewOperApplyPluginRequirements(param OperParamApplyPluginRequirements) operation.Definition {
	return &operApplyPluginRequirements{param: param}
}

type operApplyPluginRequirements struct {
	param OperParamApplyPluginRequirements
}

// OperParamApplyPluginRequirements defines the operation parameters.
type OperParamApplyPluginRequirements struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// Name returns the operation name.
func (oper *operApplyPluginRequirements) Name() string {
	return OperDefNameApplyPluginRequirements
}

// ActionDefNames returns the ordered actions.
func (oper *operApplyPluginRequirements) ActionDefNames() []string {
	return []string{
		ActionNameFetchPluginProcess,
		ActionNameInjectPluginCustomDeployConfig,
		ActionNameCheckPluginRequirements,
		ActionNameApplyPluginRequirements,
		ActionNameRestartProcess,
		ActionNameCheckPluginProcessAlive,
		ActionNameUpdateProcess,
	}
}

// DefaultParameters returns the operation defaults.
func (oper *operApplyPluginRequirements) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     20 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNameFetchPluginProcess:             true,
			ActionNameInjectPluginCustomDeployConfig: true,
			ActionNameCheckPluginRequirements:        true,
			ActionNameApplyPluginRequirements:        true,
			ActionNameRestartProcess:                 true,
			ActionNameCheckPluginProcessAlive:        true,
			ActionNameUpdateProcess:                  true,
		},
	}
}

// ExtraExecutionName returns the optional extra execution name.
func (oper *operApplyPluginRequirements) ExtraExecutionName() string {
	return ""
}
