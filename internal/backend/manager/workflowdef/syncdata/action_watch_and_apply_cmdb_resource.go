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
	"time"

	managerIface "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/iface"
	syncDataUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/cache"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/safequeue"
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
func NewActionWatchCMDBResource(capability *Capability) action.Definition {
	return &actionWatchAndApplyCMDBResource{
		cache:          capability.Cache,
		cmdbHandler:    capability.CMDBHandler,
		storageTopo:    capability.StorageTopo,
		storageProcess: capability.StoragePlugin,
		syncIface:      capability.SyncIface,
	}
}

// ActionParamWatchAndApplyCMDBResource defines the action's param.
type ActionParamWatchAndApplyCMDBResource struct {
	syncDataUtils.SyncDataActionStandardParam
}

// actionWatchAndApplyCMDBResource implements the action.Definition interface.
type actionWatchAndApplyCMDBResource struct {
	cache          cache.ICache
	cmdbHandler    cmdb.IHandler
	storageTopo    topoStg.IStorage
	storageProcess pluginStg.IDaoProcess
	syncIface      managerIface.ISyncManager

	pendingProcessEvents      *safequeue.SafeQueue[*types.HostEvent]
	waitingCreateHostMap      map[int64]*types.Host
	pendingCorrectAgentIDList []int64
}

// Name returns the name of the action.
func (act *actionWatchAndApplyCMDBResource) Name() string {
	return ActionNameWatchAndApplyCMDBResource
}

// Version returns the version of the action.
func (act *actionWatchAndApplyCMDBResource) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionWatchAndApplyCMDBResource) Description() string {
	return "watch and apply cmdb resource changes."
}

// Timeout returns the timeout of the action.
func (act *actionWatchAndApplyCMDBResource) Timeout() time.Duration {
	return 10 * time.Minute // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionWatchAndApplyCMDBResource) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount this action creates a large number of synchronization tasks,
// therefore does not allow the system to automatically retry.
func (act *actionWatchAndApplyCMDBResource) MaxRetryCount() uint {
	return 0
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionWatchAndApplyCMDBResource) DelayFn(_ int) func() {
	return func() {}
}

// Do this func define what the action will do.
func (act *actionWatchAndApplyCMDBResource) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamWatchAndApplyCMDBResource)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := syncDataUtils.NewSyncDataActionStandarder()
	if err = std.Initialize(ctx, param.SyncDataActionStandardParam); err != nil {
		return err
	}

	act.pendingProcessEvents = safequeue.NewSafeQueue[*types.HostEvent]()
	act.waitingCreateHostMap = make(map[int64]*types.Host)
	act.pendingCorrectAgentIDList = make([]int64, 0)

	gp := gopool.NewPool()
	gp.Go(func() error {
		if err := act.watchHostResource(std.Context()); err != nil {
			return fmt.Errorf("watch host resource failed: %w", err)
		}

		return nil
	})

	gp.Go(func() error {
		if err := act.watchHostRelationResource(std.Context()); err != nil {
			return fmt.Errorf("watch host relation resource failed: %w", err)
		}

		return nil
	})

	if err := gp.Wait(); err != nil {
		return fmt.Errorf("watch cmdb resource failed: %w", err)
	}

	if err = act.applyHostEvent(std); err != nil {
		return fmt.Errorf("apply host event failed: %w", err)
	}

	if err = act.tryTriggerCorrectAgentID(std); err != nil {
		return fmt.Errorf("trigger correct agent id failed: %w", err)
	}

	return nil
}

// watchHostResource watches the host resource events.
func (act *actionWatchAndApplyCMDBResource) watchHostResource(ctx contextx.IContext) error {
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

	for _, event := range events {
		if event.Detail == nil {
			continue
		}

		act.pendingProcessEvents.Enqueue(event)
	}

	if err := act.setCursor(ctx, types.HostEventCursor, events[len(events)-1].Cursor); err != nil {
		return fmt.Errorf("set host event cursor failed: %w", err)
	}

	return nil
}

