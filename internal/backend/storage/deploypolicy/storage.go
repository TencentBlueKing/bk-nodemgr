/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package deploy_policy provide deploy policy storage.
// nolint: nonamedreturns
package deploypolicy

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	daoDeployPolicy "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/deploy_policy"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// StorageName the name of storage.
const StorageName = "deploy_policy"

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

func (s *Storage) metric() *storage.MetricData {
	return storage.Metric(StorageName)
}

// CreateDeployPolicy create deploy policy.
func (s *Storage) CreateDeployPolicy(nCtx contextx.IContext, deployPolicy *types.DeployPolicy) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("create_deploy_policy")
	defer metric.End(err)

	err = s.createDeployPolicy(nCtx, deployPolicy)

	return err
}

// ListDeployPolicies list deploy policies.
func (s *Storage) ListDeployPolicies(nCtx contextx.IContext, page types.Page) (
	deployPolicies []*types.DeployPolicy, total int64, err error) {

	// record metric.
	metric := s.metric().Start("list_deploy_policies")
	defer metric.End(err)

	deployPolicies, total, err = s.listDeployPolicies(nCtx, page)

	return deployPolicies, total, err
}

// GetDeployPolicyByID get deploy policy by id.
func (s *Storage) GetDeployPolicyByID(nCtx contextx.IContext, deployPolicyID int64) (
	deployPolicy *types.DeployPolicy, err error) {

	// record metric.
	metric := s.metric().Start("get_deploy_policy_by_id")
	defer metric.End(err)

	deployPolicy, err = s.getDeployPolicyByID(nCtx, deployPolicyID)

	return deployPolicy, err
}

// UpdateDeployPolicy update deploy policy.
func (s *Storage) UpdateDeployPolicy(nCtx contextx.IContext, deployPolicyID int64, deployPolicy *types.DeployPolicy) (err error) {
	// record metric.
	metric := s.metric().Start("update_deploy_policy")
	defer metric.End(err)

	err = s.updateDeployPolicy(nCtx, deployPolicyID, deployPolicy)

	return err
}

// DeleteDeployPolicy delete deploy policy.
func (s *Storage) DeleteDeployPolicy(nCtx contextx.IContext, deployPolicyID int64) (err error) {
	// record metric.
	metric := s.metric().Start("delete_deploy_policy")
	defer metric.End(err)

	err = s.deleteDeployPolicy(nCtx, deployPolicyID)

	return err
}

// ExistDeployPolicy check deploy policy exist.
func (s *Storage) ExistDeployPolicy(nCtx contextx.IContext, deployPolicyID int64) (exist bool, err error) {
	// record metric.
	metric := s.metric().Start("exist_deploy_policy")
	defer metric.End(err)

	exist, err = s.existDeployPolicy(nCtx, deployPolicyID)

	return exist, err
}
