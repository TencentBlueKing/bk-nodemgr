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

package schedule

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/schedule/scheduleaction"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// OperDefNameScheduledSyncStoppedPluginProcessInfo scheduled sync stopped plugin process info.
const OperDefNameScheduledSyncStoppedPluginProcessInfo = "scheduled_sync_stopped_plugin_process_info"

// NewOperSyncStoppedPluginProcessInfo new an operation.
func NewOperSyncStoppedPluginProcessInfo(param OperParamSyncStoppedPluginProcessInfo) operation.Definition {
	return &operScheduledSyncStoppedPluginProcessInfo{
		param: param,
	}
}

type operScheduledSyncStoppedPluginProcessInfo struct {
	param OperParamSyncStoppedPluginProcessInfo
}

// OperParamSyncStoppedPluginProcessInfo defines the parameters for operScheduledSyncStoppedPluginProcessInfo.
type OperParamSyncStoppedPluginProcessInfo struct {
	scheduleaction.ScheduleActionStandardParam
}

// Name returns the name.
func (oper *operScheduledSyncStoppedPluginProcessInfo) Name() string {
	return OperDefNameScheduledSyncStoppedPluginProcessInfo
}

// ActionDefNames returns the action def names.
func (oper *operScheduledSyncStoppedPluginProcessInfo) ActionDefNames() []string {
	return []string{
		syncdata.ActionNameGenOperSyncStoppedPluginProcessInfo,
	}
}

// DefaultParameters returns the default parameters.
func (oper *operScheduledSyncStoppedPluginProcessInfo) DefaultParameters() operation.Param {
	return operation.Param{
		Timeout:     10 * time.Minute, // nolint:mnd
		InitContent: conv.StructToMapIgnoreError(oper.param),
	}
}

// ExtraExecutionName returns the extra execution definition name.
func (oper *operScheduledSyncStoppedPluginProcessInfo) ExtraExecutionName() string {
	return OperExtraExecutionName
}
