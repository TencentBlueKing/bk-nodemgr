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
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/schedule/utils"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameSyncSharedReleases defines the shared release synchronization action name.
	ActionNameSyncSharedReleases = "sync_shared_releases"
)

// NewActionSyncSharedReleases creates the shared release synchronization action.
func NewActionSyncSharedReleases(capability *Capability) action.Definition {
	return &actionSyncSharedReleases{handler: capability.FileHandler}
}

type actionSyncSharedReleases struct {
	handler file.ISharedReleaseSyncHandler
}

// ActionParamSyncSharedReleases defines the shared release synchronization action parameters.
type ActionParamSyncSharedReleases struct {
	utils.ScheduleActionStandardParam
}

// Name returns the name.
func (act *actionSyncSharedReleases) Name() string {
	return ActionNameSyncSharedReleases
}

// DisplayNameZh returns the Chinese display name.
func (act *actionSyncSharedReleases) DisplayNameZh() string {
	return "同步共享制品"
}

// DisplayNameEn returns the English display name.
func (act *actionSyncSharedReleases) DisplayNameEn() string {
	return "Sync Shared Releases"
}

// Version returns the version.
func (act *actionSyncSharedReleases) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description.
func (act *actionSyncSharedReleases) Description() string {
	return "synchronize shared releases from the system tenant"
}

// Timeout returns the timeout.
func (act *actionSyncSharedReleases) Timeout() time.Duration {
	return 10 * time.Minute // nolint:mnd
}

// MaxRetryCount returns the max retry count.
func (act *actionSyncSharedReleases) MaxRetryCount() uint {
	return 3 // nolint:mnd
}

// DelayFn returns the retry delay function.
func (act *actionSyncSharedReleases) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Tags returns the tags.
func (act *actionSyncSharedReleases) Tags() []action.Tag {
	return []action.Tag{}
}

// Do synchronizes shared releases for the scheduled workflow tenant.
func (act *actionSyncSharedReleases) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamSyncSharedReleases)
	if err := conv.MapToStruct(ctx.Data.Content, param); err != nil {
		return err
	}

	std := utils.NewScheduleActionStandarder()
	if err := std.Initialize(ctx, param.ScheduleActionStandardParam); err != nil {
		return err
	}

	if err := act.handler.SyncSharedReleases(std.Context()); err != nil {
		return fmt.Errorf("failed to sync shared releases: %w", err)
	}

	return nil
}
