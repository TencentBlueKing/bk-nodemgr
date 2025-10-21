/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed on the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package tenant

import (
	"context"
	"os"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// testClient creates a test handler instance
func testClient(t *testing.T) *Handler {
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

	handler := New(mongoClient.Database(os.Getenv("MONGO_DATABASE")))

	// Clean up test data before running tests
	cleanupTestData(t, handler)

	return handler
}

// cleanupTestData cleans up all test data
func cleanupTestData(t *testing.T, handler *Handler) {
	nCtx := contextx.New(context.Background())

	// Drop the entire collection to ensure clean test environment
	err := handler.dao.GetClient().Drop(nCtx)
	if err != nil {
		t.Logf("Warning: failed to drop collection: %v", err)
	}
}

// TestHandler_Create test handler Create
func TestHandler_Create(t *testing.T) {
	type args struct {
		nCtx   contextx.IContext
		tenant *types.Tenant
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal test",
			args: args{
				nCtx: contextx.New(context.Background()),
				tenant: &types.Tenant{
					ID:      uuid.NewString(),
					Name:    "test-tenant",
					Enabled: true,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			if err := h.Create(tt.args.nCtx, tt.args.tenant); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestHandler_Count test handler Count
func TestHandler_Count(t *testing.T) {
	h := testClient(t)
	testTenantID := uuid.NewString()

	// Create a test tenant first
	nCtx := contextx.New(context.Background())
	testTenant := &types.Tenant{
		ID:      testTenantID,
		Name:    "test-tenant-count",
		Enabled: true,
	}

	err := h.Create(nCtx, testTenant)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
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
			name: "count all tenants",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "count by tenant ID",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{WithID(testTenantID)},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "count by tenant name",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{WithName("test-tenant-count")},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "count by status",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{WithStatus(true)},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "count by non-existent tenant",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{WithID("non-existent-tenant")},
			},
			want:    0,
			wantErr: false,
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
	testTenantID1 := uuid.NewString()
	testTenantID2 := uuid.NewString()

	// Create test tenants first
	nCtx := contextx.New(context.Background())
	testTenant1 := &types.Tenant{
		ID:      testTenantID1,
		Name:    "test-tenant-list-1",
		Enabled: true,
	}
	testTenant2 := &types.Tenant{
		ID:      testTenantID2,
		Name:    "test-tenant-list-2",
		Enabled: false,
	}

	err := h.Create(nCtx, testTenant1)
	if err != nil {
		t.Fatalf("Failed to create test tenant 1: %v", err)
	}
	err = h.Create(nCtx, testTenant2)
	if err != nil {
		t.Fatalf("Failed to create test tenant 2: %v", err)
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
			name: "list 2 tenants",
			args: args{
				nCtx: nCtx,
				page: types.Page{
					Offset: 0,
					Limit:  2,
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
			name: "list by tenant ID",
			args: args{
				nCtx: nCtx,
				page: types.Page{
					Offset: 0,
					Limit:  10,
				},
				opts: []OptFn{WithID(testTenantID1)},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "list by status",
			args: args{
				nCtx: nCtx,
				page: types.Page{
					Offset: 0,
					Limit:  1,
				},
				opts: []OptFn{WithStatus(false)},
			},
			want:    1,
			wantErr: false,
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

// TestHandler_Exist test handler Exist
func TestHandler_Exist(t *testing.T) {
	h := testClient(t)
	testTenantID := uuid.NewString()

	// Create a test tenant first
	nCtx := contextx.New(context.Background())
	testTenant := &types.Tenant{
		ID:      testTenantID,
		Name:    "test-tenant-exist",
		Enabled: true,
	}

	err := h.Create(nCtx, testTenant)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
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
			name: "tenant exists by ID",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{WithID(testTenantID)},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "tenant exists by name",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{WithName("test-tenant-exist")},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "tenant exists by status",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{WithStatus(true)},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "tenant does not exist",
			args: args{
				nCtx: nCtx,
				opts: []OptFn{WithID("non-existent-tenant")},
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
