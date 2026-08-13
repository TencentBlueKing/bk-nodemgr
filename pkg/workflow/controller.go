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
	"maps"
	"time"

	"github.com/RichardKnop/machinery/v2/tasks"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/common"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/metric"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// IController describes the workflow controller.
type IController interface {
	// CreateTrigger creates a new trigger.
	CreateTrigger(nCtx contextx.IContext, category trigger.Category, metadata trigger.Metadata) (ITriggerCtl, error)

	// GetTrigger returns the trigger.
	GetTrigger(nCtx contextx.IContext, triggerID string) (ITriggerCtl, error)
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

	// IsActive returns whether the trigger is active.
	IsActive() bool

	// ActivateTrigger activates the trigger.
	ActivateTrigger(nCtx contextx.IContext) error

	// InactivateTrigger inactivates the trigger.
	InactivateTrigger(nCtx contextx.IContext) error

	// TryInactivateTrigger try inactivates the trigger.
	TryInactivateTrigger(nCtx contextx.IContext) error

	// CreateOperation creates a new operation under trigger.
	CreateOperation(nCtx contextx.IContext, operationDef operation.Definition, param operation.Param) (IOperationCtl, error)

	// GetOperation returns the operation.
	GetOperation(nCtx contextx.IContext, operationID string) (IOperationCtl, error)

	// ListOperation returns the operations.
	ListOperation(nCtx contextx.IContext, operationID ...string) ([]IOperationCtl, error)

	// UpdateLastTriggeredTime updates the last triggered time.
	UpdateLastTriggeredTime(nCtx contextx.IContext) error

	// UpdateMetadata updates the last triggered time.
	UpdateMetadata(nCtx contextx.IContext, metadata trigger.Metadata) error

	// UpdateOperationRetryFlag updates the operation retry flag.
	UpdateOperationRetryFlag(nCtx contextx.IContext, mode operation.RetryMode, operationID ...string) error

	// ListNeedInstantiateOperation returns the operations need to be instantiated.
	ListNeedInstantiateOperation(nCtx contextx.IContext, page types.Page) ([]IOperationCtl, error)

	// ListOperationInstances returns the operation instances with states.
	ListOperationInstances(nCtx contextx.IContext, page types.Page, states ...operation.State) ([]IOperationInstanceCtl, error)
}

// IOperationCtl describes the workflow operation handler.
type IOperationCtl interface {
	// GetOperationID returns the operation ID.
	GetOperationID() string

	// GetLastOperationInstance returns the last operation instance.
	GetLastOperationInstance(nCtx contextx.IContext) (IOperationInstanceCtl, error)

	// CreateOperationInstance creates a new operation instance.
	CreateOperationInstance(nCtx contextx.IContext) (IOperationInstanceCtl, error)

	// GetOperationInstance returns the operation instance.
	GetOperationInstance(nCtx contextx.IContext, operationInstanceID string) (IOperationInstanceCtl, error)
}

// IOperationInstanceCtl describes the workflow operation instance handler.
type IOperationInstanceCtl interface {
	// GetOperationInstanceID returns the operation instance ID.
	GetOperationInstanceID() string

	// LaunchOperationInstance launches the operation instance.
	LaunchOperationInstance(nCtx contextx.IContext) error

	// TerminateOperationInstance terminates the operation instance.
	TerminateOperationInstance(nCtx contextx.IContext) error
}

