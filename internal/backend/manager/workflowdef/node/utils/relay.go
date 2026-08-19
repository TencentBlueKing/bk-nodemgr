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

package utils

import (
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// relayInfoToEndpoint converts types.RelayInfo to discover.Endpoint with the given port.
func relayInfoToEndpoint(relayInfo *types.RelayInfo, port int64) discover.Endpoint {
	return discover.Endpoint{
		IPV4: relayInfo.AdvertiseIP,
		IPV6: relayInfo.AdvertiseIPV6,
		Port: int(port),
		Meta: nil,
	}
}

// RelayToEndpoints returns callback and download endpoints from a single relay.
func RelayToEndpoints(relay *types.RelayInfo) ([]discover.Endpoint, []discover.Endpoint) {
	return []discover.Endpoint{relayInfoToEndpoint(relay, relay.CallbackSvcPort)},
		[]discover.Endpoint{relayInfoToEndpoint(relay, relay.DownloadSvcPort)}
}

// BuildRelayServerURLs builds callback and download urls directly from the selected relay.
// Returns: (callbackURL, downloadURL).
func (std *NodeActionStandarder) BuildRelayServerURLs(relay *types.RelayInfo) (string, string) {
	callbackURL := BuildServerURLs(relayInfoToEndpoint(relay, relay.CallbackSvcPort))
	downloadURL := BuildServerURLs(relayInfoToEndpoint(relay, relay.DownloadSvcPort))

	return callbackURL, downloadURL
}

// GetSelectedRelay returns the relay that was selected by SelectRelayHost action.
func (std *NodeActionStandarder) GetSelectedRelay() (*types.RelayInfo, error) {
	relay := &std.DeployInfo().RelayInfo
	if relay.HostID <= 0 {
		return nil, errors.New("host id is required")
	}

	if relay.AgentID == "" {
		return nil, errors.New("agent id is required")
	}

	if relay.DownloadSvcPort <= 0 {
		return nil, errors.New("download service port is required")
	}

	if relay.CallbackSvcPort <= 0 {
		return nil, errors.New("callback service port is required")
	}

	if relay.AdvertiseIP == "" && relay.AdvertiseIPV6 == "" {
		return nil, errors.New("advertise ip or ipv6 is required")
	}

	return relay, nil
}
