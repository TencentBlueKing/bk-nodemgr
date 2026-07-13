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
	"slices"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/accesspoint"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/business"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/networkarea"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/networkunit"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/globalsettings"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// UpsertManyBusiness updates or inserts many business.
func (s *Storage) UpsertManyBusiness(nCtx contextx.IContext, biz ...*types.Business) error {
	if nCtx == nil {
		return basestorage.ErrNilContent()
	}

	if biz == nil {
		return basestorage.ErrUpsertNilData()
	}

	return s.WrapFn(nCtx, metricOperationUpsertManyBusiness, func(nCtx contextx.IContext) error {
		var err error
		err = s.daoBusiness.UpsertMany(nCtx, biz...)
		if err != nil {
			return fmt.Errorf("failed to upsert business: %w", err)
		}

		return nil
	})
}

// ListBusinesses lists businesses by page and conditions.
func (s *Storage) ListBusinesses(nCtx contextx.IContext, page types.Page, conditions ...*types.BusinessCondition) (
	[]*types.Business, int64, error) {

	var (
		results []*types.Business
		num     int64
		err     error
	)

	err = s.WrapFn(nCtx, metricOperationListBusiness, func(nCtx contextx.IContext) error {
		var err error
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

		results, num, err = s.daoBusiness.List(nCtx, page, opts...)

		return err
	})

	return results, num, err
}

// ListBusinessesWithoutCount lists businesses by page and conditions without count.
func (s *Storage) ListBusinessesWithoutCount(
	nCtx contextx.IContext, page types.Page, conditions ...*types.BusinessCondition) (
	[]*types.Business, error) {

	var results []*types.Business

	err := s.WrapFn(nCtx, metricOperationListBusinessWithoutCount, func(nCtx contextx.IContext) error {
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

		var err error
		results, err = s.daoBusiness.ListWithoutCount(nCtx, page, opts...)

		return err
	})

	return results, err
}

// ListNetworkArea lists networkarea by page and conditions.
// nolint: cyclop
func (s *Storage) ListNetworkArea(nCtx contextx.IContext, page types.Page, conditions ...*types.NetworkAreaCondition) (
	[]*types.NetworkArea, int64, error) {

	var (
		results []*types.NetworkArea
		num     int64
		err     error
	)

	err = s.WrapFn(nCtx, metricOperationListNetworkArea, func(nCtx contextx.IContext) error {
		var err error
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

		results, num, err = s.daoNetworkArea.List(nCtx, page, opts...)

		return err
	})

	return results, num, err
}

// ListNetworkAreaWithoutCount lists networkarea by page and conditions without count.
// nolint: cyclop
func (s *Storage) ListNetworkAreaWithoutCount(
	nCtx contextx.IContext, page types.Page, conditions ...*types.NetworkAreaCondition) (
	[]*types.NetworkArea, error) {

	var results []*types.NetworkArea

	err := s.WrapFn(nCtx, metricOperationListNetworkAreaWithoutCount, func(nCtx contextx.IContext) error {
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

		var err error
		results, err = s.daoNetworkArea.ListWithoutCount(nCtx, page, opts...)

		return err
	})

	return results, err
}

// GetNetworkArea gets networkarea by id.
func (s *Storage) GetNetworkArea(nCtx contextx.IContext, networkAreaID int64) (*types.NetworkArea, error) {
	var (
		data *types.NetworkArea
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetNetworkArea, func(nCtx contextx.IContext) error {
		var err error
		data, err = s.daoNetworkArea.Get(nCtx, networkAreaID)

		return err
	})

	return data, err
}

// UpsertManyNetworkArea updates or inserts networkarea.
func (s *Storage) UpsertManyNetworkArea(nCtx contextx.IContext, networkAreas ...*types.NetworkArea) error {
	return s.WrapFn(nCtx, metricOperationUpsertManyNetworkArea, func(nCtx contextx.IContext) error {
		var err error
		err = s.daoNetworkArea.UpsertMany(nCtx, networkAreas...)

		return err
	})
}

