/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package plugin

import (
	"fmt"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler plugin handler interface.
type IHandler interface {
	// Create create a new plugin.
	Create(nCtx contextx.IContext, plugin *types.Plugin) error

	// Count count plugins by conditions.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// List lists plugins by page and conditions.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.Plugin, int64, error)

	// Get gets a plugin by conditions.
	Get(nCtx contextx.IContext, opts ...OptFn) (*types.Plugin, error)

	// Exist check a plugin exist by conditions.
	Exist(nCtx contextx.IContext, opts ...OptFn) (bool, error)
}

var _ IHandler = &Handler{}

// Handler this is a Handler to operate plugin table.
type Handler struct {
	client *mongo.Database
	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

// Exist check a plugin exist by conditions.
func (h *Handler) Exist(nCtx contextx.IContext, opts ...OptFn) (bool, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return false, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	exist, err := h.tenantDao(nCtx.TenantID()).Exist(nCtx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check plugin exist: %w", err)
	}

	return exist, nil
}

func (h *Handler) tenantDao(tenantID string) *dao {
	tableName := TableName(tenantID)

	if d, ok := h.daoMap.Load(tableName); ok {
		return d.(*dao) // nolint: forcetypeassert
	}

	newDaoClient := newDao(h.client, tableName)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).With("table-name", tableName).Warn("failed to ensure plugin indexes")
	}

	d, _ := h.daoMap.LoadOrStore(tableName, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao) // nolint: forcetypeassert
}

// New create a new plugin handler.
func New(client *mongo.Database) *Handler {
	return &Handler{
		client: client,
		daoMap: sync.Map{},
	}
}

// Create create a new plugin.
func (h *Handler) Create(nCtx contextx.IContext, plugin *types.Plugin) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	data := convPluginFromTypes(plugin)

	if err := h.tenantDao(nCtx.TenantID()).Create(nCtx, data); err != nil {
		return fmt.Errorf("failed to create plugin, plugin-name(%s): %w", plugin.Name, err)
	}

	return nil
}

func convPluginFromTypes(plugin *types.Plugin) *Plugin {
	data := &Plugin{
		TenantID: plugin.TenantID,
		Name:     plugin.Name,
		Group:    plugin.Group,
		PkgName:  plugin.PkgName,
		Memo:     plugin.Memo,
	}

	return data
}

// Count count host by conditions.
func (h *Handler) Count(nCtx contextx.IContext, opts ...OptFn) (int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return 0, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.tenantDao(nCtx.TenantID()).Count(nCtx, filter)
}

// List list plugin by page and conditions.
func (h *Handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.Plugin, int64, error) {
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

	data, err := h.tenantDao(nCtx.TenantID()).List(nCtx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	plugins := make([]*types.Plugin, len(data))
	for idx, host := range data {
		plugins[idx] = convertPluginToTypes(host)
	}

	return plugins, num, nil
}

// Get get a plugin.
func (h *Handler) Get(nCtx contextx.IContext, opts ...OptFn) (*types.Plugin, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	data, err := h.tenantDao(nCtx.TenantID()).Get(nCtx, filter)
	if err != nil {
		return nil, err
	}

	return convertPluginToTypes(data), nil
}

func convertPluginToTypes(data *Plugin) *types.Plugin {
	plugin := &types.Plugin{
		TenantID: data.TenantID,
		Name:     data.Name,
		Group:    data.Group,
		PkgName:  data.PkgName,
		Memo:     data.Memo,
	}

	return plugin
}
