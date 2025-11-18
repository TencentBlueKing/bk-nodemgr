/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package workflow provides workflow storage interface.
package workflow

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operinstdata"
	scheduledworkflow "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/scheduled-workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/stopoperinst"
	daoTrigger "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/trigger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/scheduler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/common"
	workoper "github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/singleflight"
)

// constants ...
const (
	// StorageName defines the storage name.
	StorageName       = "workflow"
	taskInterval      = 10 * time.Second
	taskTimeout       = 20 * time.Second
	syncOperationTask = "sync stopping operation inst"
)

// NewStorage creates a new workflow storage.
func NewStorage(client *mongo.Client, database string) (IStorage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &Storage{
		Storage: basestorage.Storage{
			Name:     StorageName,
			Database: client.Database(database),
		},
	}
	if err := basestorage.InitStorage(&s.Storage,
		basestorage.WithStartFunc(s.initDao),
		basestorage.WithCheckFunc(s.check)); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to new storage")

		return nil, err
	}

	return s, nil
}

// Storage implements IStorage.
type Storage struct {
	basestorage.Storage

	daoTrigger           daoTrigger.IHandler
	daoOperation         operation.IHandler
	daoScheduledWorkflow scheduledworkflow.IHandler
	daoOperInstData      operinstdata.IHandler
	daoStopOperInst      stopoperinst.Handler

	// stop event subscriptions
	stopEventSubsMap      map[string]*StopEventSubscription
	stopEventSubsMapMutex sync.RWMutex

	stopOperInsts      map[string]struct{}
	stopOperInstsMutex sync.RWMutex

	sg singleflight.Group
}

func (s *Storage) initDao() error {
	s.daoTrigger = daoTrigger.New(s.Database)
	s.daoOperation = operation.New(s.Database)
	s.daoScheduledWorkflow = scheduledworkflow.New(s.Database)
	s.daoOperInstData = operinstdata.New(s.Database)
	s.daoStopOperInst = stopoperinst.New(s.Database)

	s.stopEventSubsMap = make(map[string]*StopEventSubscription)
	s.stopOperInsts = make(map[string]struct{})

	err := s.registerScheduler()
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to register scheduler")

		return fmt.Errorf("failed to register scheduler: %w", err)
	}

	return nil
}

func (s *Storage) check() error {
	if s.daoTrigger == nil {
		return errors.New("trigger dao is nil")
	}

	return nil
}

func (s *Storage) registerScheduler() error {
	s.Scheduler = scheduler.NewScheduler()
	err := s.Scheduler.RegisterTask(scheduler.NewTask(
		syncOperationTask,
		taskInterval,
		taskTimeout,
		s.syncStopOperInsts,
	))
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to register sync stopping operation instance task")

		return fmt.Errorf("failed to register sync stopping operation instance task: %w", err)
	}

	go s.daoStopOperInst.WatchInsert(func(stopInstID string) {
		s.stopOperInstsMutex.Lock()
		defer s.stopOperInstsMutex.Unlock()
		s.stopOperInsts[stopInstID] = struct{}{}

		go func() {
			err := s.checkNotifyStopping(contextx.New(s.Ctx))
			if err != nil {
				logger.G.Sys().WithErr(err).Warn("failed to check notify stopping")
			}
		}()
	})

	return nil
}

func (s *Storage) metric() *storage.MetricData {
	return storage.Metric(StorageName)
}

// CreateTrigger creates a new trigger.
func (s *Storage) CreateTrigger(nCtx contextx.IContext, trig *trigger.Trigger) error {
	var err error

	// record metric.
	metric := s.metric().Start("create_trigger")
	defer metric.End(err)

	if err = s.createTrigger(nCtx, trig); err != nil {
		logger.G.Sys().WithErr(err).With("trigger-id", trig.TriggerID).Error("failed to create trigger")

		return fmt.Errorf("failed to create trigger, trigger-id(%s): %w", trig.TriggerID, err)
	}

	return nil
}

