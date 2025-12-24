/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package deploypolicy

import (
	"fmt"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler deploy policy handler interface.
type IHandler interface {
	// Create create a new deploy policy.
	Create(nCtx contextx.IContext, deployPolicy *types.DeployPolicy) (int64, error)

	// Count count deploy policy by conditions.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// List lists deploy policy by page and conditions.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.DeployPolicy, int64, error)

	// Get gets a deploy policy by conditions.
	Get(nCtx contextx.IContext, opts ...OptFn) (*types.DeployPolicy, error)

	// UpdateFields update a deploy policy by fields.
	UpdateFields(nCtx contextx.IContext, fields types.DeployPolicyFields, deployPolicy ...*types.DeployPolicy) error

	// Delete delete a deploy policy by conditions.
	Delete(nCtx contextx.IContext, deployPolicyID int64) error

	// Exist check a deploy policy exist by conditions.
	Exist(nCtx contextx.IContext, opts ...OptFn) (bool, error)
}

var _ IHandler = &Handler{}

// Handler this is a Handler to operate deploy policy table.
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
		logger.G.Sys().WithErr(err).With("table-name", tableName).Warn("failed to ensure deploy policy indexes")
	}

	d, _ := h.daoMap.LoadOrStore(tableName, newDaoClient)

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao) // nolint: forcetypeassert
}

// New create a new deploy policy handler.
func New(client *mongo.Database) *Handler {
	return &Handler{
		client: client,
		daoMap: sync.Map{},
	}
}

// Create create a new deploy policy.
func (h *Handler) Create(nCtx contextx.IContext, deployPolicy *types.DeployPolicy) (int64, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return -1, err
	}

	data := convDeployPolicyFromTypes(deployPolicy, nCtx.TenantID())

	// generate deploy policy id.
	deployPolicyID, err := h.tenantDao(nCtx.TenantID()).counter.Generate(nCtx, tableNamePrefix)
	if err != nil {
		return -1, fmt.Errorf("failed to generate deploy policy id, err: %w", err)
	}
	data.DeployPolicyID = deployPolicyID
	data.Operator = nCtx.BKUsername()

	if err := h.tenantDao(nCtx.TenantID()).Create(nCtx, data); err != nil {
		return -1, fmt.Errorf("failed to create deploy policy, err: %w", err)
	}

	return deployPolicyID, nil
}

func convDeployPolicyFromTypes(deployPolicy *types.DeployPolicy, tenantID string) *DeployPolicy {
	if deployPolicy == nil {
		return nil
	}

	data := &DeployPolicy{
		TenantID:       tenantID,
		DeployPolicyID: deployPolicy.DeployPolicyID,
		Meta:           convDeployPolicyMetaFromTypes(deployPolicy.Meta),
		Specs:          convSpecsFromTypes(deployPolicy.Specs),
		Scopes:         convScopesFromTypes(deployPolicy.Scopes),
		Operator:       deployPolicy.Operator,
		Enabled:        deployPolicy.Enabled,
	}

	return data
}

func convDeployPolicyMetaFromTypes(deployPolicyMeta types.DeployPolicyMeta) Meta {
	return Meta{
		Name:        deployPolicyMeta.Name,
		Description: deployPolicyMeta.Description,
	}
}

func convSpecsFromTypes(specs []*types.DeploySpec) []*Spec {
	return conv.SliceToSlice[*types.DeploySpec, *Spec](specs, convSpecFromTypes)
}

func convSpecFromTypes(spec *types.DeploySpec) *Spec {
	if spec == nil {
		return nil
	}

	return &Spec{
		Type:  string(spec.Type),
		Param: spec.Param,
	}
}

func convScopesFromTypes(scopes []*types.Scope) []*Scope {
	return conv.SliceToSlice[*types.Scope, *Scope](scopes, convScopeFromTypes)
}

func convScopeFromTypes(scope *types.Scope) *Scope {
	if scope == nil {
		return nil
	}

	return &Scope{
		BizID:       scope.BizID,
		Type:        string(scope.Type),
		Granularity: string(scope.Granularity),
		Filter:      convTargetFilterFromTypes(scope.Filter),
		Items:       scope.Items,
	}
}

func convTargetFilterFromTypes(targetFilter *types.TargetFilter) *TargetFilter {
	if targetFilter == nil {
		return nil
	}

	return &TargetFilter{}
}

// Count count deploy policy by conditions.
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

// List list deploy policy by page and conditions.
func (h *Handler) List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.DeployPolicy, int64, error) {
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

	deployPolicies := conv.SliceToSlice[*DeployPolicy, *types.DeployPolicy](data, convDeployPolicyToTypes)

	return deployPolicies, num, nil
}

// Get get a deploy policy.
func (h *Handler) Get(nCtx contextx.IContext, opts ...OptFn) (*types.DeployPolicy, error) {
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

	return convDeployPolicyToTypes(data), nil
}

