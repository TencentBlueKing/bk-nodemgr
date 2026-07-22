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
	"fmt"
	"slices"
	"time"

	managerIface "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/iface"
	syncDataUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/batchexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameSyncHost defines the action name.
	ActionNameSyncHost = "sync_host"

	syncHostDBBatchSize = 50 // nolint: mnd
)

// NewActionSyncHost creates a new actionSyncHost.
func NewActionSyncHost(capability *Capability) action.Definition {
	return &actionSyncHost{
		cmdbHandler:    capability.CMDBHandler,
		storageHost:    capability.StorageTopo,
		storageProcess: capability.StoragePlugin,
		syncIface:      capability.SyncIface,
	}
}

// ActionParamSyncHost ...
type ActionParamSyncHost struct {
	syncDataUtils.SyncDataActionStandardParam

	BizID int64 `json:"biz_id"`
}

type actionSyncHost struct {
	cmdbHandler    cmdb.IHandler
	storageHost    topoStg.IStorageHost
	storageProcess pluginStg.IDaoProcess
	syncIface      managerIface.ISyncManager
}

// Name returns the name of the action.
func (act *actionSyncHost) Name() string {
	return ActionNameSyncHost
}

// Version returns the version of the action.
func (act *actionSyncHost) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionSyncHost) Description() string {
	return "get the host information of the designated business from cmdb, and update to the database"
}

// Timeout returns the timeout of this action.
func (act *actionSyncHost) Timeout() time.Duration {
	return 10 * time.Minute // nolint: mnd
}

// MaxRetryCount returns the max retry count of this action.
func (act *actionSyncHost) MaxRetryCount() uint {
	return 5 // nolint: mnd
}

// DelayFn returns the delay of this action.
// Uses exponential delay with jitter for staggered retry timing across instances.
func (act *actionSyncHost) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Tags returns the tags of this action.
func (act *actionSyncHost) Tags() []action.Tag {
	return []action.Tag{}
}

// Do the action.
func (act *actionSyncHost) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamSyncHost)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := syncDataUtils.NewSyncDataActionStandarder()
	if err = std.Initialize(ctx, param.SyncDataActionStandardParam); err != nil {
		return err
	}

	var cmdbData, dbData []*types.Host
	gp := gopool.NewPool()
	gp.Go(func() error {
		cond := &types.HostStaticExactCondition{
			StaticExactInclude: &types.HostStaticExactFields{
				BizID: []int64{param.BizID},
			},
		}
		cmdbData, err = act.cmdbHandler.FindHostWithCondition(std.Context(), types.UnlimitedPage(), cond)
		if err != nil {
			return fmt.Errorf("list host from cmdb failed: %w", err)
		}

		return nil
	})

	gp.Go(func() error {
		selection := &types.HostFieldSelection{
			// Sync only needs host_id for comparison and dynamic fields for repair/sync.
			HostID:        true,
			NetworkAreaID: false,
			InnerIPList:   false,
			InnerIPV6List: false,
			NodeRole:      true,
			LoginUser:     true,
			AgentID:       true,
			AdvertiseIP:   true,
			AdvertiseIPV6: true,
		}

		condition := &types.HostCondition{
			StaticExactInclude: &types.HostStaticExactFields{
				BizID: []int64{param.BizID},
			},
		}

		dbData, _, err = act.storageHost.ListHostWithFields(std.Context(), types.UnlimitedPage(), selection, condition)
		if err != nil {
			return fmt.Errorf("list host from db failed: %w", err)
		}

		return nil
	})

	if err = gp.Wait(); err != nil {
		return err
	}

	ctx.Data.Log().
		Zh("从 CMDB 获取到 %d 台主机，数据库中有 %d 台主机", len(cmdbData), len(dbData)).
		En("find %d hosts from cmdb, %d hosts in db", len(cmdbData), len(dbData)).
		Info()
	updateHosts, insertHosts, deleteHostIDs, err := act.compareData(cmdbData, dbData)
	if err != nil {
		return err
	}

	updateAgentIDHosts, err := getAndFillEmptyDynamicAgentIDByStaticSyncedAgentID(updateHosts, dbData)
	if err != nil {
		return err
	}
	updateAdvertiseIPHosts := fillDefaultAdvertiseIPs(updateHosts, dbData)
	insertAdvertiseIPHosts := fillDefaultAdvertiseIPs(insertHosts, nil)

	logHostCompareResult(ctx,
		len(updateHosts),
		len(updateAgentIDHosts),
		len(updateAdvertiseIPHosts)+len(insertAdvertiseIPHosts),
		len(insertHosts),
		len(deleteHostIDs),
	)

	if err = act.persistHostSyncChanges(ctx, std.Context(), hostSyncChanges{
		updateHosts:            updateHosts,
		insertHosts:            insertHosts,
		deleteHostIDs:          deleteHostIDs,
		updateAgentIDHosts:     updateAgentIDHosts,
		updateAdvertiseIPHosts: updateAdvertiseIPHosts,
		dbData:                 dbData,
	}); err != nil {
		return err
	}

	if err := act.tryUpdateHostProcessBizID(std, param.BizID, cmdbData...); err != nil {
		return err
	}

	if err := act.tryTriggerCorrectAgentID(ctx, std, updateHosts, dbData); err != nil {
		return err
	}

	return nil
}

