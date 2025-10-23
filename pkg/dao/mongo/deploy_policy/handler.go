/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package deploy_policy

import (
	"fmt"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IHandler deploy policy handler interface.
type IHandler interface {
	// Create create a new deploy policy.
	Create(nCtx contextx.IContext, deployPolicy *types.DeployPolicy) error

	// Count count deploy policy by conditions.
	Count(nCtx contextx.IContext, opts ...OptFn) (int64, error)

	// List lists deploy policy by page and conditions.
	List(nCtx contextx.IContext, page types.Page, opts ...OptFn) ([]*types.DeployPolicy, int64, error)

	// Get gets a deploy policy by conditions.
	Get(nCtx contextx.IContext, opts ...OptFn) (*types.DeployPolicy, error)

	// Update update a deploy policy by conditions.
	Update(nCtx contextx.IContext, deployPolicyID int64, deployPolicy *types.DeployPolicy) error

	// Delete delete a deploy policy by conditions.
	Delete(nCtx contextx.IContext, deployPolicyID int64) error

	// Exist check a deploy policy exist by conditions.
	Exist(nCtx contextx.IContext, deployPolicyID int64) (bool, error)
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
func (h *Handler) Create(nCtx contextx.IContext, deployPolicy *types.DeployPolicy) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return err
	}

	data := convDeployPolicyFromTypes(deployPolicy)

	// generate deploy policy id.
	newDeployPolicyID, err := h.tenantDao(nCtx.TenantID()).counter.Generate(nCtx, tableNamePrefix)
	if err != nil {
		return fmt.Errorf("failed to generate deploy policy id, err: %w", err)
	}
	data.DeployPolicyID = newDeployPolicyID

	if err := h.tenantDao(nCtx.TenantID()).Create(nCtx, data); err != nil {
		return fmt.Errorf("failed to create deploy policy, err: %w", err)
	}

	return nil
}

func convDeployPolicyFromTypes(deployPolicy *types.DeployPolicy) *DeployPolicy {
	data := &DeployPolicy{
		TenantID:       deployPolicy.TenantID,
		DeployPolicyID: deployPolicy.DeployPolicyID,
	}

	return data
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

	deployPolicies := make([]*types.DeployPolicy, len(data))
	for idx, dp := range data {
		deployPolicies[idx] = convertDeployPolicyToTypes(dp)
	}

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

	return convertDeployPolicyToTypes(data), nil
}

func convertDeployPolicyToTypes(data *DeployPolicy) *types.DeployPolicy {
	deployPolicy := &types.DeployPolicy{
		TenantID:       data.TenantID,
		DeployPolicyID: data.DeployPolicyID,
	}

	return deployPolicy
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
func (h *Handler) Exist(nCtx contextx.IContext, deployPolicyID int64) (bool, error) {
	if err := nCtx.CheckTenantID(); err != nil {
		return false, fmt.Errorf("failed to check tenant id: %v", err)
	}

	filter := base.AliveFilter()
	opts := []base.OptFn{
		WithDeployPolicyID(deployPolicyID),
	}
	for _, opt := range opts {
		filter = opt(filter)
	}

	exist, err := h.tenantDao(nCtx.TenantID()).Exist(nCtx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check deploy policy exist: %v", err)
	}

	return exist, nil
}

// Update update deploy policy.
func (h *Handler) Update(nCtx contextx.IContext, deployPolicyID int64, deployPolicy *types.DeployPolicy) error {
	if err := nCtx.CheckTenantID(); err != nil {
		return fmt.Errorf("failed to check tenant id: %v", err)
	}

	if deployPolicy == nil {
		return fmt.Errorf("deploy policy is nil")
	}

	if deployPolicy.DeployPolicyID != deployPolicyID {
		return fmt.Errorf("deploy policy id not match")
	}

	filter := base.AliveFilter()
	opts := []base.OptFn{
		WithDeployPolicyID(deployPolicyID),
	}

	for _, opt := range opts {
		filter = opt(filter)
	}

	data := convDeployPolicyFromTypes(deployPolicy)

	updates := []*base.DocumentFieldUpdate{
		{
			Filter: filter,
			Fields: map[string]any{
				FieldKeyDeployPolicyID: data.DeployPolicyID,
			},
		},
	}

	if err := h.tenantDao(nCtx.TenantID()).UpdateFieldsBulk(nCtx, updates); err != nil {
		return fmt.Errorf("failed to update deploy policy: %v", err)
	}

	return nil
}
