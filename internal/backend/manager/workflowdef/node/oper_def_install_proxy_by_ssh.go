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
	// OperDefNameInstallProxyBySSH is the name of the proxy SSH install operation definition.
	OperDefNameInstallProxyBySSH = "install_proxy_by_ssh"
)

// NewOperInstallProxyBySSH creates a new operation definition for proxy SSH installation.
func NewOperInstallProxyBySSH(param OperParamInstallProxyBySSH) operation.Definition {
	return &operInstallProxyBySSH{param: param}
}

type operInstallProxyBySSH struct {
	param OperParamInstallProxyBySSH
}

// OperParamInstallProxyBySSH defines the parameters for operInstallProxyBySSH.
type OperParamInstallProxyBySSH struct {
	Token     string `json:"token"`
	Operator  string `json:"operator"`
	CrossUnit bool   `json:"cross_unit"`
}

// Name returns the name.
func (oper *operInstallProxyBySSH) Name() string {
	return OperDefNameInstallProxyBySSH
}

// ActionDefNames returns the action def names.
func (oper *operInstallProxyBySSH) ActionDefNames() []string {
	if oper.param.CrossUnit {
		return []string{
			ActionNameTryReuseAgentID,
			ActionNameUpsertHostToCMDB,
			ActionNameDetectInfoBySSH,
			ActionNameInjectNodeCustomDeployConfig,
			ActionNameRenderNodeDeployment,
			ActionNameInstallProxyBySSH,
			ActionNameWaitInstallerComplete,
			ActionNameWaitGseReady,
			ActionNameSyncNodeInfo,
			ActionNameBindAgentHostRel,
			ActionNamePushHostIdentifier,
			ActionNameUpdateHost,
			ActionNameInstallPreOrderedPlugins,
		}
	}

	return []string{
		ActionNameTryReuseAgentID,
		ActionNameUpsertHostToCMDB,
		ActionNameSelectRelayHost,
		ActionNamePagentDetectInfoBySSH,
		ActionNameInjectNodeCustomDeployConfig,
		ActionNameRenderNodeDeployment,
		ActionNameEnsurePkgToRelay,
		ActionNameInstallProxyBySSH,
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
func (oper *operInstallProxyBySSH) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:         10 * time.Minute, // nolint: mnd
		InitContent:     conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: oper.retryStartPoint(),
	}
}

func (oper *operInstallProxyBySSH) retryStartPoint() map[string]bool {
	if oper.param.CrossUnit {
		return map[string]bool{
			ActionNameTryReuseAgentID:              true,
			ActionNameUpsertHostToCMDB:             true,
			ActionNameDetectInfoBySSH:              true,
			ActionNameInjectNodeCustomDeployConfig: true,
			ActionNameRenderNodeDeployment:         true,
			ActionNameInstallProxyBySSH:            true,
			ActionNameWaitInstallerComplete:        false,
			ActionNameWaitGseReady:                 false,
			ActionNameSyncNodeInfo:                 true,
			ActionNameBindAgentHostRel:             true,
			ActionNamePushHostIdentifier:           true,
			ActionNameUpdateHost:                   true,
			ActionNameInstallPreOrderedPlugins:     true,
		}
	}

	return map[string]bool{
		ActionNameTryReuseAgentID:              true,
		ActionNameUpsertHostToCMDB:             true,
		ActionNameSelectRelayHost:              true,
		ActionNameInjectNodeCustomDeployConfig: true,
		ActionNamePagentDetectInfoBySSH:        true,
		ActionNameRenderNodeDeployment:         true,
		ActionNameEnsurePkgToRelay:             true,
		ActionNameInstallProxyBySSH:            true,
		ActionNameWaitInstallerComplete:        false,
		ActionNameWaitGseReady:                 false,
		ActionNameSyncNodeInfo:                 true,
		ActionNameBindAgentHostRel:             true,
		ActionNamePushHostIdentifier:           true,
		ActionNameUpdateHost:                   true,
		ActionNameInstallPreOrderedPlugins:     true,
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operInstallProxyBySSH) ExtraExecutionName() string {
	return OperExtraExecutionNameLockAndUnlockHost
}
