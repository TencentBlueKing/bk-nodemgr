/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package workflow ...
package workflow

import (
	"errors"
	"fmt"
	"time"

	"github.com/RichardKnop/machinery/v2/tasks"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/common"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/metric"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

// IController describes the workflow controller.
type IController interface {
	// CreateTrigger creates a new trigger.
	CreateTrigger(ctx contextx.IContext, category trigger.Category, metadata trigger.Metadata) (ITriggerCtl, error)

	// GetTrigger returns the trigger.
	GetTrigger(ctx contextx.IContext, triggerID string) (ITriggerCtl, error)
}

// ITriggerCtl describes the workflow trigger handler.
// nolint: interfacebloat
type ITriggerCtl interface {
	// GetTriggerID returns the trigger ID.
	GetTriggerID() string

	// GetTriggerCategory returns the trigger category.
	GetTriggerCategory() trigger.Category

	// GetTriggerMetadata returns the trigger metadata.
	GetTriggerMetadata() trigger.Metadata

	// GetLastTriggeredAt returns the last triggered at.
	GetLastTriggeredAt() time.Time

	// RunTrigger runs the trigger.
	RunTrigger(ctx contextx.IContext) error

	// TerminateTrigger terminates the trigger.
	TerminateTrigger(ctx contextx.IContext) error

	// CreateOperation creates a new operation under trigger.
	CreateOperation(
		ctx contextx.IContext, operationDef operation.Definition, param operation.Param) (IOperationCtl, error)

	// GetOperation returns the operation.
	GetOperation(ctx contextx.IContext, operationID string) (IOperationCtl, error)

	// ListOperation returns the operations.
	ListOperation(ctx contextx.IContext, operationID ...string) ([]IOperationCtl, error)

	// UpdateLastTriggeredTime updates the last triggered time.
	UpdateLastTriggeredTime(ctx contextx.IContext) error

	// ListEmptyOperation returns the operations without instances.
	ListEmptyOperation(ctx contextx.IContext, page types.Page) ([]IOperationCtl, error)

	// ListOperationInstances returns the operation instances with states.
	ListOperationInstances(
		ctx contextx.IContext, page types.Page, states ...operation.State) ([]IOperationInstanceCtl, error)
}

// IOperationCtl describes the workflow operation handler.
type IOperationCtl interface {
	// GetOperationID returns the operation ID.
	GetOperationID() string

	// CreateOperationInstance creates a new operation instance.
	CreateOperationInstance(ctx contextx.IContext) (IOperationInstanceCtl, error)

	// CreateRetryOperationInstance creates a new retry operation instance.
	CreateRetryOperationInstance(ctx contextx.IContext, retryMod types.NodeOperationRetryMode) (IOperationInstanceCtl, error)

	// GetOperationInstance returns the operation instance.
	GetOperationInstance(ctx contextx.IContext, operationInstanceID string) (IOperationInstanceCtl, error)
}

// IOperationInstanceCtl describes the workflow operation instance handler.
type IOperationInstanceCtl interface {
	// GetOperationInstanceID returns the operation instance ID.
	GetOperationInstanceID() string

	// LaunchOperationInstance launches the operation instance.
	LaunchOperationInstance(ctx contextx.IContext) error

	// TerminateOperationInstance terminates the operation instance.
	TerminateOperationInstance(ctx contextx.IContext) error
}

// CreateTrigger creates a new trigger.
func (mgr *manager) CreateTrigger(ctx contextx.IContext, category trigger.Category, metadata trigger.Metadata) (
	ITriggerCtl, error) {

	trig := &trigger.Trigger{
		TriggerID: identifier.GenTriggerID(),
		Category:  category,
		Metadata:  metadata,
		State:     trigger.StateInit,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := mgr.stgTrigger.CreateTrigger(ctx, trig); err != nil {
		return nil, err
	}

	return &controller{
		mgr:  mgr,
		trig: trig,
	}, nil
}

// GetTrigger returns the trigger.
func (mgr *manager) GetTrigger(ctx contextx.IContext, triggerID string) (ITriggerCtl, error) {
	trig, err := mgr.stgTrigger.GetTrigger(ctx, triggerID)
	if err != nil {
		return nil, err
	}

	return &controller{
		mgr:  mgr,
		trig: trig,
	}, nil
}

type controller struct {
	mgr *manager

	trig *trigger.Trigger
	oper *operation.Operation

	// operation instance.
	operInstanceBriefData *operation.InstanceBriefData
}

// GetTriggerID returns the trigger ID.
func (ctl *controller) GetTriggerID() string {
	return ctl.trig.TriggerID
}

// GetTriggerCategory returns the trigger category.
func (ctl *controller) GetTriggerCategory() trigger.Category {
	return ctl.trig.Category
}

// GetTriggerMetadata returns the trigger metadata.
func (ctl *controller) GetTriggerMetadata() trigger.Metadata {
	return ctl.trig.Metadata
}

// GetLastTriggeredAt returns the last triggered at.
func (ctl *controller) GetLastTriggeredAt() time.Time {
	return ctl.trig.LastTriggeredAt
}

// Run runs the trigger.
func (ctl *controller) RunTrigger(ctx contextx.IContext) error {
	return ctl.mgr.stgTrigger.UpdateTriggerState(ctx, ctl.trig.TriggerID, trigger.StateRunning)
}

// Terminate terminates the trigger.
func (ctl *controller) TerminateTrigger(ctx contextx.IContext) error {
	return ctl.mgr.stgTrigger.UpdateTriggerState(ctx, ctl.trig.TriggerID, trigger.StateTerminated)
}

// CreateOperation creates a new operation under trigger.
func (ctl *controller) CreateOperation(
	ctx contextx.IContext, operationDef operation.Definition, param operation.Param) (IOperationCtl, error) {

	ctl.mgr.logger.InfoCtxf(ctx, "try to create operation. trigger-id(%s), operation-name(%s), param(%v)",
		ctl.trig.TriggerID, operationDef.Name(), param)

	oper := &operation.Operation{
		OperationID: identifier.GenOperationID(),
		TriggerID:   ctl.trig.TriggerID,
		Definition:  operationDef,
		Param:       param,
	}

	if err := ctl.mgr.stgOperation.UpsertOperation(ctx, oper); err != nil {
		ctl.mgr.logger.ErrorCtxf(ctx,
			"failed to create operation, failed to upsert. trigger-id(%s), operation-name(%s), param(%v), err(%v)",
			ctl.trig.TriggerID, operationDef.Name(), param, err)

		return nil, err
	}

	return &controller{
		mgr:  ctl.mgr,
		trig: ctl.trig,
		oper: oper,
	}, nil
}

// GetOperation returns the operation.
func (ctl *controller) GetOperation(ctx contextx.IContext, operationID string) (IOperationCtl, error) {
	oper, err := ctl.mgr.stgOperation.GetOperation(ctx, operationID)
	if err != nil {
		return nil, err
	}

	if oper.TriggerID != ctl.trig.TriggerID {
		return nil, errors.Join(common.ErrInvalidParameters(),
			fmt.Errorf("operation(%s) not belong to trigger(%s)", operationID, ctl.trig.TriggerID))
	}

	return &controller{
		mgr:  ctl.mgr,
		trig: ctl.trig,
		oper: oper,
	}, nil
}

// ListOperation returns the operations.
func (ctl *controller) ListOperation(ctx contextx.IContext, operationID ...string) ([]IOperationCtl, error) {
	opers, num, err := ctl.mgr.stgOperation.ListOperationByOperationID(ctx, operationID...)
	if err != nil {
		return nil, err
	}

	if num != int64(len(operationID)) {
		return nil, errors.New("get many operation failed, match num not equal to count")
	}

	ctls := make([]IOperationCtl, len(opers))
	for idx, oper := range opers {
		ctls[idx] = &controller{
			mgr:  ctl.mgr,
			trig: ctl.trig,
			oper: oper,
		}
	}

	return ctls, nil
}

// UpdateLastTriggeredTime updates the last triggered time.
func (ctl *controller) UpdateLastTriggeredTime(ctx contextx.IContext) error {
	ctl.trig.LastTriggeredAt = time.Now()
	if err := ctl.mgr.stgTrigger.UpdateTrigger(ctx, ctl.trig); err != nil {
		ctl.mgr.logger.ErrorCtxf(ctx,
			"failed to update last triggered time, failed to update. trigger-id(%s), err(%v)",
			ctl.trig.TriggerID, err)

		return err
	}

	return nil
}

// ListEmptyOperation returns the empty operation.
func (ctl *controller) ListEmptyOperation(ctx contextx.IContext, page types.Page) ([]IOperationCtl, error) {
	opers, _, err := ctl.mgr.stgOperation.ListEmptyOperationByTriggerID(ctx, page, ctl.trig.TriggerID)
	if err != nil {
		return nil, err
	}

	ctls := make([]IOperationCtl, len(opers))
	for idx, oper := range opers {
		ctls[idx] = &controller{
			mgr:  ctl.mgr,
			trig: ctl.trig,
			oper: oper,
		}
	}

	return ctls, nil
}

// ListOperationInstances returns the operation instances with states.
func (ctl *controller) ListOperationInstances(
	ctx contextx.IContext, page types.Page, states ...operation.State) ([]IOperationInstanceCtl, error) {

	condition := &types.OperInstDataCondition{
		ExactInclude: &types.OperInstDataExactFields{
			TriggerID: []string{ctl.trig.TriggerID},
			State:     states,
		},
	}
	instanceBriefData, _, err := ctl.mgr.stgOperationInstance.ListOperationInstanceBriefDataWithoutActionInst(
		ctx, page, condition)
	if err != nil {
		return nil, err
	}

	ctls := make([]IOperationInstanceCtl, len(instanceBriefData))
	for idx, data := range instanceBriefData {
		ctls[idx] = &controller{
			mgr:                   ctl.mgr,
			trig:                  ctl.trig,
			oper:                  ctl.oper,
			operInstanceBriefData: data,
		}
	}

	return ctls, nil
}

// GetOperationID returns the operation id.
func (ctl *controller) GetOperationID() string {
	return ctl.oper.OperationID
}

// CreateOperationInstance creates a new operation instance.
func (ctl *controller) CreateOperationInstance(ctx contextx.IContext) (IOperationInstanceCtl, error) {
	instanceData, err := ctl.createOperationInstanceBase(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &controller{
		mgr:                   ctl.mgr,
		trig:                  ctl.trig,
		oper:                  ctl.oper,
		operInstanceBriefData: &instanceData.InstanceBriefData,
	}, nil
}

// GetOperationInstance returns the operation instance.
func (ctl *controller) GetOperationInstance(
	ctx contextx.IContext, operationInstanceID string) (IOperationInstanceCtl, error) {

	instanceBriefData, err := ctl.mgr.stgOperationInstance.GetOperationInstanceBriefData(ctx, operationInstanceID)
	if err != nil {
		return nil, err
	}

	return &controller{
		mgr:                   ctl.mgr,
		trig:                  ctl.trig,
		oper:                  ctl.oper,
		operInstanceBriefData: instanceBriefData,
	}, nil
}

// GetOperationInstanceID returns the operation instance ID.
func (ctl *controller) GetOperationInstanceID() string {
	return ctl.operInstanceBriefData.Metadata.OperationInstanceID
}

// LaunchOperationInstance launches the operation instance.
// nolint: nonamedreturns
func (ctl *controller) LaunchOperationInstance(ctx contextx.IContext) (err error) {
	// record metric.
	m := metric.NewOperationInstanceLaunch(ctl.operInstanceBriefData).Start()
	defer m.End(err)

	ctl.mgr.logger.InfoCtxf(ctx, "try to launch operation instance. oper-inst-id(%s)",
		ctl.operInstanceBriefData.Metadata.OperationInstanceID)

	if ctl.operInstanceBriefData.Lifecycle.State != operation.StateInit {
		return errors.Join(common.ErrInvalidState(),
			fmt.Errorf("operation instance is not in init state: %s", ctl.operInstanceBriefData.Lifecycle.State))
	}

	// ensure operation is not nil
	if ctl.oper == nil {
		var err error
		ctl.oper, err = ctl.mgr.stgOperation.GetOperation(ctx, ctl.operInstanceBriefData.Metadata.OperationID)
		if err != nil {
			return err
		}
	}

	// update state.
	// ctl.operInstanceBriefData.Lifecycle.StartedAt = time.Now()
	ctl.operInstanceBriefData.Lifecycle.State = operation.StateLaunched
	if err := ctl.mgr.stgOperationInstance.UpdateOperationInstanceLifecycle(
		ctx, ctl.operInstanceBriefData.Metadata.OperationInstanceID, ctl.operInstanceBriefData.Lifecycle); err != nil {
		ctl.mgr.logger.ErrorCtxf(ctx, "failed to update operation instance lifecycle. oper-inst-id(%s), err(%v)",
			ctl.operInstanceBriefData.Metadata.OperationInstanceID, err)

		return err
	}

	actionNames := ctl.oper.Definition.ActionDefNames()
	if len(actionNames) == 0 {
		return errors.Join(common.ErrNoActionTodo(),
			fmt.Errorf("operation instance has no action. oper-inst-id(%s)",
				ctl.operInstanceBriefData.Metadata.OperationInstanceID))
	}

	signatures := make([]*tasks.Signature, len(actionNames))
	for idx, actionName := range actionNames {
		signatures[idx] = &tasks.Signature{
			UUID: identifier.GenActionInstanceID(),
			Name: actionName,
			Args: []tasks.Arg{
				{
					Name:  "action",
					Type:  "string",
					Value: actionName,
				},
				{
					Name:  "oper-inst-id",
					Type:  "string",
					Value: ctl.operInstanceBriefData.Metadata.OperationInstanceID,
				},
			},
		}
	}

	chain, err := tasks.NewChain(signatures...)
	if err != nil {
		return err
	}

	ctl.mgr.logger.InfoCtxf(ctx, "send chain to machinery. oper-inst-id(%s)",
		ctl.operInstanceBriefData.Metadata.OperationInstanceID)
	_, err = ctl.mgr.server.SendChainWithContext(ctx, chain)
	if err != nil {
		return fmt.Errorf("send chain to machinery failed, err: %v", err)
	}

	return nil
}

// TerminateOperationInstance terminates the operation instance.
func (ctl *controller) TerminateOperationInstance(_ contextx.IContext) error {
	return errors.New("not implemented")
}

// CreateRetryOperationInstance creates a retry operation instance.
func (ctl *controller) CreateRetryOperationInstance(ctx contextx.IContext, retryMod types.NodeOperationRetryMode) (
	IOperationInstanceCtl, error) {

	if len(ctl.oper.InstanceIDs) == 0 {
		return nil, errors.New("operation has no instance")
	}

	lastInstanceID := ctl.oper.InstanceIDs[len(ctl.oper.InstanceIDs)-1]
	prevInstance, err := ctl.mgr.stgOperationInstance.GetOperationInstanceFullData(ctx, lastInstanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get last operation instance, err: %v", err)
	}

	switch retryMod {
	case types.OperationRetryModeFull:
		return ctl.handleFullRetry(ctx)
	case types.OperationRetryModePartial:
		return ctl.handlePartialRetry(ctx, prevInstance)
	default:
		return nil, fmt.Errorf("unknown retry mode: %v", retryMod)
	}
}

func (ctl *controller) handlePartialRetry(ctx contextx.IContext, prevInstance *operation.InstanceData) (
	IOperationInstanceCtl, error) {

	if len(ctl.oper.Param.RetryStartPoint) == 0 {
		return nil, errors.New("retry start point is empty")
	}

	actionNames := ctl.oper.Definition.ActionDefNames()
	if len(actionNames) == 0 {
		return nil, errors.New("operation has no actions")
	}

	actionIndexMap := make(map[string]int, len(actionNames))
	for i, name := range actionNames {
		actionIndexMap[name] = i
	}

	failedAction, retryStartAction, startIndex, err := ctl.findRetryStartInfo(
		prevInstance, actionNames, actionIndexMap)
	if err != nil {
		return nil, err
	}

	ctl.mgr.logger.InfoCtxf(ctx, "retry start action (%s), first failed action (%s)",
		retryStartAction, failedAction)

	return ctl.createPartialRetryInstance(ctx, prevInstance, actionNames, actionIndexMap, startIndex)
}

func (ctl *controller) findRetryStartInfo(
	prevInstance *operation.InstanceData,
	actionNames []string,
	actionIndexMap map[string]int,
) (string, string, int, error) {

	var failedAction string
	failedIndex := -1
	for _, name := range actionNames {
		inst, exists := prevInstance.ActionInstanceDataMap[name]
		if !exists {
			continue
		}
		if inst.Lifecycle.State != action.StateSuccess {
			failedAction = name
			failedIndex = actionIndexMap[name]

			break
		}
	}
	if failedAction == "" {
		return "", "", -1, errors.New("no failed actions found for partial retry")
	}

	retryStartPoints := ctl.oper.Param.RetryStartPoint
	retryStartAction := ""
	startIndex := failedIndex
	for ; startIndex >= 0; startIndex-- {
		actionName := actionNames[startIndex]
		if retryStartPoints[actionName] {
			retryStartAction = actionName
			break
		}
	}
	if retryStartAction == "" {
		return "", "", -1, fmt.Errorf("no valid retry start point found for failed action (%s)", failedAction)
	}

	return failedAction, retryStartAction, startIndex, nil
}

func (ctl *controller) createPartialRetryInstance(
	ctx contextx.IContext,
	prevInstance *operation.InstanceData,
	actionNames []string,
	actionIndexMap map[string]int,
	startIndex int,
) (IOperationInstanceCtl, error) {

	stateDecider := func(actionName string) action.State {
		currentIndex, exists := actionIndexMap[actionName]
		if !exists || currentIndex < startIndex {
			return action.StateSkipped
		}

		return action.StatePending
	}

	instanceData, err := ctl.createOperationInstanceBase(ctx, stateDecider)
	if err != nil {
		return nil, err
	}

	if startIndex > 0 {
		lastActionBeforeStart := actionNames[startIndex-1]

		prevActionInst, exists := prevInstance.ActionInstanceDataMap[lastActionBeforeStart]
		if exists && prevActionInst.Content != nil {
			newActionInst, exists := instanceData.ActionInstanceDataMap[lastActionBeforeStart]
			if exists {
				newContent := make(map[string]any, len(prevActionInst.Content))
				for k, v := range prevActionInst.Content {
					newContent[k] = v
				}
				newActionInst.Content = newContent
			}
		}
	}

	if err := ctl.mgr.stgOperationInstance.UpsertOperationInstanceData(ctx, instanceData); err != nil {
		return nil, fmt.Errorf("failed to update instance with copied content, err: %w", err)
	}

	return &controller{
		mgr:                   ctl.mgr,
		trig:                  ctl.trig,
		oper:                  ctl.oper,
		operInstanceBriefData: &instanceData.InstanceBriefData,
	}, nil
}

func (ctl *controller) handleFullRetry(ctx contextx.IContext) (IOperationInstanceCtl, error) {
	instanceData, err := ctl.createOperationInstanceBase(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &controller{
		mgr:                   ctl.mgr,
		trig:                  ctl.trig,
		oper:                  ctl.oper,
		operInstanceBriefData: &instanceData.InstanceBriefData,
	}, nil
}

func (ctl *controller) createOperationInstanceBase(ctx contextx.IContext,
	stateDecider func(string) action.State,
) (*operation.InstanceData, error) {

	operationInstanceID := identifier.GenOperationInstanceID()
	actionNames := ctl.oper.Definition.ActionDefNames()

	actionInstanceDataMap := make(map[string]*action.InstanceData)
	for index, actionName := range actionNames {
		state := action.StatePending
		if stateDecider != nil {
			state = stateDecider(actionName)
		}

		actionInstanceDataMap[actionName] = &action.InstanceData{
			TriggerID:           ctl.trig.TriggerID,
			OperationID:         ctl.oper.OperationID,
			OperationDefName:    ctl.oper.Definition.Name(),
			OperationInstanceID: operationInstanceID,
			Name:                actionName,
			Index:               index,
			TotalIndex:          len(actionNames),
			Messages:            make([]common.Message, 0),
			Content:             make(map[string]any),
			PrivateData:         make(map[string]any),
			Lifecycle: &action.Lifecycle{
				CreatedAt: time.Now().Local(),
				State:     state,
			},
		}
	}

	instanceData := &operation.InstanceData{
		InstanceBriefData: operation.InstanceBriefData{
			Metadata: &operation.InstanceMetadata{
				TriggerID:              ctl.trig.TriggerID,
				OperationInstanceID:    operationInstanceID,
				OperationDefName:       ctl.oper.Definition.Name(),
				OperationID:            ctl.oper.OperationID,
				ActionNames:            actionNames,
				Index:                  len(ctl.oper.InstanceIDs),
				ParentOperationID:      ctl.oper.Param.ParentOperationID,
				Timeout:                ctl.oper.Param.Timeout,
				InitContent:            ctl.oper.Param.InitContent,
				ExtraExecutionName:     ctl.oper.Definition.ExtraExecutionName(),
				ExtraExecutionMessages: make([]common.Message, 0),
			},
			Lifecycle: &operation.Lifecycle{
				CreatedAt: time.Now().Local(),
				State:     operation.StateInit,
			},
		},
		ActionInstanceDataMap: actionInstanceDataMap,
	}

	if err := ctl.mgr.stgOperationInstance.UpsertOperationInstanceData(ctx, instanceData); err != nil {
		return nil, err
	}

	ctl.oper.InstanceIDs = append(ctl.oper.InstanceIDs, operationInstanceID)
	if err := ctl.mgr.stgOperation.UpsertOperation(ctx, ctl.oper); err != nil {
		return nil, err
	}

	return instanceData, nil
}
