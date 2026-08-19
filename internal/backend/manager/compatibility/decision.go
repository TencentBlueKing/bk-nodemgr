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

// DecideCompatibilityMode returns whether compatibility mode should be enabled for a plugin.
func DecideCompatibilityMode(policy Policy, tenantID string, bkBizID int64, pluginName string) bool {
	normalized := policy.Normalize()
	pluginName = strings.TrimSpace(pluginName)
	tenantID = strings.TrimSpace(tenantID)

	if isDisabledBiz(normalized.DisabledBiz, tenantID, bkBizID) {
		return false
	}

	for _, enabledPlugin := range normalized.EnabledPlugins {
		if enabledPlugin == pluginName {
			return true
		}
	}

	return false
}

func isDisabledBiz(disabledBiz []DisabledBiz, tenantID string, bkBizID int64) bool {
	for _, item := range disabledBiz {
		if item.TenantID == tenantID && item.BKBizID == bkBizID {
			return true
		}
	}

	return false
}
