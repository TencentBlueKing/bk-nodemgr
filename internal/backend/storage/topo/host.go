/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package topo provides topology storage for backend.
// nolint: nonamedreturns
package topo

import (
	"context"
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/business"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/host"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// GetHostByID get host by id.
func (s *Storage) GetHostByID(ctx context.Context, hostID int64) (data *types.Host, err error) {
	// record metric.
	metric := s.metric().Start("get_host_by_id")
	defer metric.End(err)

	if ctx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if hostID < 0 {
		return nil, errors.New("host id should be equal or greater than 0")
	}

	hosts, count, err := s.daoHost.List(ctx, types.Page{Limit: 1}, host.WithHostID(hostID))
	if err != nil {
		return nil, fmt.Errorf("failed to get host by id, host-id(%d): %w", hostID, err)
	}

	if count != 1 {
		return nil, fmt.Errorf("failed to get host by id, result count is not 1, host-id(%d), count(%d)",
			hostID, count)
	}

	return hosts[0], nil
}

// UpsertManyHost upserts many hosts.
func (s *Storage) UpsertManyHost(ctx context.Context, hosts ...*types.Host) (err error) {
	// record metric.
	metric := s.metric().Start("upsert_many_host")
	defer metric.End(err)

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if len(hosts) == 0 {
		return nil
	}

	if err = s.daoHost.UpsertMany(ctx, hosts...); err != nil {
		return fmt.Errorf("failed to upsert hosts: %w", err)
	}

	return nil
}

// UpsertManyHostStatic updates or inserts host statics.
func (s *Storage) UpsertManyHostStatic(ctx context.Context, hosts ...*types.Host) (err error) {
	// record metric.
	metric := s.metric().Start("upsert_many_host_static")
	defer metric.End(err)

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if len(hosts) == 0 {
		return nil
	}

	if err = s.daoHost.UpsertStaticMany(ctx, hosts...); err != nil {
		return fmt.Errorf("failed to upsert host statics: %w", err)
	}

	return nil
}

// UpdateManyHostDynamic updates host dynamics.
func (s *Storage) UpdateManyHostDynamic(ctx context.Context, hosts ...*types.Host) (err error) {
	// record metric.
	metric := s.metric().Start("update_many_host_dynamic")
	defer metric.End(err)

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if len(hosts) == 0 {
		return nil
	}

	if err = s.daoHost.UpdateDynamicMany(ctx, hosts...); err != nil {
		return fmt.Errorf("failed to upsert host dynamics: %v", err)
	}

	return nil
}

// ListHost lists hosts by page and conditions.
func (s *Storage) ListHost(ctx context.Context, page types.Page, conditions ...*types.HostCondition) (
	results []*types.Host, num int64, err error) {

	// record metric.
	metric := s.metric().Start("list_host")
	defer metric.End(err)

	opts := convertHostConditionsToOptions(conditions...)
	if results, num, err = s.daoHost.List(ctx, page, opts...); err != nil {
		return nil, 0, err
	}

	return results, num, nil
}

// ListHostOrderByUpdateTime lists hosts by page and conditions, and sort by update time.
func (s *Storage) ListHostOrderByUpdateTime(
	ctx context.Context, page types.Page, conditions ...*types.HostCondition) (results []*types.Host, num int64, err error) {

	// record metric.
	metric := s.metric().Start("list_host_order_by_updatetime")
	defer metric.End(err)

	page.Sort = types.WithSortFields(page.Sort,
		types.WithFieldDesc(base.FieldKeyUpdatedAt))
	opts := convertHostConditionsToOptions(conditions...)
	if results, num, err = s.daoHost.List(ctx, page, opts...); err != nil {
		return nil, 0, err
	}

	return results, num, nil
}

// CountHost counts host by conditions.
func (s *Storage) CountHost(ctx context.Context, conditions ...*types.HostCondition) (num int64, err error) {
	// record metric.
	metric := s.metric().Start("count_host")
	defer metric.End(err)

	opts := convertHostConditionsToOptions(conditions...)
	if num, err = s.daoHost.Count(ctx, opts...); err != nil {
		return 0, err
	}

	return num, nil
}

// DistinctHost distinct host fields.
// nolint:funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (s *Storage) DistinctHost(
	ctx context.Context, request types.HostDistinctRequest, conditions ...*types.HostCondition) (
	data *types.HostDistinctResult, err error) {

	// record metric.
	metric := s.metric().Start("distinct_host")
	defer metric.End(err)

	opts := convertHostConditionsToOptions(conditions...)
	data = new(types.HostDistinctResult)

	gp := gopool.NewPool()
	if request.BizID {
		gp.Go(func() error {
			var err error
			data.BizID, err = s.daoHost.DistinctBizID(ctx, opts...)

			return err
		})
	}
	if request.NodeRole {
		gp.Go(func() error {
			var err error
			data.NodeRole, err = s.daoHost.DistinctNodeRole(ctx, opts...)

			return err
		})
	}
	if request.NodeStatus {
		gp.Go(func() error {
			var err error
			data.NodeStatus, err = s.daoHost.DistinctNodeStatus(ctx, opts...)

			return err
		})
	}
	if request.NodeVersion {
		gp.Go(func() error {
			var err error
			data.NodeVersion, err = s.daoHost.DistinctNodeVersion(ctx, opts...)

			return err
		})
	}
	if request.DeptName {
		gp.Go(func() error {
			var err error
			data.DeptName, err = s.daoHost.DistinctDeptName(ctx, opts...)

			return err
		})
	}
	if request.OSType {
		gp.Go(func() error {
			var err error
			data.OSType, err = s.daoHost.DistinctOSType(ctx, opts...)

			return err
		})
	}
	if request.Arch {
		gp.Go(func() error {
			var err error
			data.Arch, err = s.daoHost.DistinctArch(ctx, opts...)

			return err
		})
	}
	if request.Addressing {
		gp.Go(func() error {
			var err error
			data.Addressing, err = s.daoHost.DistinctAddressing(ctx, opts...)

			return err
		})
	}
	if request.NetworkAreaID {
		gp.Go(func() error {
			var err error
			data.NetworkAreaID, err = s.daoHost.DistinctNetworkAreaID(ctx, opts...)

			return err
		})
	}
	if request.NetworkUnitID {
		gp.Go(func() error {
			var err error
			data.NetworkUnitID, err = s.daoHost.DistinctNetworkUnitID(ctx, opts...)

			return err
		})
	}
	if err = gp.Wait(); err != nil {
		return nil, err
	}

	return data, nil
}

