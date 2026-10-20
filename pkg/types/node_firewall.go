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

package types

import (
	"cmp"
	"fmt"
	"net/netip"
	"slices"
)

// FirewallFeature is a port feature predefined by the system. The ports behind a feature are computed by the
// installer, users can never specify ports, protocols or directions.
type FirewallFeature string

const (
	// FirewallFeatureGSE opens the ports used by the gse agent.
	FirewallFeatureGSE FirewallFeature = "gse"

	// FirewallFeatureNodeMgr opens the ports used to reach node manager.
	FirewallFeatureNodeMgr FirewallFeature = "nodemgr"
)

// Validate validates the firewall feature.
func (feature FirewallFeature) Validate() error {
	switch feature {
	case FirewallFeatureGSE, FirewallFeatureNodeMgr:
		return nil
	default:
		return fmt.Errorf("invalid firewall feature: %q", feature)
	}
}

// DeploymentFirewall is the local firewall setting the installer applies before installing the agent.
type DeploymentFirewall struct {
	// RemoteIPs are the CIDRs allowed by the firewall rules of all enabled features.
	RemoteIPs []string
	// Features are the enabled port features. Empty means no rule is managed.
	Features []FirewallFeature
}

// Normalize validates the firewall and returns a copy whose remote ips are masked, deduplicated and sorted,
// and whose features are deduplicated in first-seen order.
func (firewall DeploymentFirewall) Normalize() (DeploymentFirewall, error) {
	features := make([]FirewallFeature, 0, len(firewall.Features))
	for _, feature := range firewall.Features {
		if err := feature.Validate(); err != nil {
			return DeploymentFirewall{}, err
		}

		if !slices.Contains(features, feature) {
			features = append(features, feature)
		}
	}

	prefixes := make([]netip.Prefix, 0, len(firewall.RemoteIPs))
	for _, remoteIP := range firewall.RemoteIPs {
		prefix, err := parseFirewallRemoteIP(remoteIP)
		if err != nil {
			return DeploymentFirewall{}, err
		}

		prefixes = append(prefixes, prefix)
	}

	slices.SortFunc(prefixes, func(a, b netip.Prefix) int {
		return cmp.Or(a.Addr().Compare(b.Addr()), cmp.Compare(a.Bits(), b.Bits()))
	})
	prefixes = slices.Compact(prefixes)

	remoteIPs := make([]string, 0, len(prefixes))
	for _, prefix := range prefixes {
		remoteIPs = append(remoteIPs, prefix.String())
	}

	if len(features) > 0 && len(remoteIPs) == 0 {
		return DeploymentFirewall{}, fmt.Errorf("firewall features %v require at least one remote ip", features)
	}

	return DeploymentFirewall{RemoteIPs: remoteIPs, Features: features}, nil
}

// parseFirewallRemoteIP parses a CIDR and masks its host bits. A bare ip is rejected with a hint.
func parseFirewallRemoteIP(remoteIP string) (netip.Prefix, error) {
	prefix, err := netip.ParsePrefix(remoteIP)
	if err == nil {
		return prefix.Masked(), nil
	}

	addr, addrErr := netip.ParseAddr(remoteIP)
	if addrErr != nil || addr.Zone() != "" {
		return netip.Prefix{}, fmt.Errorf("invalid firewall remote ip %q, a CIDR is required: %w", remoteIP, err)
	}

	bits := 128
	if addr.Is4() {
		bits = 32
	}

	return netip.Prefix{}, fmt.Errorf("firewall remote ip %q is not a CIDR, write a single address as %s/%d",
		remoteIP, remoteIP, bits)
}
