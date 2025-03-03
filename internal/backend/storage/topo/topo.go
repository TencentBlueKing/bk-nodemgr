/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package topo provides topology storage for nodeman.
package topo

import (
	"context"
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/business"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/host"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// StorageName ...
const StorageName = "topo"

// NewStorage ...
func NewStorage(client *mongo.Client, database string, logger logger.Logger) (Storage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &storage{
		Storage: base.Storage{
			Name:     StorageName,
			Database: client.Database(database),
			Logger:   logger,
		},
	}
	err := base.InitStorage(&s.Storage,
		base.WithStartFunc(s.initDao),
		base.WithCheckFunc(s.check))
	if err != nil {
		s.Logger.Errorf("new storage failed, err: %v", err)
		return nil, err
	}

	return s, nil
}

// storage implements the Storage interface.
type storage struct {
	base.Storage

	daoBusiness business.Handler

	daoHost host.Handler
}

func (s *storage) initDao() error {
	s.daoBusiness = business.New(s.Database, s.Logger)
	s.daoHost = host.New(s.Database, s.Logger)

	return nil
}

func (s *storage) check() error {
	if s.daoBusiness == nil {
		return errors.New("dao business is nil")
	}

	return nil
}

// UpsertBusiness updates or inserts a business.
func (s *storage) UpsertBusiness(ctx context.Context, biz ...*types.Business) error {
	if ctx == nil {
		return base.ErrNilContent()
	}

	if biz == nil {
		return base.ErrUpsertNilData()
	}

	if err := s.daoBusiness.UpsertMany(ctx, biz...); err != nil {
		return fmt.Errorf("failed to upsert business: %v", err)
	}

	return nil
}

// ListBusinesses lists businesses by page and conditions.
func (s *storage) ListBusinesses(ctx context.Context, page types.Page, conditions ...BusinessCondition) (
	[]*types.Business, int64, error) {

	opts := make([]business.OptFn, 0)
	for _, condition := range conditions {
		switch condition.Type {
		case types.ConditionTypeExactInclude:
			opts = append(opts,
				business.WithBizID(condition.BizID...),
				business.WithBizName(condition.BizName...),
			)

		case types.ConditionTypeExactExclude:
			opts = append(opts,
				business.WithoutBizID(condition.BizID...),
				business.WithoutBizName(condition.BizName...),
			)

		default:
			return nil, 0, fmt.Errorf("get unexpected condition type: %s", condition.Type)
		}
	}

	return s.daoBusiness.List(ctx, page, opts...)
}

// UpsertHosts ...
func (s *storage) UpsertHosts(ctx context.Context, hosts ...*types.Host) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}

	if len(hosts) == 0 {
		return nil
	}

	if err := s.daoHost.UpsertMany(ctx, hosts...); err != nil {
		return fmt.Errorf("failed to upsert hosts: %v", err)
	}

	return nil
}

// UpsertHostStatics updates or inserts host statics.
func (s *storage) UpsertHostStatics(ctx context.Context, hosts ...*types.Host) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}

	if len(hosts) == 0 {
		return nil
	}

	if err := s.daoHost.UpsertStaticMany(ctx, hosts...); err != nil {
		return fmt.Errorf("failed to upsert host statics: %v", err)
	}

	return nil
}

// nolint:cyclop
// ListHosts lists hosts by page and conditions.
func (s *storage) ListHosts(ctx context.Context, page types.Page, conditions ...HostCondition) (
	[]*types.Host, int64, error) {

	opts := make([]host.OptFn, 0)
	for _, condition := range conditions {
		switch condition.Type {
		case types.ConditionTypeExactInclude:
			if condition.Exact != nil {
				opts = append(opts,
					host.WithHostID(condition.Exact.HostID...),
					host.WithBizID(condition.Exact.BizID...),
					host.WithNetworkAreaID(condition.Exact.NetworkAreaID...),
					host.WithOSType(condition.Exact.OSType...),
					host.WithNodeRole(condition.Exact.NodeRole...),
					host.WithNodeStatus(condition.Exact.NodeStatus...),
					host.WithNodeVersion(condition.Exact.NodeVersion...),
					host.WithAgentID(condition.Exact.AgentID...),
				)
			}

		case types.ConditionTypeExactExclude:
			if condition.Exact != nil {
				opts = append(opts,
					business.WithoutBizID(condition.Exact.HostID...),
					host.WithoutBizID(condition.Exact.BizID...),
					host.WithoutNetworkAreaID(condition.Exact.NetworkAreaID...),
					host.WithoutOSType(condition.Exact.OSType...),
					host.WithoutNodeRole(condition.Exact.NodeRole...),
					host.WithoutNodeStatus(condition.Exact.NodeStatus...),
					host.WithoutNodeVersion(condition.Exact.NodeVersion...),
					host.WithoutAgentID(condition.Exact.AgentID...),
				)
			}

		case types.ConditionTypeFuzzyInclude:
			if condition.Fuzzy != nil {
				opts = append(opts,
					host.WithFuzzyHostName(condition.Fuzzy.HostName...),
					host.WithFuzzyDeptName(condition.Fuzzy.DeptName...),
					host.WithFuzzyInnerIP(condition.Fuzzy.InnerIP...),
					host.WithFuzzyInnerIPV6(condition.Fuzzy.InnerIPV6...),
					host.WithFuzzyOuterIP(condition.Fuzzy.OuterIP...),
					host.WithFuzzyOuterIPV6(condition.Fuzzy.OuterIPV6...),
				)
			}

		case types.ConditionTypeFuzzyExclude:
			if condition.Fuzzy != nil {
				opts = append(opts,
					host.WithoutFuzzyHostName(condition.Fuzzy.HostName...),
					host.WithoutFuzzyDeptName(condition.Fuzzy.DeptName...),
					host.WithoutFuzzyInnerIP(condition.Fuzzy.InnerIP...),
					host.WithoutFuzzyInnerIPV6(condition.Fuzzy.InnerIPV6...),
					host.WithoutFuzzyOuterIP(condition.Fuzzy.OuterIP...),
					host.WithoutFuzzyOuterIPV6(condition.Fuzzy.OuterIPV6...),
				)
			}

		default:
			return nil, 0, fmt.Errorf("get unexpected condition type: %s", condition.Type)
		}
	}

	return s.daoHost.List(ctx, page, opts...)
}
