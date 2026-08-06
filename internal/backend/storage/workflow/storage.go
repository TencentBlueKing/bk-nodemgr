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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operinstdata"
	packageworkflow "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/package-workflow"
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

	metricOperationCreateTrigger                                   = "create_trigger"
	metricOperationUpdateTrigger                                   = "update_trigger"
	metricOperationSwitchTriggerAlive                              = "switch_trigger_alive"
	metricOperationGetTrigger                                      = "get_trigger"
	metricOperationListAliveTrigger                                = "list_alive_trigger"
	metricOperationListTrigger                                     = "list_trigger"
	metricOperationDeleteTriggers                                  = "delete_triggers"
	metricOperationExistTrigger                                    = "exist_trigger"
	metricOperationListScheduledWorkflow                           = "list_scheduled_workflow"
	metricOperationCountScheduledWorkflow                          = "count_scheduled_workflow"
	metricOperationGetScheduledWorkflow                            = "get_scheduled_workflow"
	metricOperationCreateScheduledWorkflow                         = "create_scheduled_workflow"
	metricOperationUpdateScheduledWorkflowTriggerID                = "update_scheduled_workflow_trigger_id"
	metricOperationUpdateScheduledWorkflowPrivateData              = "update_scheduled_workflow_private_data"
	metricOperationSwitchScheduledWorkflow                         = "switch_scheduled_workflow"
	metricOperationGetOperation                                    = "get_operation"
	metricOperationUpsertOperation                                 = "upsert_operation"
	metricOperationListOperation                                   = "list_operation"
	metricOperationCountOperation                                  = "count_operation"
	metricOperationDistinctOperation                               = "distinct_operation"
	metricOperationListOperationByOperationID                      = "list_operation_by_operation_id"
	metricOperationListOperationIDByParentOperationID              = "list_operation_id_by_parent_operation_id"
	metricOperationListOperationIDByParentOperInstID               = "list_operation_id_by_parent_oper_inst_id"
	metricOperationListNeedInstantiateOperationByTriggerID         = "list_need_instantiate_operation_by_trigger_id"
	metricOperationExistNeedInstantiateOperationByTriggerID        = "exist_need_instantiate_operation_by_trigger_id"
	metricOperationDeleteOperationsByTriggerID                     = "delete_operations_by_trigger_id"
	metricOperationPullOperationInstanceIDsFromOperation           = "pull_operation_instance_ids_from_operation"
	metricOperationUpdateOperationLatestInstBriefData              = "update_operation_latest_inst_brief_data"
	metricOperationUpdateOperInstActionStatus                      = "update_oper_inst_action_status"
	metricOperationGetActionInstanceData                           = "get_action_instance_data"
	metricOperationGetActionInstanceLifecycle                      = "get_action_instance_lifecycle"
	metricOperationGetActionInstancePrivateData                    = "get_action_instance_private_data"
	metricOperationUpdateActionInstanceLifecycle                   = "update_action_instance_lifecycle"
	metricOperationPushActionInstanceMessage                       = "push_action_instance_message"
	metricOperationGetOperationInstanceFullData                    = "get_operation_instance_full_data"
	metricOperationGetOperationInstanceBriefData                   = "get_operation_instance_brief_data"
	metricOperationListOperationInstanceBriefDataWithoutActionInst = "list_operation_instance_brief_data_without_action_inst_by_condition"
	metricOperationCountOperationInstance                          = "count_operation_instance"
	metricOperationExistOperationInstance                          = "exist_operation_instance"
	metricOperationListOperationInstanceBriefByOperationID         = "list_operation_instance_brief_without_action_inst_by_operation_id"
	metricOperationListOperationInstanceBriefByTriggerID           = "list_operation_instance_brief_without_action_inst_by_trigger_id"
	metricOperationGetLatestOperationInstanceStatusDistribution    = "get_latest_operation_instance_status_distribution_by_trigger_id"
	metricOperationUpsertOperationInstanceData                     = "upsert_operation_instance_data"
	metricOperationUpdateOperationInstanceLifecycle                = "update_operation_instance_lifecycle"
	metricOperationUpdateOperationLatestActionInstBriefData        = "update_operation_latest_action_inst_brief_data"
	metricOperationUpdateOperationInstanceExtraExecutionMessages   = "update_operation_instance_extra_execution_messages"
	metricOperationUpsertOperInstStop                              = "upsert_oper_inst_stop"
	metricOperationUpdateActionInstanceContent                     = "update_action_instance_content"
	metricOperationUpsertActionInstancePrivateData                 = "upsert_action_instance_private_data"
	metricOperationDeleteOperationInstances                        = "delete_operation_instances"
	metricOperationDeleteOperationInstancesByTriggerID             = "delete_operation_instances_by_trigger_id"
	metricOperationCreatePackageWorkflow                           = "create_package_workflow"
)