// watchHostRelationResource watches the host relation resource events.
func (act *actionWatchAndApplyCMDBResource) watchHostRelationResource(ctx contextx.IContext) error {
	cursor, err := act.getCursor(ctx, types.HostRelationEventCursor)
	if err != nil {
		return fmt.Errorf("get host event cursor failed: %w", err)
	}

	events, err := act.cmdbHandler.WatchHostRelationResourceEvent(ctx, cursor)
	if err != nil {
		return fmt.Errorf("watch host relation resource event failed: %w", err)
	}

	eventSize := len(events)
	if eventSize == 0 {
		return nil
	}

	for i := 0; i < eventSize; {
		if events[i].Detail == nil {
			i++
			continue
		}

		// when the host relation event is delete followed by create, it means the host has been moved to another business
		// we can treat it as update and only process the create event to avoid redundant processing of delete and create.
		if events[i].EventType == types.EventTypeDelete &&
			i+1 < eventSize &&
			events[i+1].Detail != nil &&
			events[i+1].Detail.HostID == events[i].Detail.HostID &&
			events[i+1].EventType == types.EventTypeCreate {

			merged := *events[i+1]
			merged.EventType = types.EventTypeUpdate
			act.pendingProcessEvents.Enqueue(&merged)
			i += 2

			continue
		}

		act.pendingProcessEvents.Enqueue(events[i])
		i++
	}

	if err := act.setCursor(ctx, types.HostRelationEventCursor, events[len(events)-1].Cursor); err != nil {
		return fmt.Errorf("set host relation event cursor failed: %w", err)
	}

	return nil
}

// applyHostEvent applies the host event to the watcher.
func (act *actionWatchAndApplyCMDBResource) applyHostEvent(std *syncDataUtils.SyncDataActionStandarder) error {
	for !act.pendingProcessEvents.IsEmpty() {
		event, ok := act.pendingProcessEvents.Dequeue()
		if !ok {
			break
		}

		switch event.Resource {
		case types.ResourceTypeHost:
			act.handleHostResource(std, event)
		case types.ResourceTypeHostRelation:
			act.handleHostRelationResource(std, event)
		default:
			return fmt.Errorf("unknown resource type: %s", event.Resource)
		}
	}

	return nil
}

// handleHostResource this func defines how to handle the host resource event.
func (act *actionWatchAndApplyCMDBResource) handleHostResource(std *syncDataUtils.SyncDataActionStandarder, event *types.HostEvent) {
	switch event.EventType {
	case types.EventTypeCreate:
		act.handleHostCreateEvent(std, event)
	case types.EventTypeUpdate:
		act.handleHostUpdateEvent(std, event)
	case types.EventTypeDelete:
		act.handleHostDeleteEvent(std, event)
	default:
		return
	}
}

func (act *actionWatchAndApplyCMDBResource) handleHostCreateEvent(std *syncDataUtils.SyncDataActionStandarder, event *types.HostEvent) {
	host, ok := act.waitingCreateHostMap[event.Detail.HostID]
	if !ok {
		// when the host synchronizes from the CMDB for the first time, the agentid needs to be updated to dynamic
		syncDataUtils.FillHostDynamicAgentID(event.Detail)
		// when the host synchronizes from the CMDB for the first time, the ops info needs to be updated to dynamic
		syncDataUtils.FillHostDynamicOpsInfo(event.Detail)
		fillDefaultAdvertiseIP(event.Detail, nil)
		act.waitingCreateHostMap[event.Detail.HostID] = event.Detail

		return
	}

	event.Detail.Static.BizID = host.Static.BizID
	event.Detail.Static.Topo = host.Static.Topo
	fillDefaultAdvertiseIP(event.Detail, nil)
	if err := act.storageTopo.CreateManyHost(std.Context(), event.Detail); err != nil {
		std.InstanceData().Log().
			Zh("创建主机失败, 主机id: %d, 错误: %v", event.Detail.HostID, err).
			En("failed to create host, host id: %d, error: %v", event.Detail.HostID, err).
			Error()

		return
	}

	delete(act.waitingCreateHostMap, event.Detail.HostID)
}

