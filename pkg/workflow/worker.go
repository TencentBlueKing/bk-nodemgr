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
	"runtime/debug"
	"time"

	"github.com/RichardKnop/machinery/v2/tasks"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/common"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/metric"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// IWorker describes the workflow worker.
type IWorker interface {
}

const (
	consumerTag         = ""
	engineMaxRetryLimit = uint(10)
)

func (mgr *manager) launchWorker() error {
	if !mgr.isRunning {
		return errors.New("operation instance manager is not running")
	}

	for _, actionDef := range mgr.registeredActionDefs {
		if err := mgr.server.RegisterTask(actionDef.Name(), mgr.do); err != nil {
			return err
		}
	}

	mgr.worker = mgr.server.NewWorker(consumerTag, mgr.WorkerNum)

	mgr.worker.SetErrorHandler(func(err error) {
		logger.G.Sys().WithErr(err).Error("worker error")
	})
	mgr.worker.SetPreTaskHandler(func(_ *tasks.Signature) {})
	mgr.worker.SetPostTaskHandler(func(_ *tasks.Signature) {})

	mgr.isConsuming = true
	go func() {
		mgr.launchWorkerErr <- mgr.worker.Launch()
		mgr.isConsuming = false
	}()

	return nil
}

// do executes the action defined by actionName for the operation instance with operationInstanceID.
// this func only accept context.Context as input, so we accept context.Context and then change it to contextx.IContext.
//
// nolint: funlen,gocognit,cyclop,gocyclo,lll,fnsize
func (mgr *manager) do(ctx context.Context, actionName string, operationInstanceID string, traceID string, spanID string) error {
	tid, err := trace.TraceIDFromHex(traceID)
	if err != nil {
		logger.G.Sys().With("trace-id", traceID).WithErr(err).Error("get invalid trace id")
	}

	sid, err := trace.SpanIDFromHex(spanID)
	if err != nil {
		logger.G.Sys().With("trace-id", traceID).WithErr(err).Error("get invalid span id")
	}

	spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    tid,
		SpanID:     sid,
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	})
	ctx = trace.ContextWithSpanContext(ctx, spanCtx)

	tracer := mgr.traceSvc.TracerProvider().Tracer(scopeNameAction)
	ctx, span := tracer.Start(ctx, fmt.Sprintf("%s %s", scopeNamePrefixAction, actionName),
		trace.WithAttributes(
			attribute.String(attributeKeyActionName, actionName),
			attribute.String(attributeKeyOperationInstanceID, operationInstanceID)),
		trace.WithSpanKind(trace.SpanKindConsumer),
	)
	defer span.End()

	actionDef, ok := mgr.registeredActionDefs[actionName]
	if !ok {
		// record metric.
		metric.ActionNotRegistered(actionName)

		return fmt.Errorf("action not registered, name(%s)", actionName)
	}

	nCtx := contextx.New(ctx, contextx.WithMessageID(actionMessageID(operationInstanceID, actionName)))

	// get action instance.
	actionInstData, err := mgr.stgActionInstance.GetActionInstanceData(nCtx, operationInstanceID, actionName)
	if err != nil {
		// record metric.
		metric.ActionDataNotFound(actionName)

		return fmt.Errorf("failed to get action instance data from operation instance. "+
			"oper-inst-id(%s), action-name(%s): %v", operationInstanceID, actionName, err)
	}

	// record metric.
	m := metric.NewActionProcess(actionInstData).Start()
	defer m.End(actionInstData.Lifecycle)

	// check if this action should be executed.
	if err := actionInstData.NeedExecuted(); err != nil {
		if errors.Is(err, common.ErrActionAlreadySucceeded()) || errors.Is(err, common.ErrActionSkipped()) {
			return nil
		}

		return err
	}

	// get operation instance.
	operInstBriefData, err := mgr.stgOperationInstance.GetOperationInstanceBriefData(nCtx, operationInstanceID)
	if err != nil {
		return fmt.Errorf("failed to get operation instance brief data. "+
			"oper-inst-id(%s): %v", operationInstanceID, err)
	}

	// handle extra execution before action executed. if retry happens, action maybe not first
	if execErr := mgr.doOperExtraExecution(nCtx, operInstBriefData); execErr != nil {
		operInstBriefData.Lifecycle.End(action.StateFailed)
		err = mgr.updateOperationInstanceLifecycle(nCtx, operationInstanceID, operInstBriefData.Lifecycle)
		if err != nil {
			return fmt.Errorf("update operation instance lifecycle failed: %w, start execution failed: %w",
				err, execErr)
		}

		return fmt.Errorf("do oper-inst-id(%s) starting extra execution failed: %w", operationInstanceID, execErr)
	}

	// handle operation instance lifecycle.
	if !operInstBriefData.Lifecycle.IsRunning() {
		operInstBriefData.Lifecycle.Start()
		if err = mgr.updateOperationInstanceLifecycle(nCtx, operationInstanceID, operInstBriefData.Lifecycle); err != nil {
			return err
		}
	}

	// handle action instance lifecycle.
	actionInstData.Lifecycle.Start()
	if err = mgr.updateActionLifecycle(nCtx, operationInstanceID, actionName, actionInstData.Lifecycle); err != nil {
		return err
	}

	// handle action content.
	if actionInstData.IsFirst() {
		// first action should get content from operation instance init data.
		actionInstData.Content = operInstBriefData.Metadata.InitContent
	} else {
		// not-first action should get content from previous action.
		preActionName := operInstBriefData.Metadata.ActionNames[actionInstData.Index-1]
		preActionInstData, err := mgr.stgActionInstance.GetActionInstanceData(nCtx, operationInstanceID, preActionName)
		if err != nil {
			return fmt.Errorf("failed to get pre action instance data from operation instance. "+
				"oper-inst-id(%s), action-name(%s): %v", operationInstanceID, preActionName, err)
		}

		actionInstData.Content = preActionInstData.Content
	}

	// execute and wait for action done.
	// nolint: contextcheck
	executeErr := mgr.executeAndWatchAction(nCtx, actionDef, operInstBriefData, actionInstData)

	// updates action instance lifecycle.
	if err = mgr.updateActionLifecycle(nCtx, operationInstanceID, actionName, actionInstData.Lifecycle); err != nil {
		return err
	}

	// updates action content.
	if err = mgr.updateActionContent(nCtx, operationInstanceID, actionName, actionInstData.Content); err != nil {
		return err
	}

	// updates action instance private data.
	if err = mgr.updateOperationInstancePrivateData(
		nCtx, operationInstanceID, actionName, actionInstData.PrivateData); err != nil {
		return err
	}

	// when action done or error happens, we need to update the state of the operation instance.
	if actionInstData.IsLast() || executeErr != nil {
		// record oper inst metric.
		defer metric.OperationInstanceProcessed(operInstBriefData)

		operInstBriefData.Lifecycle.End(actionInstData.Lifecycle.State)
		if err = mgr.updateOperationInstanceLifecycle(nCtx, operationInstanceID, operInstBriefData.Lifecycle); err != nil {
			return fmt.Errorf("failed to update operation instance lifecycle. err: %v, execution-error(%v)",
				err, executeErr)
		}

		logger.G.Sys().With("oper-inst-id", operationInstanceID, "lifecycle", operInstBriefData.Lifecycle).
			Info("updated operation instance lifecycle with terminated state")

		if err = mgr.doOperExtraExecution(nCtx, operInstBriefData); err != nil {
			return fmt.Errorf("do oper-inst-id(%s) ending extra execution failed: %w", operationInstanceID, err)
		}
	}

	return executeErr
}

