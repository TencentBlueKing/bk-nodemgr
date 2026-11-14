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
	// OperDefNameInstallPluginNoStart the name of the operation definition, it delivers the plugin
	// package and configs to the host, but does not start the plugin process.
	OperDefNameInstallPluginNoStart = "install_plugin_no_start"
)

// NewOperInstallPluginNoStart new an operation.
func NewOperInstallPluginNoStart(param OperParamInstallPluginNoStart) operation.Definition {
	return &operInstallPluginNoStart{param: param}
}

type operInstallPluginNoStart struct {
	param OperParamInstallPluginNoStart
}

// OperParamInstallPluginNoStart defines the parameters for operInstallPluginNoStart.
type OperParamInstallPluginNoStart struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// Name returns the name.
func (oper *operInstallPluginNoStart) Name() string {
	return OperDefNameInstallPluginNoStart
}

// ActionDefNames returns the action def names.
//
// Unlike install_plugin, this operation:
//   - does not check that the v2 plugin process is not alive(check_v2_plugin_no_alive);
//   - does not render the compatibility config(overwrite_plugin_config_for_compatibility);
//   - does not start the plugin process(start_process), instead it marks the process as stopped.
func (oper *operInstallPluginNoStart) ActionDefNames() []string {
	return []string{
		ActionNameTryStopProcess,
		ActionNameUpsertProcess,
		ActionNameVerifyPluginAvailability,
		ActionNameInjectPluginCustomDeployConfig,
		ActionNameInjectMainConfigDetail,
		ActionNameRenderPluginDeployment,
		ActionNameAllocatePluginBindAddress,
		ActionNameEnsureAndUpdatePluginConfigDetails,
		ActionNameRenderPluginConfig,
		ActionNameTransferPluginPkgToNode,
		ActionNameInstallPlugin,
		ActionNameWaitPluginInstallerComplete,
		ActionNameMarkPluginProcessStopped,
		ActionNameUpdateProcess,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operInstallPluginNoStart) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNameTryStopProcess:                     true,
			ActionNameUpsertProcess:                      true,
			ActionNameVerifyPluginAvailability:           true,
			ActionNameInjectPluginCustomDeployConfig:     true,
			ActionNameInjectMainConfigDetail:             true,
			ActionNameRenderPluginDeployment:             true,
			ActionNameAllocatePluginBindAddress:          true,
			ActionNameEnsureAndUpdatePluginConfigDetails: true,
			ActionNameRenderPluginConfig:                 true,
			ActionNameTransferPluginPkgToNode:            true,
			ActionNameInstallPlugin:                      true,
			ActionNameWaitPluginInstallerComplete:        false,
			ActionNameMarkPluginProcessStopped:           true,
			ActionNameUpdateProcess:                      true,
		},
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operInstallPluginNoStart) ExtraExecutionName() string {
	return ""
}
