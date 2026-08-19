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

package networkunit

import (
	"errors"
	"fmt"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler networkunit handler interface.
type IHandler interface {
	// Count counts networkunit by conditions.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// List lists networkunit by page and conditions.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.NetworkUnit, int64, error)

	// Get gets networkunit by id.
	Get(nCtx contextx.IContext, networkUnitID int64) (*types.NetworkUnit, error)

	// Create creates a networkunit.
	Create(nCtx contextx.IContext, networkUnit *types.NetworkUnit) (int64, error)

	// UpdateMany updates networkunit.
	UpdateMany(nCtx contextx.IContext, fields types.NetworkUnitUpdateFields, networkUnits ...*types.NetworkUnit) error

	// DeleteMany deletes networkunit.
	DeleteMany(nCtx contextx.IContext, networkUnitIDs ...int64) error

	// GetNetworkUnitDistributionByNetworkAreaID gets networkunit distribution by network area id.
	GetNetworkUnitDistributionByNetworkAreaID(nCtx contextx.IContext, opts ...OptFn) (map[int64]int64, error)
}

type handler struct {
	client *mongo.Database
	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

func (h *handler) tenantDao(tenantID string) *dao {
	if d, ok := h.daoMap.Load(tenantID); ok {
		return d.(*dao) // nolint: forcetypeassert
	}

	newDaoClient := newDao(tenantID, h.client)
	if err := newDaoClient.EnsureIndexes(); err != nil {
		logger.G.Sys().WithErr(err).With("tenant-id", tenantID).Warn("failed to ensure networkunit indexes")
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	// tenantDao is the sole writer of daoMap, so stored values are always *dao.
	return d.(*dao) // nolint: forcetypeassert
}

// New creates a new networkunit handler.
func New(client *mongo.Database) IHandler {
	return &handler{
		client: client,
		daoMap: sync.Map{},
	}
}

// Count counts networkunit by conditions.
func (h *handler) Count(nCtx contextx.IContext, opts ...OptFn) (int64, error) {
	if nCtx == nil {
		return 0, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return 0, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	return h.tenantDao(tenantID).Count(nCtx, filter)
}

// List lists networkunit by page and conditions.
func (h *handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) (
	[]*types.NetworkUnit, int64, error) {

	if nCtx == nil {
		return nil, 0, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, 0, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	tenantDao := h.tenantDao(tenantID)
	num, err := tenantDao.Count(nCtx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := base.ParsePage(page)

	networkUnits, err := tenantDao.List(nCtx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	data := make([]*types.NetworkUnit, len(networkUnits))
	for idx, networkunit := range networkUnits {
		data[idx] = convertNetworkUnitToTypes(networkunit)
	}

	return data, num, nil
}

// Get gets networkunit by id.
func (h *handler) Get(nCtx contextx.IContext, networkUnitID int64) (*types.NetworkUnit, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	if networkUnitID < 0 {
		return nil, base.ErrInvalidID()
	}

	filter := base.AliveFilter()
	opt := base.WithInt64Values(FieldKeyNetworkUnitID, networkUnitID)
	filter = opt(filter)

	data, err := h.tenantDao(tenantID).Get(nCtx, filter)
	if err != nil {
		return nil, err
	}

	return convertNetworkUnitToTypes(data), nil
}

// Create creates a new networkunit and return the generated id.
func (h *handler) Create(nCtx contextx.IContext, networkUnit *types.NetworkUnit) (int64, error) {
	if nCtx == nil {
		return -1, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return -1, err
	}

	tenantID := nCtx.TenantID()

	if networkUnit == nil {
		return -1, base.ErrEmptyParamData()
	}

	if err := base.CheckTenantIDMatched(tenantID, networkUnit.TenantID); err != nil {
		return -1, fmt.Errorf("failed to create networkunit: %w", err)
	}

	if networkUnit.Name == "" {
		return -1, errors.New("networkunit name is empty")
	}

	if networkUnit.NetworkAreaID < 0 {
		return -1, errors.New("networkunit networkarea-id is invalid")
	}

	return h.tenantDao(tenantID).create(nCtx, convertNetworkUnitFromTypes(networkUnit))
}

// UpdateMany updates networkunit.
func (h *handler) UpdateMany(nCtx contextx.IContext, fields types.NetworkUnitUpdateFields, networkUnits ...*types.NetworkUnit) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	if len(networkUnits) == 0 {
		return base.ErrEmptyParamData()
	}

	if !fields.Name && !fields.AccessPoints && !fields.Links && !fields.DirectEndpoints && !fields.CustomDeployConfig {
		return errors.New("no fields to update")
	}

	docs := make([]*base.DocumentFieldUpdate, 0, len(networkUnits))
	for _, networkUnit := range networkUnits {
		if networkUnit == nil {
			return base.ErrInvalidItemInParamList()
		}

		if err := base.CheckTenantIDMatched(tenantID, networkUnit.TenantID); err != nil {
			return fmt.Errorf("failed to update networkunit(%d): %w", networkUnit.ID, err)
		}

		updates := generateNetworkUnitUpdates(fields, networkUnit)
		if len(updates) == 0 {
			continue
		}

		docs = append(docs, &base.DocumentFieldUpdate{
			Filter: func() bson.D {
				filter := base.AliveFilter()
				filter = WithNetworkUnitID(networkUnit.ID)(filter)

				return filter
			}(),
			Fields: updates,
		})
	}

	if err := h.tenantDao(tenantID).UpdateOneFieldBulk(nCtx, docs); err != nil {
		return err
	}

	return nil
}

// DeleteMany deletes networkunit by ids.
func (h *handler) DeleteMany(nCtx contextx.IContext, networkUnitIDs ...int64) error {
	if nCtx == nil {
		return base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	if len(networkUnitIDs) == 0 {
		return base.ErrEmptyParamData()
	}

	if err := h.tenantDao(tenantID).deleteMany(nCtx, networkUnitIDs...); err != nil {
		return err
	}

	return nil
}

// GetNetworkUnitDistributionByNetworkAreaID gets networkunit distribution by network area id.
func (h *handler) GetNetworkUnitDistributionByNetworkAreaID(nCtx contextx.IContext, opts ...OptFn) (map[int64]int64, error) {
	if nCtx == nil {
		return nil, base.ErrInvalidContext()
	}

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, fmt.Errorf("failed to get networkunit distribution by network area id: %w", err)
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	results, err := h.tenantDao(tenantID).getNetworkUnitDistributionByNetworkAreaID(nCtx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get networkunit distribution by network area id: %w", err)
	}

	networkUnitDistribution := make(map[int64]int64)
	for _, result := range results {
		networkUnitDistribution[result.NetworkAreaID] = result.NetworkUnitCount
	}

	return networkUnitDistribution, nil
}

func convertNetworkUnitFromTypes(networkUnit *types.NetworkUnit) *NetworkUnit {
	return &NetworkUnit{
		TenantID:        networkUnit.TenantID,
		NetworkUnitID:   networkUnit.ID,
		NetworkUnitName: networkUnit.Name,
		NetworkAreaID:   networkUnit.NetworkAreaID,
		AccessPoints:    networkUnit.AccessPoints,
		Links: &Links{
			Cluster: convertLinksFromTypes(networkUnit.Links.Cluster),
			File:    convertLinksFromTypes(networkUnit.Links.File),
			Data:    convertLinksFromTypes(networkUnit.Links.Data),
		},
		IsDirect:           networkUnit.IsDirect,
		DirectEndpoints:    convertEndpointsFromTypes(networkUnit.DirectEndpoints),
		Generation:         int64(networkUnit.Generation),
		CustomDeployConfig: convertDeployConfigFromTypes(networkUnit.CustomDeployConfig),
	}
}

func convertNetworkUnitToTypes(networkArea *NetworkUnit) *types.NetworkUnit {
	var links types.Links
	if networkArea.Links != nil {
		links = types.Links{
			Cluster: convertLinksToTypes(networkArea.Links.Cluster),
			File:    convertLinksToTypes(networkArea.Links.File),
			Data:    convertLinksToTypes(networkArea.Links.Data),
		}
	}

	return &types.NetworkUnit{
		TenantID:           networkArea.TenantID,
		ID:                 networkArea.NetworkUnitID,
		Name:               networkArea.NetworkUnitName,
		NetworkAreaID:      networkArea.NetworkAreaID,
		AccessPoints:       networkArea.AccessPoints,
		Links:              links,
		IsDirect:           networkArea.IsDirect,
		DirectEndpoints:    convertEndpointsToTypes(networkArea.DirectEndpoints),
		Generation:         types.Generation(networkArea.Generation),
		CustomDeployConfig: convertDeployConfigToTypes(networkArea.CustomDeployConfig),
	}
}

func convertLinksFromTypes(link *types.Link) *Link {
	if link == nil {
		return nil
	}

	return &Link{
		NetworkAreaID: link.NetworkAreaID,
		NetworkUnitID: link.NetworkUnitID,
		AccessPointID: link.AccessPointID,
	}
}

func convertLinksToTypes(link *Link) *types.Link {
	if link == nil {
		return nil
	}

	return &types.Link{
		NetworkAreaID: link.NetworkAreaID,
		NetworkUnitID: link.NetworkUnitID,
		AccessPointID: link.AccessPointID,
	}
}

func convertEndpointsFromTypes(endpoint *types.Endpoints) *Endpoints {
	if endpoint == nil {
		return nil
	}

	return &Endpoints{
		Cluster: endpoint.Cluster,
		File:    endpoint.File,
		Data:    endpoint.Data,
	}
}

func convertEndpointsToTypes(endpoint *Endpoints) *types.Endpoints {
	if endpoint == nil {
		return nil
	}

	return &types.Endpoints{
		Cluster: endpoint.Cluster,
		File:    endpoint.File,
		Data:    endpoint.Data,
	}
}

func convertDeployConfigFromTypes(deployConfig map[criteria.OSType]types.CustomDeployConfig) map[string]CustomDeployConfig {
	if deployConfig == nil {
		return nil
	}

	data := make(map[string]CustomDeployConfig, len(deployConfig))
	for osType, config := range deployConfig {
		data[osType.String()] = CustomDeployConfig{
			InstallerRuntime: InstallerRuntime{
				BaseWorkDir: config.InstallerRuntime.BaseWorkDir,
			},
			NodeRuntime: NodeRuntime{
				BaseDeployDir: config.NodeRuntime.BaseDeployDir,
				DataIPC:       config.NodeRuntime.DataIPC,
				PluginIPC:     config.NodeRuntime.PluginIPC,
				LogDir:        config.NodeRuntime.LogDir,
				ZoneID:        config.NodeRuntime.ZoneID,
				CityID:        config.NodeRuntime.CityID,
			},
			PluginRuntime: PluginRuntime{
				BaseDeployDir: config.PluginRuntime.BaseDeployDir,
				LogDir:        config.PluginRuntime.LogDir,
			},
		}
	}

	return data
}

func convertDeployConfigToTypes(deployConfig map[string]CustomDeployConfig) map[criteria.OSType]types.CustomDeployConfig {
	if deployConfig == nil {
		return nil
	}

	data := make(map[criteria.OSType]types.CustomDeployConfig, len(deployConfig))
	for osType, config := range deployConfig {
		data[criteria.OSType(osType)] = types.CustomDeployConfig{
			InstallerRuntime: types.InstallerRuntime{
				BaseWorkDir: config.InstallerRuntime.BaseWorkDir,
			},
			NodeRuntime: types.NodeRuntime{
				BaseDeployDir: config.NodeRuntime.BaseDeployDir,
				DataIPC:       config.NodeRuntime.DataIPC,
				PluginIPC:     config.NodeRuntime.PluginIPC,
				LogDir:        config.NodeRuntime.LogDir,
				ZoneID:        config.NodeRuntime.ZoneID,
				CityID:        config.NodeRuntime.CityID,
			},
			PluginRuntime: types.PluginRuntime{
				BaseDeployDir: config.PluginRuntime.BaseDeployDir,
				LogDir:        config.PluginRuntime.LogDir,
			},
		}
	}

	return data
}

func generateNetworkUnitUpdates(fields types.NetworkUnitUpdateFields, networkUnit *types.NetworkUnit) map[string]any {
	updates := make(map[string]any)
	if fields.Name {
		updates[FieldKeyNetworkUnitName] = networkUnit.Name
	}

	if fields.AccessPoints {
		updates[FieldKeyAccessPoints] = networkUnit.AccessPoints
	}

	if fields.Links {
		updates[FieldKeyLinks] = &Links{
			Cluster: convertLinksFromTypes(networkUnit.Links.Cluster),
			File:    convertLinksFromTypes(networkUnit.Links.File),
			Data:    convertLinksFromTypes(networkUnit.Links.Data),
		}
	}

	if fields.DirectEndpoints {
		updates[FieldKeyDirectEndpoints] = convertEndpointsFromTypes(networkUnit.DirectEndpoints)
	}

	if fields.CustomDeployConfig {
		updates[FieldKeyCustomDeployConfig] = convertDeployConfigFromTypes(networkUnit.CustomDeployConfig)
	}

	return updates
}
