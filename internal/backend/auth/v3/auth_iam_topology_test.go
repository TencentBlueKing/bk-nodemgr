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

package v3

import (
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// TestBuildResourceInstancePath_WithParent tests building topology path for resources with parent.
func TestBuildResourceInstancePath_WithParent(t *testing.T) {
	tests := []struct {
		name        string
		resource    types.AuthResource
		wantPathLen int
		wantParent  types.IAMApplyResourceNode
		wantCurrent types.IAMApplyResourceNode
	}{
		{
			name: "networkunit with networkarea parent",
			resource: types.AuthResource{
				SystemID: types.SystemIDNodeMgr,
				Type:     types.AuthResourceTypeNetworkUnit,
				ID:       "456",
				Attributes: map[string]interface{}{
					"_bk_iam_path_": []string{"/networkarea,123/"},
				},
			},
			wantPathLen: 2,
			wantParent: types.IAMApplyResourceNode{
				Type: "networkarea",
				ID:   "123",
			},
			wantCurrent: types.IAMApplyResourceNode{
				Type: "networkunit",
				ID:   "456",
			},
		},
		{
			name: "package with package_type parent",
			resource: types.AuthResource{
				SystemID: types.SystemIDNodeMgr,
				Type:     types.AuthResourceTypePackage,
				ID:       "my-plugin-v1.0",
				Attributes: map[string]interface{}{
					"_bk_iam_path_": []string{"/package_type,plugin/"},
				},
			},
			wantPathLen: 2,
			wantParent: types.IAMApplyResourceNode{
				Type: "package_type",
				ID:   "plugin",
			},
			wantCurrent: types.IAMApplyResourceNode{
				Type: "package",
				ID:   "my-plugin-v1.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := buildResourceInstancePath(tt.resource)
			if len(path) != tt.wantPathLen {
				t.Fatalf("path length = %d, want %d", len(path), tt.wantPathLen)
			}
			if path[0] != tt.wantParent {
				t.Errorf("parent node = %+v, want %+v", path[0], tt.wantParent)
			}
			if path[1] != tt.wantCurrent {
				t.Errorf("current node = %+v, want %+v", path[1], tt.wantCurrent)
			}
		})
	}
}

// TestBuildResourceInstancePath_WithoutParent tests building path for resources without parent.
func TestBuildResourceInstancePath_WithoutParent(t *testing.T) {
	tests := []struct {
		name        string
		resource    types.AuthResource
		wantPathLen int
		wantNode    types.IAMApplyResourceNode
	}{
		{
			name: "networkarea without parent",
			resource: types.AuthResource{
				SystemID: types.SystemIDNodeMgr,
				Type:     types.AuthResourceTypeNetworkArea,
				ID:       "123",
			},
			wantPathLen: 1,
			wantNode: types.IAMApplyResourceNode{
				Type: "networkarea",
				ID:   "123",
			},
		},
		{
			name: "biz without parent",
			resource: types.AuthResource{
				SystemID: types.SystemIDCMDB,
				Type:     types.AuthResourceTypeBiz,
				ID:       "42",
			},
			wantPathLen: 1,
			wantNode: types.IAMApplyResourceNode{
				Type: "biz",
				ID:   "42",
			},
		},
		{
			name: "resource with nil attributes",
			resource: types.AuthResource{
				SystemID:   types.SystemIDNodeMgr,
				Type:       types.AuthResourceTypeNetworkArea,
				ID:         "789",
				Attributes: nil,
			},
			wantPathLen: 1,
			wantNode: types.IAMApplyResourceNode{
				Type: "networkarea",
				ID:   "789",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := buildResourceInstancePath(tt.resource)
			if len(path) != tt.wantPathLen {
				t.Fatalf("path length = %d, want %d", len(path), tt.wantPathLen)
			}
			if path[0] != tt.wantNode {
				t.Errorf("node = %+v, want %+v", path[0], tt.wantNode)
			}
		})
	}
}

// TestBuildIAMApplyResourceTypes_WithTopology tests complete topology path construction.
func TestBuildIAMApplyResourceTypes_WithTopology(t *testing.T) {
	rts := buildIAMApplyResourceTypes([]types.AuthResource{
		{
			SystemID: types.SystemIDNodeMgr,
			Type:     types.AuthResourceTypeNetworkUnit,
			ID:       "456",
			Attributes: map[string]interface{}{
				"_bk_iam_path_": []string{"/networkarea,123/"},
			},
		},
	})

	if len(rts) != 1 {
		t.Fatalf("expected 1 resource type, got %d", len(rts))
	}

	rt := rts[0]
	if rt.Type != string(types.AuthResourceTypeNetworkUnit) {
		t.Fatalf("expected type networkunit, got %q", rt.Type)
	}

	if len(rt.Instances) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(rt.Instances))
	}

	instance := rt.Instances[0]
	if len(instance) != 2 {
		t.Fatalf("expected instance path length 2, got %d", len(instance))
	}

	// Verify parent node
	if instance[0].Type != "networkarea" || instance[0].ID != "123" {
		t.Errorf("parent node = %+v, want {Type: networkarea, ID: 123}", instance[0])
	}

	// Verify current node
	if instance[1].Type != "networkunit" || instance[1].ID != "456" {
		t.Errorf("current node = %+v, want {Type: networkunit, ID: 456}", instance[1])
	}
}

// TestBuildIAMApplyResourceTypes_MixedTopology tests mixed resources with and without parent.
func TestBuildIAMApplyResourceTypes_MixedTopology(t *testing.T) {
	rts := buildIAMApplyResourceTypes([]types.AuthResource{
		// Resource with parent
		{
			SystemID: types.SystemIDNodeMgr,
			Type:     types.AuthResourceTypeNetworkUnit,
			ID:       "456",
			Attributes: map[string]interface{}{
				"_bk_iam_path_": []string{"/networkarea,123/"},
			},
		},
		// Resource without parent
		{
			SystemID: types.SystemIDCMDB,
			Type:     types.AuthResourceTypeBiz,
			ID:       "42",
		},
	})

	if len(rts) != 2 {
		t.Fatalf("expected 2 resource types, got %d", len(rts))
	}

	// Find networkunit resource type
	var networkunitRT *types.IAMApplyResourceType
	for i := range rts {
		if rts[i].Type == string(types.AuthResourceTypeNetworkUnit) {
			networkunitRT = &rts[i]
			break
		}
	}
	if networkunitRT == nil {
		t.Fatal("networkunit resource type not found")
	}

	// Verify networkunit has 2-node path
	if len(networkunitRT.Instances) != 1 || len(networkunitRT.Instances[0]) != 2 {
		t.Errorf("networkunit instance path length = %d, want 2", len(networkunitRT.Instances[0]))
	}

	// Find biz resource type
	var bizRT *types.IAMApplyResourceType
	for i := range rts {
		if rts[i].Type == string(types.AuthResourceTypeBiz) {
			bizRT = &rts[i]
			break
		}
	}
	if bizRT == nil {
		t.Fatal("biz resource type not found")
	}

	// Verify biz has 1-node path
	if len(bizRT.Instances) != 1 || len(bizRT.Instances[0]) != 1 {
		t.Errorf("biz instance path length = %d, want 1", len(bizRT.Instances[0]))
	}
}
