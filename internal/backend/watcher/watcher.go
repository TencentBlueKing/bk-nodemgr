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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/access"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/cache"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/scheduler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	hostEventChannelSize               = 100
	scheduleInterval                   = 1 * time.Second
	handleWatchEventTimeout            = 30 * time.Second
	cacheExpirationTime                = 2 * time.Minute
	cacheKeyPrefix                     = "bknm:backend:watcher:"
	cacheKeyLocker                     = cacheKeyPrefix + "locker"
	watcherhandleEvnetScheduleTaskName = "watcher_handle_event_schedule_task"
)

// Config defines the configuration of watcher.
type Config struct {
	CmdbHandler cmdb.IHandler
	StorageTopo topoStg.IStorage
	Cache       cache.ICache
	Logger      logger.ILogger
}

// Watcher defines a watcher manager.
type Watcher struct {
	conf Config

	id        string
	mu        sync.Mutex
	scheduler scheduler.Scheduler

	hostEventChannel           chan *types.HostEvent
	waitingCompleteDataHostMap map[int64]*types.Host
}

// NewWatcher creates a new watcher manager.
func NewWatcher(conf Config) (*Watcher, error) {
	watcher := &Watcher{
		conf:                       conf,
		id:                         identifier.GenServiceID(),
		mu:                         sync.Mutex{},
		hostEventChannel:           make(chan *types.HostEvent, hostEventChannelSize),
		waitingCompleteDataHostMap: make(map[int64]*types.Host),
	}

	watcher.scheduler = scheduler.NewScheduler(scheduler.WithLogger(watcher.conf.Logger))

	return watcher, nil
}

// Start starts the watcher manager.
func (w *Watcher) Start(ctx context.Context) error {
	w.conf.Logger.Info("started watcher manager")

	if err := w.registerHostEventScheduler(); err != nil {
		w.conf.Logger.Errorf("register host event scheduler failed: %v", err)
		return fmt.Errorf("register host event scheduler failed: %w", err)
	}

	w.scheduler.Start()

	go func() {
		w.conf.Logger.Info("apply host event started")
		w.applyHostEvent(ctx)
	}()

	return nil
}

// registerHandleEventTask registers the task to handle host events.
func (w *Watcher) registerHostEventScheduler() error {
	err := w.scheduler.RegisterTask(scheduler.NewTask(
		watcherhandleEvnetScheduleTaskName,
		scheduleInterval,
		handleWatchEventTimeout,
		func(ctx context.Context) error {
			if !w.isLock(ctx) {
				w.conf.Logger.Info("not the master, skip handling host event")
				return nil
			}

			tenantIDs := tenant.GetAllTenantIDs()
			gp := gopool.NewPool()
			for _, tenantID := range tenantIDs {
				tenantUserCtx := contextx.NewTenantUserContext(ctx, tenantID, access.GetVirtualUser())
				gp.Go(func() error {
					if err := w.watchHostResource(tenantUserCtx); err != nil {
						w.conf.Logger.Errorf("handle host event failed: %v", err)
						return err
					}

					return nil
				})

				gp.Go(func() error {
					if err := w.watchHostRelationResource(tenantUserCtx); err != nil {
						w.conf.Logger.Errorf("handle host relation event failed: %v", err)
						return err
					}

					return nil
				})
			}

			err := gp.Wait()
			if err != nil {
				w.conf.Logger.Errorf("wait for handle host event failed: %v", err)
				return fmt.Errorf("wait for handle host event failed: %w", err)
			}

			return nil
		},
	))
	if err != nil {
		w.conf.Logger.Errorf("register handle host event task failed: %v", err)
		return fmt.Errorf("register handle host event task failed: %w", err)
	}

	return nil
}

// watchHostResource watches the host resource events.
func (w *Watcher) watchHostResource(ctx contextx.ITenantUserContext) error {
	cursor, err := w.getCursor(ctx, types.HostEventCursor)
	if err != nil {
		w.conf.Logger.Errorf("get host event cursor failed: %v", err)
		return fmt.Errorf("get host event cursor failed: %w", err)
	}

	events, err := w.conf.CmdbHandler.WatchHostResourceEvent(ctx, cursor)
	if err != nil {
		w.conf.Logger.Errorf("watch host resource event failed: %v", err)
		return fmt.Errorf("watch host resource event failed: %w", err)
	}

	if len(events) == 0 {
		w.conf.Logger.Info("no host resource event found")
		return nil
	}

	for _, ev := range events {
		select {
		case w.hostEventChannel <- ev:
			w.conf.Logger.Debugf("host event sent to channel: %s", ev.Cursor)
		default:
			w.conf.Logger.Warnf("host event channel is full, dropping event: %s", ev.Cursor)
		}
	}

	if err := w.setCursor(ctx, types.HostEventCursor, events[len(events)-1].Cursor); err != nil {
		w.conf.Logger.Errorf("set host event cursor failed: %v", err)
		return fmt.Errorf("set host event cursor failed: %w", err)
	}

	return nil
}

