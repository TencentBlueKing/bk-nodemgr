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

// OperDefNameSyncNetworkArea defines the operation def name.
const OperDefNameSyncNetworkArea = "sync_networkarea_from_cmdb"

// NewOperSyncNetworkAreaFromCMDB new an operation definition.
func NewOperSyncNetworkAreaFromCMDB(param OperParamSyncNetworkAreaFromCMDB) operation.Definition {
	return &operSyncNetworkAreaFromCMDB{
		param: param,
	}
}

type operSyncNetworkAreaFromCMDB struct {
	param OperParamSyncNetworkAreaFromCMDB
}

// OperParamSyncNetworkAreaFromCMDB defines the parameters for operSyncNetworkAreaFromCMDB.
type OperParamSyncNetworkAreaFromCMDB struct {
	TenantID string `json:"tenant_id"`
}

// Name returns the name.
func (oper *operSyncNetworkAreaFromCMDB) Name() string {
	return OperDefNameSyncNetworkArea
}

// ActionDefNames returns the action def names.
func (oper *operSyncNetworkAreaFromCMDB) ActionDefNames() []string {
	return []string{
		ActionNameSyncNetworkAreaFromCMDB,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operSyncNetworkAreaFromCMDB) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     1 * time.Minute, // nolint:mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ActionRetryable checks if the action is a retry start point.
func (oper *operSyncNetworkAreaFromCMDB) ActionRetryable(_ string) bool {
	return false
}
