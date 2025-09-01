/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package syncdata

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/cache"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameWatchAndApplyCMDBResource defines the action name.
	ActionNameWatchAndApplyCMDBResource = "watch_and_apply_cmdb_resource"

	cacheExpirationTime = 2 * time.Minute
	cacheKeyPrefix      = "bknm:backend:cc:resource:cursor:"
)

// NewActionWatchCMDBResource creates a new action to watch CMDB resource changes.
func NewActionWatchCMDBResource(
	cache cache.ICache,
	cmdbHandler cmdb.IHandler,
	storageTopo topo.IStorage,
) action.Definition {

	return &actionWatchCMDBResource{
		cache:       cache,
		cmdbHandler: cmdbHandler,
		storageTopo: storageTopo,

		mu:                         sync.Mutex{},
		pendingProcessEvents:       make([]*types.HostEvent, 0),
		waitingCompleteDataHostMap: make(map[int64]*types.Host),
	}
}

// WatchCMDBResourceParam ...
type WatchCMDBResourceParam struct {
	TenantID string `json:"tenant_id"`
	Operator string `json:"operator"`
}

// actionWatchCMDBResource implements the action.Definition interface.
type actionWatchCMDBResource struct {
	cache       cache.ICache
	cmdbHandler cmdb.IHandler
	storageTopo topo.IStorage

	mu                         sync.Mutex
	pendingProcessEvents       []*types.HostEvent
	waitingCompleteDataHostMap map[int64]*types.Host
}

// Name returns the name of the action.
func (act *actionWatchCMDBResource) Name() string {
	return ActionNameWatchAndApplyCMDBResource
}

// Version returns the version of the action.
func (act *actionWatchCMDBResource) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionWatchCMDBResource) Description() string {
	return "watch and apply cmdb resource changes."
}

// Timeout returns the timeout of the action.
func (act *actionWatchCMDBResource) Timeout() time.Duration {
	return time.Second * 30 // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionWatchCMDBResource) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount this action creates a large number of synchronization tasks,
// therefore does not allow the system to automatically retry.
func (act *actionWatchCMDBResource) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionWatchCMDBResource) DelayFn() func() {
	return func() {}
}

// Do this func define what the action will do.
func (act *actionWatchCMDBResource) Do(ctx *action.InstanceContext) error {
	param := new(WatchCMDBResourceParam)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	tenantCtx, err := tenant.SetID(ctx.Ctx, param.TenantID)
	if err != nil {
		return err
	}

	tenantUserCtx := contextx.NewTenantUserContext(tenantCtx, param.TenantID, param.Operator)

	err = act.watchHostResource(tenantUserCtx)
	if err != nil {
		return fmt.Errorf("watch host resource failed: %w", err)
	}

	err = act.watchHostRelationResource(tenantUserCtx)
	if err != nil {
		return fmt.Errorf("watch host relation resource failed: %w", err)
	}

	err = act.applyHostEvent(tenantCtx)
	if err != nil {
		return fmt.Errorf("apply host event failed: %w", err)
	}

	return nil
}

// watchHostResource watches the host resource events.
func (act *actionWatchCMDBResource) watchHostResource(ctx contextx.ITenantUserContext) error {
	cursor, err := act.getCursor(ctx, types.HostEventCursor)
	if err != nil {
		return fmt.Errorf("get host event cursor failed: %w", err)
	}

	events, err := act.cmdbHandler.WatchHostResourceEvent(ctx, cursor)
	if err != nil {
		return fmt.Errorf("watch host resource event failed: %w", err)
	}

	if len(events) == 0 {
		return nil
	}

	act.pendingProcessEvents = append(act.pendingProcessEvents, events...)

	if err := act.setCursor(ctx, types.HostEventCursor, events[len(events)-1].Cursor); err != nil {
		return fmt.Errorf("set host event cursor failed: %w", err)
	}

	return nil
}

// watchHostRelationResource watches the host relation resource events.
func (act *actionWatchCMDBResource) watchHostRelationResource(ctx contextx.ITenantUserContext) error {
	cursor, err := act.getCursor(ctx, types.HostRelationEventCursor)
	if err != nil {
		return fmt.Errorf("get host event cursor failed: %w", err)
	}

	events, err := act.cmdbHandler.WatchHostRelationResourceEvent(ctx, cursor)
	if err != nil {
		return fmt.Errorf("watch host relation resource event failed: %w", err)
	}

	if len(events) == 0 {
		return nil
	}

	pendingProcess := make(map[int64]*types.HostEvent, 0)
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
		act.pendingProcessEvents = append(act.pendingProcessEvents, ev)
	}

	if err := act.setCursor(ctx, types.HostRelationEventCursor, events[len(events)-1].Cursor); err != nil {
		return fmt.Errorf("set host relation event cursor failed: %w", err)
	}

	return nil
}

