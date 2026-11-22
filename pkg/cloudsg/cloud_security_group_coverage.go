/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package cloudsg

import (
	"net/netip"
	"sort"
	"strconv"
	"strings"
)

const (
	// cloudSGPortAll is the port expression that matches every port.
	cloudSGPortAll = "ALL"
	// cloudSGProtocolAll is the protocol that matches every protocol.
	cloudSGProtocolAll = "ALL"
	// cloudSGProtocolTCP is the protocol assumed when a rule leaves the protocol empty.
	cloudSGProtocolTCP = "TCP"
)

// CloudSecurityGroupRuleCovers reports whether the covering ACCEPT rule fully covers the
// target ACCEPT rule: the target source cidr, destination ports and protocol are all subsets
// of the covering rule. It is used to skip authorizing permissions that existing entries already
// grant, for example a broader cidr or an ALL port rule.
func CloudSecurityGroupRuleCovers(covering, target CloudSecurityGroupRule) bool {
	if !cloudSGActionGrants(covering) || !cloudSGActionGrants(target) {
		return false
	}

	if !cloudSGProtocolCovers(covering.Protocol, target.Protocol) {
		return false
	}

	if !cloudSGCIDRCovers(covering.CIDR, target.CIDR) {
		return false
	}

	return cloudSGPortsCover(covering.Port, target.Port)
}

// CloudSecurityGroupRuleCoveredBy reports whether the union of the existing ACCEPT rules covers
// the target ACCEPT rule. Existing rules whose protocol and cidr each cover the target as a whole
// contribute their ports, and every target port must fall into the union of those ports, so
// scattered or adjacent port entries jointly satisfy a combined multi-port target.
// The cidr is not unioned: a target cidr split across several narrower existing cidrs is reported
// as not covered, which only causes a redundant authorization and never skips a required one.
func CloudSecurityGroupRuleCoveredBy(existing []CloudSecurityGroupRule, target CloudSecurityGroupRule) bool {
	if !cloudSGActionGrants(target) {
		return false
	}

	// an ALL port target is only covered by an ALL port rule.
	if strings.EqualFold(target.Port, cloudSGPortAll) {
		for _, covering := range existing {
			if CloudSecurityGroupRuleCovers(covering, target) {
				return true
			}
		}

		return false
	}

	targetRanges, ok := cloudSGParsePortRanges(target.Port)
	if !ok {
		return false
	}

	for _, targetRange := range targetRanges {
		if !cloudSGPortRangeCoveredByUnion(existing, target, targetRange) {
			return false
		}
	}

	return true
}

// cloudSGPortRangeCoveredByUnion reports whether the target port range is covered by the union of
// the ports of every existing rule whose protocol and cidr cover the target.
func cloudSGPortRangeCoveredByUnion(
	existing []CloudSecurityGroupRule, target CloudSecurityGroupRule, targetRange cloudSGPortRange,
) bool {

	coveringRanges := make([]cloudSGPortRange, 0, len(existing))
	for _, covering := range existing {
		if !cloudSGActionGrants(covering) {
			continue
		}

		if !cloudSGProtocolCovers(covering.Protocol, target.Protocol) {
			continue
		}

		if !cloudSGCIDRCovers(covering.CIDR, target.CIDR) {
			continue
		}

		if strings.EqualFold(covering.Port, cloudSGPortAll) {
			return true
		}

		ranges, ok := cloudSGParsePortRanges(covering.Port)
		if !ok {
			continue
		}

		coveringRanges = append(coveringRanges, ranges...)
	}

	// the union of all matching rules is evaluated together, scattered ranges
	// jointly cover the target.
	return cloudSGPortRangesUnionCovered(coveringRanges, targetRange)
}

// cloudSGActionGrants reports whether the rule grants traffic. Empty actions default to ACCEPT.
func cloudSGActionGrants(rule CloudSecurityGroupRule) bool {
	return rule.EffectiveAction() == CloudSecurityGroupRuleActionAccept
}

// cloudSGProtocolCovers reports whether the covering protocol covers the target protocol.
// ALL covers every protocol, empty protocols are treated as TCP.
func cloudSGProtocolCovers(covering, target string) bool {
	if strings.EqualFold(covering, cloudSGProtocolAll) {
		return true
	}

	covering = cloudSGDefaultProtocol(covering)
	target = cloudSGDefaultProtocol(target)

	return strings.EqualFold(covering, target)
}

