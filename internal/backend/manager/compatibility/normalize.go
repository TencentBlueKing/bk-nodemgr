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

package compatibility

import "strings"

// Normalize returns a normalized copy of the policy.
func (p Policy) Normalize() Policy {
	normalized := Policy{
		EnabledPlugins: normalizeStrings(p.EnabledPlugins),
		DisabledBiz:    make([]DisabledBiz, 0, len(p.DisabledBiz)),
	}

	seen := make(map[DisabledBiz]struct{}, len(p.DisabledBiz))
	for _, item := range p.DisabledBiz {
		normalizedItem := DisabledBiz{
			TenantID: strings.TrimSpace(item.TenantID),
			BKBizID:  item.BKBizID,
		}
		if _, ok := seen[normalizedItem]; ok {
			continue
		}
		seen[normalizedItem] = struct{}{}
		normalized.DisabledBiz = append(normalized.DisabledBiz, normalizedItem)
	}

	if normalized.EnabledPlugins == nil {
		normalized.EnabledPlugins = []string{}
	}
	if normalized.DisabledBiz == nil {
		normalized.DisabledBiz = []DisabledBiz{}
	}

	return normalized
}

func normalizeStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}

	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}

	return normalized
}
