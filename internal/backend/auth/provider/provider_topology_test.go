/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package provider

import "testing"

// TestParseParentFromIAMPath_ValidPath tests parsing valid _bk_iam_path_ attribute.
func TestParseParentFromIAMPath_ValidPath(t *testing.T) {
	tests := []struct {
		name       string
		attributes map[string]interface{}
		wantType   string
		wantID     string
	}{
		{
			name: "networkarea parent",
			attributes: map[string]interface{}{
				"_bk_iam_path_": []string{"/networkarea,123/"},
			},
			wantType: "networkarea",
			wantID:   "123",
		},
		{
			name: "package_type parent",
			attributes: map[string]interface{}{
				"_bk_iam_path_": []string{"/package_type,plugin/"},
			},
			wantType: "package_type",
			wantID:   "plugin",
		},
		{
			name: "path as []interface{}",
			attributes: map[string]interface{}{
				"_bk_iam_path_": []interface{}{"/networkarea,456/"},
			},
			wantType: "networkarea",
			wantID:   "456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := ParseParentFromIAMPath(tt.attributes)
			if node == nil {
				t.Fatal("expected non-nil node, got nil")
			}
			if node.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", node.Type, tt.wantType)
			}
			if node.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", node.ID, tt.wantID)
			}
		})
	}
}

// TestParseParentFromIAMPath_InvalidCases tests error handling for invalid inputs.
func TestParseParentFromIAMPath_InvalidCases(t *testing.T) {
	tests := []struct {
		name       string
		attributes map[string]interface{}
	}{
		{
			name:       "nil attributes",
			attributes: nil,
		},
		{
			name:       "missing _bk_iam_path_",
			attributes: map[string]interface{}{},
		},
		{
			name: "empty path array",
			attributes: map[string]interface{}{
				"_bk_iam_path_": []string{},
			},
		},
		{
			name: "invalid path format - no slashes",
			attributes: map[string]interface{}{
				"_bk_iam_path_": []string{"networkarea,123"},
			},
		},
		{
			name: "invalid path format - no comma",
			attributes: map[string]interface{}{
				"_bk_iam_path_": []string{"/networkarea123/"},
			},
		},
		{
			name: "invalid path format - too many parts",
			attributes: map[string]interface{}{
				"_bk_iam_path_": []string{"/networkarea,123,extra/"},
			},
		},
		{
			name: "wrong type",
			attributes: map[string]interface{}{
				"_bk_iam_path_": "not-an-array",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := ParseParentFromIAMPath(tt.attributes)
			if node != nil {
				t.Errorf("expected nil node, got %+v", node)
			}
		})
	}
}
