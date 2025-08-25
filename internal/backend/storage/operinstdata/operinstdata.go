/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operinstdata ...
package operinstdata

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operinstdata"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/stopoperinst"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/scheduler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/singleflight"
)

// constants ...
const (
	StorageName       = "operinstdata"
	taskInterval      = 10 * time.Second
	taskTimeout       = 20 * time.Second
	syncOperationTask = "sync stopping operation inst"
)

// NewStorage ...
func NewStorage(client *mongo.Client, database string, logger logger.ILogger) (*Storage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &Storage{
		Storage: basestorage.Storage{
			Name:     StorageName,
			Database: client.Database(database),
			Logger:   logger,
		},
	}

	err := basestorage.InitStorage(&s.Storage,
		basestorage.WithStartFunc(s.startFn),
		basestorage.WithCheckFunc(s.check))
	if err != nil {
		s.Logger.Errorf("new Storage failed, err: %v", err)
		return nil, err
	}

	return s, nil
}

// Storage defines the storage interface for operinstdata.
type Storage struct {
	basestorage.Storage

	// dao
	daoOperinstdata operinstdata.IHandler
	stopoperinstDao stopoperinst.Handler

	// stop event subscriptions
	stopEventSubsMap      map[string]*StopEventSubscription
	stopEventSubsMapMutex sync.RWMutex

	stopOperInsts      map[string]struct{}
	stopOperInstsMutex sync.RWMutex

	sg singleflight.Group
}

func (s *Storage) startFn() error {
	s.daoOperinstdata = operinstdata.New(s.Database, s.Logger)
	s.stopoperinstDao = stopoperinst.New(s.Database, s.Logger)

	s.stopEventSubsMap = make(map[string]*StopEventSubscription)
	s.stopOperInsts = make(map[string]struct{})

	err := s.registerScheduler()
	if err != nil {
		s.Logger.Errorf("failed to register scheduler, err: %v", err)
		return fmt.Errorf("failed to register scheduler, err: %w", err)
	}

	return nil
}

func (s *Storage) check() error {
	if s.daoOperinstdata == nil {
		return errors.New("operation instance dao is nil")
	}

	return nil
}

func (s *Storage) registerScheduler() error {
	s.Scheduler = scheduler.NewScheduler(scheduler.WithLogger(s.Logger))
	err := s.Scheduler.RegisterTask(scheduler.NewTask(
		syncOperationTask,
		taskInterval,
		taskTimeout,
		s.syncStopOperInsts,
	))
	if err != nil {
		s.Logger.Errorf("failed to register sync stopping operation inst task, err: %v", err)
		return fmt.Errorf("failed to register sync stopping operation inst task, err: %w", err)
	}

	go s.stopoperinstDao.WatchInsert(func(stopInstID string) {
		s.stopOperInstsMutex.Lock()
		defer s.stopOperInstsMutex.Unlock()
		s.stopOperInsts[stopInstID] = struct{}{}

		go s.checkNotifyStopping(s.Ctx) // nolint: errcheck
	})

	return nil
}

// StopEventSubscription represents the stop event subscription.
type StopEventSubscription struct {
	OperInstID string
	C          chan<- struct{}
}

// GetActionInstanceData gets full action instance data.
func (s *Storage) GetActionInstanceData(ctx context.Context, operationInstanceID, actionName string) (
	*action.InstanceData, error) {

	if ctx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if operationInstanceID == "" {
		return nil, basestorage.ErrEmptyOperaInstID()
	}

	if actionName == "" {
		return nil, basestorage.ErrEmptyActionName()
	}

	return s.daoOperinstdata.GetActionInstData(ctx, operationInstanceID, actionName)
}

// GetActionInstanceLifecycle gets action instance lifecycle.
func (s *Storage) GetActionInstanceLifecycle(ctx context.Context, operationInstanceID, actionName string) (
	*action.Lifecycle, error) {

	if ctx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if operationInstanceID == "" {
		return nil, errors.New("operation instance id is empty")
	}

	if actionName == "" {
		return nil, errors.New("actionName is empty")
	}

	return s.daoOperinstdata.GetActInstLifecycle(ctx, operationInstanceID, actionName)
}

// GetActionInstancePrivateData gets action instance private data.
func (s *Storage) GetActionInstancePrivateData(ctx context.Context, operationInstanceID, actionName string) (
	map[string]any, error) {

	if ctx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if operationInstanceID == "" {
		return nil, errors.New("operation instance id is empty")
	}

	if actionName == "" {
		return nil, errors.New("action name is empty")
	}

	return s.daoOperinstdata.GetActInstPrivateData(ctx, operationInstanceID, actionName)
}

