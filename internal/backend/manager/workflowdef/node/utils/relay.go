/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package utils

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// relayInfoToEndpoint converts types.RelayInfo to discover.Endpoint.
func relayInfoToEndpoint(relayInfo *types.RelayInfo, port int64) discover.Endpoint {
	return discover.Endpoint{
		IPV4: relayInfo.InnerIP,
		IPV6: relayInfo.InnerIPV6,
		Port: int(port),
		Nice: 0,
		Meta: nil,
	}
}

// getNetworkUnitID gets the network unit ID from the standarder.
func (std *NodeActionStandarder) getNetworkUnitID() int64 {
	networkUnitID := std.DeployInfo().Host.Dynamic.NetworkUnitID
	// For Proxy nodes, use ProxyInstallOriginUnitID
	if std.DeployInfo().Host.Dynamic.NodeRole == types.NodeRoleProxy {
		networkUnitID = std.DeployInfo().Host.Dynamic.ProxyInstallOriginUnitID
	}
	return networkUnitID
}

// GetRelayInfos queries Relay hosts and returns RelayInfo list.
// It filters RelayInfo that has download service port and selects using round-robin selector.
// If count > 0, returns up to count RelayInfos; otherwise returns a single RelayInfo.
func (std *NodeActionStandarder) GetRelayInfos(count int) ([]*types.RelayInfo, error) {
	if std.storageHost == nil {
		return nil, fmt.Errorf("storageHost is not set")
	}

	networkUnitID := std.getNetworkUnitID()

	// Query RelayInfo list
	relayInfos, err := std.storageHost.GetRelayInfosInNetworkUnit(std.Context(), networkUnitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get relay infos in network unit %d: %w", networkUnitID, err)
	}

	if len(relayInfos) == 0 {
		return nil, fmt.Errorf("no available relay host in network unit %d", networkUnitID)
	}

	// Filter RelayInfo that has download service port
	validRelayInfos := make([]*types.RelayInfo, 0, len(relayInfos))
	for _, relayInfo := range relayInfos {
		if relayInfo.DownloadSvcPort > 0 {
			validRelayInfos = append(validRelayInfos, relayInfo)
		}
	}

	if len(validRelayInfos) == 0 {
		return nil, fmt.Errorf("no relay host with download service port in network unit %d", networkUnitID)
	}

	// Convert to endpoints for selection
	endpoints := make([]discover.Endpoint, 0, len(validRelayInfos))
	relayInfoMap := make(map[string]*types.RelayInfo)
	for _, relayInfo := range validRelayInfos {
		ep := relayInfoToEndpoint(relayInfo, relayInfo.DownloadSvcPort)
		endpoints = append(endpoints, ep)
		// Use combined key (IPv4-IPv6) to avoid conflicts when multiple relays have same IPv4 but different IPv6
		key := fmt.Sprintf("%s-%s", ep.GetIPV4Address(), ep.GetIPV6Address())
		relayInfoMap[key] = relayInfo
	}

	// Select endpoints using round-robin selector
	selector := discover.NewRoundRobinSelector()
	var selectedEndpoints []discover.Endpoint
	if count > 0 {
		selectedEndpoints, err = discover.SelectEndpoints(endpoints, count, selector)
	} else {
		selectedEndpoint, err := selector.Select(endpoints)
		if err == nil {
			selectedEndpoints = []discover.Endpoint{selectedEndpoint}
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to select relay endpoints: %w", err)
	}

	// Map selected endpoints back to RelayInfo
	result := make([]*types.RelayInfo, 0, len(selectedEndpoints))
	for _, ep := range selectedEndpoints {
		// Use combined key (IPv4-IPv6) to match the key used when building the map
		key := fmt.Sprintf("%s-%s", ep.GetIPV4Address(), ep.GetIPV6Address())
		relayInfo, ok := relayInfoMap[key]
		if !ok {
			continue
		}
		result = append(result, relayInfo)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("failed to find relay info for selected endpoints")
	}

	return result, nil
}