// UpdateTrigger updates a trigger.
func (s *Storage) UpdateTrigger(nCtx contextx.IContext, trig *trigger.Trigger) error {
	var err error

	// record metric.
	metric := s.metric().Start("update_trigger")
	defer metric.End(err)

	if err = s.updateTrigger(nCtx, trig); err != nil {
		logger.G.Sys().WithErr(err).With("trigger-id", trig.TriggerID).Error("failed to update trigger")

		return fmt.Errorf("failed to update trigger, trigger-id(%s): %w", trig.TriggerID, err)
	}

	return nil
}

// SwitchTriggerActive switches a trigger active status.
func (s *Storage) SwitchTriggerActive(nCtx contextx.IContext, triggerID string, active bool) error {
	var err error

	// record metric.
	metric := s.metric().Start("switch_trigger_alive")
	defer metric.End(err)

	if err = s.switchTriggerActive(nCtx, triggerID, active); err != nil {
		logger.G.Sys().WithErr(err).With("trigger-id", triggerID).Error("failed to switch trigger active state")

		return fmt.Errorf("failed to switch trigger active state, trigger-id(%s): %w", triggerID, err)
	}

	return nil
}

// GetTrigger gets a trigger by triggerID.
func (s *Storage) GetTrigger(nCtx contextx.IContext, triggerID string) (*trigger.Trigger, error) {
	var (
		err  error
		data *trigger.Trigger
	)

	// record metric.
	metric := s.metric().Start("get_trigger")
	defer metric.End(err)

	if data, err = s.getTrigger(nCtx, triggerID); err != nil {
		logger.G.Sys().WithErr(err).With("trigger-id", triggerID).Error("failed to get trigger")

		return nil, fmt.Errorf("failed to get trigger, trigger-id(%s): %w", triggerID, err)
	}

	return data, nil
}

// ListActiveTrigger lists active triggers by given category.
func (s *Storage) ListActiveTrigger(nCtx contextx.IContext, category trigger.Category) ([]*trigger.Trigger, error) {
	var (
		err     error
		results []*trigger.Trigger
	)

	// record metric.
	metric := s.metric().Start("list_alive_trigger")
	defer metric.End(err)

	if results, err = s.listActiveTrigger(nCtx, category); err != nil {
		logger.G.Sys().WithErr(err).With("category", category).Error("failed to list active triggers")

		return nil, fmt.Errorf("failed to list active triggers, category(%s): %w", category, err)
	}

	return results, nil
}

// ListTrigger lists triggers by given category.
func (s *Storage) ListTrigger(nCtx contextx.IContext, page types.Page, category trigger.Category) ([]*trigger.Trigger, int64, error) {
	var (
		results []*trigger.Trigger
		num     int64
		err     error
	)

	// record metric.
	metric := s.metric().Start("list_trigger")
	defer metric.End(err)

	if results, num, err = s.listTrigger(nCtx, page, category); err != nil {
		logger.G.Sys().WithErr(err).With("category", category).Error("failed to list triggers")

		return nil, 0, fmt.Errorf("failed to list triggers, category(%s): %w", category, err)
	}

	return results, num, nil
}

// DeleteTriggers deletes triggers by given trigger IDs.
func (s *Storage) DeleteTriggers(nCtx contextx.IContext, triggerIDs ...string) error {
	var err error

	// record metric.
	metric := s.metric().Start("delete_triggers")
	defer metric.End(err)

	if err = s.deleteTriggers(nCtx, triggerIDs...); err != nil {
		logger.G.Sys().WithErr(err).With("trigger-ids", triggerIDs).Error("failed to delete triggers")

		return fmt.Errorf("failed to delete triggers, trigger-ids(%v): %w", triggerIDs, err)
	}

	return nil
}

// ExistTrigger checks if a trigger exists.
func (s *Storage) ExistTrigger(nCtx contextx.IContext, triggerID string) (bool, error) {
	var (
		exist bool
		err   error
	)

	// record metric.
	metric := s.metric().Start("exist_trigger")
	defer metric.End(err)

	if exist, err = s.existTrigger(nCtx, triggerID); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to check trigger exist")

		return false, fmt.Errorf("failed to check trigger exist: %w", err)
	}

	return exist, nil
}

