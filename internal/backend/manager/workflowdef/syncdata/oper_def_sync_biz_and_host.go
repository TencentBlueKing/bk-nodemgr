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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// OperDefNameSyncBizAndHost sync all biz and their host from cmdb.
const OperDefNameSyncBizAndHost = "sync_biz_and_host"

// NewOperSyncBizAndHost new an operation.
func NewOperSyncBizAndHost(param OperParamSyncBizAndHost) operation.Definition {
	return &operSyncBizAndHost{
		param: param,
	}
}

type operSyncBizAndHost struct {
	param OperParamSyncBizAndHost
}

// OperParamSyncBizAndHost defines the parameters for operSyncBizAndHost.
type OperParamSyncBizAndHost struct {
	TenantID string `json:"tenant_id"`
	Operator string `json:"operator"`
}

// Name returns the name.
func (oper *operSyncBizAndHost) Name() string {
	return OperDefNameSyncBizAndHost
}

// ActionDefNames returns the action def names.
func (oper *operSyncBizAndHost) ActionDefNames() []string {
	return []string{
		ActionNameSyncBusiness,
		ActionNameGenOperSyncHost,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operSyncBizAndHost) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint:mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operSyncBizAndHost) ExtraExecutionName() string {
	return ""
}
