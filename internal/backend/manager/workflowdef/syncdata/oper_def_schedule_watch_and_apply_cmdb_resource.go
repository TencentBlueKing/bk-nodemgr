/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package syncdata

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// OperDefNameWatchAndApplyCMDBResource defines the operation definition name.
const OperDefNameWatchAndApplyCMDBResource = "watch_and_apply_cmdb_resource"

// NewOperWatchAndApplyCMDBResource new an operation.
func NewOperWatchAndApplyCMDBResource(param OperParamWatchCMDBResource) operation.Definition {
	return &operWatchCMDBResource{
		param: param,
	}
}

// operWatchCMDBResource implements the operation.Definition interface.
type operWatchCMDBResource struct {
	param OperParamWatchCMDBResource
}

// OperParamWatchCMDBResource defines the parameters for operWatchCMDBResource.
type OperParamWatchCMDBResource struct {
	TenantID string `json:"tenant_id"`
	Operator string `json:"operator"`
}

// Name returns the name of the operation definition.
func (oper *operWatchCMDBResource) Name() string {
	return OperDefNameWatchAndApplyCMDBResource
}

// ActionDefNames returns the action def names.
func (oper *operWatchCMDBResource) ActionDefNames() []string {
	return []string{
		ActionNameWatchAndApplyCMDBResource,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operWatchCMDBResource) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     30 * time.Second, // nolint:mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operWatchCMDBResource) ExtraExecutionName() string {
	return ""
}
