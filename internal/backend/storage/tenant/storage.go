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
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// StorageName defines the storage name.
const StorageName = "tenant"

const (
	metricOperationListAllEnabledTenants = "list_all_enabled_tenants"
	metricOperationListAllTenants        = "list_all_tenants"
	metricOperationCreateManyTenant      = "create_many_tenant"
	metricOperationDeleteManyTenant      = "delete_many_tenant"
	metricOperationUpdateManyTenant      = "update_many_tenant"
)

// NewStorage creates a new release storage.
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

	daoTenant tenant.IHandler
}

func (s *Storage) initDao() error {
	s.daoTenant = tenant.New(s.Database)

	return nil
}

func (s *Storage) check() error {
	if s.daoTenant == nil {
		return errors.New("dao release is nil")
	}

	return nil
}

// ListAllEnabledTenants list all enable tenants.
func (s *Storage) ListAllEnabledTenants(nCtx contextx.IContext) ([]*types.Tenant, error) {
	var tenants []*types.Tenant

	err := s.WrapFn(nCtx, metricOperationListAllEnabledTenants, func(ctx contextx.IContext) error {
		var err error
		tenants, err = s.listAllEnabledTenants(ctx)

		return err
	})

	return tenants, err
}

// ListAllTenants list all enable tenants.
func (s *Storage) ListAllTenants(nCtx contextx.IContext) ([]*types.Tenant, error) {
	var tenants []*types.Tenant

	err := s.WrapFn(nCtx, metricOperationListAllTenants, func(ctx contextx.IContext) error {
		var err error
		tenants, err = s.listAllTenants(ctx)

		return err
	})

	return tenants, err
}

// CreateManyTenant create many tenant.
func (s *Storage) CreateManyTenant(nCtx contextx.IContext, tenants ...*types.Tenant) error {
	return s.WrapFn(nCtx, metricOperationCreateManyTenant, func(ctx contextx.IContext) error {
		var err error
		err = s.createManyTenant(ctx, tenants...)

		return err
	})
}

// DeleteManyTenant delete many tenant.
func (s *Storage) DeleteManyTenant(nCtx contextx.IContext, tenantIDs []string) error {
	return s.WrapFn(nCtx, metricOperationDeleteManyTenant, func(ctx contextx.IContext) error {
		var err error
		err = s.deleteManyTenant(ctx, tenantIDs)

		return err
	})
}

// UpdateManyTenant update many tenant.
func (s *Storage) UpdateManyTenant(nCtx contextx.IContext, tenantMap map[string]*types.Tenant) error {
	return s.WrapFn(nCtx, metricOperationUpdateManyTenant, func(ctx contextx.IContext) error {
		var err error
		err = s.updateManyTenant(ctx, tenantMap)

		return err
	})
}
