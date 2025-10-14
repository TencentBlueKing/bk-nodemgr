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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operinstdata"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/common"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// updateOperInstActionStatus update the oper inst action status.
func (s *Storage) updateOperInstActionStatus(
	nCtx contextx.IContext, operInstID string, actionName string, status action.State) error {

	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if operInstID == "" {
		return errors.New("oper inst id is empty")
	}

	if actionName == "" {
		return errors.New("action name is empty")
	}

	if err := status.Validate(); err != nil {
		return err
	}

	return s.daoOperInstData.UpdateActionInstStatus(nCtx, operInstID, actionName, status)
}

// getActionInstanceData get action instance data.
func (s *Storage) getActionInstanceData(
	nCtx contextx.IContext, operInstID, actionName string) (*action.InstanceData, error) {

	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if operInstID == "" {
		return nil, basestorage.ErrEmptyOperaInstID()
	}

	if actionName == "" {
		return nil, basestorage.ErrEmptyActionName()
	}

	return s.daoOperInstData.GetActionInstData(nCtx, operInstID, actionName)
}

// getActionInstanceLifecycle gets action instance lifecycle.
func (s *Storage) getActionInstanceLifecycle(
	nCtx contextx.IContext, operInstID, actionName string) (*action.Lifecycle, error) {

	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if operInstID == "" {
		return nil, errors.New("operation instance id is empty")
	}

	if actionName == "" {
		return nil, errors.New("action name is empty")
	}

	return s.daoOperInstData.GetActInstLifecycle(nCtx, operInstID, actionName)
}

// getActionInstancePrivateData gets action instance private data.
func (s *Storage) getActionInstancePrivateData(
	nCtx contextx.IContext, operInstID, actionName string) (map[string]any, error) {

	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if operInstID == "" {
		return nil, errors.New("operation instance id is empty")
	}

	if actionName == "" {
		return nil, errors.New("action name is empty")
	}

	return s.daoOperInstData.GetActInstPrivateData(nCtx, operInstID, actionName)
}

// updateActionInstanceLifecycle updates action instance lifecycle.
func (s *Storage) updateActionInstanceLifecycle(
	nCtx contextx.IContext, operInstID, actionName string, lifecycle *action.Lifecycle) error {

	if nCtx == nil {
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

	if err := s.existsAction(nCtx, operInstID, actionName); err != nil {
		return fmt.Errorf("action does not exist, operation-inst-id(%s), action-name(%s): %w",
			operInstID, actionName, err)
	}

	return s.daoOperInstData.UpdateActInstLifecycle(nCtx, operInstID, actionName, lifecycle)
}

// pushActionInstanceMessage pushes action instance message.
func (s *Storage) pushActionInstanceMessage(
	nCtx contextx.IContext, operInstID, actionName string, messages ...common.Message) error {

	if nCtx == nil {
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

	if err := s.existsAction(nCtx, operInstID, actionName); err != nil {
		return fmt.Errorf("action does not exist, operation-inst-id(%s), action-name(%s): %w",
			operInstID, actionName, err)
	}

	for _, msg := range messages {
		if err := s.daoOperInstData.PushActionInstanceMessage(nCtx, operInstID, actionName, msg); err != nil {
			return fmt.Errorf("failed to push action instance message, operation-inst-id(%s), action-name(%s): %w",
				operInstID, actionName, err)
		}
	}

	return nil
}

// getOperationInstanceData gets operation instance data.
func (s *Storage) getOperationInstanceData(
	nCtx contextx.IContext, operInstID string) (*operation.InstanceData, error) {

	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if operInstID == "" {
		return nil, errors.New("operation inst id is empty")
	}

	return s.daoOperInstData.FindOne(nCtx, operinstdata.WithOperInstID(operInstID))
}

// getOperationInstanceData gets operation instance data.
func (s *Storage) getOperationInstanceDataBriefData(
	nCtx contextx.IContext, operInstID string) (*operation.InstanceBriefData, error) {

	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if operInstID == "" {
		return nil, errors.New("operation inst id is empty")
	}

	oper, err := s.daoOperInstData.FindOne(nCtx, operinstdata.WithOperInstID(operInstID))
	if err != nil {
		return nil, err
	}

	return &oper.InstanceBriefData, err
}

// listOperationInstanceBriefDataWithoutActionInst lists operation instance brief data without action instance data.
func (s *Storage) listOperationInstanceBriefDataWithoutActionInst(
	nCtx contextx.IContext, page types.Page, conditions ...*types.OperInstDataCondition) (
	[]*operation.InstanceBriefData, int64, error) {

	if nCtx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	if err := page.Validate(); err != nil {
		return nil, 0, err
	}

	return s.daoOperInstData.ListWithoutActInst(nCtx, page, convertOperInstDataConditionsToOptions(conditions...)...)
}

// countOperationInstance counts operation instance.
func (s *Storage) countOperationInstance(
	nCtx contextx.IContext, triggerID string, states ...operation.State) (int64, error) {

	if nCtx == nil {
		return 0, basestorage.ErrNilContent()
	}

	if triggerID == "" {
		return 0, errors.New("trigger id is empty")
	}

	return s.daoOperInstData.Count(nCtx, operinstdata.WithTriggerID(triggerID), operinstdata.WithState(states...))
}

// listOperationInstanceBriefDataWithoutActionInstByOperationID lists operation instance brief data without action instance data.
func (s *Storage) listOperationInstanceBriefDataWithoutActionInstByOperationID(
	nCtx contextx.IContext, page types.Page, operationID ...string) (
	[]*operation.InstanceBriefData, int64, error) {

	if nCtx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	if err := page.Validate(); err != nil {
		return nil, 0, err
	}

	return s.daoOperInstData.ListWithoutActInst(nCtx, page, operinstdata.WithOperationID(operationID...))
}

// listOperationInstanceBriefDataWithoutActionInstByTriggerID lists operation instance brief data without action instance data.
func (s *Storage) listOperationInstanceBriefDataWithoutActionInstByTriggerID(
	nCtx contextx.IContext, page types.Page, triggerID ...string) (
	[]*operation.InstanceBriefData, int64, error) {

	if nCtx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	if err := page.Validate(); err != nil {
		return nil, 0, err
	}

	return s.daoOperInstData.ListWithoutActInst(nCtx, page, operinstdata.WithTriggerID(triggerID...))
}

// upsertOperationInstanceData upserts operation instance data.
func (s *Storage) upsertOperationInstanceData(nCtx contextx.IContext, operInstData *operation.InstanceData) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if operInstData == nil {
		return basestorage.ErrUpsertNilData()
	}

	return s.daoOperInstData.Upsert(nCtx, operInstData)
}

// updateOperationInstanceLifecycle updates operation instance lifecycle.
func (s *Storage) updateOperationInstanceLifecycle(
	nCtx contextx.IContext, operInstID string, lifecycle *operation.Lifecycle) error {

	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if lifecycle == nil {
		return errors.New("lifecycle is nil")
	}

	return s.daoOperInstData.UpdateLifeCycle(nCtx, operInstID, lifecycle)
}

// updateOperationInstanceExtraExecutionMessages updates operation instance execution messages.
func (s *Storage) updateOperationInstanceExtraExecutionMessages(
	nCtx contextx.IContext, operInstID string, messages ...common.Message) error {

	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if operInstID == "" {
		return basestorage.ErrEmptyOperaInstID()
	}

	if len(messages) == 0 {
		return nil
	}

	return s.daoOperInstData.UpdateExtraExecutionMessages(nCtx, operInstID, messages...)
}

// StopEventSubscription represents the stop event subscription.
type StopEventSubscription struct {
	OperInstID string
	C          chan<- struct{}
}

// syncStopOperInsts sync all stopping operation instances.
func (s *Storage) syncStopOperInsts(nCtx contextx.IContext) error {
	stopInstIDs, err := s.daoStopOperInst.FindAll(nCtx)
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
		err := s.checkNotifyStopping(nCtx)
		if err != nil {
			logger.G.Sys().WithErr(err).Error("failed to check notify stopping")
		}
	}()

	return nil
}

