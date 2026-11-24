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
	"errors"
	"fmt"
	"time"

	syncDataUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata/utils"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	workflowStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/globalsettings"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

// ActionNameCorrectRunningProcessStatus defines the V3 running-process expiry action name.
const ActionNameCorrectRunningProcessStatus = "correct_running_process_status"

// NewActionCorrectRunningProcessStatus creates the V3 running-process expiry action.
func NewActionCorrectRunningProcessStatus(capability *Capability) action.Definition {
	return &actionCorrectRunningProcessStatus{
		processStg:  capability.StoragePlugin,
		workflowStg: capability.StorageWorkflow,
	}
}

// ActionParamCorrectRunningProcessStatus defines the V3 expiry check scope.
type ActionParamCorrectRunningProcessStatus struct {
	syncDataUtils.SyncDataActionStandardParam

	HostIDs []int64 `json:"host_ids"`
}

type actionCorrectRunningProcessStatus struct {
	processStg  plugin.IDomainProcess
	workflowStg workflowStg.IStorageScheduledWorkflow
}

// Name returns the action name.
func (act *actionCorrectRunningProcessStatus) Name() string {
	return ActionNameCorrectRunningProcessStatus
}

// Version returns the action version.
func (act *actionCorrectRunningProcessStatus) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the action description.
func (act *actionCorrectRunningProcessStatus) Description() string {
	return "mark expired running V3 processes unknown for a host batch"
}

// Timeout returns the action timeout.
func (act *actionCorrectRunningProcessStatus) Timeout() time.Duration {
	return 10 * time.Minute // nolint: mnd
}

// MaxRetryCount returns the retry count.
func (act *actionCorrectRunningProcessStatus) MaxRetryCount() uint {
	return 2 // nolint: mnd
}

// DelayFn returns the retry delay function.
func (act *actionCorrectRunningProcessStatus) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Tags returns the action tags.
func (act *actionCorrectRunningProcessStatus) Tags() []action.Tag {
	return []action.Tag{}
}

// Queue returns the action queue.
func (act *actionCorrectRunningProcessStatus) Queue() action.Queue {
	return action.QueueDefault
}

// Do marks only V3 running processes unknown after their configured alive-sync cycles expire.
func (act *actionCorrectRunningProcessStatus) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamCorrectRunningProcessStatus)
	if err := conv.MapToStruct(ctx.Data.Content, param); err != nil {
		return fmt.Errorf("failed to parse running process correction parameters: %w", err)
	}
	if len(param.HostIDs) == 0 {
		return errors.New("host IDs are required for running process status correction")
	}
	std := syncDataUtils.NewSyncDataActionStandarder()
	if err := std.Initialize(ctx, param.SyncDataActionStandardParam); err != nil {
		return fmt.Errorf("failed to initialize running process correction context: %w", err)
	}

	deadline, err := processStatusExpiryDeadline(std.Context(), act.workflowStg, "sync_alive_plugin_process_info",
		globalsettings.OperCorrectRunningProcessStatusExpiryCycles, globalsettings.OperCorrectRunningProcessStatusExpiryCyclesDefault)
	if err != nil {
		return fmt.Errorf("failed to resolve V3 running process expiry: %w", err)
	}
	if err := act.processStg.MarkExpiredRunningProcessesUnknown(std.Context(), param.HostIDs, deadline); err != nil {
		return fmt.Errorf("failed to mark expired running V3 processes unknown: %w", err)
	}

	return nil
}

// DisplayNameZh returns the Chinese display name.
func (act *actionCorrectRunningProcessStatus) DisplayNameZh() string {
	return "修正过期运行中插件进程状态"
}

// DisplayNameEn returns the English display name.
func (act *actionCorrectRunningProcessStatus) DisplayNameEn() string {
	return "Correct Expired Running Plugin Process Status"
}
