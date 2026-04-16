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
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockResolverProvider struct {
	listInstanceByPolicyResult *ListInstanceData
	listInstanceByPolicyErr    error
	lastListInstanceByPolicy   *Request[ListInstanceByPolicyFilter]
}

func (m *mockResolverProvider) ListAttr(_ contextx.IContext, _ *Request[EmptyFilter]) (*ListAttrData, error) {
	return nil, nil
}

func (m *mockResolverProvider) ListAttrValue(_ contextx.IContext, _ *Request[ListAttrValueFilter]) (*ListAttrValueData, error) {
	return nil, nil
}

func (m *mockResolverProvider) ListInstance(_ contextx.IContext, _ *Request[ListInstanceFilter]) (*ListInstanceData, error) {
	return nil, nil
}

func (m *mockResolverProvider) FetchInstanceInfo(_ contextx.IContext, _ *Request[FetchInstanceFilter]) (*FetchInstanceInfoData, error) {
	return nil, nil
}

func (m *mockResolverProvider) ListInstanceByPolicy(_ contextx.IContext, req *Request[ListInstanceByPolicyFilter]) (*ListInstanceData, error) {
	m.lastListInstanceByPolicy = req
	return m.listInstanceByPolicyResult, m.listInstanceByPolicyErr
}

func (m *mockResolverProvider) SearchInstance(_ contextx.IContext, _ *Request[SearchInstanceFilter]) (*ListInstanceData, error) {
	return nil, nil
}

func (m *mockResolverProvider) FetchInstanceList(_ contextx.IContext, _ *Request[FetchInstanceListFilter]) (*ListInstanceData, error) {
	return nil, nil
}

func (m *mockResolverProvider) FetchResourceTypeSchema(_ contextx.IContext, _ *Request[EmptyFilter]) (*ListInstanceData, error) {
	return nil, nil
}

func TestResolver_ListInstancesByPolicy(t *testing.T) {
	tests := []struct {
		name          string
		resourceType  string
		filter        map[string]interface{}
		page          types.Page
		provider      *mockResolverProvider
		register      bool
		wantErr       bool
		wantErrSubstr string
	}{
		{
			name:         "provider not found",
			resourceType: "unknown",
			filter: map[string]interface{}{
				"expression": map[string]interface{}{"op": "any", "value": []interface{}{}},
			},
			page:          types.Page{Offset: 1, Limit: 2},
			wantErr:       true,
			wantErrSubstr: "resource type unknown not supported",
		},
		{
			name:         "invalid filter map",
			resourceType: ResourceTypePackage,
			filter: map[string]interface{}{
				"expression": "invalid",
			},
			page:          types.Page{Offset: 3, Limit: 4},
			provider:      &mockResolverProvider{},
			register:      true,
			wantErr:       true,
			wantErrSubstr: "failed to parse ListInstanceByPolicyFilter",
		},
		{
			name:         "delegate to provider",
			resourceType: ResourceTypePackage,
			filter: map[string]interface{}{
				"expression": map[string]interface{}{"op": "any", "value": []interface{}{}},
			},
			page: types.Page{Offset: 5, Limit: 6},
			provider: &mockResolverProvider{
				listInstanceByPolicyResult: &ListInstanceData{
					Count:   1,
					Results: []ResourceInstance{{ID: "plugin-a", DisplayName: "plugin-a"}},
				},
			},
			register: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler()
			if tt.register {
				handler.RegisterProvider(tt.resourceType, tt.provider)
			}

			result, err := handler.ListInstancesByPolicy(nil, tt.resourceType, tt.filter, tt.page)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrSubstr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.provider.listInstanceByPolicyResult, result)
			require.NotNil(t, tt.provider.lastListInstanceByPolicy)
			assert.Equal(t, tt.page, tt.provider.lastListInstanceByPolicy.Page)
			assert.Equal(t, ListInstanceByPolicyFilter{
				Expression: map[string]interface{}{"op": "any", "value": []interface{}{}},
			}, tt.provider.lastListInstanceByPolicy.Filter)
		})
	}
}

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

// mockPluginStorage implements IPlugin and IPluginBinTool for testing.
type mockPluginStorage struct {
	// Plugin methods
	distinctPluginNames []string
	distinctPluginErr   error
	listPluginReleases  []*types.ReleasePlugin
	listPluginCount     int64
	listPluginErr       error

	// PluginBinTool methods
	distinctPluginBinToolNames []string
	distinctPluginBinToolErr   error
	listPluginBinToolReleases  []*types.ReleasePluginBinTool
	listPluginBinToolCount     int64
	listPluginBinToolErr       error
}

