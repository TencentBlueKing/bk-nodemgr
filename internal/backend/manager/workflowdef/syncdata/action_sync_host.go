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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameSyncHostFromCMDB defines the action name.
	ActionNameSyncHostFromCMDB = "sync_host_from_cmdb"
)

// NewActionSyncHostFromCMDB ...
func NewActionSyncHostFromCMDB(cmdbHandler cmdb.IHandler, storageHost topo.IStorageHost) action.Definition {
	return &actionSyncHostFromCMDB{
		cmdbHandler: cmdbHandler,
		storageHost: storageHost,
	}
}

// SyncHostFromCMDBParam ...
type SyncHostFromCMDBParam struct {
	BizID    int64  `json:"biz_id"`
	TenantID string `json:"tenant_id"`
	Operator string `json:"operator"`
}

type actionSyncHostFromCMDB struct {
	cmdbHandler cmdb.IHandler
	storageHost topo.IStorageHost
}

// Name ...
func (act *actionSyncHostFromCMDB) Name() string {
	return ActionNameSyncHostFromCMDB
}

// Version ...
func (act *actionSyncHostFromCMDB) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description ...
func (act *actionSyncHostFromCMDB) Description() string {
	return "get the host information of the designated business from cmdb, and update to the database"
}

// Timeout ...
func (act *actionSyncHostFromCMDB) Timeout() time.Duration {
	return 5 * time.Minute // nolint: mnd
}

// Tags ...
func (act *actionSyncHostFromCMDB) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount ...
func (act *actionSyncHostFromCMDB) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn ...
func (act *actionSyncHostFromCMDB) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do ...
func (act *actionSyncHostFromCMDB) Do(ctx *action.InstanceContext) error {
	param := new(SyncHostFromCMDBParam)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	tenantUserCtx := contextx.NewTenantUserContext(ctx.Ctx, param.TenantID, param.Operator)

	var cmdbData, dbData []*types.Host
	gp := gopool.NewPool()
	gp.Go(func() error {
		cmdbData, err = act.cmdbHandler.ListBizHosts(tenantUserCtx, param.BizID, types.UnlimitedPage())
		if err != nil {
			return fmt.Errorf("list host from cmdb failed, err: %w", err)
		}

		return nil
	})

	tenantCtx, err := tenant.SetID(ctx.Ctx, param.TenantID)
	if err != nil {
		return err
	}

	gp.Go(func() error {
		dbData, _, err = act.storageHost.ListHost(tenantCtx, types.UnlimitedPage(), &types.HostCondition{
			ExactInclude: &types.HostExactFields{
				BizID: []int64{param.BizID},
			},
		})
		if err != nil {
			return fmt.Errorf("list host from db failed, err: %w", err)
		}

		return nil
	})

	if err = gp.Wait(); err != nil {
		return err
	}

	ctx.Data.LogI(fmt.Sprintf("find %d hosts from cmdb, %d hosts in db", len(cmdbData), len(dbData)))
	updateHosts, insertHosts, deleteHostIDs, err := act.compareData(cmdbData, dbData)
	if err != nil {
		return err
	}

	ctx.Data.LogI(
		fmt.Sprintf("comapred hosts, %d hosts need to update, %d hosts need to insert, %d hosts need to delete",
			len(updateHosts), len(insertHosts), len(deleteHostIDs)))

	if err = act.storageHost.UpsertManyHostStatic(tenantCtx, updateHosts...); err != nil {
		return err
	}

	if err = act.storageHost.UpsertManyHost(tenantCtx, insertHosts...); err != nil {
		return err
	}

	if err = act.storageHost.DeleteManyHost(tenantCtx, deleteHostIDs...); err != nil {
		return err
	}

	return nil
}

func (act *actionSyncHostFromCMDB) compareData(cmdbData, dbData []*types.Host) (
	[]*types.Host, []*types.Host, []int64, error) {

	updateHosts := make([]*types.Host, 0)
	insertHosts := make([]*types.Host, 0)
	deleteHostIDs := make([]int64, 0)

	// Convert CMDB data into maps for quick lookup
	cmdbHostMap, err := conv.SliceToMap(cmdbData, func(host *types.Host) int64 {
		return host.HostID
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("convert cmdb data to map failed, err: %w", err)
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
