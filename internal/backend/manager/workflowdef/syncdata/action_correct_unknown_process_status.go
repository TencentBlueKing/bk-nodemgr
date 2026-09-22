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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

// ActionNameCorrectUnknownProcessStatus defines the action name.
const ActionNameCorrectUnknownProcessStatus = "correct_unknown_process_status"

// Process observations older than two days no longer establish the current status.
const processStatusSyncTimeout = 48 * time.Hour

// NewActionCorrectUnknownProcessStatus creates an action that marks stale process observations unknown.
func NewActionCorrectUnknownProcessStatus(capability *Capability) action.Definition {
	return &actionCorrectUnknownProcessStatus{
		processStg: capability.StoragePlugin,
	}
}

// ActionParamCorrectUnknownProcessStatus defines the process expiry check scope.
type ActionParamCorrectUnknownProcessStatus struct {
	syncDataUtils.SyncDataActionStandardParam

	HostIDs []int64 `json:"host_ids"`
}

type actionCorrectUnknownProcessStatus struct {
	processStg plugin.IDomainProcess
}

// Name returns the action name.
func (act *actionCorrectUnknownProcessStatus) Name() string {
	return ActionNameCorrectUnknownProcessStatus
}

// Version returns the action version.
func (act *actionCorrectUnknownProcessStatus) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the action description.
func (act *actionCorrectUnknownProcessStatus) Description() string {
	return "mark expired process status unknown for a host batch"
}

// Timeout returns the action timeout.
func (act *actionCorrectUnknownProcessStatus) Timeout() time.Duration {
	return 10 * time.Minute // nolint: mnd
}

// MaxRetryCount returns the retry count.
func (act *actionCorrectUnknownProcessStatus) MaxRetryCount() uint {
	return 2 // nolint: mnd
}

// DelayFn returns the retry delay function.
func (act *actionCorrectUnknownProcessStatus) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Tags returns the action tags.
func (act *actionCorrectUnknownProcessStatus) Tags() []action.Tag {
	return []action.Tag{}
}

// Do marks expired processes unknown and records the correction time as their last sync time.
func (act *actionCorrectUnknownProcessStatus) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamCorrectUnknownProcessStatus)
	if err := conv.MapToStruct(ctx.Data.Content, param); err != nil {
		return fmt.Errorf("failed to parse process status correction parameters: %w", err)
	}
	if len(param.HostIDs) == 0 {
		return errors.New("host IDs are required for process status correction")
	}
	std := syncDataUtils.NewSyncDataActionStandarder()
	if err := std.Initialize(ctx, param.SyncDataActionStandardParam); err != nil {
		return err
	}

	deadline := time.Now().Add(-processStatusSyncTimeout)
	if err := act.processStg.MarkExpiredProcessesUnknown(std.Context(), param.HostIDs, deadline); err != nil {
		return fmt.Errorf("failed to mark expired processes unknown: %w", err)
	}

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionCorrectUnknownProcessStatus) DisplayNameZh() string {
	return "修正过期进程状态"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionCorrectUnknownProcessStatus) DisplayNameEn() string {
	return "Correct Expired Process Status"
}
