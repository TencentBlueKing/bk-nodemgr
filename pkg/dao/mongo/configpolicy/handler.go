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

package configpolicy

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

// IHandler config policy handler interface.
type IHandler interface {
	// Count count config policy by conditions.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// List lists config policy by page and conditions.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.ConfigPolicy, int64, error)

	// Get gets config policy.
	Get(nCtx contextx.IContext, configPolicyID int64) (*types.ConfigPolicy, error)

	// Create creates config policy and returns its inserted snapshot.
	Create(nCtx contextx.IContext, configPolicy *types.ConfigPolicy) (*types.ConfigPolicy, error)

	// UpdateMany returns post-update snapshots, including successful mutations before an error.
	UpdateMany(nCtx contextx.IContext, fields types.ConfigPolicyFields, configPolicies ...*types.ConfigPolicy) ([]*types.ConfigPolicy, error)

	// DeleteMany soft-deletes policies and returns pre-delete snapshots, including successes before an error.
	DeleteMany(nCtx contextx.IContext, configPolicyIDs ...int64) ([]*types.ConfigPolicy, error)

	// EnableMany returns post-enable snapshots, including successful mutations before an error.
	EnableMany(nCtx contextx.IContext, configPolicyIDs ...int64) ([]*types.ConfigPolicy, error)

	// DisableMany returns post-disable snapshots, including successful mutations before an error.
	DisableMany(nCtx contextx.IContext, configPolicyIDs ...int64) ([]*types.ConfigPolicy, error)

	// UpdatePriorityMany returns post-update snapshots, including successful mutations before an error.
	UpdatePriorityMany(nCtx contextx.IContext, priorities map[int64]int64) ([]*types.ConfigPolicy, error)
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

// Create creates config policy and returns its inserted snapshot.
func (h *handler) Create(nCtx contextx.IContext, configPolicy *types.ConfigPolicy) (*types.ConfigPolicy, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	if configPolicy == nil {
		return nil, base.ErrEmptyParamData()
	}

	if err := base.CheckTenantIDMatched(tenantID, configPolicy.TenantID); err != nil {
		return nil, err
	}

	if configPolicy.Name == "" {
		return nil, errors.New("config policy name is empty")
	}

	data := convertConfigPolicyFromTypes(configPolicy)
	data.Version = 1

	data, err := h.tenantDao(tenantID).create(nCtx, data)
	if err != nil {
		return nil, err
	}

	return convertConfigPolicyToTypes(data), nil
}

// UpdateMany updates selected fields and returns snapshots of successful mutations, even on error.
func (h *handler) UpdateMany(nCtx contextx.IContext, fields types.ConfigPolicyFields, configPolicies ...*types.ConfigPolicy) (
	[]*types.ConfigPolicy, error) {

	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	if len(configPolicies) == 0 {
		return nil, base.ErrEmptyParamData()
	}

	docs := make([]*base.DocumentFieldUpdate, 0, len(configPolicies))
	for _, configPolicy := range configPolicies {
		if configPolicy == nil {
			return nil, base.ErrInvalidItemInParamList()
		}

		if fields.Name && configPolicy.Name == "" {
			return nil, errors.New("config policy name is empty")
		}

		if err := base.CheckTenantIDMatched(tenantID, configPolicy.TenantID); err != nil {
			return nil, err
		}

		filter := append(base.AliveFilter(),
			bson.E{Key: FieldKeyConfigPolicyID, Value: configPolicy.ID},
			bson.E{Key: FieldKeyTenantID, Value: tenantID},
		)

		updateFields := generateConfigPolicyUpdates(fields, configPolicy)
		if len(updateFields) == 0 {
			return nil, base.ErrInvalidParam(
				fmt.Errorf("at least one field must be set to update， policy-id(%d), policy-name(%s)", configPolicy.ID, configPolicy.Name))
		}

		docs = append(docs, &base.DocumentFieldUpdate{
			Filter: filter,
			Fields: updateFields,
		})
	}

	data, err := h.tenantDao(tenantID).updateMany(nCtx, docs)

	return convertConfigPoliciesToTypes(data), err
}

func generateConfigPolicyUpdates(fields types.ConfigPolicyFields, configPolicy *types.ConfigPolicy) map[string]any {
	updates := make(map[string]any)

	if fields.Name {
		updates[FieldKeyConfigPolicyName] = configPolicy.Name
	}
	if fields.Remark {
		updates[FieldKeyRemark] = configPolicy.Remark
	}
	if fields.Scopes {
		updates[FieldKeyScopes] = convertConfigPolicyScopesFromTypes(configPolicy.Scopes)
	}
	if fields.TargetHostIDs {
		updates[FieldKeyTargetHostIDs] = configPolicy.TargetHostIDs
	}
	if fields.TargetPluginName {
		updates[FieldKeyTargetPluginName] = configPolicy.TargetPluginName
	}
	if fields.Configs {
		updates[FieldKeyConfigs] = configPolicy.Configs
	}
	if fields.Enabled {
		updates[FieldKeyEnabled] = configPolicy.Enabled
	}
	if fields.Priority {
		updates[FieldKeyPriority] = configPolicy.Priority
	}
	if fields.Operator {
		updates[FieldKeyOperator] = configPolicy.Operator
	}

	return updates
}

// DeleteMany soft-deletes policies and returns successful pre-delete snapshots, even on error.
func (h *handler) DeleteMany(nCtx contextx.IContext, configPolicyIDs ...int64) ([]*types.ConfigPolicy, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()

	if len(configPolicyIDs) == 0 {
		return nil, base.ErrEmptyParamData()
	}

	data, err := h.tenantDao(tenantID).deleteMany(nCtx, tenantID, configPolicyIDs...)

	return convertConfigPoliciesToTypes(data), err
}

// EnableMany enables config policies by ids and reassigns global priorities.
func (h *handler) EnableMany(nCtx contextx.IContext, configPolicyIDs ...int64) ([]*types.ConfigPolicy, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	if len(configPolicyIDs) == 0 {
		return nil, base.ErrEmptyParamData()
	}

	tenantID := nCtx.TenantID()

	data, err := h.tenantDao(tenantID).enableMany(nCtx, tenantID, configPolicyIDs...)

	return convertConfigPoliciesToTypes(data), err
}

// DisableMany disables config policies by ids and sets their priorities to disabled.
func (h *handler) DisableMany(nCtx contextx.IContext, configPolicyIDs ...int64) ([]*types.ConfigPolicy, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	if len(configPolicyIDs) == 0 {
		return nil, base.ErrEmptyParamData()
	}

	tenantID := nCtx.TenantID()

	data, err := h.tenantDao(tenantID).disableMany(nCtx, tenantID, types.ConfigPolicyPriorityDisabled, configPolicyIDs...)

	return convertConfigPoliciesToTypes(data), err
}

// UpdatePriorityMany batch-updates the priority field for the given policy IDs.
func (h *handler) UpdatePriorityMany(nCtx contextx.IContext, priorities map[int64]int64) ([]*types.ConfigPolicy, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return nil, err
	}

	tenantID := nCtx.TenantID()
	d := h.tenantDao(tenantID)

	data, err := d.updatePriorityMany(nCtx, tenantID, priorities)

	return convertConfigPoliciesToTypes(data), err
}

