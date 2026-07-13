/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package tenant

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (s *Storage) listAllEnabledTenants(nCtx contextx.IContext) ([]*types.Tenant, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	tenants, err := s.daoTenant.ListWithoutCount(nCtx, types.UnlimitedPage(), tenant.WithStatus(true))
	if err != nil {
		return nil, fmt.Errorf("failed to list all enabled tenants: %w", err)
	}

	return tenants, nil
}

func (s *Storage) listAllTenants(nCtx contextx.IContext) ([]*types.Tenant, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	tenants, err := s.daoTenant.ListWithoutCount(nCtx, types.UnlimitedPage())
	if err != nil {
		return nil, fmt.Errorf("failed to list all tenants: %w", err)
	}

	return tenants, nil
}

func (s *Storage) createManyTenant(nCtx contextx.IContext, tenants ...*types.Tenant) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if len(tenants) == 0 {
		return base.ErrInvalidParam(fmt.Errorf("tenants is empty"))
	}

	err := s.daoTenant.CreateMany(nCtx, tenants...)
	if err != nil {
		return fmt.Errorf("failed to create many tenants: %w", err)
	}

	return nil
}

func (s *Storage) deleteManyTenant(nCtx contextx.IContext, tenantIDs []string) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if len(tenantIDs) == 0 {
		return base.ErrInvalidParam(fmt.Errorf("tenantIDs is empty"))
	}

	err := s.daoTenant.DeleteMany(nCtx, tenantIDs...)
	if err != nil {
		return fmt.Errorf("failed to delete many tenants: %w", err)
	}

	return nil
}

func (s *Storage) updateManyTenant(nCtx contextx.IContext, tenantMap map[string]*types.Tenant) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if len(tenantMap) == 0 {
		return base.ErrInvalidParam(fmt.Errorf("tenantMap is empty"))
	}

	err := s.daoTenant.UpdateMany(nCtx, tenantMap)
	if err != nil {
		return fmt.Errorf("failed to update many tenants: %w", err)
	}

	return nil
}
