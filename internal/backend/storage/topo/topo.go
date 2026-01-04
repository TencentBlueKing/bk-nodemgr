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
	"fmt"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/accesspoint"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/business"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/networkarea"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/networkunit"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// UpsertManyBusiness updates or inserts many business.
func (s *Storage) UpsertManyBusiness(nCtx contextx.IContext, biz ...*types.Business) (err error) {
	// record metric.
	metric := s.metric().Start("upsert_many_business")
	defer metric.End(err)

	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if biz == nil {
		return basestorage.ErrUpsertNilData()
	}

	if err = s.daoBusiness.UpsertMany(nCtx, biz...); err != nil {
		return fmt.Errorf("failed to upsert business: %v", err)
	}

	return nil
}

// ListBusinesses lists businesses by page and conditions.
func (s *Storage) ListBusinesses(nCtx contextx.IContext, page types.Page, conditions ...*types.BusinessCondition) (
	[]*types.Business, int64, error) {

	var (
		results []*types.Business
		num     int64
		err     error
	)

	err = s.WrapFn(nCtx, "list_business", func(nCtx contextx.IContext) error {
		// record metric.
		metric := s.metric().Start("list_business")
		defer metric.End(err)

		opts := make([]business.OptFn, 0)
		for _, condition := range conditions {
			if condition == nil {
				continue
			}

			if condition.ExactInclude != nil {
				opts = append(opts,
					business.WithBizID(condition.ExactInclude.BizID...),
				)
			}

			if condition.ExactExclude != nil {
				opts = append(opts,
					business.WithoutBizID(condition.ExactExclude.BizID...),
				)
			}

			if condition.FuzzyInclude != nil {
				opts = append(opts,
					business.WithFuzzyBizName(condition.FuzzyInclude.BizName...),
				)
			}

			if condition.FuzzyExclude != nil {
				opts = append(opts,
					business.WithoutFuzzyBizName(condition.FuzzyExclude.BizName...),
				)
			}
		}

		if results, num, err = s.daoBusiness.List(nCtx, page, opts...); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	return results, num, nil
}

// ListNetworkArea lists networkarea by page and conditions.
// nolint: cyclop
func (s *Storage) ListNetworkArea(nCtx contextx.IContext, page types.Page, conditions ...*types.NetworkAreaCondition) (
	results []*types.NetworkArea, num int64, err error) {

	// record metric.
	metric := s.metric().Start("list_networkarea")
	defer metric.End(err)

	opts := make([]networkarea.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				networkarea.WithNetworkAreaID(condition.ExactInclude.NetworkAreaID...),
				networkarea.WithCloudVendor(condition.ExactInclude.CloudVendor...),
			)
		}

		if condition.ExactExclude != nil {
			opts = append(opts,
				networkarea.WithoutNetworkAreaID(condition.ExactExclude.NetworkAreaID...),
				networkarea.WithoutCloudVendor(condition.ExactExclude.CloudVendor...),
			)
		}

		if condition.FuzzyInclude != nil {
			opts = append(opts,
				networkarea.WithFuzzyNetworkAreaName(condition.FuzzyInclude.NetworkAreaName...),
			)
		}

		if condition.FuzzyExclude != nil {
			opts = append(opts,
				networkarea.WithoutFuzzyNetworkAreaName(condition.FuzzyExclude.NetworkAreaName...),
			)
		}
	}

	if results, num, err = s.daoNetworkArea.List(nCtx, page, opts...); err != nil {
		return nil, 0, err
	}

	return results, num, nil
}

// GetNetworkArea gets networkarea by id.
func (s *Storage) GetNetworkArea(nCtx contextx.IContext, networkAreaID int64) (data *types.NetworkArea, err error) {
	// record metric.
	metric := s.metric().Start("get_networkarea")
	defer metric.End(err)

	if data, err = s.daoNetworkArea.Get(nCtx, networkAreaID); err != nil {
		return nil, err
	}

	return data, nil
}

// UpsertManyNetworkArea updates or inserts networkarea.
func (s *Storage) UpsertManyNetworkArea(nCtx contextx.IContext, networkAreas ...*types.NetworkArea) (err error) {
	// record metric.
	metric := s.metric().Start("upsert_many_networkarea")
	defer metric.End(err)

	if err = s.daoNetworkArea.UpsertMany(nCtx, networkAreas...); err != nil {
		return err
	}

	return nil
}

