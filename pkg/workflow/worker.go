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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/common"
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
		mgr.logger.Errorf("worker error, err: %v", err)
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

func (mgr *manager) do(ctx context.Context, actionName string, operationInstanceID string) error {
	actionDef, ok := mgr.registeredActionDefs[actionName]
	if !ok {
		return fmt.Errorf("action not registered, name(%s)", actionName)
	}

	// get action instance.
	actionInstData, err := mgr.storage.GetActionInstanceData(ctx, operationInstanceID, actionName)
	if err != nil {
		return fmt.Errorf("failed to get action instance data from operation instance. "+
			"oper-inst-id(%s), action-name(%s), err: %v", operationInstanceID, actionName, err)
	}

	// check if this action should be executed.
	if err := actionInstData.NeedExecuted(); err != nil {
		if err == common.ErrActionAlreadySucceeded() || err == common.ErrActionSkipped() {
			return nil
		}

		return err
	}

	// get operation instance.
	operInstBriefData, err := mgr.storage.GetOperationInstanceBriefData(ctx, operationInstanceID)
	if err != nil {
		return fmt.Errorf("failed to get operation instance brief data. "+
			"oper-inst-id(%s), err: %v", operationInstanceID, err)
	}

	// handle operation instance lifecycle.
	if actionInstData.IsFirst() {
		// first action be executed, means operation instance is started.
		operInstBriefData.Lifecycle.Start()
		if err = mgr.updateOperationInstanceLifecycle(ctx, operationInstanceID, operInstBriefData.Lifecycle); err != nil {
			return err
		}
	}

	// handle action instance lifecycle.
	actionInstData.Lifecycle.Start()
	if err = mgr.updateActionLifecycle(ctx, operationInstanceID, actionName, actionInstData.Lifecycle); err != nil {
		return err
	}

	// handle action content.
	if actionInstData.IsFirst() {
		// first action should get content from operation instance init data.
		actionInstData.Content = operInstBriefData.Metadata.InitContent
	} else {
		// not-first action should get content from previous action.
		preActionName := operInstBriefData.Metadata.ActionNames[actionInstData.Index-1]
		preActionInstData, err := mgr.storage.GetActionInstanceData(ctx, operationInstanceID, preActionName)
		if err != nil {
			return fmt.Errorf("failed to get pre action instance data from operation instance. "+
				"oper-inst-id(%s), action-name(%s), err: %v", operationInstanceID, preActionName, err)
		}

		actionInstData.Content = preActionInstData.Content
	}

	// execute and wait for action done.
	executeErr := mgr.executeAndWatchAction(ctx, actionDef, operInstBriefData, actionInstData)

	// updates action instance lifecycle.
	if err = mgr.updateActionLifecycle(ctx,
		operationInstanceID, actionName, actionInstData.Lifecycle); err != nil {
		return err
	}

	// when action done or error happens, we need to update the state of the operation instance.
	if actionInstData.IsLast() || executeErr != nil {
		operInstBriefData.Lifecycle.End(actionInstData.Lifecycle.State)
		if err = mgr.updateOperationInstanceLifecycle(ctx, operationInstanceID, operInstBriefData.Lifecycle); err != nil {
			return fmt.Errorf("failed to update operation instance lifecycle. err: %v, execution-error(%v)",
				err, executeErr)
		}
	}

	return executeErr
}

func (mgr *manager) updateActionLifecycle(
	ctx context.Context,
	operationInstanceID,
	actionName string,
	actionInstLifecycle *action.Lifecycle) error {

	if err := mgr.storage.UpdateActionInstanceLifecycle(ctx,
		operationInstanceID, actionName, actionInstLifecycle); err != nil {
		return fmt.Errorf("failed to update action instance lifecycle. "+
			"oper-inst-id(%s), action-name(%s), err: %v", operationInstanceID, actionName, err)
	}

	return nil
}

