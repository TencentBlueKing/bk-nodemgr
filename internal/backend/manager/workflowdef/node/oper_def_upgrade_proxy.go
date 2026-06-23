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
	// OperDefNameUpgradeProxy the name of the operation definition.
	OperDefNameUpgradeProxy = "upgrade_proxy"
)

// NewOperUpgradeProxy new an operation.
func NewOperUpgradeProxy(param OperParamUpgradeProxy) operation.Definition {
	return &operUpgradeProxy{param: param}
}

type operUpgradeProxy struct {
	param OperParamUpgradeProxy
}

// OperParamUpgradeProxy defines the parameters for operUpgradeProxy.
type OperParamUpgradeProxy struct {
	Token    string `json:"token"`
	Operator string `json:"operator"`
}

// Name returns the name.
func (oper *operUpgradeProxy) Name() string {
	return OperDefNameUpgradeProxy
}

// ActionDefNames returns the action def names.
func (oper *operUpgradeProxy) ActionDefNames() []string {
	return []string{
		ActionNameCheckSelfRelayPluginAlive,
		ActionNameVersionCompatCheck,
		ActionNameInjectNodeCustomDeployConfig,
		ActionNameRenderNodeDeployment,
		ActionNameTransferPkgToNode,
		ActionNameUpgradeProxy,
		ActionNameWaitInstallerComplete,
		ActionNameRestartNode,
		ActionNameWaitGseReady,
		ActionNameCleanInstaller,
		ActionNameSyncNodeInfo,
		ActionNameUpdateHost,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operUpgradeProxy) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNameCheckSelfRelayPluginAlive:    true,
			ActionNameVersionCompatCheck:           true,
			ActionNameInjectNodeCustomDeployConfig: true,
			ActionNameRenderNodeDeployment:         true,
			ActionNameTransferPkgToNode:            true,
			ActionNameUpgradeProxy:                 true,
			ActionNameWaitInstallerComplete:        false,
			ActionNameRestartNode:                  true,
			ActionNameWaitGseReady:                 false,
			ActionNameCleanInstaller:               true,
			ActionNameSyncNodeInfo:                 true,
			ActionNameUpdateHost:                   true,
		},
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operUpgradeProxy) ExtraExecutionName() string {
	return OperExtraExecutionNameLockAndUnlockHost
}
