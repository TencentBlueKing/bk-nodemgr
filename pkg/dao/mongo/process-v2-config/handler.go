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

package processv2config

import (
	"errors"
	"fmt"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler defines the v2 process config handler interface.
type IHandler interface {
	// Get gets a v2 process config record by unique key.
	Get(nCtx contextx.IContext, key *types.ProcessConfigUniqueKey) (*types.ProcessConfig, error)

	// Count count v2 process config records.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// List lists v2 process config records by page and conditions.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.ProcessConfig, int64, error)

	// UpsertMany upsert many v2 process config records.
	UpsertMany(nCtx contextx.IContext, configs ...*types.ProcessConfig) error

	// DeleteMany deletes v2 process config records by conditions.
	DeleteMany(nCtx contextx.IContext, opts ...OptFn) error
}

var _ IHandler = &Handler{}

// Handler this is a Handler to operate the process v2 config table.
type Handler struct {
	client *mongo.Database
	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

// New create a new v2 process config handler.
func New(client *mongo.Database) *Handler {
	return &Handler{
		client: client,
		daoMap: sync.Map{},
	}
}

func (h *Handler) tenantDao(tenantID string) *dao {
	tableName := TableName(tenantID)

	if d, ok := h.daoMap.Load(tableName); ok {
		return d.(*dao) // nolint: forcetypeassert
	}

	newDaoClient := newDao(h.client, tableName)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).With("table-name", tableName).Warn("failed to ensure process v2 config indexes")
	}

	d, _ := h.daoMap.LoadOrStore(tableName, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao) // nolint: forcetypeassert
}

// Get gets a v2 process config record by unique key.
func (h *Handler) Get(nCtx contextx.IContext, key *types.ProcessConfigUniqueKey) (*types.ProcessConfig, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	if key == nil {
		return nil, base.ErrInvalidParam(errors.New("process v2 config unique key is nil"))
	}

	filter := WithProcessConfigUniqueKeys(key)(base.AliveFilter())

	data, err := h.tenantDao(nCtx.TenantID()).Get(nCtx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get process v2 config, err: %w", err)
	}

	return convProcessConfigToTypes(data), nil
}

// Count count v2 process config records.
func (h *Handler) Count(nCtx contextx.IContext, opts ...OptFn) (int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	count, err := h.tenantDao(nCtx.TenantID()).Count(nCtx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count process v2 config, err: %w", err)
	}

	return count, nil
}

// List lists v2 process config records by page and conditions.
func (h *Handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.ProcessConfig, int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	num, err := h.tenantDao(nCtx.TenantID()).Count(nCtx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	dataList, err := h.tenantDao(nCtx.TenantID()).List(nCtx, filter, findOpt)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list process v2 config, err: %w", err)
	}

	result := make([]*types.ProcessConfig, 0, len(dataList))
	for _, data := range dataList {
		result = append(result, convProcessConfigToTypes(data))
	}

	return result, num, nil
}

// UpsertMany upsert many v2 process config records.
func (h *Handler) UpsertMany(nCtx contextx.IContext, configs ...*types.ProcessConfig) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	if len(configs) == 0 {
		return fmt.Errorf("process v2 config is nil")
	}

	dbConfigs := make([]*ProcessConfig, 0, len(configs))
	for _, config := range configs {
		dbConfigs = append(dbConfigs, convProcessConfigFromTypes(config))
	}

	if err := h.tenantDao(nCtx.TenantID()).upsertMany(nCtx, dbConfigs...); err != nil {
		return fmt.Errorf("failed to upsert process v2 config, err: %w", err)
	}

	return nil
}

// DeleteMany deletes v2 process config records by conditions.
func (h *Handler) DeleteMany(nCtx contextx.IContext, opts ...OptFn) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	if err := h.tenantDao(nCtx.TenantID()).HardDeleteMany(nCtx, filter); err != nil {
		return fmt.Errorf("failed to delete process v2 config, err: %w", err)
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
		TemplateName:        config.TemplateName,
		ProcessName:         config.ProcessName,
		HostID:              config.HostID,
		Set:                 config.Set,
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
		TemplateName:        config.TemplateName,
		ProcessName:         config.ProcessName,
		HostID:              config.HostID,
		Set:                 config.Set,
		IsMainConfig:        config.IsMainConfig,
		Content:             config.Content,
		MD5:                 config.MD5,
		FilePath:            config.FilePath,
		CustomConfigContext: config.CustomConfigContext,
	}
}