// UpdateManyNetworkArea updates networkarea.
func (s *Storage) UpdateManyNetworkArea(nCtx contextx.IContext, networkArea ...*types.NetworkArea) (err error) {
	// record metric.
	metric := s.metric().Start("update_many_networkarea")
	defer metric.End(err)

	if err = s.daoNetworkArea.UpdateMany(nCtx, networkArea...); err != nil {
		return err
	}

	return nil
}

// DeleteManyNetworkArea deletes networkarea.
func (s *Storage) DeleteManyNetworkArea(nCtx contextx.IContext, networkAreaIDs ...int64) (err error) {
	// record metric.
	metric := s.metric().Start("delete_many_networkarea")
	defer metric.End(err)

	if err = s.daoNetworkArea.DeleteMany(nCtx, networkAreaIDs...); err != nil {
		return err
	}

	return nil
}

// ListNetworkUnit lists networkunit.
func (s *Storage) ListNetworkUnit(nCtx contextx.IContext, page types.Page, conditions ...*types.NetworkUnitCondition) (
	results []*types.NetworkUnit, num int64, err error) {

	// record metric.
	metric := s.metric().Start("list_networkunit")
	defer metric.End(err)

	opts := make([]networkunit.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				networkunit.WithNetworkUnitID(condition.ExactInclude.NetworkUnitID...),
				networkunit.WithNetworkAreaID(condition.ExactInclude.NetworkAreaID...),
				networkunit.WithIsDirect(condition.ExactInclude.IsDirect...),
			)
		}

		if condition.ExactExclude != nil {
			opts = append(opts,
				networkunit.WithoutNetworkUnitID(condition.ExactExclude.NetworkUnitID...),
				networkunit.WithoutNetworkAreaID(condition.ExactExclude.NetworkAreaID...),
				networkunit.WithoutIsDirect(condition.ExactExclude.IsDirect...),
			)
		}
	}

	if results, num, err = s.daoNetworkUnit.List(nCtx, page, opts...); err != nil {
		return nil, 0, err
	}

	return results, num, nil
}

// GetNetworkUnit gets networkunit by id.
func (s *Storage) GetNetworkUnit(nCtx contextx.IContext, networkUnitID int64) (data *types.NetworkUnit, err error) {
	// record metric.
	metric := s.metric().Start("get_networkunit")
	defer metric.End(err)

	if data, err = s.daoNetworkUnit.Get(nCtx, networkUnitID); err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Storage) checkNetworkUnitLinks(nCtx contextx.IContext, networkUnit *types.NetworkUnit) error {
	upstreamNetworkUnitIDs := make([]int64, 0)
	if networkUnit.Links.Cluster != nil {
		upstreamNetworkUnitIDs = append(upstreamNetworkUnitIDs, networkUnit.Links.Cluster.NetworkUnitID)
	}
	if networkUnit.Links.File != nil {
		upstreamNetworkUnitIDs = append(upstreamNetworkUnitIDs, networkUnit.Links.File.NetworkUnitID)
	}
	if networkUnit.Links.Data != nil {
		upstreamNetworkUnitIDs = append(upstreamNetworkUnitIDs, networkUnit.Links.Data.NetworkUnitID)
	}

	upstreamNetworkUnits, _, err := s.daoNetworkUnit.List(
		nCtx,
		types.UnlimitedPage(),
		networkunit.WithNetworkUnitID(upstreamNetworkUnitIDs...))
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to check networkunit links, failed to list upstream networkunit")

		return err
	}

	if err := findUpstreamNetworkUnitWithLink(upstreamNetworkUnits, networkUnit.Links.Cluster); err != nil {
		return err
	}

	if err := findUpstreamNetworkUnitWithLink(upstreamNetworkUnits, networkUnit.Links.File); err != nil {
		return err
	}

	if err := findUpstreamNetworkUnitWithLink(upstreamNetworkUnits, networkUnit.Links.Data); err != nil {
		return err
	}

	return nil
}

