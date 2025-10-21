/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package diff

import (
	"reflect"
	"testing"
)

// TestCompareMaps ...
func TestCompareMaps(t *testing.T) {
	tests := []struct {
		name        string
		original    map[string]int
		latest      map[string]int
		compareFunc func(oldValue, newValue int) bool
		wantAdded   map[string]int
		wantDeleted map[string]int
		wantChanged map[string]int
	}{
		{
			name:     "empty maps",
			original: map[string]int{},
			latest:   map[string]int{},
			compareFunc: func(oldValue, newValue int) bool {
				return oldValue == newValue
			},
			wantAdded:   map[string]int{},
			wantDeleted: map[string]int{},
			wantChanged: map[string]int{},
		},
		{
			name:     "all added",
			original: map[string]int{},
			latest:   map[string]int{"a": 1, "b": 2, "c": 3},
			compareFunc: func(oldValue, newValue int) bool {
				return oldValue == newValue
			},
			wantAdded:   map[string]int{"a": 1, "b": 2, "c": 3},
			wantDeleted: map[string]int{},
			wantChanged: map[string]int{},
		},
		{
			name:     "all deleted",
			original: map[string]int{"a": 1, "b": 2, "c": 3},
			latest:   map[string]int{},
			compareFunc: func(oldValue, newValue int) bool {
				return oldValue == newValue
			},
			wantAdded:   map[string]int{},
			wantDeleted: map[string]int{"a": 1, "b": 2, "c": 3},
			wantChanged: map[string]int{},
		},
		{
			name:     "no changes",
			original: map[string]int{"a": 1, "b": 2, "c": 3},
			latest:   map[string]int{"a": 1, "b": 2, "c": 3},
			compareFunc: func(oldValue, newValue int) bool {
				return oldValue == newValue
			},
			wantAdded:   map[string]int{},
			wantDeleted: map[string]int{},
			wantChanged: map[string]int{},
		},
		{
			name:     "some changed",
			original: map[string]int{"a": 1, "b": 2, "c": 3},
			latest:   map[string]int{"a": 1, "b": 20, "c": 3, "d": 4},
			compareFunc: func(oldValue, newValue int) bool {
				return oldValue == newValue
			},
			wantAdded:   map[string]int{"d": 4},
			wantDeleted: map[string]int{},
			wantChanged: map[string]int{"b": 20},
		},
		{
			name:     "mixed changes",
			original: map[string]int{"a": 1, "b": 2, "c": 3, "d": 4},
			latest:   map[string]int{"b": 20, "c": 3, "e": 5, "f": 6},
			compareFunc: func(oldValue, newValue int) bool {
				return oldValue == newValue
			},
			wantAdded:   map[string]int{"e": 5, "f": 6},
			wantDeleted: map[string]int{"a": 1, "d": 4},
			wantChanged: map[string]int{"b": 20},
		},
		{
			name:     "nil maps",
			original: nil,
			latest:   nil,
			compareFunc: func(oldValue, newValue int) bool {
				return oldValue == newValue
			},
			wantAdded:   map[string]int{},
			wantDeleted: map[string]int{},
			wantChanged: map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			added, deleted, changed := CompareMaps(tt.original, tt.latest, tt.compareFunc)

			if !reflect.DeepEqual(added, tt.wantAdded) {
				t.Errorf("CompareMaps() added = %v, want %v", added, tt.wantAdded)
			}

			if !reflect.DeepEqual(deleted, tt.wantDeleted) {
				t.Errorf("CompareMaps() deleted = %v, want %v", deleted, tt.wantDeleted)
			}

			if !reflect.DeepEqual(changed, tt.wantChanged) {
				t.Errorf("CompareMaps() changed = %v, want %v", changed, tt.wantChanged)
			}
		})
	}
}

// TestCompareMaps_StringValues ...
func TestCompareMaps_StringValues(t *testing.T) {
	original := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	latest := map[string]string{
		"key1": "value1",
		"key2": "new_value2",
		"key4": "value4",
	}

	compareFunc := func(oldValue, newValue string) bool {
		return oldValue == newValue
	}

	added, deleted, changed := CompareMaps(original, latest, compareFunc)

	expectedAdded := map[string]string{"key4": "value4"}
	expectedDeleted := map[string]string{"key3": "value3"}
	expectedChanged := map[string]string{"key2": "new_value2"}

	if !reflect.DeepEqual(added, expectedAdded) {
		t.Errorf("CompareMaps() added = %v, want %v", added, expectedAdded)
	}

	if !reflect.DeepEqual(deleted, expectedDeleted) {
		t.Errorf("CompareMaps() deleted = %v, want %v", deleted, expectedDeleted)
	}

	if !reflect.DeepEqual(changed, expectedChanged) {
		t.Errorf("CompareMaps() changed = %v, want %v", changed, expectedChanged)
	}
}
