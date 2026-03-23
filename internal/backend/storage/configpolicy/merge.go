/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package configpolicy provides the config policy storage interface.
package configpolicy

import "maps"

// deepMergeConfig merges overlay into base recursively.
func deepMergeConfig(base, overlay map[string]any) map[string]any {
	if base == nil && overlay == nil {
		return nil
	}

	result := make(map[string]any, len(base)+len(overlay))
	maps.Copy(result, base)

	for k, ov := range overlay {
		bMap, bOK := result[k].(map[string]any)
		oMap, oOK := ov.(map[string]any)
		if bOK && oOK {
			result[k] = deepMergeConfig(bMap, oMap)
		} else {
			result[k] = ov
		}
	}

	return result
}
