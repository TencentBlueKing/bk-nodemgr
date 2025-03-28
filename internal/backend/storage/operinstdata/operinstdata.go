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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/operinstdata"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/stopoperinst"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/scheduler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/singleflight"
)

// StorageName ...
const StorageName = "operinstdata"

// NewStorage ...
func NewStorage(client *mongo.Client, database string, logger logger.Logger) (IStorage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &storage{
		Storage: base.Storage{
			Name:     StorageName,
			Database: client.Database(database),
			Logger:   logger,
		},
	}

	err := base.InitStorage(&s.Storage,
		base.WithStartFunc(s.startFn),
		base.WithCheckFunc(s.check))
	if err != nil {
		s.Logger.Errorf("new storage failed, err: %v", err)
		return nil, err
	}

	return s, nil
}

type storage struct {
	base.Storage

	// dao
	operinstdataDao operinstdata.IHandler
	stopoperinstDao stopoperinst.Handler

	// stop event subscriptions
	stopEventSubsMap      map[string]*StopEventSubscription
	stopEventSubsMapMutex sync.RWMutex

	stopOperInsts      map[string]struct{}
	stopOperInstsMutex sync.RWMutex

	sg singleflight.Group
}

func (s *storage) startFn() error {
	s.operinstdataDao = operinstdata.New(s.Database, s.Logger)
	s.stopoperinstDao = stopoperinst.New(s.Database, s.Logger)

	s.stopEventSubsMap = make(map[string]*StopEventSubscription)
	s.stopOperInsts = make(map[string]struct{})

	s.registerScheduler()

	return nil
}

func (s *storage) check() error {
	if s.operinstdataDao == nil {
		return errors.New("operation instance dao is nil")
	}

	return nil
}

func (s *storage) registerScheduler() {
	s.Scheduler = scheduler.NewScheduler(scheduler.WithLogger(s.Logger), scheduler.WithInterval(time.Second*5))
	s.Scheduler.RegisterTask(&scheduler.Task{
		ID:       "sync stopping operation inst",
		Interval: 10 * time.Second,
		Timeout:  20 * time.Second,
		Fn:       s.syncStopOperInsts,
	})

	go s.stopoperinstDao.WatchInsert(func(stopInstID string) {
		s.stopOperInstsMutex.Lock()
		defer s.stopOperInstsMutex.Unlock()
		s.stopOperInsts[stopInstID] = struct{}{}

		go s.checkNotifyStopping(s.Ctx)
	})
}

// StopEventSubscription represents the stop event subscription.
type StopEventSubscription struct {
	OperInstID string
	C          chan<- struct{}
}

// GetOperInstData get task data.
func (s *storage) GetOperInstData(ctx context.Context, operInstID string) (*operengine.OperInstData, error) {
	if ctx == nil {
		return nil, base.ErrNilContent()
	}

	if operInstID == "" {
		return nil, errors.New("operation inst id is empty")
	}

	data, err := s.operinstdataDao.FindOne(ctx, operinstdata.WithOperInstID(operInstID))
	if err != nil {
		return nil, fmt.Errorf("failed to get operation inst data: %v", err)
	}

	return data, nil
}

// UpsertOperInstData update task data.
func (s *storage) UpsertOperInstData(ctx context.Context, data *operengine.OperInstData) error {
	if ctx == nil {
		return base.ErrNilContent()
	}

	if data == nil {
		return base.ErrUpsertNilData()
	}

	if err := s.operinstdataDao.Upsert(ctx, data); err != nil {
		return fmt.Errorf("failed to update operation inst data, operation-inst(%v), err: %v", data, err)
	}

	return nil
}

// MarkOperInstStopping mark task stopping.
func (s *storage) MarkOperInstStopping(ctx context.Context, operationInstID string) error {
	if ctx == nil {
		return base.ErrNilContent()
	}

	err := s.stopoperinstDao.Upsert(ctx, operationInstID)
	if err != nil {
		return fmt.Errorf("failed to mark operation inst stopping failed, operation-inst-id(%v), err: %v",
			operationInstID, err)
	}

	return nil
}