type hostSyncChanges struct {
	updateHosts            []*types.Host
	insertHosts            []*types.Host
	deleteHostIDs          []int64
	updateAgentIDHosts     []*types.Host
	updateAdvertiseIPHosts []*types.Host
	dbData                 []*types.Host
}

func (act *actionSyncHost) persistHostSyncChanges(
	ctx *action.InstanceContext,
	nCtx contextx.IContext,
	changes hostSyncChanges,
) error {

	if err := batchHandleHosts(nCtx, changes.updateHosts, func(hosts ...*types.Host) error {
		return act.updateHostsStaticWithTopo(nCtx, hosts...)
	}); err != nil {
		return err
	}

	if err := batchHandleHosts(nCtx, changes.insertHosts, func(hosts ...*types.Host) error {
		return act.storageHost.UpsertManyHost(nCtx, hosts...)
	}); err != nil {
		return err
	}

	if err := act.updateHostDynamicAdvertiseIP(nCtx, changes.updateAdvertiseIPHosts); err != nil {
		return err
	}

	if err := batchHandleHosts(nCtx, changes.updateAgentIDHosts, func(hosts ...*types.Host) error {
		return act.storageHost.UpdateHostDynamicFields(nCtx, types.HostDynamicFields{AgentID: true}, hosts...)
	}); err != nil {
		return err
	}

	if err := batchHandleHostIDs(nCtx, changes.deleteHostIDs, func(hostIDs ...int64) error {
		return act.storageHost.DeleteManyHost(nCtx, hostIDs...)
	}); err != nil {
		return err
	}

	loginUserRepairHosts := act.fillDefaultLoginUsers(
		ctx,
		slices.Concat(changes.updateHosts, changes.insertHosts),
		changes.dbData,
	)

	return batchHandleHosts(nCtx, loginUserRepairHosts, func(hosts ...*types.Host) error {
		return act.storageHost.UpdateHostDynamicFields(nCtx, types.HostDynamicFields{LoginUser: true}, hosts...)
	})
}

func logHostCompareResult(
	ctx *action.InstanceContext,
	updateCount int,
	agentIDRepairCount int,
	advertiseIPRepairCount int,
	insertCount int,
	deleteCount int,
) {

	ctx.Data.Log().
		Zh("对比完成，需更新 %d 台、需更新AgentID %d 台、需修补服务IP %d 台、新增 %d 台、删除 %d 台主机",
			updateCount, agentIDRepairCount, advertiseIPRepairCount, insertCount, deleteCount).
		En("compared hosts, %d hosts need to update, %d hosts need to update AgentID, "+
			"%d hosts need to repair advertise IP, %d hosts need to insert, %d hosts need to delete",
			updateCount, agentIDRepairCount, advertiseIPRepairCount, insertCount, deleteCount).
		Info()
}

