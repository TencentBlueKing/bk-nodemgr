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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// OperDefNameComposePluginSubConfig identifies source-to-target sub-config composition.
const OperDefNameComposePluginSubConfig = "compose_plugin_subconfig"

// OperParamComposePluginSubConfig controls the source action independently of install options.
type OperParamComposePluginSubConfig struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
	SourceOperation                       types.PluginSubConfigSourceOperation `json:"source_operation"`
}

// NewOperComposePluginSubConfig creates a sequential source and target operation.
func NewOperComposePluginSubConfig(param OperParamComposePluginSubConfig) operation.Definition {
	return &operComposePluginSubConfig{param: param}
}

type operComposePluginSubConfig struct {
	param OperParamComposePluginSubConfig
}

// Name returns the operation name.
func (oper *operComposePluginSubConfig) Name() string { return OperDefNameComposePluginSubConfig }

// ActionDefNames returns actions in dependency order.
func (oper *operComposePluginSubConfig) ActionDefNames() []string {
	return []string{ActionNamePrepareSubConfigSource, ActionNameApplyComposedPluginSubConfig}
}

// DefaultParameters returns operation parameters.
func (oper *operComposePluginSubConfig) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     20 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNamePrepareSubConfigSource:       true,
			ActionNameApplyComposedPluginSubConfig: true,
		},
	}
}

// ExtraExecutionName returns the extra execution name.
func (oper *operComposePluginSubConfig) ExtraExecutionName() string { return "" }
