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
package topo

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/host"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// GetHostByID get host by id.
func (s *Storage) GetHostByID(nCtx contextx.IContext, hostID int64) (*types.Host, error) {
	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if hostID < 0 {
		return nil, errors.New("host id should be equal or greater than 0")
	}

	var data *types.Host
	err := s.WrapFn(nCtx, metricOperationGetHostByID, func(nCtx contextx.IContext) error {
		hosts, count, err := s.daoHost.List(nCtx, types.Page{Limit: 1}, host.WithHostID(hostID))
		if err != nil {
			return fmt.Errorf("failed to get host by id, host-id(%d): %w", hostID, err)
		}

		if count != 1 {
			return fmt.Errorf("failed to get host by id, result count is not 1, host-id(%d), count(%d)",
				hostID, count)
		}

		data = hosts[0]

		return nil
	})

	return data, err
}

// UpsertManyHost upserts many hosts.
func (s *Storage) UpsertManyHost(nCtx contextx.IContext, hosts ...*types.Host) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if len(hosts) == 0 {
		return nil
	}

	return s.WrapFn(nCtx, metricOperationUpsertManyHost, func(nCtx contextx.IContext) error {
		var err error
		err = s.daoHost.UpsertMany(nCtx, hosts...)
		if err != nil {
			return fmt.Errorf("failed to upsert hosts: %w", err)
		}

		return nil
	})
}

// UpsertManyHostStatic updates or inserts host statics.
func (s *Storage) UpsertManyHostStatic(nCtx contextx.IContext, hosts ...*types.Host) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if len(hosts) == 0 {
		return nil
	}

	return s.WrapFn(nCtx, metricOperationUpsertManyHostStatic, func(nCtx contextx.IContext) error {
		var err error
		err = s.daoHost.UpsertStaticMany(nCtx, hosts...)
		if err != nil {
			return fmt.Errorf("failed to upsert host statics: %w", err)
		}

		return nil
	})
}

// UpdateManyHostDynamic updates host dynamics.
func (s *Storage) UpdateManyHostDynamic(nCtx contextx.IContext, hosts ...*types.Host) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if len(hosts) == 0 {
		return nil
	}

	return s.WrapFn(nCtx, metricOperationUpdateManyHostDynamic, func(nCtx contextx.IContext) error {
		var err error
		err = s.daoHost.UpdateDynamicMany(nCtx, hosts...)
		if err != nil {
			return fmt.Errorf("failed to upsert host dynamics: %w", err)
		}

		return nil
	})
}

// ListHost lists hosts by page and conditions.
func (s *Storage) ListHost(nCtx contextx.IContext, page types.Page, conditions ...*types.HostCondition) (
	[]*types.Host, int64, error) {

	var (
		results []*types.Host
		num     int64
	)
	err := s.WrapFn(nCtx, metricOperationListHost, func(nCtx contextx.IContext) error {
		opts := convertHostConditionsToOptions(conditions...)
		var err error
		results, num, err = s.daoHost.List(nCtx, page, opts...)

		return err
	})

	return results, num, err
}

// ListHostOrderByUpdateTime lists hosts by page and conditions, and sort by
// business operation time first, then by general update time as fallback.
func (s *Storage) ListHostOrderByUpdateTime(
	nCtx contextx.IContext, page types.Page, conditions ...*types.HostCondition) ([]*types.Host, int64, error) {

	var (
		results []*types.Host
		num     int64
	)
	err := s.WrapFn(nCtx, metricOperationListHostOrderByUpdateTime, func(nCtx contextx.IContext) error {
		page.Sort = types.WithSortFields(page.Sort,
			types.WithFieldDesc(host.FieldKeyOperationUpdatedAt),
			types.WithFieldDesc(base.FieldKeyUpdatedAt))
		opts := convertHostConditionsToOptions(conditions...)
		var err error
		results, num, err = s.daoHost.List(nCtx, page, opts...)

		return err
	})

	return results, num, err
}

