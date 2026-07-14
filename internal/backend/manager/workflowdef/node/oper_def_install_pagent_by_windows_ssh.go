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
	// OperDefNameInstallPagntNodeByWindowsSSH the name of the operation definition.
	OperDefNameInstallPagntNodeByWindowsSSH = "install_pagent_node_by_windows_ssh"
)

// NewOperInstallPagentNodeByWindowsSSH creates a dormant operation.
// Selection is deferred because the retained SSH detection seam is Unix-only and rejects Windows;
// future Windows-capable relay detection and selector wiring must replace or extend it.
func NewOperInstallPagentNodeByWindowsSSH(param OperParamInstallPagentNodeByWindowsSSH) operation.Definition {
	return &operInstallPagentNodeByWindowsSSH{param: param}
}

type operInstallPagentNodeByWindowsSSH struct {
	param OperParamInstallPagentNodeByWindowsSSH
}

// OperParamInstallPagentNodeByWindowsSSH defines the parameters for operInstallPagentNodeByWindowsSSH.
type OperParamInstallPagentNodeByWindowsSSH struct {
	Token    string `json:"token"`
	Operator string `json:"operator"`
}

// Name returns the name.
func (oper *operInstallPagentNodeByWindowsSSH) Name() string {
	return OperDefNameInstallPagntNodeByWindowsSSH
}

// ActionDefNames returns the action def names.
func (oper *operInstallPagentNodeByWindowsSSH) ActionDefNames() []string {
	return []string{
		ActionNameTryReuseAgentID,
		ActionNameUpsertHostToCMDB,
		ActionNameSelectRelayHost,
		ActionNamePagentDetectInfoBySSH,
		ActionNameInjectNodeCustomDeployConfig,
		ActionNameRenderNodeDeployment,
		ActionNameEnsurePkgToRelay,
		ActionNameInstallPagentByWindowsSSH,
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
func (oper *operInstallPagentNodeByWindowsSSH) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNameTryReuseAgentID:              true,
			ActionNameUpsertHostToCMDB:             true,
			ActionNameSelectRelayHost:              true,
			ActionNamePagentDetectInfoBySSH:        true,
			ActionNameInjectNodeCustomDeployConfig: true,
			ActionNameRenderNodeDeployment:         true,
			ActionNameEnsurePkgToRelay:             true,
			ActionNameInstallPagentByWindowsSSH:    true,
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
func (oper *operInstallPagentNodeByWindowsSSH) ExtraExecutionName() string {
	return OperExtraExecutionNameLockAndUnlockHost
}