// UpdateManyNetworkArea updates networkarea.
func (s *Storage) UpdateManyNetworkArea(nCtx contextx.IContext, networkArea ...*types.NetworkArea) error {
	return s.WrapFn(nCtx, metricOperationUpdateManyNetworkArea, func(nCtx contextx.IContext) error {
		var err error
		err = s.daoNetworkArea.UpdateMany(nCtx, networkArea...)

		return err
	})
}

// DeleteManyNetworkArea deletes networkarea.
func (s *Storage) DeleteManyNetworkArea(nCtx contextx.IContext, networkAreaIDs ...int64) error {
	return s.WrapFn(nCtx, metricOperationDeleteManyNetworkArea, func(nCtx contextx.IContext) error {
		var err error
		err = s.daoNetworkArea.DeleteMany(nCtx, networkAreaIDs...)

		return err
	})
}

// ListNetworkUnit lists networkunit.
func (s *Storage) ListNetworkUnit(nCtx contextx.IContext, page types.Page, conditions ...*types.NetworkUnitCondition) (
	[]*types.NetworkUnit, int64, error) {

	var (
		results []*types.NetworkUnit
		num     int64
		err     error
	)

	err = s.WrapFn(nCtx, metricOperationListNetworkUnit, func(nCtx contextx.IContext) error {
		var err error
		opts := convertNetworkUnitConditionsToOptions(conditions...)

		results, num, err = s.daoNetworkUnit.List(nCtx, page, opts...)

		return err
	})

	return results, num, err
}

// ListNetworkUnitWithoutCount lists networkunit by page and conditions without count.
func (s *Storage) ListNetworkUnitWithoutCount(
	nCtx contextx.IContext, page types.Page, conditions ...*types.NetworkUnitCondition) (
	[]*types.NetworkUnit, error) {

	var results []*types.NetworkUnit

	err := s.WrapFn(nCtx, metricOperationListNetworkUnitWithoutCount, func(nCtx contextx.IContext) error {
		opts := convertNetworkUnitConditionsToOptions(conditions...)

		var err error
		results, err = s.daoNetworkUnit.ListWithoutCount(nCtx, page, opts...)

		return err
	})

	return results, err
}

// GetNetworkUnitDistributionByNetworkAreaID get networkunit distribution by network area id.
func (s *Storage) GetNetworkUnitDistributionByNetworkAreaID(nCtx contextx.IContext, conditions ...*types.NetworkUnitCondition) (
	map[int64]int64, error) {

	var (
		networkUnitDistributionByNetworkAreaID map[int64]int64
		err                                    error
	)

	err = s.WrapFn(nCtx, metricOperationGetNetworkUnitDistributionByNetworkAreaID, func(nCtx contextx.IContext) error {
		var err error
		networkUnitDistributionByNetworkAreaID, err = s.getNetworkUnitDistributionByNetworkAreaID(nCtx, conditions...)

		return err
	})

	return networkUnitDistributionByNetworkAreaID, err
}

// RecommendNetworkUnitByNetworkSegment recommends network units based on network segment rules.
func (s *Storage) RecommendNetworkUnitByNetworkSegment(
	nCtx contextx.IContext,
	items ...*types.NetworkUnitSegmentRecommendationItem,
) ([]*types.NetworkUnitSegmentRecommendationResult, error) {

	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if len(items) == 0 {
		return nil, nil
	}

	var results []*types.NetworkUnitSegmentRecommendationResult
	err := s.WrapFn(nCtx, metricOperationRecommendNetworkUnitByNetworkSegment, func(nCtx contextx.IContext) error {
		var err error
		results, err = s.recommendNetworkUnitByNetworkSegment(nCtx, items...)

		return err
	})

	return results, err
}

// GetNetworkUnit gets networkunit by id.
func (s *Storage) GetNetworkUnit(nCtx contextx.IContext, networkUnitID int64) (*types.NetworkUnit, error) {
	var (
		data *types.NetworkUnit
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationGetNetworkUnit, func(nCtx contextx.IContext) error {
		var err error
		data, err = s.daoNetworkUnit.Get(nCtx, networkUnitID)

		return err
	})

	return data, err
}

