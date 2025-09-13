/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package manager

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/schedule"
)

// registerPeriodicTriggerActions registers the action definitions for periodic trigger actions.
func (mgr *Manager) registerActionDefsPeriodicOperation() error {
	if err := mgr.registerActionDefSchedule(); err != nil {
		return fmt.Errorf("register action def schedule failed, err: %w", err)
	}

	return nil
}

// registerActionDefSchedule registers the action definitions for schedule operations.
// nolint: lll
func (mgr *Manager) registerActionDefSchedule() error {
	return mgr.workflowMgr.RegisterActions(
		schedule.NewActionGenScheduleOnceTrigger(SyncCmdbHostWorkflowName, mgr.conf.StorageWorkflow, mgr.LaunchSyncBizAndHost),
		schedule.NewActionGenScheduleOnceTrigger(SyncGseAgentStateWorkflowName, mgr.conf.StorageWorkflow, mgr.LaunchSyncAllAgentState),
		schedule.NewActionGenScheduleOnceTrigger(SyncCmdbNetworkAreaWorkflowName, mgr.conf.StorageWorkflow, mgr.LaunchSyncNetworkArea),
		schedule.NewActionGenScheduleOnceTrigger(WatchAndApplyCMDBResourceWorkflowName, mgr.conf.StorageWorkflow, mgr.LaunchWatchAndApplyCMDBResource),
	)
}
