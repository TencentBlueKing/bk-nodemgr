//go:build integration

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

package processconfig

import (
	"fmt"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/testsuite/support"
	"github.com/stretchr/testify/require"
)

func testClient(t *testing.T) IHandler {
	t.Helper()

	_, db := support.RequireMongoDatabase(t)
	return New(db)
}

func TestHandlerSameNameMainAndSubConfig(t *testing.T) {
	for _, subFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("sub-first-%t", subFirst), func(t *testing.T) {
			h := testClient(t)
			nCtx := contextx.From(contextx.Background(), contextx.WithTenantID("same-name"))
			main := &types.ProcessConfig{
				Name: "same.conf", ProcessName: "plugin", HostID: 1, IsMainConfig: true, Content: "main",
			}
			sub := &types.ProcessConfig{
				Name: "same.conf", ProcessName: "plugin", HostID: 1, IsMainConfig: false, Content: "sub",
			}
			configs := []*types.ProcessConfig{main, sub}
			if subFirst {
				configs[0], configs[1] = configs[1], configs[0]
			}
			for _, config := range configs {
				require.NoError(t, h.Create(nCtx, config))
			}
			require.Error(t, h.Create(nCtx, sub), "same-type duplicate must violate the unique index")
			for _, config := range []*types.ProcessConfig{main, sub} {
				got, err := h.Get(nCtx, 1, "plugin", "same.conf", config.IsMainConfig)
				require.NoError(t, err)
				require.Equal(t, config.IsMainConfig, got.IsMainConfig)
				require.Equal(t, config.Content, got.Content)
			}

			sub.Content = "updated sub"
			require.NoError(t, h.UpsertMany(nCtx, sub))
			for _, config := range []*types.ProcessConfig{main, sub} {
				got, count, err := h.List(nCtx, types.UnlimitedPage(),
					WithHostID(1), WithProcessName("plugin"), WithName("same.conf"), WithIsMainConfig(config.IsMainConfig))
				require.NoError(t, err)
				require.EqualValues(t, 1, count)
				require.Equal(t, config.Content, got[0].Content)
			}

			require.NoError(t, h.DeleteMany(nCtx,
				WithHostID(1), WithProcessName("plugin"), WithName("same.conf"), WithIsMainConfig(false)))
			got, err := h.Get(nCtx, 1, "plugin", "same.conf", true)
			require.NoError(t, err)
			require.True(t, got.IsMainConfig)
			require.Equal(t, "main", got.Content)
		})
	}
}

func prepareData(t *testing.T, nCtx contextx.IContext) IHandler {
	t.Helper()

	h := testClient(t)
	err := h.UpsertMany(nCtx,
		&types.ProcessConfig{
			ProcessName:  "test-process-name-1",
			HostID:       2,
			Name:         "test-process-config-name-1",
			Set:          "deploy_policy",
			IsMainConfig: true,
			Content:      "test-content-1",
			MD5:          "test-md5-1",
		},
		&types.ProcessConfig{
			ProcessName:  "test-process-name-2",
			HostID:       2,
			Name:         "test-process-config-name-2",
			Set:          "manual",
			IsMainConfig: false,
			Content:      "test-content-2",
			MD5:          "test-md5-2",
		},
		&types.ProcessConfig{
			ProcessName:  "test-process-name-3",
			HostID:       2,
			Name:         "test-process-config-name-3",
			IsMainConfig: false,
			Content:      "test-content-3",
			MD5:          "test-md5-3",
		},
	)
	if err != nil {
		t.Fatalf("prepareData() error = %v", err)
	}

	return h
}

