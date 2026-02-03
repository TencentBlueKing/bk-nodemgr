/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
)

func Test_buildAddrsWithPassword(t *testing.T) {
	tests := []struct {
		name      string
		redisConf config.Redis
		want      []string
	}{
		{
			name:      "empty addrs returns nil",
			redisConf: config.Redis{Addrs: []string{}},
			want:      nil,
		},
		{
			name: "standalone mode single addr no password",
			redisConf: config.Redis{
				Type:  config.RedisTypeStandalone,
				Addrs: []string{"localhost:6379"},
			},
			want: []string{"localhost:6379"},
		},
		{
			name: "standalone mode single addr with password",
			redisConf: config.Redis{
				Type:     config.RedisTypeStandalone,
				Addrs:    []string{"localhost:6379"},
				Password: "secret",
			},
			want: []string{"secret@localhost:6379"},
		},
		{
			name: "cluster mode single addr no password - should duplicate",
			redisConf: config.Redis{
				Type:  config.RedisTypeCluster,
				Addrs: []string{"node1:6379"},
			},
			want: []string{"node1:6379", "node1:6379"},
		},
		{
			name: "cluster mode single addr with password - should duplicate",
			redisConf: config.Redis{
				Type:     config.RedisTypeCluster,
				Addrs:    []string{"node1:6379"},
				Password: "secret",
			},
			want: []string{"secret@node1:6379", "node1:6379"},
		},
		{
			name: "cluster mode multiple addrs no password",
			redisConf: config.Redis{
				Type:  config.RedisTypeCluster,
				Addrs: []string{"node1:6379", "node2:6379", "node3:6379"},
			},
			want: []string{"node1:6379", "node2:6379", "node3:6379"},
		},
		{
			name: "cluster mode multiple addrs with password",
			redisConf: config.Redis{
				Type:     config.RedisTypeCluster,
				Addrs:    []string{"node1:6379", "node2:6379"},
				Password: "secret",
			},
			want: []string{"secret@node1:6379", "node2:6379"},
		},
		{
			name: "sentinel mode with password",
			redisConf: config.Redis{
				Type:       config.RedisTypeSentinel,
				Addrs:      []string{"sentinel1:26379", "sentinel2:26379"},
				Password:   "secret",
				MasterName: "mymaster",
			},
			want: []string{"secret@sentinel1:26379", "sentinel2:26379"},
		},
		{
			name: "password with @ symbol",
			redisConf: config.Redis{
				Type:     config.RedisTypeStandalone,
				Addrs:    []string{"localhost:6379"},
				Password: "pass@word",
			},
			want: []string{"pass@word@localhost:6379"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildAddrsWithPassword(tt.redisConf)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_buildAddrsWithPassword_doesNotModifyOriginal(t *testing.T) {
	original := []string{"node1:6379"}
	redisConf := config.Redis{
		Type:     config.RedisTypeCluster,
		Addrs:    original,
		Password: "secret",
	}

	result := buildAddrsWithPassword(redisConf)

	// Original slice should not be modified
	assert.Equal(t, []string{"node1:6379"}, original)
	assert.Equal(t, []string{"secret@node1:6379", "node1:6379"}, result)
}