// CountHost counts host by conditions.
func (s *Storage) CountHost(nCtx contextx.IContext, conditions ...*types.HostCondition) (int64, error) {
	var num int64
	err := s.WrapFn(nCtx, metricOperationCountHost, func(nCtx contextx.IContext) error {
		opts := convertHostConditionsToOptions(conditions...)
		var err error
		num, err = s.daoHost.Count(nCtx, opts...)

		return err
	})

	return num, err
}

// DistinctHost distinct host fields.
// nolint:funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (s *Storage) DistinctHost(
	nCtx contextx.IContext, request types.HostDistinctRequest, conditions ...*types.HostCondition) (
	*types.HostDistinctResult, error) {

	var data *types.HostDistinctResult
	err := s.WrapFn(nCtx, metricOperationDistinctHost, func(nCtx contextx.IContext) error {
		opts := convertHostConditionsToOptions(conditions...)
		data = new(types.HostDistinctResult)

		gp := gopool.NewPool()
		s.enqueueHostDistinctTasks(gp, s.buildHostDistinctTasks(nCtx, request, data, opts))

		return gp.Wait()
	})

	return data, err
}

type hostDistinctTask struct {
	enabled bool
	run     func() error
}

func (s *Storage) buildHostDistinctTasks(
	nCtx contextx.IContext, request types.HostDistinctRequest, data *types.HostDistinctResult, opts []host.OptFn,
) []hostDistinctTask {

	tasks := []hostDistinctTask{
		{
			enabled: request.BizID,
			run: func() error {
				var err error
				data.BizID, err = s.daoHost.DistinctBizID(nCtx, opts...)

				return err
			},
		},
		{
			enabled: request.NodeRole,
			run: func() error {
				var err error
				data.NodeRole, err = s.daoHost.DistinctNodeRole(nCtx, opts...)

				return err
			},
		},
		{
			enabled: request.NodeStatus,
			run: func() error {
				var err error
				data.NodeStatus, err = s.daoHost.DistinctNodeStatus(nCtx, opts...)

				return err
			},
		},
		{
			enabled: request.NodeVersion,
			run: func() error {
				var err error
				data.NodeVersion, err = s.daoHost.DistinctNodeVersion(nCtx, opts...)

				return err
			},
		},
		{
			enabled: request.DeptName,
			run: func() error {
				var err error
				data.DeptName, err = s.daoHost.DistinctDeptName(nCtx, opts...)

				return err
			},
		},
		{
			enabled: request.OSType,
			run: func() error {
				var err error
				data.OSType, err = s.daoHost.DistinctOSType(nCtx, opts...)

				return err
			},
		},
		{
			enabled: request.Arch,
			run: func() error {
				var err error
				data.Arch, err = s.daoHost.DistinctArch(nCtx, opts...)

				return err
			},
		},
		{
			enabled: request.Addressing,
			run: func() error {
				var err error
				data.Addressing, err = s.daoHost.DistinctAddressing(nCtx, opts...)

				return err
			},
		},
		{
			enabled: request.NetworkAreaID,
			run: func() error {
				var err error
				data.NetworkAreaID, err = s.daoHost.DistinctNetworkAreaID(nCtx, opts...)

				return err
			},
		},
		{
			enabled: request.NetworkUnitID,
			run: func() error {
				var err error
				data.NetworkUnitID, err = s.daoHost.DistinctNetworkUnitID(nCtx, opts...)

				return err
			},
		},
	}

	return tasks
}