// NewStorage creates a new workflow storage.
func NewStorage(client *mongo.Client, database string) (IStorage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &Storage{
		Storage: basestorage.Storage{
			Name:      StorageName,
			Database:  client.Database(database),
			Scheduler: scheduler.NewScheduler(),
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
	daoOperInstData      operinstdata.IHandler
	daoStopOperInst      stopoperinst.IHandler
	daoScheduledWorkflow scheduledworkflow.IHandler
	daoPackageWorkflow   packageworkflow.IHandler

	// stop event subscriptions
	stopEventSubsMap      map[string]*StopEventSubscription
	stopEventSubsMapMutex sync.RWMutex

	stopOperInsts      map[string]struct{}
	stopOperInstsMutex sync.RWMutex

	sg singleflight.Group

	monitoredPackageWorkflows      map[string]*types.PackageWorkflow
	monitoredPackageWorkflowsMutex sync.RWMutex
}

func (s *Storage) initDao() error {
	s.daoTrigger = daoTrigger.New(s.Database)
	s.daoOperation = operation.New(s.Database)
	s.daoScheduledWorkflow = scheduledworkflow.New(s.Database)
	s.daoOperInstData = operinstdata.New(s.Database)
	s.daoStopOperInst = stopoperinst.New(s.Database)
	s.daoPackageWorkflow = packageworkflow.New(s.Database)

	s.stopEventSubsMap = make(map[string]*StopEventSubscription)
	s.stopOperInsts = make(map[string]struct{})

	s.monitoredPackageWorkflows = make(map[string]*types.PackageWorkflow)

	if err := s.registerStopOperInstTask(); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to register scheduler")

		return fmt.Errorf("failed to register scheduler: %w", err)
	}

	if err := s.registerPackageWorkflowScheduler(); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to register package workflow scheduler")

		return fmt.Errorf("failed to register package workflow scheduler: %w", err)
	}

	return nil
}

func (s *Storage) check() error {
	if s.daoTrigger == nil {
		return errors.New("trigger dao is nil")
	}

	if s.daoScheduledWorkflow == nil {
		return errors.New("dao scheduled workflow is nil")
	}

	if s.daoPackageWorkflow == nil {
		return errors.New("dao package workflow is nil")
	}

	return nil
}

// CreateTrigger creates a new trigger.
func (s *Storage) CreateTrigger(nCtx contextx.IContext, trig *trigger.Trigger) error {
	return s.WrapFn(nCtx, metricOperationCreateTrigger, func(nCtx contextx.IContext) error {
		var err error
		if err = s.createTrigger(nCtx, trig); err != nil {
			logger.G.Sys().WithErr(err).With("trigger-id", trig.TriggerID).Error("failed to create trigger")

			return fmt.Errorf("failed to create trigger, trigger-id(%s): %w", trig.TriggerID, err)
		}

		return nil
	})
}

// UpdateTrigger updates a trigger.
func (s *Storage) UpdateTrigger(nCtx contextx.IContext, trig *trigger.Trigger) error {
	return s.WrapFn(nCtx, metricOperationUpdateTrigger, func(nCtx contextx.IContext) error {
		var err error
		if err = s.updateTrigger(nCtx, trig); err != nil {
			logger.G.Sys().WithErr(err).With("trigger-id", trig.TriggerID).Error("failed to update trigger")

			return fmt.Errorf("failed to update trigger, trigger-id(%s): %w", trig.TriggerID, err)
		}

		return nil
	})
}

// SwitchTriggerActive switches a trigger active status.
func (s *Storage) SwitchTriggerActive(nCtx contextx.IContext, triggerID string, active bool) error {
	return s.WrapFn(nCtx, metricOperationSwitchTriggerAlive, func(nCtx contextx.IContext) error {
		var err error
		if err = s.switchTriggerActive(nCtx, triggerID, active); err != nil {
			logger.G.Sys().WithErr(err).With("trigger-id", triggerID).Error("failed to switch trigger active state")

			return fmt.Errorf("failed to switch trigger active state, trigger-id(%s): %w", triggerID, err)
		}

		return nil
	})
}

// GetTrigger gets a trigger by triggerID.
func (s *Storage) GetTrigger(nCtx contextx.IContext, triggerID string) (*trigger.Trigger, error) {
	var (
		err  error
		data *trigger.Trigger
	)

	err = s.WrapFn(nCtx, metricOperationGetTrigger, func(nCtx contextx.IContext) error {
		var err error
		if data, err = s.getTrigger(nCtx, triggerID); err != nil {
			logger.G.Sys().WithErr(err).With("trigger-id", triggerID).Error("failed to get trigger")

			return fmt.Errorf("failed to get trigger, trigger-id(%s): %w", triggerID, err)
		}

		return nil
	})

	return data, err
}

// ListActiveTrigger lists active triggers by given category.
func (s *Storage) ListActiveTrigger(nCtx contextx.IContext, page types.Page, category trigger.Category) ([]*trigger.Trigger, error) {
	var (
		err     error
		results []*trigger.Trigger
	)

	err = s.WrapFn(nCtx, metricOperationListAliveTrigger, func(nCtx contextx.IContext) error {
		var err error
		if results, err = s.listActiveTrigger(nCtx, page, category); err != nil {
			logger.G.Sys().WithErr(err).With("category", category).Error("failed to list active triggers")

			return fmt.Errorf("failed to list active triggers, category(%s): %w", category, err)
		}

		return nil
	})

	return results, err
}

