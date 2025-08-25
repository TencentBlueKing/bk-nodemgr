/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package cptemplate provides the config policy template data models.
package cptemplate

import (
	"context"
	"errors"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler config policy template handler interface.
type IHandler interface {
	// Get get config policy template by config policy id
	Get(ctx context.Context, configPolicyID int64) (*types.ConfigPolicyTemplate, error)

	// UpsertMany upsert many config policy template.
	UpsertMany(ctx context.Context, configPolicyTemplates ...*types.ConfigPolicyTemplate) error

	// DeleteMany delete many config policy template by config policy ids.
	DeleteMany(ctx context.Context, configPolicyIDs ...int64) error
}

type handler struct {
	client *mongo.Database
	logger logger.ILogger
	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

func (h *handler) tenantDao(tenantID string) *dao {
	if d, ok := h.daoMap.Load(tenantID); ok {
		return d.(*dao) // nolint: forcetypeassert
	}

	newDaoClient := newDao(tenantID, h.client, h.logger)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		h.logger.Warnf("failed to ensure config policy indexes. tenant-id(%s): %v",
			tenantID, errors.Join(base.ErrEnsureIndexesFailed(), err))
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao) // nolint: forcetypeassert
}

// New create a new accesspoint handler.
func New(client *mongo.Database, logger logger.ILogger) IHandler {
	return &handler{
		client: client,
		logger: logger,
		daoMap: sync.Map{},
	}
}

// Get gets config policy template.
func (h *handler) Get(ctx context.Context, configPolicyID int64) (*types.ConfigPolicyTemplate, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	filter := base.AliveFilter()
	filter = WithConfigPolicyID(configPolicyID)(filter)

	data, err := h.tenantDao(tenantID).Get(ctx, filter)
	if err != nil {
		return nil, err
	}

	return convertConfigPolicyTemplateToTypes(data), nil
}

// UpsertMany upsert many config policy template.
func (h *handler) UpsertMany(ctx context.Context, configPolicyTemplates ...*types.ConfigPolicyTemplate) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	data := make([]*ConfigPolicyTemplate, len(configPolicyTemplates))
	for idx, configPolicyTemplate := range configPolicyTemplates {
		if configPolicyTemplate == nil {
			return errors.New("config policy templates has nil item")
		}

		data[idx] = convertConfigPolicyTemplateFromTypes(configPolicyTemplate)
	}

	return h.tenantDao(tenantID).upsertMany(ctx, data)
}

// DeleteMany delete many config policy template.
func (h *handler) DeleteMany(ctx context.Context, configPolicyIDs ...int64) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	if len(configPolicyIDs) == 0 {
		return base.ErrEmptyParamData()
	}

	filter := base.AliveFilter()
	filter = WithConfigPolicyID(configPolicyIDs...)(filter)
	if err := h.tenantDao(tenantID).DeleteMany(ctx, filter); err != nil {
		return err
	}

	return nil
}

func convertConfigPolicyTemplateToTypes(cp *ConfigPolicyTemplate) *types.ConfigPolicyTemplate {
	return &types.ConfigPolicyTemplate{
		TenantID: cp.TenantID,
		ID:       cp.ConfigPolicyID,
		Template: cp.Template,
	}
}

func convertConfigPolicyTemplateFromTypes(cp *types.ConfigPolicyTemplate) *ConfigPolicyTemplate {
	return &ConfigPolicyTemplate{
		TenantID:       cp.TenantID,
		ConfigPolicyID: cp.ID,
		Template:       cp.Template,
	}
}
