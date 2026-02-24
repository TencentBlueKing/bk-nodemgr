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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/configpolicy"
	daoConfigPolicyEvent "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/configpolicy-event"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// StorageName defines the storage name.
const StorageName = "configpolicy"

const (
	metricOperationMatchConfigPolicyNode       = "match"
	metricOperationCountConfigPolicy           = "count"
	metricOperationListConfigPolicy            = "list"
	metricOperationGetConfigPolicy             = "get"
	metricOperationCreateConfigPolicy          = "create"
	metricOperationUpdateConfigPolicy          = "update"
	metricOperationDeleteManyConfigPolicy      = "delete_many"
	metricOperationEnableManyConfigPolicy      = "enable_many"
	metricOperationDisableManyConfigPolicy     = "disable_many"
	metricOperationCountConfigPolicyEvent      = "count_config_policy_event"
	metricOperationListConfigPolicyEvent       = "list_config_policy_event"
	metricOperationCreateManyConfigPolicyEvent = "create_many_config_policy_event"
	metricOperationDistinctConfigPolicyEvent   = "distinct_config_policy_event"
)

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

	daoConfigPolicy      configpolicy.IHandler
	daoConfigPolicyEvent daoConfigPolicyEvent.IHandler
}

func (s *Storage) initDao() error {
	s.daoConfigPolicy = configpolicy.New(s.Database)
	s.daoConfigPolicyEvent = daoConfigPolicyEvent.New(s.Database)

	return nil
}

func (s *Storage) check() error {
	if s.daoConfigPolicy == nil {
		return errors.New("dao configpolicy is nil")
	}

	if s.daoConfigPolicyEvent == nil {
		return errors.New("dao configpolicy event is nil")
	}

	return nil
}

// MatchConfigPolicyNode matches the config policy.
func (s *Storage) MatchConfigPolicyNode(nCtx contextx.IContext,
	bizID, networkAreaID, networkUnitID int64,
	osType criteria.OSType, cpuArch criteria.CPUArch, nodeRole types.NodeRole) (*types.ConfigPolicy, bool, error) {

	var result *types.ConfigPolicy
	var matched bool

	err := s.WrapFn(nCtx, metricOperationMatchConfigPolicyNode, func(nCtx contextx.IContext) error {
		var err error
		result, matched, err = s.matchConfigPolicyNode(nCtx, bizID, networkAreaID, networkUnitID, osType, cpuArch, nodeRole)
		return err
	})

	return result, matched, err
}

// CountConfigPolicy counts the config policy by conditions.
func (s *Storage) CountConfigPolicy(nCtx contextx.IContext, conditions ...*types.ConfigPolicyCondition) (int64, error) {
	var count int64

	err := s.WrapFn(nCtx, metricOperationCountConfigPolicy, func(nCtx contextx.IContext) error {
		var err error
		count, err = s.countConfigPolicy(nCtx, conditions...)
		return err
	})

	return count, err
}

// ListConfigPolicy lists the config policy by page and conditions.
func (s *Storage) ListConfigPolicy(nCtx contextx.IContext, page types.Page, conditions ...*types.ConfigPolicyCondition) (
	[]*types.ConfigPolicy, int64, error) {

	var result []*types.ConfigPolicy
	var total int64

	err := s.WrapFn(nCtx, metricOperationListConfigPolicy, func(nCtx contextx.IContext) error {
		var err error
		result, total, err = s.listConfigPolicy(nCtx, page, conditions...)
		return err
	})

	return result, total, err
}

// GetConfigPolicy gets the config policy.
func (s *Storage) GetConfigPolicy(nCtx contextx.IContext, configPolicyID int64) (*types.ConfigPolicy, error) {
	var configPolicy *types.ConfigPolicy

	err := s.WrapFn(nCtx, metricOperationGetConfigPolicy, func(nCtx contextx.IContext) error {
		var err error
		configPolicy, err = s.getConfigPolicy(nCtx, configPolicyID)
		return err
	})

	return configPolicy, err
}

// CreateConfigPolicy creates the config policy.
func (s *Storage) CreateConfigPolicy(nCtx contextx.IContext, configPolicy *types.ConfigPolicy) (int64, error) {
	var configPolicyID int64

	err := s.WrapFn(nCtx, metricOperationCreateConfigPolicy, func(nCtx contextx.IContext) error {
		var err error
		configPolicyID, err = s.createConfigPolicy(nCtx, configPolicy)
		return err
	})

	return configPolicyID, err
}

