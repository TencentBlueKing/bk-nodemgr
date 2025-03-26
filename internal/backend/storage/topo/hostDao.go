/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package topo ...
package topo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/business"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/host"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// GetHostByID get host by id.
func (s *Storage) GetHostByID(ctx context.Context, hostID int64) (*types.Host, error) {
	if ctx == nil {
		return nil, base.ErrNilContent()
	}

	if hostID < 0 {
		return nil, errors.New("host id should be equal or greater than 0")
	}

	hosts, count, err := s.daoHost.List(ctx, types.Page{}, host.WithHostID(hostID))
	if err != nil {
		return nil, fmt.Errorf("failed to get host by id, host-id(%d), err: %w", hostID, err)
	}

	if count != 1 {
		return nil, fmt.Errorf("failed to get host by id, result count is not 1, host-id(%d), count(%d)",
			hostID, count)
	}

	return hosts[0], nil
}

// UpsertManyHost upserts many hosts.
func (s *Storage) UpsertManyHost(ctx context.Context, hosts ...*types.Host) error {
	if ctx == nil {
		return base.ErrNilContent()
	}

	if len(hosts) == 0 {
		return nil
	}

	if err := s.daoHost.UpsertMany(ctx, hosts...); err != nil {
		return fmt.Errorf("failed to upsert hosts: %v", err)
	}

	return nil
}

// UpsertManyHostStatic updates or inserts host statics.
func (s *Storage) UpsertManyHostStatic(ctx context.Context, hosts ...*types.Host) error {
	if ctx == nil {
		return base.ErrNilContent()
	}

	if len(hosts) == 0 {
		return nil
	}

	if err := s.daoHost.UpsertStaticMany(ctx, hosts...); err != nil {
		return fmt.Errorf("failed to upsert host statics: %v", err)
	}

	return nil
}

// UpdateManyHostDynamic updates host dynamics.
func (s *Storage) UpdateManyHostDynamic(ctx context.Context, hosts ...*types.Host) error {
	if ctx == nil {
		return base.ErrNilContent()
	}

	if len(hosts) == 0 {
		return nil
	}

	if err := s.daoHost.UpdateDynamicMany(ctx, hosts...); err != nil {
		return fmt.Errorf("failed to upsert host dynamics: %v", err)
	}

	return nil
}

// nolint:cyclop
// ListHost lists hosts by page and conditions.
func (s *Storage) ListHost(ctx context.Context, page types.Page, conditions ...types.HostCondition) (
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
					host.WithNetworkUnitID(condition.Exact.NetworkUnitID...),
					host.WithOSType(condition.Exact.OSType...),
					host.WithNodeRole(condition.Exact.NodeRole...),
					host.WithNodeStatus(condition.Exact.NodeStatus...),
					host.WithNodeVersion(condition.Exact.NodeVersion...),
					host.WithNodeGeneration(condition.Exact.NodeGeneration...),
					host.WithAgentID(condition.Exact.AgentID...),
				)
			}

		case types.ConditionTypeExactExclude:
			if condition.Exact != nil {
				opts = append(opts,
					business.WithoutBizID(condition.Exact.HostID...),
					host.WithoutBizID(condition.Exact.BizID...),
					host.WithoutNetworkAreaID(condition.Exact.NetworkAreaID...),
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

// nolint:cyclop
// CountHost counts host by conditions.
func (s *Storage) CountHost(ctx context.Context, conditions ...types.HostCondition) (int64, error) {
	opts := make([]host.OptFn, 0)
	for _, condition := range conditions {
		switch condition.Type {
		case types.ConditionTypeExactInclude:
			if condition.Exact != nil {
				opts = append(opts,
					host.WithHostID(condition.Exact.HostID...),
					host.WithBizID(condition.Exact.BizID...),
					host.WithNetworkAreaID(condition.Exact.NetworkAreaID...),
					host.WithNetworkUnitID(condition.Exact.NetworkUnitID...),
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
					host.WithoutNetworkUnitID(condition.Exact.NetworkUnitID...),
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
			return 0, fmt.Errorf("get unexpected condition type: %s", condition.Type)
		}
	}

	return s.daoHost.Count(ctx, opts...)
}

// GetAgentAccessEndpoints get agent access endpoints by unit id.
// nolint: nonamedreturns
func (s *Storage) GetAgentAccessEndpoints(ctx context.Context, unitID int64) (clusterEndpoints []string,
	dataEndpoints []string, fileEndpoints []string, err error) {

	if ctx == nil {
		return nil, nil, nil, base.ErrNilContent()
	}

	if unitID < 0 {
		return nil, nil, nil, errors.New("unit id should be equal or greater than 0")
	}

	hosts, count, err := s.daoHost.List(ctx, types.UnlimitedPage(),
		host.WithNetworkUnitID(unitID),
		host.WithNodeRole(types.NodeRoleProxy),
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get host by unit id, unit-id(%d), err: %w", unitID, err)
	}
	if count == 0 {
		return nil, nil, nil, fmt.Errorf("failed to get host by unit id, result count is 0, unit-id(%d)", unitID)
	}

	clusterEndpoints = make([]string, 0)
	dataEndpoints = make([]string, 0)
	fileEndpoints = make([]string, 0)
	for idx := range hosts {
		innerIps := strings.Split(hosts[idx].Static.InnerIP, ",")
		for _, ip := range innerIps {
			clusterEndpoints = append(clusterEndpoints, fmt.Sprintf("%s:%d", ip, hosts[idx].Dynamic.ProxyClusterPort))
			dataEndpoints = append(dataEndpoints, fmt.Sprintf("%s:%d", ip, hosts[idx].Dynamic.ProxyDataPort))
			fileEndpoints = append(fileEndpoints, fmt.Sprintf("%s:%d", ip, hosts[idx].Dynamic.ProxyFilePort))
		}
	}

	return clusterEndpoints, dataEndpoints, fileEndpoints, nil
}
