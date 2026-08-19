/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package workflow

import (
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/scheduler"
	"github.com/google/uuid"
)

const (
	stopPollInterval  = time.Second
	stopPollTimeout   = 2 * time.Second
	pollOperationTask = "poll stopping operation inst"
)

// registerStopOperInstTask registers the stop operation instance task.
func (s *Storage) registerStopOperInstTask() error {
	if s.Scheduler == nil {
		return errors.New("scheduler is not initialized")
	}

	err := s.Scheduler.RegisterTask(scheduler.NewTask(
		syncOperationTask,
		taskInterval,
		taskTimeout,
		s.syncStopOperInsts,
	))
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to register sync stopping operation instance task")

		return fmt.Errorf("failed to register sync stopping operation instance task: %w", err)
	}

	err = s.Scheduler.RegisterTask(scheduler.NewTask(
		pollOperationTask,
		stopPollInterval,
		stopPollTimeout,
		s.syncSubscribedStopOperInsts,
	))
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to register poll stopping operation instance task")

		return fmt.Errorf("failed to register poll stopping operation instance task: %w", err)
	}

	return nil
}

// StopEventSubscription represents the stop event subscription.
type StopEventSubscription struct {
	OperInstID string
	C          chan<- struct{}
}

// watchOperInstStopping watches operation instance stopping.
func (s *Storage) watchOperInstStopping(nCtx contextx.IContext, operInstID string) <-chan struct{} {
	channel := make(chan struct{}, 1)
	subscription := &StopEventSubscription{
		OperInstID: operInstID,
		C:          channel,
	}

	subscriptionID := uuid.New().String()
	s.stopEventSubsMapMutex.Lock()
	s.stopEventSubsMap[subscriptionID] = subscription
	s.stopEventSubsMapMutex.Unlock()

	go func() {
		<-nCtx.Done()

		s.stopEventSubsMapMutex.Lock()
		delete(s.stopEventSubsMap, subscriptionID)
		s.stopEventSubsMapMutex.Unlock()
	}()

	go func() {
		err := s.checkNotifyStopping(nCtx)
		if err != nil {
			logger.G.Sys().WithErr(err).Error("watch operation instance stopping event succeed, but check notify stopping failed")
		}
	}()

	return channel
}

// syncStopOperInsts sync all stopping operation instances.
func (s *Storage) syncStopOperInsts(nCtx contextx.IContext) error {
	stopInstIDs, err := s.daoStopOperInst.FindAll(nCtx)
	if err != nil {
		return fmt.Errorf("failed to find all stopping operation instances: %w", err)
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

func (s *Storage) syncSubscribedStopOperInsts(nCtx contextx.IContext) error {
	operInstIDs := s.subscribedOperInstIDs()
	if len(operInstIDs) == 0 {
		return nil
	}

	stopInstIDs, err := s.daoStopOperInst.FindByIDs(nCtx, operInstIDs...)
	if err != nil {
		return fmt.Errorf("failed to find subscribed stopping operation instances: %w", err)
	}

	if len(stopInstIDs) == 0 {
		return nil
	}

	s.addStopOperInsts(stopInstIDs...)

	return s.checkNotifyStopping(nCtx)
}

func (s *Storage) subscribedOperInstIDs() []string {
	s.stopEventSubsMapMutex.RLock()

	operInstIDSet := make(map[string]struct{}, len(s.stopEventSubsMap))
	for _, subscription := range s.stopEventSubsMap {
		operInstIDSet[subscription.OperInstID] = struct{}{}
	}
	s.stopEventSubsMapMutex.RUnlock()

	return conv.MapKeyToSlice(operInstIDSet)
}

func (s *Storage) addStopOperInsts(operInstIDs ...string) {
	s.stopOperInstsMutex.Lock()
	defer s.stopOperInstsMutex.Unlock()

	for _, operInstID := range operInstIDs {
		s.stopOperInsts[operInstID] = struct{}{}
	}
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

// processStoppingEvents process stopping events.
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

// upsertNeedStopOperInst upserts need stop operation instance.
func (s *Storage) upsertNeedStopOperInst(nCtx contextx.IContext, operInstID string) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if operInstID == "" {
		return errors.New("operation instance id is empty")
	}

	err := s.daoStopOperInst.Upsert(nCtx, operInstID)
	if err != nil {
		return fmt.Errorf("failed to upsert stop operation instance, operInstID(%s): %w", operInstID, err)
	}

	s.addStopOperInsts(operInstID)

	if err := s.checkNotifyStopping(nCtx); err != nil {
		logger.G.Sys().WithErr(err).Warn("failed to notify local stopping operation instance")
	}

	return nil
}
