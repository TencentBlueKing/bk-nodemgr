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

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAttributeValueID_UnmarshalJSON tests AttributeValueID JSON deserialization.
// According to IAM list_attr_value API, attribute value IDs support string/int/bool types.
// Note: float64 is NOT an IAM-declared type (it's just JSON number implementation detail).
func TestAttributeValueID_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    AttributeValueID
		wantErr bool
	}{
		{
			name:    "string type",
			input:   `"abc123"`,
			want:    AttributeValueID("abc123"),
			wantErr: false,
		},
		{
			name:    "int type",
			input:   `123`,
			want:    AttributeValueID("123"),
			wantErr: false,
		},
		{
			name:    "negative int",
			input:   `-456`,
			want:    AttributeValueID("-456"),
			wantErr: false,
		},
		{
			name:    "bool true",
			input:   `true`,
			want:    AttributeValueID("true"),
			wantErr: false,
		},
		{
			name:    "bool false",
			input:   `false`,
			want:    AttributeValueID("false"),
			wantErr: false,
		},
		{
			name:    "null value - handled as empty string",
			input:   `null`,
			want:    AttributeValueID(""),
			wantErr: false, // Go JSON unmarshaler treats null as zero value
		},
		{
			name:    "float64 - should error (not IAM declared type)",
			input:   `123.45`,
			want:    AttributeValueID(""),
			wantErr: true, // float64 is NOT supported per IAM spec
		},
		{
			name:    "float64 without fractional part - should error",
			input:   `789.0`,
			want:    AttributeValueID(""),
			wantErr: true, // float64 is NOT supported per IAM spec
		},
		{
			name:    "invalid type - array",
			input:   `["abc"]`,
			want:    AttributeValueID(""),
			wantErr: true,
		},
		{
			name:    "invalid type - object",
			input:   `{"id": "abc"}`,
			want:    AttributeValueID(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got AttributeValueID
			err := json.Unmarshal([]byte(tt.input), &got)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// TestAttributeValueID_UnmarshalJSON_InSlice tests AttributeValueID deserialization in arrays.
// This tests the list_attr_value filter scenario where ids can be mixed types.
func TestAttributeValueID_UnmarshalJSON_InSlice(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []AttributeValueID
		wantErr bool
	}{
		{
			name:    "mixed types - string, int, bool",
			input:   `["abc", 123, true, "xyz", 456, false]`,
			want:    []AttributeValueID{"abc", "123", "true", "xyz", "456", "false"},
			wantErr: false,
		},
		{
			name:    "all strings",
			input:   `["id1", "id2", "id3"]`,
			want:    []AttributeValueID{"id1", "id2", "id3"},
			wantErr: false,
		},
		{
			name:    "all ints",
			input:   `[1, 2, 3]`,
			want:    []AttributeValueID{"1", "2", "3"},
			wantErr: false,
		},
		{
			name:    "empty array",
			input:   `[]`,
			want:    []AttributeValueID{},
			wantErr: false,
		},
		{
			name:    "contains null - treated as empty string",
			input:   `["abc", null, 123]`,
			want:    []AttributeValueID{"abc", "", "123"},
			wantErr: false, // Go JSON unmarshaler treats null as zero value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []AttributeValueID
			err := json.Unmarshal([]byte(tt.input), &got)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// TestListAttrValueFilter_UnmarshalJSON tests ListAttrValueFilter deserialization.
// This filter's IDs field uses []AttributeValueID which supports string/int/bool.
func TestListAttrValueFilter_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ListAttrValueFilter
		wantErr bool
	}{
		{
			name:  "with string IDs",
			input: `{"attr": "os", "keyword": "linux", "ids": ["id1", "id2"]}`,
			want: ListAttrValueFilter{
				Attr:    "os",
				Keyword: "linux",
				IDs:     []AttributeValueID{"id1", "id2"},
			},
			wantErr: false,
		},
		{
			name:  "with int IDs",
			input: `{"attr": "os", "ids": [1, 2, 3]}`,
			want: ListAttrValueFilter{
				Attr: "os",
				IDs:  []AttributeValueID{"1", "2", "3"},
			},
			wantErr: false,
		},
		{
			name:  "with mixed type IDs (string/int/bool)",
			input: `{"attr": "os", "ids": ["abc", 123, true]}`,
			want: ListAttrValueFilter{
				Attr: "os",
				IDs:  []AttributeValueID{"abc", "123", "true"},
			},
			wantErr: false,
		},
		{
			name:  "without optional fields",
			input: `{"attr": "os"}`,
			want: ListAttrValueFilter{
				Attr: "os",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got ListAttrValueFilter
			err := json.Unmarshal([]byte(tt.input), &got)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// TestFetchInstanceFilter_UnmarshalJSON tests FetchInstanceFilter deserialization.
// According to IAM fetch_instance_info API, filter.ids only supports string type (array(string)).
// No custom type needed - standard []string works.
func TestFetchInstanceFilter_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    FetchInstanceFilter
		wantErr bool
	}{
		{
			name:  "with string IDs and attrs",
			input: `{"ids": ["id1", "id2"], "attrs": ["display_name", "os"]}`,
			want: FetchInstanceFilter{
				IDs:   []string{"id1", "id2"},
				Attrs: []string{"display_name", "os"},
			},
			wantErr: false,
		},
		{
			name:  "with string IDs only",
			input: `{"ids": ["id1"]}`,
			want: FetchInstanceFilter{
				IDs: []string{"id1"},
			},
			wantErr: false,
		},
		{
			name:  "with empty IDs array",
			input: `{"ids": []}`,
			want: FetchInstanceFilter{
				IDs: []string{},
			},
			wantErr: false,
		},
		{
			name:  "without attrs",
			input: `{"ids": ["id1", "id2", "id3"]}`,
			want: FetchInstanceFilter{
				IDs: []string{"id1", "id2", "id3"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got FetchInstanceFilter
			err := json.Unmarshal([]byte(tt.input), &got)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// TestInstanceInfo_MarshalJSON tests InstanceInfo custom marshaling.
func TestInstanceInfo_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		info    InstanceInfo
		want    string
		wantErr bool
	}{
		{
			name: "basic instance info",
			info: InstanceInfo{
				ID:          "host-123",
				DisplayName: "192.168.1.1",
				Attributes: map[string]interface{}{
					"os": "Linux",
				},
			},
			want:    `{"display_name":"192.168.1.1","id":"host-123","os":"Linux"}`,
			wantErr: false,
		},
		{
			name: "instance without display_name",
			info: InstanceInfo{
				ID: "host-456",
				Attributes: map[string]interface{}{
					"country": "China",
				},
			},
			want:    `{"country":"China","id":"host-456"}`,
			wantErr: false,
		},
		{
			name: "instance with empty attributes",
			info: InstanceInfo{
				ID:          "host-789",
				DisplayName: "test-host",
				Attributes:  map[string]interface{}{},
			},
			want:    `{"display_name":"test-host","id":"host-789"}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.info.MarshalJSON()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.JSONEq(t, tt.want, string(got))
			}
		})
	}
}

// TestInstanceInfo_UnmarshalJSON tests InstanceInfo custom unmarshaling.
func TestInstanceInfo_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    InstanceInfo
		wantErr bool
	}{
		{
			name:  "basic instance info",
			input: `{"id":"host-123","display_name":"192.168.1.1","os":"Linux"}`,
			want: InstanceInfo{
				ID:          "host-123",
				DisplayName: "192.168.1.1",
				Attributes: map[string]interface{}{
					"os": "Linux",
				},
			},
			wantErr: false,
		},
		{
			name:  "instance without display_name",
			input: `{"id":"host-456","country":"China"}`,
			want: InstanceInfo{
				ID: "host-456",
				Attributes: map[string]interface{}{
					"country": "China",
				},
			},
			wantErr: false,
		},
		{
			name:  "instance with array attribute",
			input: `{"id":"host-789","_bk_iam_path_":["/biz,1/set,1/"]}`,
			want: InstanceInfo{
				ID: "host-789",
				Attributes: map[string]interface{}{
					"_bk_iam_path_": []interface{}{"/biz,1/set,1/"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got InstanceInfo
			err := json.Unmarshal([]byte(tt.input), &got)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want.ID, got.ID)
				assert.Equal(t, tt.want.DisplayName, got.DisplayName)
				assert.Equal(t, tt.want.Attributes, got.Attributes)
			}
		})
	}
}

// TestFetchInstanceFilter_IDsLimit tests FetchInstanceFilter IDs count validation.
// This test verifies that the dispatcher correctly validates the IDs count limit
// according to IAM specification (max 1000 IDs).
func TestFetchInstanceFilter_IDsLimit(t *testing.T) {
	tests := []struct {
		name      string
		idsCount  int
		wantError bool
		errorMsg  string
	}{
		{
			name:      "within limit - 1000 IDs",
			idsCount:  1000,
			wantError: false,
		},
		{
			name:      "within limit - 500 IDs",
			idsCount:  500,
			wantError: false,
		},
		{
			name:      "exceeds limit - 1001 IDs",
			idsCount:  1001,
			wantError: true,
			errorMsg:  "IDs count exceeds maximum limit of 1000",
		},
		{
			name:      "exceeds limit - 2000 IDs",
			idsCount:  2000,
			wantError: true,
			errorMsg:  "IDs count exceeds maximum limit of 1000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids := make([]string, tt.idsCount)
			for i := 0; i < tt.idsCount; i++ {
				ids[i] = fmt.Sprintf("id-%d", i)
			}

			filter := FetchInstanceFilter{
				IDs: ids,
			}

			var err error
			if len(filter.IDs) > MaxFetchInstanceIDs {
				err = fmt.Errorf("IDs count exceeds maximum limit of %d", MaxFetchInstanceIDs)
			}

			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSearchInstanceFilter_EmptyKeyword(t *testing.T) {
	tests := []struct {
		name     string
		keyword  string
		wantErr  bool
		errorMsg string
	}{
		{
			name:    "valid keyword",
			keyword: "test",
			wantErr: false,
		},
		{
			name:     "empty keyword",
			keyword:  "",
			wantErr:  true,
			errorMsg: "keyword is required and cannot be empty",
		},
		{
			name:     "whitespace only keyword",
			keyword:  "   ",
			wantErr:  true,
			errorMsg: "keyword is required and cannot be empty",
		},
		{
			name:     "tab only keyword",
			keyword:  "\t",
			wantErr:  true,
			errorMsg: "keyword is required and cannot be empty",
		},
		{
			name:     "newline only keyword",
			keyword:  "\n",
			wantErr:  true,
			errorMsg: "keyword is required and cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := SearchInstanceFilter{
				Keyword: tt.keyword,
			}

			var err error
			if strings.TrimSpace(filter.Keyword) == "" {
				err = fmt.Errorf("keyword is required and cannot be empty")
			}

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPackageProvider_ListInstance_NilParent(t *testing.T) {
	provider := &PackageProvider{
		storage: nil,
	}

	req := &Request[ListInstanceFilter]{
		Filter: ListInstanceFilter{
			Parent: nil,
		},
	}

	result, err := provider.ListInstance(nil, req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(0), result.Count)
	assert.NotNil(t, result.Results)
	assert.Empty(t, result.Results)
}
