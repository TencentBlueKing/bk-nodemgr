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

// Package tenant provides tenant storage for application service.
package tenant

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	tenantDao "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	pkgTenant "github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// StorageName defines the storage name.
	StorageName = "tenant"

	metricOperationListEnabledTenantIDs = "list_enabled_tenant_ids"
	metricOperationEnsureReservedTenant = "ensure_reserved_tenant"
)

// IStorage defines application tenant storage interface.
type IStorage interface {
	basestorage.Interface
	pkgTenant.ITenantIDProvider

	// EnsureReservedTenant ensures the reserved tenant exists in storage.
	EnsureReservedTenant(nCtx contextx.IContext, tenant *types.Tenant) error
}

// NewStorage creates a new tenant storage.
func NewStorage(client *mongo.Client, database string) (*Storage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &Storage{
		Storage: basestorage.Storage{
			Name:     StorageName,
			Database: client.Database(database),
		},
	}
	err := basestorage.InitStorage(&s.Storage,
		basestorage.WithStartFunc(s.initDao),
		basestorage.WithCheckFunc(s.check))
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to new storage")

		return nil, err
	}

	return s, nil
}

var _ IStorage = &Storage{}

// Storage implements IStorage.
type Storage struct {
	basestorage.Storage

	daoTenant tenantDao.IHandler
}

func (s *Storage) initDao() error {
	s.daoTenant = tenantDao.New(s.Database)

	return nil
}

func (s *Storage) check() error {
	if s.daoTenant == nil {
		return errors.New("dao tenant is nil")
	}

	return nil
}

// ListEnabledTenantIDs lists all enabled tenant IDs.
func (s *Storage) ListEnabledTenantIDs(nCtx contextx.IContext) ([]string, error) {
	var tenantIDs []string

	err := s.WrapFn(nCtx, metricOperationListEnabledTenantIDs, func(ctx contextx.IContext) error {
		var err error
		tenantIDs, err = s.listEnabledTenantIDs(ctx)

		return err
	})

	return tenantIDs, err
}

// EnsureReservedTenant ensures the reserved tenant exists in storage.
func (s *Storage) EnsureReservedTenant(nCtx contextx.IContext, tenant *types.Tenant) error {
	return s.WrapFn(nCtx, metricOperationEnsureReservedTenant, func(ctx contextx.IContext) error {
		return s.ensureReservedTenant(ctx, tenant)
	})
}

func (s *Storage) listEnabledTenantIDs(nCtx contextx.IContext) ([]string, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	tenants, _, err := s.daoTenant.List(nCtx, types.UnlimitedPage(), tenantDao.WithStatus(true))
	if err != nil {
		return nil, fmt.Errorf("failed to list all enabled tenants: %w", err)
	}

	tenantIDs := make([]string, 0, len(tenants))
	for _, tenant := range tenants {
		if tenant == nil {
			continue
		}

		tenantIDs = append(tenantIDs, tenant.ID)
	}

	return tenantIDs, nil
}

func (s *Storage) ensureReservedTenant(nCtx contextx.IContext, reservedTenant *types.Tenant) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if reservedTenant == nil {
		return base.ErrInvalidParam(fmt.Errorf("tenant is nil"))
	}

	if reservedTenant.ID == "" {
		return base.ErrInvalidParam(fmt.Errorf("tenant id is empty"))
	}

	tenants, _, err := s.daoTenant.List(nCtx, types.UnlimitedPage())
	if err != nil {
		return fmt.Errorf("failed to list all tenants: %w", err)
	}

	for _, tenant := range tenants {
		if tenant == nil || tenant.ID != reservedTenant.ID {
			continue
		}

		if tenant.Name == reservedTenant.Name && tenant.Enabled == reservedTenant.Enabled {
			return nil
		}

		if err := s.daoTenant.UpdateMany(nCtx, map[string]*types.Tenant{reservedTenant.ID: reservedTenant}); err != nil {
			return fmt.Errorf("failed to update reserved tenant: %w", err)
		}

		return nil
	}

	if err := s.daoTenant.CreateMany(nCtx, reservedTenant); err != nil {
		return fmt.Errorf("failed to create reserved tenant: %w", err)
	}

	return nil
}
