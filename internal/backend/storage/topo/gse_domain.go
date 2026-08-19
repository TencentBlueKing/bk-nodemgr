/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

// Package topo provides topology storage for backend.
// nolint: nonamedreturns
package topo

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/basestorage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/accesspoint"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/host"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// GetV4AgentAccessEndpoints get v4 agent access endpoints by networkunit id.
func (s *Storage) GetV4AgentAccessEndpoints(nCtx contextx.IContext, networkUnitID int64) (
	[]string, []string, []string, error) {

	var (
		cluster []string
		file    []string
		data    []string
		err     error
	)

	err = s.WrapFn(nCtx, metricOperationGetV4AgentAccessEndpoints, func(nCtx contextx.IContext) error {
		var err error
		cluster, file, data, err = s.getV4AgentAccessEndpoints(nCtx, networkUnitID)

		return err
	})

	return cluster, file, data, err
}

// nolint: nonamedreturns
func (s *Storage) getV4AgentAccessEndpoints(
	nCtx contextx.IContext, networkUnitID int64) (
	clusterEndpoints []string, fileEndpoints []string, dataEndpoints []string, err error) {

	if nCtx == nil {
		return nil, nil, nil, basestorage.ErrNilContent()
	}

	if networkUnitID < 0 {
		return nil, nil, nil, errors.New("unit id should be equal or greater than 0")
	}

	networkUnit, err := s.daoNetworkUnit.Get(nCtx, networkUnitID)
	if err != nil {
		return nil, nil, nil,
			fmt.Errorf("failed to get networkunit by id, networkunit-id(%d): %w", networkUnitID, err)
	}

	// direct unit return direct endpoints.
	if networkUnit.IsDirect {
		if networkUnit.DirectEndpoints == nil {
			return nil, nil, nil,
				fmt.Errorf("networkunit-id(%d) is direct unit, but direct endpoints is nil", networkUnitID)
		}

		return networkUnit.DirectEndpoints.Cluster, networkUnit.DirectEndpoints.File, networkUnit.DirectEndpoints.Data, nil
	}

	hosts, count, err := s.daoHost.List(nCtx, types.UnlimitedPage(),
		host.WithDynamicNetworkUnitID(networkUnitID),
		host.WithDynamicNodeRole(types.NodeRoleProxy),
		host.WithDynamicNodeStatus(types.NodeStatusRunning),
		host.WithDynamicProxyAccessDisabled(false),
	)
	if err != nil {
		return nil, nil, nil,
			fmt.Errorf("failed to get host by networkunit id, networkunit-id(%d): %w", networkUnitID, err)
	}
	if count == 0 {
		return nil, nil, nil,
			fmt.Errorf("failed to get host by networkunit id, result count is 0, networkunit-id(%d)", networkUnitID)
	}

	clusterMap := make(map[string]struct{})
	fileMap := make(map[string]struct{})
	dataMap := make(map[string]struct{})
	for _, host := range hosts {
		if host.Dynamic.ProxySupportCluster() {
			clusterMap[fmt.Sprintf("%s:%d", host.Dynamic.AdvertiseIP, host.Dynamic.ProxyClusterPort)] = struct{}{}
		}

		if host.Dynamic.ProxySupportFile() {
			fileMap[fmt.Sprintf("%s:%d", host.Dynamic.AdvertiseIP, host.Dynamic.ProxyFilePort)] = struct{}{}
		}

		if host.Dynamic.ProxySupportData() {
			dataMap[fmt.Sprintf("%s:%d", host.Dynamic.AdvertiseIP, host.Dynamic.ProxyDataPort)] = struct{}{}
		}
	}

	return conv.MapKeyToSlice(clusterMap), conv.MapKeyToSlice(fileMap), conv.MapKeyToSlice(dataMap), nil
}

// GetV6AgentAccessEndpoints get v6 agent access endpoints by networkunit id.
func (s *Storage) GetV6AgentAccessEndpoints(nCtx contextx.IContext, networkUnitID int64) (
	[]string, []string, []string, error) {

	var (
		cluster []string
		file    []string
		data    []string
		err     error
	)

	err = s.WrapFn(nCtx, metricOperationGetV6AgentAccessEndpoints, func(nCtx contextx.IContext) error {
		var err error
		cluster, file, data, err = s.getV6AgentAccessEndpoints(nCtx, networkUnitID)

		return err
	})

	return cluster, file, data, err
}