// CreateTrigger creates a new trigger.
func (mgr *manager) CreateTrigger(nCtx contextx.IContext, category trigger.Category, metadata trigger.Metadata) (
	ITriggerCtl, error) {

	trig := &trigger.Trigger{
		TriggerID: identifier.GenTriggerID(),
		Category:  category,
		Metadata:  metadata,
		Active:    false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := mgr.stgTrigger.CreateTrigger(nCtx, trig); err != nil {
		return nil, err
	}

	return &controller{
		mgr:  mgr,
		trig: trig,
	}, nil
}

// GetTrigger returns the trigger.
func (mgr *manager) GetTrigger(nCtx contextx.IContext, triggerID string) (ITriggerCtl, error) {
	trig, err := mgr.stgTrigger.GetTrigger(nCtx, triggerID)
	if err != nil {
		return nil, err
	}

	return &controller{
		mgr:  mgr,
		trig: trig,
	}, nil
}

var _ ITriggerCtl = &controller{}

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

// IsActive returns whether the trigger is active.
func (ctl *controller) IsActive() bool {
	return ctl.trig.Active
}

// Activate activates the trigger.
func (ctl *controller) ActivateTrigger(nCtx contextx.IContext) error {
	return ctl.mgr.stgTrigger.SwitchTriggerActive(nCtx, ctl.trig.TriggerID, true)
}

// Inactivate inactivates the trigger.
func (ctl *controller) InactivateTrigger(nCtx contextx.IContext) error {
	return ctl.mgr.stgTrigger.SwitchTriggerActive(nCtx, ctl.trig.TriggerID, false)
}

// TryInactivateTrigger try inactivates the trigger.
func (ctl *controller) TryInactivateTrigger(nCtx contextx.IContext) error {
	operNeedInstantiated, err := ctl.mgr.stgOperation.ExistNeedInstantiateOperationByTriggerID(nCtx, ctl.trig.TriggerID)
	if err != nil {
		return err
	}

	operInstNeedLauched, err := ctl.mgr.stgOperationInstance.ExistOperationInstanceByState(nCtx, ctl.trig.TriggerID, operation.StateInit)
	if err != nil {
		return err
	}

	if operNeedInstantiated || operInstNeedLauched {
		logger.G.Sys().With("trigger-id", ctl.trig.TriggerID).
			Info("trigger need instantiate operations or need lauched operation instances, skip inactivate")

		return nil
	}

	return ctl.mgr.stgTrigger.SwitchTriggerActive(nCtx, ctl.trig.TriggerID, false)
}

// CreateOperation creates a new operation under trigger.
func (ctl *controller) CreateOperation(nCtx contextx.IContext, operationDef operation.Definition, param operation.Param) (IOperationCtl, error) {
	logger.G.Sys().With("trigger-id", ctl.trig.TriggerID, "operation", operationDef.Name(), "param", param).Info("try to create operation")

	oper := &operation.Operation{
		OperationID: identifier.GenOperationID(),
		TriggerID:   ctl.trig.TriggerID,
		Definition:  operationDef,
		Param:       param,
		CreateTime:  time.Now(),
	}

	if err := ctl.mgr.stgOperation.UpsertOperation(nCtx, oper); err != nil {
		logger.G.Sys().
			WithErr(err).
			With("trigger-id", ctl.trig.TriggerID, "operation", operationDef.Name(), "param", param).
			Error("failed to create operation")

		return nil, err
	}

	return &controller{
		mgr:  ctl.mgr,
		trig: ctl.trig,
		oper: oper,
	}, nil
}

// GetOperation returns the operation.
func (ctl *controller) GetOperation(nCtx contextx.IContext, operationID string) (IOperationCtl, error) {
	oper, err := ctl.mgr.stgOperation.GetOperation(nCtx, operationID)
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
func (ctl *controller) ListOperation(nCtx contextx.IContext, operationID ...string) ([]IOperationCtl, error) {
	opers, num, err := ctl.mgr.stgOperation.ListOperationByOperationID(nCtx, operationID...)
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
func (ctl *controller) UpdateLastTriggeredTime(nCtx contextx.IContext) error {
	ctl.trig.LastTriggeredAt = time.Now()
	if err := ctl.mgr.stgTrigger.UpdateTrigger(nCtx, ctl.trig); err != nil {
		logger.G.Sys().WithErr(err).With("trigger-id", ctl.trig.TriggerID).Error("failed to update last triggered time")

		return err
	}

	return nil
}

// UpdateMetadata updates the metadata.
func (ctl *controller) UpdateMetadata(nCtx contextx.IContext, metadata trigger.Metadata) error {
	ctl.trig.Metadata = metadata
	if err := ctl.mgr.stgTrigger.UpdateTrigger(nCtx, ctl.trig); err != nil {
		logger.G.Sys().WithErr(err).With("trigger-id", ctl.trig.TriggerID).Error("failed to update metadata")

		return err
	}

	return nil
}

// UpdateOperationRetryFlag updates the operation retry flag.
// nolint: gocognit
func (ctl *controller) UpdateOperationRetryFlag(nCtx contextx.IContext, mode operation.RetryMode, operationID ...string) error {
	operations, _, err := ctl.mgr.stgOperation.ListOperationByOperationID(nCtx, operationID...)
	if err != nil {
		return err
	}

	for _, oper := range operations {
		if err := oper.CheckEnforceability(); err != nil {
			return err
		}
	}

	gp := gopool.NewPool()
	for _, oper := range operations {
		tempOper := oper
		fn := func() error {
			lastInstID := tempOper.GetLastInstanceID()
			if lastInstID == "" {
				return fmt.Errorf("operation has no instance, can not retry, operation-id(%s)", tempOper.OperationID)
			}

			lastInst, err := ctl.mgr.stgOperationInstance.GetOperationInstanceBriefData(nCtx, lastInstID)
			if err != nil {
				return err
			}

			if mode == operation.RetryModePartial && lastInst.Lifecycle.IsTerminated() {
				return fmt.Errorf("failed to partial retry operation-id(%s), last operation instance is terminated", tempOper.OperationID)
			}

			if tempOper.GetLastRetryFlag() != nil && tempOper.GetLastRetryFlag().RetryInstanceID == "" {
				logger.G.Biz(nCtx).With("operation-id", tempOper.OperationID).With("last retry has not been processed, activate trigger directly")

				return nil
			}

			tempOper.RetryFlags = append(tempOper.RetryFlags, operation.RetryFlag{
				Mode:             mode,
				SourceInstanceID: tempOper.GetLastInstanceID(),
				RetryInstanceID:  "",
			})

			if err := ctl.mgr.stgOperation.UpsertOperation(nCtx, tempOper); err != nil {
				return err
			}

			return nil
		}

		gp.Go(fn)
	}

	if err := gp.Wait(); err != nil {
		return err
	}

	return nil
}

// ListNeedInstantiateOperation returns the operations need to be instantiated.
func (ctl *controller) ListNeedInstantiateOperation(nCtx contextx.IContext, page types.Page) ([]IOperationCtl, error) {
	opers, _, err := ctl.mgr.stgOperation.ListNeedInstantiateOperationByTriggerID(nCtx, page, ctl.trig.TriggerID)
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

// GetLastOperationInstance returns the last operation instance.
func (ctl *controller) GetLastOperationInstance(nCtx contextx.IContext) (IOperationInstanceCtl, error) {
	lastInstanceID := ctl.oper.GetLastInstanceID()
	if lastInstanceID == "" {
		return nil, fmt.Errorf("operation has no instance, operation-id(%s)", ctl.oper.OperationID)
	}

	return ctl.GetOperationInstance(nCtx, lastInstanceID)
}

// ListOperationInstances returns the operation instances with states.
func (ctl *controller) ListOperationInstances(nCtx contextx.IContext, page types.Page, states ...operation.State) ([]IOperationInstanceCtl, error) {
	condition := &types.OperInstDataCondition{
		ExactInclude: &types.OperInstDataExactFields{
			TriggerID: []string{ctl.trig.TriggerID},
			State:     states,
		},
	}
	instanceBriefData, _, err := ctl.mgr.stgOperationInstance.ListOperationInstanceBriefDataWithoutActionInst(
		nCtx, page, condition)
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
func (ctl *controller) CreateOperationInstance(nCtx contextx.IContext) (IOperationInstanceCtl, error) {
	if ctl.oper == nil {
		return nil, errors.New("operation is nil")
	}

	var (
		instanceData *operation.InstanceData
		err          error
	)
	if ctl.oper.GetLastRetryFlag() != nil {
		instanceData, err = ctl.generateRetryOperationInstance(nCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to generate retry operation instance: %w", err)
		}
	} else {
		instanceData, err = ctl.generateOperationInstance(nCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to generate operation instance: %w", err)
		}
	}

	return &controller{
		mgr:                   ctl.mgr,
		trig:                  ctl.trig,
		oper:                  ctl.oper,
		operInstanceBriefData: &instanceData.InstanceBriefData,
	}, nil
}

// GetOperationInstance returns the operation instance.
func (ctl *controller) GetOperationInstance(nCtx contextx.IContext, operationInstanceID string) (IOperationInstanceCtl, error) {
	instanceBriefData, err := ctl.mgr.stgOperationInstance.GetOperationInstanceBriefData(nCtx, operationInstanceID)
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
// nolint: funlen,nonamedreturns
func (ctl *controller) LaunchOperationInstance(nCtx contextx.IContext) (err error) {
	// record metric.
	m := metric.NewOperationInstanceLaunch(ctl.operInstanceBriefData).Start()
	defer m.End(err)

	logger.G.Sys().With("oper-inst-id", ctl.operInstanceBriefData.Metadata.OperationInstanceID).Info("try to launch operation instance")

	if ctl.operInstanceBriefData.Lifecycle.State != operation.StateInit {
		return fmt.Errorf("operation instance state(%s) is not init", ctl.operInstanceBriefData.Lifecycle.State)
	}

	// ensure operation is not nil
	if ctl.oper == nil {
		if ctl.oper, err = ctl.mgr.stgOperation.GetOperation(nCtx, ctl.operInstanceBriefData.Metadata.OperationID); err != nil {
			return err
		}
	}

	actionNames := ctl.oper.Definition.ActionDefNames()
	if len(actionNames) == 0 {
		return fmt.Errorf("operation-instance-id(%s) has no action", ctl.operInstanceBriefData.Metadata.OperationInstanceID)
	}

	tracer := ctl.mgr.traceSvc.TracerProvider().Tracer(scopeNameOperationInstance)
	traceCtx, span := tracer.Start(nCtx, fmt.Sprintf("%s %s", scopeNamePrefixOperationInstance, ctl.operInstanceBriefData.Metadata.OperationDefName),
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String(attributeKeyTriggerID, ctl.operInstanceBriefData.Metadata.TriggerID),
			attribute.String(attributeKeyOperationID, ctl.operInstanceBriefData.Metadata.OperationID),
			attribute.String(attributeKeyOperationDefName, ctl.operInstanceBriefData.Metadata.OperationDefName),
			attribute.String(attributeKeyOperationInstanceID, ctl.operInstanceBriefData.Metadata.OperationInstanceID),
		),
	)
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, fmt.Sprintf("launch operation instance failed: %s", err))
			span.RecordError(err)
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

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
				{
					Name:  "trace-id",
					Type:  "string",
					Value: span.SpanContext().TraceID().String(),
				},
				{
					Name:  "span-id",
					Type:  "string",
					Value: span.SpanContext().SpanID().String(),
				},
			},
		}
	}

	logger.G.Sys().With("oper-inst-id", ctl.operInstanceBriefData.Metadata.OperationInstanceID).Info("send chain to machinery")

	// update state.
	ctl.operInstanceBriefData.Lifecycle.Launch()
	if err := ctl.mgr.stgOperationInstance.UpdateOperationInstanceLifecycle(
		nCtx, ctl.operInstanceBriefData.Metadata.OperationInstanceID, ctl.operInstanceBriefData.Lifecycle); err != nil {
		logger.G.Sys().
			WithErr(err).
			With("oper-inst-id", ctl.operInstanceBriefData.Metadata.OperationInstanceID).
			Error("failed to update operation instance lifecycle")

		return err
	}

	chain, err := tasks.NewChain(signatures...)
	if err != nil {
		return fmt.Errorf("failed to create chain: %w", err)
	}

	if _, err = ctl.mgr.server.SendChainWithContext(traceCtx, chain); err != nil {
		ctl.operInstanceBriefData.Lifecycle.EndByActionState(action.StateFailed)
		if updateErr := ctl.mgr.stgOperationInstance.UpdateOperationInstanceLifecycle(
			nCtx, ctl.operInstanceBriefData.Metadata.OperationInstanceID, ctl.operInstanceBriefData.Lifecycle); updateErr != nil {
			logger.G.Sys().
				WithErr(updateErr).
				With("oper-inst-id", ctl.operInstanceBriefData.Metadata.OperationInstanceID).
				Error("failed to fail operation instance lifecycle after send chain failure")
		}
		if updateErr := ctl.mgr.stgOperation.UpdateOperationLatestInstBriefData(
			nCtx, ctl.operInstanceBriefData.Metadata.OperationID, ctl.operInstanceBriefData); updateErr != nil {
			logger.G.Sys().
				WithErr(updateErr).
				With("oper-id", ctl.operInstanceBriefData.Metadata.OperationID,
					"oper-inst-id", ctl.operInstanceBriefData.Metadata.OperationInstanceID).
				Error("failed to update operation latest brief data after send chain failure")
		}

		return fmt.Errorf("failed to send chain to machinery: %w", err)
	}

	return nil
}

// TerminateOperationInstance terminates the operation instance.
func (ctl *controller) TerminateOperationInstance(nCtx contextx.IContext) error {
	if operation.CheckStateFinished(ctl.operInstanceBriefData.Lifecycle.State) {
		logger.G.Sys().With("oper-inst-id", ctl.operInstanceBriefData.Metadata.OperationInstanceID).
			Info("operation instance already stop, skip terminate")

		return nil
	}

	if err := ctl.mgr.stgOperationInstance.UpsertNeedStopOperInst(nCtx, ctl.operInstanceBriefData.Metadata.OperationInstanceID); err != nil {
		return fmt.Errorf("failed to upsert need stop oper inst: %w", err)
	}

	logger.G.Sys().With("oper-inst-id", ctl.operInstanceBriefData.Metadata.OperationInstanceID).
		Info("operation instance terminated")

	return nil
}

func (ctl *controller) generateOperationInstance(nCtx contextx.IContext) (*operation.InstanceData, error) {
	inst, err := ctl.createOperationInstanceBase(nCtx, nil)
	if err != nil {
		return nil, err
	}

	if err := ctl.mgr.stgOperationInstance.UpsertOperationInstanceData(nCtx, inst); err != nil {
		return nil, err
	}

	ctl.oper.InstanceIDs = append(ctl.oper.InstanceIDs, inst.Metadata.OperationInstanceID)
	if len(ctl.oper.InstanceIDs) > operation.MaxInstanceNum {
		ctl.oper.InstanceIDs = ctl.oper.InstanceIDs[len(ctl.oper.InstanceIDs)-operation.MaxInstanceNum:]
	}
	ctl.oper.LatestInstBriefData = &inst.InstanceBriefData
	if err := ctl.mgr.stgOperation.UpsertOperation(nCtx, ctl.oper); err != nil {
		return nil, err
	}

	logger.G.Sys().With("oper-id", ctl.oper.OperationID,
		"oper-inst-id", inst.Metadata.OperationInstanceID).
		Debug("created operation instance")

	return inst, nil
}

func (ctl *controller) generateRetryOperationInstance(nCtx contextx.IContext) (*operation.InstanceData, error) {
	lastRetryFlag := ctl.oper.GetLastRetryFlag()
	if lastRetryFlag == nil {
		return nil, errors.New("no retry flag found for operation")
	}

	var (
		retryInstance *operation.InstanceData
		err           error
	)

	switch lastRetryFlag.Mode {
	case operation.RetryModeAll:
		retryInstance, err = ctl.handleFullRetry(nCtx)
		if err != nil {
			return nil, err
		}

	case operation.RetryModePartial:
		retryInstance, err = ctl.handlePartialRetry(nCtx)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unknown retry mode: %v", lastRetryFlag.Mode)
	}

	if err := ctl.mgr.stgOperationInstance.UpsertOperationInstanceData(nCtx, retryInstance); err != nil {
		return nil, err
	}

	ctl.oper.InstanceIDs = append(ctl.oper.InstanceIDs, retryInstance.Metadata.OperationInstanceID)
	if len(ctl.oper.InstanceIDs) > operation.MaxInstanceNum {
		ctl.oper.InstanceIDs = ctl.oper.InstanceIDs[len(ctl.oper.InstanceIDs)-operation.MaxInstanceNum:]
	}
	ctl.oper.RetryFlags[len(ctl.oper.RetryFlags)-1].RetryInstanceID = retryInstance.Metadata.OperationInstanceID
	ctl.oper.LatestInstBriefData = &retryInstance.InstanceBriefData
	if err := ctl.mgr.stgOperation.UpsertOperation(nCtx, ctl.oper); err != nil {
		return nil, err
	}

	logger.G.Sys().With("oper-id", ctl.oper.OperationID,
		"retry-mode", lastRetryFlag.Mode,
		"source-inst-id", lastRetryFlag.SourceInstanceID,
		"retry-inst-id", retryInstance.Metadata.OperationInstanceID).
		Debug("created retry operation instance")

	return retryInstance, nil
}

func (ctl *controller) handlePartialRetry(nCtx contextx.IContext) (*operation.InstanceData, error) {
	if len(ctl.oper.Param.RetryStartPoint) == 0 {
		return nil, errors.New("retry start point is empty")
	}

	if len(ctl.oper.Definition.ActionDefNames()) == 0 {
		return nil, errors.New("operation has no actions")
	}

	prevInstanceID := ctl.oper.RetryFlags[len(ctl.oper.RetryFlags)-1].SourceInstanceID
	prevInstance, err := ctl.mgr.stgOperationInstance.GetOperationInstanceFullData(nCtx, prevInstanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get previous instance(%s): %w", prevInstanceID, err)
	}

	startIndex, err := ctl.findPartialRetryStartIndex(prevInstance)
	if err != nil {
		return nil, err
	}

	return ctl.createPartialRetryInstance(nCtx, prevInstance, startIndex)
}

func (ctl *controller) findPartialRetryStartIndex(prevInstance *operation.InstanceData) (int, error) {
	if len(ctl.oper.InstanceIDs) == 0 {
		return -1, errors.New("operation has no instance")
	}

	actionNames := ctl.oper.Definition.ActionDefNames()
	var failed *action.InstanceData
	for _, name := range actionNames {
		inst, ok := prevInstance.ActionInstanceDataMap[name]
		if !ok {
			return -1, fmt.Errorf("previous instance missing action(%s)", name)
		}

		if inst.Lifecycle.State == action.StateTerminated {
			return -1, fmt.Errorf("previous instance action(%s) is terminated, can not partial retry", name)
		}

		if inst.Lifecycle.IsFailed() {
			failed = inst
			break
		}
	}
	if failed == nil {
		// operation-level failure (e.g. operation extra execution failed before/after actions)
		// leaves no failed action anchor, fall back to retrying the operation from the beginning.
		if prevInstance.Lifecycle.IsFailed() {
			logger.G.Sys().With("oper-inst-id", prevInstance.Metadata.OperationInstanceID).
				Info("operation failed at operation level, no failed action found, retry from beginning")

			return 0, nil
		}

		return -1, errors.New("all actions succeeded, nothing to retry")
	}

	retryStartPoints := ctl.oper.Param.RetryStartPoint
	for idx := failed.Index; idx >= 0; idx-- {
		name := actionNames[idx]
		if retryStartPoints[name] {
			logger.G.Sys().With("retry-action", name, "failed-action", failed.Name).Info("partial retry start action decided")
			return idx, nil
		}
	}

	return -1, fmt.Errorf("no valid retry start point found for failed action(%s)", failed.Name)
}

func (ctl *controller) createPartialRetryInstance(
	nCtx contextx.IContext, prevInstance *operation.InstanceData, startIndex int) (*operation.InstanceData, error) {

	if startIndex < 0 {
		return nil, errors.New("invalid start index for partial retry")
	}

	stateDecider := func(actionName string) action.State {
		actionInst, ok := prevInstance.ActionInstanceDataMap[actionName]
		if !ok || actionInst.Index < startIndex {
			return action.StateSkipped
		}

		return action.StatePending
	}

	instanceData, err := ctl.createOperationInstanceBase(nCtx, stateDecider)
	if err != nil {
		return nil, err
	}

	if startIndex > 0 {
		actionNames := ctl.oper.Definition.ActionDefNames()
		lastActionBeforeStart := actionNames[startIndex-1]

		// in the same operation, last operation instance's action should be exist in new instance
		prevActInst, okPrev := prevInstance.ActionInstanceDataMap[lastActionBeforeStart]
		newActInst, okNew := instanceData.ActionInstanceDataMap[lastActionBeforeStart]
		if okPrev && okNew {
			newActInst.Content = maps.Clone(prevActInst.Content)
		}
	}

	return instanceData, nil
}

func (ctl *controller) handleFullRetry(nCtx contextx.IContext) (*operation.InstanceData, error) {
	return ctl.createOperationInstanceBase(nCtx, nil)
}

func (ctl *controller) createOperationInstanceBase(_ contextx.IContext, stateDecider func(string) action.State) (*operation.InstanceData, error) {
	operationInstanceID := identifier.GenOperationInstanceID()
	actionNames := ctl.oper.Definition.ActionDefNames()

	actionInstanceDataMap := make(map[string]*action.InstanceData)
	for index, actionName := range actionNames {
		state := action.StatePending
		if stateDecider != nil {
			state = stateDecider(actionName)
		}

		// get action Definition for display names
		actionDef, ok := ctl.mgr.registeredActionDefs[actionName]
		if !ok {
			// record metric.
			metric.ActionNotRegistered(actionName)

			return nil, fmt.Errorf("action not registered, name(%s)", actionName)
		}

		actionInstanceDataMap[actionName] = &action.InstanceData{
			TriggerID:           ctl.trig.TriggerID,
			OperationID:         ctl.oper.OperationID,
			OperationDefName:    ctl.oper.Definition.Name(),
			OperationInstanceID: operationInstanceID,
			Name:                actionName,
			DisplayNameZh:       actionDef.DisplayNameZh(),
			DisplayNameEn:       actionDef.DisplayNameEn(),
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

	return &operation.InstanceData{
		InstanceBriefData: operation.InstanceBriefData{
			Metadata: &operation.InstanceMetadata{
				TriggerID:              ctl.trig.TriggerID,
				OperationInstanceID:    operationInstanceID,
				OperationDefName:       ctl.oper.Definition.Name(),
				OperationID:            ctl.oper.OperationID,
				ActionNames:            actionNames,
				ParentOperationID:      ctl.oper.Param.ParentOperationID,
				ParentOperInstID:       ctl.oper.Param.ParentOperInstID,
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
	}, nil
}