// checkNotifyStopping check and notify the stopping event.
func (s *Storage) checkNotifyStopping(nCtx contextx.IContext) error {
	_, err, _ := s.sg.Do("checkNotifyStopping", func() (interface{}, error) {
		err := s.processStoppingEvents(nCtx)
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
func (s *Storage) processStoppingEvents(nCtx contextx.IContext) error {
	nCtx, cancel := contextx.WithTimeout(contextx.From(nCtx), 5*time.Second) // nolint: mnd
	defer cancel()

	notifications := s.getNotifications()

	for _, notify := range notifications {
		select {
		case <-nCtx.Done():
			return nCtx.Err()
		case notify.Subscription.C <- struct{}{}:
			s.stopEventSubsMapMutex.Lock()
			delete(s.stopEventSubsMap, notify.Key)
			s.stopEventSubsMapMutex.Unlock()
		default:
			logger.G.Sys().With("notify", notify).Error("failed to notify stopping event")
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
	nCtx contextx.IContext, operInstID string, actionName string, content map[string]any) error {

	if nCtx == nil {
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

	if err := s.existsAction(nCtx, operInstID, actionName); err != nil {
		return fmt.Errorf("action does not exist, operation-inst-id(%s), action-name(%s): %w",
			operInstID, actionName, err)
	}

	return s.daoOperInstData.UpdateActionInstContent(nCtx, operInstID, actionName, content)
}

// existsAction checks if the action exists in the operation instance.
func (s *Storage) existsAction(nCtx contextx.IContext, operInstID string, actionName string) error {
	oper, err := s.daoOperInstData.FindOneWithoutActionData(nCtx, operinstdata.WithOperInstID(operInstID))
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
	nCtx contextx.IContext, operInstID string, actionName string, privateData map[string]any) error {

	if nCtx == nil {
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

	if err := s.existsAction(nCtx, operInstID, actionName); err != nil {
		return fmt.Errorf("action does not exist, operation-inst-id(%s), action-name(%s): %w",
			operInstID, actionName, err)
	}

	if len(privateData) == 0 {
		return nil
	}

	return s.daoOperInstData.PushActInstPrivateData(nCtx, operInstID, actionName, privateData)
}

// deleteOperationInstances deletes operation instances by given operation instance IDs.
func (s *Storage) deleteOperationInstances(nCtx contextx.IContext, operInstID ...string) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if len(operInstID) == 0 {
		return nil
	}

	return s.daoOperInstData.Delete(nCtx, operInstID...)
}

// deleteOperationInstancesByTriggerID deletes operation instances by given triggerID IDs.
func (s *Storage) deleteOperationInstancesByTriggerID(nCtx contextx.IContext, triggerID ...string) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if len(triggerID) == 0 {
		return nil
	}

	return s.daoOperInstData.Delete(nCtx, triggerID...)
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