// applyHostEvent applies the host event to the watcher.
func (act *actionWatchCMDBResource) applyHostEvent(ctx context.Context) error {
	for _, event := range act.pendingProcessEvents {
		switch event.Resource {
		case types.ResourceTypeHost:
			if err := act.handleHostResource(ctx, event); err != nil {
				return err
			}
		case types.ResourceTypeHostRelation:
			if err := act.handleHostRelationResource(ctx, event); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown resource type: %s", event.Resource)
		}
	}

	return nil
}

// handleHostResource this func defines how to handle the host resource event.
func (act *actionWatchCMDBResource) handleHostResource(ctx context.Context, event *types.HostEvent) error {
	switch event.EventType {
	case types.EventTypeCreate:
		host, ok := act.waitingCompleteDataHostMap[event.Detail.HostID]
		if !ok {
			act.mu.Lock()
			defer act.mu.Unlock()
			// when the host synchronizes from the CMDB for the first time, the agentid needs to be updated to dynamic
			event.Detail.Dynamic.AgentID = event.Detail.Static.SyncedAgentID
			act.waitingCompleteDataHostMap[event.Detail.HostID] = event.Detail

			return nil
		}

		event.Detail.Static.BizID = host.Static.BizID
		err := act.storageTopo.UpsertManyHost(ctx, event.Detail)
		if err != nil {
			return fmt.Errorf("UpsertManyHost failed: %w", err)
		}

		act.mu.Lock()
		defer act.mu.Unlock()
		delete(act.waitingCompleteDataHostMap, event.Detail.HostID)

		return nil
	case types.EventTypeUpdate:
		dbHost, err := act.storageTopo.GetHostByID(ctx, event.Detail.HostID)
		if err != nil {
			return fmt.Errorf("GetHostByID failed: %w", err)
		}

		event.Detail.Static.BizID = dbHost.Static.BizID
		err = act.storageTopo.UpsertManyHostStatic(ctx, event.Detail)
		if err != nil {
			return fmt.Errorf("UpsertManyHostStatic failed: %w", err)
		}

		return nil
	case types.EventTypeDelete:
		err := act.storageTopo.DeleteManyHost(ctx, event.Detail.HostID)
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
func (act *actionWatchCMDBResource) handleHostRelationResource(ctx context.Context, event *types.HostEvent) error {
	switch event.EventType {
	case types.EventTypeCreate:
		host, ok := act.waitingCompleteDataHostMap[event.Detail.HostID]
		if !ok {
			act.mu.Lock()
			defer act.mu.Unlock()
			act.waitingCompleteDataHostMap[event.Detail.HostID] = event.Detail

			return nil
		}

		host.Static.BizID = event.Detail.Static.BizID
		err := act.storageTopo.UpsertManyHost(ctx, host)
		if err != nil {
			return fmt.Errorf("UpsertManyHost failed: %w", err)
		}

		act.mu.Lock()
		defer act.mu.Unlock()
		delete(act.waitingCompleteDataHostMap, event.Detail.HostID)

		return nil
	case types.EventTypeUpdate:
		dbHost, err := act.storageTopo.GetHostByID(ctx, event.Detail.HostID)
		if err != nil {
			return fmt.Errorf("GetHostByID failed: %w", err)
		}

		dbHost.Static.BizID = event.Detail.Static.BizID
		err = act.storageTopo.UpsertManyHostStatic(ctx, dbHost)
		if err != nil {
			return fmt.Errorf("UpsertManyHostStatic failed: %w", err)
		}

		return nil
	case types.EventTypeDelete:
		err := act.storageTopo.DeleteManyHost(ctx, event.Detail.HostID)
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

// getCursor retrieves the cursor for the given key from the cache.
func (act *actionWatchCMDBResource) getCursor(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", errors.New("get cursor from cache, key cannot be empty")
	}

	key = cacheKeyPrefix + key
	exist, err := act.cache.Exists(ctx, key)
	if err != nil {
		return "", fmt.Errorf("check cursor exist failed: %w", err)
	}

	if !exist {
		return "", nil
	}

	result, err := act.cache.Get(ctx, key)
	if err != nil {
		return "", fmt.Errorf("get cursor failed: %w", err)
	}

	return string(result), nil
}

// setCursor sets the cursor for the given key in the cache.
func (act *actionWatchCMDBResource) setCursor(ctx context.Context, key, value string) error {
	if key == "" {
		return errors.New("set cursor to cache, key cannot be empty")
	}

	key = cacheKeyPrefix + key
	if err := act.cache.SetWithExpiration(ctx, key, []byte(value), cacheExpirationTime); err != nil {
		return fmt.Errorf("set cursor failed: %w", err)
	}

	return nil
}