func (s *Storage) loadNetworkUnitSegmentRuleConfig(nCtx contextx.IContext) (types.NetworkUnitSegmentRuleConfig, error) {
	setting, err := s.daoGlobalSettings.Get(nCtx, globalsettings.NetworkUnitSegmentRules)
	if err != nil {
		if errors.Is(err, base.ErrRecordNoFound()) {
			return types.NetworkUnitSegmentRuleConfig{}, nil
		}

		return nil, err
	}

	if setting == nil || setting.Value == "" {
		return types.NetworkUnitSegmentRuleConfig{}, nil
	}

	return parseNetworkUnitSegmentRuleConfig(setting.Value)
}

func (s *Storage) recommendNetworkUnitByNetworkSegment(
	nCtx contextx.IContext,
	items ...*types.NetworkUnitSegmentRecommendationItem,
) ([]*types.NetworkUnitSegmentRecommendationResult, error) {

	rules, err := s.loadNetworkUnitSegmentRuleConfig(nCtx)
	if err != nil {
		results := make([]*types.NetworkUnitSegmentRecommendationResult, len(items))
		for idx, item := range items {
			results[idx] = &types.NetworkUnitSegmentRecommendationResult{NetworkUnitID: -1}
			if item != nil {
				results[idx].NetworkAreaID = item.NetworkAreaID
				results[idx].IP = item.IP
			}
			results[idx].Message = fmt.Sprintf("invalid rule config: %v", err)
		}

		return results, nil
	}

	areaIDs := make(map[int64]struct{})
	for _, item := range items {
		if item == nil {
			continue
		}
		areaIDs[item.NetworkAreaID] = struct{}{}
	}

	conditions := make([]*types.NetworkUnitCondition, 0, 1)
	if len(areaIDs) > 0 {
		networkAreaIDs := make([]int64, 0, len(areaIDs))
		for areaID := range areaIDs {
			networkAreaIDs = append(networkAreaIDs, areaID)
		}
		conditions = append(conditions, &types.NetworkUnitCondition{
			ExactInclude: &types.NetworkUnitExactFields{NetworkAreaID: networkAreaIDs},
		})
	}

	networkUnits, err := s.ListNetworkUnitWithoutCount(nCtx, types.UnlimitedPage(), conditions...)
	if err != nil {
		return nil, err
	}

	networkUnitAreaMap := make(map[int64]int64, len(networkUnits))
	for _, networkUnit := range networkUnits {
		if networkUnit == nil {
			continue
		}
		networkUnitAreaMap[networkUnit.ID] = networkUnit.NetworkAreaID
	}

	return recommendNetworkUnitsBySegment(rules, items, networkUnitAreaMap)
}