func (s *Storage) enqueueHostDistinctTasks(gp gopool.Pool, tasks []hostDistinctTask) {
	for _, task := range tasks {
		if !task.enabled {
			continue
		}

		gp.Go(task.run)
	}
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
func (s *Storage) DeleteManyHost(nCtx contextx.IContext, hostIDs ...int64) error {
	if nCtx == nil {
		return errors.New("nCtx is nil")
	}

	if len(hostIDs) == 0 {
		return nil
	}

	return s.WrapFn(nCtx, metricOperationDeleteManyHost, func(nCtx contextx.IContext) error {
		var err error
		err = s.daoHost.DeleteMany(nCtx, hostIDs...)

		return err
	})
}

// FindHostWithDynamic finds hosts with dynamic fields.
func (s *Storage) FindHostWithDynamic(nCtx contextx.IContext, page types.Page, conditions ...*types.HostCondition) (
	[]*types.Host, error) {

	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	var results []*types.Host
	err := s.WrapFn(nCtx, metricOperationFindHostWithDynamic, func(nCtx contextx.IContext) error {
		opts := convertHostConditionsToOptions(conditions...)
		var err error
		results, err = s.daoHost.FindWithDynamic(nCtx, page, opts...)

		return err
	})

	return results, err
}

// TouchHostOperationTime marks the given hosts as recently operated by a user
// or API action. This updates the dedicated business-operation timestamp used
// to sort the host list, without affecting the general updated_at field.
func (s *Storage) TouchHostOperationTime(nCtx contextx.IContext, hostIDs ...int64) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if len(hostIDs) == 0 {
		return nil
	}

	return s.WrapFn(nCtx, metricOperationTouchHostOperationTime, func(nCtx contextx.IContext) error {
		return s.daoHost.TouchOperationUpdatedAt(nCtx, hostIDs...)
	})
}

// UpdateHostDynamicFields updates the dynamic fields of a host.
func (s *Storage) UpdateHostDynamicFields(nCtx contextx.IContext, fields types.HostDynamicFields, hosts ...*types.Host) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	return s.WrapFn(nCtx, metricOperationUpdateHostDynamicFields, func(nCtx contextx.IContext) error {
		var err error
		err = s.daoHost.UpdateDynamicFields(nCtx, fields, hosts...)
		if err != nil {
			return fmt.Errorf("failed to update host dynamic fields: %w", err)
		}

		return nil
	})
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

// getHostDistributionByNodeVersion ...
func (s *Storage) getHostDistributionByNodeVersion(nCtx contextx.IContext, conditions ...*types.HostCondition) (
	map[string]int64, error) {

	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	opts := convertHostConditionsToOptions(conditions...)
	hostDistributionByNodeVersion, err := s.daoHost.GetHostDistributionByNodeVersion(nCtx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get host distribution by node version: %w", err)
	}

	return hostDistributionByNodeVersion, nil
}

// getNetworkUnitDistributionByNetworkAreaID ...
func (s *Storage) getNetworkUnitDistributionByNetworkAreaID(nCtx contextx.IContext, conditions ...*types.NetworkUnitCondition) (
	map[int64]int64, error) {

	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	opts := convertNetworkUnitConditionsToOptions(conditions...)
	networkUnitDistributionByNetworkAreaID, err := s.daoNetworkUnit.GetNetworkUnitDistributionByNetworkAreaID(nCtx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get networkunit distribution by network area id: %w", err)
	}

	return networkUnitDistributionByNetworkAreaID, nil
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

	return results, err
}

func (s *Storage) getHostBizMapping(nCtx contextx.IContext, hostIDs []int64) (map[int64]int64, error) {
	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	hostIDs = conv.SliceUnique(hostIDs)
	selection := &types.HostFieldSelection{
		HostID: true,
		BizID:  true,
	}
	results, _, err := s.daoHost.ListWithFields(nCtx, types.UnlimitedPage(), selection, host.WithHostID(hostIDs...))
	if err != nil {
		return nil, fmt.Errorf("failed to get host biz mapping: %w", err)
	}

	hostBizMapping := make(map[int64]int64, len(results))
	for _, host := range results {
		if host.HostID == 0 {
			return nil, fmt.Errorf("invalid host id in host biz mapping result: %d", host.HostID)
		}

		if _, ok := hostBizMapping[host.HostID]; ok {
			return nil, fmt.Errorf("duplicate host id in host biz mapping result: %d", host.HostID)
		}

		hostBizMapping[host.HostID] = host.Static.BizID
	}

	for _, id := range hostIDs {
		if _, ok := hostBizMapping[id]; !ok {
			return nil, fmt.Errorf("host id %d not found in host biz mapping result", id)
		}
	}

	return hostBizMapping, nil
}