// ListScheduledWorkflow lists scheduled workflow by page and conditions.
func (s *Storage) ListScheduledWorkflow(
	nCtx contextx.IContext, page types.Page, conditions ...*types.ScheduledWorkflowCondition) (
	[]*types.ScheduledWorkflow, int64, error) {

	var (
		results []*types.ScheduledWorkflow
		num     int64
		err     error
	)

	// record metric.
	metric := s.metric().Start("list_scheduled_workflow")
	defer metric.End(err)

	if results, num, err = s.listScheduledWorkflow(nCtx, page, conditions...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to list scheduled workflows")

		return nil, 0, fmt.Errorf("failed to list scheduled workflows: %w", err)
	}

	return results, num, nil
}

// CountScheduledWorkflow counts scheduled workflow by conditions.
func (s *Storage) CountScheduledWorkflow(
	nCtx contextx.IContext, conditions ...*types.ScheduledWorkflowCondition) (
	int64, error) {

	var (
		num int64
		err error
	)

	// record metric.
	metric := s.metric().Start("count_scheduled_workflow")
	defer metric.End(err)

	if num, err = s.countScheduledWorkflow(nCtx, conditions...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to count scheduled workflows")

		return 0, fmt.Errorf("failed to count scheduled workflows: %w", err)
	}

	return num, nil
}

// GetScheduledWorkflow gets a scheduled workflow by workflow-id.
func (s *Storage) GetScheduledWorkflow(nCtx contextx.IContext, workflowID string) (*types.ScheduledWorkflow, error) {
	var (
		workflow *types.ScheduledWorkflow
		err      error
	)

	// record metric.
	metric := s.metric().Start("get_scheduled_workflow")
	defer metric.End(err)

	if workflow, err = s.getScheduledWorkflow(nCtx, workflowID); err != nil {
		logger.G.Sys().WithErr(err).With("workflow-id", workflowID).Error("failed to get scheduled workflow")

		return nil, fmt.Errorf("failed to get workflow, workflow-id(%s): %w", workflowID, err)
	}

	return workflow, nil
}

// CreateScheduledWorkflow creates a new scheduled workflow.
func (s *Storage) CreateScheduledWorkflow(nCtx contextx.IContext, workflow *types.ScheduledWorkflow) error {
	var err error

	// record metric.
	metric := s.metric().Start("create_scheduled_workflow")
	defer metric.End(err)

	if err = s.createScheduledWorkflow(nCtx, workflow); err != nil {
		logger.G.Sys().WithErr(err).With("workflow-id", workflow.WorkflowID).Error("failed to create scheduled workflow")

		return fmt.Errorf("failed to create scheduled workflow, workflow-id(%s): %w", workflow.WorkflowID, err)
	}

	return nil
}

// UpdateScheduledWorkflowTriggerID updates a scheduled workflow's trigger ID.
func (s *Storage) UpdateScheduledWorkflowTriggerID(nCtx contextx.IContext, workflowID, triggerID string) error {
	var err error

	// record metric.
	metric := s.metric().Start("update_scheduled_workflow_trigger_id")
	defer metric.End(err)

	if err = s.updateScheduledWorkflowTriggerID(nCtx, workflowID, triggerID); err != nil {
		logger.G.Sys().WithErr(err).With("workflow-id", workflowID).Error("failed to update scheduled workflow trigger id")

		return err
	}

	return nil
}

// UpdateScheduledWorkflowPrivateData updates a scheduled workflow's private data.
func (s *Storage) UpdateScheduledWorkflowPrivateData(nCtx contextx.IContext, workflowID string, privateData map[string]any) error {
	var err error

	// record metric.
	metric := s.metric().Start("update_scheduled_workflow_private_data")
	defer metric.End(err)

	if err = s.updateScheduledWorkflowPrivateData(nCtx, workflowID, privateData); err != nil {
		logger.G.Sys().WithErr(err).With("workflow-id", workflowID).Error("failed to update scheduled workflow private data")

		return err
	}

	return nil
}

// SwitchScheduleWorkflow enables or disables a scheduled workflow.
func (s *Storage) SwitchScheduleWorkflow(nCtx contextx.IContext, workflowID string, enable bool) error {
	var err error

	// record metric.
	metric := s.metric().Start("switch_scheduled_workflow")
	defer metric.End(err)

	if err := s.switchScheduleWorkflow(nCtx, workflowID, enable); err != nil {
		logger.G.Sys().WithErr(err).With("workflow-id", workflowID).Error("failed to switch scheduled workflow")

		return err
	}

	return nil
}

