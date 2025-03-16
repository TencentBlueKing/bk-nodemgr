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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/accesspoint"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/business"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/host"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/networkarea"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/networkunit"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/topoevent"
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

	daoNetworkArea networkarea.Handler

	daoNetworkUnit networkunit.Handler

	daoAccessPoint accesspoint.Handler

	daoTopoEvent topoevent.Handler
}

func (s *storage) initDao() error {
	s.daoBusiness = business.New(s.Database, s.Logger)
	s.daoHost = host.New(s.Database, s.Logger)
	s.daoNetworkArea = networkarea.New(s.Database, s.Logger)
	s.daoNetworkUnit = networkunit.New(s.Database, s.Logger)
	s.daoAccessPoint = accesspoint.New(s.Database, s.Logger)
	s.daoTopoEvent = topoevent.New(s.Database, s.Logger)

	return nil
}

func (s *storage) check() error {
	if s.daoBusiness == nil {
		return errors.New("dao business is nil")
	}

	return nil
}

// UpsertManyBusiness updates or inserts many business.
func (s *storage) UpsertManyBusiness(ctx context.Context, biz ...*types.Business) error {
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
func (s *storage) ListBusinesses(ctx context.Context, page types.Page, conditions ...types.BusinessCondition) (
	[]*types.Business, int64, error) {

	opts := make([]business.OptFn, 0)
	for _, condition := range conditions {
		switch condition.Type {
		case types.ConditionTypeExactInclude:
			if condition.Exact != nil {
				opts = append(opts,
					business.WithBizID(condition.Exact.BizID...),
				)
			}

		case types.ConditionTypeExactExclude:
			if condition.Exact != nil {
				opts = append(opts,
					business.WithoutBizID(condition.Exact.BizID...),
				)
			}

		case types.ConditionTypeFuzzyInclude:
			if condition.Fuzzy != nil {
				opts = append(opts,
					business.WithFuzzyBizName(condition.Fuzzy.BizName...),
				)
			}

		case types.ConditionTypeFuzzyExclude:
			if condition.Fuzzy != nil {
				opts = append(opts,
					business.WithoutFuzzyBizName(condition.Fuzzy.BizName...),
				)
			}

		default:
			return nil, 0, fmt.Errorf("get unexpected condition type: %s", condition.Type)
		}
	}

	return s.daoBusiness.List(ctx, page, opts...)
}

// UpsertManyHost upserts many hosts.
func (s *storage) UpsertManyHost(ctx context.Context, hosts ...*types.Host) error {
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

// UpsertManyHostStatic updates or inserts host statics.
func (s *storage) UpsertManyHostStatic(ctx context.Context, hosts ...*types.Host) error {
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
// ListHost lists hosts by page and conditions.
func (s *storage) ListHost(ctx context.Context, page types.Page, conditions ...types.HostCondition) (
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
func (s *storage) CountHost(ctx context.Context, conditions ...types.HostCondition) (int64, error) {
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

// nolint:cyclop
// ListNetworkArea lists networkarea by page and conditions.
func (s *storage) ListNetworkArea(ctx context.Context, page types.Page, conditions ...types.NetworkAreaCondition) (
	[]*types.NetworkArea, int64, error) {

	opts := make([]networkarea.OptFn, 0)
	for _, condition := range conditions {
		switch condition.Type {
		case types.ConditionTypeExactInclude:
			if condition.Exact != nil {
				opts = append(opts,
					networkarea.WithNetworkAreaID(condition.Exact.NetworkAreaID...),
				)
			}

		case types.ConditionTypeExactExclude:
			if condition.Exact != nil {
				opts = append(opts,
					networkarea.WithoutNetworkAreaID(condition.Exact.NetworkAreaID...),
				)
			}

		case types.ConditionTypeFuzzyInclude:
			if condition.Fuzzy != nil {
				opts = append(opts,
					networkarea.WithFuzzyNetworkAreaName(condition.Fuzzy.NetworkAreaName...),
				)
			}

		case types.ConditionTypeFuzzyExclude:
			if condition.Fuzzy != nil {
				opts = append(opts,
					networkarea.WithoutFuzzyNetworkAreaName(condition.Fuzzy.NetworkAreaName...),
				)
			}

		default:
			return nil, 0, fmt.Errorf("get unexpected condition type: %s", condition.Type)
		}
	}

	return s.daoNetworkArea.List(ctx, page, opts...)
}

// GetNetworkArea gets networkarea by id.
func (s *storage) GetNetworkArea(ctx context.Context, networkAreaID int64) (*types.NetworkArea, error) {
	return s.daoNetworkArea.Get(ctx, networkAreaID)
}

// UpsertManyNetworkArea updates or inserts networkarea.
func (s *storage) UpsertManyNetworkArea(ctx context.Context, networkAreas ...*types.NetworkArea) error {
	return s.daoNetworkArea.UpsertMany(ctx, networkAreas...)
}

// UpdateManyNetworkArea updates networkarea.
func (s *storage) UpdateManyNetworkArea(ctx context.Context, networkArea ...*types.NetworkArea) error {
	return s.daoNetworkArea.UpdateMany(ctx, networkArea...)
}

// DeleteManyNetworkArea deletes networkarea.
func (s *storage) DeleteManyNetworkArea(ctx context.Context, networkAreaIDs ...int64) error {
	return s.daoNetworkArea.DeleteMany(ctx, networkAreaIDs...)
}

// ListNetworkUnit lists networkunit.
func (s *storage) ListNetworkUnit(ctx context.Context, page types.Page, conditions ...types.NetworkUnitCondition) (
	[]*types.NetworkUnit, int64, error) {

	opts := make([]networkunit.OptFn, 0)
	for _, condition := range conditions {
		switch condition.Type {
		case types.ConditionTypeExactInclude:
			if condition.Exact != nil {
				opts = append(opts,
					networkunit.WithNetworkUnitID(condition.Exact.NetworkUnitID...),
					networkunit.WithNetworkAreaID(condition.Exact.NetworkAreaID...),
				)
			}

		case types.ConditionTypeExactExclude:
			if condition.Exact != nil {
				opts = append(opts,
					networkunit.WithoutNetworkUnitID(condition.Exact.NetworkUnitID...),
					networkunit.WithoutNetworkAreaID(condition.Exact.NetworkAreaID...),
				)
			}

		case types.ConditionTypeFuzzyInclude:
		case types.ConditionTypeFuzzyExclude:

		default:
			return nil, 0, fmt.Errorf("get unexpected condition type: %s", condition.Type)
		}
	}

	return s.daoNetworkUnit.List(ctx, page, opts...)
}

// GetNetworkUnit gets networkunit by id.
func (s *storage) GetNetworkUnit(ctx context.Context, networkUnitID int64) (*types.NetworkUnit, error) {
	return s.daoNetworkUnit.Get(ctx, networkUnitID)
}

func (s *storage) checkNetworkUnitLinks(ctx context.Context, networkUnit *types.NetworkUnit) error {
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
		ctx,
		types.Page{Limit: 3},
		networkunit.WithNetworkUnitID(upstreamNetworkUnitIDs...))
	if err != nil {
		s.Logger.Errorf("failed to check networkunit links, failed to list upstream networkunit: %v", err.Error())
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
				return fmt.Errorf("upstream link networkarea not matched. networkarea-id(%d), networkunit-id(%d), accesspoint-id(%d)",
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
func (s *storage) CreateNetworkUnit(
	ctx context.Context,
	networkUnit *types.NetworkUnit,
	accessPoints ...*types.AccessPoint) (int64, *AccessPointResult, error) {

	if err := s.checkNetworkUnitLinks(ctx, networkUnit); err != nil {
		return -1, nil, err
	}

	if len(accessPoints) == 0 {
		networkUnit.AccessPoints = nil

		networkUnitID, err := s.daoNetworkUnit.Create(ctx, networkUnit)
		if err != nil {
			return -1, nil, err
		}

		return networkUnitID, &AccessPointResult{}, nil
	}

	// create accesspoints first.
	accessPointIDs, err := s.daoAccessPoint.CreateMany(ctx, accessPoints...)
	if err != nil {
		s.Logger.Errorf("failed to create networkunit, failed to create accesspoint: %v", err.Error())

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
	networkUnitID, err := s.daoNetworkUnit.Create(ctx, networkUnit)
	if err != nil {
		return -1, nil, err
	}

	return networkUnitID, &AccessPointResult{
		Created: accessPoints,
	}, nil
}

// UpdateNetworkUnit updates networkunit.
func (s *storage) UpdateNetworkUnit(
	ctx context.Context,
	networkUnit *types.NetworkUnit,
	accessPoints ...*types.AccessPoint) (*AccessPointResult, error) {

	if err := s.checkNetworkUnitLinks(ctx, networkUnit); err != nil {
		return nil, err
	}

	if len(accessPoints) == 0 {
		networkUnit.AccessPoints = nil

		if err := s.daoNetworkUnit.UpdateMany(ctx, networkUnit); err != nil {
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
		err := s.daoAccessPoint.UpdateMany(ctx, oldAccessPoints...)
		if err != nil {
			s.Logger.Errorf("failed to create networkunit, failed to update accesspoint: %v", err.Error())

			return nil, err
		}
	}
	if len(newAccessPoints) > 0 {
		createdAccessPointIDs, err := s.daoAccessPoint.CreateMany(ctx, newAccessPoints...)
		if err != nil {
			s.Logger.Errorf("failed to create networkunit, failed to create accesspoint: %v", err.Error())

			return nil, err
		}

		accessPointIDs = append(accessPointIDs, createdAccessPointIDs...)
		for idx, accessPointID := range accessPointIDs {
			if idx > len(newAccessPoints) {
				break
			}

			newAccessPoints[idx].ID = accessPointID
		}
	}
	networkUnit.AccessPoints = accessPointIDs

	if err := s.daoNetworkUnit.UpdateMany(ctx, networkUnit); err != nil {
		return nil, err
	}

	return &AccessPointResult{
		Created: newAccessPoints,
		Updated: oldAccessPoints,
		Deleted: nil,
	}, nil
}

// DeleteManyNetworkUnit deletes networkunit.
func (s *storage) DeleteManyNetworkUnit(ctx context.Context, networkUnitIDs ...int64) error {
	return s.daoNetworkUnit.DeleteMany(ctx, networkUnitIDs...)
}

// ListAccessPoint lists accesspoint.
func (s *storage) ListAccessPoint(ctx context.Context, page types.Page, conditions ...types.AccessPointCondition) (
	[]*types.AccessPoint, int64, error) {

	opts := make([]accesspoint.OptFn, 0)
	for _, condition := range conditions {
		switch condition.Type {
		case types.ConditionTypeExactInclude:
			if condition.Exact != nil {
				opts = append(opts,
					accesspoint.WithAccessPointID(condition.Exact.AccessPointID...),
					accesspoint.WithNetworkAreaID(condition.Exact.NetworkAreaID...),
				)
			}

		case types.ConditionTypeExactExclude:
			if condition.Exact != nil {
				opts = append(opts,
					accesspoint.WithoutAccessPointID(condition.Exact.AccessPointID...),
					accesspoint.WithoutNetworkAreaID(condition.Exact.NetworkAreaID...),
				)
			}

		case types.ConditionTypeFuzzyInclude:
		case types.ConditionTypeFuzzyExclude:

		default:
			return nil, 0, fmt.Errorf("get unexpected condition type: %s", condition.Type)
		}
	}

	return s.daoAccessPoint.List(ctx, page, opts...)
}

// CountTopoEvent counts topo events.
func (s *storage) CountTopoEvent(ctx context.Context, conditions ...types.TopoEventCondition) (int64, error) {
	opts := make([]topoevent.OptFn, 0)
	for _, condition := range conditions {
		switch condition.Type {
		case types.ConditionTypeExactInclude:
			if condition.Exact != nil {
				opts = append(opts,
					topoevent.WithNetworkAreaID(condition.Exact.NetworkAreaID...),
					topoevent.WithNetworkUnitID(condition.Exact.NetworkUnitID...),
					topoevent.WithAccessPointID(condition.Exact.AccessPointID...),
					topoevent.WithType(condition.Exact.Type...),
					topoevent.WithOperator(condition.Exact.Operator...),
				)
			}

		case types.ConditionTypeExactExclude:
			if condition.Exact != nil {
				opts = append(opts,
					topoevent.WithoutNetworkAreaID(condition.Exact.NetworkAreaID...),
					topoevent.WithoutNetworkUnitID(condition.Exact.NetworkUnitID...),
					topoevent.WithoutAccessPointID(condition.Exact.AccessPointID...),
					topoevent.WithoutType(condition.Exact.Type...),
					topoevent.WithoutOperator(condition.Exact.Operator...),
				)
			}

		case types.ConditionTypeFuzzyInclude:
		case types.ConditionTypeFuzzyExclude:

		default:
			return 0, fmt.Errorf("get unexpected condition type: %s", condition.Type)
		}
	}

	return s.daoTopoEvent.Count(ctx, opts...)
}

// ListTopoEvent lists topo events.
func (s *storage) ListTopoEvent(ctx context.Context, page types.Page, conditions ...types.TopoEventCondition) (
	[]*types.TopoEvent, int64, error) {

	opts := make([]topoevent.OptFn, 0)
	for _, condition := range conditions {
		switch condition.Type {
		case types.ConditionTypeExactInclude:
			if condition.Exact != nil {
				opts = append(opts,
					topoevent.WithNetworkAreaID(condition.Exact.NetworkAreaID...),
					topoevent.WithNetworkUnitID(condition.Exact.NetworkUnitID...),
					topoevent.WithAccessPointID(condition.Exact.AccessPointID...),
				)
			}

		case types.ConditionTypeExactExclude:
			if condition.Exact != nil {
				opts = append(opts,
					topoevent.WithoutNetworkAreaID(condition.Exact.NetworkAreaID...),
					topoevent.WithoutNetworkUnitID(condition.Exact.NetworkUnitID...),
					topoevent.WithoutAccessPointID(condition.Exact.AccessPointID...),
				)
			}

		case types.ConditionTypeFuzzyInclude:
		case types.ConditionTypeFuzzyExclude:

		default:
			return nil, 0, fmt.Errorf("get unexpected condition type: %s", condition.Type)
		}
	}

	return s.daoTopoEvent.List(ctx, page, opts...)
}

// CreateManyTopoEvent creates topo events.
func (s *storage) CreateManyTopoEvent(ctx context.Context, events ...*types.TopoEvent) error {
	return s.daoTopoEvent.CreateMany(ctx, events...)
}

// AccessPointResult describes the accesspoint result in networkunit handlers.
type AccessPointResult struct {
	Created []*types.AccessPoint
	Updated []*types.AccessPoint
	Deleted []*types.AccessPoint
}
