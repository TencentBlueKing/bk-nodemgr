/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package plugin

import (
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

const (
	// OperDefNameMigrateProcessFromPluginV2 the name of the operation definition.
	OperDefNameMigrateProcessFromPluginV2 = "migrate_process_from_plugin_v2"
)

// NewOperMigratePluginProcessFromV2 new an operation.
func NewOperMigratePluginProcessFromV2(param OperParamMigratePluginProcessFromV2) operation.Definition {
	return &operMigrateProcessFromPluginV2{param: param}
}

type operMigrateProcessFromPluginV2 struct {
	param OperParamMigratePluginProcessFromV2
}

// OperParamMigratePluginProcessFromV2 defines the parameters for operMigrateProcessFromPluginV2.
type OperParamMigratePluginProcessFromV2 struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// Name returns the name.
func (oper *operMigrateProcessFromPluginV2) Name() string {
	return OperDefNameMigrateProcessFromPluginV2
}

// ActionDefNames returns the action def names.
func (oper *operMigrateProcessFromPluginV2) ActionDefNames() []string {
	return []string{
		ActionNameFetchPluginProcess,
		ActionNameInjectPluginCustomDeployConfig,
		ActionNameVerifyPluginAvailability,
		ActionNameRenderPluginDeployment,
		ActionNameEnsureAndUpdatePluginConfigDetails,
		ActionNameRenderPluginConfig,
		ActionNamePushPluginConfig,
		ActionNameStopPluginV2Process,
		ActionNameRestartProcess,
		ActionNameUpdateProcess,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operMigrateProcessFromPluginV2) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNameFetchPluginProcess:                 true,
			ActionNameInjectPluginCustomDeployConfig:     true,
			ActionNameVerifyPluginAvailability:           true,
			ActionNameRenderPluginDeployment:             true,
			ActionNameEnsureAndUpdatePluginConfigDetails: true,
			ActionNameRenderPluginConfig:                 true,
			ActionNamePushPluginConfig:                   true,
			ActionNameStopPluginV2Process:                true,
			ActionNameRestartProcess:                     true,
			ActionNameUpdateProcess:                      true,
		},
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operMigrateProcessFromPluginV2) ExtraExecutionName() string {
	return ""
}