// GetOperation get operation by operationID.
func (s *Storage) GetOperation(nCtx contextx.IContext, operationID string) (*workoper.Operation, error) {
	var (
		oper *workoper.Operation
		err  error
	)
	// record metric.
	metric := s.metric().Start("get_operation")
	defer metric.End(err)

	if oper, err = s.getOperation(nCtx, operationID); err != nil {
		logger.G.Sys().WithErr(err).With("operation-id", operationID).Error("failed to get operations")

		return nil, fmt.Errorf("failed to get operations, operationID(%s): %w", operationID, err)
	}

	return oper, nil
}

// UpsertOperation upsert operation.
func (s *Storage) UpsertOperation(nCtx contextx.IContext, operation *workoper.Operation) error {
	var err error

	// record metric.
	metric := s.metric().Start("upsert_operation")
	defer metric.End(err)

	if err = s.upsertOperation(nCtx, operation); err != nil {
		logger.G.Sys().WithErr(err).With("operation-id", operation.OperationID).Error("failed to upsert operation")

		return fmt.Errorf("failed to upsert operation, operation-id(%s): %w", operation.OperationID, err)
	}

	return nil
}

// ListOperationByTriggerID lists operation by trigger id.
func (s *Storage) ListOperationByTriggerID(
	nCtx contextx.IContext, page types.Page, triggerID ...string) (
	[]*workoper.Operation, int64, error) {

	var (
		opers []*workoper.Operation
		num   int64
		err   error
	)

	// record metric.
	metric := s.metric().Start("list_operation_by_trigger_id")
	defer metric.End(err)

	if opers, num, err = s.listOperationByTriggerID(nCtx, page, triggerID...); err != nil {
		logger.G.Sys().WithErr(err).With("trigger-ids", triggerID).Error("failed to list operations by trigger id")

		return nil, 0, fmt.Errorf("failed to list operations by trigger id, trigger-ids(%v): %w", triggerID, err)
	}

	return opers, num, nil
}

// ListOperationByOperationID lists operation by operation id.
func (s *Storage) ListOperationByOperationID(
	nCtx contextx.IContext, operationID ...string) (
	[]*workoper.Operation, int64, error) {

	var (
		opers []*workoper.Operation
		num   int64
		err   error
	)

	// record metric.
	metric := s.metric().Start("list_operation_by_operation_id")
	defer metric.End(err)

	if opers, num, err = s.listOperationByOperationID(nCtx, operationID...); err != nil {
		logger.G.Sys().WithErr(err).With("operation-ids", operationID).Error("failed to list operations by operation id")

		return nil, 0, fmt.Errorf("failed to list operations by operation id, operation-ids(%v): %w", operationID, err)
	}

	return opers, num, nil
}

// ListOperationByParentOperationID lists operation by parent operation ID.
func (s *Storage) ListOperationByParentOperationID(
	nCtx contextx.IContext, page types.Page, parentID ...string) ([]*workoper.Operation, int64, error) {

	var (
		opers []*workoper.Operation
		num   int64
		err   error
	)

	// record metric.
	metric := s.metric().Start("list_operation_by_parent_operation_id")
	defer metric.End(err)

	if opers, num, err = s.listOperationByParentOperationID(nCtx, page, parentID...); err != nil {
		logger.G.Sys().WithErr(err).With("parent-ids", parentID).Error("failed to list operations by parent operation id")

		return nil, 0, fmt.Errorf("failed to list operations by parent operation id, parent-ids(%v): %w", parentID, err)
	}

	return opers, num, nil
}

// ListOperationByNodeWorkflowOperationCondition lists operation by node workflow operation condition.
func (s *Storage) ListOperationByNodeWorkflowOperationCondition(
	nCtx contextx.IContext, page types.Page, condition ...*types.NodeWorkflowOperationCondition) (
	[]*workoper.Operation, int64, error) {

	var (
		opers []*workoper.Operation
		num   int64
		err   error
	)

	// record metric.
	metric := s.metric().Start("list_operation_by_node_workflow_operation_condition")
	defer metric.End(err)

	if opers, num, err = s.listOperationByNodeWorkflowOperationCondition(nCtx, page, condition...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to list operations by node workflow operation condition")

		return nil, 0, fmt.Errorf("failed to list operations by node workflow operation condition: %w", err)
	}

	return opers, num, nil
}

