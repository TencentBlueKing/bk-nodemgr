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

package workflow

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
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

type fakeActionDef struct {
	name  string
	queue action.Queue
}

func (def fakeActionDef) Name() string                       { return def.name }
func (def fakeActionDef) DisplayNameZh() string              { return def.name }
func (def fakeActionDef) DisplayNameEn() string              { return def.name }
func (def fakeActionDef) Version() string                    { return "1.0.0" }
func (def fakeActionDef) Description() string                { return def.name }
func (def fakeActionDef) Timeout() time.Duration             { return time.Minute }
func (def fakeActionDef) Tags() []action.Tag                 { return nil }
func (def fakeActionDef) Queue() action.Queue                { return def.queue }
func (def fakeActionDef) MaxRetryCount() uint                { return 0 }
func (def fakeActionDef) DelayFn(_ int) func()               { return nil }
func (def fakeActionDef) Do(_ *action.InstanceContext) error { return nil }

func Test_resolveConsumeQueueNames(t *testing.T) {
	newManager := func(consumeQueues []string, defs ...action.Definition) *manager {
		mgr := &manager{
			registeredActionDefs: make(map[string]action.Definition),
			consumeQueues:        consumeQueues,
		}
		for _, def := range defs {
			mgr.registeredActionDefs[def.Name()] = def
		}

		return mgr
	}

	tests := []struct {
		name          string
		consumeQueues []string
		defs          []action.Definition
		want          []string
	}{
		{
			name: "default consumes default queue only when no dedicated queue",
			defs: []action.Definition{fakeActionDef{name: "a", queue: action.QueueDefault}},
			want: []string{"default"},
		},
		{
			name: "default consumes default queue and all dedicated queues",
			defs: []action.Definition{
				fakeActionDef{name: "a", queue: action.QueueDefault},
				fakeActionDef{name: "b", queue: "ssh"},
				fakeActionDef{name: "c", queue: "credit"},
			},
			want: []string{"default", "ssh", "credit"},
		},
		{
			name: "default deduplicates dedicated queues",
			defs: []action.Definition{
				fakeActionDef{name: "a", queue: action.QueueDefault},
				fakeActionDef{name: "b", queue: "ssh"},
				fakeActionDef{name: "c", queue: "ssh"},
			},
			want: []string{"default", "ssh"},
		},
		{
			name:          "configured consumes only the configured queues",
			consumeQueues: []string{"ssh"},
			defs:          []action.Definition{fakeActionDef{name: "b", queue: "ssh"}},
			want:          []string{"ssh"},
		},
		{
			name:          "configured queues are consumed literally without registration check",
			consumeQueues: []string{"operation_inst_engine_queue", "ssh", "ssh"},
			defs:          []action.Definition{fakeActionDef{name: "b", queue: "ssh"}},
			want:          []string{"operation_inst_engine_queue", "ssh"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newManager(tt.consumeQueues, tt.defs...).resolveConsumeQueueNames()
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}
