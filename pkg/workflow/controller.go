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
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RichardKnop/machinery/v2/tasks"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/common"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

// IController describes the workflow controller.
type IController interface {
	// CreateTrigger creates a new trigger.
	CreateTrigger(ctx context.Context, category trigger.Category, metadata trigger.Metadata) (ITriggerCtl, error)

	// GetTrigger returns the trigger.
	GetTrigger(ctx context.Context, triggerID string) (ITriggerCtl, error)
}

// ITriggerCtl describes the workflow trigger handler.
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
	RunTrigger(ctx context.Context) error

	// TerminateTrigger terminates the trigger.
	TerminateTrigger(ctx context.Context) error

	// CreateOperation creates a new operation under trigger.
	CreateOperation(
		ctx context.Context, operationDef operation.Definition, param operation.Param) (IOperationCtl, error)

	// GetOperation returns the operation.
	GetOperation(ctx context.Context, operationID string) (IOperationCtl, error)

	// UpdateLastTriggeredTime updates the last triggered time.
	UpdateLastTriggeredTime(ctx context.Context) error

	// ListEmptyOperation returns the operations without instances.
	ListEmptyOperation(ctx context.Context, page types.Page) ([]IOperationCtl, error)

	// ListOperationInstances returns the operation instances with states.
	ListOperationInstances(
		ctx context.Context, page types.Page, states ...operation.State) ([]IOperationInstanceCtl, error)
}

// IOperationCtl describes the workflow operation handler.
type IOperationCtl interface {
	// GetOperationID returns the operation ID.
	GetOperationID() string

	// CreateOperationInstance creates a new operation instance.
	CreateOperationInstance(ctx context.Context) (IOperationInstanceCtl, error)

	// GetOperationInstance returns the operation instance.
	GetOperationInstance(ctx context.Context, operationInstanceID string) (IOperationInstanceCtl, error)
}

// IOperationInstanceCtl describes the workflow operation instance handler.
type IOperationInstanceCtl interface {
	// GetOperationInstanceID returns the operation instance ID.
	GetOperationInstanceID() string

	// LaunchOperationInstance launches the operation instance.
	LaunchOperationInstance(ctx context.Context) error

	// TerminateOperationInstance terminates the operation instance.
	TerminateOperationInstance(ctx context.Context) error
}

// CreateTrigger creates a new trigger.
func (mgr *manager) CreateTrigger(ctx context.Context, category trigger.Category, metadata trigger.Metadata) (
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
func (mgr *manager) GetTrigger(ctx context.Context, triggerID string) (ITriggerCtl, error) {
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
func (ctl *controller) RunTrigger(ctx context.Context) error {
	return ctl.mgr.stgTrigger.UpdateTriggerState(ctx, ctl.trig.TriggerID, trigger.StateRunning)
}

// Terminate terminates the trigger.
func (ctl *controller) TerminateTrigger(ctx context.Context) error {
	return ctl.mgr.stgTrigger.UpdateTriggerState(ctx, ctl.trig.TriggerID, trigger.StateTerminated)
}

// CreateOperation creates a new operation under trigger.
func (ctl *controller) CreateOperation(
	ctx context.Context, operationDef operation.Definition, param operation.Param) (IOperationCtl, error) {

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
func (ctl *controller) GetOperation(ctx context.Context, operationID string) (IOperationCtl, error) {
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

// UpdateLastTriggeredTime updates the last triggered time.
func (ctl *controller) UpdateLastTriggeredTime(ctx context.Context) error {
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
func (ctl *controller) ListEmptyOperation(ctx context.Context, page types.Page) ([]IOperationCtl, error) {
	opers, _, err := ctl.mgr.stgOperation.ListEmptyOperation(ctx, page, ctl.trig.TriggerID)
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
	ctx context.Context, page types.Page, states ...operation.State) ([]IOperationInstanceCtl, error) {

	condition := operation.ListOperationInstanceCondition{
		TriggerIDs: []string{ctl.trig.TriggerID},
		States:     states,
	}

	instanceBriefData, _, err := ctl.mgr.stgOperationInstance.ListOperationInstanceBriefData(ctx, page, condition)
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
func (ctl *controller) CreateOperationInstance(ctx context.Context) (IOperationInstanceCtl, error) {
	operationInstanceID := identifier.GenOperationInstanceID()
	actionNames := ctl.oper.Definition.ActionDefNames()

	actionInstanceDataMap := make(map[string]*action.InstanceData)
	for index, actionName := range actionNames {
		actionInstanceDataMap[actionName] = &action.InstanceData{
			TriggerID:           ctl.trig.TriggerID,
			OperationID:         ctl.oper.OperationID,
			OperationDefName:    ctl.oper.Definition.Name(),
			OperationInstanceID: operationInstanceID,

			Name:        actionName,
			Index:       index,
			TotalIndex:  len(actionNames),
			Messages:    make([]action.Message, 0),
			Content:     make(map[string]any),
			PrivateData: make(map[string]any),
			Lifecycle: &action.Lifecycle{
				CreatedAt: time.Now().Local(),
				State:     action.StatePending,
			},
		}
	}

	instanceData := &operation.InstanceData{
		InstanceBriefData: operation.InstanceBriefData{
			Metadata: &operation.InstanceMetadata{
				TriggerID:           ctl.trig.TriggerID,
				OperationInstanceID: operationInstanceID,
				OperationDefName:    ctl.oper.Definition.Name(),
				OperationID:         ctl.oper.OperationID,
				ActionNames:         actionNames,
				Index:               len(ctl.oper.InstanceIDs),
				ParentOperationID:   ctl.oper.Param.ParentOperationID,
				Timeout:             ctl.oper.Param.Timeout,
				InitContent:         ctl.oper.Param.InitContent,
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

	return &controller{
		mgr:                   ctl.mgr,
		trig:                  ctl.trig,
		oper:                  ctl.oper,
		operInstanceBriefData: &instanceData.InstanceBriefData,
	}, nil
}

// GetOperationInstance returns the operation instance.
func (ctl *controller) GetOperationInstance(
	ctx context.Context, operationInstanceID string) (IOperationInstanceCtl, error) {

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
func (ctl *controller) LaunchOperationInstance(ctx context.Context) error {
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
	ctl.operInstanceBriefData.Lifecycle.StartedAt = time.Now()
	ctl.operInstanceBriefData.Lifecycle.State = operation.StateRunning
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
func (ctl *controller) TerminateOperationInstance(_ context.Context) error {
	return errors.New("not implemented")
}
