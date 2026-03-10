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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// relayInfoToEndpoint converts types.RelayInfo to discover.Endpoint.
func relayInfoToEndpoint(relayInfo *types.RelayInfo, port int64) discover.Endpoint {
	return discover.Endpoint{
		IPV4: relayInfo.InnerIP,
		IPV6: relayInfo.InnerIPV6,
		Port: int(port),
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
func (std *NodeActionStandarder) GetRelayInfos() ([]*types.RelayInfo, error) {
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

	relayInfoMap, err := conv.SliceToMap(relayInfos, func(v *types.RelayInfo) int64 {
		return v.HostID
	})
	if err != nil {
		return nil, fmt.Errorf("failed to convert relay infos to map: %w", err)
	}

	for _, relayInfo := range relayInfoMap {
		if relayInfo.DownloadSvcPort <= 0 || relayInfo.CallbackSvcPort <= 0 {
			delete(relayInfoMap, relayInfo.HostID)
		}
	}

	validRelayInfos := conv.MapValueToSlice(relayInfoMap)
	if len(validRelayInfos) == 0 {
		return nil, fmt.Errorf("no relay host with download service port in network unit %d", networkUnitID)
	}

	return validRelayInfos, nil
}

// RelayInfosToEndpoints converts relay infos to callback and download endpoint slices.
// nolint: nonamedreturns
func RelayInfosToEndpoints(infos []*types.RelayInfo) (callbacks []discover.Endpoint, downloads []discover.Endpoint) {
	callbacks = make([]discover.Endpoint, len(infos))
	downloads = make([]discover.Endpoint, len(infos))
	for i, info := range infos {
		callbacks[i] = relayInfoToEndpoint(info, info.CallbackSvcPort)
		downloads[i] = relayInfoToEndpoint(info, info.DownloadSvcPort)
	}

	return callbacks, downloads
}

// GetRelayEndpoints queries Relay hosts and returns Endpoint list, callback and download endpoints.
// Returns: (callback endpoints, download endpoints, error).
func (std *NodeActionStandarder) GetRelayEndpoints() ([]discover.Endpoint, []discover.Endpoint, error) {
	infos, err := std.GetRelayInfos()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get relay infos: %w", err)
	}

	callbacks, downloads := RelayInfosToEndpoints(infos)

	return callbacks, downloads, nil
}