// GetNetworkUnitIDsByAccessPoints returns NetworkUnit IDs that contain the given AccessPoint IDs.
func (s *Storage) GetNetworkUnitIDsByAccessPoints(nCtx contextx.IContext, accessPointIDs []int64) ([]int64, error) {
	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if len(accessPointIDs) == 0 {
		return nil, nil
	}

	var networkUnitIDs []int64

	err := s.WrapFn(nCtx, metricOperationGetNetworkUnitIDsByAccessPoints, func(nCtx contextx.IContext) error {
		// Query all NetworkUnits and filter those containing the requested AccessPoint IDs.
		networkUnits, err := s.daoNetworkUnit.ListWithoutCount(nCtx, types.UnlimitedPage())
		if err != nil {
			return fmt.Errorf("list networkunits failed: %w", err)
		}

		// Build a set of requested AccessPoint IDs for fast lookup.
		requestedSet := make(map[int64]struct{}, len(accessPointIDs))
		for _, id := range accessPointIDs {
			requestedSet[id] = struct{}{}
		}

		// Collect NetworkUnit IDs that contain any of the requested AccessPoint IDs.
		unitIDSet := make(map[int64]struct{})
		for _, unit := range networkUnits {
			if unit == nil {
				continue
			}

			for _, apID := range unit.AccessPoints {
				if _, found := requestedSet[apID]; !found {
					continue
				}
				unitIDSet[unit.ID] = struct{}{}

				break
			}
		}

		// Convert set to slice.
		networkUnitIDs = make([]int64, 0, len(unitIDSet))
		for id := range unitIDSet {
			networkUnitIDs = append(networkUnitIDs, id)
		}

		return nil
	})

	return networkUnitIDs, err
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

	upstreamNetworkUnits, err := s.daoNetworkUnit.ListWithoutCount(
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

func convertNetworkUnitConditionsToOptions(conditions ...*types.NetworkUnitCondition) []networkunit.OptFn {
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
				networkunit.WithGeneration(condition.ExactInclude.Generation...),
			)
		}

		if condition.ExactExclude != nil {
			opts = append(opts,
				networkunit.WithoutNetworkUnitID(condition.ExactExclude.NetworkUnitID...),
				networkunit.WithoutNetworkAreaID(condition.ExactExclude.NetworkAreaID...),
				networkunit.WithoutIsDirect(condition.ExactExclude.IsDirect...),
				networkunit.WithoutGeneration(condition.ExactExclude.Generation...),
			)
		}

		if condition.FuzzyInclude != nil {
			opts = append(opts,
				networkunit.WithFuzzyNetworkUnitName(condition.FuzzyInclude.NetworkUnitName...),
			)
		}

		if condition.FuzzyExclude != nil {
			opts = append(opts,
				networkunit.WithoutFuzzyNetworkUnitName(condition.FuzzyExclude.NetworkUnitName...),
			)
		}
	}

	return opts
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
	int64, *AccessPointResult, error) {

	var (
		networkUnitID int64
		data          *AccessPointResult
		err           error
	)

	err = s.WrapFn(nCtx, metricOperationCreateNetworkUnit, func(nCtx contextx.IContext) error {
		var err error
		if !networkUnit.IsDirect {
			if err = s.checkNetworkUnitLinks(nCtx, networkUnit); err != nil {
				return err
			}
		}

		if len(accessPoints) == 0 {
			networkUnit.AccessPoints = nil

			networkUnitID, err = s.daoNetworkUnit.Create(nCtx, networkUnit)
			if err != nil {
				return err
			}

			data = &AccessPointResult{}

			return nil
		}

		// create accesspoints first.
		var accessPointIDs []int64
		accessPointIDs, err = s.daoAccessPoint.CreateMany(nCtx, accessPoints...)
		if err != nil {
			logger.G.Sys().WithErr(err).Error("failed to create networkunit, failed to create accesspoint")

			return err
		}

		networkUnit.AccessPoints = accessPointIDs
		for idx, accessPointID := range accessPointIDs {
			if idx > len(accessPointIDs) {
				break
			}

			accessPoints[idx].ID = accessPointID
		}

		// create networkunit.
		networkUnitID, err = s.daoNetworkUnit.Create(nCtx, networkUnit)
		if err != nil {
			return err
		}

		data = &AccessPointResult{
			Created: accessPoints,
		}

		return nil
	})

	return networkUnitID, data, err
}

// UpdateNetworkUnit updates networkunit.
func (s *Storage) UpdateNetworkUnit(
	nCtx contextx.IContext, fields types.NetworkUnitUpdateFields, networkUnit *types.NetworkUnit, accessPoints ...*types.AccessPoint) (
	*AccessPointResult, error) {

	var (
		data *AccessPointResult
		err  error
	)

	err = s.WrapFn(nCtx, metricOperationUpdateNetworkUnit, func(nCtx contextx.IContext) error {
		var err error
		err = s.prepareNetworkUnitUpdate(nCtx, networkUnit)
		if err != nil {
			return err
		}

		if len(accessPoints) == 0 {
			data, err = s.updateNetworkUnitWithoutAccessPoints(nCtx, fields, networkUnit)
			return err
		}

		data, err = s.updateNetworkUnitWithAccessPoints(nCtx, fields, networkUnit, accessPoints...)

		return err
	})

	return data, err
}

func (s *Storage) prepareNetworkUnitUpdate(nCtx contextx.IContext, networkUnit *types.NetworkUnit) error {
	if networkUnit.IsDirect {
		return nil
	}

	return s.checkNetworkUnitLinks(nCtx, networkUnit)
}