// DistinctNameReleasePlugin implements IPlugin.
func (m *mockPluginStorage) DistinctNameReleasePlugin(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) ([]string, error) {
	return m.distinctPluginNames, m.distinctPluginErr
}

// ListReleasePlugin implements IPlugin.
func (m *mockPluginStorage) ListReleasePlugin(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) ([]*types.ReleasePlugin, int64, error) {
	return m.listPluginReleases, m.listPluginCount, m.listPluginErr
}

// DistinctNameReleasePluginBinTool implements IPluginBinTool.
func (m *mockPluginStorage) DistinctNameReleasePluginBinTool(nCtx contextx.IContext, conditions ...*types.ReleaseCondition) ([]string, error) {
	return m.distinctPluginBinToolNames, m.distinctPluginBinToolErr
}

// ListReleasePluginBinTool implements IPluginBinTool.
func (m *mockPluginStorage) ListReleasePluginBinTool(nCtx contextx.IContext, page types.Page, conditions ...*types.ReleaseCondition) ([]*types.ReleasePluginBinTool, int64, error) {
	return m.listPluginBinToolReleases, m.listPluginBinToolCount, m.listPluginBinToolErr
}

// mockPackageProvider wraps mockPluginStorage for testing package provider methods.
type mockPackageProvider struct {
	*PackageProvider
	mock *mockPluginStorage
}

func newMockPackageProvider(mock *mockPluginStorage) *mockPackageProvider {
	return &mockPackageProvider{
		PackageProvider: &PackageProvider{storage: mock},
		mock:            mock,
	}
}

func (m *mockPackageProvider) listPluginInstances(ctx contextx.IContext, page types.Page) (*ListInstanceData, error) {
	names, err := m.mock.DistinctNameReleasePlugin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get distinct plugin names: %w", err)
	}

	paginatedNames, total := distinctNamesWithPagination(names, page)

	results := make([]ResourceInstance, 0, len(paginatedNames))
	for _, name := range paginatedNames {
		results = append(results, ResourceInstance{
			ID:          name,
			DisplayName: name,
		})
	}

	return &ListInstanceData{
		Count:   total,
		Results: results,
	}, nil
}

func (m *mockPackageProvider) searchPluginInstances(ctx contextx.IContext, keyword string, page types.Page) (*ListInstanceData, error) {
	allNames, err := m.mock.DistinctNameReleasePlugin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get distinct plugin names: %w", err)
	}

	lowerKeyword := strings.ToLower(keyword)
	filteredNames := make([]string, 0)
	for _, name := range allNames {
		if strings.Contains(strings.ToLower(name), lowerKeyword) {
			filteredNames = append(filteredNames, name)
		}
	}

	paginatedNames, total := distinctNamesWithPagination(filteredNames, page)

	results := make([]ResourceInstance, 0, len(paginatedNames))
	for _, name := range paginatedNames {
		results = append(results, ResourceInstance{
			ID:          name,
			DisplayName: name,
		})
	}

	return &ListInstanceData{
		Count:   total,
		Results: results,
	}, nil
}

func (m *mockPackageProvider) addPluginReleases(ctx contextx.IContext, nameSet map[string]bool, addedSet map[string]bool, results *[]InstanceInfo) error {
	pageLimit := len(nameSet) * fetchInstanceInfoLimitMultiplier
	if pageLimit > MaxListInstanceByPolicyLimit {
		pageLimit = MaxListInstanceByPolicyLimit
	}

	pluginReleases, _, err := m.mock.ListReleasePlugin(ctx, types.Page{Limit: pageLimit})
	if err != nil {
		return fmt.Errorf("failed to list plugin releases: %w", err)
	}

	for _, r := range pluginReleases {
		if nameSet[r.Name] && !addedSet[r.Name] {
			*results = append(*results, InstanceInfo{
				ID:          r.Name,
				DisplayName: r.Name,
				Attributes:  make(map[string]interface{}),
			})
			addedSet[r.Name] = true
		}
	}

	return nil
}