func (act *actionWatchAndApplyCMDBResource) handleHostUpdateEvent(std *syncDataUtils.SyncDataActionStandarder, event *types.HostEvent) {
	exists, err := act.storageTopo.ExistHost(std.Context(), &types.HostCondition{
		StaticExactInclude: &types.HostStaticExactFields{
			HostID: []int64{event.Detail.HostID},
		},
	})
	if err != nil {
		std.InstanceData().Log().
			Zh("检查主机是否存在失败, 主机id: %d, 错误: %v", event.Detail.HostID, err).
			En("failed to check host existence, host id: %d, error: %v", event.Detail.HostID, err).
			Error()

		return
	}

	if !exists {
		std.InstanceData().Log().
			Zh("主机不存在于本地数据库, 尝试从CMDB获取主机信息, 主机id: %d", event.Detail.HostID).
			En("host not found in local db, try to get host info from cmdb, host id: %d", event.Detail.HostID).
			Info()

		if err := act.tryCreateHostFromCMDB(std, event.Detail.HostID, event.Detail.Static.BizID); err != nil {
			std.InstanceData().Log().
				Zh("从CMDB获取主机信息失败, 主机id: %d, 错误: %v", event.Detail.HostID, err).
				En("failed to get host info from cmdb, host id: %d, error: %v", event.Detail.HostID, err).
				Error()

			return
		}

		return
	}

	dbHost, err := act.storageTopo.GetHostByID(std.Context(), event.Detail.HostID)
	if err != nil {
		std.InstanceData().Log().
			Zh("通过主机id获取主机信息失败, 主机id: %d, 错误: %v", event.Detail.HostID, err).
			En("failed to get host info by host id, host id: %d, error: %v", event.Detail.HostID, err).
			Error()

		return
	}

	fields := types.UpdateAllHostStaticFields()
	fields.Topo = false  // topo field should not be updated by host update event, it should be updated by host relation event
	fields.BizID = false // biz id field should not be updated by host update event, it should be updated by host relation event
	if err := act.storageTopo.UpdateHostStaticFields(std.Context(), fields, event.Detail); err != nil {
		std.InstanceData().Log().
			Zh("更新主机静态信息失败, 主机id: %d, 错误: %v", event.Detail.HostID, err).
			En("failed to update host static info, host id: %d, error: %v", event.Detail.HostID, err).
			Error()

		return
	}

	if dbHost.Dynamic.AgentID == "" && event.Detail.Static.SyncedAgentID != "" {
		event.Detail.Dynamic.AgentID = event.Detail.Static.SyncedAgentID
		if err := act.storageTopo.UpdateHostDynamicFields(std.Context(), types.HostDynamicFields{AgentID: true}, event.Detail); err != nil {
			std.InstanceData().Log().
				Zh("更新主机动态信息失败, 主机id: %d, 错误: %v", event.Detail.HostID, err).
				En("failed to update host dynamic info, host id: %d, error: %v", event.Detail.HostID, err).
				Error()

			return
		}
	}

	if fillDefaultAdvertiseIP(event.Detail, dbHost) {
		if err := act.storageTopo.UpdateHostDynamicFields(std.Context(), types.HostDynamicFields{
			AdvertiseIP:   true,
			AdvertiseIPV6: true,
		}, event.Detail); err != nil {
			std.InstanceData().Log().
				Zh("更新主机动态服务IP失败, 主机id: %d, 错误: %v", event.Detail.HostID, err).
				En("failed to update host dynamic advertise ip, host id: %d, error: %v", event.Detail.HostID, err).
				Error()

			return
		}
	}

	if act.checkHostNeedCorrectAgentID(dbHost, event.Detail) {
		act.pendingCorrectAgentIDList = append(act.pendingCorrectAgentIDList, event.Detail.HostID)
	}
}

