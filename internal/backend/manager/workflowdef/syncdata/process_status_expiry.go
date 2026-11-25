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
	"fmt"
	"math"
	"strconv"
	"time"

	workflowStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/globalsettings"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/scheduler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/robfig/cron/v3"
)

// processStatusExpiryDeadline resolves the tenant's current alive-sync interval on every execution.
func processStatusExpiryDeadline(nCtx contextx.IContext, workflowStorage workflowStg.IStorageScheduledWorkflow,
	workflowName, settingName, defaultCycles string) (time.Time, error) {

	cycles, err := strconv.ParseInt(globalsettings.Get(nCtx, settingName, defaultCycles), 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse %s: %w", settingName, err)
	}
	if cycles <= 0 {
		return time.Time{}, fmt.Errorf("%s must be positive, got %d", settingName, cycles)
	}

	workflows, total, err := workflowStorage.ListScheduledWorkflow(nCtx, types.SingleItemPage(),
		&types.ScheduledWorkflowCondition{ExactInclude: &types.ScheduledWorkflowExactFields{
			WorkflowName: []string{workflowName},
		}})
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to list alive-sync workflow %s: %w", workflowName, err)
	}
	if total != 1 || len(workflows) != 1 {
		return time.Time{}, fmt.Errorf("expected one alive-sync workflow %s, got %d", workflowName, total)
	}

	schedule, err := scheduler.CronParser().Parse(workflows[0].Interval)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse alive-sync interval %q: %w", workflows[0].Interval, err)
	}
	interval, ok := schedule.(cron.ConstantDelaySchedule)
	if !ok {
		return time.Time{}, fmt.Errorf("alive-sync workflow %s requires an @every interval for process expiry", workflowName)
	}
	if interval.Delay <= 0 || cycles > math.MaxInt64/int64(interval.Delay) {
		return time.Time{}, fmt.Errorf("invalid process expiry duration for workflow %s and %d cycles", workflowName, cycles)
	}

	return time.Now().Add(-interval.Delay * time.Duration(cycles)), nil
}
