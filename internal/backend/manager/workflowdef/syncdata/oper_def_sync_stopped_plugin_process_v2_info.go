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
	// OperDefNameSyncStoppedPluginProcessV2Info defines the operation def name.
	OperDefNameSyncStoppedPluginProcessV2Info = "sync_stopped_plugin_process_v2_info"

	// OperDefNameSyncStoppedPluginProcessV2InfoTimeout defines the timeout for each sync stopped v2 plugin process
	// info operation.
	OperDefNameSyncStoppedPluginProcessV2InfoTimeout = 10 * time.Minute
)

// NewOperSyncStoppedPluginProcessV2Info new an operation.
func NewOperSyncStoppedPluginProcessV2Info(param OperParamSyncStoppedPluginProcessV2Info) operation.Definition {
	return &operSyncStoppedPluginProcessV2Info{
		param: param,
	}
}

type operSyncStoppedPluginProcessV2Info struct {
	param OperParamSyncStoppedPluginProcessV2Info
}

// OperParamSyncStoppedPluginProcessV2Info defines the parameters for operSyncStoppedPluginProcessV2Info.
type OperParamSyncStoppedPluginProcessV2Info struct {
	TenantID string  `json:"tenant_id"`
	Operator string  `json:"operator"`
	HostIDs  []int64 `json:"host_ids"`
}

// Name returns the name.
func (oper *operSyncStoppedPluginProcessV2Info) Name() string {
	return OperDefNameSyncStoppedPluginProcessV2Info
}

// ActionDefNames returns the action def names.
func (oper *operSyncStoppedPluginProcessV2Info) ActionDefNames() []string {
	return []string{
		ActionNameSyncStoppedPluginProcessV2Info,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operSyncStoppedPluginProcessV2Info) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     OperDefNameSyncStoppedPluginProcessV2InfoTimeout,
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operSyncStoppedPluginProcessV2Info) ExtraExecutionName() string {
	return ""
}
