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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operinstdata"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/scheduleworkflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/stopoperinst"
	daoTrigger "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/trigger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/scheduler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/common"
	workoper "github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/schedule"
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
func NewStorage(client *mongo.Client, database string, logger logger.ILogger) (IStorage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &Storage{
		Storage: basestorage.Storage{
			Name:     StorageName,
			Database: client.Database(database),
			Logger:   logger,
		},
	}
	if err := basestorage.InitStorage(&s.Storage,
		basestorage.WithStartFunc(s.initDao),
		basestorage.WithCheckFunc(s.check)); err != nil {
		s.Logger.Errorf("new storage failed, err: %v", err)
		return nil, err
	}

	return s, nil
}

// Storage implements IStorage.
type Storage struct {
	basestorage.Storage

	daoTrigger          daoTrigger.IHandler
	daoOperation        operation.IHandler
	daoScheduleWorkflow scheduleworkflow.IHandler
	daoOperInstData     operinstdata.IHandler
	daoStopOperInst     stopoperinst.Handler

	// stop event subscriptions
	stopEventSubsMap      map[string]*StopEventSubscription
	stopEventSubsMapMutex sync.RWMutex

	stopOperInsts      map[string]struct{}
	stopOperInstsMutex sync.RWMutex

	sg singleflight.Group
}

