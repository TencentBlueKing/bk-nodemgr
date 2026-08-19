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

package node

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

const (
	// OperDefNameInstallPagentByManual the name of the operation definition.
	OperDefNameInstallPagentByManual = "install_pagent_by_manual"
)

// NewOperInstallPagentByManual new an operation.
func NewOperInstallPagentByManual(param OperParamInstallPagentByManual) operation.Definition {
	return &operInstallPagentByManual{param: param}
}

type operInstallPagentByManual struct {
	param OperParamInstallPagentByManual
}

// OperParamInstallPagentByManual defines the parameters for operInstallPagentByManual.
type OperParamInstallPagentByManual struct {
	Token    string `json:"token"`
	Operator string `json:"operator"`
}

// Name returns the name.
func (oper *operInstallPagentByManual) Name() string {
	return OperDefNameInstallPagentByManual
}

// ActionDefNames returns the action def names.
func (oper *operInstallPagentByManual) ActionDefNames() []string {
	return []string{
		ActionNameTryReuseAgentID,
		ActionNameUpsertHostToCMDB,
		ActionNameSelectRelayHost,
		ActionNameGenManualBootstrapCommand,
		ActionNameWaitDetectInfoByManual,
		ActionNameInjectNodeCustomDeployConfig,
		ActionNameRenderNodeDeployment,
		ActionNameEnsurePkgToRelay,
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
func (oper *operInstallPagentByManual) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     20 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNameTryReuseAgentID:              true,
			ActionNameUpsertHostToCMDB:             true,
			ActionNameSelectRelayHost:              true,
			ActionNameGenManualBootstrapCommand:    true,
			ActionNameWaitDetectInfoByManual:       false,
			ActionNameInjectNodeCustomDeployConfig: false,
			ActionNameRenderNodeDeployment:         false,
			ActionNameEnsurePkgToRelay:             true,
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
func (oper *operInstallPagentByManual) ExtraExecutionName() string {
	return OperExtraExecutionNameLockAndUnlockHost
}
