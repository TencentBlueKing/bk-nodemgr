/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package periodictask implements non business periodic tasks.
package periodictask

import (
	"context"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/schedule"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

const (
	deleteNonLatestScheduleWorkflowOperInstRecordsTimeout  = 5 * time.Minute
	deleteNonLatestScheduleWorkflowOperInstRecordsTaskName = "delete_nonlatest_schedule_workflow_operinst_records"
)

// DeleteNonLatestWorkflowScheduleOperInstRecords delete non-latest workflow schedule operation instance records.
func (pt *PeriodicTask) DeleteNonLatestWorkflowScheduleOperInstRecords(ctx context.Context) error {
	sw, cnt, err := pt.conf.StgScheduleWorkflow.ListScheduleWorkflow(ctx, types.UnlimitedPage(), nil)
	if err != nil {
		pt.conf.Logger.Errorf("list schedule workflow failed: %v", err)
		return fmt.Errorf("list schedule workflow failed: %w", err)
	}

	if cnt == 0 || len(sw) == 0 {
		pt.conf.Logger.Info("no schedule workflow found, skip delete non-latest workflow schedule oper-inst records")
		return nil
	}

	for _, s := range sw {
		opers, cnt, err := pt.conf.StgOperation.ListOperationByTrigger(ctx, types.UnlimitedPage(), s.TriggerID)
		if err != nil {
			pt.conf.Logger.Errorf("list operation by trigger-id(%s) failed: %v", s.TriggerID, err)
			return fmt.Errorf("list operation by trigger-id(%s) failed: %v", s.TriggerID, err)
		}

		if cnt != 1 {
			continue
		}

		triggerIDs, err := pt.getOperationInstanceRelatedTriggerID(ctx, opers[0], s.WorkflowName)
		if err != nil {
			pt.conf.Logger.Errorf("get operation-id(%s) related trigger id failed: %v", opers[0].OperationID, err)
			return fmt.Errorf("get operation-id(%s) related trigger id failed: %w", opers[0].OperationID, err)
		}

		for _, triggerID := range triggerIDs {
			err = pt.deleteOnceTriggerAndRelationd(ctx, triggerID)
			if err != nil {
				pt.conf.Logger.Errorf("delete once trigger-id(%s) and related failed: %v", triggerID, err)
				return fmt.Errorf("delete once trigger-id(%s) and related failed: %w", triggerID, err)
			}
		}

		err = pt.conf.StgOperInst.DeleteOperationInstances(ctx, opers[0].GetNonLastInstanceIDs()...)
		if err != nil {
			pt.conf.Logger.Errorf("delete oper-inst-ids(%v) failed: %v", opers[0].GetNonLastInstanceIDs(), err)
			return fmt.Errorf("delete oper-inst-ids(%v) failed: %w", opers[0].GetNonLastInstanceIDs(), err)
		}

		opers[0].InstanceIDs = []string{opers[0].GetLastInstanceID()}
		err = pt.conf.StgOperation.UpsertOperation(ctx, opers[0])
		if err != nil {
			pt.conf.Logger.Errorf("upsert operation-id(%s) failed: %v", opers[0].OperationID, err)
			return err
		}
	}

	pt.conf.Logger.Info("delete non-latest workflow schedule oper-inst records success")

	return nil
}

// getOperationInstanceRelatedTriggerID get operation instance related trigger ids.
// schedule workflow action name is `gen_once_trigger_` + workflow name.
func (pt *PeriodicTask) getOperationInstanceRelatedTriggerID(
	ctx context.Context, oper *operation.Operation, scheduleWorkflowName string) ([]string, error) {

	actionName := fmt.Sprintf(schedule.ActionNameGenScheduleOnceTrigger, scheduleWorkflowName)
	triggerIDs := make([]string, 0)
	for _, operInstID := range oper.GetNonLastInstanceIDs() {
		privateData, err := pt.conf.StgOperInst.GetActionInstancePrivateData(ctx, operInstID, actionName)
		if err != nil {
			return nil, err
		}

		triggerID, ok := privateData["child_trigger_id"]
		if !ok {
			continue
		}

		triggerIDStr, ok := triggerID.(string)
		if !ok || triggerIDStr == "" {
			continue
		}

		triggerIDs = append(triggerIDs, triggerIDStr)
	}

	return triggerIDs, nil
}

// deleteOnceTriggerAndRelationd delete once trigger and its related operations and operation instances.
func (pt *PeriodicTask) deleteOnceTriggerAndRelationd(ctx context.Context, triggerID string) error {
	trig, err := pt.conf.StgTrigger.GetTrigger(ctx, triggerID)
	if err != nil {
		return err
	}

	if trig.Category != trigger.CategoryOnce {
		return fmt.Errorf("trigger-id(%s) is not an once trigger", triggerID)
	}

	opers, cnt, err := pt.conf.StgOperation.ListOperationByTrigger(ctx, types.UnlimitedPage(), triggerID)
	if err != nil {
		return err
	}

	needDeleteOperIDs := make([]string, 0, cnt)
	needDeleteOperInstIDs := make([]string, 0)
	for _, oper := range opers {
		needDeleteOperIDs = append(needDeleteOperIDs, oper.OperationID)
		needDeleteOperInstIDs = append(needDeleteOperInstIDs, oper.InstanceIDs...)
	}

	err = pt.conf.StgOperInst.DeleteOperationInstances(ctx, needDeleteOperInstIDs...)
	if err != nil {
		return err
	}

	err = pt.conf.StgOperation.DeleteOperations(ctx, needDeleteOperIDs...)
	if err != nil {
		return err
	}

	err = pt.conf.StgTrigger.DeleteTriggers(ctx, triggerID)
	if err != nil {
		return err
	}

	return nil
}