// watchHostRelationResource watches the host relation resource events.
func (w *Watcher) watchHostRelationResource(ctx contextx.ITenantUserContext) error {
	cursor, err := w.getCursor(ctx, types.HostRelationEventCursor)
	if err != nil {
		w.conf.Logger.Errorf("get host event cursor failed: %v", err)
		return fmt.Errorf("get host event cursor failed: %w", err)
	}

	events, err := w.conf.CmdbHandler.WatchHostRelationResourceEvent(ctx, cursor)
	if err != nil {
		w.conf.Logger.Errorf("watch host relation resource event failed: %v", err)
		return fmt.Errorf("watch host relation resource event failed: %w", err)
	}

	if len(events) == 0 {
		w.conf.Logger.Info("no host relation resource event found")
		return nil
	}

	for _, ev := range mergeHostEvent(events...) {
		select {
		case w.hostEventChannel <- ev:
			w.conf.Logger.Debugf("host relation event sent to channel: %v", ev)
		default:
			w.conf.Logger.Warnf("host relation event channel is full, dropping event: %v", ev)
		}
	}

	if err := w.setCursor(ctx, types.HostRelationEventCursor, events[len(events)-1].Cursor); err != nil {
		w.conf.Logger.Errorf("set host relation event cursor failed: %v", err)
		return fmt.Errorf("set host relation event cursor failed: %w", err)
	}

	return nil
}

// applyHostEvent applies the host event to the watcher.
func (w *Watcher) applyHostEvent(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			w.conf.Logger.Info("apply host event context done")
		case event := <-w.hostEventChannel:
			switch event.Resource {
			case types.ResourceTypeHost:
				if err := w.handleHostResource(ctx, event); err != nil {
					w.conf.Logger.Errorf("handle host resource failed: %v", err)
				}
			case types.ResourceTypeHostRelation:
				if err := w.handleHostRelationResource(ctx, event); err != nil {
					w.conf.Logger.Errorf("handle host relation resource failed: %v", err)
				}
			default:
				w.conf.Logger.Warnf("unknown resource type: %s", event.Resource)
			}
		default:
			w.conf.Logger.Debug("no host event to apply")
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
	case types.EventTypeBlank:
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
	case types.EventTypeBlank:
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
		if event.Detail == nil {
			continue
		}

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

// getCursor retrieves the cursor for the given key from the cache.
func (w *Watcher) getCursor(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", errors.New("get cursor from cache, key cannot be empty")
	}

	key = cacheKeyPrefix + key
	exist, err := w.conf.Cache.Exists(ctx, key)
	if err != nil {
		return "", fmt.Errorf("check cursor exist failed: %w", err)
	}

	if !exist {
		return "", nil
	}

	result, err := w.conf.Cache.Get(ctx, key)
	if err != nil {
		return "", fmt.Errorf("get cursor failed: %w", err)
	}

	return string(result), nil
}

// setCursor sets the cursor for the given key in the cache.
func (w *Watcher) setCursor(ctx context.Context, key, value string) error {
	if key == "" {
		return errors.New("set cursor to cache, key cannot be empty")
	}

	key = cacheKeyPrefix + key
	if err := w.conf.Cache.SetWithExpiration(ctx, key, []byte(value), cacheExpirationTime); err != nil {
		return fmt.Errorf("set cursor failed: %w", err)
	}

	return nil
}

// isLock checks if the watcher is the master by trying to acquire a lock.
func (w *Watcher) isLock(ctx context.Context) bool {
	exist, err := w.conf.Cache.Exists(ctx, cacheKeyLocker)
	if err != nil {
		w.conf.Logger.Errorf("check locker failed: %v", err)
		return false
	}

	if exist {
		id, err := w.conf.Cache.Get(ctx, cacheKeyLocker)
		if err != nil {
			w.conf.Logger.Errorf("get locker id failed: %v", err)
			return false
		}

		if string(id) != w.id {
			return false
		}

		err = w.conf.Cache.SetWithExpiration(ctx, cacheKeyLocker, []byte(w.id), cacheExpirationTime)
		if err != nil {
			w.conf.Logger.Errorf("set locker by id(%s) failed: %v", w.id, err)
			return false
		}

		return true
	}

	locked, err := w.conf.Cache.SetNXWithExpiration(ctx, cacheKeyLocker, []byte(w.id), cacheExpirationTime)
	if err != nil {
		w.conf.Logger.Errorf("set master lock failed: %v", err)
		return false
	}

	return locked
}
