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
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/diff"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/usermanager"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameSyncTenant defines the action name.
	ActionNameSyncTenant = "sync_tenant"
)

// NewActionSyncTenant creates a new actionSyncTenant.
func NewActionSyncTenant(capability *Capability) action.Definition {
	return &actionSyncTenant{
		userManagerHandler: capability.UserManagerHandler,
		daoTenant:          capability.StorageTenant,
	}
}

// ActionParamSyncTenant the action's param.
type ActionParamSyncTenant struct {
	syncDataUtils.SyncDataActionStandardParam
}

type actionSyncTenant struct {
	userManagerHandler usermanager.IHandler
	daoTenant          tenant.IStorage
}

// Name returns the name of the action.
func (act *actionSyncTenant) Name() string {
	return ActionNameSyncTenant
}

// Version returns the version of the action.
func (act *actionSyncTenant) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionSyncTenant) Description() string {
	return "sync tenant info from user manager and update to storage"
}

// Timeout returns the timeout of this action.
func (act *actionSyncTenant) Timeout() time.Duration {
	return 1 * time.Minute
}

// MaxRetryCount returns the max retry count of this action.
func (act *actionSyncTenant) MaxRetryCount() uint {
	return 2 // nolint: mnd
}

// DelayFn returns the delay of this action.
func (act *actionSyncTenant) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Tags returns the tags of this action.
func (act *actionSyncTenant) Tags() []action.Tag {
	return []action.Tag{}
}

// Do the action.
// nolint: funlen,gocognit
func (act *actionSyncTenant) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamSyncTenant)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := syncDataUtils.NewSyncDataActionStandarder()
	if err = std.Initialize(ctx, param.SyncDataActionStandardParam); err != nil {
		return err
	}

	nCtx := std.Context()
	gp := gopool.NewPool()

	var (
		latestTenants   []*types.Tenant
		originTenants   []*types.Tenant
		latestTenantMap map[string]*types.Tenant
		originTenantMap map[string]*types.Tenant
	)

	gp.Go(func() error {
		latestTenants, err = act.userManagerHandler.ListALLTenants(nCtx)
		if err != nil {
			return fmt.Errorf("failed to list all tenants from user manager: %w", err)
		}

		latestTenantMap, err = conv.SliceToMap(latestTenants, func(v *types.Tenant) string {
			return v.ID
		})
		if err != nil {
			return fmt.Errorf("failed to convert latest tenants to map: %w", err)
		}

		return nil
	})

	gp.Go(func() error {
		originTenants, err = act.daoTenant.ListAllEnabledTenants(nCtx)
		if err != nil {
			return fmt.Errorf("failed to list all enabled tenants from storage: %w", err)
		}

		originTenantMap, err = conv.SliceToMap(originTenants, func(v *types.Tenant) string {
			return v.ID
		})
		if err != nil {
			return fmt.Errorf("failed to convert origin tenants to map: %w", err)
		}

		return nil
	})

	if err := gp.Wait(); err != nil {
		return err
	}

	added, deleted, changed := diff.CompareMaps(originTenantMap, latestTenantMap, func(oldTenant, newTenant *types.Tenant) bool {
		// if new tenant is nil means it's deleted;
		// if old tenant is nil means it's a new tenant.
		if newTenant == nil || oldTenant == nil {
			return true
		}

		if newTenant.Name != oldTenant.Name {
			return true
		}

		if newTenant.Enabled != oldTenant.Enabled {
			return true
		}

		return false
	})

	if len(added) > 0 {
		addedTenants := conv.MapValueToSlice(added)
		if err := act.daoTenant.CreateManyTenant(nCtx, addedTenants...); err != nil {
			return fmt.Errorf("failed to create new tenants: %w", err)
		}

		ctx.Data.LogI(fmt.Sprintf("added tenants num: %d", len(added)))
	}

	if len(deleted) > 0 {
		deletedTenantIDs := conv.MapKeyToSlice(deleted)
		if err := act.daoTenant.DeleteManyTenant(nCtx, deletedTenantIDs); err != nil {
			return fmt.Errorf("failed to delete deleted tenants: %w", err)
		}

		ctx.Data.LogI(fmt.Sprintf("deleted tenants num: %d", len(deleted)))
	}

	if len(changed) > 0 {
		if err := act.daoTenant.UpdateManyTenant(nCtx, changed); err != nil {
			return fmt.Errorf("failed to update changed tenants: %w", err)
		}

		ctx.Data.LogI(fmt.Sprintf("changed tenants num: %d", len(changed)))
	}

	return nil
}
