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

package plugin

import (
	"context"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var instance struct {
	TenantID string
	sync.Once
}

func tenantID() string {
	instance.Once.Do(func() {
		instance.TenantID = uuid.NewString()
	})

	return instance.TenantID
}

// testClient creates a test handler instance
func testClient(t *testing.T) IHandler {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	nCtx := context.Background()
	mongoClient, err := mongo.Connect(
		nCtx,
		&options.ClientOptions{
			Hosts: []string{
				os.Getenv("MONGO_ADDRESS"),
			},
			Auth: &options.Credential{
				Username:      os.Getenv("MONGO_USER"),
				Password:      os.Getenv("MONGO_PASSWORD"),
				AuthSource:    os.Getenv("MONGO_AUTH_SOURCE"),
				AuthMechanism: os.Getenv("MONGO_AUTH_MECHANISM"),
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	return New(mongoClient.Database(os.Getenv("MONGO_DATABASE")))
}

// TestHandler_Create test handler Create
func TestHandler_Create(t *testing.T) {
	type args struct {
		nCtx   contextx.IContext
		plugin *types.Plugin
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal test",
			args: args{
				nCtx: contextx.New(context.Background(), contextx.WithTenantID(tenantID())),
				plugin: &types.Plugin{
					TenantID: tenantID(),
					Name:     "test-plugin",
					Group:    "test-group",
					PkgName:  "test-plugin-pkg",
				},
			},
			wantErr: false,
		},
		{
			name: "missing tenant ID",
			args: args{
				nCtx: contextx.New(context.Background()),
				plugin: &types.Plugin{
					Name:    "test-plugin",
					Group:   "test-group",
					PkgName: "test-plugin-pkg",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			if err := h.Create(tt.args.nCtx, tt.args.plugin); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestHandler_Count test handler Count
func TestHandler_Count(t *testing.T) {
	h := testClient(t)
	testTenantID := tenantID()

	// Create a test plugin first
	nCtx := contextx.New(context.Background(), contextx.WithTenantID(testTenantID))
	testPlugin := &types.Plugin{
		TenantID: testTenantID,
		Name:     "test-plugin-count",
		Group:    "test-group",
		PkgName:  "test-plugin-pkg",
	}

	err := h.Create(nCtx, testPlugin)
	if err != nil {
		t.Fatalf("Failed to create test plugin: %v", err)
	}

	type args struct {
		nCtx contextx.IContext
		opts []OptFn
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "count all plugins",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "count by plugin name",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{WithName("test-plugin-count")},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "count by group",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{WithGroup("test-group")},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "count by non-existent plugin",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{WithName("non-existent-plugin")},
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "missing tenant ID",
			args: args{
				nCtx: contextx.New(context.Background()),
				opts: []OptFn{},
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.Count(tt.args.nCtx, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Count() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Count() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestHandler_List test handler List
func TestHandler_List(t *testing.T) {
	h := testClient(t)
	testTenantID := tenantID()
	// Create test plugins first
	nCtx := contextx.New(context.Background(), contextx.WithTenantID(testTenantID))
	testPlugin1 := &types.Plugin{
		TenantID: testTenantID,
		Name:     "test-plugin-list-1",
		Group:    "test-group",
		PkgName:  "test-plugin-pkg-1",
	}
	testPlugin2 := &types.Plugin{
		TenantID: testTenantID,
		Name:     "test-plugin-list-2",
		Group:    "test-group",
		PkgName:  "test-plugin-pkg-2",
	}

	err := h.Create(nCtx, testPlugin1)
	if err != nil {
		t.Fatalf("Failed to create test plugin 1: %v", err)
	}
	err = h.Create(nCtx, testPlugin2)
	if err != nil {
		t.Fatalf("Failed to create test plugin 2: %v", err)
	}

	type args struct {
		nCtx contextx.IContext
		page types.Page
		opts []OptFn
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "list all plugins",
			args: args{
				nCtx: nCtx,
				page: types.Page{
					Offset: 0,
					Limit:  10,
				},
				opts: []OptFn{},
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "list with pagination",
			args: args{
				nCtx: nCtx,
				page: types.Page{
					Offset: 0,
					Limit:  1,
				},
				opts: []OptFn{},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "list by group",
			args: args{
				nCtx: nCtx,
				page: types.Page{
					Offset: 0,
					Limit:  10,
				},
				opts: []OptFn{WithGroup("test-group")},
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "missing tenant ID",
			args: args{
				nCtx: contextx.New(context.Background()),
				page: types.Page{
					Offset: 0,
					Limit:  10,
				},
				opts: []OptFn{},
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := h.List(tt.args.nCtx, tt.args.page, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				expectedCount := int(tt.want)
				if len(got) != expectedCount {
					t.Errorf("List() got count = %v, want %v", len(got), expectedCount)
				}
			}
		})
	}
}

// TestHandler_Get test handler Get
func TestHandler_Get(t *testing.T) {
	h := testClient(t)
	testTenantID := tenantID()

	// Create a test plugin first
	nCtx := contextx.New(context.Background(), contextx.WithTenantID(testTenantID))
	testPlugin := &types.Plugin{
		TenantID: testTenantID,
		Name:     "test-plugin-get",
		Group:    "test-group",
		PkgName:  "test-plugin-pkg",
	}

	err := h.Create(nCtx, testPlugin)
	if err != nil {
		t.Fatalf("Failed to create test plugin: %v", err)
	}

	type args struct {
		nCtx contextx.IContext
		opts []OptFn
	}
	tests := []struct {
		name    string
		args    args
		want    *types.Plugin
		wantErr bool
	}{
		{
			name: "get by plugin name",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{WithName("test-plugin-get")},
			},
			want: &types.Plugin{
				TenantID: testTenantID,
				Name:     "test-plugin-get",
				Group:    "test-group",
				PkgName:  "test-plugin-pkg",
			},
			wantErr: false,
		},
		{
			name: "get non-existent plugin",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{WithName("non-existent-plugin")},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.Get(tt.args.nCtx, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestHandler_Exist test handler Exist
func TestHandler_Exist(t *testing.T) {
	h := testClient(t)
	testTenantID := tenantID()

	// Create a test plugin first
	nCtx := contextx.New(context.Background(), contextx.WithTenantID(testTenantID))
	testPlugin := &types.Plugin{
		TenantID: testTenantID,
		Name:     "test-plugin-exist",
		Group:    "test-group",
		PkgName:  "test-plugin-pkg",
	}

	err := h.Create(nCtx, testPlugin)
	if err != nil {
		t.Fatalf("Failed to create test plugin: %v", err)
	}

	type args struct {
		nCtx contextx.IContext
		opts []OptFn
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "plugin exists by name",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{WithName("test-plugin-exist")},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "plugin exists by group",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{WithGroup("test-group")},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "plugin does not exist",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{WithName("non-existent-plugin")},
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.Exist(tt.args.nCtx, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Exist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Exist() got = %v, want %v", got, tt.want)
			}
		})
	}
}
