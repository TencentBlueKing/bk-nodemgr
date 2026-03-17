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
	// OperDefNameUninstallPagent the name of the operation definition.
	OperDefNameUninstallPagent = "uninstall_pagent"
)

// NewOperUninstallPagent new an operation.
func NewOperUninstallPagent(param OperParamUninstallPagent) operation.Definition {
	return &operUninstallPagent{param: param}
}

type operUninstallPagent struct {
	param OperParamUninstallPagent
}

// OperParamUninstallPagent defines the parameters for operUninstallPagent.
type OperParamUninstallPagent struct {
	Token    string `json:"token"`
	Operator string `json:"operator"`
}

// Name returns the name.
func (oper *operUninstallPagent) Name() string {
	return OperDefNameUninstallPagent
}

// ActionDefNames returns the action def names.
func (oper *operUninstallPagent) ActionDefNames() []string {
	return []string{
		ActionNameSelectRelayHost,
		ActionNameTransferPkgToNode,
		ActionNameUninstallPagent,
		ActionNameWaitInstallerComplete,
		ActionNameResetNodeDynamic,
		ActionNameUpdateHost,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operUninstallPagent) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNameSelectRelayHost:       true,
			ActionNameTransferPkgToNode:     true,
			ActionNameUninstallPagent:       true,
			ActionNameWaitInstallerComplete: false,
			ActionNameResetNodeDynamic:      true,
			ActionNameUpdateHost:            true,
		},
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operUninstallPagent) ExtraExecutionName() string {
	return OperExtraExecutionName
}
