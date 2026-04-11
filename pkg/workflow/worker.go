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

	// get operation instance.
	operInstBriefData, err := mgr.stgOperationInstance.GetOperationInstanceBriefData(nCtx, operationInstanceID)
	if err != nil {
		logger.G.Sys().Ctx(nCtx).WithErr(err).With("oper-inst-id", operationInstanceID).Error("failed to get operation instance brief data")

		return fmt.Errorf("failed to get operation instance brief data. "+
			"oper-inst-id(%s): %v", operationInstanceID, err)
	}

	// nolint: contextcheck
	refreshMsgCtx, refreshMsgCancel := contextx.WithCancel(contextx.From(nCtx))
	defer refreshMsgCancel()
	go mgr.autoRefreshExtraExecutionMsg(refreshMsgCtx, operInstBriefData)

	if execErr := mgr.doOperExtraExecution(nCtx, operInstBriefData); execErr != nil {
		logger.G.Sys().Ctx(nCtx).WithErr(execErr).
			With("oper-inst-id", operationInstanceID).
			Error("failed to do operation instance extra execution before do action")

		return fmt.Errorf("failed to do operation instance extra execution. oper-inst-id(%s): %v", operationInstanceID, execErr)
	}

	// nolint: contextcheck
	actionErr := mgr.doAction(nCtx, actionDef, operInstBriefData)

	// when action finished, we also need to execute the extra execution if needed, and update the operation instance state.
	if execErr := mgr.doOperExtraExecution(nCtx, operInstBriefData); execErr != nil {
		logger.G.Sys().Ctx(nCtx).WithErr(execErr).
			With("oper-inst-id", operationInstanceID).
			Error("failed to do operation instance extra execution after do action")

		if actionErr != nil {
			bothErr := fmt.Errorf("do action error: %v; after action finished do extra execution error: %v", actionErr, execErr)
			logger.G.Sys().Ctx(nCtx).WithErr(bothErr).
				With("oper-inst-id", operationInstanceID).
				Error("failed to do action and operation instance extra execution after do action")

			return fmt.Errorf("failed to do extra execution after do action.oper-inst-id(%s): both error: %v", operationInstanceID, bothErr)
		}

		return fmt.Errorf("failed to do operation instance extra execution after action execution. oper-inst-id(%s): %v",
			operationInstanceID, execErr)
	}

	return actionErr
}

