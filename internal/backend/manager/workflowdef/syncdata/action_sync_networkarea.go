/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package syncdata

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"

	syncDataUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata/utils"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameSyncNetworkArea defines the action name.
	ActionNameSyncNetworkArea = "sync_networkarea"
)

// NewActionSyncNetworkArea get a new action.
func NewActionSyncNetworkArea(capability *Capability) action.Definition {
	return &actionSyncNetworkArea{
		cmdbHandler:        capability.CMDBHandler,
		storageNetworkArea: capability.StorageTopo,
	}
}

// ActionParamSyncNetworkArea describes the parameters.
type ActionParamSyncNetworkArea struct {
	syncDataUtils.SyncDataActionStandardParam
}

type actionSyncNetworkArea struct {
	cmdbHandler        cmdb.IHandler
	storageNetworkArea topoStg.IStorageNetworkArea
}

// Name returns the name of the action.
func (act *actionSyncNetworkArea) Name() string {
	return ActionNameSyncNetworkArea
}

// Version returns the version of the action.
func (act *actionSyncNetworkArea) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionSyncNetworkArea) Description() string {
	return "get the networkareas which also called cloudarea from cmdb, and update to the database"
}

// Timeout returns the timeout of the action.
func (act *actionSyncNetworkArea) Timeout() time.Duration {
	return 10 * time.Minute // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionSyncNetworkArea) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the retry count of the action.
func (act *actionSyncNetworkArea) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn returns the delay function.
func (act *actionSyncNetworkArea) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Do does the action.
func (act *actionSyncNetworkArea) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamSyncNetworkArea)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := syncDataUtils.NewSyncDataActionStandarder()
	if err = std.Initialize(ctx, param.SyncDataActionStandardParam); err != nil {
		return err
	}

	executor := pageexecutor.NewPageExecutor[*types.NetworkArea](500, 1*time.Hour) // nolint: mnd
	fn := func(nCtx contextx.IContext, p types.Page) ([]*types.NetworkArea, error) {
		networkAreas, err := act.cmdbHandler.SearchNetworkArea(nCtx, p)
		if err != nil {
			return nil, err
		}

		return networkAreas, nil
	}

	result, err := executor.Execute(std.Context(), types.UnlimitedPage(), fn)
	if err != nil {
		return err
	}

	if err = act.storageNetworkArea.UpsertManyNetworkArea(std.Context(), result.Items...); err != nil {
		return err
	}

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionSyncNetworkArea) DisplayNameZh() string { return "同步管控区域" }

// DisplayNameEn returns the English display name of the action.
func (act *actionSyncNetworkArea) DisplayNameEn() string { return "Sync Network Area" }