func (act *actionWatchAndApplyCMDBResource) handleHostDeleteEvent(std *syncDataUtils.SyncDataActionStandarder, event *types.HostEvent) {
	if err := act.storageTopo.DeleteManyHost(std.Context(), event.Detail.HostID); err != nil {
		std.InstanceData().Log().
			Zh("删除主机失败, 主机id: %d, 错误: %v", event.Detail.HostID, err).
			En("failed to delete host, host id: %d, error: %v", event.Detail.HostID, err).
			Error()

		return
	}
}

// handleHostRelationResource this func defines how to handle the host relation resource event.
// if the host relation is deleted, it means the host is need deleted
// host event will handle the delete logic, so we can ignore the delete event of host relation.
func (act *actionWatchAndApplyCMDBResource) handleHostRelationResource(std *syncDataUtils.SyncDataActionStandarder, event *types.HostEvent) {
	switch event.EventType {
	case types.EventTypeCreate:
		act.handleHostRelationCreateEvent(std, event)
	case types.EventTypeUpdate:
		act.handleHostRelationUpdateEvent(std, event)
	case types.EventTypeDelete:
		act.handleHostRelationDeleteEvent(std, event)
	default:
		return
	}
}

func (act *actionWatchAndApplyCMDBResource) handleHostRelationCreateEvent(std *syncDataUtils.SyncDataActionStandarder, event *types.HostEvent) {
	host, ok := act.waitingCreateHostMap[event.Detail.HostID]
	existHost, err := act.storageTopo.ExistHost(std.Context(), &types.HostCondition{
		StaticExactInclude: &types.HostStaticExactFields{HostID: []int64{event.Detail.HostID}},
	})
	if err != nil {
		std.InstanceData().Log().
			Zh("检查主机是否存在失败, 主机id: %d, 错误: %v", event.Detail.HostID, err).
			En("failed to check host existence, host id: %d, error: %v", event.Detail.HostID, err).
			Error()

		return
	}

	if !ok && !existHost {
		act.waitingCreateHostMap[event.Detail.HostID] = event.Detail

		return
	}

	if ok && !existHost {
		host.Static.BizID = event.Detail.Static.BizID
		host.Static.Topo = event.Detail.Static.Topo
		fillDefaultAdvertiseIP(host, nil)
		if err := act.storageTopo.CreateManyHost(std.Context(), host); err != nil {
			std.InstanceData().Log().
				Zh("创建主机失败, 主机id: %d, 错误: %v", host.HostID, err).
				En("failed to create host, host id: %d, error: %v", host.HostID, err).
				Error()

			return
		}

		delete(act.waitingCreateHostMap, event.Detail.HostID)

		return
	}

	hostRelation := &types.HostTopoRelation{
		HostID: event.Detail.HostID,
		BizID:  event.Detail.Static.BizID,
		Topo:   event.Detail.Static.Topo,
	}
	if err := act.storageTopo.UpsertHostTopo(std.Context(), hostRelation); err != nil {
		std.InstanceData().Log().
			Zh("更新主机拓扑信息失败, 主机id: %d, 错误: %v", event.Detail.HostID, err).
			En("failed to upsert host topo, host id: %d, error: %v", event.Detail.HostID, err).
			Error()

		return
	}

	if err := act.storageTopo.UpdateHostStaticFields(std.Context(), types.HostStaticFields{BizID: true}, event.Detail); err != nil {
		std.InstanceData().Log().
			Zh("更新主机业务信息失败, 主机id: %d, 错误: %v", event.Detail.HostID, err).
			En("failed to update biz id of host, host id: %d, error: %v", event.Detail.HostID, err).
			Error()

		return
	}
}