func (act *actionSyncHost) updateHostDynamicAdvertiseIP(nCtx contextx.IContext, hosts []*types.Host) error {
	return batchHandleHosts(nCtx, hosts, func(batchHosts ...*types.Host) error {
		return act.storageHost.UpdateHostDynamicFields(nCtx, types.HostDynamicFields{
			AdvertiseIP:   true,
			AdvertiseIPV6: true,
		}, batchHosts...)
	})
}

func batchHandleHosts(nCtx contextx.IContext, hosts []*types.Host, fn func(hosts ...*types.Host) error) error {
	return batchexecutor.Execute(nCtx, hosts, func(_ contextx.IContext, batchHosts []*types.Host) error {
		return fn(batchHosts...)
	}, batchexecutor.WithBatchSize(syncHostDBBatchSize), batchexecutor.WithTimeout(10*time.Minute)) // nolint: mnd
}

func batchHandleHostIDs(nCtx contextx.IContext, hostIDs []int64, fn func(hostIDs ...int64) error) error {
	return batchexecutor.Execute(nCtx, hostIDs, func(_ contextx.IContext, batchHostIDs []int64) error {
		return fn(batchHostIDs...)
	}, batchexecutor.WithBatchSize(syncHostDBBatchSize), batchexecutor.WithTimeout(10*time.Minute)) // nolint: mnd
}

func (act *actionSyncHost) updateHostsStaticWithTopo(nCtx contextx.IContext, hosts ...*types.Host) error {
	if err := act.storageHost.UpdateHostStaticFields(nCtx, types.UpdateAllHostStaticFields(), hosts...); err != nil {
		return err
	}

	hostRelations := conv.SliceToSlice(hosts, func(host *types.Host) *types.HostTopoRelation {
		return &types.HostTopoRelation{
			HostID: host.HostID,
			Topo:   host.Static.Topo,
		}
	})
	if err := act.storageHost.UpsertHostTopo(nCtx, hostRelations...); err != nil {
		return err
	}

	return nil
}

