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
	// OperDefNameStopPluginV2 the name of the operation definition.
	OperDefNameStopPluginV2 = "stop_plugin_v2"
)

// NewOperStopPluginV2 new an operation.
func NewOperStopPluginV2(param OperParamStopPluginV2) operation.Definition {
	return &operStopPluginV2{param: param}
}

type operStopPluginV2 struct {
	param OperParamStopPluginV2
}

// OperParamStopPluginV2 defines the parameters for operStopPluginV2.
type OperParamStopPluginV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

// Name returns the name.
func (oper *operStopPluginV2) Name() string {
	return OperDefNameStopPluginV2
}

// ActionDefNames returns the action def names.
func (oper *operStopPluginV2) ActionDefNames() []string {
	return []string{
		ActionNameInjectPluginBaseRuntimeV2,
		ActionNameFetchPluginProcessV2,
		ActionNameStopProcessV2,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operStopPluginV2) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNameInjectPluginBaseRuntimeV2: true,
			ActionNameFetchPluginProcessV2:      true,
			ActionNameStopProcessV2:             true,
		},
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operStopPluginV2) ExtraExecutionName() string {
	return ""
}