// ListTrigger lists triggers by given category.
func (s *Storage) ListTrigger(nCtx contextx.IContext, page types.Page, category trigger.Category) ([]*trigger.Trigger, int64, error) {
	var (
		results []*trigger.Trigger
		num     int64
		err     error
	)

	err = s.WrapFn(nCtx, metricOperationListTrigger, func(nCtx contextx.IContext) error {
		var err error
		if results, num, err = s.listTrigger(nCtx, page, category); err != nil {
			logger.G.Sys().WithErr(err).With("category", category).Error("failed to list triggers")

			return fmt.Errorf("failed to list triggers, category(%s): %w", category, err)
		}

		return nil
	})

	return results, num, err
}

// DeleteTriggers deletes triggers by given trigger IDs.
func (s *Storage) DeleteTriggers(nCtx contextx.IContext, triggerIDs ...string) error {
	return s.WrapFn(nCtx, metricOperationDeleteTriggers, func(nCtx contextx.IContext) error {
		var err error
		if err = s.deleteTriggers(nCtx, triggerIDs...); err != nil {
			logger.G.Sys().WithErr(err).With("trigger-ids", triggerIDs).Error("failed to delete triggers")

			return fmt.Errorf("failed to delete triggers, trigger-ids(%v): %w", triggerIDs, err)
		}

		return nil
	})
}

// ExistTrigger checks if a trigger exists.
func (s *Storage) ExistTrigger(nCtx contextx.IContext, triggerID string) (bool, error) {
	var (
		exist bool
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationExistTrigger, func(nCtx contextx.IContext) error {
		var err error
		if exist, err = s.existTrigger(nCtx, triggerID); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to check trigger exist")

			return fmt.Errorf("failed to check trigger exist: %w", err)
		}

		return nil
	})

	return exist, err
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

	err = s.WrapFn(nCtx, metricOperationListScheduledWorkflow, func(nCtx contextx.IContext) error {
		var err error
		if results, num, err = s.listScheduledWorkflow(nCtx, page, conditions...); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to list scheduled workflows")

			return fmt.Errorf("failed to list scheduled workflows: %w", err)
		}

		return nil
	})

	return results, num, err
}

// CountScheduledWorkflow counts scheduled workflow by conditions.
func (s *Storage) CountScheduledWorkflow(
	nCtx contextx.IContext, conditions ...*types.ScheduledWorkflowCondition) (
	int64, error) {

	var (
		num int64
		err error
	)

	err = s.WrapFn(nCtx, metricOperationCountScheduledWorkflow, func(nCtx contextx.IContext) error {
		var err error
		if num, err = s.countScheduledWorkflow(nCtx, conditions...); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to count scheduled workflows")

			return fmt.Errorf("failed to count scheduled workflows: %w", err)
		}

		return nil
	})

	return num, err
}

// GetScheduledWorkflow gets a scheduled workflow by workflow-id.
func (s *Storage) GetScheduledWorkflow(nCtx contextx.IContext, workflowID string) (*types.ScheduledWorkflow, error) {
	var (
		workflow *types.ScheduledWorkflow
		err      error
	)

	err = s.WrapFn(nCtx, metricOperationGetScheduledWorkflow, func(nCtx contextx.IContext) error {
		var err error
		if workflow, err = s.getScheduledWorkflow(nCtx, workflowID); err != nil {
			logger.G.Sys().WithErr(err).With("workflow-id", workflowID).Error("failed to get scheduled workflow")

			return fmt.Errorf("failed to get workflow, workflow-id(%s): %w", workflowID, err)
		}

		return nil
	})

	return workflow, err
}

// CreateScheduledWorkflow creates a new scheduled workflow.
func (s *Storage) CreateScheduledWorkflow(nCtx contextx.IContext, workflow *types.ScheduledWorkflow) error {
	return s.WrapFn(nCtx, metricOperationCreateScheduledWorkflow, func(nCtx contextx.IContext) error {
		var err error
		if err = s.createScheduledWorkflow(nCtx, workflow); err != nil {
			logger.G.Sys().WithErr(err).With("workflow-id", workflow.WorkflowID).Error("failed to create scheduled workflow")

			return fmt.Errorf("failed to create scheduled workflow, workflow-id(%s): %w", workflow.WorkflowID, err)
		}

		return nil
	})
}

// UpdateScheduledWorkflowTriggerID updates a scheduled workflow's trigger ID.
func (s *Storage) UpdateScheduledWorkflowTriggerID(nCtx contextx.IContext, workflowID, triggerID string) error {
	return s.WrapFn(nCtx, metricOperationUpdateScheduledWorkflowTriggerID, func(nCtx contextx.IContext) error {
		var err error
		if err = s.updateScheduledWorkflowTriggerID(nCtx, workflowID, triggerID); err != nil {
			logger.G.Sys().WithErr(err).With("workflow-id", workflowID).Error("failed to update scheduled workflow trigger id")

			return err
		}

		return nil
	})
}

// UpdateScheduledWorkflowPrivateData updates a scheduled workflow's private data.
func (s *Storage) UpdateScheduledWorkflowPrivateData(nCtx contextx.IContext, workflowID string, privateData map[string]any) error {
	return s.WrapFn(nCtx, metricOperationUpdateScheduledWorkflowPrivateData, func(nCtx contextx.IContext) error {
		var err error
		if err = s.updateScheduledWorkflowPrivateData(nCtx, workflowID, privateData); err != nil {
			logger.G.Sys().WithErr(err).With("workflow-id", workflowID).Error("failed to update scheduled workflow private data")

			return err
		}

		return nil
	})
}

