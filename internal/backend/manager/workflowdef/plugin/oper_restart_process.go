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
	// OperDefNameRestartProcess the name of the operation definition.
	OperDefNameRestartProcess = "restart_process"
)

// NewOperRestartProcess new an operation.
func NewOperRestartProcess(param OperParamRestartProcess) operation.Definition {
	return &operRestartProcess{param: param}
}

type operRestartProcess struct {
	param OperParamRestartProcess
}

// OperParamRestartProcess defines the parameters for operRestartProcess.
type OperParamRestartProcess struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// Name returns the name.
func (oper *operRestartProcess) Name() string {
	return OperDefNameRestartProcess
}

// ActionDefNames returns the action def names.
func (oper *operRestartProcess) ActionDefNames() []string {
	return []string{
		ActionNameFetchPluginProcess,
		ActionNameRestartProcess,
		ActionNameCheckPluginProcessAlive,
		ActionNameUpdateProcess,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operRestartProcess) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNameFetchPluginProcess:      true,
			ActionNameRestartProcess:          true,
			ActionNameCheckPluginProcessAlive: true,
			ActionNameUpdateProcess:           true,
		},
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operRestartProcess) ExtraExecutionName() string {
	return ""
}
