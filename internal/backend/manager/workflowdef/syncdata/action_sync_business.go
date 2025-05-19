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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameSyncBizFromCMDB defines the action name.
	ActionNameSyncBizFromCMDB = "sync_biz_from_cmdb"
)

// NewActionSyncBusinessFromCMDB creates a new syncBusinessFromCMDB.
func NewActionSyncBusinessFromCMDB(cmdbHandler cmdb.IHandler, topoStorage topoStg.IStorage,
	logger logger.Logger) action.Definition {

	return &actionSyncBusinessFromCMDB{
		cmdbHandler: cmdbHandler,
		topoStorage: topoStorage,
		logger:      logger,
	}
}

// SyncBizFromCMDBParam the action's param.
type SyncBizFromCMDBParam struct {
	TenantID string `json:"tenant_id"`
}

type actionSyncBusinessFromCMDB struct {
	cmdbHandler cmdb.IHandler
	topoStorage topoStg.IStorage
	logger      logger.Logger
}

// Name returns the name of the action.
func (act *actionSyncBusinessFromCMDB) Name() string {
	return ActionNameSyncBizFromCMDB
}

// Version returns the version of the action.
func (act *actionSyncBusinessFromCMDB) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionSyncBusinessFromCMDB) Description() string {
	return "sync business info from cmdb and update to storage"
}

// Timeout returns the timeout of this action.
func (act *actionSyncBusinessFromCMDB) Timeout() time.Duration {
	return 1 * time.Minute
}

// MaxRetryCount returns the max retry count of this action.
func (act *actionSyncBusinessFromCMDB) MaxRetryCount() uint {
	return 2 // nolint: mnd
}

// DelayFn returns the delay of this action.
func (act *actionSyncBusinessFromCMDB) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Tags returns the tags of this action.
func (act *actionSyncBusinessFromCMDB) Tags() []action.Tag {
	return []action.Tag{}
}

// Do the action.
func (act *actionSyncBusinessFromCMDB) Do(ctx *action.InstanceContext) error {
	param := new(SyncBizFromCMDBParam)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	gp := gopool.NewPool()
	gp.SetLimit(10) // nolint: mnd

	executor := runtime.NewPageExecutor[*types.Business](500, 1*time.Hour) // nolint: mnd
	fn := func(ctx context.Context, p types.Page) ([]*types.Business, error) {
		bizs, err := act.cmdbHandler.SearchBusiness(ctx, p)
		if err != nil {
			return nil, err
		}

		return bizs, nil
	}

	tenantCtx, _ := tenant.SetID(ctx.Ctx, param.TenantID)
	result, err := executor.Execute(tenantCtx, types.UnlimitedPage(), fn)
	if err != nil {
		return err
	}

	if err = act.topoStorage.UpsertManyBusiness(tenantCtx, result.Items...); err != nil {
		return err
	}

	return nil
}
