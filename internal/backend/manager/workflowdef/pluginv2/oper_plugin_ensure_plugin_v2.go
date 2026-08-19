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

package pluginv2

import (
	"time"

	pluginV2Utils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/pluginv2/utils"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

const (
	// OperDefNamePluginEnsurePluginV2 the name of the operation definition.
	OperDefNamePluginEnsurePluginV2 = "ensure_plugin_v2"
)

// NewOperPluginEnsurePluginV2 new an operation.
func NewOperPluginEnsurePluginV2(param OperParamPluginEnsurePluginV2) operation.Definition {
	return &operPluginEnsurePluginV2{param: param}
}

type operPluginEnsurePluginV2 struct {
	param OperParamPluginEnsurePluginV2
}

// OperParamPluginEnsurePluginV2 defines the parameters for operPluginEnsurePluginV2.
type OperParamPluginEnsurePluginV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

// Name returns the name.
func (oper *operPluginEnsurePluginV2) Name() string {
	return OperDefNamePluginEnsurePluginV2
}

// ActionDefNames returns the action def names.
func (oper *operPluginEnsurePluginV2) ActionDefNames() []string {
	return []string{
		ActionNameInjectPluginBaseRuntimeV2,
		ActionNameFinishIfPluginProcessV2Alive,
		ActionNameVerifyPluginAvailabilityV2,
		ActionNameRenderPluginDeploymentV2,
		ActionNameEnsureAndUpdatePluginConfigDetailsV2,
		ActionNameRenderPluginConfigV2,
		ActionNameTransferPluginPkgToNodeV2,
		ActionNameInstallPluginV2,
		ActionNameWaitPluginInstallerCompleteV2,
		ActionNameStartProcessV2,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operPluginEnsurePluginV2) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNameInjectPluginBaseRuntimeV2:            true,
			ActionNameFinishIfPluginProcessV2Alive:         true,
			ActionNameVerifyPluginAvailabilityV2:           true,
			ActionNameRenderPluginDeploymentV2:             true,
			ActionNameEnsureAndUpdatePluginConfigDetailsV2: true,
			ActionNameRenderPluginConfigV2:                 true,
			ActionNameTransferPluginPkgToNodeV2:            true,
			ActionNameInstallPluginV2:                      true,
			ActionNameWaitPluginInstallerCompleteV2:        false,
			ActionNameStartProcessV2:                       true,
		},
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operPluginEnsurePluginV2) ExtraExecutionName() string {
	return ""
}
