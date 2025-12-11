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
	// OperDefNameInstallPlugin the name of the operation definition.
	OperDefNameInstallPlugin = "install_plugin"
)

// NewOperInstallPlugin new an operation.
func NewOperInstallPlugin(param OperParamInstallPlugin) operation.Definition {
	return &operInstallPlugin{param: param}
}

type operInstallPlugin struct {
	param OperParamInstallPlugin
}

// OperParamInstallPlugin defines the parameters for operInstallPlugin.
type OperParamInstallPlugin struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// Name returns the name.
func (oper *operInstallPlugin) Name() string {
	return OperDefNameInstallPlugin
}

// ActionDefNames returns the action def names.
func (oper *operInstallPlugin) ActionDefNames() []string {
	return []string{
		ActionNameUpsertProcess,
		ActionNameRenderPluginDeployment,
		ActionNameEnsureAndUpdatePluginConfigDetails,
		ActionNameRenderPluginConfig,
		ActionNameTransferPluginPkgToNode,
		ActionNameInstallPlugin,
		ActionNameWaitInstallerComplete,
		ActionNameTrusteeshipProcess,
		ActionNameUpdateProcess,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operInstallPlugin) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNameRenderPluginDeployment:  true,
			ActionNameTransferPluginPkgToNode: true,
			ActionNameInstallPlugin:           true,
			ActionNameWaitInstallerComplete:   false,
			ActionNameTrusteeshipProcess:      true,
		},
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operInstallPlugin) ExtraExecutionName() string {
	return ""
}
