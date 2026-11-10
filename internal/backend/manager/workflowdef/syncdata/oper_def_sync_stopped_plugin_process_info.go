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
	// OperDefNameSyncStoppedPluginProcessInfo defines the operation def name.
	OperDefNameSyncStoppedPluginProcessInfo = "sync_stopped_plugin_process_info"

	// OperDefNameSyncStoppedPluginProcessInfoTimeout defines the timeout for each sync stopped plugin process info operation.
	OperDefNameSyncStoppedPluginProcessInfoTimeout = 10 * time.Minute
)

// NewOperSyncStoppedPluginProcessInfo new an operation.
func NewOperSyncStoppedPluginProcessInfo(param OperParamSyncStoppedPluginProcessInfo) operation.Definition {
	return &operSyncStoppedPluginProcessInfo{
		param: param,
	}
}

type operSyncStoppedPluginProcessInfo struct {
	param OperParamSyncStoppedPluginProcessInfo
}

// OperParamSyncStoppedPluginProcessInfo defines the parameters for operSyncStoppedPluginProcessInfo.
type OperParamSyncStoppedPluginProcessInfo struct {
	TenantID string  `json:"tenant_id"`
	Operator string  `json:"operator"`
	HostIDs  []int64 `json:"host_ids"`
}

// Name returns the name.
func (oper *operSyncStoppedPluginProcessInfo) Name() string {
	return OperDefNameSyncStoppedPluginProcessInfo
}

// ActionDefNames returns the action def names.
func (oper *operSyncStoppedPluginProcessInfo) ActionDefNames() []string {
	return []string{
		ActionNameSyncStoppedPluginProcessInfo,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operSyncStoppedPluginProcessInfo) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     OperDefNameSyncStoppedPluginProcessInfoTimeout,
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operSyncStoppedPluginProcessInfo) ExtraExecutionName() string {
	return ""
}
