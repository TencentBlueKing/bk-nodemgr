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
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/scheduler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	// HandleResourceEvent is the watcher handle event.
	HandleResourceEvent = "handle_resource_event"
)

// Config defines the configuration of watcher.
type Config struct {
	CmdbHandler cmdb.IHandler
	TopoStorage topoStg.IStorage
	Manager     manager.Manager
}

// Watcher defines a watcher manager.
type Watcher struct {
	// config.
	conf Config

	// logger.
	logger logger.Logger

	// scheduler.
	scheduler scheduler.Scheduler

	// mutex
	mu sync.Mutex

	// biz id set.
	needSyncBizIDSet map[int64]struct{}
}

// NewWatcher creates a new watcher manager.
func NewWatcher(conf Config, logger logger.Logger) (*Watcher, error) {
	return &Watcher{
		conf:             conf,
		logger:           logger,
		scheduler:        scheduler.NewScheduler(scheduler.WithLogger(logger)),
		needSyncBizIDSet: make(map[int64]struct{}),
	}, nil
}

// Start starts the watcher manager.
func (w *Watcher) Start(ctx context.Context) error {
	w.logger.Info("started watcher manager")

	watcher, err := w.conf.CmdbHandler.NewWatcher()
	if err != nil {
		w.logger.Errorf("failed to create a new watcher, error: %v", err)
		return err
	}

	hostChannel, err := watcher.WatchHost()
	if err != nil {
		w.logger.Errorf("watch host failed, error: %v", err)
		return err
	}

	hostRelChannel, err := watcher.WatchHostRelation()
	if err != nil {
		w.logger.Errorf("watch host relation failed, error: %v", err)
		return err
	}

	err = watcher.Start(ctx)
	if err != nil {
		w.logger.Errorf("start watcher failed, error: %v", err)
		return err
	}

	err = w.registerHandleEventTask(ctx)
	if err != nil {
		w.logger.Errorf("register handle event task failed, error: %v", err)
		return err
	}

	go w.collectEvents(ctx, hostChannel, hostRelChannel)

	return nil
}

// collectEvents collects events from the watcher.
func (w *Watcher) collectEvents(ctx context.Context, hostChannel <-chan *types.ChangeEvent[*types.Host],
	hostRelChannel <-chan *types.ChangeEvent[*types.HostRel]) {

	for {
		select {
		case event := <-hostChannel:
			switch event.ChangeType {
			// Only handle update events which do not involve changes in biz IDs
			case types.ChangeTypeUpdate:
				w.logger.Infof("update host event: %v", event)
				host, err := w.conf.TopoStorage.GetHostByID(ctx, event.Detail.HostID)
				if err != nil {
					w.logger.Errorf("get host by id failed, error: %v", err)
					continue
				}

				w.mu.Lock()
				w.needSyncBizIDSet[host.Static.BizID] = struct{}{}
				w.mu.Unlock()
			default:
				continue
			}
		case event := <-hostRelChannel:
			switch event.ChangeType {
			// Only handle create and delete event which indicates a change in the host under the biz
			case types.ChangeTypeCreate, types.ChangeTypeDelete:
				w.logger.Infof("host relation event: %v", event)

				w.mu.Lock()
				w.needSyncBizIDSet[event.Detail.BizID] = struct{}{}
				w.mu.Unlock()
			default:
				continue
			}
		case <-ctx.Done():
			w.logger.Info("stopped watcher manager")
			return
		}
	}
}

// registerHandleEventTask registers the handle event task.
func (w *Watcher) registerHandleEventTask(ctx context.Context) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		w.logger.Errorf("failed to get tenant id from context: %v", err)
		return err
	}

	w.scheduler.RegisterTask(&scheduler.Task{
		ID:       HandleResourceEvent,
		Interval: 10 * time.Second, // nolint: mnd
		Timeout:  time.Minute,
		Fn: func(ctx context.Context) error {
			w.mu.Lock()
			bizSet := conv.MapKeyToSlice(w.needSyncBizIDSet)
			// clean up the need sync biz id set
			w.needSyncBizIDSet = make(map[int64]struct{})
			w.mu.Unlock()

			for _, bizID := range bizSet {
				triggerID, err := w.conf.Manager.Execute(ctx, &syncdata.OperSyncHost{
					TenantID: tenantID,
					BizID:    bizID,
				})

				if err != nil {
					w.logger.Errorf("failed to start sync cmdb host operation, biz id: %v, err: %v", bizID, err)
					continue
				}

				w.logger.Infof("start sync cmdb host operation, trigger id: %s, biz id: %d", triggerID, bizID)
			}

			return nil
		},
	})

	return nil
}