// UpdateActionInstanceLifecycle updates action instance lifecycle.
func (s *Storage) UpdateActionInstanceLifecycle(
	ctx context.Context, operationInstanceID, actionName string, lifecycle *action.Lifecycle) error {

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if operationInstanceID == "" {
		return errors.New("operation instance id is empty")
	}

	if actionName == "" {
		return errors.New("actionName is empty")
	}

	if lifecycle == nil {
		return errors.New("lifecycle is nil")
	}

	if err := s.existsAction(ctx, operationInstanceID, actionName); err != nil {
		return err
	}

	if err := s.daoOperinstdata.UpdateActInstLifecycle(ctx, operationInstanceID, actionName, lifecycle); err != nil {
		return err
	}

	return nil
}

// PushActionInstanceMessage pushes action instance message.
func (s *Storage) PushActionInstanceMessage(
	ctx context.Context, operationInstanceID, actionName string, messages ...action.Message) error {

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if operationInstanceID == "" {
		return errors.New("operation instance id is empty")
	}

	if actionName == "" {
		return errors.New("actionName is empty")
	}

	if len(messages) == 0 {
		return nil
	}

	if err := s.existsAction(ctx, operationInstanceID, actionName); err != nil {
		return err
	}

	for _, msg := range messages {
		if err := s.daoOperinstdata.PushActionInstanceMessage(ctx, operationInstanceID, actionName, msg); err != nil {
			return fmt.Errorf("push action instance msg failed, err(%v)", err)
		}
	}

	return nil
}

// GetOperationInstanceFullData gets full operation instance data.
func (s *Storage) GetOperationInstanceFullData(ctx context.Context, operationInstanceID string) (
	*operation.InstanceData, error) {

	if ctx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if operationInstanceID == "" {
		return nil, errors.New("operation inst id is empty")
	}

	data, err := s.daoOperinstdata.FindOne(ctx, operinstdata.WithOperInstID(operationInstanceID))
	if err != nil {
		return nil, fmt.Errorf("failed to get operation inst data: %v", err)
	}

	return data, nil
}

// GetOperationInstanceBriefData gets brief operation instance data.
func (s *Storage) GetOperationInstanceBriefData(ctx context.Context, operationInstanceID string) (
	*operation.InstanceBriefData, error) {

	if ctx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if operationInstanceID == "" {
		return nil, errors.New("operation inst id is empty")
	}

	data, err := s.daoOperinstdata.FindOne(ctx, operinstdata.WithOperInstID(operationInstanceID))
	if err != nil {
		return nil, fmt.Errorf("failed to get operation inst data: %v", err)
	}

	return &data.InstanceBriefData, nil
}

// ListOperationInstanceBriefData lists operation instance brief data. without action instance data.
func (s *Storage) ListOperationInstanceBriefData(
	ctx context.Context, _ types.Page, condition operation.ListOperationInstanceCondition) (
	[]*operation.InstanceBriefData, int64, error) {

	if ctx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	operInstData, num, err := s.daoOperinstdata.ListWithoutActInst(ctx, types.UnlimitedPage(),
		operinstdata.WithTriggerID(condition.TriggerIDs...), operinstdata.WithState(condition.States...))
	if err != nil {
		return nil, 0, err
	}

	if num == 0 {
		return nil, 0, nil
	}

	return operInstData, num, nil
}

// CountOperationInstance counts operation instance.
func (s *Storage) CountOperationInstance(ctx context.Context, triggerID string,
	states ...operation.State) (int64, error) {

	if ctx == nil {
		return 0, basestorage.ErrNilContent()
	}

	if triggerID == "" {
		return 0, errors.New("trigger id is empty")
	}

	num, err := s.daoOperinstdata.Count(ctx, operinstdata.WithTriggerID(triggerID), operinstdata.WithState(states...))
	if err != nil {
		return 0, fmt.Errorf("failed to count operation instance: %v", err)
	}

	return num, nil
}

// ListOperInstanceBriefByOperation lists operation instance brief data.
func (s *Storage) ListOperInstanceBriefByOperation(
	ctx context.Context, page types.Page, operationID ...string) (
	[]*operation.InstanceBriefData, int64, error) {

	if ctx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	if err := page.Validate(); err != nil {
		return nil, 0, err
	}

	operationData, num, err := s.daoOperinstdata.ListWithoutActInst(ctx, page,
		operinstdata.WithOperationID(operationID...))
	if err != nil {
		return nil, 0, err
	}

	if num == 0 {
		return nil, 0, nil
	}

	return operationData, num, nil
}

// UpsertOperationInstanceData upserts operation instance data.
func (s *Storage) UpsertOperationInstanceData(ctx context.Context,
	operationInstanceData *operation.InstanceData) error {

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if operationInstanceData == nil {
		return basestorage.ErrUpsertNilData()
	}

	if err := s.daoOperinstdata.Upsert(ctx, operationInstanceData); err != nil {
		return fmt.Errorf("failed to update operation inst data, operation-inst(%v), err: %v", operationInstanceData, err)
	}

	return nil
}

