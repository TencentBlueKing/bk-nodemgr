/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package configpolicy

import (
	"errors"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler config policy handler interface.
type IHandler interface {
	// Count count config policy by conditions.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// List lists config policy by page and conditions.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.ConfigPolicy, int64, error)

	// Get gets config policy.
	Get(nCtx contextx.IContext, configPolicyID int64) (*types.ConfigPolicy, error)

	// Create creates config policy.
	Create(nCtx contextx.IContext, configPolicy *types.ConfigPolicy) (int64, error)

	// UpdateMany updates config policies.
	UpdateMany(nCtx contextx.IContext, configPolicies ...*types.ConfigPolicy) error

	// DeleteMany deletes config policies by ids.
	DeleteMany(nCtx contextx.IContext, configPolicyIDs ...int64) error

	// EnableMany enables config policies by ids.
	EnableMany(nCtx contextx.IContext, configPolicyIDs ...int64) error

	// DisableMany disables config policies by ids.
	DisableMany(nCtx contextx.IContext, configPolicyIDs ...int64) error
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
		logger.G.Sys().WithErr(err).With("tenant-id", tenantID).Warn("failed to ensure config policy indexes")
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao) // nolint: forcetypeassert
}

// New create a new accesspoint handler.
func New(client *mongo.Database) IHandler {
	return &handler{
		client: client,
		daoMap: sync.Map{},
	}
}

// Count count config policy by conditions.
func (h *handler) Count(nCtx contextx.IContext, opts ...OptFn) (int64, error) {
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

// List lists config policy by page and conditions.
func (h *handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.ConfigPolicy, int64, error) {
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

	configPolicies, err := h.tenantDao(tenantID).List(nCtx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}

	data := make([]*types.ConfigPolicy, len(configPolicies))
	for idx, configPolicy := range configPolicies {
		data[idx] = convertConfigPolicyToTypes(configPolicy)
	}

	return data, num, nil
}

// Get gets config policy.
func (h *handler) Get(nCtx contextx.IContext, configPolicyID int64) (*types.ConfigPolicy, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	filter := base.AliveFilter()
	filter = WithConfigPolicyID(configPolicyID)(filter)

	data, err := h.tenantDao(tenantID).Get(nCtx, filter)
	if err != nil {
		return nil, err
	}

	return convertConfigPolicyToTypes(data), nil
}

// Create creates config policy.
func (h *handler) Create(nCtx contextx.IContext, configPolicy *types.ConfigPolicy) (int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return -1, err
	}

	tenantID := nCtx.TenantID()

	if configPolicy == nil {
		return -1, base.ErrEmptyParamData()
	}

	if err := base.CheckTenantIDMatched(tenantID, configPolicy.TenantID); err != nil {
		return -1, err
	}

	if configPolicy.Name == "" {
		return -1, errors.New("config policy name is empty")
	}

	data := convertConfigPolicyFromTypes(configPolicy)
	data.Version = 1

	return h.tenantDao(tenantID).create(nCtx, data)
}

// UpdateMany updates config policies.
func (h *handler) UpdateMany(nCtx contextx.IContext, configPolicies ...*types.ConfigPolicy) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	if len(configPolicies) == 0 {
		return base.ErrEmptyParamData()
	}

	data := make([]*ConfigPolicy, len(configPolicies))
	for idx, configPolicy := range configPolicies {
		if configPolicy == nil {
			return base.ErrInvalidItemInParamList()
		}

		if configPolicy.Name == "" {
			return errors.New("config policy name is empty")
		}

		data[idx] = convertConfigPolicyFromTypes(configPolicy)

		if err := base.CheckTenantIDMatched(tenantID, data[idx].Raw.TenantID); err != nil {
			return err
		}
	}

	return h.tenantDao(tenantID).updateMany(nCtx, tenantID, data)
}

// DeleteMany deletes config policies by ids.
func (h *handler) DeleteMany(nCtx contextx.IContext, configPolicyIDs ...int64) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	if len(configPolicyIDs) == 0 {
		return base.ErrEmptyParamData()
	}

	return h.tenantDao(tenantID).deleteMany(nCtx, tenantID, configPolicyIDs...)
}

// EnableMany enables config policies by ids.
func (h *handler) EnableMany(nCtx contextx.IContext, configPolicyIDs ...int64) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	if len(configPolicyIDs) == 0 {
		return base.ErrEmptyParamData()
	}

	return h.tenantDao(tenantID).setEnabledMany(nCtx, tenantID, true, configPolicyIDs...)
}

// DisableMany disables config policies by ids.
func (h *handler) DisableMany(nCtx contextx.IContext, configPolicyIDs ...int64) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	tenantID := nCtx.TenantID()

	if len(configPolicyIDs) == 0 {
		return base.ErrEmptyParamData()
	}

	return h.tenantDao(tenantID).setEnabledMany(nCtx, tenantID, false, configPolicyIDs...)
}

func convertConfigPolicyToTypes(cp *ConfigPolicy) *types.ConfigPolicy {
	scopes := make([]types.ConfigPolicyScope, len(cp.Raw.Scopes))
	for idx, scope := range cp.Raw.Scopes {
		scopes[idx] = types.ConfigPolicyScope{
			NetworkAreaID: scope.NetworkAreaID,
			NetworkUnitID: scope.NetworkUnitID,
			NodeOsType:    criteria.OSType(scope.NodeOsType),
			NodeCPUArch:   criteria.CPUArch(scope.NodeCPUArch),
		}
	}

	return &types.ConfigPolicy{
		TenantID:      cp.Raw.TenantID,
		Version:       cp.Version,
		ID:            cp.Raw.ConfigPolicyID,
		Name:          cp.Raw.ConfigPolicyName,
		Type:          types.ConfigPolicyType(cp.Raw.ConfigPolicyType),
		BizID:         cp.Raw.BizID,
		Remark:        cp.Raw.Remark,
		Scopes:        scopes,
		TargetHostIDs: cp.Raw.TargetHostIDs,
		Configs:       cp.Raw.Configs,
		Enabled:       cp.Raw.Enabled,
		Priority:      cp.Raw.Priority,
		UpdatedAt:     cp.Raw.UpdatedAt,
		Operator:      cp.Raw.Operator,
	}
}

func convertConfigPolicyFromTypes(cp *types.ConfigPolicy) *ConfigPolicy {
	scopes := make([]Scope, len(cp.Scopes))
	for idx, scope := range cp.Scopes {
		scopes[idx] = Scope{
			NetworkAreaID: scope.NetworkAreaID,
			NetworkUnitID: scope.NetworkUnitID,
			NodeOsType:    string(scope.NodeOsType),
			NodeCPUArch:   string(scope.NodeCPUArch),
		}
	}

	return &ConfigPolicy{
		Version: cp.Version,
		Raw: RawData{
			TenantID:         cp.TenantID,
			ConfigPolicyID:   cp.ID,
			ConfigPolicyName: cp.Name,
			ConfigPolicyType: string(cp.Type),
			BizID:            cp.BizID,
			Remark:           cp.Remark,
			Scopes:           scopes,
			TargetHostIDs:    cp.TargetHostIDs,
			Configs:          cp.Configs,
			Enabled:          cp.Enabled,
			Priority:         cp.Priority,
			UpdatedAt:        cp.UpdatedAt,
			Operator:         cp.Operator,
		},
	}
}