func convertConfigPoliciesToTypes(data []*ConfigPolicy) []*types.ConfigPolicy {
	policies := make([]*types.ConfigPolicy, len(data))
	for idx, policy := range data {
		policies[idx] = convertConfigPolicyToTypes(policy)
	}

	return policies
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
		TenantID:         cp.Raw.TenantID,
		Version:          cp.Version,
		ID:               cp.Raw.ConfigPolicyID,
		Name:             cp.Raw.ConfigPolicyName,
		Type:             types.ConfigPolicyType(cp.Raw.ConfigPolicyType),
		BizID:            cp.Raw.BizID,
		Remark:           cp.Raw.Remark,
		Scopes:           scopes,
		TargetHostIDs:    cp.Raw.TargetHostIDs,
		TargetPluginName: cp.Raw.TargetPluginName,
		Configs:          cp.Raw.Configs,
		Enabled:          cp.Raw.Enabled,
		Priority:         cp.Raw.Priority,
		UpdatedAt:        cp.Raw.UpdatedAt,
		Operator:         cp.Raw.Operator,
	}
}

func convertConfigPolicyFromTypes(cp *types.ConfigPolicy) *ConfigPolicy {
	scopes := convertConfigPolicyScopesFromTypes(cp.Scopes)

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
			TargetPluginName: cp.TargetPluginName,
			Configs:          cp.Configs,
			Enabled:          cp.Enabled,
			Priority:         cp.Priority,
			UpdatedAt:        cp.UpdatedAt,
			Operator:         cp.Operator,
		},
	}
}

func convertConfigPolicyScopesFromTypes(scopes []types.ConfigPolicyScope) []Scope {
	data := make([]Scope, len(scopes))
	for idx, scope := range scopes {
		data[idx] = Scope{
			NetworkAreaID: scope.NetworkAreaID,
			NetworkUnitID: scope.NetworkUnitID,
			NodeOsType:    string(scope.NodeOsType),
			NodeCPUArch:   string(scope.NodeCPUArch),
		}
	}

	return data
}
