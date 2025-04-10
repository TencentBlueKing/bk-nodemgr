/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cmdb

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/scheduler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IWatcher defines the interface of watcher.
// Notice: need watch first, then start.
type IWatcher interface {
	// WatchHost watch resource host id.
	WatchHost() (<-chan *types.ChangeEvent[*types.Host], error)

	// WatchHostRelation watch resource host relation id.
	WatchHostRelation() (<-chan *types.ChangeEvent[*types.HostRel], error)

	// Start start the watcher.
	Start(ctx context.Context) error
}

// Watcher implements the IWatcher interface.
type Watcher struct {
	scheduler scheduler.Scheduler
	handler   *Handler
	logger    logger.Logger

	done chan struct{}

	resourceHost struct {
		sync.Once
		cursor  string
		channel chan *types.ChangeEvent[*types.Host]
	}

	resourceHostRel struct {
		sync.Once
		cursor  string
		channel chan *types.ChangeEvent[*types.HostRel]
	}
}

// NewWatcher create a new watcher.
func NewWatcher(handler *Handler) *Watcher {
	w := &Watcher{
		logger:  handler.logger,
		handler: handler,
	}

	w.resourceHost.channel = make(chan *types.ChangeEvent[*types.Host], resourceChanBuffer)
	w.resourceHostRel.channel = make(chan *types.ChangeEvent[*types.HostRel], resourceChanBuffer)

	return w
}

// resourceChanBuffer is the buffer size of the channel.
const resourceChanBuffer = 100

// WatchHost watch resource host id.
func (w *Watcher) WatchHost() (<-chan *types.ChangeEvent[*types.Host], error) {
	valid := false
	w.resourceHost.Once.Do(func() {
		valid = true
	})

	if !valid {
		return nil, errors.New("host watcher already be watched")
	}

	return w.resourceHost.channel, nil
}

// WatchHostRelation watch resource host relation id.
func (w *Watcher) WatchHostRelation() (<-chan *types.ChangeEvent[*types.HostRel], error) {
	valid := false
	w.resourceHostRel.Once.Do(func() {
		valid = true
	})

	if !valid {
		return nil, errors.New("host relation watcher already be watched")
	}

	return w.resourceHostRel.channel, nil
}

// Start start the watcher.
// nolint: gocognit
func (w *Watcher) Start(ctx context.Context) error {
	w.done = make(chan struct{})

	registerHost, registerHostRel := true, true
	w.resourceHost.Once.Do(func() {
		registerHost = false
	})
	w.resourceHostRel.Once.Do(func() {
		registerHostRel = false
	})

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return fmt.Errorf("get tenant id failed, err: %v", err)
	}

	if registerHost {
		w.scheduler.RegisterTask(&scheduler.Task{
			ID:       WatchResourceHost,
			Interval: time.Second,
			Timeout:  time.Minute,
			Fn: func(ctx context.Context) error {
				ctx, err := tenant.SetID(ctx, tenantID)
				if err != nil {
					return fmt.Errorf("set tenant id failed, err: %v", err)
				}

				events, cursor, err := w.getHostResourceByWatch(ctx, w.resourceHost.cursor)
				if err != nil {
					return fmt.Errorf("get host resource by watch failed, err: %v", err)
				}

				w.resourceHost.cursor = cursor
				if len(events) == 0 {
					return nil
				}

				for _, event := range events {
					w.resourceHost.channel <- &types.ChangeEvent[*types.Host]{
						ChangeType: types.ChangeType(event.BKEventType),
						Detail:     w.handler.convHostInfoToTypes(tenantID, event.BKDetail, CCNoBusinessID),
					}
				}

				return nil
			},
		})
	}

	if registerHostRel {
		w.scheduler.RegisterTask(&scheduler.Task{
			ID:       WatchResourceHostRelation,
			Interval: time.Second,
			Timeout:  time.Minute,
			Fn: func(ctx context.Context) error {
				ctx, err := tenant.SetID(ctx, tenantID)
				if err != nil {
					return fmt.Errorf("set tenant id failed, err: %v", err)
				}

				events, cursor, err := w.getHostRelationResourceByWatch(ctx, w.resourceHostRel.cursor)
				if err != nil {
					return fmt.Errorf("get host relation resource by watch failed, err: %v", err)
				}

				w.resourceHostRel.cursor = cursor
				if len(events) == 0 {
					return nil
				}

				for _, event := range events {
					w.resourceHostRel.channel <- &types.ChangeEvent[*types.HostRel]{
						ChangeType: types.ChangeType(event.BKEventType),
						Detail:     convHostTopoRelationToTypes(event.BKDetail),
					}
				}

				return nil
			},
		})
	}

	w.scheduler.Start()

	return nil
}

