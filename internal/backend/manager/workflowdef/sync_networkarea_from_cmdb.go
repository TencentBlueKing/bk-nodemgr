/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package workflowdef

import (
	"context"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
)

// NewActionSyncNetworkAreaFromCMDB get a new action.
func NewActionSyncNetworkAreaFromCMDB(cmdbHandler cmdb.Handler, topoStorage topo.Storage) operengine.ActionDef {
	return &syncNetworkAreaFromCMDB{
		cmdbHandler: cmdbHandler,
		topoStorage: topoStorage,
	}
}

// SyncNetworkAreaFromCMDBParam describes the parameters.
type SyncNetworkAreaFromCMDBParam struct {
	TenantID string `json:"tenant_id"`
}

// syncNetworkAreaFromCMDB defines the action.
type syncNetworkAreaFromCMDB struct {
	cmdbHandler cmdb.Handler
	topoStorage topo.Storage
}

// Name returns the name of the action.
func (s *syncNetworkAreaFromCMDB) Name() string {
	return SyncNetworkAreaFromCMDB
}

// Version returns the version of the action.
func (s *syncNetworkAreaFromCMDB) Version() string {
	return "v1"
}

// Description returns the description of the action.
func (s *syncNetworkAreaFromCMDB) Description() string {
	return "Get the networkareas which also called cloudarea from CMDB, and update to the database."
}

// Timeout returns the timeout of the action.
func (s *syncNetworkAreaFromCMDB) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (s *syncNetworkAreaFromCMDB) Tags() []operengine.ActionTag {
	return []operengine.ActionTag{}
}

// MaxRetryCount returns the retry count of the action.
func (s *syncNetworkAreaFromCMDB) MaxRetryCount() uint {
	return 3
}

// DelayFn returns the delay function.
func (s *syncNetworkAreaFromCMDB) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do does the action.
func (s *syncNetworkAreaFromCMDB) Do(ctx *operengine.ActionInstContext) error {
	param := new(SyncNetworkAreaFromCMDBParam)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	tenantCtx, err := tenant.SetID(ctx.Ctx, param.TenantID)
	if err != nil {
		return err
	}

	executor := runtime.NewPageExecutor[*types.NetworkArea](500, 1*time.Hour)
	fn := func(ctx context.Context, p types.Page) ([]*types.NetworkArea, error) {
		networkareas, err := s.cmdbHandler.SearchNetworkArea(ctx, p)
		if err != nil {
			return nil, err
		}

		return networkareas, nil
	}

	result, err := executor.Execute(tenantCtx, types.UnlimitedPage(), fn)
	if err != nil {
		return err
	}

	if err = s.topoStorage.UpsertManyNetworkArea(tenantCtx, result.Items...); err != nil {
		return err
	}

	return nil
}
