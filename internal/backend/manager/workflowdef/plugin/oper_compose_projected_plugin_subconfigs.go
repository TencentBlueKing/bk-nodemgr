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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// OperDefNameComposeProjectedPluginSubConfigs identifies a single-host remote projection batch.
const OperDefNameComposeProjectedPluginSubConfigs = "compose_projected_plugin_subconfigs"

// OperParamComposeProjectedPluginSubConfigs persists the ordered source deployments of a batch.
// No individual source process represents the batch itself.
type OperParamComposeProjectedPluginSubConfigs struct {
	TenantID        string                               `json:"tenant_id"`
	Operator        string                               `json:"operator"`
	HostID          int64                                `json:"host_id"`
	SourceTokens    []string                             `json:"source_tokens"`
	SourceOperation types.PluginSubConfigSourceOperation `json:"source_operation"`
}

// NewOperComposeProjectedPluginSubConfigs creates a batch source preparation and sequential target operation.
func NewOperComposeProjectedPluginSubConfigs(param OperParamComposeProjectedPluginSubConfigs) operation.Definition {
	return &operComposeProjectedPluginSubConfigs{param: param}
}

type operComposeProjectedPluginSubConfigs struct {
	param OperParamComposeProjectedPluginSubConfigs
}

// Name returns the operation name.
func (oper *operComposeProjectedPluginSubConfigs) Name() string {
	return OperDefNameComposeProjectedPluginSubConfigs
}

// ActionDefNames returns actions in dependency order.
func (oper *operComposeProjectedPluginSubConfigs) ActionDefNames() []string {
	return []string{ActionNamePrepareProjectionSources, ActionNameApplyProjectedPluginSubConfigs}
}

// DefaultParameters returns the persisted batch parameters.
func (oper *operComposeProjectedPluginSubConfigs) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     20 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNamePrepareProjectionSources:       true,
			ActionNameApplyProjectedPluginSubConfigs: true,
		},
	}
}

// ExtraExecutionName returns the extra execution name.
func (oper *operComposeProjectedPluginSubConfigs) ExtraExecutionName() string { return "" }
