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
	// OperDefNameUninstallPlugin the name of the operation definition.
	OperDefNameUninstallPlugin = "uninstall_plugin"
)

// NewOperUninstallPlugin new an operation.
func NewOperUninstallPlugin(param OperParamUninstallPlugin) operation.Definition {
	return &operUninstallPlugin{param: param}
}

type operUninstallPlugin struct {
	param OperParamUninstallPlugin
}

// OperParamUninstallPlugin defines the parameters for operUninstallPlugin.
type OperParamUninstallPlugin struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// Name returns the name.
func (oper *operUninstallPlugin) Name() string {
	return OperDefNameUninstallPlugin
}

// ActionDefNames returns the action def names.
func (oper *operUninstallPlugin) ActionDefNames() []string {
	return []string{
		ActionNameFetchPluginProcess,
		ActionNameCheckPluginProcessAlive,
		ActionNameStopProcess,
		ActionNameInjectPluginCustomDeployConfig,
		ActionNameTransferPluginPkgToNode,
		ActionNameUninstallPlugin,
		ActionNameWaitPluginInstallerComplete,
		ActionNameDeleteProcess,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operUninstallPlugin) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNameFetchPluginProcess:             true,
			ActionNameCheckPluginProcessAlive:        true,
			ActionNameStopProcess:                    true,
			ActionNameInjectPluginCustomDeployConfig: true,
			ActionNameTransferPluginPkgToNode:        true,
			ActionNameUninstallPlugin:                true,
			ActionNameWaitPluginInstallerComplete:    false,
			ActionNameDeleteProcess:                  true,
		},
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operUninstallPlugin) ExtraExecutionName() string {
	return ""
}
