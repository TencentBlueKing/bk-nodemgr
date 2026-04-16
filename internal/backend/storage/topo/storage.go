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
package topo

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/accesspoint"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/business"
	gsdao "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/globalsettings"
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

const (
	metricOperationUpsertManyBusiness                        = "upsert_many_business"
	metricOperationListBusiness                              = "list_business"
	metricOperationListNetworkArea                           = "list_networkarea"
	metricOperationGetNetworkArea                            = "get_networkarea"
	metricOperationUpsertManyNetworkArea                     = "upsert_many_networkarea"
	metricOperationUpdateManyNetworkArea                     = "update_many_networkarea"
	metricOperationDeleteManyNetworkArea                     = "delete_many_networkarea"
	metricOperationListNetworkUnit                           = "list_networkunit"
	metricOperationGetNetworkUnitDistributionByNetworkAreaID = "get_networkunit_distribution_by_networkarea_id"
	metricOperationGetNetworkUnit                            = "get_networkunit"
	metricOperationCreateNetworkUnit                         = "create_networkunit"
	metricOperationUpdateNetworkUnit                         = "update_networkunit"
	metricOperationDeleteNetworkUnit                         = "delete_networkunit"
	metricOperationCountAccessPoint                          = "count_accesspoint"
	metricOperationListAccessPoint                           = "list_accesspoint"
	metricOperationGetHostDistributionByNodeRole             = "get_host_distribution_by_node_role"
	metricOperationGetHostDistributionByNetworkAreaID        = "get_host_distribution_by_networkarea_id"
	metricOperationGetHostsByAreaAndInnerIP                  = "get_hosts_by_area_and_inner_ip"
	metricOperationExistDedicatedInstallerProxyHost          = "exist_dedicated_installer_proxy_host"
	metricOperationGetNetworkUnitByIDs                       = "get_networkunit_by_ids"
	metricOperationListHostWithFields                        = "list_host_with_fields"
	metricOperationGetRelayInfosInNetworkUnit                = "get_relay_infos_in_network_unit"
	metricOperationGetHostByID                               = "get_host_by_id"
	metricOperationUpsertManyHost                            = "upsert_many_host"
	metricOperationUpsertManyHostStatic                      = "upsert_many_host_static"
	metricOperationUpdateManyHostDynamic                     = "update_many_host_dynamic"
	metricOperationListHost                                  = "list_host"
	metricOperationListHostOrderByUpdateTime                 = "list_host_order_by_updatetime"
	metricOperationCountHost                                 = "count_host"
	metricOperationDistinctHost                              = "distinct_host"
	metricOperationRecommendNetworkUnitByNetworkSegment      = "recommend_networkunit_by_network_segment"
	metricOperationDeleteManyHost                            = "delete_many_host"
	metricOperationFindHostWithDynamic                       = "find_host_with_dynamic"
	metricOperationUpdateHostDynamicFields                   = "update_host_dynamic_fields"
	metricOperationTouchHostOperationTime                    = "touch_host_operation_time"
	metricOperationGetV4AgentAccessEndpoints                 = "get_v4_agent_access_endpoints"
	metricOperationGetV6AgentAccessEndpoints                 = "get_v6_agent_access_endpoints"
	metricOperationGetProxyUpstreamAccessPoints              = "get_proxy_upstream_accesspoints"
	metricOperationNeedStaticAccess                          = "need_static_access"
	metricOperationCountTopoEvent                            = "count_topo_event"
	metricOperationListTopoEvent                             = "list_topo_event"
	metricOperationCreateManyTopoEvent                       = "create_many_topo_event"
	metricOperationDistinctTopoEvent                         = "distinct_topo_event"
	metricOperationGetNetworkUnitCustomDeployConfig          = "get_networkunit_custom_deploy_config"
	metricOperationGetNetworkUnitIDsByAccessPoints           = "get_networkunit_ids_by_accesspoints"
)

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

	daoGlobalSettings gsdao.IHandler

	daoHost host.IHandler

	daoNetworkArea networkarea.IHandler

	daoNetworkUnit networkunit.IHandler

	daoAccessPoint accesspoint.IHandler

	daoTopoEvent topoevent.IHandler
}

func (s *Storage) initDao() error {
	s.daoBusiness = business.New(s.Database)
	s.daoGlobalSettings = gsdao.New(s.Database)
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
	if s.daoGlobalSettings == nil {
		return errors.New("dao global settings is nil")
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

// GetHostsByAreaAndInnerIP get hosts by area and inner ip.
func (s *Storage) GetHostsByAreaAndInnerIP(nCtx contextx.IContext, networkAreaID int64, innerip string) (
	[]*types.Host, error) {

	var results []*types.Host

	err := s.WrapFn(nCtx, metricOperationGetHostsByAreaAndInnerIP, func(nCtx contextx.IContext) error {
		var err error
		results, err = s.getHostsByAreaAndInnerIP(nCtx, networkAreaID, innerip)

		return err
	})

	return results, err
}

// ExistDedicatedInstallerProxyHost exists dedicated installer proxy host by network unit id.
func (s *Storage) ExistDedicatedInstallerProxyHost(nCtx contextx.IContext, networkUnitIDs []int64) (
	map[int64]bool, error) {

	var exist map[int64]bool

	err := s.WrapFn(nCtx, metricOperationExistDedicatedInstallerProxyHost, func(nCtx contextx.IContext) error {
		var err error
		exist, err = s.existDedicatedInstallerProxyHost(nCtx, networkUnitIDs)

		return err
	})

	return exist, err
}

// GetNetworkUnitByIDs list network unit by unit ids.
func (s *Storage) GetNetworkUnitByIDs(nCtx contextx.IContext, networkUnitIDs []int64) (
	[]*types.NetworkUnit, error) {

	var results []*types.NetworkUnit

	err := s.WrapFn(nCtx, metricOperationGetNetworkUnitByIDs, func(nCtx contextx.IContext) error {
		var err error
		results, err = s.getNetworkUnitByIDs(nCtx, networkUnitIDs)

		return err
	})

	return results, err
}

// ListHostWithFields lists hosts with fields.
func (s *Storage) ListHostWithFields(nCtx contextx.IContext, page types.Page,
	selection *types.HostFieldSelection, conditions ...*types.HostCondition) (
	[]*types.Host, int64, error) {

	var (
		results []*types.Host
		num     int64
	)

	err := s.WrapFn(nCtx, metricOperationListHostWithFields, func(nCtx contextx.IContext) error {
		var err error
		results, num, err = s.listHostWithFields(nCtx, page, selection, conditions...)

		return err
	})

	return results, num, err
}

// GetRelayInfosInNetworkUnit gets available Relay Infos in the specified network unit.
// Returns RelayInfo list with DedicatedInstaller tag and Running status.
func (s *Storage) GetRelayInfosInNetworkUnit(nCtx contextx.IContext, networkUnitID int64) (
	[]*types.RelayInfo, error) {

	var results []*types.RelayInfo

	err := s.WrapFn(nCtx, metricOperationGetRelayInfosInNetworkUnit, func(nCtx contextx.IContext) error {
		var err error
		results, err = s.getRelayInfosInNetworkUnit(nCtx, networkUnitID)

		return err
	})

	return results, err
}