// SwitchScheduleWorkflow enables or disables a scheduled workflow.
func (s *Storage) SwitchScheduleWorkflow(nCtx contextx.IContext, workflowID string, enable bool) error {
	return s.WrapFn(nCtx, metricOperationSwitchScheduledWorkflow, func(nCtx contextx.IContext) error {
		var err error
		if err = s.switchScheduleWorkflow(nCtx, workflowID, enable); err != nil {
			logger.G.Sys().WithErr(err).With("workflow-id", workflowID).Error("failed to switch scheduled workflow")

			return err
		}

		return nil
	})
}

// GetOperation get operation by operationID.
func (s *Storage) GetOperation(nCtx contextx.IContext, operationID string) (*workoper.Operation, error) {
	var (
		oper *workoper.Operation
		err  error
	)
	err = s.WrapFn(nCtx, metricOperationGetOperation, func(nCtx contextx.IContext) error {
		var err error
		if oper, err = s.getOperation(nCtx, operationID); err != nil {
			logger.G.Sys().WithErr(err).With("operation-id", operationID).Error("failed to get operations")

			return fmt.Errorf("failed to get operations, operationID(%s): %w", operationID, err)
		}

		return nil
	})

	return oper, err
}

// UpsertOperation upsert operation.
func (s *Storage) UpsertOperation(nCtx contextx.IContext, operation *workoper.Operation) error {
	return s.WrapFn(nCtx, metricOperationUpsertOperation, func(nCtx contextx.IContext) error {
		var err error
		if err = s.upsertOperation(nCtx, operation); err != nil {
			logger.G.Sys().WithErr(err).With("operation-id", operation.OperationID).Error("failed to upsert operation")

			return fmt.Errorf("failed to upsert operation, operation-id(%s): %w", operation.OperationID, err)
		}

		return nil
	})
}

// ListOperation lists operation by page and condition.
func (s *Storage) ListOperation(
	nCtx contextx.IContext, page types.Page, conditions ...*types.OperationCondition) (
	[]*workoper.Operation, int64, error) {

	var (
		opers []*workoper.Operation
		num   int64
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationListOperation, func(nCtx contextx.IContext) error {
		var err error
		opers, num, err = s.listOperation(nCtx, page, conditions...)

		return err
	})

	return opers, num, err
}

// CountOperation counts operation by condition.
func (s *Storage) CountOperation(
	nCtx contextx.IContext, conditions ...*types.OperationCondition) (int64, error) {

	var (
		num int64
		err error
	)

	err = s.WrapFn(nCtx, metricOperationCountOperation, func(nCtx contextx.IContext) error {
		var err error
		num, err = s.countOperation(nCtx, conditions...)

		return err
	})

	return num, err
}

// DistinctOperation distincts operation fields by conditions.
func (s *Storage) DistinctOperation(
	nCtx contextx.IContext,
	selector types.WorkflowOperationDistinctSelector,
	conditions ...*types.OperationCondition) (*types.WorkflowOperationDistinctResult, error) {

	var (
		result *types.WorkflowOperationDistinctResult
		err    error
	)

	err = s.WrapFn(nCtx, metricOperationDistinctOperation, func(nCtx contextx.IContext) error {
		var err error
		result, err = s.distinctOperation(nCtx, selector, conditions...)

		return err
	})

	return result, err
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

	err = s.WrapFn(nCtx, metricOperationListOperationByOperationID, func(nCtx contextx.IContext) error {
		var err error
		if opers, num, err = s.listOperationByOperationID(nCtx, operationID...); err != nil {
			logger.G.Sys().WithErr(err).With("operation-ids", operationID).Error("failed to list operations by operation id")

			return fmt.Errorf("failed to list operations by operation id, operation-ids(%v): %w", operationID, err)
		}

		return nil
	})

	return opers, num, err
}

// ListOperationIDByParentOperationID lists operation IDs by parent operation ID.
func (s *Storage) ListOperationIDByParentOperationID(
	nCtx contextx.IContext, page types.Page, parentID ...string) ([]string, error) {

	var (
		ids []string
		err error
	)

	err = s.WrapFn(nCtx, metricOperationListOperationIDByParentOperationID, func(nCtx contextx.IContext) error {
		var err error
		if ids, err = s.listOperationIDByParentOperationID(nCtx, page, parentID...); err != nil {
			logger.G.Sys().WithErr(err).With("parent-ids", parentID).Error("failed to list operation ids by parent operation id")

			return fmt.Errorf("failed to list operation ids by parent operation id, parent-ids(%v): %w", parentID, err)
		}

		return nil
	})

	return ids, err
}

// ListOperationIDByParentOperInstID lists operation IDs by parent operation instance ID.
func (s *Storage) ListOperationIDByParentOperInstID(
	nCtx contextx.IContext, page types.Page, parentID ...string) ([]string, error) {

	var (
		ids []string
		err error
	)

	err = s.WrapFn(nCtx, metricOperationListOperationIDByParentOperInstID, func(nCtx contextx.IContext) error {
		var err error
		if ids, err = s.listOperationIDByParentOperInstID(nCtx, page, parentID...); err != nil {
			logger.G.Sys().WithErr(err).With("parent-ids", parentID).Error("failed to list operation ids by parent operation instance id")

			return fmt.Errorf("failed to list operation ids by parent operation instance id, parent-ids(%v): %w", parentID, err)
		}

		return nil
	})

	return ids, err
}

