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
	"time"

	syncDataUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameSyncHost defines the action name.
	ActionNameSyncHost = "sync_host"
)

// NewActionSyncHost creates a new actionSyncHost.
func NewActionSyncHost(capability *Capability) action.Definition {
	return &actionSyncHost{
		cmdbHandler:    capability.CMDBHandler,
		storageHost:    capability.StorageTopo,
		storageProcess: capability.StoragePlugin,
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
	return 5 * time.Minute // nolint: mnd
}

// MaxRetryCount returns the max retry count of this action.
func (act *actionSyncHost) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn returns the delay of this action.
func (act *actionSyncHost) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
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
		cmdbData, err = act.cmdbHandler.ListBizHosts(std.Context(), param.BizID, types.UnlimitedPage())
		if err != nil {
			return fmt.Errorf("list host from cmdb failed: %w", err)
		}

		return nil
	})

	gp.Go(func() error {
		selection := &types.HostFieldSelection{
			// in compare logic we only need the host_id field.
			HostID:        true,
			NetworkAreaID: false,
			InnerIPList:   false,
			InnerIPV6List: false,
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

	ctx.Data.Log().
		Zh("对比完成，需更新 %d 台、新增 %d 台、删除 %d 台主机", len(updateHosts), len(insertHosts), len(deleteHostIDs)).
		En("compared hosts, %d hosts need to update, %d hosts need to insert, %d hosts need to delete",
			len(updateHosts), len(insertHosts), len(deleteHostIDs)).
		Info()

	if err = act.storageHost.UpsertManyHostStatic(std.Context(), updateHosts...); err != nil {
		return err
	}

	if err = act.storageHost.UpsertManyHost(std.Context(), insertHosts...); err != nil {
		return err
	}

	if err = act.storageHost.DeleteManyHost(std.Context(), deleteHostIDs...); err != nil {
		return err
	}

	if err := act.tryUpdateHostProcessBizID(std, param.BizID, cmdbData...); err != nil {
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
	cmdbHostMap, err := conv.SliceToMap(cmdbData, func(host *types.Host) int64 {
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
		if host.Static.SyncedAgentID != "" {
			host.Dynamic.AgentID = host.Static.SyncedAgentID
		}
		insertHosts = append(insertHosts, host)
	}

	return updateHosts, insertHosts, deleteHostIDs, nil
}

func (act *actionSyncHost) tryUpdateHostProcessBizID(std *syncDataUtils.SyncDataActionStandarder, bizID int64, cmdbData ...*types.Host) error {
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
