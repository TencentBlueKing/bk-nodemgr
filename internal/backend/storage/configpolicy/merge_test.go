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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeepMergeConfig(t *testing.T) {
	tests := []struct {
		name    string
		base    map[string]any
		overlay map[string]any
		want    map[string]any
	}{
		{
			name:    "both nil",
			base:    nil,
			overlay: nil,
			want:    nil,
		},
		{
			name:    "nil base",
			base:    nil,
			overlay: map[string]any{"a": 1},
			want:    map[string]any{"a": 1},
		},
		{
			name:    "nil overlay",
			base:    map[string]any{"a": 1},
			overlay: nil,
			want:    map[string]any{"a": 1},
		},
		{
			name:    "flat override",
			base:    map[string]any{"a": 1, "b": "hello"},
			overlay: map[string]any{"b": "world", "c": true},
			want:    map[string]any{"a": 1, "b": "world", "c": true},
		},
		{
			name: "nested merge",
			base: map[string]any{
				"logging": map[string]any{
					"level": "info",
					"path":  "/var/log",
				},
				"timeout": 30,
			},
			overlay: map[string]any{
				"logging": map[string]any{
					"level": "debug",
				},
			},
			want: map[string]any{
				"logging": map[string]any{
					"level": "debug",
					"path":  "/var/log",
				},
				"timeout": 30,
			},
		},
		{
			name: "overlay replaces map with scalar",
			base: map[string]any{
				"a": map[string]any{"nested": true},
			},
			overlay: map[string]any{
				"a": "flat",
			},
			want: map[string]any{
				"a": "flat",
			},
		},
		{
			name: "overlay replaces scalar with map",
			base: map[string]any{
				"a": "flat",
			},
			overlay: map[string]any{
				"a": map[string]any{"nested": true},
			},
			want: map[string]any{
				"a": map[string]any{"nested": true},
			},
		},
		{
			name:    "empty maps",
			base:    map[string]any{},
			overlay: map[string]any{},
			want:    map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deepMergeConfig(tt.base, tt.overlay)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDeepMergeConfig_DoesNotMutateInputs(t *testing.T) {
	base := map[string]any{
		"a": 1,
		"nested": map[string]any{
			"x": 10,
		},
	}
	overlay := map[string]any{
		"a": 2,
		"nested": map[string]any{
			"y": 20,
		},
	}

	_ = deepMergeConfig(base, overlay)

	assert.Equal(t, map[string]any{"a": 1, "nested": map[string]any{"x": 10}}, base)
	assert.Equal(t, map[string]any{"a": 2, "nested": map[string]any{"y": 20}}, overlay)
}

func TestDeepMergeConfig_MultiLayer(t *testing.T) {
	defaultCfg := map[string]any{
		"log_level":   "info",
		"concurrency": 4,
		"paths": map[string]any{
			"data": "/data",
			"log":  "/var/log",
		},
	}
	lowPriority := map[string]any{
		"concurrency": 8,
		"paths": map[string]any{
			"log": "/tmp/log",
		},
	}
	highPriority := map[string]any{
		"log_level": "debug",
	}

	result := deepMergeConfig(defaultCfg, lowPriority)
	result = deepMergeConfig(result, highPriority)

	assert.Equal(t, map[string]any{
		"log_level":   "debug",
		"concurrency": 8,
		"paths": map[string]any{
			"data": "/data",
			"log":  "/tmp/log",
		},
	}, result)
}
