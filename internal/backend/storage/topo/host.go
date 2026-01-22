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
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/host"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// GetHostByID get host by id.
func (s *Storage) GetHostByID(nCtx contextx.IContext, hostID int64) (data *types.Host, err error) {
	// record metric.
	metric := s.metric().Start("get_host_by_id")
	defer metric.End(err)

	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if hostID < 0 {
		return nil, errors.New("host id should be equal or greater than 0")
	}

	hosts, count, err := s.daoHost.List(nCtx, types.Page{Limit: 1}, host.WithHostID(hostID))
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
func (s *Storage) UpsertManyHost(nCtx contextx.IContext, hosts ...*types.Host) (err error) {
	// record metric.
	metric := s.metric().Start("upsert_many_host")
	defer metric.End(err)

	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if len(hosts) == 0 {
		return nil
	}

	if err = s.daoHost.UpsertMany(nCtx, hosts...); err != nil {
		return fmt.Errorf("failed to upsert hosts: %w", err)
	}

	return nil
}

// UpsertManyHostStatic updates or inserts host statics.
func (s *Storage) UpsertManyHostStatic(nCtx contextx.IContext, hosts ...*types.Host) (err error) {
	// record metric.
	metric := s.metric().Start("upsert_many_host_static")
	defer metric.End(err)

	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if len(hosts) == 0 {
		return nil
	}

	if err = s.daoHost.UpsertStaticMany(nCtx, hosts...); err != nil {
		return fmt.Errorf("failed to upsert host statics: %w", err)
	}

	return nil
}

// UpdateManyHostDynamic updates host dynamics.
func (s *Storage) UpdateManyHostDynamic(nCtx contextx.IContext, hosts ...*types.Host) (err error) {
	// record metric.
	metric := s.metric().Start("update_many_host_dynamic")
	defer metric.End(err)

	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if len(hosts) == 0 {
		return nil
	}

	if err = s.daoHost.UpdateDynamicMany(nCtx, hosts...); err != nil {
		return fmt.Errorf("failed to upsert host dynamics: %v", err)
	}

	return nil
}

// ListHost lists hosts by page and conditions.
func (s *Storage) ListHost(nCtx contextx.IContext, page types.Page, conditions ...*types.HostCondition) (
	results []*types.Host, num int64, err error) {

	// record metric.
	metric := s.metric().Start("list_host")
	defer metric.End(err)

	opts := convertHostConditionsToOptions(conditions...)
	if results, num, err = s.daoHost.List(nCtx, page, opts...); err != nil {
		return nil, 0, err
	}

	return results, num, nil
}

// ListHostOrderByUpdateTime lists hosts by page and conditions, and sort by update time.
func (s *Storage) ListHostOrderByUpdateTime(
	nCtx contextx.IContext, page types.Page, conditions ...*types.HostCondition) (results []*types.Host, num int64, err error) {

	// record metric.
	metric := s.metric().Start("list_host_order_by_updatetime")
	defer metric.End(err)

	page.Sort = types.WithSortFields(page.Sort,
		types.WithFieldDesc(base.FieldKeyUpdatedAt))
	opts := convertHostConditionsToOptions(conditions...)
	if results, num, err = s.daoHost.List(nCtx, page, opts...); err != nil {
		return nil, 0, err
	}

	return results, num, nil
}

// CountHost counts host by conditions.
func (s *Storage) CountHost(nCtx contextx.IContext, conditions ...*types.HostCondition) (num int64, err error) {
	// record metric.
	metric := s.metric().Start("count_host")
	defer metric.End(err)

	opts := convertHostConditionsToOptions(conditions...)
	if num, err = s.daoHost.Count(nCtx, opts...); err != nil {
		return 0, err
	}

	return num, nil
}

// DistinctHost distinct host fields.
// nolint:funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (s *Storage) DistinctHost(
	nCtx contextx.IContext, request types.HostDistinctRequest, conditions ...*types.HostCondition) (
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
			data.BizID, err = s.daoHost.DistinctBizID(nCtx, opts...)

			return err
		})
	}
	if request.NodeRole {
		gp.Go(func() error {
			var err error
			data.NodeRole, err = s.daoHost.DistinctNodeRole(nCtx, opts...)

			return err
		})
	}
	if request.NodeStatus {
		gp.Go(func() error {
			var err error
			data.NodeStatus, err = s.daoHost.DistinctNodeStatus(nCtx, opts...)

			return err
		})
	}
	if request.NodeVersion {
		gp.Go(func() error {
			var err error
			data.NodeVersion, err = s.daoHost.DistinctNodeVersion(nCtx, opts...)

			return err
		})
	}
	if request.DeptName {
		gp.Go(func() error {
			var err error
			data.DeptName, err = s.daoHost.DistinctDeptName(nCtx, opts...)

			return err
		})
	}
	if request.OSType {
		gp.Go(func() error {
			var err error
			data.OSType, err = s.daoHost.DistinctOSType(nCtx, opts...)

			return err
		})
	}
	if request.Arch {
		gp.Go(func() error {
			var err error
			data.Arch, err = s.daoHost.DistinctArch(nCtx, opts...)

			return err
		})
	}
	if request.Addressing {
		gp.Go(func() error {
			var err error
			data.Addressing, err = s.daoHost.DistinctAddressing(nCtx, opts...)

			return err
		})
	}
	if request.NetworkAreaID {
		gp.Go(func() error {
			var err error
			data.NetworkAreaID, err = s.daoHost.DistinctNetworkAreaID(nCtx, opts...)

			return err
		})
	}
	if request.NetworkUnitID {
		gp.Go(func() error {
			var err error
			data.NetworkUnitID, err = s.daoHost.DistinctNetworkUnitID(nCtx, opts...)

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

		if condition.StaticExactInclude != nil {
			opts = append(opts,
				host.WithHostID(condition.StaticExactInclude.HostID...),
				host.WithStaticBizID(condition.StaticExactInclude.BizID...),
				host.WithStaticNetworkAreaID(condition.StaticExactInclude.NetworkAreaID...),
				host.WithStaticAddressing(condition.StaticExactInclude.Addressing...),
				host.WithStaticInnerIPList(condition.StaticExactInclude.InnerIP...),
				host.WithStaticInnerIPV6List(condition.StaticExactInclude.InnerIPV6...),
			)
		}

		if condition.DynamicExactInclude != nil {
			opts = append(opts,
				host.WithDynamicNetworkUnitID(condition.DynamicExactInclude.NetworkUnitID...),
				host.WithDynamicNodeOsType(condition.DynamicExactInclude.OSType...),
				host.WithDynamicCPUArch(condition.DynamicExactInclude.Arch...),
				host.WithDynamicNodeRole(condition.DynamicExactInclude.NodeRole...),
				host.WithDynamicNodeStatus(condition.DynamicExactInclude.NodeStatus...),
				host.WithDynamicNodeVersion(condition.DynamicExactInclude.NodeVersion...),
				host.WithDynamicAgentID(condition.DynamicExactInclude.AgentID...),
				host.WithDynamicNodeGeneration(condition.DynamicExactInclude.NodeGeneration...),
			)
		}

		if condition.StaticExactExclude != nil {
			opts = append(opts,
				host.WithoutHostID(condition.StaticExactExclude.HostID...),
				host.WithoutStaticBizID(condition.StaticExactExclude.BizID...),
				host.WithoutStaticNetworkAreaID(condition.StaticExactExclude.NetworkAreaID...),
				host.WithoutStaticAddressing(condition.StaticExactExclude.Addressing...),
				host.WithoutStaticInnerIPList(condition.StaticExactExclude.InnerIP...),
			)
		}

		if condition.DynamicExactExclude != nil {
			opts = append(opts,
				host.WithoutDynamicNetworkUnitID(condition.DynamicExactExclude.NetworkUnitID...),
				host.WithoutDynamicNodeOsType(condition.DynamicExactExclude.OSType...),
				host.WithoutDynamicCPUArch(condition.DynamicExactExclude.Arch...),
				host.WithoutDynamicNodeRole(condition.DynamicExactExclude.NodeRole...),
				host.WithoutDynamicNodeStatus(condition.DynamicExactExclude.NodeStatus...),
				host.WithoutDynamicNodeVersion(condition.DynamicExactExclude.NodeVersion...),
				host.WithoutDynamicAgentID(condition.DynamicExactExclude.AgentID...),
				host.WithoutDynamicNodeGeneration(condition.DynamicExactExclude.NodeGeneration...),
			)
		}

		if condition.StaticFuzzyInclude != nil {
			opts = append(opts,
				host.WithFuzzyStaticHostName(condition.StaticFuzzyInclude.HostName...),
				host.WithFuzzyStaticDeptName(condition.StaticFuzzyInclude.DeptName...),
				host.WithFuzzyStaticInnerIPList(condition.StaticFuzzyInclude.InnerIP...),
				host.WithFuzzyStaticInnerIPV6List(condition.StaticFuzzyInclude.InnerIPV6...),
				host.WithFuzzyStaticOuterIPList(condition.StaticFuzzyInclude.OuterIP...),
				host.WithFuzzyStaticOuterIPV6List(condition.StaticFuzzyInclude.OuterIPV6...),
			)
		}

		if condition.StaticFuzzyExclude != nil {
			opts = append(opts,
				host.WithoutFuzzyStaticHostName(condition.StaticFuzzyExclude.HostName...),
				host.WithoutFuzzyStaticDeptName(condition.StaticFuzzyExclude.DeptName...),
				host.WithoutFuzzyStaticInnerIPList(condition.StaticFuzzyExclude.InnerIP...),
				host.WithoutFuzzyStaticInnerIPV6List(condition.StaticFuzzyExclude.InnerIPV6...),
				host.WithoutFuzzyStaticOuterIPList(condition.StaticFuzzyExclude.OuterIP...),
				host.WithoutFuzzyStaticOuterIPV6List(condition.StaticFuzzyExclude.OuterIPV6...),
			)
		}
	}

	return opts
}

// DeleteManyHost delete many hosts by hostIDs.
func (s *Storage) DeleteManyHost(nCtx contextx.IContext, hostIDs ...int64) (err error) {
	// record metric.
	metric := s.metric().Start("delete_many_host")
	defer metric.End(err)

	if nCtx == nil {
		return errors.New("nCtx is nil")
	}

	if len(hostIDs) == 0 {
		return nil
	}

	if err = s.daoHost.DeleteMany(nCtx, hostIDs...); err != nil {
		return err
	}

	return nil
}

// FindHostWithDynamic finds hosts with dynamic fields.
func (s *Storage) FindHostWithDynamic(nCtx contextx.IContext, page types.Page, conditions ...*types.HostCondition) (
	results []*types.Host, err error) {

	// record metric.
	metric := s.metric().Start("find_host_with_dynamic")
	defer metric.End(err)

	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	opts := convertHostConditionsToOptions(conditions...)
	if results, err = s.daoHost.FindWithDynamic(nCtx, page, opts...); err != nil {
		return nil, err
	}

	return results, nil
}

// UpdateHostDynamicFields updates the dynamic fields of a host.
func (s *Storage) UpdateHostDynamicFields(nCtx contextx.IContext, fields types.HostDynamicFields, hosts ...*types.Host) (err error) {
	// record metric.
	metric := s.metric().Start("update_host_dynamic_fields")
	defer metric.End(err)

	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if err = s.daoHost.UpdateDynamicFields(nCtx, fields, hosts...); err != nil {
		return fmt.Errorf("failed to update host dynamic fields: %w", err)
	}

	return nil
}

// getHostDistributionByNodeRole ...
func (s *Storage) getHostDistributionByNodeRole(nCtx contextx.IContext, conditions ...*types.HostCondition) (
	map[string]int64, error) {

	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	opts := convertHostConditionsToOptions(conditions...)
	hostDistributionByNodeRole, err := s.daoHost.GetHostDistributionByNodeRole(nCtx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get host distribution by node role: %w", err)
	}

	return hostDistributionByNodeRole, nil
}

// getHostDistributionByNetworkAreaID ...
func (s *Storage) getHostDistributionByNetworkAreaID(nCtx contextx.IContext, conditions ...*types.HostCondition) (
	map[int64]int64, error) {

	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	opts := convertHostConditionsToOptions(conditions...)
	hostDistributionByNetworkAreaID, err := s.daoHost.GetHostDistributionByNetworkAreaID(nCtx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get host distribution by node role: %w", err)
	}

	return hostDistributionByNetworkAreaID, nil
}

func (s *Storage) listHostWithFields(nCtx contextx.IContext, page types.Page,
	selection *types.HostFieldSelection, conditions ...*types.HostCondition) (
	[]*types.Host, int64, error) {

	if nCtx == nil {
		return nil, 0, basestorage.ErrNilContent()
	}

	opts := convertHostConditionsToOptions(conditions...)
	hosts, num, err := s.daoHost.ListWithFields(nCtx, page, selection, opts...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list host with fields: %w", err)
	}

	return hosts, num, nil
}

func (s *Storage) getRelayInfosInNetworkUnit(nCtx contextx.IContext, networkUnitID int64) (
	[]*types.RelayInfo, error) {

	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	results, err := s.daoHost.GetRelayInfosInNetworkUnit(nCtx, networkUnitID)
	if err != nil {
		return nil, err
	}

	return results, nil
}