func convertHostConditionsToOptions(conditions ...*types.HostCondition) []host.OptFn {
	opts := make([]host.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				host.WithHostID(condition.ExactInclude.HostID...),
				host.WithBizID(condition.ExactInclude.BizID...),
				host.WithNetworkAreaID(condition.ExactInclude.NetworkAreaID...),
				host.WithNetworkUnitID(condition.ExactInclude.NetworkUnitID...),
				host.WithOSType(condition.ExactInclude.OSType...),
				host.WithArch(condition.ExactInclude.Arch...),
				host.WithNodeRole(condition.ExactInclude.NodeRole...),
				host.WithNodeStatus(condition.ExactInclude.NodeStatus...),
				host.WithNodeVersion(condition.ExactInclude.NodeVersion...),
				host.WithAgentID(condition.ExactInclude.AgentID...),
				host.WithNodeGeneration(condition.ExactInclude.NodeGeneration...),
				host.WithStaticAddressing(condition.ExactInclude.Addressing...),
				host.WithStaticInnerIP(condition.ExactInclude.InnerIP...),
			)
		}

		if condition.ExactExclude != nil {
			opts = append(opts,
				business.WithoutBizID(condition.ExactExclude.HostID...),
				host.WithoutBizID(condition.ExactExclude.BizID...),
				host.WithoutNetworkAreaID(condition.ExactExclude.NetworkAreaID...),
				host.WithoutNetworkUnitID(condition.ExactExclude.NetworkUnitID...),
				host.WithoutOSType(condition.ExactExclude.OSType...),
				host.WithoutArch(condition.ExactExclude.Arch...),
				host.WithoutNodeRole(condition.ExactExclude.NodeRole...),
				host.WithoutNodeStatus(condition.ExactExclude.NodeStatus...),
				host.WithoutNodeVersion(condition.ExactExclude.NodeVersion...),
				host.WithoutAgentID(condition.ExactExclude.AgentID...),
				host.WithNodeGeneration(condition.ExactExclude.NodeGeneration...),
				host.WithStaticAddressing(condition.ExactExclude.Addressing...),
				host.WithStaticInnerIP(condition.ExactExclude.InnerIP...),
			)
		}

		if condition.FuzzyInclude != nil {
			opts = append(opts,
				host.WithFuzzyHostName(condition.FuzzyInclude.HostName...),
				host.WithFuzzyDeptName(condition.FuzzyInclude.DeptName...),
				host.WithFuzzyInnerIP(condition.FuzzyInclude.InnerIP...),
				host.WithFuzzyInnerIPV6(condition.FuzzyInclude.InnerIPV6...),
				host.WithFuzzyOuterIP(condition.FuzzyInclude.OuterIP...),
				host.WithFuzzyOuterIPV6(condition.FuzzyInclude.OuterIPV6...),
			)
		}

		if condition.FuzzyExclude != nil {
			opts = append(opts,
				host.WithoutFuzzyHostName(condition.FuzzyExclude.HostName...),
				host.WithoutFuzzyDeptName(condition.FuzzyExclude.DeptName...),
				host.WithoutFuzzyInnerIP(condition.FuzzyExclude.InnerIP...),
				host.WithoutFuzzyInnerIPV6(condition.FuzzyExclude.InnerIPV6...),
				host.WithoutFuzzyOuterIP(condition.FuzzyExclude.OuterIP...),
				host.WithoutFuzzyOuterIPV6(condition.FuzzyExclude.OuterIPV6...),
			)
		}
	}

	return opts
}