// getHostResourceByWatch get host resource by watch.
func (w *Watcher) getHostResourceByWatch(ctx context.Context, cursor string) ([]*HostEventInfo, string, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, "", err
	}

	req := &ResourceWatchReq{
		TenantID:   tenantID,
		BKCursor:   cursor,
		BKResource: WatchResourceHost,
		BKFields:   ccHostFields(),
	}

	resp, err := w.handler.cli.resourceWatch(ctx, req)
	if err != nil {
		return nil, "", err
	}

	newCursor := ""
	result := make([]*HostEventInfo, 0)
	for _, hostEvent := range resp.BKEvents {
		hostData := new(HostEventInfo)
		if err := conv.MapToStruct(*hostEvent, hostData); err != nil {
			return nil, "", err
		}

		newCursor = hostData.BKCursor

		if !resp.BKWatched {
			break
		}

		if hostData.BKDetail == nil {
			continue
		}

		result = append(result, hostData)
	}

	return result, newCursor, nil
}

// getHostRelationResourceByWatch get host relation resource by watch.
func (w *Watcher) getHostRelationResourceByWatch(ctx context.Context, cursor string) (
	[]*HostRelationEventInfo, string, error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, "", err
	}

	req := &ResourceWatchReq{
		TenantID:   tenantID,
		BKCursor:   cursor,
		BKResource: WatchResourceHostRelation,
	}

	resp, err := w.handler.cli.resourceWatch(ctx, req)
	if err != nil {
		return nil, "", err
	}

	newCursor := ""
	result := make([]*HostRelationEventInfo, 0)
	for _, relationEvent := range resp.BKEvents {
		relationData := new(HostRelationEventInfo)
		if err := conv.MapToStruct(*relationEvent, relationData); err != nil {
			return nil, "", err
		}

		newCursor = relationData.BKCursor

		if !resp.BKWatched {
			break
		}

		if relationData.BKDetail == nil {
			continue
		}

		result = append(result, relationData)
	}

	return result, newCursor, nil
}

// getProcessResourceByWatch get process resource by watch.
// nolint: unused
func (w *Watcher) getProcessResourceByWatch(ctx context.Context, cursor string) ([]*ProcessEventInfo, string,
	error) {

	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, "", err
	}

	req := &ResourceWatchReq{
		TenantID:   tenantID,
		BKCursor:   cursor,
		BKResource: WatchResourceProcess,
	}

	resp, err := w.handler.cli.resourceWatch(ctx, req)
	if err != nil {
		return nil, "", err
	}

	newCursor := ""
	result := make([]*ProcessEventInfo, len(resp.BKEvents))
	for _, processEvent := range resp.BKEvents {
		processData := new(ProcessEventInfo)
		if err := conv.MapToStruct(*processEvent, processData); err != nil {
			return nil, "", err
		}

		newCursor = processData.BKCursor

		if !resp.BKWatched {
			break
		}

		if processData.BKDetail == nil {
			continue
		}

		result = append(result, processData)
	}

	return result, newCursor, nil
}

const (
	// WatchResourceHost describe the resource watch's resource host.
	WatchResourceHost = "host"

	// WatchResourceHostRelation describe the resource watch's resource host relation.
	WatchResourceHostRelation = "host_relation"

	// WatchResourceProcess describe the resource watch's resource process.
	WatchResourceProcess = "process"
)

// ResourceWatchEventType represents the event type of watch event.
type ResourceWatchEventType string

const (
	// ResourceWatchEventTypeCreate represents the create event type.
	ResourceWatchEventTypeCreate ResourceWatchEventType = "create"

	// ResourceWatchEventTypeUpdate represents the update event type.
	ResourceWatchEventTypeUpdate ResourceWatchEventType = "update"

	// ResourceWatchEventTypeDelete represents the delete event type.
	ResourceWatchEventTypeDelete ResourceWatchEventType = "delete"
)

// EventInfo describe the event info define by cmdb.
type EventInfo[T any] struct {
	BKCursor    string `json:"bk_cursor,omitempty"`
	BKResource  string `json:"bk_resource"`
	BKEventType string `json:"bk_event_type,omitempty"`
	BKDetail    T      `json:"bk_detail"`
}

// HostEventInfo describe the host event info define by cmdb.
type HostEventInfo = EventInfo[*HostInfo]

// HostRelationEventInfo describe the host relation event info define by cmdb.
type HostRelationEventInfo = EventInfo[*HostTopoRelation]

// ProcessEventInfo describe the host relation event info define by cmdb.
type ProcessEventInfo = EventInfo[*ProcessProperty]
