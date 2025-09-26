/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package configpolicy provides the config policy storage interface.
package configpolicy

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/configpolicy"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// StorageName defines the storage name.
const StorageName = "configpolicy"

// NewStorage creates a new release storage.
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

// Storage implements IStorage.
type Storage struct {
	basestorage.Storage

	daoConfigPolicy configpolicy.IHandler
}

func (s *Storage) initDao() error {
	s.daoConfigPolicy = configpolicy.New(s.Database)

	return nil
}

func (s *Storage) check() error {
	if s.daoConfigPolicy == nil {
		return errors.New("dao configpolicy is nil")
	}

	return nil
}

// MatchConfigPolicy matches the config policy.
func (s *Storage) MatchConfigPolicy(nCtx contextx.IContext,
	bizID, networkAreaID, networkUnitID int64,
	osType criteria.OSType, cpuArch criteria.CPUArch) (*types.ConfigPolicy, bool, error) {

	var (
		result  *types.ConfigPolicy
		matched bool
		err     error
	)

	// record metric.
	metric := s.metric().Start("match")
	defer metric.End(err)

	result, matched, err = s.matchConfigPolicy(nCtx, bizID, networkAreaID, networkUnitID, osType, cpuArch)

	return result, matched, err
}

// CountConfigPolicy counts the config policy by conditions.
func (s *Storage) CountConfigPolicy(nCtx contextx.IContext, conditions ...*types.ConfigPolicyCondition) (int64, error) {
	var (
		count int64
		err   error
	)

	// record metric.
	metric := s.metric().Start("count")
	defer metric.End(err)

	count, err = s.countConfigPolicy(nCtx, conditions...)

	return count, err
}

// ListConfigPolicy lists the config policy by page and conditions.
func (s *Storage) ListConfigPolicy(nCtx contextx.IContext, page types.Page, conditions ...*types.ConfigPolicyCondition) (
	[]*types.ConfigPolicy, int64, error) {

	var (
		result []*types.ConfigPolicy
		total  int64
		err    error
	)

	// record metric.
	metric := s.metric().Start("list")
	defer metric.End(err)

	result, total, err = s.listConfigPolicy(nCtx, page, conditions...)

	return result, total, err
}

// GetConfigPolicy gets the config policy.
func (s *Storage) GetConfigPolicy(nCtx contextx.IContext, configPolicyID int64) (*types.ConfigPolicy, error) {
	var (
		configPolicy *types.ConfigPolicy
		err          error
	)

	// record metric.
	metric := s.metric().Start("get")
	defer metric.End(err)

	configPolicy, err = s.getConfigPolicy(nCtx, configPolicyID)

	return configPolicy, err
}

// CreateConfigPolicy creates the config policy.
func (s *Storage) CreateConfigPolicy(nCtx contextx.IContext, configPolicy *types.ConfigPolicy) (int64, error) {
	var (
		configPolicyID int64
		err            error
	)

	// record metric.
	metric := s.metric().Start("create")
	defer metric.End(err)

	configPolicyID, err = s.createConfigPolicy(nCtx, configPolicy)

	return configPolicyID, err
}

// UpdateConfigPolicy updates the config policy.
func (s *Storage) UpdateConfigPolicy(nCtx contextx.IContext, configPolicy *types.ConfigPolicy) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("update")
	defer metric.End(err)

	err = s.updateConfigPolicy(nCtx, configPolicy)

	return err
}

// DeleteManyConfigPolicy deletes the config policies.
func (s *Storage) DeleteManyConfigPolicy(nCtx contextx.IContext, configPolicyIDs ...int64) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("delete_many")
	defer metric.End(err)

	err = s.deleteManyConfigPolicy(nCtx, configPolicyIDs...)

	return err
}

// EnableManyConfigPolicy enables the config policies.
func (s *Storage) EnableManyConfigPolicy(nCtx contextx.IContext, configPolicyIDs ...int64) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("enable_many")
	defer metric.End(err)

	err = s.enableManyConfigPolicy(nCtx, configPolicyIDs...)

	return err
}

// DisableManyConfigPolicy disables the config policies.
func (s *Storage) DisableManyConfigPolicy(nCtx contextx.IContext, configPolicyIDs ...int64) error {
	var (
		err error
	)

	// record metric.
	metric := s.metric().Start("disable_many")
	defer metric.End(err)

	err = s.disableManyConfigPolicy(nCtx, configPolicyIDs...)

	return err
}

func (s *Storage) metric() *storage.MetricData {
	return storage.Metric(StorageName)
}