// ListNeedInstantiateOperationByTriggerID lists operations need to be instantiated by trigger id.
func (s *Storage) ListNeedInstantiateOperationByTriggerID(nCtx contextx.IContext, page types.Page, triggerID string) (
	[]*workoper.Operation, int64, error) {

	var (
		opers []*workoper.Operation
		num   int64
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationListNeedInstantiateOperationByTriggerID, func(nCtx contextx.IContext) error {
		var err error
		if opers, num, err = s.listNeedInstantiateOperationByTriggerID(nCtx, page, triggerID); err != nil {
			logger.G.Sys().WithErr(err).With("trigger-id", triggerID).Error("failed to list need instantiate operations")

			return fmt.Errorf("failed to list need instantiate operations, trigger-id(%s): %w", triggerID, err)
		}

		return nil
	})

	return opers, num, err
}

// ExistNeedInstantiateOperationByTriggerID checks whether there are operations need to be instantiated by trigger id.
func (s *Storage) ExistNeedInstantiateOperationByTriggerID(nCtx contextx.IContext, triggerID string) (bool, error) {
	var (
		exist bool
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationExistNeedInstantiateOperationByTriggerID, func(nCtx contextx.IContext) error {
		var err error
		if exist, err = s.existNeedInstantiateOperationByTriggerID(nCtx, triggerID); err != nil {
			logger.G.Sys().WithErr(err).With("trigger-id", triggerID).Error("failed to check whether exist need instantiate operations")

			return fmt.Errorf("failed to check whether exist need instantiate operations, trigger-id(%s): %w", triggerID, err)
		}

		return nil
	})

	return exist, err
}

// DeleteOperationsByTriggerID deletes operations by trigger ID.
func (s *Storage) DeleteOperationsByTriggerID(ctx contextx.IContext, triggerID ...string) error {
	return s.WrapFn(ctx, metricOperationDeleteOperationsByTriggerID, func(nCtx contextx.IContext) error {
		var err error
		if err = s.deleteOperationsByTriggerID(nCtx, triggerID...); err != nil {
			logger.G.Sys().WithErr(err).With("trigger-id", triggerID).Error("failed to delete operations by trigger id")

			return fmt.Errorf("failed to delete operations by trigger id, trigger-id(%s): %w", triggerID, err)
		}

		return nil
	})
}

// PullOperationInstanceIDsFromOperation pulls operation instance IDs from operation.
func (s *Storage) PullOperationInstanceIDsFromOperation(
	nCtx contextx.IContext, operationID string, operInstIDs ...string) error {

	return s.WrapFn(nCtx, metricOperationPullOperationInstanceIDsFromOperation, func(nCtx contextx.IContext) error {
		var err error
		if err = s.pullOperationInstanceIDs(nCtx, operationID, operInstIDs...); err != nil {
			logger.G.Sys().WithErr(err).With("operation-id", operationID, "oper-inst-ids", operInstIDs).Error("failed to pull operation instance ids")

			return fmt.Errorf("failed to pull operation instance ids, operation-id(%s), oper-inst-ids(%v): %w",
				operationID, operInstIDs, err)
		}

		return nil
	})
}

// UpdateOperationLatestInstBriefData updates operation's latest instance brief data.
func (s *Storage) UpdateOperationLatestInstBriefData(
	nCtx contextx.IContext, operationID string, briefData *workoper.InstanceBriefData) error {

	return s.WrapFn(nCtx, metricOperationUpdateOperationLatestInstBriefData, func(nCtx contextx.IContext) error {
		var err error
		if err = s.updateLatestInstBriefData(nCtx, operationID, briefData); err != nil {
			logger.G.Sys().WithErr(err).With("operation-id", operationID).Error("failed to update operation latest inst brief data")

			return fmt.Errorf("failed to update operation latest inst brief data, operation-id(%s): %w",
				operationID, err)
		}

		return nil
	})
}

// UpdateOperInstActionStatus update the oper inst action status.
func (s *Storage) UpdateOperInstActionStatus(
	nCtx contextx.IContext, operInstID string, actionName string, status action.State) error {

	return s.WrapFn(nCtx, metricOperationUpdateOperInstActionStatus, func(nCtx contextx.IContext) error {
		var err error
		if err = s.updateOperInstActionStatus(nCtx, operInstID, actionName, status); err != nil {
			logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID, "action-name", actionName).Error("failed to update oper inst action status")

			return fmt.Errorf("failed to update oper inst action status, oper-inst-id(%s), action-name(%s): %w",
				operInstID, actionName, err)
		}

		return nil
	})
}

// GetActionInstanceData gets full action instance data.
func (s *Storage) GetActionInstanceData(
	nCtx contextx.IContext, operInstID, actionName string) (*action.InstanceData, error) {

	var (
		data *action.InstanceData
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetActionInstanceData, func(nCtx contextx.IContext) error {
		var err error
		if data, err = s.getActionInstanceData(nCtx, operInstID, actionName); err != nil {
			logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID, "action-name", actionName).Error("failed to get action instance data")

			return fmt.Errorf("failed to get action instance data, oper-inst-id(%s), action-name(%s): %w",
				operInstID, actionName, err)
		}

		return nil
	})

	return data, err
}

