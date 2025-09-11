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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/common"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

const (
	// OperDefNameUpgradeNode the name of the operation definition.
	OperDefNameUpgradeNode = "upgrade_node"
)

// NewOperUpgradeNode new an operation.
func NewOperUpgradeNode(param OperParamUpgradeNode) operation.Definition {
	return &operUpgradeNode{param: param}
}

type operUpgradeNode struct {
	param OperParamUpgradeNode
}

// OperParamUpgradeNode defines the parameters for operUpgradeNode.
type OperParamUpgradeNode struct {
	Token    string `json:"token"`
	Operator string `json:"operator"`
}

// Name returns the name.
func (oper *operUpgradeNode) Name() string {
	return OperDefNameUpgradeNode
}

// ActionDefNames returns the action def names.
func (oper *operUpgradeNode) ActionDefNames() []string {
	return []string{
		ActionNameVersionCompatCheck,
		ActionNameRenderNodeDeployment,
		ActionNameTransferPkgToNode,
		ActionNameUpgradeNode,
		common.ActionNameWaitInstallerComplete,
		ActionNameRestartNode,
		ActionNameWaitGseReady,
		ActionNameCleanInstaller,
		ActionNameSyncNodeInfo,
		ActionNameUpdateHost,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operUpgradeNode) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNameVersionCompatCheck:           true,
			ActionNameRenderNodeDeployment:         true,
			ActionNameTransferPkgToNode:            true,
			ActionNameUpgradeNode:                  true,
			common.ActionNameWaitInstallerComplete: false,
			ActionNameRestartNode:                  true,
			ActionNameWaitGseReady:                 false,
			ActionNameCleanInstaller:               true,
			ActionNameSyncNodeInfo:                 true,
			ActionNameUpdateHost:                   true,
		},
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operUpgradeNode) ExtraExecutionName() string {
	return OperExtraExecutionName
}
