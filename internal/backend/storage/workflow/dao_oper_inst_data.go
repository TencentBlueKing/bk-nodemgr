/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package workflow

import (
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operinstdata"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/common"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// updateOperInstActionStatus update the oper inst action status.
func (s *Storage) updateOperInstActionStatus(
	ctx contextx.IContext, operInstID string, actionName string, status action.State) error {

	return s.daoOperInstData.UpdateActionInstStatus(ctx, operInstID, actionName, status)
}

// getActionInstanceData get action instance data.
func (s *Storage) getActionInstanceData(
	ctx contextx.IContext, operInstID, actionName string) (*action.InstanceData, error) {

	return s.daoOperInstData.GetActionInstData(ctx, operInstID, actionName)
}

// getActionInstanceLifecycle gets action instance lifecycle.
func (s *Storage) getActionInstanceLifecycle(
	ctx contextx.IContext, operInstID, actionName string) (*action.Lifecycle, error) {

	return s.daoOperInstData.GetActInstLifecycle(ctx, operInstID, actionName)
}

// getActionInstancePrivateData gets action instance private data.
func (s *Storage) getActionInstancePrivateData(
	ctx contextx.IContext, operInstID, actionName string) (map[string]any, error) {

	return s.daoOperInstData.GetActInstPrivateData(ctx, operInstID, actionName)
}

// updateActionInstanceLifecycle updates action instance lifecycle.
func (s *Storage) updateActionInstanceLifecycle(
	ctx contextx.IContext, operInstID, actionName string, lifecycle *action.Lifecycle) error {

	if err := s.existsAction(ctx, operInstID, actionName); err != nil {
		return fmt.Errorf("action does not exist, operation-inst-id(%s), action-name(%s): %w",
			operInstID, actionName, err)
	}

	return s.daoOperInstData.UpdateActInstLifecycle(ctx, operInstID, actionName, lifecycle)
}

// pushActionInstanceMessage pushes action instance message.
func (s *Storage) pushActionInstanceMessage(
	ctx contextx.IContext, operInstID, actionName string, messages ...common.Message) error {

	if err := s.existsAction(ctx, operInstID, actionName); err != nil {
		return fmt.Errorf("action does not exist, operation-inst-id(%s), action-name(%s): %w",
			operInstID, actionName, err)
	}

	for _, msg := range messages {
		if err := s.daoOperInstData.PushActionInstanceMessage(ctx, operInstID, actionName, msg); err != nil {
			return fmt.Errorf("failed to push action instance message, operation-inst-id(%s), action-name(%s): %w",
				operInstID, actionName, err)
		}
	}

	return nil
}

// getOperationInstanceData gets operation instance data.
func (s *Storage) getOperationInstanceData(
	ctx contextx.IContext, conditions ...*types.OperInstDataCondition) (*operation.InstanceData, error) {

	return s.daoOperInstData.FindOne(ctx, convertOperInstDataConditionsToOptions(conditions...)...)
}

// listOperationInstanceBriefDataWithoutActionInst lists operation instance brief data without action instance data.
func (s *Storage) listOperationInstanceBriefDataWithoutActionInst(
	ctx contextx.IContext, page types.Page, conditions ...*types.OperInstDataCondition) (
	[]*operation.InstanceBriefData, int64, error) {

	return s.daoOperInstData.ListWithoutActInst(ctx, page, convertOperInstDataConditionsToOptions(conditions...)...)
}

// countOperationInstance counts operation instance.
func (s *Storage) countOperationInstance(
	ctx contextx.IContext, conditions ...*types.OperInstDataCondition) (int64, error) {

	return s.daoOperInstData.Count(ctx, convertOperInstDataConditionsToOptions(conditions...)...)
}

// upsertOperationInstanceData upserts operation instance data.
func (s *Storage) upsertOperationInstanceData(ctx contextx.IContext, operInstData *operation.InstanceData) error {
	return s.daoOperInstData.Upsert(ctx, operInstData)
}

// updateOperationInstanceLifecycle updates operation instance lifecycle.
func (s *Storage) updateOperationInstanceLifecycle(
	ctx contextx.IContext, operInstID string, lifecycle *operation.Lifecycle) error {

	return s.daoOperInstData.UpdateLifeCycle(ctx, operInstID, lifecycle)
}

// updateOperationInstanceExtraExecutionMessages updates operation instance execution messages.
func (s *Storage) updateOperationInstanceExtraExecutionMessages(
	ctx contextx.IContext, operInstID string, messages ...common.Message) error {

	return s.daoOperInstData.UpdateExtraExecutionMessages(ctx, operInstID, messages...)
}

// StopEventSubscription represents the stop event subscription.
type StopEventSubscription struct {
	OperInstID string
	C          chan<- struct{}
}

// syncStopOperInsts sync all stopping operation instances.
func (s *Storage) syncStopOperInsts(ctx contextx.IContext) error {
	stopInstIDs, err := s.daoStopOperInst.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to find all stopping operation instances: %v", err)
	}

	stopInstMap := make(map[string]struct{}, len(stopInstIDs))
	for _, stopInstID := range stopInstIDs {
		stopInstMap[stopInstID] = struct{}{}
	}

	s.stopOperInstsMutex.Lock()
	s.stopOperInsts = stopInstMap
	s.stopOperInstsMutex.Unlock()

	go func() {
		err := s.checkNotifyStopping(ctx)
		if err != nil {
			s.Logger.Errorf("sync stopping event succeed, but check notify stopping failed, err: %v", err)
		}
	}()

	return nil
}