// nolint: nonamedreturns
func (s *Storage) getV6AgentAccessEndpoints(
	nCtx contextx.IContext, networkUnitID int64) (
	clusterEndpoints []string, fileEndpoints []string, dataEndpoints []string, err error) {

	if nCtx == nil {
		return nil, nil, nil, basestorage.ErrNilContent()
	}

	if networkUnitID < 0 {
		return nil, nil, nil, errors.New("unit id should be equal or greater than 0")
	}

	networkUnit, err := s.daoNetworkUnit.Get(nCtx, networkUnitID)
	if err != nil {
		return nil, nil, nil,
			fmt.Errorf("failed to get networkunit by id, networkunit-id(%d): %w", networkUnitID, err)
	}

	// direct unit return direct endpoints.
	if networkUnit.IsDirect {
		if networkUnit.DirectEndpoints == nil {
			return nil, nil, nil,
				fmt.Errorf("networkunit-id(%d) is direct unit, but direct endpoints is nil", networkUnitID)
		}

		return networkUnit.DirectEndpoints.Cluster, networkUnit.DirectEndpoints.File, networkUnit.DirectEndpoints.Data, nil
	}

	hosts, count, err := s.daoHost.List(nCtx, types.UnlimitedPage(),
		host.WithDynamicNetworkUnitID(networkUnitID),
		host.WithDynamicNodeRole(types.NodeRoleProxy),
		host.WithDynamicNodeStatus(types.NodeStatusRunning),
		host.WithDynamicProxyAccessDisabled(false),
	)
	if err != nil {
		return nil, nil, nil,
			fmt.Errorf("failed to get host by networkunit id, networkunit-id(%d): %w", networkUnitID, err)
	}
	if count == 0 {
		return nil, nil, nil,
			fmt.Errorf("failed to get host by networkunit id, result count is 0, networkunit-id(%d)", networkUnitID)
	}

	clusterMap := make(map[string]struct{})
	fileMap := make(map[string]struct{})
	dataMap := make(map[string]struct{})
	for _, host := range hosts {
		if host.Dynamic.ProxySupportCluster() {
			clusterMap[fmt.Sprintf("%s:%d", host.Dynamic.AdvertiseIPV6, host.Dynamic.ProxyClusterPort)] = struct{}{}
		}

		if host.Dynamic.ProxySupportFile() {
			fileMap[fmt.Sprintf("%s:%d", host.Dynamic.AdvertiseIPV6, host.Dynamic.ProxyFilePort)] = struct{}{}
		}

		if host.Dynamic.ProxySupportData() {
			dataMap[fmt.Sprintf("%s:%d", host.Dynamic.AdvertiseIPV6, host.Dynamic.ProxyDataPort)] = struct{}{}
		}
	}

	return conv.MapKeyToSlice(clusterMap), conv.MapKeyToSlice(fileMap), conv.MapKeyToSlice(dataMap), nil
}

// GetProxyUpstreamAccessEndpoints gets proxy upstream accesspoint.
func (s *Storage) GetProxyUpstreamAccessEndpoints(nCtx contextx.IContext, networkUnitID int64) (
	[]string, []string, []string, error) {

	if nCtx == nil {
		return nil, nil, nil, basestorage.ErrNilContent()
	}

	if networkUnitID < 0 {
		return nil, nil, nil, errors.New("unit id should be equal or greater than 0")
	}

	var (
		clusterEndpoints []string
		fileEndpoints    []string
		dataEndpoints    []string
		err              error
	)

	err = s.WrapFn(nCtx, metricOperationGetProxyUpstreamAccessPoints, func(nCtx contextx.IContext) error {
		var err error
		clusterEndpoints, fileEndpoints, dataEndpoints, err = s.getProxyUpstreamEndpoints(nCtx, networkUnitID)

		return err
	})

	return clusterEndpoints, fileEndpoints, dataEndpoints, err
}

func (s *Storage) getProxyUpstreamEndpoints(nCtx contextx.IContext, networkUnitID int64) ([]string, []string, []string, error) {
	clusterAPID, fileAPID, dataAPID, err := s.getProxyUpstreamAccessPointIDs(nCtx, networkUnitID)
	if err != nil {
		return nil, nil, nil, err
	}

	accesspoints, _, err := s.daoAccessPoint.List(
		nCtx, types.UnlimitedPage(), accesspoint.WithAccessPointID(clusterAPID, fileAPID, dataAPID))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get upstreams accesspoint, networkunit-id(%d): %w", networkUnitID, err)
	}

	return resolveProxyUpstreamEndpoints(types.AccessPointList(accesspoints), clusterAPID, fileAPID, dataAPID, networkUnitID)
}

