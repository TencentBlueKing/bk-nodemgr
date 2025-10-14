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

	syncDataUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata/utils"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameSyncBusiness defines the action name.
	ActionNameSyncBusiness = "sync_business"
)

// NewActionSyncBusiness creates a new actionSyncBusiness.
func NewActionSyncBusiness(capability *Capability) action.Definition {
	return &actionSyncBusiness{
		cmdbHandler:     capability.CMDBHandler,
		storageBusiness: capability.StorageTopo,
	}
}

// ActionParamSyncBusiness the action's param.
type ActionParamSyncBusiness struct {
	syncDataUtils.SyncDataActionStandardParam
}

type actionSyncBusiness struct {
	cmdbHandler     cmdb.IHandler
	storageBusiness topoStg.IStorageBusiness
}

// Name returns the name of the action.
func (act *actionSyncBusiness) Name() string {
	return ActionNameSyncBusiness
}

// Version returns the version of the action.
func (act *actionSyncBusiness) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionSyncBusiness) Description() string {
	return "sync business info from cmdb and update to storage"
}

// Timeout returns the timeout of this action.
func (act *actionSyncBusiness) Timeout() time.Duration {
	return 1 * time.Minute
}

// MaxRetryCount returns the max retry count of this action.
func (act *actionSyncBusiness) MaxRetryCount() uint {
	return 2 // nolint: mnd
}

// DelayFn returns the delay of this action.
func (act *actionSyncBusiness) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Tags returns the tags of this action.
func (act *actionSyncBusiness) Tags() []action.Tag {
	return []action.Tag{}
}

// Do the action.
func (act *actionSyncBusiness) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamSyncBusiness)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := syncDataUtils.NewSyncDataActionStandarder()
	if err = std.Initialize(ctx, param.SyncDataActionStandardParam); err != nil {
		return err
	}

	gp := gopool.NewPool()
	gp.SetLimit(10) // nolint: mnd

	executor := pageexecutor.NewPageExecutor[*types.Business](500, 1*time.Hour) // nolint: mnd
	fn := func(_ context.Context, p types.Page) ([]*types.Business, error) {
		bizs, err := act.cmdbHandler.SearchBusiness(std.Context(), p)
		if err != nil {
			return nil, err
		}

		return bizs, nil
	}

	result, err := executor.Execute(ctx.Ctx, types.UnlimitedPage(), fn)
	if err != nil {
		return err
	}

	if err = act.storageBusiness.UpsertManyBusiness(std.Context(), result.Items...); err != nil {
		return err
	}

	return nil
}