func (mgr *manager) updateActionLifecycle(
	ctx contextx.IContext,
	operationInstanceID,
	actionName string,
	actionInstLifecycle *action.Lifecycle) error {

	if err := mgr.stgActionInstance.UpdateActionInstanceLifecycle(ctx,
		operationInstanceID, actionName, actionInstLifecycle); err != nil {
		return fmt.Errorf("failed to update action instance lifecycle. "+
			"oper-inst-id(%s), action-name(%s): %v", operationInstanceID, actionName, err)
	}

	return nil
}

func (mgr *manager) updateActionContent(
	ctx contextx.IContext,
	operationInstanceID,
	actionName string,
	content map[string]any) error {

	if err := mgr.stgActionInstance.UpdateActionInstanceContent(ctx,
		operationInstanceID, actionName, content); err != nil {
		return fmt.Errorf("failed to update action instance content. "+
			"oper-inst-id(%s), action-name(%s): %v", operationInstanceID, actionName, err)
	}

	return nil
}

func (mgr *manager) updateOperationInstanceLifecycle(
	ctx contextx.IContext,
	operationInstanceID string,
	operInstLifecycle *operation.Lifecycle) error {

	if err := mgr.stgOperationInstance.UpdateOperationInstanceLifecycle(ctx,
		operationInstanceID, operInstLifecycle); err != nil {
		return fmt.Errorf("failed to update operation instance lifecycle. "+
			"oper-inst-id(%s): %v", operationInstanceID, err)
	}

	return nil
}