func (s *Storage) initDao() error {
	s.daoTrigger = daoTrigger.New(s.Database, s.Logger)
	s.daoOperation = operation.New(s.Database, s.Logger)
	s.daoScheduleWorkflow = scheduleworkflow.New(s.Database, s.Logger)
	s.daoOperInstData = operinstdata.New(s.Database, s.Logger)
	s.daoStopOperInst = stopoperinst.New(s.Database, s.Logger)

	s.stopEventSubsMap = make(map[string]*StopEventSubscription)
	s.stopOperInsts = make(map[string]struct{})

	err := s.registerScheduler()
	if err != nil {
		s.Logger.Errorf("failed to register scheduler, err: %v", err)
		return fmt.Errorf("failed to register scheduler, err: %w", err)
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
	s.Scheduler = scheduler.NewScheduler(scheduler.WithLogger(s.Logger))
	err := s.Scheduler.RegisterTask(scheduler.NewTask(
		syncOperationTask,
		taskInterval,
		taskTimeout,
		s.syncStopOperInsts,
	))
	if err != nil {
		s.Logger.Errorf("failed to register sync stopping operation inst task, err: %v", err)
		return fmt.Errorf("failed to register sync stopping operation inst task, err: %w", err)
	}

	go s.daoStopOperInst.WatchInsert(func(stopInstID string) {
		s.stopOperInstsMutex.Lock()
		defer s.stopOperInstsMutex.Unlock()
		s.stopOperInsts[stopInstID] = struct{}{}

		go func() {
			err := s.checkNotifyStopping(contextx.NewContext(s.Ctx, map[string]any{}))
			if err != nil {
				s.Logger.Warnf("failed to notify stopping operation inst, err: %v", err)
			}
		}()
	})

	return nil
}

func (s *Storage) metric() *storage.MetricData {
	return storage.Metric(StorageName)
}

// CreateTrigger creates a new trigger.
func (s *Storage) CreateTrigger(ctx contextx.IContext, trig *trigger.Trigger) error {
	var err error

	// record metric.
	metric := s.metric().Start("create")
	defer metric.End(err)

	if trig == nil {
		return errors.New("trigger is nil")
	}

	if err = s.createTrigger(ctx, trig); err != nil {
		s.Logger.Errorf("failed to create trigger, trigger-id(%s): %v", trig.TriggerID, err)
		return fmt.Errorf("failed to create trigger, trigger-id(%s): %w", trig.TriggerID, err)
	}

	return nil
}

// UpdateTrigger updates a trigger.
func (s *Storage) UpdateTrigger(ctx contextx.IContext, trig *trigger.Trigger) error {
	var err error

	// record metric.
	metric := s.metric().Start("update")
	defer metric.End(err)

	if trig == nil {
		return errors.New("trigger is nil")
	}

	if err = s.updateTrigger(ctx, trig); err != nil {
		s.Logger.Errorf("failed to update trigger, trigger-id(%s): %v", trig.TriggerID, err)
		return fmt.Errorf("failed to update trigger, trigger-id(%s): %w", trig.TriggerID, err)
	}

	return nil
}

// UpdateTriggerState updates a trigger's state.
func (s *Storage) UpdateTriggerState(ctx contextx.IContext, triggerID string, state trigger.State) error {
	var err error

	// record metric.
	metric := s.metric().Start("update_state")
	defer metric.End(err)

	if err = s.updateTriggerState(ctx, triggerID, state); err != nil {
		s.Logger.Errorf("failed to update trigger state, trigger-id(%s): %v", triggerID, err)
		return fmt.Errorf("failed to update trigger state, trigger-id(%s): %w", triggerID, err)
	}

	return nil
}

// GetTrigger gets a trigger by triggerID.
func (s *Storage) GetTrigger(ctx contextx.IContext, triggerID string) (*trigger.Trigger, error) {
	var (
		err  error
		data *trigger.Trigger
	)

	// record metric.
	metric := s.metric().Start("get")
	defer metric.End(err)

	if data, err = s.getTrigger(ctx, triggerID); err != nil {
		s.Logger.Errorf("failed to get trigger, trigger-id(%s): %v", triggerID, err)
		return nil, fmt.Errorf("failed to get trigger, trigger-id(%s): %w", triggerID, err)
	}

	return data, nil
}

// ListAliveTrigger lists alive triggers by category.
func (s *Storage) ListAliveTrigger(ctx contextx.IContext, category trigger.Category) ([]*trigger.Trigger, error) {
	var (
		err     error
		results []*trigger.Trigger
	)

	// record metric.
	metric := s.metric().Start("list_alive")
	defer metric.End(err)

	if results, err = s.listAliveTrigger(ctx, category); err != nil {
		s.Logger.Errorf("failed to list alive triggers, category(%s): %v", category, err)
		return nil, fmt.Errorf("failed to list alive triggers, category(%s): %w", category, err)
	}

	return results, nil
}

// DeleteTriggers deletes triggers by given trigger IDs.
func (s *Storage) DeleteTriggers(ctx contextx.IContext, triggerIDs ...string) error {
	var err error

	// record metric.
	metric := s.metric().Start("delete_many")
	defer metric.End(err)

	if len(triggerIDs) == 0 {
		return errors.New("triggerIDs is empty")
	}

	if err = s.deleteTriggers(ctx, triggerIDs...); err != nil {
		s.Logger.Errorf("failed to delete triggers, trigger-ids(%v): %v", triggerIDs, err)
		return fmt.Errorf("failed to delete triggers, trigger-ids(%v): %w", triggerIDs, err)
	}

	return nil
}

// ListScheduleWorkflow lists schedule workflow by page and conditions.
func (s *Storage) ListScheduleWorkflow(
	ctx contextx.IContext, page types.Page, conditions ...*types.ScheduleWorkflowCondition) (
	[]*schedule.Schedule, int64, error) {

	var (
		results []*schedule.Schedule
		num     int64
		err     error
	)

	// record metric.
	metric := s.metric().Start("list")
	defer metric.End(err)

	if results, num, err = s.listScheduleWorkflow(ctx, page, conditions...); err != nil {
		s.Logger.Errorf("failed to list schedule workflows: %v", err)
		return nil, 0, fmt.Errorf("failed to list schedule workflows: %w", err)
	}

	return results, num, nil
}

// CountScheduleWorkflow counts schedule workflow by conditions.
func (s *Storage) CountScheduleWorkflow(
	ctx contextx.IContext, conditions ...*types.ScheduleWorkflowCondition) (
	int64, error) {

	var (
		num int64
		err error
	)

	// record metric.
	metric := s.metric().Start("count")
	defer metric.End(err)

	if num, err = s.countScheduleWorkflow(ctx, conditions...); err != nil {
		s.Logger.Errorf("failed to count schedule workflows: %v", err)
		return 0, fmt.Errorf("failed to count schedule workflows: %w", err)
	}

	return num, nil
}

// GetScheduleWorkflow gets a schedule workflow by workflow-id.
func (s *Storage) GetScheduleWorkflow(ctx contextx.IContext, workflowID string) (*schedule.Schedule, error) {
	var (
		workflow *schedule.Schedule
		err      error
	)

	// record metric.
	metric := s.metric().Start("get")
	defer metric.End(err)

	if ctx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if workflowID == "" {
		return nil, errors.New("workflow id cannot be empty")
	}

	if workflow, err = s.getScheduleWorkflow(ctx, workflowID); err != nil {
		s.Logger.Errorf("failed to get schedule workflow, workflow-id(%s): %v", workflowID, err)
		return nil, fmt.Errorf("failed to get workflow, workflow-id(%s): %w", workflowID, err)
	}

	return workflow, nil
}

// CreateScheduleWorkflow creates a new schedule workflow.
func (s *Storage) CreateScheduleWorkflow(ctx contextx.IContext, workflow *schedule.Schedule) error {
	var err error

	// record metric.
	metric := s.metric().Start("create")
	defer metric.End(err)

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if workflow == nil {
		return errors.New("schedule workflow cannot be nil")
	}

	if workflow.WorkflowID == "" {
		return errors.New("schedule workflow id cannot be empty")
	}

	if err = s.createScheduleWorkflow(ctx, workflow); err != nil {
		s.Logger.Errorf("failed to create schedule workflow, workflow-id(%s): %v", workflow.WorkflowID, err)
		return fmt.Errorf("failed to create schedule workflow, workflow-id(%s): %w", workflow.WorkflowID, err)
	}

	return nil
}

// GetOperation get operation by operationID.
func (s *Storage) GetOperation(ctx contextx.IContext, operationID string) (*workoper.Operation, error) {
	var (
		opers []*workoper.Operation
		err   error
	)
	// record metric.
	metric := s.metric().Start("get")
	defer metric.End(err)

	if ctx == nil {
		return nil, basestorage.ErrNilContent()
	}

	cond := &types.OperationCondition{
		ExactInclude: &types.OperationExactFields{
			OperationID: []string{operationID},
		},
	}
	if opers, _, err = s.listOperation(ctx, types.SingleItemPage(), cond); err != nil {
		s.Logger.Errorf("failed to get operations, operationID(%s): %v", operationID, err)
		return nil, fmt.Errorf("failed to get operations, operationID(%s): %w", operationID, err)
	}

	if len(opers) != 1 {
		return nil, fmt.Errorf("get operation failed, match num not 1, operation-id(%s), count(%d)",
			operationID, len(opers))
	}

	return opers[0], nil
}

// UpsertOperation upsert operation.
func (s *Storage) UpsertOperation(ctx contextx.IContext, operation *workoper.Operation) error {
	var err error

	// record metric.
	metric := s.metric().Start("upsert")
	defer metric.End(err)

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if operation == nil {
		return basestorage.ErrUpsertNilData()
	}

	if err = s.upsertOperation(ctx, operation); err != nil {
		s.Logger.Errorf("failed to upsert operation, operation-id(%s): %v", operation.OperationID, err)
		return fmt.Errorf("failed to upsert operation, operation-id(%s): %w", operation.OperationID, err)
	}

	return nil
}

// ListOperationByTriggerID lists operation by trigger id.
func (s *Storage) ListOperationByTriggerID(
	ctx contextx.IContext, page types.Page, triggerID ...string) (
	[]*workoper.Operation, int64, error) {

	var (
		opers []*workoper.Operation
		num   int64
		err   error
	)

	// record metric.
	metric := s.metric().Start("list_by_trigger_id")
	defer metric.End(err)

	if ctx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	if err = page.Validate(); err != nil {
		return nil, 0, err
	}

	if len(triggerID) == 0 {
		return nil, 0, basestorage.ErrEmptyTriggerID()
	}

	cond := &types.OperationCondition{
		ExactInclude: &types.OperationExactFields{
			TriggerID: triggerID,
		},
	}
	if opers, num, err = s.listOperation(ctx, page, cond); err != nil {
		s.Logger.Errorf("failed to list operations by trigger id, trigger-ids(%v): %v", triggerID, err)
		return nil, 0, fmt.Errorf("failed to list operations by trigger id, trigger-ids(%v): %w", triggerID, err)
	}

	return opers, num, nil
}

// ListOperationByOperationID lists operation by operation id.
func (s *Storage) ListOperationByOperationID(
	ctx contextx.IContext, operationID ...string) (
	[]*workoper.Operation, int64, error) {

	var (
		opers []*workoper.Operation
		num   int64
		err   error
	)

	// record metric.
	metric := s.metric().Start("list_by_operation_id")
	defer metric.End(err)

	if ctx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	if len(operationID) == 0 {
		return nil, 0, basestorage.ErrEmptyOperationID()
	}

	cond := &types.OperationCondition{
		ExactInclude: &types.OperationExactFields{
			OperationID: operationID,
		},
	}
	if opers, num, err = s.listOperation(ctx, types.UnlimitedPage(), cond); err != nil {
		s.Logger.Errorf("failed to list operations by operation id, operation-ids(%v): %v", operationID, err)
		return nil, 0, fmt.Errorf("failed to list operations by operation id, operation-ids(%v): %w", operationID, err)
	}

	return opers, num, nil
}

// ListOperationByNodeWorkflowOperationCondition lists operation by node workflow operation condition.
func (s *Storage) ListOperationByNodeWorkflowOperationCondition(
	ctx contextx.IContext, page types.Page, condition ...*types.NodeWorkflowOperationCondition) (
	[]*workoper.Operation, int64, error) {

	var (
		opers []*workoper.Operation
		num   int64
		err   error
	)

	// record metric.
	metric := s.metric().Start("list_by_node_workflow_operation_condition")
	defer metric.End(err)

	if opers, num, err = s.listOperationByNodeWorkflowOperationCondition(ctx, page, condition...); err != nil {
		s.Logger.Errorf("failed to list operations by node workflow operation condition: %v", err)
		return nil, 0, fmt.Errorf("failed to list operations by node workflow operation condition: %w", err)
	}

	return opers, num, nil
}

// ListEmptyOperation lists empty operation by trigger id.
func (s *Storage) ListEmptyOperation(
	ctx contextx.IContext, page types.Page, triggerID string) (
	[]*workoper.Operation, int64, error) {

	var (
		opers []*workoper.Operation
		num   int64
		err   error
	)

	// record metric.
	metric := s.metric().Start("list_empty")
	defer metric.End(err)

	if ctx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	if err := page.Validate(); err != nil {
		return nil, 0, err
	}

	if triggerID == "" {
		return nil, 0, basestorage.ErrEmptyTriggerID()
	}

	cond := &types.OperationCondition{
		ExactInclude: &types.OperationExactFields{
			TriggerID:     []string{triggerID},
			OperInstEmpty: true,
		},
	}
	if opers, num, err = s.listOperation(ctx, page, cond); err != nil {
		s.Logger.Errorf("failed to list empty operations, trigger-id(%s): %v", triggerID, err)
		return nil, 0, fmt.Errorf("failed to list empty operations, trigger-id(%s): %w", triggerID, err)
	}

	return opers, num, nil
}

// DeleteOperations deletes operations by operation ids.
func (s *Storage) DeleteOperations(ctx contextx.IContext, operationID ...string) error {
	var err error

	// record metric.
	metric := s.metric().Start("delete")
	defer metric.End(err)

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if len(operationID) == 0 {
		return basestorage.ErrEmptyOperationID()
	}

	if err = s.deleteOperations(ctx, operationID...); err != nil {
		s.Logger.Errorf("failed to delete operations, operation-ids(%v): %v", operationID, err)
		return fmt.Errorf("failed to delete operations, operation-ids(%v): %w", operationID, err)
	}

	return nil
}

// PullOperationInstanceIDsFromOperation pulls operation instance IDs from operation.
func (s *Storage) PullOperationInstanceIDsFromOperation(
	ctx contextx.IContext, operationID string, operInstIDs ...string) error {

	var err error

	// record metric.
	metric := s.metric().Start("pull_instance_ids_from_operation")
	defer metric.End(err)

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if len(operationID) == 0 {
		return basestorage.ErrEmptyOperationID()
	}

	if err = s.pullOperationInstanceIDs(ctx, operationID, operInstIDs...); err != nil {
		s.Logger.Errorf("failed to pull operation instance ids, operation-id(%s), oper-inst-ids(%v): %v",
			operationID, operInstIDs, err)

		return fmt.Errorf("failed to pull operation instance ids, operation-id(%s), oper-inst-ids(%v): %w",
			operationID, operInstIDs, err)
	}

	return nil
}

// UpdateOperInstActionStatus update the oper inst action status.
func (s *Storage) UpdateOperInstActionStatus(
	ctx contextx.IContext, operInstID string, actionName string, status action.State) error {

	var err error

	// record metric.
	metric := s.metric().Start("update_oper_inst_action_status")
	defer metric.End(err)

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if operInstID == "" {
		return errors.New("oper inst id is empty")
	}

	if actionName == "" {
		return errors.New("action name is empty")
	}

	if err = status.Validate(); err != nil {
		return err
	}

	if err = s.updateOperInstActionStatus(ctx, operInstID, actionName, status); err != nil {
		s.Logger.Errorf("failed to update oper inst action status, oper-inst-id(%s), action-name(%s): %v",
			operInstID, actionName, err)

		return fmt.Errorf("failed to update oper inst action status, oper-inst-id(%s), action-name(%s): %w",
			operInstID, actionName, err)
	}

	return nil
}

// GetActionInstanceData gets full action instance data.
func (s *Storage) GetActionInstanceData(
	ctx contextx.IContext, operInstID, actionName string) (*action.InstanceData, error) {

	var (
		data *action.InstanceData
		err  error
	)

	// record metric.
	metric := s.metric().Start("get_action")
	defer metric.End(err)

	if ctx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if operInstID == "" {
		return nil, basestorage.ErrEmptyOperaInstID()
	}

	if actionName == "" {
		return nil, basestorage.ErrEmptyActionName()
	}

	if data, err = s.getActionInstanceData(ctx, operInstID, actionName); err != nil {
		s.Logger.Errorf("failed to get action instance data, oper-inst-id(%s), action-name(%s): %v",
			operInstID, actionName, err)

		return nil, fmt.Errorf("failed to get action instance data, oper-inst-id(%s), action-name(%s): %w",
			operInstID, actionName, err)
	}

	return data, nil
}

// GetActionInstanceLifecycle gets action instance lifecycle.
func (s *Storage) GetActionInstanceLifecycle(
	ctx contextx.IContext, operInstID, actionName string) (*action.Lifecycle, error) {

	var (
		lifecycle *action.Lifecycle
		err       error
	)

	// record metric.
	metric := s.metric().Start("get_action_lifecycle")
	defer metric.End(err)

	if ctx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if operInstID == "" {
		return nil, errors.New("operation instance id is empty")
	}

	if actionName == "" {
		return nil, errors.New("action name is empty")
	}

	if lifecycle, err = s.getActionInstanceLifecycle(ctx, operInstID, actionName); err != nil {
		s.Logger.Errorf("failed to get action instance lifecycle, oper-inst-id(%s), action-name(%s): %v",
			operInstID, actionName, err)

		return nil, fmt.Errorf("failed to get action instance lifecycle, oper-inst-id(%s), action-name(%s): %w",
			operInstID, actionName, err)
	}

	return lifecycle, nil
}

// GetActionInstancePrivateData gets action instance private data.
func (s *Storage) GetActionInstancePrivateData(
	ctx contextx.IContext, operInstID, actionName string) (map[string]any, error) {

	var (
		data map[string]any
		err  error
	)

	// record metric.
	metric := s.metric().Start("get_action_private_data")
	defer metric.End(err)

	if ctx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if operInstID == "" {
		return nil, errors.New("operation instance id is empty")
	}

	if actionName == "" {
		return nil, errors.New("action name is empty")
	}

	if data, err = s.getActionInstancePrivateData(ctx, operInstID, actionName); err != nil {
		s.Logger.Errorf("failed to get action instance private data, oper-inst-id(%s), action-name(%s): %v",
			operInstID, actionName, err)

		return nil, fmt.Errorf("failed to get action instance private data, oper-inst-id(%s), action-name(%s): %w",
			operInstID, actionName, err)
	}

	return data, nil
}

// UpdateActionInstanceLifecycle updates action instance lifecycle.
func (s *Storage) UpdateActionInstanceLifecycle(
	ctx contextx.IContext, operInstID, actionName string, lifecycle *action.Lifecycle) error {

	var err error

	// record metric.
	metric := s.metric().Start("update_action_lifecycle")
	defer metric.End(err)

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if operInstID == "" {
		return errors.New("operation instance id is empty")
	}

	if actionName == "" {
		return errors.New("action name is empty")
	}

	if lifecycle == nil {
		return errors.New("lifecycle is nil")
	}

	if err = s.updateActionInstanceLifecycle(ctx, operInstID, actionName, lifecycle); err != nil {
		s.Logger.Errorf("failed to update action instance lifecycle, oper-inst-id(%s), action-name(%s): %v",
			operInstID, actionName, err)

		return fmt.Errorf("failed to update action instance lifecycle, oper-inst-id(%s), action-name(%s): %w",
			operInstID, actionName, err)
	}

	return nil
}

// PushActionInstanceMessage pushes action instance message.
func (s *Storage) PushActionInstanceMessage(
	ctx contextx.IContext, operInstID, actionName string, messages ...common.Message) error {

	var err error

	// record metric.
	metric := s.metric().Start("push_action_message")
	defer metric.End(err)

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if operInstID == "" {
		return errors.New("operation instance id is empty")
	}

	if actionName == "" {
		return errors.New("action name is empty")
	}

	if len(messages) == 0 {
		return nil
	}

	if err = s.pushActionInstanceMessage(ctx, operInstID, actionName, messages...); err != nil {
		s.Logger.Errorf("failed to push action instance message, oper-inst-id(%s), action-name(%s): %v",
			operInstID, actionName, err)

		return fmt.Errorf("failed to push action instance message, oper-inst-id(%s), action-name(%s): %w",
			operInstID, actionName, err)
	}

	return nil
}

// GetOperationInstanceFullData gets full operation instance data.
func (s *Storage) GetOperationInstanceFullData(
	ctx contextx.IContext, operInstID string) (*workoper.InstanceData, error) {

	var (
		data *workoper.InstanceData
		err  error
	)

	// record metric.
	metric := s.metric().Start("get_oper_inst_full_data")
	defer metric.End(err)

	if ctx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if operInstID == "" {
		return nil, errors.New("operation inst id is empty")
	}

	cond := &types.OperInstDataCondition{
		ExactInclude: &types.OperInstDataExactFields{
			OperInstID: []string{operInstID},
		},
	}
	if data, err = s.getOperationInstanceData(ctx, cond); err != nil {
		s.Logger.Errorf("failed to get operation inst data, oper-inst-id(%s): %v", operInstID, err)
		return nil, fmt.Errorf("failed to get operation inst data, oper-inst-id(%s): %w", operInstID, err)
	}

	return data, nil
}

// GetOperationInstanceBriefData gets brief operation instance data.
func (s *Storage) GetOperationInstanceBriefData(
	ctx contextx.IContext, operInstID string) (*workoper.InstanceBriefData, error) {

	var (
		data *workoper.InstanceData
		err  error
	)

	// record metric.
	metric := s.metric().Start("get_oper_inst_brief_data")
	defer metric.End(err)

	if ctx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if operInstID == "" {
		return nil, errors.New("operation inst id is empty")
	}

	cond := &types.OperInstDataCondition{
		ExactInclude: &types.OperInstDataExactFields{
			OperInstID: []string{operInstID},
		},
	}
	if data, err = s.getOperationInstanceData(ctx, cond); err != nil {
		s.Logger.Errorf("failed to get operation inst data, oper-inst-id(%s): %v", operInstID, err)
		return nil, fmt.Errorf("failed to get operation inst data, oper-inst-id(%s): %w", operInstID, err)
	}

	return &data.InstanceBriefData, nil
}

// ListOperationInstanceBriefData lists operation instance brief data. without action instance data.
func (s *Storage) ListOperationInstanceBriefData(
	ctx contextx.IContext, page types.Page, conditions ...*types.OperInstDataCondition) (
	[]*workoper.InstanceBriefData, int64, error) {

	var (
		results []*workoper.InstanceBriefData
		num     int64
		err     error
	)

	// record metric.
	metric := s.metric().Start("list_oper_inst_brief_data_by_condition")
	defer metric.End(err)

	if ctx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	if results, num, err = s.listOperationInstanceBriefDataWithoutActionInst(ctx, page, conditions...); err != nil {
		s.Logger.Errorf("failed to list operation instance brief data: %v", err)
		return nil, 0, fmt.Errorf("failed to list operation instance brief data: %w", err)
	}

	return results, num, nil
}

// CountOperationInstance counts operation instance.
func (s *Storage) CountOperationInstance(
	ctx contextx.IContext, triggerID string, states ...workoper.State) (int64, error) {

	var (
		num int64
		err error
	)

	// record metric.
	metric := s.metric().Start("count_oper_inst")
	defer metric.End(err)

	if ctx == nil {
		return 0, basestorage.ErrNilContent()
	}

	if triggerID == "" {
		return 0, errors.New("trigger id is empty")
	}

	cond := &types.OperInstDataCondition{
		ExactInclude: &types.OperInstDataExactFields{
			TriggerID: []string{triggerID},
			State:     states,
		},
	}
	if num, err = s.countOperationInstance(ctx, cond); err != nil {
		return 0, fmt.Errorf("failed to count operation instance: %v", err)
	}

	return num, nil
}

// ListOperInstanceBriefByOperationID lists operation instance brief data.
func (s *Storage) ListOperInstanceBriefByOperationID(ctx contextx.IContext, page types.Page, operationID ...string) (
	[]*workoper.InstanceBriefData, int64, error) {

	var (
		results []*workoper.InstanceBriefData
		num     int64
		err     error
	)

	// record metric.
	metric := s.metric().Start("list_oper_inst_brief_data_by_operation")
	defer metric.End(err)

	if ctx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	if err = page.Validate(); err != nil {
		return nil, 0, err
	}

	cond := &types.OperInstDataCondition{
		ExactInclude: &types.OperInstDataExactFields{
			OperationID: operationID,
		},
	}
	if results, num, err = s.listOperationInstanceBriefDataWithoutActionInst(ctx, page, cond); err != nil {
		s.Logger.Errorf("failed to list operation instance brief data by operation: %v", err)
		return nil, 0, fmt.Errorf("failed to list operation instance brief data by operation: %w", err)
	}

	return results, num, nil
}

// UpsertOperationInstanceData upserts operation instance data.
func (s *Storage) UpsertOperationInstanceData(
	ctx contextx.IContext, operInstData *workoper.InstanceData) error {

	var err error

	// record metric.
	metric := s.metric().Start("upsert_oper_inst_data")
	defer metric.End(err)

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if operInstData == nil {
		return basestorage.ErrUpsertNilData()
	}

	if err = s.upsertOperationInstanceData(ctx, operInstData); err != nil {
		s.Logger.Errorf("failed to update operation inst data, operation-inst(%v): %v", operInstData, err)
		return fmt.Errorf("failed to update operation inst data, operation-inst(%v): %w", operInstData, err)
	}

	return nil
}

// UpdateOperationInstanceLifecycle updates operation instance lifecycle.
func (s *Storage) UpdateOperationInstanceLifecycle(
	ctx contextx.IContext, operInstID string, lifecycle *workoper.Lifecycle) error {

	var err error

	// record metric.
	metric := s.metric().Start("update_oper_inst_lifecycle")
	defer metric.End(err)

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if lifecycle == nil {
		return errors.New("lifecycle is nil")
	}

	if err = s.updateOperationInstanceLifecycle(ctx, operInstID, lifecycle); err != nil {
		s.Logger.Errorf("failed to update operation inst lifecycle, oper-inst-id(%s): %v", operInstID, err)
		return fmt.Errorf("failed to update operation inst lifecycle, oper-inst-id(%s): %w", operInstID, err)
	}

	return nil
}

// UpdateOperationInstanceExtraExecutionMessages updates operation instance execution messages.
func (s *Storage) UpdateOperationInstanceExtraExecutionMessages(
	ctx contextx.IContext, operInstID string, messages ...common.Message) error {

	var err error

	// record metric.
	metric := s.metric().Start("update_oper_inst_extra_execution_messages")
	defer metric.End(err)

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if operInstID == "" {
		return basestorage.ErrEmptyOperaInstID()
	}

	if len(messages) == 0 {
		return nil
	}

	if err = s.updateOperationInstanceExtraExecutionMessages(ctx, operInstID, messages...); err != nil {
		s.Logger.Errorf("failed to update operation inst extra execution messages, oper-inst-id(%s): %v",
			operInstID, err)

		return fmt.Errorf("failed to update operation inst extra execution messages, oper-inst-id(%s): %w",
			operInstID, err)
	}

	return nil
}

// WatchOperInstStopping watches operation instance stopping.
func (s *Storage) WatchOperInstStopping(ctx contextx.IContext, operInstID string) <-chan struct{} {
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
		<-ctx.Done()

		s.stopEventSubsMapMutex.Lock()
		delete(s.stopEventSubsMap, subscriptionID)
		s.stopEventSubsMapMutex.Unlock()
	}()

	go func() {
		err := s.checkNotifyStopping(ctx)
		if err != nil {
			s.Logger.Errorf("watch operation instance stopping event succeed, "+
				"but check notify stopping failed, err: %v", err)
		}
	}()

	return channel
}

