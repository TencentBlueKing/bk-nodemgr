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

const (
	// OperDefNameSyncAlivePluginProcessV2Info defines the operation def name.
	OperDefNameSyncAlivePluginProcessV2Info = "sync_alive_plugin_process_v2_info"

	// OperDefNameSyncAlivePluginProcessV2InfoTimeout defines the timeout for each sync alive v2 plugin process info operation.
	OperDefNameSyncAlivePluginProcessV2InfoTimeout = 10 * time.Minute
)

// NewOperSyncAlivePluginProcessV2Info new an operation.
func NewOperSyncAlivePluginProcessV2Info(param OperParamSyncAlivePluginProcessV2Info) operation.Definition {
	return &operSyncAlivePluginProcessV2Info{
		param: param,
	}
}

type operSyncAlivePluginProcessV2Info struct {
	param OperParamSyncAlivePluginProcessV2Info
}

// OperParamSyncAlivePluginProcessV2Info defines the parameters for operSyncAlivePluginProcessV2Info.
type OperParamSyncAlivePluginProcessV2Info struct {
	TenantID string  `json:"tenant_id"`
	Operator string  `json:"operator"`
	HostIDs  []int64 `json:"host_ids"`
}

// Name returns the name.
func (oper *operSyncAlivePluginProcessV2Info) Name() string {
	return OperDefNameSyncAlivePluginProcessV2Info
}

// ActionDefNames returns the action def names.
func (oper *operSyncAlivePluginProcessV2Info) ActionDefNames() []string {
	return []string{
		ActionNameSyncAlivePluginProcessV2Info,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operSyncAlivePluginProcessV2Info) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     OperDefNameSyncAlivePluginProcessV2InfoTimeout,
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operSyncAlivePluginProcessV2Info) ExtraExecutionName() string {
	return ""
}