// GetActionInstanceLifecycle gets action instance lifecycle.
func (s *Storage) GetActionInstanceLifecycle(
	nCtx contextx.IContext, operInstID, actionName string) (*action.Lifecycle, error) {

	var (
		lifecycle *action.Lifecycle
		err       error
	)

	err = s.WrapFn(nCtx, metricOperationGetActionInstanceLifecycle, func(nCtx contextx.IContext) error {
		var err error
		if lifecycle, err = s.getActionInstanceLifecycle(nCtx, operInstID, actionName); err != nil {
			logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID, "action-name", actionName).Error("failed to get action instance lifecycle")

			return fmt.Errorf("failed to get action instance lifecycle, oper-inst-id(%s), action-name(%s): %w",
				operInstID, actionName, err)
		}

		return nil
	})

	return lifecycle, err
}

// GetActionInstancePrivateData gets action instance private data.
func (s *Storage) GetActionInstancePrivateData(
	nCtx contextx.IContext, operInstID, actionName string) (map[string]any, error) {

	var (
		data map[string]any
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetActionInstancePrivateData, func(nCtx contextx.IContext) error {
		var err error
		if data, err = s.getActionInstancePrivateData(nCtx, operInstID, actionName); err != nil {
			logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID, "action-name", actionName).Error("failed to get action instance private data")

			return fmt.Errorf("failed to get action instance private data, oper-inst-id(%s), action-name(%s): %w",
				operInstID, actionName, err)
		}

		return nil
	})

	return data, err
}

// UpdateActionInstanceLifecycle updates action instance lifecycle.
func (s *Storage) UpdateActionInstanceLifecycle(
	nCtx contextx.IContext, operInstID, actionName string, lifecycle *action.Lifecycle) error {

	return s.WrapFn(nCtx, metricOperationUpdateActionInstanceLifecycle, func(nCtx contextx.IContext) error {
		var err error
		if err = s.updateActionInstanceLifecycle(nCtx, operInstID, actionName, lifecycle); err != nil {
			logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID, "action-name", actionName).Error("failed to update action instance lifecycle")

			return fmt.Errorf("failed to update action instance lifecycle, oper-inst-id(%s), action-name(%s): %w",
				operInstID, actionName, err)
		}

		return nil
	})
}

// PushActionInstanceMessage pushes action instance message.
func (s *Storage) PushActionInstanceMessage(
	nCtx contextx.IContext, operInstID, actionName string, messages ...common.Message) error {

	return s.WrapFn(nCtx, metricOperationPushActionInstanceMessage, func(nCtx contextx.IContext) error {
		var err error
		if err = s.pushActionInstanceMessage(nCtx, operInstID, actionName, messages...); err != nil {
			logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID, "action-name", actionName).Error("failed to push action instance message")

			return fmt.Errorf("failed to push action instance message, oper-inst-id(%s), action-name(%s): %w",
				operInstID, actionName, err)
		}

		return nil
	})
}

// GetOperationInstanceFullData gets full operation instance data.
func (s *Storage) GetOperationInstanceFullData(
	nCtx contextx.IContext, operInstID string) (*workoper.InstanceData, error) {

	var (
		data *workoper.InstanceData
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetOperationInstanceFullData, func(nCtx contextx.IContext) error {
		var err error
		if data, err = s.getOperationInstanceData(nCtx, operInstID); err != nil {
			logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID).Error("failed to get operation instance full data")

			return fmt.Errorf("failed to get operation instance data, oper-inst-id(%s): %w", operInstID, err)
		}

		return nil
	})

	return data, err
}

// GetOperationInstanceBriefData gets brief operation instance data.
func (s *Storage) GetOperationInstanceBriefData(
	nCtx contextx.IContext, operInstID string) (*workoper.InstanceBriefData, error) {

	var (
		briefData *workoper.InstanceBriefData
		err       error
	)

	err = s.WrapFn(nCtx, metricOperationGetOperationInstanceBriefData, func(nCtx contextx.IContext) error {
		var err error
		if briefData, err = s.getOperationInstanceDataBriefData(nCtx, operInstID); err != nil {
			logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID).Error("failed to get operation instance brief data")

			return fmt.Errorf("failed to get operation instance data, oper-inst-id(%s): %w", operInstID, err)
		}

		return nil
	})

	return briefData, err
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

	err = s.WrapFn(nCtx, metricOperationListOperationInstanceBriefDataWithoutActionInst, func(nCtx contextx.IContext) error {
		var err error
		if results, num, err = s.listOperationInstanceBriefDataWithoutActionInst(
			nCtx, page, conditions...); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to list operation instance brief data")

			return fmt.Errorf("failed to list operation instance brief data: %w", err)
		}

		return nil
	})

	return results, num, err
}

