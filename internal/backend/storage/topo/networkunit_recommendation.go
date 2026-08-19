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

package topo

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"strconv"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func parseNetworkUnitSegmentRuleConfig(value string) (types.NetworkUnitSegmentRuleConfig, error) {
	var cfg types.NetworkUnitSegmentRuleConfig
	if err := json.Unmarshal([]byte(value), &cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func recommendNetworkUnitsBySegment(
	rules types.NetworkUnitSegmentRuleConfig,
	items []*types.NetworkUnitSegmentRecommendationItem,
	networkUnitAreaMap map[int64]int64,
) ([]*types.NetworkUnitSegmentRecommendationResult, error) {

	result := make([]*types.NetworkUnitSegmentRecommendationResult, len(items))

	for idx, item := range items {
		result[idx] = recommendOneNetworkUnitBySegment(rules, item, networkUnitAreaMap)
	}

	return result, nil
}

func recommendOneNetworkUnitBySegment(
	rules types.NetworkUnitSegmentRuleConfig,
	item *types.NetworkUnitSegmentRecommendationItem,
	networkUnitAreaMap map[int64]int64,
) *types.NetworkUnitSegmentRecommendationResult {

	if item == nil {
		return &types.NetworkUnitSegmentRecommendationResult{
			NetworkUnitID: -1,
			Message:       "empty recommendation item",
		}
	}

	result := &types.NetworkUnitSegmentRecommendationResult{
		NetworkAreaID: item.NetworkAreaID,
		IP:            item.IP,
		NetworkUnitID: -1,
	}

	ip, err := netip.ParseAddr(item.IP)
	if err != nil {
		result.Message = "invalid ip"
		return result
	}
	ip = ip.Unmap()
	if !ip.Is4() {
		result.Message = "invalid ip"
		return result
	}

	areaRules, ok := rules[strconv.FormatInt(item.NetworkAreaID, 10)]
	if !ok || len(areaRules.Rules) == 0 {
		result.Message = "no rule matched"
		return result
	}

	for _, rule := range areaRules.Rules {
		matched, err := matchRuleCIDRs(rule.CIDRs, ip)
		if err != nil {
			result.Message = fmt.Sprintf("invalid rule cidr: %v", err)
			return result
		}
		if !matched {
			continue
		}

		if networkUnitAreaMap[rule.NetworkUnitID] != item.NetworkAreaID {
			result.Message = "invalid recommended network unit"

			return result
		}

		result.NetworkUnitID = rule.NetworkUnitID
		result.Message = "matched"

		return result
	}

	result.Message = "no rule matched"

	return result
}

func matchRuleCIDRs(cidrs []string, ip netip.Addr) (bool, error) {
	for _, cidr := range cidrs {
		prefix, err := netip.ParsePrefix(cidr)
		if err != nil {
			return false, err
		}
		if prefix.Masked().Contains(ip) {
			return true, nil
		}
	}

	return false, nil
}
