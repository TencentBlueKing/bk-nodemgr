/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package watcher provides the internal watcher.
package watcher

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/globalsettings"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/locker"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/scheduler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	// HandleHostEvent is the watcher handle host resource.
	HandleHostEvent = "handle_host_event"

	// ScheduleInterval is the interval for the watcher scheduler.
	ScheduleInterval = 1 * time.Second

	// HandleWatchEventTimeout is the timeout for handling watch events.
	HandleWatchEventTimeout = 30 * time.Second

	// WatcherLockKey is the key for the watcher lock.
	WatcherLockKey = "watcher_lock_key"

	// HostEventChannelSize is the size of the host event channel.
	HostEventChannelSize = 100
)

// Config defines the configuration of watcher.
type Config struct {
	CmdbHandler cmdb.IHandler
	StorageTopo topoStg.IStorage
}

// Watcher defines a watcher manager.
type Watcher struct {
	conf Config

	id            string
	logger        logger.Logger
	mu            sync.Mutex
	redsynclocker locker.MutexFactory
	scheduler     scheduler.Scheduler

	hostEventChannel           chan *types.HostEvent
	waitingCompleteDataHostMap map[int64]*types.Host
}

// OptionFunc defines the function type for options.
type OptionFunc func(*Watcher)

// WithLogger sets the logger for the watcher.
func WithLogger(logger logger.Logger) OptionFunc {
	return func(w *Watcher) {
		w.logger = logger
	}
}

// WithDistributedLocker sets the distributed locker for the watcher.
func WithDistributedLocker(locker locker.MutexFactory) OptionFunc {
	return func(w *Watcher) {
		w.redsynclocker = locker
	}
}

// NewWatcher creates a new watcher manager.
func NewWatcher(conf Config, opts ...OptionFunc) (*Watcher, error) {
	watcher := &Watcher{
		conf:                       conf,
		id:                         identifier.GenServiceID(),
		logger:                     logger.LoggerDefault{},
		mu:                         sync.Mutex{},
		redsynclocker:              nil,
		hostEventChannel:           make(chan *types.HostEvent, HostEventChannelSize),
		waitingCompleteDataHostMap: make(map[int64]*types.Host),
	}

	for _, opt := range opts {
		opt(watcher)
	}

	watcher.scheduler = scheduler.NewScheduler(scheduler.WithLogger(watcher.logger))

	if watcher.redsynclocker == nil {
		return nil, errors.New("redsync locker is required")
	}

	return watcher, nil
}

// Start starts the watcher manager.
func (w *Watcher) Start(ctx context.Context) error {
	w.logger.Info("started watcher manager")
	err := globalsettings.GetInstance().DeleteMany(ctx, types.HostEventCursor, types.HostRelationEventCursor)
	if err != nil {
		w.logger.Errorf("delete global settings failed: %v", err)
		return err
	}

	if err := w.registerHostEventScheduler(); err != nil {
		w.logger.Errorf("register host event scheduler failed: %v", err)
		return fmt.Errorf("register host event scheduler failed: %w", err)
	}

	w.scheduler.Start()

	go func() {
		w.logger.Info("apply host event started")
		w.applyHostEvent(ctx)
	}()

	return nil
}

// registerHandleEventTask registers the task to handle host events.
func (w *Watcher) registerHostEventScheduler() error {
	err := w.scheduler.RegisterTask(scheduler.NewTask(
		HandleHostEvent,
		ScheduleInterval,
		HandleWatchEventTimeout,
		func(ctx context.Context) error {
			err := w.redsynclocker.NewMutex(WatcherLockKey).TryLock()
			if err != nil {
				w.logger.Warnf("try acquire distributed lock failed, watcher(%s) is not the main watcher", w.id)
				return err
			}

			gp := gopool.NewPool()
			gp.Go(func() error {
				if err := w.watchHostResource(ctx); err != nil {
					w.logger.Errorf("handle host event failed: %v", err)
					return err
				}

				return nil
			})

			gp.Go(func() error {
				if err := w.watchHostRelationResource(ctx); err != nil {
					w.logger.Errorf("handle host relation event failed: %v", err)
					return err
				}

				return nil
			})

			_ = gp.Wait()

			return nil
		},
	))
	if err != nil {
		w.logger.Errorf("register handle host event task failed: %v", err)
		return fmt.Errorf("register handle host event task failed: %w", err)
	}

	return nil
}

