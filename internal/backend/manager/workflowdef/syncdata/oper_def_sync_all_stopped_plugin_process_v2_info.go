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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// OperDefNameSyncAllStoppedPluginProcessV2Info defines the operation def name.
const OperDefNameSyncAllStoppedPluginProcessV2Info = "sync_all_stopped_plugin_process_v2_info"

// NewOperSyncAllStoppedPluginProcessV2Info new an operation.
func NewOperSyncAllStoppedPluginProcessV2Info(
	param OperParamSyncAllStoppedPluginProcessV2Info) operation.Definition {

	return &operSyncAllStoppedPluginProcessV2Info{
		param: param,
	}
}

type operSyncAllStoppedPluginProcessV2Info struct {
	param OperParamSyncAllStoppedPluginProcessV2Info
}

// OperParamSyncAllStoppedPluginProcessV2Info defines the parameters for operSyncAllStoppedPluginProcessV2Info.
type OperParamSyncAllStoppedPluginProcessV2Info struct {
	TenantID string `json:"tenant_id"`
	Operator string `json:"operator"`
}

// Name returns the name.
func (oper *operSyncAllStoppedPluginProcessV2Info) Name() string {
	return OperDefNameSyncAllStoppedPluginProcessV2Info
}

// ActionDefNames returns the action def names.
func (oper *operSyncAllStoppedPluginProcessV2Info) ActionDefNames() []string {
	return []string{
		ActionNameGenOperSyncStoppedPluginProcessV2Info,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operSyncAllStoppedPluginProcessV2Info) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     OperDefNameSyncStoppedPluginProcessV2InfoTimeout,
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operSyncAllStoppedPluginProcessV2Info) ExtraExecutionName() string {
	return ""
}