// UpdateOperationInstanceLifecycle updates operation instance lifecycle.
func (s *Storage) UpdateOperationInstanceLifecycle(ctx context.Context,
	operationInstanceID string, lifecycle *operation.Lifecycle) error {

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if lifecycle == nil {
		return errors.New("lifecycle is nil")
	}

	if err := s.daoOperinstdata.UpdateLifeCycle(ctx, operationInstanceID, lifecycle); err != nil {
		return err
	}

	return nil
}

// WatchOperInstStopping watches operation instance stopping.
func (s *Storage) WatchOperInstStopping(ctx context.Context, operationInstanceID string) <-chan struct{} {
	channel := make(chan struct{}, 1)
	subscription := &StopEventSubscription{
		OperInstID: operationInstanceID,
		C:          channel,
	}

	subscriptionID := uuid.New().String()
	s.stopEventSubsMapMutex.Lock()
	s.stopEventSubsMap[subscriptionID] = subscription
	s.stopEventSubsMapMutex.Unlock()

	go func() {
		<-ctx.Done()

		s.stopEventSubsMapMutex.Lock()
		delete(s.stopEventSubsMap, subscriptionID)
		s.stopEventSubsMapMutex.Unlock()
	}()

	go func() {
		err := s.checkNotifyStopping(ctx)
		if err != nil {
			s.Logger.Errorf("watch operation instance stopping event succeed, "+
				"but check notify stopping failed, err: %v", err)
		}
	}()

	return channel
}

// MarkOperInstStopping mark task stopping.
func (s *Storage) MarkOperInstStopping(ctx context.Context, operationInstID string) error {
	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	err := s.stopoperinstDao.Upsert(ctx, operationInstID)
	if err != nil {
		return fmt.Errorf("failed to mark operation inst stopping failed, operation-inst-id(%v), err: %v",
			operationInstID, err)
	}

	return nil
}

// syncStopOperInsts sync all stopping operation instances.
func (s *Storage) syncStopOperInsts(ctx context.Context) error {
	stopInstIDs, err := s.stopoperinstDao.FindAll(ctx)
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
func (s *Storage) checkNotifyStopping(ctx context.Context) error {
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
func (s *Storage) processStoppingEvents(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second) // nolint: mnd
	defer cancel()

	notifications := s.getNotifications()

	for _, notify := range notifications {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case notify.Subscription.C <- struct{}{}:
			s.removeSubscription(notify.Key)
		default:
			s.Logger.Errorf("failed to notify stopping event, the channel is full, notify: %+v", notify)
		}
	}

	return nil
}

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

func (s *Storage) removeSubscription(key string) {
	s.stopEventSubsMapMutex.Lock()
	defer s.stopEventSubsMapMutex.Unlock()

	delete(s.stopEventSubsMap, key)
}

// UpdateActionInstanceContent update action instance content.
func (s *Storage) UpdateActionInstanceContent(ctx context.Context, operInstID string, actionName string,
	content map[string]any) error {

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if operInstID == "" {
		return errors.New("operation instance id is empty")
	}

	if actionName == "" {
		return errors.New("actionName is empty")
	}

	if len(content) == 0 {
		return errors.New("content is empty")
	}

	if err := s.existsAction(ctx, operInstID, actionName); err != nil {
		return err
	}

	if err := s.daoOperinstdata.UpdateActionInstContent(ctx, operInstID, actionName, content); err != nil {
		return fmt.Errorf("update action instance content failed, err(%v)", err)
	}

	return nil
}

func (s *Storage) existsAction(ctx context.Context, operInstID string, actionName string) error {
	operation, err := s.daoOperinstdata.FindOneWithoutActionData(ctx, operinstdata.WithOperInstID(operInstID))
	if err != nil {
		return err
	}
	if operation == nil {
		return errors.New("no found operation")
	}

	for _, act := range operation.Metadata.ActionNames {
		if act == actionName {
			return nil
		}
	}

	return errors.New("no found action")
}

// UpsertActionInstancePrivateData upserts action instance private data.
func (s *Storage) UpsertActionInstancePrivateData(
	ctx context.Context, operInstID string, actionName string, privateData map[string]any) error {

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if operInstID == "" {
		return errors.New("operation instance id is empty")
	}

	if actionName == "" {
		return errors.New("actionName is empty")
	}

	if err := s.existsAction(ctx, operInstID, actionName); err != nil {
		return fmt.Errorf("action does not exist, operation-inst-id(%s), action-name(%s), err: %w",
			operInstID, actionName, err)
	}

	if len(privateData) == 0 {
		return nil
	}

	if err := s.daoOperinstdata.PushActInstPrivateData(ctx, operInstID, actionName, privateData); err != nil {
		return fmt.Errorf(
			"failed to update operation instance private data, operation-inst-id(%s), action-name(%s), err: %w",
			operInstID, actionName, err)
	}

	return nil
}