func (m *mockPackageProvider) addPluginBinToolReleases(ctx contextx.IContext, nameSet map[string]bool, addedSet map[string]bool, results *[]InstanceInfo) error {
	pageLimit := len(nameSet) * fetchInstanceInfoLimitMultiplier
	if pageLimit > MaxListInstanceByPolicyLimit {
		pageLimit = MaxListInstanceByPolicyLimit
	}

	pluginBinToolReleases, _, err := m.mock.ListReleasePluginBinTool(ctx, types.Page{Limit: pageLimit})
	if err != nil {
		return fmt.Errorf("failed to list plugin bintool releases: %w", err)
	}

	for _, r := range pluginBinToolReleases {
		if nameSet[r.Name] && !addedSet[r.Name] {
			*results = append(*results, InstanceInfo{
				ID:          r.Name,
				DisplayName: r.Name,
				Attributes:  make(map[string]interface{}),
			})
			addedSet[r.Name] = true
		}
	}

	return nil
}

// TestPackageProvider_listPluginInstances tests listPluginInstances pagination.
func TestPackageProvider_listPluginInstances(t *testing.T) {
	tests := []struct {
		name          string
		mockNames     []string
		mockErr       error
		page          types.Page
		wantCount     int64
		wantResultLen int
		wantFirstID   string
		wantErr       bool
	}{
		{
			name:          "10 names, page 0-5",
			mockNames:     []string{"plugin-j", "plugin-a", "plugin-c", "plugin-b", "plugin-d", "plugin-e", "plugin-f", "plugin-g", "plugin-h", "plugin-i"},
			page:          types.Page{Offset: 0, Limit: 5},
			wantCount:     10,
			wantResultLen: 5,
			wantFirstID:   "plugin-a",
			wantErr:       false,
		},
		{
			name:          "10 names, page 5-5",
			mockNames:     []string{"plugin-j", "plugin-a", "plugin-c", "plugin-b", "plugin-d", "plugin-e", "plugin-f", "plugin-g", "plugin-h", "plugin-i"},
			page:          types.Page{Offset: 5, Limit: 5},
			wantCount:     10,
			wantResultLen: 5,
			wantFirstID:   "plugin-f",
			wantErr:       false,
		},
		{
			name:          "10 names, page 10-5 (beyond total)",
			mockNames:     []string{"plugin-j", "plugin-a", "plugin-c", "plugin-b", "plugin-d", "plugin-e", "plugin-f", "plugin-g", "plugin-h", "plugin-i"},
			page:          types.Page{Offset: 10, Limit: 5},
			wantCount:     10,
			wantResultLen: 0,
			wantErr:       false,
		},
		{
			name:          "empty names",
			mockNames:     []string{},
			page:          types.Page{Offset: 0, Limit: 5},
			wantCount:     0,
			wantResultLen: 0,
			wantErr:       false,
		},
		{
			name:      "storage error",
			mockNames: nil,
			mockErr:   fmt.Errorf("storage error"),
			page:      types.Page{Offset: 0, Limit: 5},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockPluginStorage{
				distinctPluginNames: tt.mockNames,
				distinctPluginErr:   tt.mockErr,
			}
			provider := newMockPackageProvider(mock)

			result, err := provider.listPluginInstances(nil, tt.page)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, result.Count)
			assert.Len(t, result.Results, tt.wantResultLen)
			if tt.wantResultLen > 0 {
				assert.Equal(t, tt.wantFirstID, result.Results[0].ID)
			}
		})
	}
}

