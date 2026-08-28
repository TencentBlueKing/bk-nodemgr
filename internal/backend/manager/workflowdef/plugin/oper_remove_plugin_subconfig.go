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

package plugin

import (
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

const (
	// OperDefNameRemovePluginSubConfig the name of the operation definition.
	OperDefNameRemovePluginSubConfig = "remove_plugin_subconfig"
)

// NewOperRemovePluginSubConfig new an operation.
func NewOperRemovePluginSubConfig(param OperParamRemovePluginSubConfig) operation.Definition {
	return &operRemovePluginSubConfig{param: param}
}

type operRemovePluginSubConfig struct {
	param OperParamRemovePluginSubConfig
}

// OperParamRemovePluginSubConfig defines the parameters for operRemovePluginSubConfig.
type OperParamRemovePluginSubConfig struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// Name returns the name.
func (oper *operRemovePluginSubConfig) Name() string {
	return OperDefNameRemovePluginSubConfig
}

// ActionDefNames returns the action def names.
func (oper *operRemovePluginSubConfig) ActionDefNames() []string {
	return []string{
		ActionNameFetchPluginProcess,
		ActionNameVerifyPluginAvailability,
		ActionNameCheckPluginProcessAlive,
		ActionNameInjectPluginCustomDeployConfig,
		ActionNameRemovePluginSubConfig,
		ActionNameReloadProcess,
		ActionNameDeleteProcessConfigRecord,
		ActionNameUpdateProcess,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operRemovePluginSubConfig) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNameFetchPluginProcess:             true,
			ActionNameVerifyPluginAvailability:       true,
			ActionNameCheckPluginProcessAlive:        true,
			ActionNameInjectPluginCustomDeployConfig: true,
			ActionNameRemovePluginSubConfig:          true,
			ActionNameReloadProcess:                  true,
			ActionNameDeleteProcessConfigRecord:      true,
			ActionNameUpdateProcess:                  true,
		},
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operRemovePluginSubConfig) ExtraExecutionName() string {
	return ""
}