func (act *actionSyncHost) compareData(cmdbData, dbData []*types.Host) (
	[]*types.Host, []*types.Host, []int64, error) {

	updateHosts := make([]*types.Host, 0)
	insertHosts := make([]*types.Host, 0)
	deleteHostIDs := make([]int64, 0)

	// Convert CMDB data into maps for quick lookup
	// some time we will meet duplicate hostid, but this is acceptable, we will update it in the next loop.
	cmdbHostMap, err := conv.SliceToMapIgnore(cmdbData, func(host *types.Host) int64 {
		return host.HostID
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("convert cmdb data to map failed: %w", err)
	}

	// Handle hosts in the database
	for _, host := range dbData {
		if cmdbHost, exists := cmdbHostMap[host.HostID]; exists {
			// The host exists in the CMDB and is added to the update list
			// Note: cmdbHost is used here instead of host, because we want to use the CMDB data as the prevailing one
			updateHosts = append(updateHosts, cmdbHost)
			delete(cmdbHostMap, host.HostID)
		} else {
			// The host does not exist in the CMDB and should be removed from the database
			deleteHostIDs = append(deleteHostIDs, host.HostID)
		}
	}

	// Handling Hosts that Only Exist in the CMDB (New Hosts)
	for _, host := range cmdbHostMap {
		// when the host synchronizes from the CMDB for the first time, the agentid needs to be updated to dynamic
		syncDataUtils.FillHostDynamicAgentID(host)

		// when the host synchronizes from the CMDB for the first time, the ops info needs to be updated to dynamic
		syncDataUtils.FillHostDynamicOpsInfo(host)

		insertHosts = append(insertHosts, host)
	}

	return updateHosts, insertHosts, deleteHostIDs, nil
}

func fillDefaultAdvertiseIP(host, dbHost *types.Host) bool {
	if host == nil || host.Dynamic == nil || host.Static == nil {
		return false
	}
	if !isProxyHost(host, dbHost) {
		return false
	}

	if dbHost != nil && dbHost.Dynamic != nil {
		if dbHost.Dynamic.AdvertiseIP != "" {
			host.Dynamic.AdvertiseIP = dbHost.Dynamic.AdvertiseIP
		}
		if dbHost.Dynamic.AdvertiseIPV6 != "" {
			host.Dynamic.AdvertiseIPV6 = dbHost.Dynamic.AdvertiseIPV6
		}
	}

	ipv4Updated := fillDefaultAdvertiseIPv4(host, dbHost)
	ipv6Updated := fillDefaultAdvertiseIPv6(host, dbHost)

	return ipv4Updated || ipv6Updated
}

func fillDefaultAdvertiseIPv4(host, dbHost *types.Host) bool {
	if host.Dynamic.AdvertiseIP == "" && len(host.Static.InnerIPList) > 0 {
		host.Dynamic.AdvertiseIP = host.Static.InnerIPList[0]
		return true
	}
	if dbHost != nil && host.Dynamic.AdvertiseIP != "" &&
		(dbHost.Dynamic == nil || dbHost.Dynamic.AdvertiseIP == "") {

		return true
	}

	return false
}

func fillDefaultAdvertiseIPv6(host, dbHost *types.Host) bool {
	if host.Dynamic.AdvertiseIPV6 == "" && len(host.Static.InnerIPV6List) > 0 {
		host.Dynamic.AdvertiseIPV6 = host.Static.InnerIPV6List[0]
		return true
	}
	if dbHost != nil && host.Dynamic.AdvertiseIPV6 != "" &&
		(dbHost.Dynamic == nil || dbHost.Dynamic.AdvertiseIPV6 == "") {

		return true
	}

	return false
}

func isProxyHost(host, dbHost *types.Host) bool {
	if host.Dynamic.NodeRole == types.NodeRoleProxy {
		return true
	}
	if dbHost != nil && dbHost.Dynamic != nil {
		return dbHost.Dynamic.NodeRole == types.NodeRoleProxy
	}

	return false
}

func fillDefaultAdvertiseIPs(hosts, dbData []*types.Host) []*types.Host {
	dbHostMap := make(map[int64]*types.Host, len(dbData))
	for _, host := range dbData {
		dbHostMap[host.HostID] = host
	}

	backfillHosts := make([]*types.Host, 0)
	for _, host := range hosts {
		if fillDefaultAdvertiseIP(host, dbHostMap[host.HostID]) {
			backfillHosts = append(backfillHosts, host)
		}
	}

	return backfillHosts
}

func (act *actionSyncHost) fillDefaultLoginUsers(ctx *action.InstanceContext, hosts, dbData []*types.Host) []*types.Host {
	dbHostMap := make(map[int64]*types.Host, len(dbData))
	for _, host := range dbData {
		dbHostMap[host.HostID] = host
	}

	backfillHosts := make([]*types.Host, 0)
	for _, host := range hosts {
		if dbHost, exists := dbHostMap[host.HostID]; exists && dbHost.Dynamic.LoginUser != "" {
			host.Dynamic.LoginUser = dbHost.Dynamic.LoginUser
			continue
		}

		if host.Dynamic.LoginUser != "" {
			continue
		}

		defaultUser, err := criteria.DefaultAdminUser(criteria.OSType(host.Static.OSType))
		if err != nil {
			logger.G.Sys().With("host-id", host.HostID).WithErr(err).Warn("failed to get default user for synced host")
			continue
		}

		host.Dynamic.LoginUser = string(defaultUser)
		backfillHosts = append(backfillHosts, host)
	}

	ctx.Data.Log().
		Zh("主机登录用户修复完成，需修补 %d 台", len(backfillHosts)).
		En("host login user repair done, %d hosts need repair", len(backfillHosts)).
		Info()

	return backfillHosts
}

func getAndFillEmptyDynamicAgentIDByStaticSyncedAgentID(
	updateData, dbData []*types.Host) (
	[]*types.Host, error) {

	dbDataMap, err := conv.SliceToMapIgnore(dbData, func(host *types.Host) int64 {
		return host.HostID
	})
	if err != nil {
		return nil, fmt.Errorf("convert cmdb data to map failed: %w", err)
	}

	hosts := make([]*types.Host, 0)
	for _, host := range updateData {
		dbHost, exists := dbDataMap[host.HostID]
		if !exists {
			continue
		}

		if dbHost.Dynamic.AgentID == "" && host.Static.SyncedAgentID != "" {
			host.Dynamic.AgentID = host.Static.SyncedAgentID
			hosts = append(hosts, host)
		}
	}

	return hosts, nil
}

func (act *actionSyncHost) tryTriggerCorrectAgentID(
	ctx *action.InstanceContext,
	std *syncDataUtils.SyncDataActionStandarder,
	updateHosts, dbData []*types.Host,
) error {

	hostIDs := filterNeedCorrectAgentIDHostIDs(updateHosts, dbData)
	if len(hostIDs) == 0 {
		return nil
	}

	ctx.Data.Log().
		Zh("发现 %d 台主机需要修正 Agent ID，触发修正任务", len(hostIDs)).
		En("found %d hosts need to correct agent id, triggering correction task", len(hostIDs)).
		Info()

	triggerID, err := act.syncIface.LaunchSyncCorrectAgentID(std.Context(), hostIDs...)
	if err != nil {
		logger.G.Sys().Ctx(std.Context()).WithErr(err).Error("failed to launch correct agent id task")
		return err
	}

	logger.G.Sys().Ctx(std.Context()).
		With("trigger-id", triggerID, "host-count", len(hostIDs)).
		Info("launched correct agent id task")

	return nil
}

func filterNeedCorrectAgentIDHostIDs(updateHosts, dbData []*types.Host) []int64 {
	dbDataMap := make(map[int64]*types.Host, len(dbData))
	for _, host := range dbData {
		dbDataMap[host.HostID] = host
	}

	hostIDs := make([]int64, 0)
	for _, host := range updateHosts {
		dbHost, exists := dbDataMap[host.HostID]
		if !exists {
			continue
		}
		if dbHost.Dynamic == nil || dbHost.Dynamic.AgentID == "" {
			continue
		}
		if host.Static == nil || host.Static.SyncedAgentID == "" {
			continue
		}
		if host.Static.SyncedAgentID == dbHost.Dynamic.AgentID {
			continue
		}
		if dbHost.Dynamic.NodeStatus == types.NodeStatusRunning {
			continue
		}

		hostIDs = append(hostIDs, host.HostID)
	}

	return hostIDs
}

func (act *actionSyncHost) tryUpdateHostProcessBizID(std *syncDataUtils.SyncDataActionStandarder, bizID int64,
	cmdbData ...*types.Host) error {

	hostIDs := conv.SliceToSlice(cmdbData, func(host *types.Host) int64 {
		return host.HostID
	})

	if err := act.storageProcess.UpdateProcessManyHostBizID(std.Context(), bizID, hostIDs...); err != nil {
		std.InstanceData().Log().
			Zh("更新主机关联的进程业务ID失败，主机ID列表: %v, 错误: %v", hostIDs, err).
			En("failed to update process bizID related to hosts, hostIDs: %v, error: %v", hostIDs, err).
			Error()

		return err
	}

	std.InstanceData().Log().
		Zh("成功更新主机关联的进程业务ID，主机ID列表: %v", hostIDs).
		En("successfully updated process bizID related to hosts, hostIDs: %v", hostIDs).
		Info()

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionSyncHost) DisplayNameZh() string { return "同步主机" }

// DisplayNameEn returns the English display name of the action.
func (act *actionSyncHost) DisplayNameEn() string { return "Sync Host" }
