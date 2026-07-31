/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package node

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

const (
	// OperDefNameInstallNodeByWindowsAuto the name of the operation definition.
	OperDefNameInstallNodeByWindowsAuto = "install_node_by_windows_auto"
)

// NewOperInstallNodeByWindowsAuto new an operation.
func NewOperInstallNodeByWindowsAuto(param OperParamInstallNodeByWindowsAuto) operation.Definition {
	return &operInstallNodeByWindowsAuto{param: param}
}

type operInstallNodeByWindowsAuto struct {
	param OperParamInstallNodeByWindowsAuto
}

// OperParamInstallNodeByWindowsAuto defines the parameters for operInstallNodeByWindowsAuto.
type OperParamInstallNodeByWindowsAuto struct {
	Token    string `json:"token"`
	Operator string `json:"operator"`
}

// Name returns the name.
func (oper *operInstallNodeByWindowsAuto) Name() string {
	return OperDefNameInstallNodeByWindowsAuto
}

// ActionDefNames returns the action def names.
func (oper *operInstallNodeByWindowsAuto) ActionDefNames() []string {
	return []string{
		ActionNameTryReuseAgentID,
		ActionNameUpsertHostToCMDB,
		ActionNameDetectInfoByWindowsAuto,
		ActionNameInjectNodeCustomDeployConfig,
		ActionNameRenderNodeDeployment,
		ActionNameInstallNodeByWindowsAuto,
		ActionNameWaitInstallerComplete,
		ActionNameWaitGseReady,
		ActionNameSyncNodeInfo,
		ActionNameBindAgentHostRel,
		ActionNamePushHostIdentifier,
		ActionNameUpdateHost,
		ActionNameInstallPreOrderedPlugins,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operInstallNodeByWindowsAuto) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNameTryReuseAgentID:              true,
			ActionNameUpsertHostToCMDB:             true,
			ActionNameDetectInfoByWindowsAuto:      true,
			ActionNameInjectNodeCustomDeployConfig: true,
			ActionNameRenderNodeDeployment:         true,
			ActionNameInstallNodeByWindowsAuto:     true,
			ActionNameWaitInstallerComplete:        false,
			ActionNameWaitGseReady:                 false,
			ActionNameSyncNodeInfo:                 true,
			ActionNameBindAgentHostRel:             true,
			ActionNamePushHostIdentifier:           true,
			ActionNameUpdateHost:                   true,
			ActionNameInstallPreOrderedPlugins:     true,
		},
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operInstallNodeByWindowsAuto) ExtraExecutionName() string {
	return OperExtraExecutionNameLockAndUnlockHost
}
