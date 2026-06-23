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
	// OperDefNameReconfigProxy the name of the operation definition.
	OperDefNameReconfigProxy = "reconfig_proxy"
)

// NewOperReconfigProxy new an operation.
func NewOperReconfigProxy(param OperParamReconfigProxy) operation.Definition {
	return &operReconfigProxy{param: param}
}

type operReconfigProxy struct {
	param OperParamReconfigProxy
}

// OperParamReconfigProxy defines the parameters for operReconfigProxy.
type OperParamReconfigProxy struct {
	Token    string `json:"token"`
	Operator string `json:"operator"`
}

// Name returns the name.
func (oper *operReconfigProxy) Name() string {
	return OperDefNameReconfigProxy
}

// ActionDefNames returns the action def names.
func (oper *operReconfigProxy) ActionDefNames() []string {
	return []string{
		ActionNameCheckSelfRelayPluginAlive,
		ActionNameVersionCompatCheck,
		ActionNameInjectNodeCustomDeployConfig,
		ActionNameRenderNodeDeployment,
		ActionNameTransferPkgToNode,
		ActionNameReconfigProxy,
		ActionNameWaitInstallerComplete,
		ActionNameRestartNode,
		ActionNameWaitGseReady,
		ActionNameCleanInstaller,
		ActionNameSyncNodeInfo,
		ActionNameUpdateHost,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operReconfigProxy) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNameCheckSelfRelayPluginAlive:    true,
			ActionNameVersionCompatCheck:           true,
			ActionNameInjectNodeCustomDeployConfig: true,
			ActionNameRenderNodeDeployment:         true,
			ActionNameTransferPkgToNode:            true,
			ActionNameReconfigProxy:                true,
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
func (oper *operReconfigProxy) ExtraExecutionName() string {
	return OperExtraExecutionNameLockAndUnlockHost
}