func convDeployPolicyToTypes(data *DeployPolicy) *types.DeployPolicy {
	if data == nil {
		return nil
	}

	deployPolicy := &types.DeployPolicy{
		DeployPolicyID: data.DeployPolicyID,
		Operator:       data.Operator,
		Enabled:        data.Enabled,
	}

	deployPolicy.Meta = types.DeployPolicyMeta{
		Name:        data.Meta.Name,
		Description: data.Meta.Description,
	}

	if len(data.Scopes) > 0 {
		deployPolicy.Scopes = conv.SliceToSlice[*Scope, *types.Scope](data.Scopes, convScopeToTypes)
	}

	if len(data.Specs) > 0 {
		deployPolicy.Specs = conv.SliceToSlice[*Spec, *types.DeploySpec](data.Specs, convSpecToTypes)
	}

	return deployPolicy
}

func convScopeToTypes(data *Scope) *types.Scope {
	if data == nil {
		return nil
	}

	return &types.Scope{
		BizID:       data.BizID,
		Type:        types.ScopeType(data.Type),
		Granularity: types.TargetGranularity(data.Granularity),
		Filter:      convTargetFilterToTypes(data.Filter),
		Items:       data.Items,
	}
}

func convTargetFilterToTypes(data *TargetFilter) *types.TargetFilter {
	if data == nil {
		return nil
	}

	// TODO: implement me.

	return &types.TargetFilter{}
}

func convSpecToTypes(data *Spec) *types.DeploySpec {
	if data == nil {
		return nil
	}

	return &types.DeploySpec{
		Type:  types.DeploySpecType(data.Type),
		Param: data.Param,
	}
}

// Delete delete deploy policy.
func (h *Handler) Delete(nCtx contextx.IContext, deployPolicyID int64) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return fmt.Errorf("failed to check tenant id: %v", err)
	}

	filter := base.AliveFilter()
	opts := []base.OptFn{
		WithDeployPolicyID(deployPolicyID),
	}

	for _, opt := range opts {
		filter = opt(filter)
	}

	err := h.tenantDao(nCtx.TenantID()).DeleteMany(nCtx, filter)

	if err != nil {
		logger.G.Sys().With("deploy-policy-id", deployPolicyID).WithErr(err).Error("failed to delete deploy policy")

		return fmt.Errorf("failed to delete deploy policy: %v", err)
	}

	return nil
}

// Exist check deploy policy exist.
func (h *Handler) Exist(nCtx contextx.IContext, opts ...OptFn) (bool, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return false, fmt.Errorf("failed to check tenant id: %v", err)
	}

	filter := base.AliveFilter()
	for _, opt := range opts {
		filter = opt(filter)
	}

	exist, err := h.tenantDao(nCtx.TenantID()).Exist(nCtx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check deploy policy exist: %v", err)
	}

	return exist, nil
}

// UpdateFields update a deploy policy by fields.
func (h *Handler) UpdateFields(nCtx contextx.IContext, fields types.DeployPolicyFields, deployPolicy ...*types.DeployPolicy) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return fmt.Errorf("failed to check tenant id: %v", err)
	}

	if len(deployPolicy) == 0 {
		return base.ErrInvalidParam(fmt.Errorf("deploy policy list is empty"))
	}

	// Check if at least one field is set to update
	if !fields.Meta && !fields.Scopes && !fields.Specs && !fields.Enabled {
		return base.ErrInvalidParam(fmt.Errorf("at least one field must be set to update"))
	}

	tenantID := nCtx.TenantID()

	docs := make([]*base.DocumentFieldUpdate, 0, len(deployPolicy))
	for _, policy := range deployPolicy {
		if policy == nil {
			return base.ErrInvalidItemInParamList()
		}

		updates := generateDeployPolicyUpdates(fields, policy)
		if len(updates) == 0 {
			continue
		}

		updates[FieldKeyOperator] = nCtx.BKUsername()

		docs = append(docs, &base.DocumentFieldUpdate{
			Filter: func() bson.D {
				filter := base.AliveFilter()
				filter = WithDeployPolicyID(policy.DeployPolicyID)(filter)

				return filter
			}(),
			Fields: updates,
		})
	}

	if len(docs) == 0 {
		return base.ErrInvalidParam(fmt.Errorf("no valid updates to apply"))
	}

	if err := h.tenantDao(tenantID).UpdateFieldsBulk(nCtx, docs); err != nil {
		return fmt.Errorf("failed to update deploy policy fields: %v", err)
	}

	return nil
}

// generateDeployPolicyUpdates generate deploy policy updates.
func generateDeployPolicyUpdates(fields types.DeployPolicyFields, deployPolicy *types.DeployPolicy) map[string]any {
	updates := make(map[string]any)

	if fields.Meta {
		updates[FieldKeyMeta] = deployPolicy.Meta
	}

	if fields.Scopes {
		updates[FieldKeyScopes] = deployPolicy.Scopes
	}

	if fields.Specs {
		updates[FieldKeySpecs] = deployPolicy.Specs
	}

	if fields.Enabled {
		updates[FieldKeyEnabled] = deployPolicy.Enabled
	}

	return updates
}
