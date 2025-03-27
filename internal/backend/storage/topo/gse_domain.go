/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package topo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/accesspoint"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/host"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// GetV4AgentAccessEndpoints get v4 agent access endpoints by networkunit id.
func (s *Storage) GetV4AgentAccessEndpoints(ctx context.Context, networkUnitID int64) (
	[]string, []string, []string, error) {

	return s.getAgentAccessEndpoints(ctx, networkUnitID, func(static *types.HostStatic) string {
		return static.InnerIP
	})
}

// GetV6AgentAccessEndpoints get v6 agent access endpoints by networkunit id.
func (s *Storage) GetV6AgentAccessEndpoints(ctx context.Context, networkUnitID int64) (
	[]string, []string, []string, error) {

	return s.getAgentAccessEndpoints(ctx, networkUnitID, func(static *types.HostStatic) string {
		return static.InnerIPV6
	})
}

// nolint: nonamedreturns
func (s *Storage) getAgentAccessEndpoints(
	ctx context.Context, networkUnitID int64, ipSelector func(static *types.HostStatic) string) (
	clusterEndpoints []string, fileEndpoints []string, dataEndpoints []string, err error) {

	if ctx == nil {
		return nil, nil, nil, base.ErrNilContent()
	}

	if networkUnitID < 0 {
		return nil, nil, nil, errors.New("unit id should be equal or greater than 0")
	}

	networkUnit, err := s.daoNetworkUnit.Get(ctx, networkUnitID)
	if err != nil {
		return nil, nil, nil,
			fmt.Errorf("failed to get networkunit by id, networkunit-id(%d), err: %w", networkUnitID, err)
	}

	// direct unit return direct endpoints.
	if networkUnit.IsDirect {
		if networkUnit.DirectEndpoints == nil {
			return nil, nil, nil,
				fmt.Errorf("networkunit-id(%d) is direct unit, but direct endpoints is nil", networkUnitID)
		}

		return networkUnit.DirectEndpoints.Cluster, networkUnit.DirectEndpoints.File, networkUnit.DirectEndpoints.Data, nil
	}

	hosts, count, err := s.daoHost.List(ctx, types.UnlimitedPage(),
		host.WithNetworkUnitID(networkUnitID),
		host.WithNodeRole(types.NodeRoleProxy),
	)
	if err != nil {
		return nil, nil, nil,
			fmt.Errorf("failed to get host by networkunit id, networkunit-id(%d), err: %w", networkUnitID, err)
	}
	if count == 0 {
		return nil, nil, nil,
			fmt.Errorf("failed to get host by networkunit id, result count is 0, networkunit-id(%d)", networkUnitID)
	}

	clusterMap := make(map[string]bool)
	fileMap := make(map[string]bool)
	dataMap := make(map[string]bool)
	for idx := range hosts {
		ips := strings.Split(ipSelector(hosts[idx].Static), ",")
		for _, ip := range ips {
			clusterMap[fmt.Sprintf("%s:%d", ip, hosts[idx].Dynamic.ProxyClusterPort)] = true
			fileMap[fmt.Sprintf("%s:%d", ip, hosts[idx].Dynamic.ProxyFilePort)] = true
			dataMap[fmt.Sprintf("%s:%d", ip, hosts[idx].Dynamic.ProxyDataPort)] = true
		}
	}

	return conv.MapKeyToSlice(clusterMap), conv.MapKeyToSlice(fileMap), conv.MapKeyToSlice(dataMap), nil
}

// GetProxyUpstreamAccessEndpoints gets proxy upstream accesspoint.
// nolint: nonamedreturns
func (s *Storage) GetProxyUpstreamAccessEndpoints(ctx context.Context, networkUnitID int64) (
	clusterEndpoints []string, fileEndpoints []string, dataEndpoints []string, err error) {

	if ctx == nil {
		return nil, nil, nil, base.ErrNilContent()
	}

	if networkUnitID < 0 {
		return nil, nil, nil, errors.New("unit id should be equal or greater than 0")
	}

	networkUnit, err := s.daoNetworkUnit.Get(ctx, networkUnitID)
	if err != nil {
		return nil, nil, nil,
			fmt.Errorf("failed to get networkunit by id, networkunit-id(%d), err: %w", networkUnitID, err)
	}

	// direct unit return direct endpoints.
	if networkUnit.IsDirect {
		return nil, nil, nil,
			fmt.Errorf("networkunit-id(%d) is direct unit, can not have proxy", networkUnitID)
	}

	if networkUnit.Links.Cluster == nil || networkUnit.Links.File == nil || networkUnit.Links.Data == nil {
		return nil, nil, nil,
			fmt.Errorf("networkunit-id(%d) links is invalid, cluster or file or data have empty upstreams", networkUnitID)
	}

	accesspoints, _, err := s.daoAccessPoint.List(ctx, types.UnlimitedPage(), accesspoint.WithAccessPointID(
		networkUnit.Links.Cluster.AccessPointID,
		networkUnit.Links.File.AccessPointID,
		networkUnit.Links.Data.AccessPointID,
	))
	if err != nil {
		return nil, nil, nil,
			fmt.Errorf("failed to get upstreams accesspoint, networkunit-id(%d), err: %w", networkUnitID, err)
	}

	apList := types.AccessPointList(accesspoints)

	if ap, ok := apList.Found(networkUnit.Links.Cluster.AccessPointID); ok {
		clusterEndpoints = ap.Endpoints.Cluster
	} else {
		return nil, nil, nil, fmt.Errorf("networkunit-id(%d) links cluster accesspoint not found", networkUnitID)
	}

	if ap, ok := apList.Found(networkUnit.Links.File.AccessPointID); ok {
		fileEndpoints = ap.Endpoints.File
	} else {
		return nil, nil, nil, fmt.Errorf("networkunit-id(%d) links cluster accesspoint not found", networkUnitID)
	}

	if ap, ok := apList.Found(networkUnit.Links.Data.AccessPointID); ok {
		dataEndpoints = ap.Endpoints.Data
	} else {
		return nil, nil, nil, fmt.Errorf("networkunit-id(%d) links cluster accesspoint not found", networkUnitID)
	}

	return clusterEndpoints, fileEndpoints, dataEndpoints, nil
}