func (act *actionWatchAndApplyCMDBResource) handleHostRelationUpdateEvent(std *syncDataUtils.SyncDataActionStandarder, event *types.HostEvent) {
	exists, err := act.storageTopo.ExistHost(std.Context(), &types.HostCondition{
		StaticExactInclude: &types.HostStaticExactFields{HostID: []int64{event.Detail.HostID}},
	})
	if err != nil {
		std.InstanceData().Log().
			Zh("检查主机是否存在失败, 主机id: %d, 错误: %v", event.Detail.HostID, err).
			En("failed to check host existence, host id: %d, error: %v", event.Detail.HostID, err).
			Error()

		return
	}

	if !exists {
		std.InstanceData().Log().
			Zh("主机不存在于本地数据库, 尝试从CMDB获取主机信息, 主机id: %d", event.Detail.HostID).
			En("host not found in local db, try to get host info from cmdb, host id: %d", event.Detail.HostID).
			Info()
		if err := act.tryCreateHostFromCMDB(std, event.Detail.HostID, event.Detail.Static.BizID); err != nil {
			std.InstanceData().Log().
				Zh("从CMDB获取主机信息失败, 主机id: %d, 错误: %v", event.Detail.HostID, err).
				En("failed to get host info from cmdb, host id: %d, error: %v", event.Detail.HostID, err).
				Error()

			return
		}

		return
	}

	hostRelation := &types.HostTopoRelation{
		HostID: event.Detail.HostID,
		BizID:  event.Detail.Static.BizID,
		Topo:   event.Detail.Static.Topo,
	}
	if err = act.storageTopo.UpsertHostTopo(std.Context(), hostRelation); err != nil {
		std.InstanceData().Log().
			Zh("更新主机拓扑信息失败, 主机id: %d, 错误: %v", event.Detail.HostID, err).
			En("failed to upsert host topo, host id: %d, error: %v", event.Detail.HostID, err).
			Error()

		return
	}

	if err := act.storageTopo.UpdateHostStaticFields(std.Context(), types.HostStaticFields{BizID: true}, event.Detail); err != nil {
		std.InstanceData().Log().
			Zh("更新主机业务信息失败, 主机id: %d, 错误: %v", event.Detail.HostID, err).
			En("failed to update biz id of host, host id: %d, error: %v", event.Detail.HostID, err).
			Error()

		return
	}

	if err := act.storageProcess.UpdateProcessManyHostBizID(std.Context(), event.Detail.Static.BizID, event.Detail.HostID); err != nil {
		std.InstanceData().Log().
			Zh("更新主机相关进程的业务id失败, 主机id: %d, 业务id: %d, 错误: %v", event.Detail.HostID, event.Detail.Static.BizID, err).
			En("failed to update host related process biz id, host id: %d, biz id: %d, error: %v",
				event.Detail.HostID, event.Detail.Static.BizID, err).
			Error()

		return
	}
}

func (act *actionWatchAndApplyCMDBResource) handleHostRelationDeleteEvent(std *syncDataUtils.SyncDataActionStandarder, event *types.HostEvent) {
	hostRelation := &types.HostTopoRelation{
		HostID: event.Detail.HostID,
		BizID:  event.Detail.Static.BizID,
		Topo:   event.Detail.Static.Topo,
	}
	if err := act.storageTopo.PopHostTopo(std.Context(), hostRelation); err != nil {
		std.InstanceData().Log().
			Zh("从主机中移除主机拓扑失败, 主机id: %d, 错误: %v", event.Detail.HostID, err).
			En("failed to remove host topo from host, host id: %d, error: %v", event.Detail.HostID, err).
			Error()

		return
	}
}