func (s *Storage) updateNetworkUnitWithoutAccessPoints(
	nCtx contextx.IContext, fields types.NetworkUnitUpdateFields, networkUnit *types.NetworkUnit,
) (*AccessPointResult, error) {

	networkUnit.AccessPoints = nil
	if err := s.daoNetworkUnit.UpdateMany(nCtx, fields, networkUnit); err != nil {
		return nil, err
	}

	return &AccessPointResult{}, nil
}

func (s *Storage) updateNetworkUnitWithAccessPoints(
	nCtx contextx.IContext, fields types.NetworkUnitUpdateFields, networkUnit *types.NetworkUnit, accessPoints ...*types.AccessPoint,
) (*AccessPointResult, error) {

	accessPointIDs, oldAccessPoints, newAccessPoints := splitAccessPoints(accessPoints)

	accessPointIDs, err := s.upsertNetworkUnitAccessPoints(nCtx, accessPointIDs, oldAccessPoints, newAccessPoints)
	if err != nil {
		return nil, err
	}

	networkUnit.AccessPoints = accessPointIDs
	if err := s.daoNetworkUnit.UpdateMany(nCtx, fields, networkUnit); err != nil {
		return nil, err
	}

	return &AccessPointResult{
		Created: newAccessPoints,
		Updated: oldAccessPoints,
		Deleted: nil,
	}, nil
}

func splitAccessPoints(accessPoints []*types.AccessPoint) ([]int64, []*types.AccessPoint, []*types.AccessPoint) {
	accessPointIDs := make([]int64, 0, len(accessPoints))
	oldAccessPoints := make([]*types.AccessPoint, 0, len(accessPoints))
	newAccessPoints := make([]*types.AccessPoint, 0, len(accessPoints))
	for _, accessPoint := range accessPoints {
		if accessPoint.ID >= 0 {
			accessPointIDs = append(accessPointIDs, accessPoint.ID)
			oldAccessPoints = append(oldAccessPoints, accessPoint)

			continue
		}

		newAccessPoints = append(newAccessPoints, accessPoint)
	}

	return accessPointIDs, oldAccessPoints, newAccessPoints
}

func (s *Storage) upsertNetworkUnitAccessPoints(
	nCtx contextx.IContext, accessPointIDs []int64, oldAccessPoints, newAccessPoints []*types.AccessPoint,
) ([]int64, error) {

	if len(oldAccessPoints) > 0 {
		var err error
		err = s.daoAccessPoint.UpdateMany(nCtx, oldAccessPoints...)
		if err != nil {
			logger.G.Sys().WithErr(err).Error("failed to update networkunit, failed to update accesspoint")
			return nil, err
		}
	}

	if len(newAccessPoints) == 0 {
		return accessPointIDs, nil
	}

	createdAccessPointIDs, err := s.daoAccessPoint.CreateMany(nCtx, newAccessPoints...)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to update networkunit, failed to create accesspoint")
		return nil, err
	}

	accessPointIDs = append(accessPointIDs, createdAccessPointIDs...)
	for idx, accessPointID := range createdAccessPointIDs {
		newAccessPoints[idx].ID = accessPointID
	}

	return accessPointIDs, nil
}

// DeleteManyNetworkUnit deletes networkunit.
func (s *Storage) DeleteManyNetworkUnit(nCtx contextx.IContext, networkUnitIDs ...int64) error {
	return s.WrapFn(nCtx, metricOperationDeleteNetworkUnit, func(nCtx contextx.IContext) error {
		if len(networkUnitIDs) == 0 {
			return nil
		}

		if err := s.checkNetworkUnitDeleteProtection(nCtx, networkUnitIDs...); err != nil {
			return err
		}

		return s.daoNetworkUnit.DeleteMany(nCtx, networkUnitIDs...)
	})
}

func (s *Storage) checkNetworkUnitDeleteProtection(nCtx contextx.IContext, networkUnitIDs ...int64) error {
	if err := s.checkRunningHostNetworkUnitUsage(nCtx, networkUnitIDs...); err != nil {
		return err
	}

	if err := s.checkLinkedNetworkUnitUsage(nCtx, networkUnitIDs...); err != nil {
		return err
	}

	return nil
}