// CountOperationInstanceByState counts operation instance by state.
func (s *Storage) CountOperationInstanceByState(nCtx contextx.IContext, triggerID string, states ...workoper.State) (int64, error) {
	var (
		num int64
		err error
	)

	err = s.WrapFn(nCtx, metricOperationCountOperationInstance, func(nCtx contextx.IContext) error {
		var err error
		num, err = s.countOperationInstanceByState(nCtx, triggerID, states...)
		if err != nil {
			logger.G.Sys().WithErr(err).With("trigger-id", triggerID, "states", states).Error("failed to count operation instance")

			return fmt.Errorf("failed to count operation instance, trigger-id(%s), states(%v): %w",
				triggerID, states, err)
		}

		return nil
	})

	return num, err
}

// ExistOperationInstanceByState checks whether operation instance exists by state.
func (s *Storage) ExistOperationInstanceByState(nCtx contextx.IContext, triggerID string, states ...workoper.State) (bool, error) {
	var (
		exist bool
		err   error
	)

	err = s.WrapFn(nCtx, metricOperationExistOperationInstance, func(nCtx contextx.IContext) error {
		var err error
		exist, err = s.existOperationInstanceByState(nCtx, triggerID, states...)
		if err != nil {
			logger.G.Sys().WithErr(err).With("trigger-id", triggerID, "states", states).Error("failed to check whether operation instance exists")

			return fmt.Errorf("failed to check whether operation instance exists, trigger-id(%s), states(%v): %w",
				triggerID, states, err)
		}

		return nil
	})

	return exist, err
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

	err = s.WrapFn(nCtx, metricOperationListOperationInstanceBriefByOperationID, func(nCtx contextx.IContext) error {
		var err error
		results, num, err = s.listOperationInstanceBriefDataWithoutActionInstByOperationID(
			nCtx, page, operationID...)
		if err != nil {
			logger.G.Sys().WithErr(err).Error("failed to list operation instance brief data by operation")

			return fmt.Errorf("failed to list operation instance brief data by operation: %w", err)
		}

		return nil
	})

	return results, num, err
}

// ListOperInstanceBriefWithoutActionInstByTriggerID lists operation instance brief data.
func (s *Storage) ListOperInstanceBriefWithoutActionInstByTriggerID(nCtx contextx.IContext, page types.Page, triggerID ...string) (
	[]*workoper.InstanceBriefData, int64, error) {

	var (
		results []*workoper.InstanceBriefData
		num     int64
		err     error
	)

	err = s.WrapFn(nCtx, metricOperationListOperationInstanceBriefByTriggerID, func(nCtx contextx.IContext) error {
		var err error
		results, num, err = s.listOperationInstanceBriefDataWithoutActionInstByTriggerID(
			nCtx, page, triggerID...)
		if err != nil {
			logger.G.Sys().WithErr(err).Error("failed to list operation instance brief data by trigger")

			return fmt.Errorf("failed to list operation instance brief data by trigger: %w", err)
		}

		return nil
	})

	return results, num, err
}

// GetLatestOperationInstanceStatusDistributionByTriggerID gets the latest operation instance status distribution by trigger ID.
func (s *Storage) GetLatestOperationInstanceStatusDistributionByTriggerID(nCtx contextx.IContext, triggerID ...string) (
	map[string]*workoper.InstanceStatusDistribution, error) {

	var (
		distribution map[string]*workoper.InstanceStatusDistribution
		err          error
	)

	err = s.WrapFn(nCtx, metricOperationGetLatestOperationInstanceStatusDistribution, func(nCtx contextx.IContext) error {
		var err error
		distribution, err = s.getLatestOperationInstanceStatusDistributionByTriggerID(nCtx, triggerID...)

		return err
	})

	return distribution, err
}

// UpsertOperationInstanceData upserts operation instance data.
func (s *Storage) UpsertOperationInstanceData(
	nCtx contextx.IContext, operInstData *workoper.InstanceData) error {

	return s.WrapFn(nCtx, metricOperationUpsertOperationInstanceData, func(nCtx contextx.IContext) error {
		var err error
		if err = s.upsertOperationInstanceData(nCtx, operInstData); err != nil {
			logger.G.Sys().WithErr(err).With("oper-inst", operInstData).Error("failed to update operation instance data")

			return fmt.Errorf("failed to update operation instance data, operation-inst(%v): %w", operInstData, err)
		}

		return nil
	})
}

// UpdateOperationInstanceLifecycle updates operation instance lifecycle.
func (s *Storage) UpdateOperationInstanceLifecycle(
	nCtx contextx.IContext, operInstID string, lifecycle *workoper.Lifecycle) error {

	return s.WrapFn(nCtx, metricOperationUpdateOperationInstanceLifecycle, func(nCtx contextx.IContext) error {
		var err error
		if err = s.updateOperationInstanceLifecycle(nCtx, operInstID, lifecycle); err != nil {
			logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID).Error("failed to update operation instance lifecycle")

			return fmt.Errorf("failed to update operation instance lifecycle, oper-inst-id(%s): %w", operInstID, err)
		}

		return nil
	})
}

// UpdateOperationLatestActionInstBriefData updates operation latest action inst brief data.
func (s *Storage) UpdateOperationLatestActionInstBriefData(
	ctx contextx.IContext, operationInstanceID string, briefData *action.InstanceBriefData) error {

	return s.WrapFn(ctx, metricOperationUpdateOperationLatestActionInstBriefData, func(nCtx contextx.IContext) error {
		var err error
		if err = s.updateOperationLatestActionInstBriefData(nCtx, operationInstanceID, briefData); err != nil {
			logger.G.Sys().WithErr(err).With("oper-inst-id", operationInstanceID).Error("failed to update operation latest action inst brief data")

			return fmt.Errorf("failed to update operation latest action inst brief data, oper-inst-id(%s): %w", operationInstanceID, err)
		}

		return nil
	})
}