// tryCreateHostFromCMDB depends on the host id to get the host information from CMDB.
func (act *actionWatchAndApplyCMDBResource) tryCreateHostFromCMDB(std *syncDataUtils.SyncDataActionStandarder, hostID int64, bizID int64) error {
	cond := &types.HostStaticExactCondition{
		StaticExactInclude: &types.HostStaticExactFields{HostID: []int64{hostID}},
	}
	if bizID != cmdb.CCNoBusinessID {
		cond.StaticExactInclude.BizID = []int64{bizID}
	}

	hosts, err := act.cmdbHandler.FindHostWithCondition(std.Context(), types.SingleItemPage(), cond)
	if err != nil {
		std.InstanceData().Log().
			Zh("通过主机id从CMDB获取主机信息失败, 主机id: %d, 错误: %v", hostID, err).
			En("failed to get host info from cmdb by host id, host id: %d, error: %v", hostID, err).
			Error()

		return err
	}

	if len(hosts) == 0 || hosts[0] == nil {
		std.InstanceData().Log().
			Zh("通过主机id从CMDB未找到主机信息, 跳过主机关联关系信息更新, 主机id: %d", hostID).
			En("host info not found in cmdb by host id, skip host relation info update, host id: %d", hostID).
			Info()

		return nil
	}

	host := hosts[0]
	// when the host synchronizes from the CMDB for the first time, the agentid needs to be updated to dynamic
	syncDataUtils.FillHostDynamicAgentID(host)
	// when the host synchronizes from the CMDB for the first time, the ops info needs to be updated to dynamic
	syncDataUtils.FillHostDynamicOpsInfo(host)
	fillDefaultAdvertiseIP(host, nil)
	if err := act.storageTopo.CreateManyHost(std.Context(), host); err != nil {
		std.InstanceData().Log().
			Zh("创建主机失败, 主机id: %d, 错误: %v", hostID, err).
			En("failed to create host info, host id: %d, error: %v", hostID, err).
			Error()

		return err
	}

	if err := act.storageProcess.UpdateProcessManyHostBizID(std.Context(), host.Static.BizID, host.HostID); err != nil {
		std.InstanceData().Log().
			Zh("更新主机相关进程的业务id失败, 主机id: %d, 业务id: %d, 错误: %v", host.HostID, host.Static.BizID, err).
			En("failed to update host related process biz id, host id: %d, biz id: %d, error: %v", host.HostID, host.Static.BizID, err).
			Error()

		return err
	}

	return nil
}

func (act *actionWatchAndApplyCMDBResource) checkHostNeedCorrectAgentID(dbHost, cmdbHost *types.Host) bool {
	if dbHost.Dynamic == nil || dbHost.Dynamic.AgentID == "" {
		return false
	}
	if cmdbHost.Static == nil || cmdbHost.Static.SyncedAgentID == "" {
		return false
	}
	if cmdbHost.Static.SyncedAgentID == dbHost.Dynamic.AgentID {
		return false
	}
	if dbHost.Dynamic.NodeStatus == types.NodeStatusRunning {
		return false
	}

	return true
}

func (act *actionWatchAndApplyCMDBResource) tryTriggerCorrectAgentID(std *syncDataUtils.SyncDataActionStandarder) error {
	if len(act.pendingCorrectAgentIDList) == 0 {
		return nil
	}

	std.InstanceData().Log().
		Zh("发现 %d 台主机需要修正 Agent ID，触发修正任务", len(act.pendingCorrectAgentIDList)).
		En("found %d hosts need to correct agent id, triggering correction task", len(act.pendingCorrectAgentIDList)).
		Info()

	triggerID, err := act.syncIface.LaunchSyncCorrectAgentID(std.Context(), act.pendingCorrectAgentIDList...)
	if err != nil {
		logger.G.Sys().Ctx(std.Context()).WithErr(err).Error("failed to launch correct agent id task")
		return err
	}

	logger.G.Sys().Ctx(std.Context()).
		With("trigger-id", triggerID, "host-count", len(act.pendingCorrectAgentIDList)).
		Info("launched correct agent id task")

	return nil
}

// getCursor retrieves the cursor for the given key from the cache.
func (act *actionWatchAndApplyCMDBResource) getCursor(ctx context.Context, key string) (string, error) {
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
func (act *actionWatchAndApplyCMDBResource) setCursor(ctx context.Context, key, value string) error {
	if key == "" {
		return errors.New("set cursor to cache, key cannot be empty")
	}

	key = cacheKeyPrefix + key
	if err := act.cache.SetWithExpiration(ctx, key, []byte(value), cacheExpirationTime); err != nil {
		return fmt.Errorf("set cursor failed: %w", err)
	}

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionWatchAndApplyCMDBResource) DisplayNameZh() string {
	return "监听并应用 CMDB 资源变更"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionWatchAndApplyCMDBResource) DisplayNameEn() string {
	return "Watch and Apply CMDB Resource"
}