// ListNeedInstantiateOperationByTriggerID lists operations need to be instantiated by trigger id.
func (s *Storage) ListNeedInstantiateOperationByTriggerID(nCtx contextx.IContext, page types.Page, triggerID string) (
	[]*workoper.Operation, int64, error) {

	var (
		opers []*workoper.Operation
		num   int64
		err   error
	)

	// record metric.
	metric := s.metric().Start("list_need_instantiate_operation_by_trigger_id")
	defer metric.End(err)

	if opers, num, err = s.listNeedInstantiateOperationByTriggerID(nCtx, page, triggerID); err != nil {
		logger.G.Sys().WithErr(err).With("trigger-id", triggerID).Error("failed to list need instantiate operations")

		return nil, 0, fmt.Errorf("failed to list need instantiate operations, trigger-id(%s): %w", triggerID, err)
	}

	return opers, num, nil
}

// DeleteOperationsByTriggerID deletes operations by trigger ID.
func (s *Storage) DeleteOperationsByTriggerID(ctx contextx.IContext, triggerID ...string) error {
	var err error

	// record metric.
	metric := s.metric().Start("delete_operations_by_trigger_id")
	defer metric.End(err)

	if err = s.deleteOperationsByTriggerID(ctx, triggerID...); err != nil {
		logger.G.Sys().WithErr(err).With("trigger-id", triggerID).Error("failed to delete operations by trigger id")

		return fmt.Errorf("failed to delete operations by trigger id, trigger-id(%s): %w", triggerID, err)
	}

	return nil
}

// PullOperationInstanceIDsFromOperation pulls operation instance IDs from operation.
func (s *Storage) PullOperationInstanceIDsFromOperation(
	nCtx contextx.IContext, operationID string, operInstIDs ...string) error {

	var err error

	// record metric.
	metric := s.metric().Start("pull_operation_instance_ids_from_operation")
	defer metric.End(err)

	if err = s.pullOperationInstanceIDs(nCtx, operationID, operInstIDs...); err != nil {
		logger.G.Sys().WithErr(err).With("operation-id", operationID, "oper-inst-ids", operInstIDs).Error("failed to pull operation instance ids")

		return fmt.Errorf("failed to pull operation instance ids, operation-id(%s), oper-inst-ids(%v): %w",
			operationID, operInstIDs, err)
	}

	return nil
}

// UpdateOperInstActionStatus update the oper inst action status.
func (s *Storage) UpdateOperInstActionStatus(
	nCtx contextx.IContext, operInstID string, actionName string, status action.State) error {

	var err error

	// record metric.
	metric := s.metric().Start("update_oper_inst_action_status")
	defer metric.End(err)

	if err = s.updateOperInstActionStatus(nCtx, operInstID, actionName, status); err != nil {
		logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID, "action-name", actionName).Error("failed to update oper inst action status")

		return fmt.Errorf("failed to update oper inst action status, oper-inst-id(%s), action-name(%s): %w",
			operInstID, actionName, err)
	}

	return nil
}

// GetActionInstanceData gets full action instance data.
func (s *Storage) GetActionInstanceData(
	nCtx contextx.IContext, operInstID, actionName string) (*action.InstanceData, error) {

	var (
		data *action.InstanceData
		err  error
	)

	// record metric.
	metric := s.metric().Start("get_action_instance_data")
	defer metric.End(err)

	if data, err = s.getActionInstanceData(nCtx, operInstID, actionName); err != nil {
		logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID, "action-name", actionName).Error("failed to get action instance data")

		return nil, fmt.Errorf("failed to get action instance data, oper-inst-id(%s), action-name(%s): %w",
			operInstID, actionName, err)
	}

	return data, nil
}

