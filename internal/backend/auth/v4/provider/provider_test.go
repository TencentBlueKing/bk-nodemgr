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

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchInstanceFilter_UnmarshalJSON(t *testing.T) {
	for _, ids := range [][]string{{"id1", "id2"}, {"id1"}, {}, {"id1", "id2", "id3"}} {
		input, err := json.Marshal(FetchInstanceFilter{IDs: ids})
		require.NoError(t, err)
		var got FetchInstanceFilter
		require.NoError(t, json.Unmarshal(input, &got))
		assert.Equal(t, ids, got.IDs)
	}
}

func TestInstanceInfo_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		info InstanceInfo
		want string
	}{
		{"basic instance info", InstanceInfo{ID: "host-123", DisplayName: "127.0.0.1", Attributes: map[string]interface{}{"os": "Linux"}}, `{"display_name":"127.0.0.1","id":"host-123","os":"Linux"}`},
		{"instance without display_name", InstanceInfo{ID: "host-456", Attributes: map[string]interface{}{"country": "China"}}, `{"country":"China","id":"host-456"}`},
		{"instance with empty attributes", InstanceInfo{ID: "host-789", DisplayName: "test-host", Attributes: map[string]interface{}{}}, `{"display_name":"test-host","id":"host-789"}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.info.MarshalJSON()
			require.NoError(t, err)
			assert.JSONEq(t, test.want, string(got))
		})
	}
}

func TestInstanceInfo_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input string
		want  InstanceInfo
	}{
		{`{"id":"host-123","display_name":"127.0.0.1","os":"Linux"}`, InstanceInfo{ID: "host-123", DisplayName: "127.0.0.1", Attributes: map[string]interface{}{"os": "Linux"}}},
		{`{"id":"host-456","country":"China"}`, InstanceInfo{ID: "host-456", Attributes: map[string]interface{}{"country": "China"}}},
		{`{"id":"host-789","_bk_iam_path_":["/biz,1/set,1/"]}`, InstanceInfo{ID: "host-789", Attributes: map[string]interface{}{AttrIAMPath: []interface{}{"/biz,1/set,1/"}}}},
	}
	for _, test := range tests {
		var got InstanceInfo
		require.NoError(t, json.Unmarshal([]byte(test.input), &got))
		assert.Equal(t, test.want, got)
	}
}

func TestFetchInstanceFilter_IDsLimit(t *testing.T) {
	handler := NewHandler()
	handler.RegisterProvider(ResourceTypePackageType, NewPackageTypeProvider())
	for _, count := range []int{1000, 500, 1001, 2000} {
		ids := make([]string, count)
		for index := range ids {
			ids[index] = fmt.Sprintf("id-%d", index)
		}
		_, err := handler.FetchInstanceInfo(contextx.New(context.Background()), ResourceTypePackageType, &Request[FetchInstanceFilter]{Filter: FetchInstanceFilter{IDs: ids}})
		if count > MaxFetchInstanceIDs {
			require.ErrorIs(t, err, ErrInvalidArgument)
			continue
		}
		require.NoError(t, err)
	}
}

type mockPluginStorage struct {
	distinctPluginNames        []string
	distinctPluginErr          error
	distinctPluginBinToolNames []string
	distinctPluginBinToolErr   error
}

func (m *mockPluginStorage) DistinctNameReleasePlugin(_ contextx.IContext, _ ...*types.ReleaseCondition) ([]string, error) {
	return m.distinctPluginNames, m.distinctPluginErr
}

func (m *mockPluginStorage) DistinctNameReleasePluginBinTool(_ contextx.IContext, _ ...*types.ReleaseCondition) ([]string, error) {
	return m.distinctPluginBinToolNames, m.distinctPluginBinToolErr
}

func TestPackageProvider_ListInstance_NilParent(t *testing.T) {
	provider := NewPackageProvider(&mockPluginStorage{})
	result, err := provider.ListInstance(contextx.New(context.Background()), &Request[ListInstanceFilter]{Page: types.Page{Limit: 10}})
	require.NoError(t, err)
	assert.Equal(t, int64(4), result.Count)
	assert.Len(t, result.Results, 4)
}

func TestPackageProvider_listPluginInstances(t *testing.T) {
	names := []string{"plugin-j", "plugin-a", "plugin-c", "plugin-b", "plugin-d", "plugin-e", "plugin-f", "plugin-g", "plugin-h", "plugin-i"}
	tests := []struct {
		name       string
		names      []string
		storageErr error
		page       types.Page
		count      int64
		length     int
		first      string
	}{
		{"first page", names, nil, types.Page{Limit: 5}, 10, 5, "plugin-a"},
		{"second page", names, nil, types.Page{Offset: 5, Limit: 5}, 10, 5, "plugin-f"},
		{"beyond total", names, nil, types.Page{Offset: 10, Limit: 5}, 10, 0, ""},
		{"empty names", []string{}, nil, types.Page{Limit: 5}, 0, 0, ""},
		{"storage error", nil, fmt.Errorf("storage error"), types.Page{Limit: 5}, 0, 0, ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := NewPackageProvider(&mockPluginStorage{distinctPluginNames: test.names, distinctPluginErr: test.storageErr})
			result, err := provider.ListInstance(contextx.New(context.Background()), &Request[ListInstanceFilter]{
				Filter: ListInstanceFilter{Parent: &ParentFilter{Type: ResourceTypePackageType, ID: string(types.ReleaseTypePlugin)}}, Page: test.page,
			})
			if test.storageErr != nil {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.count, result.Count)
			assert.Len(t, result.Results, test.length)
			if test.length > 0 {
				assert.Equal(t, test.first, result.Results[0].ID)
			}
		})
	}
}

func TestPackageProvider_searchPluginInstances(t *testing.T) {
	tests := []struct {
		name       string
		names      []string
		keyword    string
		storageErr error
		wantIDs    []string
	}{
		{"keyword matches", []string{"plugin-a", "plugin-b", "tool-c"}, "plugin", nil, []string{"plugin-a", "plugin-b"}},
		{"no matches", []string{"plugin-a", "plugin-b", "tool-c"}, "xyz", nil, []string{}},
		{"empty keyword", []string{"plugin-a", "plugin-b", "tool-c"}, "", nil, []string{"plugin-a", "plugin-b", "tool-c"}},
		{"case insensitive", []string{"Plugin-A", "PLUGIN-B", "tool-c"}, "plugin", nil, []string{"PLUGIN-B", "Plugin-A"}},
		{"storage error", nil, "plugin", fmt.Errorf("storage error"), nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := NewPackageProvider(&mockPluginStorage{distinctPluginNames: test.names, distinctPluginErr: test.storageErr})
			result, err := provider.ListInstance(contextx.New(context.Background()), &Request[ListInstanceFilter]{
				Filter: ListInstanceFilter{Parent: &ParentFilter{Type: ResourceTypePackageType, ID: string(types.ReleaseTypePlugin)}, Keyword: test.keyword}, Page: types.Page{Limit: 10},
			})
			if test.storageErr != nil {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, int64(len(test.wantIDs)), result.Count)
			ids := make([]string, 0, len(result.Results))
			for _, result := range result.Results {
				ids = append(ids, result.ID)
			}
			assert.Equal(t, test.wantIDs, ids)
		})
	}
}

func TestPackageProvider_FetchInstanceInfo_ErrorPropagation(t *testing.T) {
	tests := []struct {
		name       string
		ids        []string
		pluginErr  error
		binToolErr error
	}{
		{"plugin storage error", []string{"plugin-a"}, fmt.Errorf("plugin storage error"), nil},
		{"bintool storage error", []string{"bintool-a"}, nil, fmt.Errorf("bintool storage error")},
		{"no error", []string{"plugin-a"}, nil, nil},
		{"empty ids", []string{}, nil, nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := NewPackageProvider(&mockPluginStorage{distinctPluginErr: test.pluginErr, distinctPluginBinToolErr: test.binToolErr})
			_, err := provider.FetchInstanceInfo(contextx.New(context.Background()), &Request[FetchInstanceFilter]{Filter: FetchInstanceFilter{IDs: test.ids}})
			if test.pluginErr != nil {
				require.ErrorIs(t, err, test.pluginErr)
				return
			}
			if test.binToolErr != nil {
				require.ErrorIs(t, err, test.binToolErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