func findUpstreamNetworkUnitWithLink(upstreamNetworkUnits []*types.NetworkUnit, link *types.Link) error {
	if link == nil {
		return nil
	}

	// find networkunit.
	for _, upstreamNetworkUnit := range upstreamNetworkUnits {
		if upstreamNetworkUnit.ID == link.NetworkUnitID {
			// check networkarea.
			if upstreamNetworkUnit.NetworkAreaID != link.NetworkAreaID {
				return fmt.Errorf("upstream link networkarea not matched. "+
					"networkarea-id(%d), networkunit-id(%d), accesspoint-id(%d)",
					link.NetworkAreaID, link.NetworkUnitID, link.AccessPointID)
			}

			// check accesspoint.
			for _, apID := range upstreamNetworkUnit.AccessPoints {
				if apID == link.AccessPointID {
					return nil
				}
			}

			return fmt.Errorf("upstream link accesspoint not found. networkarea-id(%d), networkunit-id(%d), accesspoint-id(%d)",
				link.NetworkAreaID, link.NetworkUnitID, link.AccessPointID)
		}
	}

	return fmt.Errorf("upstream link networkunit not found. networkarea-id(%d), networkunit-id(%d), accesspoint-id(%d)",
		link.NetworkAreaID, link.NetworkUnitID, link.AccessPointID)
}

// CreateNetworkUnit creates networkunit.
func (s *Storage) CreateNetworkUnit(nCtx contextx.IContext, networkUnit *types.NetworkUnit, accessPoints ...*types.AccessPoint) (
	networkUnitID int64, data *AccessPointResult, err error) {

	// record metric.
	metric := s.metric().Start("create_networkunit")
	defer metric.End(err)

	if !networkUnit.IsDirect {
		if err = s.checkNetworkUnitLinks(nCtx, networkUnit); err != nil {
			return -1, nil, err
		}
	}

	if len(accessPoints) == 0 {
		networkUnit.AccessPoints = nil

		networkUnitID, err = s.daoNetworkUnit.Create(nCtx, networkUnit)
		if err != nil {
			return -1, nil, err
		}

		return networkUnitID, &AccessPointResult{}, nil
	}

	// create accesspoints first.
	var accessPointIDs []int64
	accessPointIDs, err = s.daoAccessPoint.CreateMany(nCtx, accessPoints...)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to create networkunit, failed to create accesspoint")

		return -1, nil, err
	}

	networkUnit.AccessPoints = accessPointIDs
	for idx, accessPointID := range accessPointIDs {
		if idx > len(accessPointIDs) {
			break
		}

		accessPoints[idx].ID = accessPointID
	}

	// create networkunit.
	if networkUnitID, err = s.daoNetworkUnit.Create(nCtx, networkUnit); err != nil {
		return -1, nil, err
	}

	return networkUnitID, &AccessPointResult{
		Created: accessPoints,
	}, nil
}

// UpdateNetworkUnit updates networkunit.
func (s *Storage) UpdateNetworkUnit(nCtx contextx.IContext, networkUnit *types.NetworkUnit, accessPoints ...*types.AccessPoint) (
	data *AccessPointResult, err error) {

	// record metric.
	metric := s.metric().Start("update_networkunit")
	defer metric.End(err)

	if !networkUnit.IsDirect {
		if err := s.checkNetworkUnitLinks(nCtx, networkUnit); err != nil {
			return nil, err
		}
	}

	if len(accessPoints) == 0 {
		networkUnit.AccessPoints = nil

		if err = s.daoNetworkUnit.UpdateMany(nCtx, networkUnit); err != nil {
			return nil, err
		}

		return &AccessPointResult{}, nil
	}

	// update old accesspoints, create new accesspoints.
	accessPointIDs := make([]int64, 0)
	oldAccessPoints := make([]*types.AccessPoint, 0)
	newAccessPoints := make([]*types.AccessPoint, 0)
	for _, accessPoint := range accessPoints {
		if accessPoint.ID >= 0 {
			accessPointIDs = append(accessPointIDs, accessPoint.ID)
			oldAccessPoints = append(oldAccessPoints, accessPoint)

			continue
		}

		newAccessPoints = append(newAccessPoints, accessPoint)
	}

	if len(oldAccessPoints) > 0 {
		if err = s.daoAccessPoint.UpdateMany(nCtx, oldAccessPoints...); err != nil {
			logger.G.Sys().WithErr(err).Error("failed to update networkunit, failed to update accesspoint")

			return nil, err
		}
	}
	if len(newAccessPoints) > 0 {
		createdAccessPointIDs, err := s.daoAccessPoint.CreateMany(nCtx, newAccessPoints...)
		if err != nil {
			logger.G.Sys().WithErr(err).Error("failed to update networkunit, failed to create accesspoint")

			return nil, err
		}
		accessPointIDs = append(accessPointIDs, createdAccessPointIDs...)

		for idx, accessPointID := range createdAccessPointIDs {
			newAccessPoints[idx].ID = accessPointID
		}
	}
	networkUnit.AccessPoints = accessPointIDs

	if err = s.daoNetworkUnit.UpdateMany(nCtx, networkUnit); err != nil {
		return nil, err
	}

	return &AccessPointResult{
		Created: newAccessPoints,
		Updated: oldAccessPoints,
		Deleted: nil,
	}, nil
}