// nolint: funlen,gocognit,cyclop,gocyclo,lll
func (mgr *manager) doAction(nCtx contextx.IContext, actionDef action.Definition, operInstBriefData *operation.InstanceBriefData) error {
	operationInstanceID := operInstBriefData.Metadata.OperationInstanceID

	// get action instance.
	actionInstData, err := mgr.stgActionInstance.GetActionInstanceData(nCtx, operationInstanceID, actionDef.Name())
	if err != nil {
		// record metric.
		metric.ActionDataNotFound(actionDef.Name())

		operInstBriefData.Lifecycle.End(action.StateFailed)
		if refreshErr := mgr.refreshOperationInstanceState(nCtx, operInstBriefData); refreshErr != nil {
			logger.G.Sys().Ctx(nCtx).WithErr(refreshErr).With(
				"oper-inst-id", operationInstanceID, "action-name", actionDef.Name()).
				Error("failed to refresh operation instance state after failed to get action instance data")

			return fmt.Errorf("failed to get action instance data from operation instance. "+
				"oper-inst-id(%s), action-name(%s): %w; additionally failed to refresh operation instance state: %v",
				operationInstanceID, actionDef.Name(), err, refreshErr)
		}

		return fmt.Errorf("failed to get action instance data from operation instance. "+
			"oper-inst-id(%s), action-name(%s): %w", operationInstanceID, actionDef.Name(), err)
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

	// handle operation instance lifecycle.
	if !operInstBriefData.Lifecycle.IsRunning() {
		operInstBriefData.Lifecycle.Start()
		if refreshErr := mgr.refreshOperationInstanceState(nCtx, operInstBriefData); refreshErr != nil {
			logger.G.Sys().Ctx(nCtx).WithErr(refreshErr).With(
				"oper-inst-id", operationInstanceID, "action-name", actionDef.Name()).
				Error("failed to refresh operation instance state after failed to start")

			return fmt.Errorf("failed to start operation instance. "+
				"oper-inst-id(%s), action-name(%s): %v", operationInstanceID, actionDef.Name(), refreshErr)
		}
	}

	// handle action instance lifecycle.
	actionInstData.Lifecycle.Start()
	if err = mgr.updateActionLifecycle(nCtx, operationInstanceID, actionDef.Name(), actionInstData.Lifecycle); err != nil {
		return err
	}

	// update latest action brief data to instance and operation.
	if err = mgr.updateOperationLatestActionBriefData(nCtx, operationInstanceID, actionDef, operInstBriefData); err != nil {
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
	if err = mgr.updateActionLifecycle(nCtx, operationInstanceID, actionDef.Name(), actionInstData.Lifecycle); err != nil {
		return err
	}

	// updates action content.
	if err = mgr.updateActionContent(nCtx, operationInstanceID, actionDef.Name(), actionInstData.Content); err != nil {
		return err
	}

	// updates action instance private data.
	if err = mgr.updateOperationInstancePrivateData(
		nCtx, operationInstanceID, actionDef.Name(), actionInstData.PrivateData); err != nil {
		return err
	}

	// when action done or error happens, we need to update the state of the operation instance.
	if actionInstData.IsLast() || executeErr != nil {
		// record oper inst metric.
		defer metric.OperationInstanceProcessed(operInstBriefData)

		operInstBriefData.Lifecycle.End(actionInstData.Lifecycle.State)
		if refreshErr := mgr.refreshOperationInstanceState(nCtx, operInstBriefData); refreshErr != nil {
			logger.G.Sys().Ctx(nCtx).WithErr(refreshErr).With(
				"oper-inst-id", operationInstanceID, "action-name", actionDef.Name()).
				Error("failed to refresh operation instance state after failed to end")

			return fmt.Errorf("failed to end operation instance. "+
				"oper-inst-id(%s), action-name(%s): %v", operationInstanceID, actionDef.Name(), refreshErr)
		}

		logger.G.Sys().Ctx(nCtx).With("oper-inst-id", operationInstanceID, "lifecycle", operInstBriefData.Lifecycle).
			Info("updated operation instance lifecycle with terminated state")
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

// refreshOperationInstanceState refreshes operation instance state and latest instance brief data.
func (mgr *manager) refreshOperationInstanceState(
	ctx contextx.IContext,
	operInstBriefData *operation.InstanceBriefData) error {

	operationID := operInstBriefData.Metadata.OperationID
	operationInstanceID := operInstBriefData.Metadata.OperationInstanceID
	// update operation instance lifecycle.
	if err := mgr.stgOperationInstance.UpdateOperationInstanceLifecycle(ctx,
		operationInstanceID, operInstBriefData.Lifecycle); err != nil {
		return fmt.Errorf("failed to update operation instance lifecycle. "+
			"oper-inst-id(%s): %v", operationInstanceID, err)
	}

	// update operation's latest instance brief data.
	if err := mgr.stgOperation.UpdateOperationLatestInstBriefData(ctx,
		operationID, operInstBriefData); err != nil {
		return fmt.Errorf("failed to update operation latest inst brief data. "+
			"operation-id(%s): %v", operationID, err)
	}

	return nil
}

func (mgr *manager) updateOperationLatestActionBriefData(
	ctx contextx.IContext,
	operationInstanceID string,
	actionDef action.Definition,
	operInstBriefData *operation.InstanceBriefData) error {

	latestActionBriefData := &action.InstanceBriefData{
		Name: actionDef.Name(),
		Tags: actionDef.Tags(),
	}

	// update instance.
	if err := mgr.stgOperationInstance.UpdateOperationLatestActionInstBriefData(ctx,
		operationInstanceID, latestActionBriefData); err != nil {
		return fmt.Errorf("failed to update instance table latest action brief data. "+
			"oper-inst-id(%s): %v", operationInstanceID, err)
	}

	operInstBriefData.LatestActionInstBriefData = latestActionBriefData
	// update operation.
	if err := mgr.stgOperation.UpdateOperationLatestInstBriefData(ctx,
		operInstBriefData.Metadata.OperationID, operInstBriefData); err != nil {
		return fmt.Errorf("failed to update operation table latest inst brief data. "+
			"operation-id(%s): %v", operInstBriefData.Metadata.OperationID, err)
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
		logger.G.Sys().Ctx(ctx).
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

			logger.G.Sys().Ctx(actionInstCtx.Ctx).
				WithErr(err).
				With("info", actionInstCtx.Data.Info(), "recover", r, "stack", stack).
				Info("failed to execute action, recover from panic")

			actionInstCtx.Data.Log().
				Zh("步骤发生异常: recover(%v), stack(%s)", r, stack).
				En("action panic: recover(%v), stack(%s)", r, stack).
				Error()
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
				logger.G.Sys().Ctx(ctx).WithErr(err).With("action", data.Name).Error("failed to refresh action inst data messages")
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
				logger.G.Sys().Ctx(ctx).WithErr(err).With("action", data.Name).Error("failed to refresh action inst data messages")
			}

			continue
		}
	}
}

// callActionDefWithRetry do action with retry.
func (mgr *manager) callActionDefWithRetry(actionInstCtx *action.InstanceContext, actionDef action.Definition) error {
	var doErr error

	for retryNum := uint(0); retryNum <= actionDef.MaxRetryCount() && retryNum < engineMaxRetryLimit; retryNum++ {
		logger.G.Sys().Ctx(actionInstCtx.Ctx).
			With("operation", actionInstCtx.Data.OperationDefName).
			With("oper-inst-id", actionInstCtx.Data.OperationInstanceID, "action", actionInstCtx.Data.Name, "retry", retryNum).
			Info("started action")

		if retryNum > 0 {
			actionInstCtx.Data.Log().
				Zh("第%d次重试步骤", retryNum).
				En("retried action, for the No.%d attempt", retryNum).
				Info()
		}

		doErr = actionDef.Do(actionInstCtx)

		if doErr != nil {
			logger.G.Sys().Ctx(actionInstCtx.Ctx).
				WithErr(doErr).
				With("operation", actionInstCtx.Data.OperationDefName).
				With("oper-inst-id", actionInstCtx.Data.OperationInstanceID, "action", actionInstCtx.Data.Name, "retry", retryNum).
				Error("failed to do action")

			if retryNum < actionDef.MaxRetryCount() && retryNum+1 < engineMaxRetryLimit {
				actionInstCtx.Data.Log().Zh("步骤执行失败, 即将重试: %v", doErr).En("action failed, about to retry: %v", doErr).Warn()
			}

			delayFn := actionDef.DelayFn()
			if delayFn != nil {
				delayFn()
			}

			continue
		}

		logger.G.Sys().Ctx(actionInstCtx.Ctx).
			With("operation", actionInstCtx.Data.OperationDefName).
			With("oper-inst-id", actionInstCtx.Data.OperationInstanceID, "action", actionInstCtx.Data.Name, "retry", retryNum).
			Info("done action")

		break
	}

	if doErr != nil {
		logger.G.Sys().Ctx(actionInstCtx.Ctx).
			WithErr(doErr).
			With("operation", actionInstCtx.Data.OperationDefName).
			With("oper-inst-id", actionInstCtx.Data.OperationInstanceID, "action", actionInstCtx.Data.Name).
			Error("failed to do action with all attempts")

		actionInstCtx.Data.Log().Zh("步骤最终失败: %v", doErr).En("action failed: %v", doErr).Error()
	}

	return doErr
}

// doOperExtraExecution executes the extra action for the operation instance.
func (mgr *manager) doOperExtraExecution(nCtx contextx.IContext, oper *operation.InstanceBriefData) error {
	if oper.Metadata.ExtraExecutionName == "" {
		return nil
	}

	if !oper.Lifecycle.NeedExecutedExtraExecution() {
		return nil
	}

	actionDef, ok := mgr.registeredOperExtraExecutionDefs[oper.Metadata.ExtraExecutionName]
	if !ok {
		return fmt.Errorf("extra action not registered, name(%s)", oper.Metadata.ExtraExecutionName)
	}

	if execErr := actionDef.Do(nCtx, oper); execErr != nil {
		oper.Lifecycle.End(action.StateFailed)
		if refreshErr := mgr.refreshOperationInstanceState(nCtx, oper); refreshErr != nil {
			logger.G.Sys().Ctx(nCtx).WithErr(refreshErr).With(
				"oper-inst-id", oper.Metadata.OperationInstanceID, "extra-execution-name", oper.Metadata.ExtraExecutionName).
				Error("failed to refresh operation instance state after failed to execute extra execution")

			return fmt.Errorf("failed to execute operation instance extra execution. "+
				"oper-inst-id(%s), extra-execution-name(%s): %v; additionally failed to refresh operation instance state: %v",
				oper.Metadata.OperationInstanceID, oper.Metadata.ExtraExecutionName, execErr, refreshErr)
		}

		return fmt.Errorf("failed to execute operation instance extra execution. "+
			"oper-inst-id(%s), extra-execution-name(%s): %v",
			oper.Metadata.OperationInstanceID, oper.Metadata.ExtraExecutionName, execErr)
	}

	return nil
}

func (mgr *manager) autoRefreshExtraExecutionMsg(nCtx contextx.IContext, oper *operation.InstanceBriefData) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	idx := len(oper.Metadata.ExtraExecutionMessages)

	// refresh action inst data messages to storage
	for {
		select {
		case <-nCtx.Done():
			// when finished, push the rest messages to storage.
			msgs := oper.Metadata.ExtraExecutionMessages[idx:]

			// need to make sure the db operation done, so in this way we use mgr.ctx instead of ctx.
			// nolint: contextcheck
			err := mgr.stgOperationInstance.UpdateOperationInstanceExtraExecutionMessages(mgr.ctx,
				oper.Metadata.OperationInstanceID,
				msgs...)
			if err != nil {
				logger.G.Sys().Ctx(nCtx).WithErr(err).
					With("operation-extra-execution", oper.Metadata.ExtraExecutionName).
					Error("failed to refresh operation extra execution messages")
			}

			return

		case <-ticker.C:
			msgs := oper.Metadata.ExtraExecutionMessages[idx:]
			idx += len(msgs)

			if len(msgs) == 0 {
				continue
			}

			// need to make sure the db operation done, so in this way we use mgr.ctx instead of ctx.
			// nolint: contextcheck
			err := mgr.stgOperationInstance.UpdateOperationInstanceExtraExecutionMessages(
				mgr.ctx,
				oper.Metadata.OperationInstanceID,
				msgs...)
			if err != nil {
				logger.G.Sys().Ctx(nCtx).WithErr(err).
					With("operation-extra-execution", oper.Metadata.ExtraExecutionName).
					Error("failed to refresh operation extra execution messages")
			}

			continue
		}
	}
}

func actionMessageID(operInstID, actionName string) string {
	return fmt.Sprintf("%s|%s", operInstID, actionName)
}
