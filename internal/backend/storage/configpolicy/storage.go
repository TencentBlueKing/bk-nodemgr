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
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/configpolicy"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// IStorage defines the config policy storage interface.
type IStorage interface {
	basestorage.Interface

	// MatchConfigPolicy matches the config policy.
	MatchConfigPolicy(ctx context.Context,
		bizID, networkAreaID, networkUnitID int64,
		osType criteria.OSType, cpuArch criteria.CPUArch) (*types.ConfigPolicy, bool, error)

	// CountConfigPolicy counts the config policy by conditions.
	CountConfigPolicy(ctx context.Context, conditions ...*types.ConfigPolicyCondition) (int64, error)

	// ListConfigPolicy lists the config policy by page and conditions.
	ListConfigPolicy(ctx context.Context, page types.Page, conditions ...*types.ConfigPolicyCondition) (
		[]*types.ConfigPolicy, int64, error)

	// GetConfigPolicy gets the config policy.
	GetConfigPolicy(ctx context.Context, configPolicyID int64) (*types.ConfigPolicy, error)

	// CreateConfigPolicy creates the config policy.
	CreateConfigPolicy(ctx context.Context, configPolicy *types.ConfigPolicy) (int64, error)

	// UpdateConfigPolicy updates the config policy.
	UpdateConfigPolicy(ctx context.Context, configPolicy *types.ConfigPolicy) error

	// DeleteManyConfigPolicy deletes the config policies.
	DeleteManyConfigPolicy(ctx context.Context, configPolicyIDs ...int64) error

	// EnableManyConfigPolicy enables the config policies.
	EnableManyConfigPolicy(ctx context.Context, configPolicyIDs ...int64) error

	// DisableManyConfigPolicy disables the config policies.
	DisableManyConfigPolicy(ctx context.Context, configPolicyIDs ...int64) error
}

// StorageName defines the storage name.
const StorageName = "configpolicy"

// NewStorage creates a new release storage.
func NewStorage(client *mongo.Client, database string, logger logger.Logger) (*Storage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &Storage{
		Storage: basestorage.Storage{
			Name:     StorageName,
			Database: client.Database(database),
			Logger:   logger,
		},
	}
	err := basestorage.InitStorage(&s.Storage,
		basestorage.WithStartFunc(s.initDao),
		basestorage.WithCheckFunc(s.check))
	if err != nil {
		s.Logger.Errorf("new storage failed, err: %v", err)
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
	s.daoConfigPolicy = configpolicy.New(s.Database, s.Logger)

	return nil
}

func (s *Storage) check() error {
	if s.daoConfigPolicy == nil {
		return errors.New("dao configpolicy is nil")
	}

	return nil
}

// MatchConfigPolicy matches the config policy.
func (s *Storage) MatchConfigPolicy(ctx context.Context,
	bizID, networkAreaID, networkUnitID int64,
	osType criteria.OSType, cpuArch criteria.CPUArch) (*types.ConfigPolicy, bool, error) {

	results, _, err := s.daoConfigPolicy.List(ctx,
		types.Page{
			Limit: 1,
			Sort:  "-" + configpolicy.FieldKeyUpdatedAt,
		},
		configpolicy.WithEnabledScope(bizID, networkAreaID, networkUnitID, osType, cpuArch),
	)
	if err != nil {
		return nil, false, fmt.Errorf("list config policy failed, err: %w")
	}

	if len(results) == 0 {
		return nil, false, nil
	}

	return results[0], true, nil
}

// CountConfigPolicy counts the config policy by conditions.
func (s *Storage) CountConfigPolicy(ctx context.Context, conditions ...*types.ConfigPolicyCondition) (int64, error) {
	opts, err := convertConfigPolicyConditionsToOptions(conditions...)
	if err != nil {
		return 0, err
	}

	return s.daoConfigPolicy.Count(ctx, opts...)
}

// ListConfigPolicy lists the config policy by page and conditions.
func (s *Storage) ListConfigPolicy(ctx context.Context, page types.Page, conditions ...*types.ConfigPolicyCondition) (
	[]*types.ConfigPolicy, int64, error) {

	opts, err := convertConfigPolicyConditionsToOptions(conditions...)
	if err != nil {
		return nil, 0, err
	}

	return s.daoConfigPolicy.List(ctx, page, opts...)
}

// GetConfigPolicy gets the config policy.
func (s *Storage) GetConfigPolicy(ctx context.Context, configPolicyID int64) (*types.ConfigPolicy, error) {
	return s.daoConfigPolicy.Get(ctx, configPolicyID)
}

// CreateConfigPolicy creates the config policy.
func (s *Storage) CreateConfigPolicy(ctx context.Context, configPolicy *types.ConfigPolicy) (int64, error) {
	configPolicy.UpdatedAt = time.Now()

	return s.daoConfigPolicy.Create(ctx, configPolicy)
}

// UpdateConfigPolicy updates the config policy.
func (s *Storage) UpdateConfigPolicy(ctx context.Context, configPolicy *types.ConfigPolicy) error {
	configPolicy.UpdatedAt = time.Now()

	return s.daoConfigPolicy.UpdateMany(ctx, configPolicy)
}

// DeleteManyConfigPolicy deletes the config policies.
func (s *Storage) DeleteManyConfigPolicy(ctx context.Context, configPolicyIDs ...int64) error {
	return s.daoConfigPolicy.DeleteMany(ctx, configPolicyIDs...)
}

// EnableManyConfigPolicy enables the config policies.
func (s *Storage) EnableManyConfigPolicy(ctx context.Context, configPolicyIDs ...int64) error {
	return s.daoConfigPolicy.EnableMany(ctx, configPolicyIDs...)
}

// DisableManyConfigPolicy disables the config policies.
func (s *Storage) DisableManyConfigPolicy(ctx context.Context, configPolicyIDs ...int64) error {
	return s.daoConfigPolicy.DisableMany(ctx, configPolicyIDs...)
}

func convertConfigPolicyConditionsToOptions(conditions ...*types.ConfigPolicyCondition) ([]configpolicy.OptFn, error) {
	opts := make([]configpolicy.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				configpolicy.WithConfigPolicyID(condition.ExactInclude.ConfigPolicyID...),
				configpolicy.WithBizID(condition.ExactInclude.BizID...),
				configpolicy.WithNodeRole(condition.ExactInclude.NodeRole...),
				configpolicy.WithEnabled(condition.ExactInclude.Enabled...))
		}

		if condition.FuzzyInclude != nil {
			opts = append(opts,
				configpolicy.WithFuzzyConfigPolicyName(condition.FuzzyInclude.ConfigPolicyName...),
				configpolicy.WithFuzzyOperator(condition.FuzzyInclude.Operator...))
		}

		if condition.ExactExclude != nil {
			return nil, errors.New("exact exclude is not supported")
		}

		if condition.FuzzyExclude != nil {
			return nil, errors.New("fuzzy exclude is not supported")
		}
	}

	return opts, nil
}
