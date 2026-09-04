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

package dpmgr

import (
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestGetTaskConfigSource(t *testing.T) {
	tests := []struct {
		name       string
		task       *ChangeTask
		wantHostID int64
	}{
		{
			name: "empty task",
		},
		{
			name: "same host only source is omitted",
			task: &ChangeTask{
				Target:       &types.Target{Host: types.Host{HostID: 1}},
				ConfigSource: &types.Target{Host: types.Host{HostID: 1}},
			},
		},
		{
			name: "same host service source is kept",
			task: &ChangeTask{
				Target: &types.Target{Host: types.Host{HostID: 1}},
				ConfigSource: &types.Target{
					Host:            types.Host{HostID: 1},
					ServiceInstance: types.ServiceInstance{ID: 2},
				},
			},
			wantHostID: 1,
		},
		{
			name: "same host topo relation source is kept",
			task: &ChangeTask{
				Target: &types.Target{Host: types.Host{HostID: 1}},
				ConfigSource: &types.Target{
					Host:                 types.Host{HostID: 1},
					MatchedTopoRelations: []types.TargetMatchedTopoRelation{{TopoObjID: "module", TopoInstID: 3}},
				},
			},
			wantHostID: 1,
		},
		{
			name: "different host source is kept",
			task: &ChangeTask{
				Target:       &types.Target{Host: types.Host{HostID: 1}},
				ConfigSource: &types.Target{Host: types.Host{HostID: 2}},
			},
			wantHostID: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getTaskConfigSource(tt.task)
			if got.Host.HostID != tt.wantHostID {
				t.Fatalf("getTaskConfigSource() HostID = %d, want %d", got.Host.HostID, tt.wantHostID)
			}
		})
	}
}