func (mgr *manager) updateOperationInstanceLifecycle(
	ctx context.Context,
	operationInstanceID string,
	operInstLifecycle *operation.Lifecycle) error {

	if err := mgr.storage.UpdateOperationInstanceLifecycle(ctx,
		operationInstanceID, operInstLifecycle); err != nil {
		return fmt.Errorf("failed to update operation instance lifecycle. "+
			"oper-inst-id(%s), err: %v", operationInstanceID, err)
	}

	return nil
}

func (mgr *manager) executeAndWatchAction(ctx context.Context,
	actionDef action.Definition,
	operInstBriefData *operation.InstanceBriefData,
	actionInstData *action.InstanceData) error {

	actionTimeoutCtx, actionTimeoutCancel := context.WithTimeout(ctx, actionDef.Timeout())
	defer actionTimeoutCancel()

	operationTimeoutCtx, operationTimeoutCancel := context.WithDeadline(
		ctx, operInstBriefData.Lifecycle.StartedAt.Add(operInstBriefData.Metadata.Timeout))
	defer operationTimeoutCancel()

	// watch storage for stopping event.
	terminatingC := mgr.storage.WatchOperInstStopping(actionTimeoutCtx, operInstBriefData.Metadata.OperationID)

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
			mgr.logger.Errorf("action panic, info(%v), revoer(%v), stack(%s)",
				actionInstCtx.Data.Info(), r, debug.Stack())
			err = fmt.Errorf("action panic, info(%v), revoer(%v), stack(%s)",
				actionInstCtx.Data.Info(), r, debug.Stack())
		}

		doResult <- err
	}()

	ctx, cancel := context.WithCancel(actionInstCtx.Ctx)
	defer cancel()

	go mgr.autoRefreshActionDataMsg(ctx, actionInstCtx.Data)

	err = mgr.callActionDefWithRetry(actionInstCtx, actionDef)
}

func (mgr *manager) autoRefreshActionDataMsg(ctx context.Context, data *action.InstanceData) {
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
			err := mgr.storage.PushActionInstanceMessage(mgr.ctx, data.OperationID, data.Name, msgs...) // nolint: contextcheck
			if err != nil {
				mgr.logger.Errorf("failed to refresh action inst data messages, action-name(%s), err: %v",
					data.Name, err)
			}

			return

		case <-ticker.C:
			msgs := data.Messages[idx:]
			idx += len(msgs)

			// need to make sure the db operation done, so in this way we use mgr.ctx instead of ctx.
			err := mgr.storage.PushActionInstanceMessage(mgr.ctx, data.OperationID, data.Name, msgs...) // nolint: contextcheck
			if err != nil {
				mgr.logger.Errorf("failed to refresh action inst data messages, action-name(%s), err: %v",
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
		mgr.logger.Infof("started action, action-name(%s), oper-def-name(%s), retry-num(%d)",
			actionInstCtx.Data.Name, actionInstCtx.Data.OperationDefName, retryNum)
		actionInstCtx.Data.Log(fmt.Sprintf("started action, action-name(%s), retry-num(%d)",
			actionInstCtx.Data.Name, retryNum))

		doErr = actionDef.Do(actionInstCtx)

		if doErr != nil {
			mgr.logger.Errorf("failed to do action, operinst-id(%s), action-name(%s), retry-num(%d), err: %v",
				actionInstCtx.Data.OperationID, actionInstCtx.Data.Name, retryNum, doErr)
			actionInstCtx.Data.Log(fmt.Sprintf("failed to do action, action-name(%s), retry-num(%d), err: %v",
				actionInstCtx.Data.Name, retryNum, doErr))

			actionDef.DelayFn()

			continue
		}

		mgr.logger.Infof("done action, action-name(%s), oper-def-name(%s), retry-num(%d)",
			actionInstCtx.Data.Name, actionInstCtx.Data.OperationDefName, retryNum)
		actionInstCtx.Data.Log(fmt.Sprintf("done action, action-name(%s), oper-def-name(%s), retry-num(%d)",
			actionInstCtx.Data.Name, actionInstCtx.Data.OperationDefName, retryNum))

		break
	}

	return doErr
}