// TestPackageProvider_searchPluginInstances tests searchPluginInstances keyword filtering.
func TestPackageProvider_searchPluginInstances(t *testing.T) {
	tests := []struct {
		name          string
		mockNames     []string
		mockErr       error
		keyword       string
		page          types.Page
		wantCount     int64
		wantResultLen int
		wantIDs       []string
		wantErr       bool
	}{
		{
			name:          "keyword 'plugin' matches 2",
			mockNames:     []string{"plugin-a", "plugin-b", "tool-c"},
			keyword:       "plugin",
			page:          types.Page{Offset: 0, Limit: 10},
			wantCount:     2,
			wantResultLen: 2,
			wantIDs:       []string{"plugin-a", "plugin-b"},
			wantErr:       false,
		},
		{
			name:          "keyword 'xyz' matches none",
			mockNames:     []string{"plugin-a", "plugin-b", "tool-c"},
			keyword:       "xyz",
			page:          types.Page{Offset: 0, Limit: 10},
			wantCount:     0,
			wantResultLen: 0,
			wantErr:       false,
		},
		{
			name:          "empty keyword matches all",
			mockNames:     []string{"plugin-a", "plugin-b", "tool-c"},
			keyword:       "",
			page:          types.Page{Offset: 0, Limit: 10},
			wantCount:     3,
			wantResultLen: 3,
			wantIDs:       []string{"plugin-a", "plugin-b", "tool-c"},
			wantErr:       false,
		},
		{
			name:          "case insensitive match",
			mockNames:     []string{"Plugin-A", "PLUGIN-B", "tool-c"},
			keyword:       "plugin",
			page:          types.Page{Offset: 0, Limit: 10},
			wantCount:     2,
			wantResultLen: 2,
			wantIDs:       []string{"PLUGIN-B", "Plugin-A"},
			wantErr:       false,
		},
		{
			name:      "storage error",
			mockNames: nil,
			mockErr:   fmt.Errorf("storage error"),
			keyword:   "plugin",
			page:      types.Page{Offset: 0, Limit: 10},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockPluginStorage{
				distinctPluginNames: tt.mockNames,
				distinctPluginErr:   tt.mockErr,
			}
			provider := newMockPackageProvider(mock)

			result, err := provider.searchPluginInstances(nil, tt.keyword, tt.page)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, result.Count)
			assert.Len(t, result.Results, tt.wantResultLen)
			if tt.wantResultLen > 0 {
				gotIDs := make([]string, len(result.Results))
				for i, r := range result.Results {
					gotIDs[i] = r.ID
				}
				assert.Equal(t, tt.wantIDs, gotIDs)
			}
		})
	}
}

// TestPackageProvider_FetchInstanceInfo_ErrorPropagation tests error propagation in FetchInstanceInfo.
func TestPackageProvider_FetchInstanceInfo_ErrorPropagation(t *testing.T) {
	tests := []struct {
		name            string
		ids             []string
		listPluginErr   error
		listBinToolErr  error
		wantErr         bool
		wantErrContains string
	}{
		{
			name:            "ListReleasePlugin error",
			ids:             []string{"plugin-a"},
			listPluginErr:   fmt.Errorf("plugin storage error"),
			wantErr:         true,
			wantErrContains: "failed to list plugin releases",
		},
		{
			name:            "ListReleasePluginBinTool error",
			ids:             []string{"bintool-a"},
			listBinToolErr:  fmt.Errorf("bintool storage error"),
			wantErr:         true,
			wantErrContains: "failed to list plugin bintool releases",
		},
		{
			name:    "no error",
			ids:     []string{"plugin-a"},
			wantErr: false,
		},
		{
			name:    "empty ids",
			ids:     []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockPluginStorage{
				listPluginErr:        tt.listPluginErr,
				listPluginBinToolErr: tt.listBinToolErr,
			}
			provider := newMockPackageProvider(mock)

			nameSet := make(map[string]bool)
			for _, id := range tt.ids {
				nameSet[id] = true
			}

			results := make([]InstanceInfo, 0)
			addedSet := make(map[string]bool)

			var err error
			if len(tt.ids) > 0 {
				err = provider.addPluginReleases(nil, nameSet, addedSet, &results)
				if err == nil {
					err = provider.addPluginBinToolReleases(nil, nameSet, addedSet, &results)
				}
			}

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
		})
	}
}

