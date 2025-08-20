/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodeinstall

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

const (
	// OperDefNameInstallPagntNodeBySSH the name of the operation definition.
	OperDefNameInstallPagntNodeBySSH = "install_pagent_node_by_ssh"
)

// NewOperInstallPagentNodeBySSH new an operation.
func NewOperInstallPagentNodeBySSH(param OperParamInstallPagentNodeBySSH) operation.Definition {
	return &operInstallPagentNodeBySSH{param: param}
}

type operInstallPagentNodeBySSH struct {
	param OperParamInstallPagentNodeBySSH
}

// OperParamInstallPagentNodeBySSH defines the parameters for operInstallNodeBySSH.
type OperParamInstallPagentNodeBySSH struct {
	Token    string `json:"token"`
	Operator string `json:"operator"`
}

// Name returns the name.
func (oper *operInstallPagentNodeBySSH) Name() string {
	return OperDefNameInstallPagntNodeBySSH
}

// ActionDefNames returns the action def names.
func (oper *operInstallPagentNodeBySSH) ActionDefNames() []string {
	return []string{
		ActionNameEnsurePkgToRelay,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operInstallPagentNodeBySSH) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
		RetryStartPoint: map[string]bool{
			ActionNameEnsurePkgToRelay: true,
		},
	}
}
