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

// Package tenant provides tenant related operations.
package tenant

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	pkgTenant "github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IStorage defines tenant storage interface.
type IStorage interface {
	basestorage.Interface
	pkgTenant.ITenantIDProvider

	IDaoTenant
	IDomainTenant
}

// IDomainTenant defines tenant domain interface.
type IDomainTenant interface {
	// EnsureReservedTenant ensures the reserved tenant exists in storage.
	EnsureReservedTenant(nCtx contextx.IContext, tenant *types.Tenant) error
}

// IDaoTenant defines tenant dao interface.
type IDaoTenant interface {
	// CreateManyTenant create many tenant.
	CreateManyTenant(nCtx contextx.IContext, tenant ...*types.Tenant) error

	// DeleteManyTenant delete many tenant.
	DeleteManyTenant(nCtx contextx.IContext, tenantID []string) error

	// UpdateManyTenant update many tenant.
	UpdateManyTenant(nCtx contextx.IContext, tenantMap map[string]*types.Tenant) error

	// ListAllEnabledTenants list all enabled tenants.
	ListAllEnabledTenants(nCtx contextx.IContext) ([]*types.Tenant, error)

	// ListAllTenants list all enabled tenants.
	ListAllTenants(nCtx contextx.IContext) ([]*types.Tenant, error)
}
