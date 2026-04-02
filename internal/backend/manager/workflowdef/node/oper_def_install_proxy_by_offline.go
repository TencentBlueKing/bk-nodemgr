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
	// OperDefNameInstallProxyByOffline is the name of the offline proxy install operation definition.
	OperDefNameInstallProxyByOffline = "install_proxy_by_offline"
)

// NewOperInstallProxyByOffline creates a new operation definition for offline proxy installation.
func NewOperInstallProxyByOffline(param OperParamInstallProxyByOffline) operation.Definition {
	return &operInstallProxyByOffline{param: param}
}

type operInstallProxyByOffline struct {
	param OperParamInstallProxyByOffline
}

// OperParamInstallProxyByOffline defines the parameters for operInstallProxyByOffline.
type OperParamInstallProxyByOffline struct {
	Token    string `json:"token"`
	Operator string `json:"operator"`
}

// Name returns the name.
func (oper *operInstallProxyByOffline) Name() string {
	return OperDefNameInstallProxyByOffline
}

// ActionDefNames returns the action def names.
func (oper *operInstallProxyByOffline) ActionDefNames() []string {
	return []string{
		ActionNameTryReuseAgentID,
		ActionNameUpsertHostToCMDB,
		ActionNameResolveOfflineDetectInfo,
		ActionNameInjectNodeCustomDeployConfig,
		ActionNameRenderNodeDeployment,
		ActionNameWaitOfflineManualInstall,
		ActionNameWaitGseReady,
		ActionNameSyncNodeInfo,
		ActionNameBindAgentHostRel,
		ActionNamePushHostIdentifier,
		ActionNameUpdateHost,
		ActionNameInstallPreOrderedPlugins,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operInstallProxyByOffline) DefaultParameters() operation.Param {
	// Timeout must be >= WaitGseReady timeout (24h) + WaitOfflineManualInstall timeout (3h).
	return operation.Param{
		Timeout:     30 * time.Hour, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNameTryReuseAgentID:              true,
			ActionNameUpsertHostToCMDB:             true,
			ActionNameResolveOfflineDetectInfo:     true,
			ActionNameInjectNodeCustomDeployConfig: true,
			ActionNameRenderNodeDeployment:         true,
			ActionNameWaitOfflineManualInstall:     false,
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
func (oper *operInstallProxyByOffline) ExtraExecutionName() string {
	return OperExtraExecutionName
}
