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

package cmdb

import (
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func TestBuildModuleMatchedTopoRelationsByModuleID(t *testing.T) {
	relationsByModuleID := buildModuleMatchedTopoRelationsByModuleID([]int64{101, 101, 102})

	if len(relationsByModuleID) != 2 {
		t.Fatalf("expected 2 module relations, got %d", len(relationsByModuleID))
	}
	assertTargetMatchedTopoRelation(t, relationsByModuleID[101], TopoNodeObjIDModule, 101)
	assertTargetMatchedTopoRelation(t, relationsByModuleID[102], TopoNodeObjIDModule, 102)
}

func TestTopoNodeMatchesModulePaths(t *testing.T) {
	modulePaths := [][]*Node{
		{
			{BKObjID: TopoNodeObjIDBiz, BKInstID: 2},
			{BKObjID: TopoNodeObjIDSet, BKInstID: 3},
			{BKObjID: "custom", BKInstID: 4},
			{BKObjID: TopoNodeObjIDModule, BKInstID: 5},
		},
	}

	cases := []struct {
		name     string
		topoNode *types.ScopeTopoNode
		matched  bool
	}{
		{name: "biz", topoNode: &types.ScopeTopoNode{TopoObjID: TopoNodeObjIDBiz, TopoInstID: 2}, matched: true},
		{name: "set", topoNode: &types.ScopeTopoNode{TopoObjID: TopoNodeObjIDSet, TopoInstID: 3}, matched: true},
		{name: "custom", topoNode: &types.ScopeTopoNode{TopoObjID: "custom", TopoInstID: 4}, matched: true},
		{name: "module", topoNode: &types.ScopeTopoNode{TopoObjID: TopoNodeObjIDModule, TopoInstID: 5}, matched: true},
		{name: "wrong set", topoNode: &types.ScopeTopoNode{TopoObjID: TopoNodeObjIDSet, TopoInstID: 6}, matched: false},
		{name: "nil", matched: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			matched := topoNodeMatchesModulePaths(2, 5, modulePaths, tc.topoNode)
			if matched != tc.matched {
				t.Fatalf("expected matched %v, got %v", tc.matched, matched)
			}
		})
	}
}

func TestAssignAndFilterMatchedTopoRelationsByModuleID(t *testing.T) {
	targets := []*types.Target{
		{ServiceInstance: types.ServiceInstance{ModuleID: 101}},
		{ServiceInstance: types.ServiceInstance{ModuleID: 102}},
		{ServiceInstance: types.ServiceInstance{ModuleID: 103}},
		nil,
	}
	relationsByModuleID := map[int64][]types.TargetMatchedTopoRelation{
		101: {{TopoObjID: TopoNodeObjIDModule, TopoInstID: 101}},
		102: {{TopoObjID: TopoNodeObjIDSet, TopoInstID: 201}},
	}

	assignMatchedTopoRelationsByModuleID(targets, relationsByModuleID)

	assertTargetMatchedTopoRelation(t, targets[0].MatchedTopoRelations, TopoNodeObjIDModule, 101)
	assertTargetMatchedTopoRelation(t, targets[1].MatchedTopoRelations, TopoNodeObjIDSet, 201)
	if len(targets[2].MatchedTopoRelations) != 0 {
		t.Fatalf("expected no relations for unmatched module, got %v", targets[2].MatchedTopoRelations)
	}

	filteredTargets := filterTargetsWithMatchedTopoRelations(targets)
	if len(filteredTargets) != 2 {
		t.Fatalf("expected 2 filtered targets, got %d", len(filteredTargets))
	}
}

func assertTargetMatchedTopoRelation(
	t *testing.T,
	relations []types.TargetMatchedTopoRelation,
	topoObjID string,
	topoInstID int64,
) {
	t.Helper()
	if len(relations) != 1 {
		t.Fatalf("expected one relation, got %d", len(relations))
	}
	if relations[0].TopoObjID != topoObjID || relations[0].TopoInstID != topoInstID {
		t.Fatalf("expected relation %s/%d, got %s/%d",
			topoObjID, topoInstID, relations[0].TopoObjID, relations[0].TopoInstID)
	}
}
