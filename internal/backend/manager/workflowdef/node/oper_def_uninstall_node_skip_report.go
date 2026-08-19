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
	// OperDefNameUninstallNodeSkipReport is the operation definition name for uninstall without installer report.
	OperDefNameUninstallNodeSkipReport = "uninstall_node_skip_report"
)

// NewOperUninstallNodeSkipReport creates an uninstall operation that skips installer report.
func NewOperUninstallNodeSkipReport(param OperParamUninstallNode) operation.Definition {
	return &operUninstallNodeSkipReport{param: param}
}

type operUninstallNodeSkipReport struct {
	param OperParamUninstallNode
}

// Name returns the name.
func (oper *operUninstallNodeSkipReport) Name() string {
	return OperDefNameUninstallNodeSkipReport
}

// ActionDefNames returns the action def names.
func (oper *operUninstallNodeSkipReport) ActionDefNames() []string {
	return []string{
		ActionNameInjectNodeCustomDeployConfig,
		ActionNameTransferPkgToNode,
		ActionNameUninstallNodeSkipReport,
		ActionNameWaitGseNotAlive,
		ActionNameResetNodeDynamic,
		ActionNameUpdateHost,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operUninstallNodeSkipReport) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNameInjectNodeCustomDeployConfig: true,
			ActionNameTransferPkgToNode:            true,
			ActionNameUninstallNodeSkipReport:      true,
			ActionNameWaitGseNotAlive:              false,
			ActionNameResetNodeDynamic:             true,
			ActionNameUpdateHost:                   true,
		},
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operUninstallNodeSkipReport) ExtraExecutionName() string {
	return OperExtraExecutionNameLockAndUnlockHost
}