// checkNotifyStopping check and notify the stopping event.
func (s *Storage) checkNotifyStopping(ctx contextx.IContext) error {
	_, err, _ := s.sg.Do("checkNotifyStopping", func() (interface{}, error) {
		err := s.processStoppingEvents(ctx)
		if err != nil {
			return nil, err
		}

		return nil, nil // nolint: nilnil
	})
	if err != nil {
		return err
	}

	return nil
}

// processStoppingEvents ...
func (s *Storage) processStoppingEvents(ctx contextx.IContext) error {
	ctx, cancel := contextx.WithTimeout(ctx, 5*time.Second) // nolint: mnd
	defer cancel()

	notifications := s.getNotifications()

	for _, notify := range notifications {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case notify.Subscription.C <- struct{}{}:
			s.stopEventSubsMapMutex.Lock()
			delete(s.stopEventSubsMap, notify.Key)
			s.stopEventSubsMapMutex.Unlock()
		default:
			s.Logger.Errorf("failed to notify stopping event, the channel is full, notify: %+v", notify)
		}
	}

	return nil
}

// notifyItem notify item.
type notifyItem struct {
	Key          string
	Subscription *StopEventSubscription
}

// getNotifications get the notifications.
func (s *Storage) getNotifications() []notifyItem {
	s.stopEventSubsMapMutex.Lock()
	defer s.stopEventSubsMapMutex.Unlock()
	s.stopOperInstsMutex.Lock()
	defer s.stopOperInstsMutex.Unlock()

	notifications := make([]notifyItem, 0, len(s.stopEventSubsMap))
	for key, subscription := range s.stopEventSubsMap {
		_, ok := s.stopOperInsts[subscription.OperInstID]
		if ok {
			notifications = append(notifications, notifyItem{
				Key:          key,
				Subscription: subscription,
			})
		}
	}

	return notifications
}

// UpdateActionInstanceContent update action instance content.
func (s *Storage) updateActionInstanceContent(
	ctx contextx.IContext, operInstID string, actionName string, content map[string]any) error {

	if err := s.existsAction(ctx, operInstID, actionName); err != nil {
		return fmt.Errorf("action does not exist, operation-inst-id(%s), action-name(%s): %w",
			operInstID, actionName, err)
	}

	return s.daoOperInstData.UpdateActionInstContent(ctx, operInstID, actionName, content)
}

// existsAction checks if the action exists in the operation instance.
func (s *Storage) existsAction(ctx contextx.IContext, operInstID string, actionName string) error {
	oper, err := s.daoOperInstData.FindOneWithoutActionData(ctx, operinstdata.WithOperInstID(operInstID))
	if err != nil {
		return err
	}
	if oper == nil {
		return errors.New("no found operation")
	}

	for _, act := range oper.Metadata.ActionNames {
		if act == actionName {
			return nil
		}
	}

	return errors.New("no found action")
}

// upsertActionInstancePrivateData upserts action instance private data.
func (s *Storage) upsertActionInstancePrivateData(
	ctx contextx.IContext, operInstID string, actionName string, privateData map[string]any) error {

	if err := s.existsAction(ctx, operInstID, actionName); err != nil {
		return fmt.Errorf("action does not exist, operation-inst-id(%s), action-name(%s), err: %w",
			operInstID, actionName, err)
	}

	if len(privateData) == 0 {
		return nil
	}

	return s.daoOperInstData.PushActInstPrivateData(ctx, operInstID, actionName, privateData)
}

// deleteOperationInstances deletes operation instances by given operation instance IDs.
func (s *Storage) deleteOperationInstances(ctx contextx.IContext, operInstID ...string) error {
	return s.daoOperInstData.Delete(ctx, operInstID...)
}

// convertOperInstDataConditionsToOptions converts OperInstDataCondition to OptFn.
func convertOperInstDataConditionsToOptions(conditions ...*types.OperInstDataCondition) []operinstdata.OptFn {
	opts := make([]operinstdata.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				operinstdata.WithTriggerID(condition.ExactInclude.TriggerID...),
				operinstdata.WithOperationID(condition.ExactInclude.OperationID...),
				operinstdata.WithOperInstID(condition.ExactInclude.OperInstID...),
				operinstdata.WithState(condition.ExactInclude.State...),
			)
		}
	}

	return opts
}