func (s *Storage) checkRunningHostNetworkUnitUsage(nCtx contextx.IContext, networkUnitIDs ...int64) error {
	opts := convertHostConditionsToOptions(&types.HostCondition{
		DynamicExactInclude: &types.HostDynamicExactFields{
			NetworkUnitID: networkUnitIDs,
			NodeStatus:    []types.NodeStatus{types.NodeStatusRunning},
		},
	})
	hostCounts, err := s.daoHost.CountGroupByNetworkUnitID(nCtx, opts...)
	if err != nil {
		return fmt.Errorf("count running hosts bound to networkunit: %w", err)
	}

	for _, networkUnitID := range networkUnitIDs {
		count := hostCounts[networkUnitID]
		if count > 0 {
			return fmt.Errorf("%w: networkunit %d is bound by %d running hosts",
				ErrNetworkUnitDeleteProtected, networkUnitID, count)
		}
	}

	return nil
}

func (s *Storage) checkLinkedNetworkUnitUsage(nCtx contextx.IContext, networkUnitIDs ...int64) error {
	linkedNetworkUnits, err := s.daoNetworkUnit.ListWithoutCount(
		nCtx,
		types.SingleItemPage(),
		networkunit.WithLinkedNetworkUnitID(networkUnitIDs...))
	if err != nil {
		return fmt.Errorf("list networkunits linked to networkunit: %w", err)
	}

	if len(linkedNetworkUnits) == 0 {
		return nil
	}

	linkedNetworkUnit := linkedNetworkUnits[0]
	if linkedNetworkUnit == nil {
		return nil
	}

	linkedID, ok := findLinkedNetworkUnitID(linkedNetworkUnit.Links, networkUnitIDs...)
	if !ok {
		return fmt.Errorf("%w: networkunit is referenced by networkunit %d",
			ErrNetworkUnitDeleteProtected, linkedNetworkUnit.ID)
	}

	return fmt.Errorf("%w: networkunit %d is referenced by networkunit %d",
		ErrNetworkUnitDeleteProtected, linkedID, linkedNetworkUnit.ID)
}

func findLinkedNetworkUnitID(links types.Links, networkUnitIDs ...int64) (int64, bool) {
	for _, link := range []*types.Link{links.Cluster, links.File, links.Data} {
		if link != nil && slices.Contains(networkUnitIDs, link.NetworkUnitID) {
			return link.NetworkUnitID, true
		}
	}

	return 0, false
}

// CountAccessPoint counts accesspoint.
func (s *Storage) CountAccessPoint(nCtx contextx.IContext, conditions ...*types.AccessPointCondition) (int64, error) {
	var (
		num int64
		err error
	)

	err = s.WrapFn(nCtx, metricOperationCountAccessPoint, func(nCtx contextx.IContext) error {
		var err error
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

		num, err = s.daoAccessPoint.Count(nCtx, opts...)

		return err
	})

	return num, err
}

// ListAccessPoint lists accesspoint.
func (s *Storage) ListAccessPoint(nCtx contextx.IContext, page types.Page, conditions ...*types.AccessPointCondition) (
	[]*types.AccessPoint, int64, error) {

	var (
		results []*types.AccessPoint
		num     int64
		err     error
	)

	err = s.WrapFn(nCtx, metricOperationListAccessPoint, func(nCtx contextx.IContext) error {
		var err error
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

		results, num, err = s.daoAccessPoint.List(nCtx, page, opts...)

		return err
	})

	return results, num, err
}

// ListAccessPointWithoutCount lists accesspoint by page and conditions without count.
func (s *Storage) ListAccessPointWithoutCount(
	nCtx contextx.IContext, page types.Page, conditions ...*types.AccessPointCondition) (
	[]*types.AccessPoint, error) {

	var results []*types.AccessPoint

	err := s.WrapFn(nCtx, metricOperationListAccessPointWithoutCount, func(nCtx contextx.IContext) error {
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

		var err error
		results, err = s.daoAccessPoint.ListWithoutCount(nCtx, page, opts...)

		return err
	})

	return results, err
}

// AccessPointResult describes the accesspoint result in networkunit handlers.
type AccessPointResult struct {
	Created []*types.AccessPoint
	Updated []*types.AccessPoint
	Deleted []*types.AccessPoint
}
