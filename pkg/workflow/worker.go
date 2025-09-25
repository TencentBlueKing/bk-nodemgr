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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/common"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/metric"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
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
		mgr.logger.Errorf("worker error: %v", err)
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

// doAction executes the action defined by actionName for the operation instance with operationInstanceID.
// nolint: funlen,gocognit,cyclop,gocyclo
// NOCC: golint/fnsize(func design is not suitable for splitting).
// notice: this func only accept context.Context as input, so we accept context.Context and then change it to contextx.IContext.
func (mgr *manager) do(ctx context.Context, actionName string, operationInstanceID string) error {
	var nCtx contextx.IContext
	nCtx = contextx.New(ctx, contextx.WithMessageID(actionMessageID(operationInstanceID, actionName)))

	actionDef, ok := mgr.registeredActionDefs[actionName]
	if !ok {
		// record metric.
		metric.ActionNotRegistered(actionName)

		return fmt.Errorf("action not registered, name(%s)", actionName)
	}

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

	// handle extra execution before action executed.
	// if retry happens, action maybe not first
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
	if actionInstData.IsFirst() {
		// first action be executed, means operation instance is started.
		operInstBriefData.Lifecycle.State = operation.StateRunning
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

		mgr.logger.InfoCtxf(nCtx, "updated operation instance lifecycle with terminated state. oper-inst-id(%s), lifecycle(%+v)",
			operationInstanceID, operInstBriefData.Lifecycle)

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
		mgr.logger.DebugCtxf(ctx, "no private data to update, oper-inst-id(%s), action-name(%s), private-data(%v)",
			operationInstanceID, actionName, privateData)

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

func (mgr *manager) executeAndWatchAction(ctx contextx.IContext,
	actionDef action.Definition,
	operInstBriefData *operation.InstanceBriefData,
	actionInstData *action.InstanceData) error {

	actionTimeoutCtx, actionTimeoutCancel := contextx.WithTimeout(contextx.From(ctx), actionDef.Timeout())
	defer actionTimeoutCancel()

	operationTimeoutCtx, operationTimeoutCancel := context.WithDeadline(
		ctx, operInstBriefData.Lifecycle.StartedAt.Add(operInstBriefData.Metadata.Timeout))
	defer operationTimeoutCancel()

	// watch storage for stopping event.
	terminatingC := mgr.stgOperationInstance.WatchOperInstStopping(
		actionTimeoutCtx, operInstBriefData.Metadata.OperationID)

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

	case <-operationTimeoutCtx.Done():
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

	case <-ctx.Done():
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
			mgr.logger.ErrorCtxf(actionInstCtx.Ctx, "action panic, info(%v), revoer(%v), stack(%s)",
				actionInstCtx.Data.Info(), r, debug.Stack())
			err = fmt.Errorf("action panic, info(%v), revoer(%v), stack(%s)",
				actionInstCtx.Data.Info(), r, debug.Stack())
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
				mgr.logger.ErrorCtxf(ctx, "failed to refresh action inst data messages, action-name(%s): %v",
					data.Name, err)
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
				mgr.logger.ErrorCtxf(ctx, "failed to refresh action inst data messages, action-name(%s): %v",
					data.Name, err)
			}

			continue
		}
	}
}

// callActionDefWithRetry do action with retry.
func (mgr *manager) callActionDefWithRetry(actionInstCtx *action.InstanceContext, actionDef action.Definition) error {
	var doErr error

	for retryNum := uint(0); retryNum <= actionDef.MaxRetryCount() && retryNum < engineMaxRetryLimit; retryNum++ {
		mgr.logger.InfoCtxf(actionInstCtx.Ctx, "started action, action-name(%s), oper-def-name(%s), retry-num(%d)",
			actionInstCtx.Data.Name, actionInstCtx.Data.OperationDefName, retryNum)
		actionInstCtx.Data.LogI(fmt.Sprintf("started action, action-name(%s), retry-num(%d)",
			actionInstCtx.Data.Name, retryNum))

		doErr = actionDef.Do(actionInstCtx)

		if doErr != nil {
			mgr.logger.ErrorCtxf(actionInstCtx.Ctx, "failed to do action, operinst-id(%s), action-name(%s), retry-num(%d): %v",
				actionInstCtx.Data.OperationInstanceID, actionInstCtx.Data.Name, retryNum, doErr)
			actionInstCtx.Data.LogW(fmt.Sprintf("failed to do action, action-name(%s), retry-num(%d): %v",
				actionInstCtx.Data.Name, retryNum, doErr))

			actionDef.DelayFn()

			continue
		}

		mgr.logger.InfoCtxf(actionInstCtx.Ctx, "done action, action-name(%s), oper-def-name(%s), retry-num(%d)",
			actionInstCtx.Data.Name, actionInstCtx.Data.OperationDefName, retryNum)
		actionInstCtx.Data.LogI(fmt.Sprintf("done action, action-name(%s), oper-def-name(%s), retry-num(%d)",
			actionInstCtx.Data.Name, actionInstCtx.Data.OperationDefName, retryNum))

		break
	}

	if doErr != nil {
		mgr.logger.ErrorCtxf(actionInstCtx.Ctx, "action failed, action-name(%s), oper-def-name(%s), err(%v)",
			actionInstCtx.Data.Name, actionInstCtx.Data.OperationDefName, doErr)
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
		mgr.logger.Errorf("refresh operation-extra-execution(%s) message failed: %v",
			oper.Metadata.ExtraExecutionName, err)
	}

	return err
}

func actionMessageID(operInstID, actionName string) string {
	return fmt.Sprintf("%s|%s", operInstID, actionName)
}
