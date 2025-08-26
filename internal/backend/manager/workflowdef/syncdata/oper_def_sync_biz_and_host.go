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

// OperDefNameSyncBizAndHostFromCMDB sync all biz and their host from cmdb.
const OperDefNameSyncBizAndHostFromCMDB = "sync_biz_and_host_from_cmdb"

// NewOperSyncBizAndHostFromCMDB new an operation.
func NewOperSyncBizAndHostFromCMDB(param OperParamSyncBizAndHostFromCMDB) operation.Definition {
	return &operSyncBizAndHostFromCMDB{
		param: param,
	}
}

type operSyncBizAndHostFromCMDB struct {
	param OperParamSyncBizAndHostFromCMDB
}

// OperParamSyncBizAndHostFromCMDB defines the parameters for operSyncBizAndHostFromCMDB.
type OperParamSyncBizAndHostFromCMDB struct {
	TenantID string `json:"tenant_id"`
	Operator string `json:"operator"`
}

// Name returns the name.
func (oper *operSyncBizAndHostFromCMDB) Name() string {
	return OperDefNameSyncBizAndHostFromCMDB
}

// ActionDefNames returns the action def names.
func (oper *operSyncBizAndHostFromCMDB) ActionDefNames() []string {
	return []string{
		ActionNameSyncBizFromCMDB,
		ActionNameGenOperSyncHost,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operSyncBizAndHostFromCMDB) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint:mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operSyncBizAndHostFromCMDB) ExtraExecutionName() string {
	return ""
}
