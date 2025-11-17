/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package topo provide topology storage.
// nolint: nonamedreturns
package topo

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/accesspoint"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/business"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/host"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/networkarea"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/networkunit"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/topoevent"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// StorageName ...
const StorageName = "topo"

// NewStorage ...
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

// Storage implements the IStorage interface.
type Storage struct {
	basestorage.Storage

	daoBusiness business.IHandler

	daoHost host.IHandler

	daoNetworkArea networkarea.IHandler

	daoNetworkUnit networkunit.IHandler

	daoAccessPoint accesspoint.IHandler

	daoTopoEvent topoevent.IHandler
}

func (s *Storage) initDao() error {
	s.daoBusiness = business.New(s.Database)
	s.daoHost = host.New(s.Database)
	s.daoNetworkArea = networkarea.New(s.Database)
	s.daoNetworkUnit = networkunit.New(s.Database)
	s.daoAccessPoint = accesspoint.New(s.Database)
	s.daoTopoEvent = topoevent.New(s.Database)

	return nil
}

func (s *Storage) check() error {
	if s.daoBusiness == nil {
		return errors.New("dao business is nil")
	}

	if s.daoHost == nil {
		return errors.New("dao host is nil")
	}
	if s.daoNetworkArea == nil {
		return errors.New("dao network area is nil")
	}
	if s.daoNetworkUnit == nil {
		return errors.New("dao network unit is nil")
	}
	if s.daoAccessPoint == nil {
		return errors.New("dao access point is nil")
	}
	if s.daoTopoEvent == nil {
		return errors.New("dao topo event is nil")
	}

	return nil
}

func (s *Storage) metric() *storage.MetricData {
	return storage.Metric(StorageName)
}

// GetHostsByAreaAndInnerIP get hosts by area and inner ip.
func (s *Storage) GetHostsByAreaAndInnerIP(nCtx contextx.IContext, networkAreaID int64, innerip string) (
	results []*types.Host, err error) {

	// record metric.
	metric := s.metric().Start("get_hosts_by_area_and_inner_ip")
	defer metric.End(err)

	results, err = s.getHostsByAreaAndInnerIP(nCtx, networkAreaID, innerip)

	return results, err
}

// ExistDedicatedInstallerProxyHost exists dedicated installer proxy host by network unit id.
func (s *Storage) ExistDedicatedInstallerProxyHost(nCtx contextx.IContext, networkUnitID int64) (
	exist bool, err error) {

	// record metric.
	metric := s.metric().Start("exist_dedicated_installer_proxy_host")
	defer metric.End(err)

	exist, err = s.existDedicatedInstallerProxyHost(nCtx, networkUnitID)

	return exist, err
}

// GetNetworkUnitByIDs list network unit by unit ids.
func (s *Storage) GetNetworkUnitByIDs(nCtx contextx.IContext, networkUnitIDs []int64) (
	results []*types.NetworkUnit, err error) {

	// record metric.
	metric := s.metric().Start("get_network_unit_by_ids")
	defer metric.End(err)

	results, err = s.getNetworkUnitByIDs(nCtx, networkUnitIDs)

	return results, err
}