// watchHostResource watches the host resource events.
func (w *Watcher) watchHostResource(ctx context.Context) error {
	// TODO: use cache to store cursor instead of global settings
	cursor, err := globalsettings.GetInstance().Get(ctx, types.HostEventCursor)
	if err != nil {
		w.logger.Errorf("get host event cursor failed: %v", err)
		return fmt.Errorf("get host event cursor failed: %w", err)
	}

	events, err := w.conf.CmdbHandler.WatchHostResourceEvent(ctx, cursor)
	if err != nil {
		w.logger.Errorf("watch host resource event failed: %v", err)
		return fmt.Errorf("watch host resource event failed: %w", err)
	}

	if len(events) == 0 {
		w.logger.Info("no host resource event found")
		return nil
	}

	for _, ev := range events {
		select {
		case w.hostEventChannel <- ev:
			w.logger.Debugf("host event sent to channel: %s", ev.Cursor)
		default:
			w.logger.Warnf("host event channel is full, dropping event: %s", ev.Cursor)
		}
	}

	err = globalsettings.GetInstance().UpsertMany(ctx, &types.GlobalSettings{
		SettingName: types.HostEventCursor,
		Value:       events[len(events)-1].Cursor,
	})
	if err != nil {
		w.logger.Errorf("upsert host event cursor failed: %v", err)
		return fmt.Errorf("upsert host event cursor failed: %w", err)
	}

	return nil
}

// watchHostRelationResource watches the host relation resource events.
func (w *Watcher) watchHostRelationResource(ctx context.Context) error {
	// TODO: use cache to store cursor instead of global settings
	cursor, err := globalsettings.GetInstance().Get(ctx, types.HostRelationEventCursor)
	if err != nil {
		w.logger.Errorf("get host relation event cursor failed: %v", err)
		return fmt.Errorf("get host relation event cursor failed: %w", err)
	}

	events, err := w.conf.CmdbHandler.WatchHostRelationResourceEvent(ctx, cursor)
	if err != nil {
		w.logger.Errorf("watch host relation resource event failed: %v", err)
		return fmt.Errorf("watch host relation resource event failed: %w", err)
	}

	if len(events) == 0 {
		w.logger.Info("no host relation resource event found")
		return nil
	}

	for _, ev := range mergeHostEvent(events...) {
		select {
		case w.hostEventChannel <- ev:
			w.logger.Debugf("host relation event sent to channel: %v", ev)
		default:
			w.logger.Warnf("host relation event channel is full, dropping event: %v", ev)
		}
	}

	err = globalsettings.GetInstance().UpsertMany(ctx, &types.GlobalSettings{
		SettingName: types.HostRelationEventCursor,
		Value:       events[len(events)-1].Cursor,
	})
	if err != nil {
		w.logger.Errorf("upsert host event cursor failed: %v", err)
		return fmt.Errorf("upsert host event cursor failed: %w", err)
	}

	return nil
}

// applyHostEvent applies the host event to the watcher.
func (w *Watcher) applyHostEvent(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			w.logger.Info("apply host event context done")
		case event := <-w.hostEventChannel:
			switch event.Resource {
			case types.ResourceTypeHost:
				if err := w.handleHostResource(ctx, event); err != nil {
					w.logger.Errorf("handle host resource failed: %v", err)
				}
			case types.ResourceTypeHostRelation:
				if err := w.handleHostRelationResource(ctx, event); err != nil {
					w.logger.Errorf("handle host relation resource failed: %v", err)
				}
			default:
				w.logger.Warnf("unknown resource type: %s", event.Resource)
			}
		default:
			w.logger.Debug("no host event to apply")
		}
	}
}