func cloudSGDefaultProtocol(protocol string) string {
	if protocol == "" {
		return cloudSGProtocolTCP
	}

	return protocol
}

// cloudSGCIDRCovers reports whether the covering cidr contains the whole target cidr.
// Unparseable values never cover anything, keeping the authorize behavior conservative.
func cloudSGCIDRCovers(covering, target string) bool {
	coveringPrefix, ok := cloudSGParsePrefix(covering)
	if !ok {
		return false
	}

	targetPrefix, ok := cloudSGParsePrefix(target)
	if !ok {
		return false
	}

	return coveringPrefix.Bits() <= targetPrefix.Bits() && coveringPrefix.Contains(targetPrefix.Addr())
}

// cloudSGParsePrefix parses a cidr or a bare ip into a prefix.
func cloudSGParsePrefix(value string) (netip.Prefix, bool) {
	if value == "" {
		return netip.Prefix{}, false
	}

	if prefix, err := netip.ParsePrefix(value); err == nil {
		return prefix.Masked(), true
	}

	addr, err := netip.ParseAddr(value)
	if err != nil {
		return netip.Prefix{}, false
	}

	return netip.PrefixFrom(addr, addr.BitLen()), true
}

// cloudSGPortRange is one closed port interval.
type cloudSGPortRange struct {
	lo uint16
	hi uint16
}

// cloudSGPortsCover reports whether every target port is covered by the covering port
// expression. Supported expressions: all, single port, discrete list "a,b,c", range "a-b".
// ALL is only covered by ALL.
func cloudSGPortsCover(covering, target string) bool {
	if strings.EqualFold(covering, cloudSGPortAll) {
		return true
	}

	if strings.EqualFold(target, cloudSGPortAll) {
		return false
	}

	coveringRanges, ok := cloudSGParsePortRanges(covering)
	if !ok {
		return false
	}

	targetRanges, ok := cloudSGParsePortRanges(target)
	if !ok {
		return false
	}

	for _, targetRange := range targetRanges {
		if !cloudSGPortRangesUnionCovered(coveringRanges, targetRange) {
			return false
		}
	}

	return true
}

// cloudSGPortRangesUnionCovered reports whether the target range is fully covered by the
// union of the covering ranges. Contiguous and overlapping ranges jointly cover the target,
// evaluated in sorted order with interval subtraction.
func cloudSGPortRangesUnionCovered(covering []cloudSGPortRange, target cloudSGPortRange) bool {
	if len(covering) == 0 {
		return false
	}

	sorted := append([]cloudSGPortRange(nil), covering...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].lo != sorted[j].lo {
			return sorted[i].lo < sorted[j].lo
		}

		return sorted[i].hi < sorted[j].hi
	})

	cursor := int32(target.lo)
	for _, coveringRange := range sorted {
		if int32(coveringRange.hi) < cursor {
			continue
		}

		// a gap before the cursor means the target is not fully covered.
		if int32(coveringRange.lo) > cursor {
			return false
		}

		cursor = int32(coveringRange.hi) + 1
		if cursor > int32(target.hi) {
			return true
		}
	}

	return cursor > int32(target.hi)
}

// portRangeBoundCount is the bound count of one port range expression "lo-hi".
const portRangeBoundCount = 2

// cloudSGParsePortRanges parses a port expression into closed intervals.
func cloudSGParsePortRanges(expression string) ([]cloudSGPortRange, bool) {
	if expression == "" {
		return nil, false
	}

	parts := strings.Split(expression, ",")
	ranges := make([]cloudSGPortRange, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)

		bounds := strings.SplitN(part, "-", portRangeBoundCount)
		lo, err := strconv.ParseUint(bounds[0], 10, 16)
		if err != nil {
			return nil, false
		}

		hi := lo
		if len(bounds) == portRangeBoundCount {
			hi, err = strconv.ParseUint(bounds[1], 10, 16)
			if err != nil {
				return nil, false
			}
		}

		if hi < lo {
			return nil, false
		}

		ranges = append(ranges, cloudSGPortRange{lo: uint16(lo), hi: uint16(hi)})
	}

	return ranges, true
}
