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
	// OperDefNameInstallNodeByManual the name of the operation definition.
	OperDefNameInstallNodeByManual = "install_node_by_manual"
)

// NewOperInstallNodeByManual new an operation.
func NewOperInstallNodeByManual(param OperParamInstallNodeByManual) operation.Definition {
	return &operInstallNodeByManual{param: param}
}

type operInstallNodeByManual struct {
	param OperParamInstallNodeByManual
}

// OperParamInstallNodeByManual defines the parameters for operInstallNodeByManual.
type OperParamInstallNodeByManual struct {
	Token    string `json:"token"`
	Operator string `json:"operator"`
}

// Name returns the name.
func (oper *operInstallNodeByManual) Name() string {
	return OperDefNameInstallNodeByManual
}

// ActionDefNames returns the action def names.
func (oper *operInstallNodeByManual) ActionDefNames() []string {
	return []string{
		ActionNameTryReuseAgentID,
		ActionNameUpsertHostToCMDB,
		ActionNameGenManualBootstrapCommand,
		ActionNameWaitDetectInfoByManual,
		ActionNameInjectNodeCustomDeployConfig,
		ActionNameRenderNodeDeployment,
		ActionNameInstallNodeByManual,
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
func (oper *operInstallNodeByManual) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     20 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNameTryReuseAgentID:              true,
			ActionNameUpsertHostToCMDB:             true,
			ActionNameGenManualBootstrapCommand:    true,
			ActionNameWaitDetectInfoByManual:       false,
			ActionNameInjectNodeCustomDeployConfig: false,
			ActionNameRenderNodeDeployment:         false,
			ActionNameInstallNodeByManual:          false,
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
func (oper *operInstallNodeByManual) ExtraExecutionName() string {
	return OperExtraExecutionNameLockAndUnlockHost
}
