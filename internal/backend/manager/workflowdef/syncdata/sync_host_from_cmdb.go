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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"time"

	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
)

// NewActionSyncHostFromCMDB ...
func NewActionSyncHostFromCMDB(cmdbHandler cmdb.IHandler, iDaoHost topoStg.IDaoHost) operengine.ActionDef {
	return &syncHostFromCMDB{
		cmdbHandler: cmdbHandler,
		iDaoHost:    iDaoHost,
	}
}

// SyncHostFromCMDBParam ...
type SyncHostFromCMDBParam struct {
	BizID    int64  `json:"biz_id"`
	TenantID string `json:"tenant_id"`
}

// syncHostFromCMDB ...
type syncHostFromCMDB struct {
	cmdbHandler cmdb.IHandler
	iDaoHost    topoStg.IDaoHost
}

// Name ...
func (action *syncHostFromCMDB) Name() string {
	return ActionNameSyncHostFromCMDB
}

// Version ...
func (action *syncHostFromCMDB) Version() string {
	return "v1"
}

// Description ...
func (action *syncHostFromCMDB) Description() string {
	return "Get the host information of the designated business from CMDB, and update to the database."
}

// Timeout ...
func (action *syncHostFromCMDB) Timeout() time.Duration {
	return 5 * time.Minute // nolint: mnd
}

// Tags ...
func (action *syncHostFromCMDB) Tags() []operengine.ActionTag {
	return []operengine.ActionTag{}
}

// MaxRetryCount ...
func (action *syncHostFromCMDB) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn ...
func (action *syncHostFromCMDB) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do ...
func (action *syncHostFromCMDB) Do(ctx *operengine.ActionInstContext) error {
	param := new(SyncHostFromCMDBParam)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	tenantCtx, err := tenant.SetID(ctx.Ctx, param.TenantID)
	if err != nil {
		return err
	}

	var cmdbData, dbData []*types.Host

	gp := gopool.NewPool()
	gp.Go(func() error {
		cmdbData, err = action.cmdbHandler.ListBizHosts(tenantCtx, param.BizID, types.UnlimitedPage())
		if err != nil {
			return fmt.Errorf("list host from cmdb failed, err: %w", err)
		}

		return nil
	})

	gp.Go(func() error {
		dbData, _, err = action.iDaoHost.ListHost(tenantCtx, types.UnlimitedPage(), &types.HostCondition{
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

	upsertHosts, deleteHostIDs, err := action.compareData(cmdbData, dbData)
	if err != nil {
		return err
	}

	if err = action.iDaoHost.UpsertManyHostStatic(tenantCtx, upsertHosts...); err != nil {
		return err
	}

	if err = action.iDaoHost.DeleteManyHost(tenantCtx, deleteHostIDs...); err != nil {
		return err
	}

	return nil
}

func (action *syncHostFromCMDB) compareData(cmdbData, dbData []*types.Host) ([]*types.Host, []int64, error) {
	upsertHosts := make([]*types.Host, 0)
	deleteHostIDs := make([]int64, 0)

	// Convert CMDB data into maps for quick lookup
	cmdbHostMap, err := conv.SliceToMap(cmdbData, func(host *types.Host) int64 {
		return host.HostID
	})
	if err != nil {
		return nil, nil, fmt.Errorf("convert cmdb data to map failed, err: %w", err)
	}

	// Handle hosts in the database
	for _, host := range dbData {
		if cmdbHost, exists := cmdbHostMap[host.HostID]; exists {
			// The host exists in the CMDB and is added to the update list
			// Note: cmdbHost is used here instead of host, because we want to use the CMDB data as the prevailing one
			upsertHosts = append(upsertHosts, cmdbHost)
			delete(cmdbHostMap, host.HostID)
		} else {
			// The host does not exist in the CMDB and should be removed from the database
			deleteHostIDs = append(deleteHostIDs, host.HostID)
		}
	}

	// Handling Hosts that Only Exist in the CMDB (New Hosts)
	for _, host := range cmdbHostMap {
		upsertHosts = append(upsertHosts, host)
	}

	return upsertHosts, deleteHostIDs, nil
}
