/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package deploypolicy provide deploy policy storage.
// nolint: nonamedreturns
package deploypolicy

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	daoDeployPolicy "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/deploypolicy"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// StorageName the name of storage.
const StorageName = "deploypolicy"

// NewStorage ...
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

// Storage this is a storage to operate deploy policy table.
type Storage struct {
	basestorage.Storage

	// dao
	daoDeployPolicy daoDeployPolicy.IHandler
}

func (s *Storage) initDao() error {
	s.daoDeployPolicy = daoDeployPolicy.New(s.Database)

	return nil
}

func (s *Storage) check() error {
	if s.daoDeployPolicy == nil {
		return errors.New("dao deploy policy is nil")
	}

	return nil
}

const (
	metricCreateDeployPolicy                     = "create_deploy_policy"
	metricListDeployPolicies                     = "list_deploy_policies"
	metricGetDeployPolicyByID                    = "get_deploy_policy_by_id"
	metricUpdateDeployPolicyFields               = "update_deploy_policy_fields"
	metricDeleteDeployPolicy                     = "delete_deploy_policy"
	metricExistDeployPolicy                      = "exist_deploy_policy"
	metricDiscoverEnabledPoliciesBySpecifyPlugin = "discover_enabled_policies_by_specify_plugin"
	metricRefreshExecuteInfo                     = "refresh_execute_info"
)

// CreateDeployPolicy create deploy policy.
func (s *Storage) CreateDeployPolicy(nCtx contextx.IContext, deployPolicy *types.DeployPolicy) (int64, error) {
	var (
		deployPolicyID int64
		err            error
	)

	err = s.WrapFn(nCtx, metricCreateDeployPolicy, func(nCtx contextx.IContext) error {
		deployPolicyID, err = s.createDeployPolicy(nCtx, deployPolicy)

		return err
	})
	if err != nil {
		return -1, err
	}

	return deployPolicyID, nil
}

// ListDeployPolicies list deploy policies.
func (s *Storage) ListDeployPolicies(nCtx contextx.IContext, page types.Page, condition *types.DeployPolicyCondition) (
	[]*types.DeployPolicy, int64, error) {

	var (
		deployPolicies []*types.DeployPolicy
		total          int64
		err            error
	)

	err = s.WrapFn(nCtx, metricListDeployPolicies, func(nCtx contextx.IContext) error {
		deployPolicies, total, err = s.listDeployPolicies(nCtx, page, condition)

		return err
	})
	if err != nil {
		return nil, 0, err
	}

	return deployPolicies, total, nil
}

// GetDeployPolicyByID get deploy policy by id.
func (s *Storage) GetDeployPolicyByID(nCtx contextx.IContext, deployPolicyID int64) (
	*types.DeployPolicy, error) {

	var (
		deployPolicy *types.DeployPolicy
		err          error
	)

	err = s.WrapFn(nCtx, metricGetDeployPolicyByID, func(nCtx contextx.IContext) error {
		deployPolicy, err = s.getDeployPolicyByID(nCtx, deployPolicyID)

		return err
	})
	if err != nil {
		return nil, err
	}

	return deployPolicy, nil
}

// UpdateDeployPolicyFields update deploy policy fields.
func (s *Storage) UpdateDeployPolicyFields(nCtx contextx.IContext, fields types.DeployPolicyFields, deployPolicy ...*types.DeployPolicy) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricUpdateDeployPolicyFields, func(nCtx contextx.IContext) error {
		return s.updateDeployPolicyFields(nCtx, fields, deployPolicy...)
	})
	if err != nil {
		return err
	}

	return nil
}

// DeleteDeployPolicy delete deploy policy.
func (s *Storage) DeleteDeployPolicy(nCtx contextx.IContext, deployPolicyID int64) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricDeleteDeployPolicy, func(nCtx contextx.IContext) error {
		err = s.deleteDeployPolicy(nCtx, deployPolicyID)

		return err
	})
	if err != nil {
		return err
	}

	return nil
}

// ExistDeployPolicy check deploy policy exist.
func (s *Storage) ExistDeployPolicy(nCtx contextx.IContext, condition *types.DeployPolicyCondition) (bool, error) {
	var (
		exist bool
		err   error
	)

	err = s.WrapFn(nCtx, metricExistDeployPolicy, func(nCtx contextx.IContext) error {
		exist, err = s.existDeployPolicy(nCtx, condition)

		return err
	})
	if err != nil {
		return false, err
	}

	return exist, nil
}

// ==================== IDomainDeployPolicyMgr Functions ====================

// DiscoverEnabledPoliciesBySpecifyPlugin discover enabled policies by specify plugin.
func (s *Storage) DiscoverEnabledPoliciesBySpecifyPlugin(nCtx contextx.IContext, param *types.SpecifyPluginParam) ([]*types.DeployPolicy, error) {
	var (
		policies []*types.DeployPolicy
		err      error
	)

	err = s.WrapFn(nCtx, metricDiscoverEnabledPoliciesBySpecifyPlugin, func(nCtx contextx.IContext) error {
		policies, err = s.discoverEnabledPoliciesBySpecifyPlugin(nCtx, param)

		return err
	})
	if err != nil {
		return nil, err
	}

	return policies, nil
}

// RefreshExecuteInfo refresh execute info.
func (s *Storage) RefreshExecuteInfo(nCtx contextx.IContext, deployPolicy ...*types.DeployPolicy) error {
	var (
		err error
	)

	err = s.WrapFn(nCtx, metricRefreshExecuteInfo, func(nCtx contextx.IContext) error {
		err = s.refreshExecuteInfo(nCtx, deployPolicy...)

		return err
	})
	if err != nil {
		return err
	}

	return nil
}
