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
	// OperDefNameUpgradePlugin the name of the operation definition.
	OperDefNameUpgradePlugin = "upgrade_plugin"
)

// NewOperUpgradePlugin new an operation.
func NewOperUpgradePlugin(param OperParamUpgradePlugin) operation.Definition {
	return &operUpgradePlugin{param: param}
}

type operUpgradePlugin struct {
	param OperParamUpgradePlugin
}

// OperParamUpgradePlugin defines the parameters for operUpgradePlugin.
type OperParamUpgradePlugin struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// Name returns the name.
func (oper *operUpgradePlugin) Name() string {
	return OperDefNameUpgradePlugin
}

// ActionDefNames returns the action def names.
func (oper *operUpgradePlugin) ActionDefNames() []string {
	return []string{
		ActionNameFetchPluginProcess,
		ActionNameVerifyPluginAvailability,
		ActionNameCheckPluginProcessAlive,
		ActionNameStopProcess,
		ActionNameInjectPluginCustomDeployConfig,
		ActionNameRenderPluginDeployment,
		ActionNameFetchProcessSubConfigIntoDeployment,
		ActionNameEnsureAndUpdatePluginConfigDetails,
		ActionNameRenderPluginConfig,
		ActionNameOverwritePluginConfigForCompatibility,
		ActionNameTransferPluginPkgToNode,
		ActionNameUpgradePlugin,
		ActionNameWaitPluginInstallerComplete,
		ActionNamePushPluginConfig,
		ActionNameStartProcess,
		ActionNameUpdateProcess,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operUpgradePlugin) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNameFetchPluginProcess:                    true,
			ActionNameVerifyPluginAvailability:              true,
			ActionNameCheckPluginProcessAlive:               true,
			ActionNameStopProcess:                           true,
			ActionNameInjectPluginCustomDeployConfig:        true,
			ActionNameRenderPluginDeployment:                true,
			ActionNameEnsureAndUpdatePluginConfigDetails:    true,
			ActionNameRenderPluginConfig:                    true,
			ActionNameOverwritePluginConfigForCompatibility: true,
			ActionNameTransferPluginPkgToNode:               true,
			ActionNameUpgradePlugin:                         true,
			ActionNameWaitPluginInstallerComplete:           false,
			ActionNameFetchProcessSubConfigIntoDeployment:   true,
			ActionNamePushPluginConfig:                      true,
			ActionNameStartProcess:                          true,
			ActionNameUpdateProcess:                         true,
		},
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operUpgradePlugin) ExtraExecutionName() string {
	return ""
}
