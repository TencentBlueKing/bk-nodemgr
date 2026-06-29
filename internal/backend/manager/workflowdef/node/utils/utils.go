/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package utils use to provide some common utils for node actions.
package utils

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/installer"
)

const (
	// DefaultEndpointSelectionCount is the default count when selecting endpoints.
	DefaultEndpointSelectionCount = 3
)

// NodeInstallerEndpointSource defines where installer callback/download endpoints should come from.
type NodeInstallerEndpointSource string

const (
	// NodeInstallerEndpointSourceServer uses backend/file service endpoints selected from discover.
	NodeInstallerEndpointSourceServer NodeInstallerEndpointSource = "server"
	// NodeInstallerEndpointSourceRelay uses the selected relay's callback/download endpoints.
	NodeInstallerEndpointSourceRelay NodeInstallerEndpointSource = "relay"
	// NodeInstallerEndpointSourceProxySelf uses the proxy host's own relay callback/download endpoints.
	NodeInstallerEndpointSourceProxySelf NodeInstallerEndpointSource = "proxy_self"
)

func buildServerURL(needV4, needV6 bool, endpoints ...discover.Endpoint) []string {
	if len(endpoints) == 0 {
		return nil
	}

	addrs := make([]string, 0, len(endpoints))
	for _, ep := range endpoints {
		if needV4 && ep.IPV4 != "" {
			addr := ep.GetIPV4Address()
			if addr != "" {
				addrs = append(addrs, "http://"+addr)
			}
		}

		if needV6 && ep.IPV6 != "" {
			addr := ep.GetIPV6Address()
			if addr != "" {
				addrs = append(addrs, "http://"+addr)
			}
		}
	}

	if len(addrs) == 0 {
		return nil
	}

	rand.Shuffle(len(addrs), func(i, j int) {
		addrs[i], addrs[j] = addrs[j], addrs[i]
	})

	return addrs
}

// BuildServerURLs builds a comma-separated string of multiple addresses in the format: http://addr1,http://addr2,http://addr3.
// The installer already supports this format and will use utils.SplitServerAddrs() to split and process it.
// If endpoints is empty, returns an empty string.
func BuildServerURLs(endpoints ...discover.Endpoint) string {
	addrs := buildServerURL(true, true, endpoints...)
	if len(addrs) == 0 {
		return ""
	}

	return strings.Join(addrs, installer.ServerAddrSeparator)
}

// SelectOneServerV4URL selects one ipv4 address from the given endpoints.
func SelectOneServerV4URL(endpoints []discover.Endpoint) string {
	addrs := buildServerURL(true, false, endpoints...)
	if len(addrs) == 0 {
		return ""
	}

	return addrs[0]
}

// SelectOneServerV6URL selects one ipv6 address from the given endpoints.
func SelectOneServerV6URL(endpoints []discover.Endpoint) string {
	addrs := buildServerURL(false, true, endpoints...)
	if len(addrs) == 0 {
		return ""
	}

	return addrs[0]
}

// SelectInstallEndpointSource selects install callback/download endpoint source from control mode.
func SelectInstallEndpointSource(std *NodeActionStandarder) NodeInstallerEndpointSource {
	if std.DeployInfo().InstallOptions.DirectInstall {
		return NodeInstallerEndpointSourceServer
	}

	return NodeInstallerEndpointSourceRelay
}

// GenerateNodeInstallerServerEndpoints generates installer callback/download endpoints from the given source.
// Returns: (callback endpoints, download endpoints).
func GenerateNodeInstallerServerEndpoints(
	std *NodeActionStandarder,
	provider discover.Discover,
	source NodeInstallerEndpointSource) (
	[]discover.Endpoint, []discover.Endpoint, error) {
		
	switch source {
	case NodeInstallerEndpointSourceServer:
		return generateDirectServerEndpoints(provider)
	case NodeInstallerEndpointSourceRelay:
		return generateInDirectServerEndpoints(std)
	case NodeInstallerEndpointSourceProxySelf:
		return generateProxySelfEndpoints(std)
	default:
		return nil, nil, fmt.Errorf("unsupported installer endpoint source: %s", source)
	}
}

func generateInDirectServerEndpoints(std *NodeActionStandarder) ([]discover.Endpoint, []discover.Endpoint, error) {
	relay, err := std.GetSelectedRelay()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get selected relay: %w", err)
	}

	callbackEndpoints, downloadEndpoints := RelayToEndpoints(relay)

	// shuffle callback endpoints.
	rand.Shuffle(len(callbackEndpoints), func(i, j int) {
		callbackEndpoints[i], callbackEndpoints[j] = callbackEndpoints[j], callbackEndpoints[i]
	})

	if len(callbackEndpoints) > DefaultEndpointSelectionCount {
		callbackEndpoints = callbackEndpoints[:DefaultEndpointSelectionCount]
	}

	// shuffle download endpoints.
	rand.Shuffle(len(downloadEndpoints), func(i, j int) {
		downloadEndpoints[i], downloadEndpoints[j] = downloadEndpoints[j], downloadEndpoints[i]
	})

	if len(downloadEndpoints) > DefaultEndpointSelectionCount {
		downloadEndpoints = downloadEndpoints[:DefaultEndpointSelectionCount]
	}

	return callbackEndpoints, downloadEndpoints, nil
}

func generateDirectServerEndpoints(provider discover.Discover) ([]discover.Endpoint, []discover.Endpoint, error) {
	callbackEndpoints, err := provider.SelectEndpoints(
		discover.ServiceNameBackend,
		discover.EndpointNameBackendCallback,
		DefaultEndpointSelectionCount,
		discover.NewRoundRobinSelector())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to select backend callback endpoints: %w", err)
	}

	downloadEndpoints, err := provider.SelectEndpoints(
		discover.ServiceNameFile,
		discover.EndpointNameFileDownload,
		DefaultEndpointSelectionCount,
		discover.NewRoundRobinSelector())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to select file endpoints: %w", err)
	}

	return callbackEndpoints, downloadEndpoints, nil
}

func generateProxySelfEndpoints(std *NodeActionStandarder) ([]discover.Endpoint, []discover.Endpoint, error) {
	host := std.DeployInfo().Host
	if host.Dynamic.RelayCallbackPort <= 0 || host.Dynamic.RelayDownloadPort <= 0 {
		return nil, nil, fmt.Errorf("host relay ports are not set")
	}
	innerIP := ""
	if len(host.Static.InnerIPList) > 0 {
		innerIP = host.Static.InnerIPList[0]
	}

	innerIPV6 := ""
	if len(host.Static.InnerIPV6List) > 0 {
		innerIPV6 = host.Static.InnerIPV6List[0]
	}

	if innerIP == "" && innerIPV6 == "" {
		return nil, nil, fmt.Errorf("host inner ip and ipv6 are not set")
	}

	callbackEndpoints := []discover.Endpoint{
		{
			IPV4: innerIP,
			IPV6: innerIPV6,
			Port: int(host.Dynamic.RelayCallbackPort),
		},
	}

	downloadEndpoints := []discover.Endpoint{
		{
			IPV4: innerIP,
			IPV6: innerIPV6,
			Port: int(host.Dynamic.RelayDownloadPort),
		},
	}

	return callbackEndpoints, downloadEndpoints, nil
}
