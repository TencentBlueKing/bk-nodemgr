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
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameSyncHostTopoRelation defines the action name.
	ActionNameSyncHostTopoRelation = "sync_host_topo_relation"
)

// NewActionSyncHostTopoRelation creates a new actionSyncHostTopoRelation.
func NewActionSyncHostTopoRelation(capability *Capability) action.Definition {
	return &actionSyncHostTopoRelation{
		cmdbHandler: capability.CMDBHandler,
		storageHost: capability.StorageTopo,
	}
}

// ActionParamSyncHostTopoRelation defines the action's param.
type ActionParamSyncHostTopoRelation struct {
	syncDataUtils.SyncDataActionStandardParam

	BizID int64 `json:"biz_id"`
}

type actionSyncHostTopoRelation struct {
	cmdbHandler cmdb.IHandler
	storageHost topoStg.IStorageHost
}

// Name returns the name of the action.
func (act *actionSyncHostTopoRelation) Name() string {
	return ActionNameSyncHostTopoRelation
}

// Version returns the version of the action.
func (act *actionSyncHostTopoRelation) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionSyncHostTopoRelation) Description() string {
	return "get the host topo relation of the designated business from cmdb, and update module and set to the database"
}

// Timeout returns the timeout of this action.
func (act *actionSyncHostTopoRelation) Timeout() time.Duration {
	return 10 * time.Minute // nolint: mnd
}

// MaxRetryCount returns the max retry count of this action.
func (act *actionSyncHostTopoRelation) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn returns the delay of this action.
func (act *actionSyncHostTopoRelation) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Minute)
	}
}

// Tags returns the tags of this action.
func (act *actionSyncHostTopoRelation) Tags() []action.Tag {
	return []action.Tag{}
}

// Do the action.
func (act *actionSyncHostTopoRelation) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamSyncHostTopoRelation)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	std := syncDataUtils.NewSyncDataActionStandarder()
	if err = std.Initialize(ctx, param.SyncDataActionStandardParam); err != nil {
		return err
	}

	hosts, err := act.cmdbHandler.ListBizHostTopoRelations(std.Context(), param.BizID, types.UnlimitedPage())
	if err != nil {
		return fmt.Errorf("list host topo relation from cmdb failed: %w", err)
	}

	ctx.Data.Log().
		Zh("从 CMDB 获取到 %d 条主机拓扑关系", len(hosts)).
		En("find %d host topo relations from cmdb", len(hosts)).
		Info()

	hosts = mergeHostTopoRelations(hosts)
	if err = batchHandleHosts(std.Context(), hosts, func(hosts ...*types.Host) error {
		return act.storageHost.UpdateHostStaticFields(std.Context(), types.HostStaticFields{
			BizID:    true,
			ModuleID: true,
			SetID:    true,
			Topo:     true,
		}, hosts...)
	}); err != nil {
		return err
	}

	return nil
}

func mergeHostTopoRelations(hosts []*types.Host) []*types.Host {
	hostMap := make(map[int64]*types.Host)
	for _, host := range hosts {
		if host == nil || host.Static == nil {
			continue
		}

		merged, ok := hostMap[host.HostID]
		if !ok {
			merged = &types.Host{
				TenantID: host.TenantID,
				HostID:   host.HostID,
				Static: &types.HostStatic{
					BizID:    host.Static.BizID,
					SetID:    host.Static.SetID,
					ModuleID: host.Static.ModuleID,
					Topo:     make([]types.HostTopo, 0, len(host.Static.Topo)),
				},
			}
			hostMap[host.HostID] = merged
		}

		merged.Static.Topo = append(merged.Static.Topo, host.Static.Topo...)
	}

	data := make([]*types.Host, 0, len(hostMap))
	for _, host := range hostMap {
		data = append(data, host)
	}

	return data
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionSyncHostTopoRelation) DisplayNameZh() string { return "同步主机拓扑关系" }

// DisplayNameEn returns the English display name of the action.
func (act *actionSyncHostTopoRelation) DisplayNameEn() string { return "Sync Host Topo Relation" }