// WatchOperInstStopping watch operation instance stopping event.
func (s *storage) WatchOperInstStopping(ctx context.Context, operInstID string) <-chan struct{} {
	c := make(chan struct{}, 1)
	subscription := &StopEventSubscription{
		OperInstID: operInstID,
		C:          c,
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

	return c
}

// syncStopOperInsts sync all stopping operation instances.
func (s *storage) syncStopOperInsts(ctx context.Context) error {
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
func (s *storage) checkNotifyStopping(ctx context.Context) error {
	_, err, _ := s.sg.Do("checkNotifyStopping", func() (interface{}, error) {
		err := s.processStoppingEvents(ctx)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// TODO: 此处有坑，需要重新测试
func (s *storage) processStoppingEvents(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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
func (s *storage) getNotifications() []notifyItem {
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

func (s *storage) removeSubscription(key string) {
	s.stopEventSubsMapMutex.Lock()
	defer s.stopEventSubsMapMutex.Unlock()

	delete(s.stopEventSubsMap, key)
}

// UpdateActInstLifecycle update operation instance's action instance lifecycle.
func (s *storage) UpdateActInstLifecycle(ctx context.Context, operInstID string, actionName string,
	lifecycle *operengine.ActInstLifeCycle) error {

	if ctx == nil {
		return base.ErrNilContent()
	}

	if operInstID == "" {
		return errors.New("operation instance id is empty")
	}

	if actionName == "" {
		return errors.New("actionName is empty")
	}

	if lifecycle == nil {
		return errors.New("lifecycle is nil")
	}

	if err := s.operinstdataDao.UpdateActInstLifecycle(ctx, operInstID, actionName, lifecycle); err != nil {
		return err
	}

	return nil
}

// UpdateLifecycle update operation instance's lifecycle.
func (s *storage) UpdateLifecycle(ctx context.Context, operInstID string, lifecycle *operengine.Lifecycle) error {
	if ctx == nil {
		return base.ErrNilContent()
	}

	if lifecycle == nil {
		return errors.New("lifecycle is nil")
	}

	if err := s.operinstdataDao.UpdateLifecycle(ctx, operInstID, lifecycle); err != nil {
		return err
	}

	return nil
}

// GetOperInstDataWithoutActionData find one operation instance data.
func (s *storage) GetOperInstDataWithoutActionData(ctx context.Context, operInstID string,
) (*operengine.OperInstData, error) {

	if ctx == nil {
		return nil, base.ErrNilContent()
	}

	if operInstID == "" {
		return nil, errors.New("operation instance id is empty")
	}

	operInstData, err := s.operinstdataDao.FindOneWithoutActionData(ctx, operinstdata.WithOperInstID(operInstID))
	if err != nil {
		return nil, err
	}

	return operInstData, nil
}

// GetActionInstData find one action instance data.
func (s *storage) GetActionInstData(ctx context.Context, operInstID string,
	actionName string) (*operengine.ActionInstData, error) {

	if ctx == nil {
		return nil, base.ErrNilContent()
	}

	if operInstID == "" {
		return nil, errors.New("operation instance id is empty")
	}

	if actionName == "" {
		return nil, errors.New("actionName is empty")
	}

	return s.operinstdataDao.GetActionInstData(ctx, operInstID, actionName)
}

// GetActInstLifecycle get action instance's lifecycle.
func (s *storage) GetActInstLifecycle(ctx context.Context, operInstID string, actionName string) (
	*operengine.ActInstLifeCycle, error) {

	if ctx == nil {
		return nil, base.ErrNilContent()
	}

	if operInstID == "" {
		return nil, errors.New("operation instance id is empty")
	}

	if actionName == "" {
		return nil, errors.New("actionName is empty")
	}

	return s.operinstdataDao.GetActInstLifecycle(ctx, operInstID, actionName)
}

// PushActInstMsgs push action instance msgs.
func (s *storage) PushActInstMsgs(ctx context.Context, operInstID string, actionName string,
	msgs ...operengine.Message) error {

	if ctx == nil {
		return base.ErrNilContent()
	}

	if operInstID == "" {
		return errors.New("operation instance id is empty")
	}

	if actionName == "" {
		return errors.New("actionName is empty")
	}

	if len(msgs) == 0 {
		return nil
	}

	for _, msg := range msgs {
		if err := s.operinstdataDao.PushActInstMsgs(ctx, operInstID, actionName, msg); err != nil {
			return fmt.Errorf("push action instance msg failed, err(%v)", err)
		}
	}

	return nil
}

// UpdateActionInstContent update action instance content.
func (s *storage) UpdateActionInstContent(ctx context.Context, operInstID string, actionName string,
	content map[string]any) error {

	if ctx == nil {
		return base.ErrNilContent()
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

	if err := s.operinstdataDao.UpdateActionInstContent(ctx, operInstID, actionName, content); err != nil {
		return fmt.Errorf("update action instance content failed, err(%v)", err)
	}

	return nil
}