func Test_handler_Create(t *testing.T) {
	nCtx := contextx.From(contextx.Background(), contextx.WithTenantID("test-tenant-id"))

	type args struct {
		nCtx   contextx.IContext
		config *types.ProcessConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "create process config",
			args: args{
				nCtx: nCtx,
				config: &types.ProcessConfig{
					ProcessName:  "test-process-name",
					HostID:       2,
					Name:         "test-process-config-name",
					IsMainConfig: false,
					MD5:          "test-md5",
					Content:      "test-content",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := prepareData(t, nCtx)
			if err := h.Create(tt.args.nCtx, tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("Create() success for name(%s), ProcessName(%s)", tt.args.config.Name, tt.args.config.ProcessName)
		})
	}
}

func Test_handler_UpsertMany(t *testing.T) {
	nCtx := contextx.From(contextx.Background(), contextx.WithTenantID("test-tenant-id"))

	type args struct {
		nCtx   contextx.IContext
		config []*types.ProcessConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "upsert many process config",
			args: args{
				nCtx: nCtx,
				config: []*types.ProcessConfig{
					{
						ProcessName:  "test-process-name-1",
						HostID:       2,
						Name:         "test-process-config-name-1",
						IsMainConfig: true,
						Content:      "updated-test-content-1",
						MD5:          "updated-test-md5-1",
					},
					{
						ProcessName:  "test-process-name-4",
						HostID:       2,
						Name:         "test-process-config-name-4",
						IsMainConfig: false,
						Content:      "test-content-4",
						MD5:          "test-md5-4",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := prepareData(t, nCtx)
			if err := h.UpsertMany(tt.args.nCtx, tt.args.config...); (err != nil) != tt.wantErr {
				t.Errorf("UpsertMany() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("UpsertMany() success")
		})
	}
}

func Test_handler_DeleteMany(t *testing.T) {
	nCtx := contextx.From(contextx.Background(), contextx.WithTenantID("test-tenant-id"))

	type args struct {
		nCtx contextx.IContext
		opts []base.OptFn
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "delete many process config",
			args: args{
				nCtx: nCtx,
				opts: []base.OptFn{
					WithName("test-process-config-name-2", "test-process-config-name-3"),
					WithProcessName("test-process-name-2", "test-process-name-3"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := prepareData(t, nCtx)
			if err := h.DeleteMany(tt.args.nCtx, tt.args.opts...); (err != nil) != tt.wantErr {
				t.Errorf("DeleteMany() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("DeleteMany() success")
		})
	}
}

func Test_handler_Get(t *testing.T) {
	nCtx := contextx.From(contextx.Background(), contextx.WithTenantID("test-tenant-id"))

	type args struct {
		nCtx         contextx.IContext
		name         string
		processName  string
		hostID       int64
		isMainConfig bool
	}
	tests := []struct {
		name    string
		args    args
		want    *types.ProcessConfig
		wantErr bool
	}{
		{
			name: "get main process config",
			args: args{
				nCtx:         nCtx,
				name:         "test-process-config-name-1",
				processName:  "test-process-name-1",
				hostID:       2,
				isMainConfig: true,
			},
			want: &types.ProcessConfig{
				ProcessName:  "test-process-name-1",
				HostID:       2,
				Name:         "test-process-config-name-1",
				Set:          "deploy_policy",
				IsMainConfig: true,
			},
			wantErr: false,
		},
		{
			name: "get sub process config",
			args: args{
				nCtx:         nCtx,
				name:         "test-process-config-name-2",
				processName:  "test-process-name-2",
				hostID:       2,
				isMainConfig: false,
			},
			want: &types.ProcessConfig{
				ProcessName:  "test-process-name-2",
				HostID:       2,
				Name:         "test-process-config-name-2",
				Set:          "manual",
				IsMainConfig: false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := prepareData(t, nCtx)
			got, err := h.Get(tt.args.nCtx, tt.args.hostID, tt.args.processName, tt.args.name, tt.args.isMainConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compareProcessConfig(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}

			t.Logf("Get() success")
		})
	}
}

func compareProcessConfig(a, b *types.ProcessConfig) bool {
	if a == nil || b == nil {
		return a == b
	}

	return a.Name == b.Name &&
		a.ProcessName == b.ProcessName &&
		a.IsMainConfig == b.IsMainConfig &&
		a.Set == b.Set
}

func Test_handler_List(t *testing.T) {
	nCtx := contextx.From(contextx.Background(), contextx.WithTenantID("test-tenant-id"))

	type args struct {
		nCtx contextx.IContext
		opts []base.OptFn
	}
	tests := []struct {
		name    string
		args    args
		wantLen int64
		wantErr bool
	}{
		{
			name: "list process config with no filters",
			args: args{
				nCtx: nCtx,
				opts: []base.OptFn{},
			},
			wantLen: 3,
			wantErr: false,
		},
		{
			name: "list process config with name filter",
			args: args{
				nCtx: nCtx,
				opts: []base.OptFn{
					WithName("test-process-config-name-1"),
				},
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "list process config with process name filter",
			args: args{
				nCtx: nCtx,
				opts: []base.OptFn{
					WithProcessName("test-process-name-1"),
				},
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "list process config with host id filter",
			args: args{
				nCtx: nCtx,
				opts: []base.OptFn{
					WithHostID(2),
				},
			},
			wantLen: 3,
			wantErr: false,
		},
		{
			name: "list process config with name and process name filter",
			args: args{
				nCtx: nCtx,
				opts: []base.OptFn{
					WithName("test-process-config-name-1"),
					WithProcessName("test-process-name-1"),
				},
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "list process config with is main config filter",
			args: args{
				nCtx: nCtx,
				opts: []base.OptFn{
					WithIsMainConfig(true),
				},
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "list process config with set filter",
			args: args{
				nCtx: nCtx,
				opts: []base.OptFn{
					WithSet("deploy_policy"),
				},
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "list process config without set filter",
			args: args{
				nCtx: nCtx,
				opts: []base.OptFn{
					WithoutSet("deploy_policy"),
				},
			},
			wantLen: 2,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := prepareData(t, nCtx)
			got, cnt, err := h.List(tt.args.nCtx, types.UnlimitedPage(), tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if cnt != tt.wantLen {
				t.Errorf("List() got len = %v, want %v", len(got), tt.wantLen)
			}

			t.Logf("List() success")
		})
	}
}
