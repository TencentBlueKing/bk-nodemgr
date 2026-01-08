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
	// OperDefNameApplyPluginSubConfig the name of the operation definition.
	OperDefNameApplyPluginSubConfig = "apply_plugin_subconfig"
)

// NewOperApplyPluginSubConfig new an operation.
func NewOperApplyPluginSubConfig(param OperParamApplyPluginSubConfig) operation.Definition {
	return &operApplyPluginSubConfig{param: param}
}

type operApplyPluginSubConfig struct {
	param OperParamApplyPluginSubConfig
}

// OperParamApplyPluginSubConfig defines the parameters for operApplyPluginSubConfig.
type OperParamApplyPluginSubConfig struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// Name returns the name.
func (oper *operApplyPluginSubConfig) Name() string {
	return OperDefNameApplyPluginSubConfig
}

// ActionDefNames returns the action def names.
func (oper *operApplyPluginSubConfig) ActionDefNames() []string {
	return []string{
		ActionNameCheckPluginProcessAlive,
		ActionNameVerifyPluginAvailability,
		ActionNameRenderPluginDeployment,
		ActionNameEnsureAndUpdatePluginConfigDetails,
		ActionNameRenderPluginConfig,
		ActionNamePushPluginConfig,
		ActionNameReloadProcess,
		ActionNameUpdateProcess,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operApplyPluginSubConfig) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNameCheckPluginProcessAlive:            true,
			ActionNameVerifyPluginAvailability:           true,
			ActionNameRenderPluginDeployment:             true,
			ActionNameEnsureAndUpdatePluginConfigDetails: true,
			ActionNameRenderPluginConfig:                 true,
			ActionNamePushPluginConfig:                   true,
			ActionNameReloadProcess:                      true,
			ActionNameUpdateProcess:                      true,
		},
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operApplyPluginSubConfig) ExtraExecutionName() string {
	return ""
}