func (mgr *manager) updateOperationInstancePrivateData(
	ctx contextx.IContext,
	operationInstanceID string,
	actionName string,
	privateData map[string]any) error {

	if len(privateData) == 0 {
		// no private data to update, skip.
		logger.G.Sys().
			With("oper-inst-id", operationInstanceID, "action-name", actionName, "private-data", privateData).
			Debug("no private data to update")

		return nil
	}

	if err := mgr.stgActionInstance.UpsertActionInstancePrivateData(ctx,
		operationInstanceID, actionName, privateData); err != nil {
		return fmt.Errorf("failed to upsert action instance private data. "+
			"oper-inst-id(%s), action-name(%s), private-data(%v): %v",
			operationInstanceID, actionName, privateData, err)
	}

	return nil
}

func (mgr *manager) executeAndWatchAction(nCtx contextx.IContext,
	actionDef action.Definition,
	operInstBriefData *operation.InstanceBriefData,
	actionInstData *action.InstanceData) error {

	actionTimeoutCtx, actionTimeoutCancel := contextx.WithTimeout(contextx.From(nCtx), actionDef.Timeout())
	defer actionTimeoutCancel()

	operInstTimeoutCtx, operInstTimeoutCancel := context.WithDeadline(
		nCtx, operInstBriefData.Lifecycle.StartedAt.Add(operInstBriefData.Metadata.Timeout))
	defer operInstTimeoutCancel()

	// watch storage for stopping event.
	terminatingC := mgr.stgOperationInstance.WatchOperInstStopping(actionTimeoutCtx, operInstBriefData.Metadata.OperationInstanceID)

	doResult := make(chan error, 1)
	actionInstCtx := &action.InstanceContext{
		Ctx:  actionTimeoutCtx,
		Data: actionInstData,
	}

	go mgr.executeAction(doResult, actionInstCtx, actionDef) // nolint:contextcheck

	select {
	case err := <-doResult:
		{
			actionInstData.Lifecycle.EndWithErr(err)

			return err
		}

	case <-actionTimeoutCtx.Done():
		{
			actionInstData.Lifecycle.EndWithTimeout()

			return fmt.Errorf("action timeout, oper-inst-id(%s), action-name(%s)",
				operInstBriefData.Metadata.OperationID, actionInstData.Name)
		}

	case <-operInstTimeoutCtx.Done():
		{
			actionInstData.Lifecycle.EndWithTimeout()

			return fmt.Errorf("operation instance timeout, oper-inst-id(%s), action-name(%s)",
				operInstBriefData.Metadata.OperationID, actionInstData.Name)
		}

	case <-terminatingC:
		{
			actionInstData.Lifecycle.EndWithTerminated()

			return fmt.Errorf("operation instance has been terminated, oper-inst-id(%s), action-name(%s)",
				operInstBriefData.Metadata.OperationID, actionInstData.Name)
		}

	case <-nCtx.Done():
		{
			// TODO: 考虑关闭 worker 时，worker 退出时，action 未完成，如何处理
			actionInstData.Lifecycle.EndWithTerminated()

			return fmt.Errorf("operation operInstMgr context done, oper-inst-id(%s), action-name(%s)",
				operInstBriefData.Metadata.OperationID, actionInstData.Name)
		}
	}
}

func (mgr *manager) executeAction(
	doResult chan error, actionInstCtx *action.InstanceContext, actionDef action.Definition) {

	var err error

	defer func() {
		if r := recover(); r != nil {
			err = errors.New("action panic")

			stack := string(debug.Stack())

			logger.G.Sys().
				WithErr(err).
				With("info", actionInstCtx.Data.Info(), "recover", r, "stack", stack).
				Info("failed to execute action, recover from panic")

			actionInstCtx.Data.LogE(fmt.Sprintf("action panic: revoer(%v), stack(%s)", r, stack))
		}

		doResult <- err
	}()

	ctx, cancel := contextx.WithCancel(contextx.From(actionInstCtx.Ctx))
	defer cancel()

	go mgr.autoRefreshActionDataMsg(ctx, actionInstCtx.Data)

	err = mgr.callActionDefWithRetry(actionInstCtx, actionDef)
}

