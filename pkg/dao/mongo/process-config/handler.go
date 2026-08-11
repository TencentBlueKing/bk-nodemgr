/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package processconfig

import (
	"fmt"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler defines the process config handler interface.
type IHandler interface {
	// Create create process config record.
	Create(nCtx contextx.IContext, config *types.ProcessConfig) error

	// Get get process config record.
	Get(nCtx contextx.IContext, hostID int64, processName, name string) (*types.ProcessConfig, error)

	// Count count process config records.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// List list process config records.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.ProcessConfig, int64, error)

	// UpsertMany upsert many process config record.
	UpsertMany(nCtx contextx.IContext, configs ...*types.ProcessConfig) error

	// DeleteMany delete many process config record.
	DeleteMany(nCtx contextx.IContext, opts ...OptFn) error
}

var _ IHandler = &Handler{}

// Handler this is a Handler to operate process config table.
type Handler struct {
	client *mongo.Database
	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

func (h *Handler) tenantDao(tenantID string) *dao {
	tableName := TableName(tenantID)

	if d, ok := h.daoMap.Load(tableName); ok {
		return d.(*dao) // nolint: forcetypeassert
	}

	newDaoClient := newDao(h.client, tableName)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).With("table-name", tableName).Warn("failed to ensure process indexes")
	}

	d, _ := h.daoMap.LoadOrStore(tableName, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao) // nolint: forcetypeassert
}

// New create a new process config handler.
func New(client *mongo.Database) *Handler {
	return &Handler{
		client: client,
		daoMap: sync.Map{},
	}
}

// Create create process config record.
func (h *Handler) Create(nCtx contextx.IContext, config *types.ProcessConfig) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	if config == nil {
		return fmt.Errorf("process config is nil")
	}

	tenantID := nCtx.TenantID()

	data := convProcessConfigFromTypes(config)
	if err := h.tenantDao(tenantID).Create(nCtx, data); err != nil {
		return fmt.Errorf("failed to create process, err: %w", err)
	}

	return nil
}

// Get get process config record.
func (h *Handler) Get(nCtx contextx.IContext, hostID int64, processName, name string) (*types.ProcessConfig, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	filter := base.AliveFilter()
	filter = WithName(name)(filter)
	filter = WithProcessName(processName)(filter)
	filter = WithHostID(hostID)(filter)

	data, err := h.tenantDao(tenantID).Get(nCtx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get process config, err: %w", err)
	}

	return convProcessConfigToTypes(data), nil
}

// Count count process config records.
func (h *Handler) Count(nCtx contextx.IContext, opts ...OptFn) (int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return 0, err
	}

	tenantID := nCtx.TenantID()
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	count, err := h.tenantDao(tenantID).Count(nCtx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count process config, err: %w", err)
	}

	return count, nil
}

// List list process config records.
func (h *Handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.ProcessConfig, int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	tenantID := nCtx.TenantID()
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.tenantDao(tenantID).Count(nCtx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	dataList, err := h.tenantDao(tenantID).List(nCtx, filter, findOpt)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list process config, err: %w", err)
	}

	result := make([]*types.ProcessConfig, 0, len(dataList))
	for _, data := range dataList {
		result = append(result, convProcessConfigToTypes(data))
	}

	return result, num, nil
}

// UpsertMany upsert many process config record.
func (h *Handler) UpsertMany(nCtx contextx.IContext, configs ...*types.ProcessConfig) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	if len(configs) == 0 {
		return fmt.Errorf("process config is nil")
	}

	tenantID := nCtx.TenantID()
	dbConfigs := make([]*ProcessConfig, 0, len(configs))
	for _, config := range configs {
		dbConfigs = append(dbConfigs, convProcessConfigFromTypes(config))
	}
	if err := h.tenantDao(tenantID).upsertMany(nCtx, dbConfigs...); err != nil {
		return fmt.Errorf("failed to upsert process config, err: %w", err)
	}

	return nil
}

// DeleteMany delete many process config record.
func (h *Handler) DeleteMany(nCtx contextx.IContext, opts ...OptFn) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()
	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	if err := h.tenantDao(tenantID).HardDeleteMany(nCtx, filter); err != nil {
		return fmt.Errorf("failed to delete process config, err: %w", err)
	}

	return nil
}

// convProcessConfigToTypes convert ProcessConfig to types.ProcessConfig.
func convProcessConfigToTypes(config *ProcessConfig) *types.ProcessConfig {
	if config == nil {
		return nil
	}

	return &types.ProcessConfig{
		Name:                config.Name,
		ProcessName:         config.ProcessName,
		HostID:              config.HostID,
		IsMainConfig:        config.IsMainConfig,
		Content:             config.Content,
		MD5:                 config.MD5,
		FilePath:            config.FilePath,
		CustomConfigContext: config.CustomConfigContext,
	}
}

// convProcessConfigFromTypes convert types.ProcessConfig to ProcessConfig.
func convProcessConfigFromTypes(config *types.ProcessConfig) *ProcessConfig {
	if config == nil {
		return nil
	}

	return &ProcessConfig{
		Name:                config.Name,
		ProcessName:         config.ProcessName,
		HostID:              config.HostID,
		IsMainConfig:        config.IsMainConfig,
		Content:             config.Content,
		MD5:                 config.MD5,
		FilePath:            config.FilePath,
		CustomConfigContext: config.CustomConfigContext,
	}
}
