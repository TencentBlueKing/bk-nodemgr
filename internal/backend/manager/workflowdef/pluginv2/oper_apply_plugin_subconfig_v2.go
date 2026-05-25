/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pluginv2

import (
	"time"

	pluginV2Utils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/pluginv2/utils"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

const (
	// OperDefNameApplyPluginSubConfigV2 the name of the operation definition.
	OperDefNameApplyPluginSubConfigV2 = "apply_plugin_subconfig_v2"
)

// NewOperApplyPluginSubConfigV2 new an operation.
func NewOperApplyPluginSubConfigV2(param OperParamApplyPluginSubConfigV2) operation.Definition {
	return &operApplyPluginSubConfigV2{param: param}
}

type operApplyPluginSubConfigV2 struct {
	param OperParamApplyPluginSubConfigV2
}

// OperParamApplyPluginSubConfigV2 defines the parameters for operApplyPluginSubConfigV2.
type OperParamApplyPluginSubConfigV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

// Name returns the name.
func (oper *operApplyPluginSubConfigV2) Name() string {
	return OperDefNameApplyPluginSubConfigV2
}

// ActionDefNames returns the action def names.
func (oper *operApplyPluginSubConfigV2) ActionDefNames() []string {
	return []string{
		ActionNameFetchPluginProcessV2,
		ActionNameVerifyPluginAvailabilityV2,
		ActionNameCheckPluginProcessAliveV2,
		ActionNameInjectPluginBaseRuntimeV2,
		ActionNameRenderPluginDeploymentV2,
		ActionNameEnsureAndUpdatePluginConfigDetailsV2,
		ActionNameRenderPluginConfigV2,
		ActionNamePushPluginConfigV2,
		ActionNameReloadProcessV2,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operApplyPluginSubConfigV2) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNameFetchPluginProcessV2:                 true,
			ActionNameVerifyPluginAvailabilityV2:           true,
			ActionNameCheckPluginProcessAliveV2:            true,
			ActionNameInjectPluginBaseRuntimeV2:            true,
			ActionNameRenderPluginDeploymentV2:             true,
			ActionNameEnsureAndUpdatePluginConfigDetailsV2: true,
			ActionNameRenderPluginConfigV2:                 true,
			ActionNamePushPluginConfigV2:                   true,
			ActionNameReloadProcessV2:                      true,
		},
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operApplyPluginSubConfigV2) ExtraExecutionName() string {
	return ""
}
