/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package topo provides topology Storage for nodeman.
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
func NewStorage(client *mongo.Client, database string, logger logger.Logger) (*Storage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &Storage{
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

// Storage implements the IStorage interface.
type Storage struct {
	base.Storage

	daoBusiness business.IHandler

	daoHost host.IHandler

	daoNetworkArea networkarea.Handler

	daoNetworkUnit networkunit.Handler

	daoAccessPoint accesspoint.Handler

	daoTopoEvent topoevent.IHandler
}

func (s *Storage) initDao() error {
	s.daoBusiness = business.New(s.Database, s.Logger)
	s.daoHost = host.New(s.Database, s.Logger)
	s.daoNetworkArea = networkarea.New(s.Database, s.Logger)
	s.daoNetworkUnit = networkunit.New(s.Database, s.Logger)
	s.daoAccessPoint = accesspoint.New(s.Database, s.Logger)
	s.daoTopoEvent = topoevent.New(s.Database, s.Logger)

	return nil
}

func (s *Storage) check() error {
	if s.daoBusiness == nil {
		return errors.New("dao business is nil")
	}

	return nil
}

// UpsertManyBusiness updates or inserts many business.
func (s *Storage) UpsertManyBusiness(ctx context.Context, biz ...*types.Business) error {
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
func (s *Storage) ListBusinesses(ctx context.Context, page types.Page, conditions ...*types.BusinessCondition) (
	[]*types.Business, int64, error) {

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

	return s.daoBusiness.List(ctx, page, opts...)
}

// ListNetworkArea lists networkarea by page and conditions.
// nolint: cyclop
func (s *Storage) ListNetworkArea(ctx context.Context, page types.Page, conditions ...*types.NetworkAreaCondition) (
	[]*types.NetworkArea, int64, error) {

	opts := make([]networkarea.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				networkarea.WithNetworkAreaID(condition.ExactInclude.NetworkAreaID...),
			)
		}

		if condition.ExactExclude != nil {
			opts = append(opts,
				networkarea.WithoutNetworkAreaID(condition.ExactExclude.NetworkAreaID...),
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

	return s.daoNetworkArea.List(ctx, page, opts...)
}

// GetNetworkArea gets networkarea by id.
func (s *Storage) GetNetworkArea(ctx context.Context, networkAreaID int64) (*types.NetworkArea, error) {
	return s.daoNetworkArea.Get(ctx, networkAreaID)
}

// UpsertManyNetworkArea updates or inserts networkarea.
func (s *Storage) UpsertManyNetworkArea(ctx context.Context, networkAreas ...*types.NetworkArea) error {
	return s.daoNetworkArea.UpsertMany(ctx, networkAreas...)
}

// UpdateManyNetworkArea updates networkarea.
func (s *Storage) UpdateManyNetworkArea(ctx context.Context, networkArea ...*types.NetworkArea) error {
	return s.daoNetworkArea.UpdateMany(ctx, networkArea...)
}

// DeleteManyNetworkArea deletes networkarea.
func (s *Storage) DeleteManyNetworkArea(ctx context.Context, networkAreaIDs ...int64) error {
	return s.daoNetworkArea.DeleteMany(ctx, networkAreaIDs...)
}

// ListNetworkUnit lists networkunit.
func (s *Storage) ListNetworkUnit(ctx context.Context, page types.Page, conditions ...*types.NetworkUnitCondition) (
	[]*types.NetworkUnit, int64, error) {

	opts := make([]networkunit.OptFn, 0)
	for _, condition := range conditions {
		if condition == nil {
			continue
		}

		if condition.ExactInclude != nil {
			opts = append(opts,
				networkunit.WithNetworkUnitID(condition.ExactInclude.NetworkUnitID...),
				networkunit.WithNetworkAreaID(condition.ExactInclude.NetworkAreaID...),
			)
		}

		if condition.ExactExclude != nil {
			opts = append(opts,
				networkunit.WithoutNetworkUnitID(condition.ExactExclude.NetworkUnitID...),
				networkunit.WithoutNetworkAreaID(condition.ExactExclude.NetworkAreaID...),
			)
		}
	}

	return s.daoNetworkUnit.List(ctx, page, opts...)
}

// GetNetworkUnit gets networkunit by id.
func (s *Storage) GetNetworkUnit(ctx context.Context, networkUnitID int64) (*types.NetworkUnit, error) {
	return s.daoNetworkUnit.Get(ctx, networkUnitID)
}

func (s *Storage) checkNetworkUnitLinks(ctx context.Context, networkUnit *types.NetworkUnit) error {
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
func (s *Storage) CreateNetworkUnit(
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
func (s *Storage) UpdateNetworkUnit(
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

		for idx, accessPointID := range createdAccessPointIDs {
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
func (s *Storage) DeleteManyNetworkUnit(ctx context.Context, networkUnitIDs ...int64) error {
	return s.daoNetworkUnit.DeleteMany(ctx, networkUnitIDs...)
}

// CountAccessPoint counts accesspoint.
func (s *Storage) CountAccessPoint(ctx context.Context, conditions ...*types.AccessPointCondition) (int64, error) {
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

	return s.daoAccessPoint.Count(ctx, opts...)
}

// ListAccessPoint lists accesspoint.
func (s *Storage) ListAccessPoint(ctx context.Context, page types.Page, conditions ...*types.AccessPointCondition) (
	[]*types.AccessPoint, int64, error) {

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

	return s.daoAccessPoint.List(ctx, page, opts...)
}

// AccessPointResult describes the accesspoint result in networkunit handlers.
type AccessPointResult struct {
	Created []*types.AccessPoint
	Updated []*types.AccessPoint
	Deleted []*types.AccessPoint
}
