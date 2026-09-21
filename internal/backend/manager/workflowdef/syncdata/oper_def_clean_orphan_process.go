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

package syncdata

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

const (
	// OperDefNameCleanOrphanProcess defines the operation def name.
	OperDefNameCleanOrphanProcess = "clean_orphan_process"
)

// NewOperCleanOrphanProcess new an operation.
func NewOperCleanOrphanProcess(param OperParamCleanOrphanProcess) operation.Definition {
	return &operCleanOrphanProcess{
		param: param,
	}
}

type operCleanOrphanProcess struct {
	param OperParamCleanOrphanProcess
}

// OperParamCleanOrphanProcess defines the parameters for operCleanOrphanProcess.
type OperParamCleanOrphanProcess struct {
	TenantID  string           `json:"tenant_id"`
	Operator  string           `json:"operator"`
	Processes []*types.Process `json:"processes"`
}

// Name returns the name.
func (oper *operCleanOrphanProcess) Name() string {
	return OperDefNameCleanOrphanProcess
}

// ActionDefNames returns the action def names.
func (oper *operCleanOrphanProcess) ActionDefNames() []string {
	return []string{
		ActionNameCleanOrphanProcess,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operCleanOrphanProcess) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     30 * time.Minute, // nolint: mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operCleanOrphanProcess) ExtraExecutionName() string {
	return ""
}