// UpdateActionInstanceContent update action instance content.
func (s *Storage) UpdateActionInstanceContent(
	ctx contextx.IContext, operInstID string, actionName string, content map[string]any) error {

	var err error

	// record metric.
	metric := s.metric().Start("update_action_content")
	defer metric.End(err)

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if operInstID == "" {
		return errors.New("operation instance id is empty")
	}

	if actionName == "" {
		return errors.New("action name is empty")
	}

	if len(content) == 0 {
		return errors.New("content is empty")
	}

	if err = s.updateActionInstanceContent(ctx, operInstID, actionName, content); err != nil {
		s.Logger.Errorf("failed to update action instance content, oper-inst-id(%s), action-name(%s): %v",
			operInstID, actionName, err)

		return fmt.Errorf("failed to update action instance content, oper-inst-id(%s), action-name(%s): %w",
			operInstID, actionName, err)
	}

	return nil
}

// UpsertActionInstancePrivateData upserts action instance private data.
func (s *Storage) UpsertActionInstancePrivateData(
	ctx contextx.IContext, operInstID string, actionName string, privateData map[string]any) error {

	var err error

	// record metric.
	metric := s.metric().Start("upsert_action_private_data")
	defer metric.End(err)

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if operInstID == "" {
		return errors.New("operation instance id is empty")
	}

	if actionName == "" {
		return errors.New("action name is empty")
	}

	if len(privateData) == 0 {
		return nil
	}

	if err = s.upsertActionInstancePrivateData(ctx, operInstID, actionName, privateData); err != nil {
		s.Logger.Errorf("failed to update operation instance private data, operation-inst-id(%s), action-name(%s): %v",
			operInstID, actionName, err)

		return fmt.Errorf(
			"failed to update operation instance private data, operation-inst-id(%s), action-name(%s), err: %w",
			operInstID, actionName, err)
	}

	return nil
}

// DeleteOperationInstances deletes operation instances by given operation instance IDs.
func (s *Storage) DeleteOperationInstances(ctx contextx.IContext, operInstID ...string) error {
	var err error

	// record metric.
	metric := s.metric().Start("delete_oper_inst")
	defer metric.End(err)

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if len(operInstID) == 0 {
		return nil
	}

	if err = s.deleteOperationInstances(ctx, operInstID...); err != nil {
		s.Logger.Errorf("failed to delete operation instances, oper-inst-ids(%v): %v", operInstID, err)
		return fmt.Errorf("failed to delete operation instances, oper-inst-ids(%v): %w", operInstID, err)
	}

	return nil
}
