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

// Package firewall configures the Windows firewall before a full node install.
package firewall

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"
)

const (
	featureGSE     = "gse"
	featureNodeMgr = "nodemgr"
)

// Step applies the configured firewall rules.
type Step struct {
	args StepArgs
}

// StepArgs defines the firewall configuration passed by full-install.
type StepArgs struct {
	DeployEnv string
	Features  string
	RemoteIPs string
}

// NewStep creates a firewall configuration step.
func NewStep(args StepArgs) *Step {
	return &Step{args: args}
}

// Run applies firewall rules before the full install download step.
func (step *Step) Run(ctx context.Context) error {
	return runFirewall(ctx, step.args)
}

type firewallRule struct {
	Name      string
	Direction string
	Protocol  string
	PortType  string
	Ports     string
	RemoteIPs string
}

type featureRule struct {
	suffix    string
	direction string
	protocol  string
	portType  string
	ports     string
}

var featureRules = map[string][]featureRule{
	featureGSE: {
		{suffix: "Inbound_GSE2_Rule_TCP", direction: "in", protocol: "TCP", portType: "localport", ports: "60020-60030,20020,20030"},
		{suffix: "Outbound_GSE2_Rule_TCP", direction: "out", protocol: "TCP", portType: "remoteport", ports: "28625,28668,28925,20020,20030,10205,60020-60030,7070,8080,80"},
		{suffix: "Inbound_GSE2_Rule_UDP", direction: "in", protocol: "UDP", portType: "localport", ports: "60020-60030,20020,20030,10020,10030"},
		{suffix: "Outbound_GSE2_Rule_UDP", direction: "out", protocol: "UDP", portType: "remoteport", ports: "60020-60030,20020,20030,10020,10030"},
	},
	featureNodeMgr: {
		{suffix: "Outbound_NodeMgr_Rule_TCP", direction: "out", protocol: "TCP", portType: "remoteport", ports: "80"},
	},
}

func parseFeatures(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	seen := make(map[string]struct{})
	features := make([]string, 0)
	for _, item := range strings.Split(raw, ",") {
		feature := strings.ToLower(strings.TrimSpace(item))
		if feature == "" {
			return nil, fmt.Errorf("firewall feature is empty")
		}
		if _, ok := featureRules[feature]; !ok {
			return nil, fmt.Errorf("unsupported firewall feature: %s", feature)
		}
		if _, ok := seen[feature]; ok {
			continue
		}
		seen[feature] = struct{}{}
		features = append(features, feature)
	}

	sort.Strings(features)
	return features, nil
}

func parseRemoteIPs(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	seen := make(map[string]struct{})
	remoteIPs := make([]string, 0)
	for _, item := range strings.Split(raw, ",") {
		cidr := strings.TrimSpace(item)
		if cidr == "" {
			return nil, fmt.Errorf("firewall remote IP is empty")
		}
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, fmt.Errorf("invalid firewall remote IP CIDR %q: %w", cidr, err)
		}
		normalized := network.String()
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		remoteIPs = append(remoteIPs, normalized)
	}

	sort.Strings(remoteIPs)
	return remoteIPs, nil
}

func buildRules(deployEnv string, features, remoteIPs []string) ([]firewallRule, error) {
	if len(features) > 0 && len(remoteIPs) == 0 {
		return nil, fmt.Errorf("firewall remote IPs are required when firewall features are set")
	}

	ipv4, ipv6 := splitRemoteIPs(remoteIPs)
	rules := make([]firewallRule, 0)
	for _, feature := range features {
		for _, definition := range featureRules[feature] {
			if len(ipv4) > 0 {
				rules = append(rules, firewallRule{
					Name:      deployEnv + "_" + definition.suffix + "_IPv4",
					Direction: definition.direction,
					Protocol:  definition.protocol,
					PortType:  definition.portType,
					Ports:     definition.ports,
					RemoteIPs: strings.Join(ipv4, ","),
				})
			}
			if len(ipv6) > 0 {
				rules = append(rules, firewallRule{
					Name:      deployEnv + "_" + definition.suffix + "_IPv6",
					Direction: definition.direction,
					Protocol:  definition.protocol,
					PortType:  definition.portType,
					Ports:     definition.ports,
					RemoteIPs: strings.Join(ipv6, ","),
				})
			}
		}
	}

	return rules, nil
}

func splitRemoteIPs(remoteIPs []string) (ipv4, ipv6 []string) {
	for _, cidr := range remoteIPs {
		ip, _, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if ip.To4() != nil {
			ipv4 = append(ipv4, cidr)
			continue
		}
		ipv6 = append(ipv6, cidr)
	}
	return ipv4, ipv6
}