// DeleteManyNetworkUnit deletes networkunit.
func (s *Storage) DeleteManyNetworkUnit(nCtx contextx.IContext, networkUnitIDs ...int64) (err error) {
	// record metric.
	metric := s.metric().Start("delete_networkunit")
	defer metric.End(err)

	if err = s.daoNetworkUnit.DeleteMany(nCtx, networkUnitIDs...); err != nil {
		return err
	}

	return nil
}

// CountAccessPoint counts accesspoint.
func (s *Storage) CountAccessPoint(nCtx contextx.IContext, conditions ...*types.AccessPointCondition) (num int64, err error) {
	// record metric.
	metric := s.metric().Start("count_accesspoint")
	defer metric.End(err)

	opts := make([]accesspoint.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				accesspoint.WithAccessPointID(condition.ExactInclude.AccessPointID...),
				accesspoint.WithNetworkAreaID(condition.ExactInclude.NetworkAreaID...),
			)
		}

		if condition.ExactExclude != nil {
			opts = append(opts,
				accesspoint.WithoutAccessPointID(condition.ExactExclude.AccessPointID...),
				accesspoint.WithoutNetworkAreaID(condition.ExactExclude.NetworkAreaID...),
			)
		}
	}

	if num, err = s.daoAccessPoint.Count(nCtx, opts...); err != nil {
		return 0, err
	}

	return num, nil
}

// ListAccessPoint lists accesspoint.
func (s *Storage) ListAccessPoint(nCtx contextx.IContext, page types.Page, conditions ...*types.AccessPointCondition) (
	results []*types.AccessPoint, num int64, err error) {

	// record metric.
	metric := s.metric().Start("list_accesspoint")
	defer metric.End(err)

	opts := make([]accesspoint.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				accesspoint.WithAccessPointID(condition.ExactInclude.AccessPointID...),
				accesspoint.WithNetworkAreaID(condition.ExactInclude.NetworkAreaID...),
			)
		}

		if condition.ExactExclude != nil {
			opts = append(opts,
				accesspoint.WithoutAccessPointID(condition.ExactExclude.AccessPointID...),
				accesspoint.WithoutNetworkAreaID(condition.ExactExclude.NetworkAreaID...),
			)
		}
	}

	if results, num, err = s.daoAccessPoint.List(nCtx, page, opts...); err != nil {
		return nil, 0, err
	}

	return results, num, nil
}

// AccessPointResult describes the accesspoint result in networkunit handlers.
type AccessPointResult struct {
	Created []*types.AccessPoint
	Updated []*types.AccessPoint
	Deleted []*types.AccessPoint
}

// GetHostDistributionByNodeRole ...
func (s *Storage) GetHostDistributionByNodeRole(nCtx contextx.IContext, conditions ...*types.HostCondition) (
	hostDistributionByNodeRole map[string]int64, err error) {

	// record metric.
	metric := s.metric().Start("get_host_distribution_by_node_role")
	defer metric.End(err)

	hostDistributionByNodeRole, err = s.getHostDistributionByNodeRole(nCtx, conditions...)

	return hostDistributionByNodeRole, err
}

// GetHostDistributionByNetworkAreaID ...
func (s *Storage) GetHostDistributionByNetworkAreaID(nCtx contextx.IContext, conditions ...*types.HostCondition) (
	hostDistributionByNetworkAreaID map[int64]int64, err error) {

	// record metric.
	metric := s.metric().Start("get_host_distribution_by_networkarea_id")
	defer metric.End(err)

	hostDistributionByNetworkAreaID, err = s.getHostDistributionByNetworkAreaID(nCtx, conditions...)

	return hostDistributionByNetworkAreaID, err
}
