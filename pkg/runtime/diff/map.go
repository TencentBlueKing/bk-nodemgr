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

package diff

// CompareMaps compares the maps and returns the added, deleted and changed items' keys.
// nolint: nonamedreturns
func CompareMaps[K comparable, V any](original map[K]V, latest map[K]V, compareFunc func(oldValue, newValue V) bool,
) (added map[K]V, deleted map[K]V, changed map[K]V) {

	added = make(map[K]V)
	deleted = make(map[K]V)
	changed = make(map[K]V)

	for key, newValue := range latest {
		if oldValue, exists := original[key]; exists {
			// the key exists in both maps, check if they are equal.
			if !compareFunc(oldValue, newValue) {
				changed[key] = newValue
			}
		} else {
			// the key only exists in the latest map, it's added.
			added[key] = newValue
		}
	}

	// check keys only exists in original map
	for key := range original {
		if _, exists := latest[key]; !exists {
			// the key only exists in the original map, it's deleted.
			deleted[key] = original[key]
		}
	}

	return added, deleted, changed
}