// GetActionInstanceLifecycle gets action instance lifecycle.
func (s *Storage) GetActionInstanceLifecycle(
	nCtx contextx.IContext, operInstID, actionName string) (*action.Lifecycle, error) {

	var (
		lifecycle *action.Lifecycle
		err       error
	)

	// record metric.
	metric := s.metric().Start("get_action_instance_lifecycle")
	defer metric.End(err)

	if lifecycle, err = s.getActionInstanceLifecycle(nCtx, operInstID, actionName); err != nil {
		logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID, "action-name", actionName).Error("failed to get action instance lifecycle")

		return nil, fmt.Errorf("failed to get action instance lifecycle, oper-inst-id(%s), action-name(%s): %w",
			operInstID, actionName, err)
	}

	return lifecycle, nil
}

// GetActionInstancePrivateData gets action instance private data.
func (s *Storage) GetActionInstancePrivateData(
	nCtx contextx.IContext, operInstID, actionName string) (map[string]any, error) {

	var (
		data map[string]any
		err  error
	)

	// record metric.
	metric := s.metric().Start("get_action_instance_private_data")
	defer metric.End(err)

	if data, err = s.getActionInstancePrivateData(nCtx, operInstID, actionName); err != nil {
		logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID, "action-name", actionName).Error("failed to get action instance private data")

		return nil, fmt.Errorf("failed to get action instance private data, oper-inst-id(%s), action-name(%s): %w",
			operInstID, actionName, err)
	}

	return data, nil
}

// UpdateActionInstanceLifecycle updates action instance lifecycle.
func (s *Storage) UpdateActionInstanceLifecycle(
	nCtx contextx.IContext, operInstID, actionName string, lifecycle *action.Lifecycle) error {

	var err error

	// record metric.
	metric := s.metric().Start("update_action_instance_lifecycle")
	defer metric.End(err)

	if err = s.updateActionInstanceLifecycle(nCtx, operInstID, actionName, lifecycle); err != nil {
		logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID, "action-name", actionName).Error("failed to update action instance lifecycle")

		return fmt.Errorf("failed to update action instance lifecycle, oper-inst-id(%s), action-name(%s): %w",
			operInstID, actionName, err)
	}

	return nil
}

// PushActionInstanceMessage pushes action instance message.
func (s *Storage) PushActionInstanceMessage(
	nCtx contextx.IContext, operInstID, actionName string, messages ...common.Message) error {

	var err error

	// record metric.
	metric := s.metric().Start("push_action_instance_message")
	defer metric.End(err)

	if err = s.pushActionInstanceMessage(nCtx, operInstID, actionName, messages...); err != nil {
		logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID, "action-name", actionName).Error("failed to push action instance message")

		return fmt.Errorf("failed to push action instance message, oper-inst-id(%s), action-name(%s): %w",
			operInstID, actionName, err)
	}

	return nil
}

// GetOperationInstanceFullData gets full operation instance data.
func (s *Storage) GetOperationInstanceFullData(
	nCtx contextx.IContext, operInstID string) (*workoper.InstanceData, error) {

	var (
		data *workoper.InstanceData
		err  error
	)

	// record metric.
	metric := s.metric().Start("get_operation_instance_full_data")
	defer metric.End(err)

	if data, err = s.getOperationInstanceData(nCtx, operInstID); err != nil {
		logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID).Error("failed to get operation instance full data")

		return nil, fmt.Errorf("failed to get operation instance data, oper-inst-id(%s): %w", operInstID, err)
	}

	return data, nil
}

// GetOperationInstanceBriefData gets brief operation instance data.
func (s *Storage) GetOperationInstanceBriefData(
	nCtx contextx.IContext, operInstID string) (*workoper.InstanceBriefData, error) {

	var (
		briefData *workoper.InstanceBriefData
		err       error
	)

	// record metric.
	metric := s.metric().Start("get_operation_instance_brief_data")
	defer metric.End(err)

	if briefData, err = s.getOperationInstanceDataBriefData(nCtx, operInstID); err != nil {
		logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID).Error("failed to get operation instance brief data")

		return nil, fmt.Errorf("failed to get operation instance data, oper-inst-id(%s): %w", operInstID, err)
	}

	return briefData, nil
}