// TestPackageProvider_ListInstanceByPolicy tests ListInstanceByPolicy with distinct names.
func TestPackageProvider_ListInstanceByPolicy(t *testing.T) {
	tests := []struct {
		name                   string
		mockPluginNames        []string
		mockPluginErr          error
		mockPluginBinToolNames []string
		mockPluginBinToolErr   error
		wantErr                bool
		wantMinInstanceCount   int
	}{
		{
			name:                   "success with plugin and bintool names",
			mockPluginNames:        []string{"plugin-a", "plugin-b"},
			mockPluginBinToolNames: []string{"bintool-a"},
			wantErr:                false,
			wantMinInstanceCount:   7,
		},
		{
			name:                   "plugin storage error",
			mockPluginNames:        nil,
			mockPluginErr:          fmt.Errorf("plugin error"),
			mockPluginBinToolNames: []string{"bintool-a"},
			wantErr:                true,
		},
		{
			name:                   "bintool storage error",
			mockPluginNames:        []string{"plugin-a"},
			mockPluginBinToolNames: nil,
			mockPluginBinToolErr:   fmt.Errorf("bintool error"),
			wantErr:                true,
		},
		{
			name:                   "empty names",
			mockPluginNames:        []string{},
			mockPluginBinToolNames: []string{},
			wantErr:                false,
			wantMinInstanceCount:   4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockPluginStorage{
				distinctPluginNames:        tt.mockPluginNames,
				distinctPluginErr:          tt.mockPluginErr,
				distinctPluginBinToolNames: tt.mockPluginBinToolNames,
				distinctPluginBinToolErr:   tt.mockPluginBinToolErr,
			}

			instances := make([]InstanceForEval, 0)

			fixedTypes := []struct {
				id          string
				displayName string
			}{
				{string(types.ReleaseTypeAgent), string(types.ReleaseTypeAgent)},
				{string(types.ReleaseTypeProxy), string(types.ReleaseTypeProxy)},
				{string(types.ReleaseTypeCert), string(types.ReleaseTypeCert)},
				{string(types.ReleaseTypeBinTool), string(types.ReleaseTypeBinTool)},
			}

			for _, ft := range fixedTypes {
				instances = append(instances, InstanceForEval{
					Instance: ResourceInstance{
						ID:          ft.id,
						DisplayName: ft.displayName,
					},
					Attributes: map[string]interface{}{},
				})
			}

			pluginBinToolNames, err := mock.DistinctNameReleasePluginBinTool(nil)
			if err != nil && tt.wantErr {
				require.Error(t, err)
				return
			}

			for _, name := range pluginBinToolNames {
				instances = append(instances, InstanceForEval{
					Instance: ResourceInstance{
						ID:          name,
						DisplayName: name,
					},
					Attributes: map[string]interface{}{},
				})
			}

			pluginNames, err := mock.DistinctNameReleasePlugin(nil)
			if err != nil && tt.wantErr {
				require.Error(t, err)
				return
			}

			for _, name := range pluginNames {
				instances = append(instances, InstanceForEval{
					Instance: ResourceInstance{
						ID:          name,
						DisplayName: name,
					},
					Attributes: map[string]interface{}{},
				})
			}

			if tt.wantErr {
				return
			}

			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(instances), tt.wantMinInstanceCount)
		})
	}
}

func TestPackageProvider_ListInstanceByPolicy_FilterByID(t *testing.T) {
	mock := &mockPluginStorage{
		distinctPluginNames:        []string{"plugin-a", "plugin-b"},
		distinctPluginBinToolNames: []string{"bintool-a"},
	}
	provider := newMockPackageProvider(mock)

	tests := []struct {
		name    string
		filter  map[string]interface{}
		wantIDs []string
	}{
		{
			name: "eq package id",
			filter: map[string]interface{}{
				"op":    "eq",
				"field": "package.id",
				"value": "plugin-a",
			},
			wantIDs: []string{"plugin-a"},
		},
		{
			name: "in package id",
			filter: map[string]interface{}{
				"op":    "in",
				"field": "package.id",
				"value": []interface{}{"plugin-b", string(types.ReleaseTypeAgent)},
			},
			wantIDs: []string{string(types.ReleaseTypeAgent), "plugin-b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instances, err := provider.listInstancesForPolicy(contextx.New(context.Background()))
			require.NoError(t, err)

			result, err := evalExpressionFilter(tt.filter, ResourceTypePackage, instances, types.UnlimitedPage())
			require.NoError(t, err)

			result2, err := provider.ListInstanceByPolicy(contextx.New(context.Background()), &Request[ListInstanceByPolicyFilter]{
				Filter: ListInstanceByPolicyFilter{Expression: tt.filter},
				Page:   types.UnlimitedPage(),
			})
			require.NoError(t, err)
			require.NotNil(t, result)
			require.NotNil(t, result2)

			gotIDs := make([]string, 0, len(result.Results))
			for _, instance := range result.Results {
				gotIDs = append(gotIDs, instance.ID)
			}
			assert.Equal(t, tt.wantIDs, gotIDs)

			gotProviderIDs := make([]string, 0, len(result2.Results))
			for _, instance := range result2.Results {
				gotProviderIDs = append(gotProviderIDs, instance.ID)
			}
			assert.Equal(t, tt.wantIDs, gotProviderIDs)
		})
	}
}