// UpdateOperationInstanceExtraExecutionMessages updates operation instance execution messages.
func (s *Storage) UpdateOperationInstanceExtraExecutionMessages(
	nCtx contextx.IContext, operInstID string, messages ...common.Message) error {

	return s.WrapFn(nCtx, metricOperationUpdateOperationInstanceExtraExecutionMessages, func(nCtx contextx.IContext) error {
		var err error
		if err = s.updateOperationInstanceExtraExecutionMessages(nCtx, operInstID, messages...); err != nil {
			logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID).Error("failed to update operation instance extra execution messages")

			return fmt.Errorf("failed to update operation instance extra execution messages, oper-inst-id(%s): %w",
				operInstID, err)
		}

		return nil
	})
}

// WatchOperInstStopping watches operation instance stopping.
func (s *Storage) WatchOperInstStopping(nCtx contextx.IContext, operInstID string) <-chan struct{} {
	return s.watchOperInstStopping(nCtx, operInstID)
}

// UpsertNeedStopOperInst upserts need stop operation instance.
func (s *Storage) UpsertNeedStopOperInst(nCtx contextx.IContext, operInstID string) error {
	return s.WrapFn(nCtx, metricOperationUpsertOperInstStop, func(nCtx contextx.IContext) error {
		var err error
		if err = s.upsertNeedStopOperInst(nCtx, operInstID); err != nil {
			logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID).Error("failed to upsert operation instance stop record")

			return fmt.Errorf("failed to upsert operation instance stop record, oper-inst-id(%s): %w", operInstID, err)
		}

		return nil
	})
}

// UpdateActionInstanceContent update action instance content.
func (s *Storage) UpdateActionInstanceContent(
	nCtx contextx.IContext, operInstID string, actionName string, content map[string]any) error {

	return s.WrapFn(nCtx, metricOperationUpdateActionInstanceContent, func(nCtx contextx.IContext) error {
		var err error
		if err = s.updateActionInstanceContent(nCtx, operInstID, actionName, content); err != nil {
			logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID, "action-name", actionName).Error("failed to update action instance content")

			return fmt.Errorf("failed to update action instance content, oper-inst-id(%s), action-name(%s): %w",
				operInstID, actionName, err)
		}

		return nil
	})
}

// UpsertActionInstancePrivateData upserts action instance private data.
func (s *Storage) UpsertActionInstancePrivateData(
	nCtx contextx.IContext, operInstID string, actionName string, privateData map[string]any) error {

	return s.WrapFn(nCtx, metricOperationUpsertActionInstancePrivateData, func(nCtx contextx.IContext) error {
		var err error
		if err = s.upsertActionInstancePrivateData(nCtx, operInstID, actionName, privateData); err != nil {
			logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID, "action-name", actionName).Error("failed to update operation instance private data")

			return fmt.Errorf(
				"failed to update operation instance private data, operation-inst-id(%s), action-name(%s): %w",
				operInstID, actionName, err)
		}

		return nil
	})
}

// DeleteOperationInstances deletes operation instances by given operation instance IDs.
func (s *Storage) DeleteOperationInstances(nCtx contextx.IContext, operInstID ...string) error {
	return s.WrapFn(nCtx, metricOperationDeleteOperationInstances, func(nCtx contextx.IContext) error {
		var err error
		if err = s.deleteOperationInstances(nCtx, operInstID...); err != nil {
			logger.G.Sys().WithErr(err).With("oper-inst-id", operInstID).Error("failed to delete operation instances")

			return fmt.Errorf("failed to delete operation instances, oper-inst-ids(%v): %w", operInstID, err)
		}

		return nil
	})
}

// DeleteOperationInstancesByTriggerID deletes operation instances by trigger ID.
func (s *Storage) DeleteOperationInstancesByTriggerID(ctx contextx.IContext, triggerID ...string) error {
	return s.WrapFn(ctx, metricOperationDeleteOperationInstancesByTriggerID, func(nCtx contextx.IContext) error {
		var err error
		if err = s.deleteOperationInstancesByTriggerID(nCtx, triggerID...); err != nil {
			logger.G.Sys().WithErr(err).With("trigger-id", triggerID).Error("failed to delete operation instances by trigger ID")

			return fmt.Errorf("failed to delete operation instances by trigger ID, trigger-ids(%v): %w", triggerID, err)
		}

		return nil
	})
}

// CreatePackageWorkflow creates a package workflow record.
func (s *Storage) CreatePackageWorkflow(nCtx contextx.IContext, workflow *types.PackageWorkflow) error {
	return s.WrapFn(nCtx, metricOperationCreatePackageWorkflow, func(nCtx contextx.IContext) error {
		if err := s.createPackageWorkflow(nCtx, workflow); err != nil {
			logger.G.Sys().WithErr(err).With("workflow-id", workflow.WorkflowID).Error("failed to create package workflow")

			return fmt.Errorf("failed to create package workflow, workflow-id(%s): %w", workflow.WorkflowID, err)
		}

		return nil
	})
}