func (s *Storage) getProxyUpstreamAccessPointIDs(nCtx contextx.IContext, networkUnitID int64) (int64, int64, int64, error) {
	networkUnit, err := s.daoNetworkUnit.Get(nCtx, networkUnitID)
	if err != nil {
		return -1, -1, -1, fmt.Errorf("failed to get networkunit by id, networkunit-id(%d): %w", networkUnitID, err)
	}

	if networkUnit.IsDirect {
		return -1, -1, -1, fmt.Errorf("networkunit-id(%d) is direct unit, can not have proxy", networkUnitID)
	}

	if networkUnit.Links.Cluster == nil || networkUnit.Links.File == nil || networkUnit.Links.Data == nil {
		return -1, -1, -1, fmt.Errorf(
			"networkunit-id(%d) links is invalid, cluster or file or data have empty upstreams", networkUnitID)
	}

	return networkUnit.Links.Cluster.AccessPointID,
		networkUnit.Links.File.AccessPointID,
		networkUnit.Links.Data.AccessPointID,
		nil
}

func resolveProxyUpstreamEndpoints(
	apList types.AccessPointList, clusterAPID, fileAPID, dataAPID, networkUnitID int64,
) ([]string, []string, []string, error) {

	clusterEndpoints, err := getProxyUpstreamEndpoint(apList, clusterAPID, networkUnitID, func(ap *types.AccessPoint) []string {
		return ap.Endpoints.Cluster
	})
	if err != nil {
		return nil, nil, nil, err
	}

	fileEndpoints, err := getProxyUpstreamEndpoint(apList, fileAPID, networkUnitID, func(ap *types.AccessPoint) []string {
		return ap.Endpoints.File
	})
	if err != nil {
		return nil, nil, nil, err
	}

	dataEndpoints, err := getProxyUpstreamEndpoint(apList, dataAPID, networkUnitID, func(ap *types.AccessPoint) []string {
		return ap.Endpoints.Data
	})

	return clusterEndpoints, fileEndpoints, dataEndpoints, err
}

func getProxyUpstreamEndpoint(
	apList types.AccessPointList, accessPointID, networkUnitID int64, endpointSelector func(*types.AccessPoint) []string,
) ([]string, error) {

	ap, ok := apList.Found(accessPointID)
	if !ok {
		return nil, fmt.Errorf("networkunit-id(%d) links accesspoint-id(%d) not found", networkUnitID, accessPointID)
	}

	return endpointSelector(ap), nil
}

// NeedStaticAccess check host is need static access or not.
func (s *Storage) NeedStaticAccess(nCtx contextx.IContext, networkUnitID int64) (bool, error) {
	if nCtx == nil {
		return false, basestorage.ErrNilContent()
	}

	if networkUnitID < 0 {
		return false, errors.New("unit id should be equal or greater than 0")
	}

	var (
		result bool
		err    error
	)

	err = s.WrapFn(nCtx, metricOperationNeedStaticAccess, func(nCtx contextx.IContext) error {
		var err error
		var networkUnit *types.NetworkUnit
		if networkUnit, err = s.daoNetworkUnit.Get(nCtx, networkUnitID); err != nil {
			return fmt.Errorf("failed to get networkunit by id, networkunit-id(%d): %w", networkUnitID, err)
		}

		need := false
		if networkUnit.NetworkAreaID == types.DefaultNetworkAreaID && !networkUnit.IsDirect {
			need = true
		}

		result = need

		return nil
	})
	if err != nil {
		return false, err
	}

	return result, nil
}

// GetNetworkUnitCustomDeployConfig gets network unit custom deploy config by network unit id and os type.
func (s *Storage) GetNetworkUnitCustomDeployConfig(nCtx contextx.IContext, networkUnitID int64, osType criteria.OSType) (
	*types.CustomDeployConfig, error) {

	if nCtx == nil {
		return nil, basestorage.ErrNilContent()
	}

	if networkUnitID < 0 {
		return nil, errors.New("unit id should be equal or greater than 0")
	}

	customDeployConfig := &types.CustomDeployConfig{}
	err := s.WrapFn(nCtx, metricOperationGetNetworkUnitCustomDeployConfig, func(nCtx contextx.IContext) error {
		var err error
		var networkUnit *types.NetworkUnit
		if networkUnit, err = s.daoNetworkUnit.Get(nCtx, networkUnitID); err != nil {
			return fmt.Errorf("failed to get networkunit by id, networkunit-id(%d): %w", networkUnitID, err)
		}

		if networkUnit.CustomDeployConfig == nil {
			return nil
		}

		config, ok := networkUnit.CustomDeployConfig[osType]
		if !ok {
			return nil
		}

		customDeployConfig = &config

		return nil
	})
	if err != nil {
		return nil, err
	}

	return customDeployConfig, nil
}