func (mgr *manager) autoRefreshActionDataMsg(ctx contextx.IContext, data *action.InstanceData) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	idx := 0

	// refresh action inst data messages to storage
	for {
		select {
		case <-ctx.Done():
			// when finished, push the rest messages to storage.
			msgs := data.Messages[idx:]

			// need to make sure the db operation done, so in this way we use mgr.ctx instead of ctx.
			// nolint: contextcheck
			err := mgr.stgActionInstance.PushActionInstanceMessage(mgr.ctx,
				data.OperationInstanceID,
				data.Name,
				msgs...)
			if err != nil {
				logger.G.Sys().WithErr(err).With("action", data.Name).Error("failed to refresh action inst data messages")
			}

			return

		case <-ticker.C:
			msgs := data.Messages[idx:]
			idx += len(msgs)

			// need to make sure the db operation done, so in this way we use mgr.ctx instead of ctx.
			// nolint: contextcheck
			err := mgr.stgActionInstance.PushActionInstanceMessage(
				mgr.ctx,
				data.OperationInstanceID,
				data.Name,
				msgs...)
			if err != nil {
				logger.G.Sys().WithErr(err).With("action", data.Name).Error("failed to refresh action inst data messages")
			}

			continue
		}
	}
}

// callActionDefWithRetry do action with retry.
func (mgr *manager) callActionDefWithRetry(actionInstCtx *action.InstanceContext, actionDef action.Definition) error {
	var doErr error

	for retryNum := uint(0); retryNum <= actionDef.MaxRetryCount() && retryNum < engineMaxRetryLimit; retryNum++ {
		logger.G.Sys().
			With("operation", actionInstCtx.Data.OperationDefName).
			With("oper-inst-id", actionInstCtx.Data.OperationInstanceID, "action", actionInstCtx.Data.Name, "retry", retryNum).
			Info("started action")

		actionInstCtx.Data.LogI(fmt.Sprintf("started action, action-name(%s), retry-num(%d)",
			actionInstCtx.Data.Name, retryNum))

		doErr = actionDef.Do(actionInstCtx)

		if doErr != nil {
			logger.G.Sys().
				WithErr(doErr).
				With("operation", actionInstCtx.Data.OperationDefName).
				With("oper-inst-id", actionInstCtx.Data.OperationInstanceID, "action", actionInstCtx.Data.Name, "retry", retryNum).
				Error("failed to do action")

			actionInstCtx.Data.LogW(fmt.Sprintf("failed to do action, action-name(%s), retry-num(%d): %v",
				actionInstCtx.Data.Name, retryNum, doErr))

			delayFn := actionDef.DelayFn()
			if delayFn != nil {
				delayFn()
			}

			continue
		}

		logger.G.Sys().
			With("operation", actionInstCtx.Data.OperationDefName).
			With("oper-inst-id", actionInstCtx.Data.OperationInstanceID, "action", actionInstCtx.Data.Name, "retry", retryNum).
			Info("done action")

		actionInstCtx.Data.LogI(fmt.Sprintf("done action, action-name(%s), oper-def-name(%s), retry-num(%d)",
			actionInstCtx.Data.Name, actionInstCtx.Data.OperationDefName, retryNum))

		break
	}

	if doErr != nil {
		logger.G.Sys().
			WithErr(doErr).
			With("operation", actionInstCtx.Data.OperationDefName).
			With("oper-inst-id", actionInstCtx.Data.OperationInstanceID, "action", actionInstCtx.Data.Name).
			Error("failed to do action with all attempts")

		actionInstCtx.Data.LogE(fmt.Sprintf("action failed, action-name(%s), oper-def-name(%s), err(%v)",
			actionInstCtx.Data.Name, actionInstCtx.Data.OperationDefName, doErr))
	}

	return doErr
}

// doOperExtraExecution executes the extra action for the operation instance.
func (mgr *manager) doOperExtraExecution(ctx contextx.IContext, oper *operation.InstanceBriefData) error {
	if oper.Metadata.ExtraExecutionName == "" {
		return nil
	}

	if oper.Lifecycle.State == operation.StateInit || oper.Lifecycle.State == operation.StateRunning {
		return nil
	}

	actionDef, ok := mgr.registeredOperExtraExecutionDefs[oper.Metadata.ExtraExecutionName]
	if !ok {
		return fmt.Errorf("extra action not registered, name(%s)", oper.Metadata.ExtraExecutionName)
	}

	msgIdx := len(oper.Metadata.ExtraExecutionMessages)
	err := actionDef.Do(ctx, oper)
	updateErr := mgr.stgOperationInstance.UpdateOperationInstanceExtraExecutionMessages(
		ctx,
		oper.Metadata.OperationInstanceID,
		oper.Metadata.ExtraExecutionMessages[msgIdx:]...)
	if updateErr != nil {
		logger.G.Sys().
			WithErr(updateErr).
			With("operation-extra-execution", oper.Metadata.ExtraExecutionName).
			Error("failed to refresh operation extra execution message")
	}

	return err
}

func actionMessageID(operInstID, actionName string) string {
	return fmt.Sprintf("%s|%s", operInstID, actionName)
}
