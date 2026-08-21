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
	// OperDefNameDebugPlugin defines the debug plugin operation.
	OperDefNameDebugPlugin = "debug_plugin"
)

// NewOperDebugPlugin creates a debug plugin operation.
func NewOperDebugPlugin(param OperParamDebugPlugin) operation.Definition {
	return &operDebugPlugin{param: param}
}

// OperParamDebugPlugin defines parameters for a debug plugin operation.
type OperParamDebugPlugin struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

type operDebugPlugin struct {
	param OperParamDebugPlugin
}

// Name returns the operation name.
func (oper *operDebugPlugin) Name() string {
	return OperDefNameDebugPlugin
}

// ActionDefNames returns action definitions in execution order.
func (oper *operDebugPlugin) ActionDefNames() []string {
	return []string{
		ActionNameTryStopProcess,
		ActionNameUpsertProcess,
		ActionNameVerifyPluginAvailability,
		ActionNameInjectPluginCustomDeployConfig,
		ActionNameRenderPluginDeployment,
		ActionNameEnsureAndUpdatePluginConfigDetails,
		ActionNameRenderPluginConfig,
		ActionNameTransferPluginPkgToNode,
		ActionNameInstallPlugin,
		ActionNameWaitPluginInstallerComplete,
		ActionNamePushPluginConfig,
		ActionNamePrepareDebugProcess,
		ActionNameRunDebugPlugin,
		ActionNameCleanDebugPlugin,
		ActionNameDeleteProcess,
	}
}

// DefaultParameters returns the default operation parameters.
func (oper *operDebugPlugin) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNameTryStopProcess:                     true,
			ActionNameUpsertProcess:                      true,
			ActionNameVerifyPluginAvailability:           true,
			ActionNameInjectPluginCustomDeployConfig:     true,
			ActionNameRenderPluginDeployment:             true,
			ActionNameEnsureAndUpdatePluginConfigDetails: true,
			ActionNameRenderPluginConfig:                 true,
			ActionNameTransferPluginPkgToNode:            true,
			ActionNameInstallPlugin:                      true,
			ActionNameWaitPluginInstallerComplete:        false,
			ActionNamePushPluginConfig:                   true,
			ActionNamePrepareDebugProcess:                true,
			ActionNameRunDebugPlugin:                     false,
			ActionNameCleanDebugPlugin:                   true,
			ActionNameDeleteProcess:                      true,
		},
	}
}

// ExtraExecutionName returns the operation extra execution name.
func (oper *operDebugPlugin) ExtraExecutionName() string {
	return ""
}
