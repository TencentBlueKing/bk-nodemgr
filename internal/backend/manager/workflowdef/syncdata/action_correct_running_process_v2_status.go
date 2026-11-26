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

// ActionNameCorrectRunningProcessV2Status defines the V2 running-process expiry action name.
const ActionNameCorrectRunningProcessV2Status = "correct_running_process_v2_status"

// NewActionCorrectRunningProcessV2Status creates the V2 running-process expiry action.
func NewActionCorrectRunningProcessV2Status(capability *Capability) action.Definition {
	return &actionCorrectRunningProcessV2Status{
		processStg:  capability.StoragePlugin,
		workflowStg: capability.StorageWorkflow,
	}
}

// ActionParamCorrectRunningProcessV2Status defines the V2 expiry check scope.
type ActionParamCorrectRunningProcessV2Status struct {
	syncDataUtils.SyncDataActionStandardParam

	HostIDs []int64 `json:"host_ids"`
}

type actionCorrectRunningProcessV2Status struct {
	processStg  plugin.IDomainProcessV2
	workflowStg workflowStg.IStorageScheduledWorkflow
}

// Name returns the action name.
func (act *actionCorrectRunningProcessV2Status) Name() string {
	return ActionNameCorrectRunningProcessV2Status
}

// Version returns the action version.
func (act *actionCorrectRunningProcessV2Status) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the action description.
func (act *actionCorrectRunningProcessV2Status) Description() string {
	return "mark expired running V2 processes unknown for a host batch"
}

// Timeout returns the action timeout.
func (act *actionCorrectRunningProcessV2Status) Timeout() time.Duration {
	return 10 * time.Minute // nolint: mnd
}

// MaxRetryCount returns the retry count.
func (act *actionCorrectRunningProcessV2Status) MaxRetryCount() uint {
	return 2 // nolint: mnd
}

// DelayFn returns the retry delay function.
func (act *actionCorrectRunningProcessV2Status) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Tags returns the action tags.
func (act *actionCorrectRunningProcessV2Status) Tags() []action.Tag {
	return []action.Tag{}
}

// Queue returns the action queue.
func (act *actionCorrectRunningProcessV2Status) Queue() action.Queue {
	return action.QueueDefault
}

// Do marks only V2 running processes unknown after their configured alive-sync cycles expire.
func (act *actionCorrectRunningProcessV2Status) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamCorrectRunningProcessV2Status)
	if err := conv.MapToStruct(ctx.Data.Content, param); err != nil {
		return fmt.Errorf("failed to parse running V2 process correction parameters: %w", err)
	}
	if len(param.HostIDs) == 0 {
		return errors.New("host IDs are required for running V2 process status correction")
	}
	std := syncDataUtils.NewSyncDataActionStandarder()
	if err := std.Initialize(ctx, param.SyncDataActionStandardParam); err != nil {
		return fmt.Errorf("failed to initialize running V2 process correction context: %w", err)
	}

	deadline, err := processStatusExpiryDeadline(std.Context(), act.workflowStg, "sync_alive_plugin_process_v2_info",
		globalsettings.OperCorrectRunningProcessV2StatusExpiryCycles, globalsettings.OperCorrectRunningProcessV2StatusExpiryCyclesDefault)
	if err != nil {
		return fmt.Errorf("failed to resolve V2 running process expiry: %w", err)
	}
	if err := act.processStg.MarkExpiredRunningProcessesV2Unknown(std.Context(), param.HostIDs, deadline); err != nil {
		return fmt.Errorf("failed to mark expired running V2 processes unknown: %w", err)
	}

	return nil
}

// DisplayNameZh returns the Chinese display name.
func (act *actionCorrectRunningProcessV2Status) DisplayNameZh() string {
	return "修正过期运行中旧版插件进程状态"
}

// DisplayNameEn returns the English display name.
func (act *actionCorrectRunningProcessV2Status) DisplayNameEn() string {
	return "Correct Expired Running V2 Plugin Process Status"
}