// DeleteManyHost delete many hosts by hostIDs.
func (s *Storage) DeleteManyHost(ctx context.Context, hostIDs ...int64) (err error) {
	// record metric.
	metric := s.metric().Start("delete_many_host")
	defer metric.End(err)

	if ctx == nil {
		return errors.New("ctx is nil")
	}

	if len(hostIDs) == 0 {
		return nil
	}

	if err = s.daoHost.DeleteMany(ctx, hostIDs...); err != nil {
		return err
	}

	return nil
}

// FindHostWithDynamic finds hosts with dynamic fields.
func (s *Storage) FindHostWithDynamic(ctx context.Context, page types.Page, conditions ...*types.HostCondition) (
	results []*types.Host, err error) {

	// record metric.
	metric := s.metric().Start("find_host_with_dynamic")
	defer metric.End(err)

	if ctx == nil {
		return nil, basestorage.ErrNilContent()
	}

	opts := convertHostConditionsToOptions(conditions...)
	if results, err = s.daoHost.FindWithDynamic(ctx, page, opts...); err != nil {
		return nil, err
	}

	return results, nil
}

// UpdateHostDynamicFields updates the dynamic fields of a host.
func (s *Storage) UpdateHostDynamicFields(ctx context.Context, fields types.HostDynamicFields, hosts ...*types.Host) (err error) {
	// record metric.
	metric := s.metric().Start("update_host_dynamic_fields")
	defer metric.End(err)

	if ctx == nil {
		return basestorage.ErrNilContent()
	}

	if err = s.daoHost.UpdateDynamicFields(ctx, fields, hosts...); err != nil {
		return fmt.Errorf("failed to update host dynamic fields: %w", err)
	}

	return nil
}
