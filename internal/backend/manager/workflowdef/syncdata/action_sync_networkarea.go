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
	"time"

	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameSyncNetworkAreaFromCMDB defines the action name.
	ActionNameSyncNetworkAreaFromCMDB = "sync_networkarea_from_cmdb"
)

// NewActionSyncNetworkAreaFromCMDB get a new action.
func NewActionSyncNetworkAreaFromCMDB(cmdbHandler cmdb.IHandler,
	storageNetworkArea topoStg.IStorageNetworkArea) action.Definition {

	return &actionSyncNetworkAreaFromCMDB{
		cmdbHandler:        cmdbHandler,
		storageNetworkArea: storageNetworkArea,
	}
}

// SyncNetworkAreaFromCMDBParam describes the parameters.
type SyncNetworkAreaFromCMDBParam struct {
	TenantID string `json:"tenant_id"`
	Operator string `json:"operator"`
}

type actionSyncNetworkAreaFromCMDB struct {
	cmdbHandler        cmdb.IHandler
	storageNetworkArea topoStg.IStorageNetworkArea
}

// Name returns the name of the action.
func (act *actionSyncNetworkAreaFromCMDB) Name() string {
	return ActionNameSyncNetworkAreaFromCMDB
}

// Version returns the version of the action.
func (act *actionSyncNetworkAreaFromCMDB) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionSyncNetworkAreaFromCMDB) Description() string {
	return "get the networkareas which also called cloudarea from cmdb, and update to the database"
}

// Timeout returns the timeout of the action.
func (act *actionSyncNetworkAreaFromCMDB) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionSyncNetworkAreaFromCMDB) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the retry count of the action.
func (act *actionSyncNetworkAreaFromCMDB) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn returns the delay function.
func (act *actionSyncNetworkAreaFromCMDB) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do does the action.
func (act *actionSyncNetworkAreaFromCMDB) Do(ctx *action.InstanceContext) error {
	param := new(SyncNetworkAreaFromCMDBParam)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	nCtx := contextx.New(ctx.Ctx, contextx.WithTenantID(param.TenantID), contextx.WithBKUsername(param.Operator))
	executor := pageexecutor.NewPageExecutor[*types.NetworkArea](500, 1*time.Hour) // nolint: mnd
	fn := func(_ context.Context, p types.Page) ([]*types.NetworkArea, error) {
		networkareas, err := act.cmdbHandler.SearchNetworkArea(nCtx, p)
		if err != nil {
			return nil, err
		}

		return networkareas, nil
	}

	result, err := executor.Execute(nCtx, types.UnlimitedPage(), fn)
	if err != nil {
		return err
	}

	if err = act.storageNetworkArea.UpsertManyNetworkArea(nCtx, result.Items...); err != nil {
		return err
	}

	return nil
}