// UpdateConfigPolicy updates the config policy.
func (s *Storage) UpdateConfigPolicy(nCtx contextx.IContext, configPolicy *types.ConfigPolicy) error {
	return s.WrapFn(nCtx, metricOperationUpdateConfigPolicy, func(nCtx contextx.IContext) error {
		return s.updateConfigPolicy(nCtx, configPolicy)
	})
}

// DeleteManyConfigPolicy deletes the config policies.
func (s *Storage) DeleteManyConfigPolicy(nCtx contextx.IContext, configPolicyIDs ...int64) error {
	return s.WrapFn(nCtx, metricOperationDeleteManyConfigPolicy, func(nCtx contextx.IContext) error {
		return s.deleteManyConfigPolicy(nCtx, configPolicyIDs...)
	})
}

// EnableManyConfigPolicy enables the config policies.
func (s *Storage) EnableManyConfigPolicy(nCtx contextx.IContext, configPolicyIDs ...int64) error {
	return s.WrapFn(nCtx, metricOperationEnableManyConfigPolicy, func(nCtx contextx.IContext) error {
		return s.enableManyConfigPolicy(nCtx, configPolicyIDs...)
	})
}

// DisableManyConfigPolicy disables the config policies.
func (s *Storage) DisableManyConfigPolicy(nCtx contextx.IContext, configPolicyIDs ...int64) error {
	return s.WrapFn(nCtx, metricOperationDisableManyConfigPolicy, func(nCtx contextx.IContext) error {
		return s.disableManyConfigPolicy(nCtx, configPolicyIDs...)
	})
}

// CountConfigPolicyEvent counts policy events.
func (s *Storage) CountConfigPolicyEvent(nCtx contextx.IContext, conditions ...*types.ConfigPolicyEventCondition) (int64, error) {
	var num int64

	err := s.WrapFn(nCtx, metricOperationCountConfigPolicyEvent, func(nCtx contextx.IContext) error {
		var err error
		if num, err = s.countConfigPolicyEvent(nCtx, conditions...); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to count config policy event")
			return fmt.Errorf("failed to count config policy event: %w", err)
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	return num, nil
}

// ListConfigPolicyEvent lists policy events.
func (s *Storage) ListConfigPolicyEvent(nCtx contextx.IContext, page types.Page, conditions ...*types.ConfigPolicyEventCondition) (
	[]*types.ConfigPolicyEvent, int64, error) {

	var results []*types.ConfigPolicyEvent
	var num int64

	err := s.WrapFn(nCtx, metricOperationListConfigPolicyEvent, func(nCtx contextx.IContext) error {
		var err error
		if results, num, err = s.listConfigPolicyEvent(nCtx, page, conditions...); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to list config policy event")
			return fmt.Errorf("failed to list config policy event: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	return results, num, nil
}

// CreateManyConfigPolicyEvent creates policy events.
func (s *Storage) CreateManyConfigPolicyEvent(nCtx contextx.IContext, events ...*types.ConfigPolicyEvent) error {
	return s.WrapFn(nCtx, metricOperationCreateManyConfigPolicyEvent, func(nCtx contextx.IContext) error {
		if err := s.createManyConfigPolicyEvent(nCtx, events...); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to create many config policy event")
			return fmt.Errorf("failed to create many config policy event: %w", err)
		}
		return nil
	})
}

// DistinctConfigPolicyEvent distincts release by conditions.
func (s *Storage) DistinctConfigPolicyEvent(
	nCtx contextx.IContext, request types.ConfigPolicyEventDistinctRequest, conditions ...*types.ConfigPolicyEventCondition) (
	*types.ConfigPolicyEventDistinctResult, error) {

	var data *types.ConfigPolicyEventDistinctResult

	err := s.WrapFn(nCtx, metricOperationDistinctConfigPolicyEvent, func(nCtx contextx.IContext) error {
		var err error
		if data, err = s.distinctConfigPolicyEvent(nCtx, request, conditions...); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to distinct config policy event")
			return fmt.Errorf("failed to distinct config policy event: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return data, nil
}