// ListOperationInstanceBriefDataWithoutActionInst lists operation instance brief data. without action instance data.
func (s *Storage) ListOperationInstanceBriefDataWithoutActionInst(
	nCtx contextx.IContext, page types.Page, conditions ...*types.OperInstDataCondition) (
	[]*workoper.InstanceBriefData, int64, error) {

	var (
		results []*workoper.InstanceBriefData
		num     int64
		err     error
	)

	// record metric.
	metric := s.metric().Start("list_operation_instance_brief_data_without_action_inst_by_condition")
	defer metric.End(err)

	if results, num, err = s.listOperationInstanceBriefDataWithoutActionInst(
		nCtx, page, conditions...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to list operation instance brief data")

		return nil, 0, fmt.Errorf("failed to list operation instance brief data: %w", err)
	}

	return results, num, nil
}

// CountOperationInstance counts operation instance.
func (s *Storage) CountOperationInstance(
	nCtx contextx.IContext, triggerID string, states ...workoper.State) (int64, error) {

	var (
		num int64
		err error
	)

	// record metric.
	metric := s.metric().Start("count_operation_instance")
	defer metric.End(err)

	if num, err = s.countOperationInstance(nCtx, triggerID, states...); err != nil {
		logger.G.Sys().WithErr(err).With("trigger-id", triggerID, "states", states).Error("failed to count operation instance")

		return 0, fmt.Errorf("failed to count operation instance, trigger-id(%s), states(%v): %w",
			triggerID, states, err)
	}

	return num, nil
}

// ListOperInstanceBriefWithoutActionInstByOperationID lists operation instance brief data.
func (s *Storage) ListOperInstanceBriefWithoutActionInstByOperationID(
	nCtx contextx.IContext, page types.Page, operationID ...string) (
	[]*workoper.InstanceBriefData, int64, error) {

	var (
		results []*workoper.InstanceBriefData
		num     int64
		err     error
	)

	// record metric.
	metric := s.metric().Start("list_operation_instance_brief_without_action_inst_by_operation_id")
	defer metric.End(err)

	if results, num, err = s.listOperationInstanceBriefDataWithoutActionInstByOperationID(
		nCtx, page, operationID...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to list operation instance brief data by operation")

		return nil, 0, fmt.Errorf("failed to list operation instance brief data by operation: %w", err)
	}

	return results, num, nil
}

// ListOperInstanceBriefWithoutActionInstByTriggerID lists operation instance brief data.
func (s *Storage) ListOperInstanceBriefWithoutActionInstByTriggerID(nCtx contextx.IContext, page types.Page, triggerID ...string) (
	[]*workoper.InstanceBriefData, int64, error) {

	var (
		results []*workoper.InstanceBriefData
		num     int64
		err     error
	)

	// record metric.
	metric := s.metric().Start("list_operation_instance_brief_without_action_inst_by_trigger_id")
	defer metric.End(err)

	if results, num, err = s.listOperationInstanceBriefDataWithoutActionInstByTriggerID(
		nCtx, page, triggerID...); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to list operation instance brief data by trigger")

		return nil, 0, fmt.Errorf("failed to list operation instance brief data by trigger: %w", err)
	}

	return results, num, nil
}

// UpsertOperationInstanceData upserts operation instance data.
func (s *Storage) UpsertOperationInstanceData(
	nCtx contextx.IContext, operInstData *workoper.InstanceData) error {

	var err error

	// record metric.
	metric := s.metric().Start("upsert_operation_instance_data")
	defer metric.End(err)

	if err = s.upsertOperationInstanceData(nCtx, operInstData); err != nil {
		logger.G.Sys().WithErr(err).With("oper-inst", operInstData).Error("failed to update operation instance data")

		return fmt.Errorf("failed to update operation instance data, operation-inst(%v): %w", operInstData, err)
	}

	return nil
}

// UpdateOperationInstanceLifecycle updates operation instance lifecycle.
func (s *Storage) UpdateOperationInstanceLifecycle(
	nCtx contextx.IContext, operInstID string, lifecycle *workoper.Lifecycle) error {

	var err error

	// record metric.
	metric := s.metric().Start("update_operation_instance_lifecycle")
	defer metric.End(err)

	if err = s.updateOperationInstanceLifecycle(nCtx, operInstID, lifecycle); err != nil {
		logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID).Error("failed to update operation instance lifecycle")

		return fmt.Errorf("failed to update operation instance lifecycle, oper-inst-id(%s): %w", operInstID, err)
	}

	return nil
}

