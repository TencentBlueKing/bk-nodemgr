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

// OperDefNameSyncHost defines the operation def name.
const OperDefNameSyncHost = "sync_host_from_cmdb"

// NewOperSyncHostFromCMDB new an operation.
func NewOperSyncHostFromCMDB(param OperParamSyncHostFromCMDB) operation.Definition {
	return &operSyncHostFromCMDB{
		param: param,
	}
}

type operSyncHostFromCMDB struct {
	param OperParamSyncHostFromCMDB
}

// OperParamSyncHostFromCMDB defines the parameters for operSyncHostFromCMDB.
type OperParamSyncHostFromCMDB struct {
	BizID    int64  `json:"biz_id"`
	TenantID string `json:"tenant_id"`
	Operator string `json:"operator"`
}

// Name returns the name.
func (oper *operSyncHostFromCMDB) Name() string {
	return OperDefNameSyncHost
}

// ActionDefNames returns the action def names.
func (oper *operSyncHostFromCMDB) ActionDefNames() []string {
	return []string{
		ActionNameSyncHostFromCMDB,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operSyncHostFromCMDB) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint:mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}