// handleHostResource this func defines how to handle the host resource event.
func (w *Watcher) handleHostResource(ctx context.Context, event *types.HostEvent) error {
	switch event.EventType {
	case types.EventTypeCreate:
		host, ok := w.waitingCompleteDataHostMap[event.Detail.HostID]
		if !ok {
			w.mu.Lock()
			defer w.mu.Unlock()
			// when the host synchronizes from the CMDB for the first time, the agentid needs to be updated to dynamic
			event.Detail.Dynamic.AgentID = event.Detail.Static.SyncedAgentID
			w.waitingCompleteDataHostMap[event.Detail.HostID] = event.Detail

			return nil
		}

		event.Detail.Static.BizID = host.Static.BizID
		err := w.conf.StorageTopo.UpsertManyHost(ctx, event.Detail)
		if err != nil {
			return fmt.Errorf("UpsertManyHost failed: %w", err)
		}

		w.mu.Lock()
		defer w.mu.Unlock()
		delete(w.waitingCompleteDataHostMap, event.Detail.HostID)

		return nil
	case types.EventTypeUpdate:
		dbHost, err := w.conf.StorageTopo.GetHostByID(ctx, event.Detail.HostID)
		if err != nil {
			return fmt.Errorf("GetHostByID failed: %w", err)
		}

		event.Detail.Static.BizID = dbHost.Static.BizID
		err = w.conf.StorageTopo.UpsertManyHostStatic(ctx, event.Detail)
		if err != nil {
			return fmt.Errorf("UpsertManyHostStatic failed: %w", err)
		}

		return nil
	case types.EventTypeDelete:
		err := w.conf.StorageTopo.DeleteManyHost(ctx, event.Detail.HostID)
		if err != nil {
			return fmt.Errorf("DeleteManyHost failed: %w", err)
		}

		return nil
	default:
		return fmt.Errorf("unknown event type: %s", event.EventType)
	}
}

// handleHostRelationResource this func defines how to handle the host relation resource event.
func (w *Watcher) handleHostRelationResource(ctx context.Context, event *types.HostEvent) error {
	switch event.EventType {
	case types.EventTypeCreate:
		host, ok := w.waitingCompleteDataHostMap[event.Detail.HostID]
		if !ok {
			w.mu.Lock()
			defer w.mu.Unlock()
			w.waitingCompleteDataHostMap[event.Detail.HostID] = event.Detail

			return nil
		}

		host.Static.BizID = event.Detail.Static.BizID
		err := w.conf.StorageTopo.UpsertManyHost(ctx, host)
		if err != nil {
			return fmt.Errorf("UpsertManyHost failed: %w", err)
		}

		w.mu.Lock()
		defer w.mu.Unlock()
		delete(w.waitingCompleteDataHostMap, event.Detail.HostID)

		return nil
	case types.EventTypeUpdate:
		dbHost, err := w.conf.StorageTopo.GetHostByID(ctx, event.Detail.HostID)
		if err != nil {
			return fmt.Errorf("GetHostByID failed: %w", err)
		}

		dbHost.Static.BizID = event.Detail.Static.BizID
		err = w.conf.StorageTopo.UpsertManyHostStatic(ctx, dbHost)
		if err != nil {
			return fmt.Errorf("UpsertManyHostStatic failed: %w", err)
		}

		return nil
	case types.EventTypeDelete:
		err := w.conf.StorageTopo.DeleteManyHost(ctx, event.Detail.HostID)
		if err != nil {
			return fmt.Errorf("DeleteManyHost failed: %w", err)
		}

		return nil
	default:
		return fmt.Errorf("unknown event type: %s", event.EventType)
	}
}

// mergeHostEvent this func defines how to merge host events.
func mergeHostEvent(events ...*types.HostEvent) []*types.HostEvent {
	if len(events) == 0 {
		return nil
	}

	pendingProcess := make(map[int64]*types.HostEvent, 0)
	result := make([]*types.HostEvent, 0, len(events))
	for _, event := range events {
		ev, ok := pendingProcess[event.Detail.HostID]
		if !ok {
			pendingProcess[event.Detail.HostID] = event
			continue
		}

		if ev.EventType == types.EventTypeDelete && event.EventType == types.EventTypeCreate {
			event.EventType = types.EventTypeUpdate
			pendingProcess[event.Detail.HostID] = event

			continue
		}
	}

	for _, ev := range pendingProcess {
		result = append(result, ev)
	}

	return result
}
