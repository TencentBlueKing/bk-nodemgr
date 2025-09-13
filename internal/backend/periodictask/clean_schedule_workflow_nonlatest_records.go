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
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/schedule"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	deleteNonLatestScheduleWorkflowOperInstRecordsTimeout  = 5 * time.Minute
	deleteNonLatestScheduleWorkflowOperInstRecordsTaskName = "delete_nonlatest_schedule_workflow_operinst_records"
)

// DeleteNonLatestWorkflowScheduleOperInstRecords delete non-latest workflow schedule operation instance records.
func (pt *PeriodicTask) DeleteNonLatestWorkflowScheduleOperInstRecords(ctx contextx.IContext) error {
	scheduleWorkflows, cnt, err := pt.conf.StgWorkflow.ListScheduleWorkflow(ctx, types.UnlimitedPage(), nil)
	if err != nil {
		pt.conf.Logger.Errorf("list schedule workflow failed: %v", err)
		return fmt.Errorf("list schedule workflow failed: %w", err)
	}

	if cnt == 0 || len(scheduleWorkflows) == 0 {
		pt.conf.Logger.Info("no schedule workflow found, skip delete non-latest workflow schedule oper-inst records")
		return nil
	}

	for _, sw := range scheduleWorkflows {
		opers, _, err := pt.conf.StgWorkflow.ListOperationByTriggerID(ctx, types.UnlimitedPage(), sw.TriggerID)
		if err != nil {
			pt.conf.Logger.Errorf("list operation by trigger-id(%s) failed: %v", sw.TriggerID, err)
			return fmt.Errorf("list operation by trigger-id(%s) failed: %w", sw.TriggerID, err)
		}

		if len(opers) != 1 || len(opers[0].GetNonLastInstanceIDs()) == 0 {
			continue
		}

		nonLatestOperInstIDs := opers[0].GetNonLastInstanceIDs()
		triggerIDs, err := pt.getOperationInstanceRelatedTriggerID(ctx, sw.WorkflowName, nonLatestOperInstIDs...)
		if err != nil {
			pt.conf.Logger.Errorf("get operation-id(%s) related trigger id failed: %v", opers[0].OperationID, err)
			return fmt.Errorf("get operation-id(%s) related trigger id failed: %w", opers[0].OperationID, err)
		}

		for _, triggerID := range triggerIDs {
			if err := pt.deleteOnceTriggerAndRelationd(ctx, triggerID); err != nil {
				pt.conf.Logger.Errorf("delete operation-id(%s) once trigger-id(%s) and related failed: %v",
					opers[0].OperationID, triggerID, err)

				return fmt.Errorf("delete operation-id(%s) once trigger-id(%s) and related failed: %w",
					opers[0].OperationID, triggerID, err)
			}
		}

		if err := pt.conf.StgWorkflow.DeleteOperationInstances(ctx, nonLatestOperInstIDs...); err != nil {
			pt.conf.Logger.Errorf("delete oper-inst-ids(%v) failed: %v", nonLatestOperInstIDs, err)
			return fmt.Errorf("delete oper-inst-ids(%v) failed: %w", nonLatestOperInstIDs, err)
		}

		if err := pt.conf.StgWorkflow.PullOperationInstanceIDsFromOperation(
			ctx, opers[0].OperationID, nonLatestOperInstIDs...); err != nil {
			pt.conf.Logger.Errorf("delete operation-id(%s) related oper-inst-ids(%v) failed: %v",
				opers[0].OperationID, nonLatestOperInstIDs, err)

			return fmt.Errorf("delete operation-id(%s) related oper-inst-ids(%v) failed: %w",
				opers[0].OperationID, nonLatestOperInstIDs, err)
		}
	}

	pt.conf.Logger.Info("delete non-latest workflow schedule oper-inst records success")

	return nil
}

// getOperationInstanceRelatedTriggerID get operation instance related trigger ids.
// schedule workflow action name is `gen_once_trigger_` + workflow name.
func (pt *PeriodicTask) getOperationInstanceRelatedTriggerID(
	ctx contextx.IContext, scheduleWorkflowName string, operInstIDs ...string) ([]string, error) {

	actionName := fmt.Sprintf(schedule.ActionNameGenScheduleOnceTrigger, scheduleWorkflowName)
	triggerIDs := make([]string, 0)
	for _, operInstID := range operInstIDs {
		privateData, err := pt.conf.StgWorkflow.GetActionInstancePrivateData(ctx, operInstID, actionName)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				pt.conf.Logger.Warnf("oper-inst-id(%s) already delete, skip it", operInstID)
				continue
			}

			return nil, fmt.Errorf("get action private data by oper-inst-id(%s) and action-name(%s) failed: %w",
				operInstID, actionName, err)
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
func (pt *PeriodicTask) deleteOnceTriggerAndRelationd(ctx contextx.IContext, triggerID string) error {
	trig, err := pt.conf.StgWorkflow.GetTrigger(ctx, triggerID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			pt.conf.Logger.Warnf("trigger-id(%s) already delete, skip it", triggerID)
			return nil
		}

		return err
	}

	if trig.Category != trigger.CategoryOnce {
		return fmt.Errorf("trigger-id(%s) is not an once trigger", triggerID)
	}

	opers, _, err := pt.conf.StgWorkflow.ListOperationByTriggerID(ctx, types.UnlimitedPage(), triggerID)
	if err != nil {
		return err
	}

	needDeleteOperIDs := make([]string, 0, len(opers))
	needDeleteOperInstIDs := make([]string, 0)
	for _, oper := range opers {
		needDeleteOperIDs = append(needDeleteOperIDs, oper.OperationID)
		needDeleteOperInstIDs = append(needDeleteOperInstIDs, oper.InstanceIDs...)
	}

	err = pt.conf.StgWorkflow.DeleteOperationInstances(ctx, needDeleteOperInstIDs...)
	if err != nil {
		return err
	}

	err = pt.conf.StgWorkflow.DeleteOperations(ctx, needDeleteOperIDs...)
	if err != nil {
		return err
	}

	err = pt.conf.StgWorkflow.DeleteTriggers(ctx, triggerID)
	if err != nil {
		return err
	}

	return nil
}