// UpdateOperationInstanceExtraExecutionMessages updates operation instance execution messages.
func (s *Storage) UpdateOperationInstanceExtraExecutionMessages(
	nCtx contextx.IContext, operInstID string, messages ...common.Message) error {

	var err error

	// record metric.
	metric := s.metric().Start("update_operation_instance_extra_execution_messages")
	defer metric.End(err)

	if err = s.updateOperationInstanceExtraExecutionMessages(nCtx, operInstID, messages...); err != nil {
		logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID).Error("failed to update operation instance extra execution messages")

		return fmt.Errorf("failed to update operation instance extra execution messages, oper-inst-id(%s): %w",
			operInstID, err)
	}

	return nil
}

// WatchOperInstStopping watches operation instance stopping.
func (s *Storage) WatchOperInstStopping(nCtx contextx.IContext, operInstID string) <-chan struct{} {
	channel := make(chan struct{}, 1)
	subscription := &StopEventSubscription{
		OperInstID: operInstID,
		C:          channel,
	}

	subscriptionID := uuid.New().String()
	s.stopEventSubsMapMutex.Lock()
	s.stopEventSubsMap[subscriptionID] = subscription
	s.stopEventSubsMapMutex.Unlock()

	go func() {
		<-nCtx.Done()

		s.stopEventSubsMapMutex.Lock()
		delete(s.stopEventSubsMap, subscriptionID)
		s.stopEventSubsMapMutex.Unlock()
	}()

	go func() {
		err := s.checkNotifyStopping(nCtx)
		if err != nil {
			logger.G.Sys().WithErr(err).Error("watch operation instance stopping event succeed, but check notify stopping failed")
		}
	}()

	return channel
}

// UpdateActionInstanceContent update action instance content.
func (s *Storage) UpdateActionInstanceContent(
	nCtx contextx.IContext, operInstID string, actionName string, content map[string]any) error {

	var err error

	// record metric.
	metric := s.metric().Start("update_action_instance_content")
	defer metric.End(err)

	if err = s.updateActionInstanceContent(nCtx, operInstID, actionName, content); err != nil {
		logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID, "action-name", actionName).Error("failed to update action instance content")

		return fmt.Errorf("failed to update action instance content, oper-inst-id(%s), action-name(%s): %w",
			operInstID, actionName, err)
	}

	return nil
}

// UpsertActionInstancePrivateData upserts action instance private data.
func (s *Storage) UpsertActionInstancePrivateData(
	nCtx contextx.IContext, operInstID string, actionName string, privateData map[string]any) error {

	var err error

	// record metric.
	metric := s.metric().Start("upsert_action_instance_private_data")
	defer metric.End(err)

	if err = s.upsertActionInstancePrivateData(nCtx, operInstID, actionName, privateData); err != nil {
		logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID, "action-name", actionName).Error("failed to update operation instance private data")

		return fmt.Errorf(
			"failed to update operation instance private data, operation-inst-id(%s), action-name(%s): %w",
			operInstID, actionName, err)
	}

	return nil
}

// DeleteOperationInstances deletes operation instances by given operation instance IDs.
func (s *Storage) DeleteOperationInstances(nCtx contextx.IContext, operInstID ...string) error {
	var err error

	// record metric.
	metric := s.metric().Start("delete_operation_instances")
	defer metric.End(err)

	if err = s.deleteOperationInstances(nCtx, operInstID...); err != nil {
		logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID).Error("failed to delete operation instances")

		return fmt.Errorf("failed to delete operation instances, oper-inst-ids(%v): %w", operInstID, err)
	}

	return nil
}

// DeleteOperationInstancesByTriggerID deletes operation instances by trigger ID.
func (s *Storage) DeleteOperationInstancesByTriggerID(ctx contextx.IContext, triggerID ...string) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("delete_operation_instances_by_trigger_id")
	defer metric.End(err)

	if err = s.deleteOperationInstancesByTriggerID(ctx, triggerID...); err != nil {
		logger.G.Sys().WithErr(err).With("trigger-id", triggerID).Error("failed to delete operation instances by trigger ID")

		return fmt.Errorf("failed to delete operation instances by trigger ID, trigger-ids(%v): %w", triggerID, err)
	}

	return nil
}
