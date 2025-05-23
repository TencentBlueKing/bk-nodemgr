/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package networkarea

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// testClient ...
func testClient(t *testing.T) IHandler {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	mongoClient, err := mongo.Connect(
		ctx,
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

	return New(mongoClient.Database(os.Getenv("MONGO_DATABASE")), logger.LoggerDefault{})
}

var once = sync.Once{}

// prepareData for all tests.
func prepareData(t *testing.T, ctx context.Context) {
	once.Do(func() {
		tenantID, _ := tenant.GetID(ctx)

		// pre insert.
		h := testClient(t)
		err := h.UpsertMany(ctx,
			&types.NetworkArea{
				TenantID: tenantID,
				ID:       90001,
				Name:     "test-name-90001",
			},
			&types.NetworkArea{
				TenantID: tenantID,
				ID:       90002,
				Name:     "test-name-same",
			},
			&types.NetworkArea{
				TenantID: tenantID,
				ID:       90003,
				Name:     "test-name-same",
			},
		)
		if err != nil {
			t.Errorf("prepareData() error = %v", err)
		}

		systemCtx, _ := tenant.SetID(context.Background(), "system_tenant")
		err = h.UpsertMany(systemCtx,
			&types.NetworkArea{
				TenantID: "system_tenant",
				ID:       base.GlobalNetworkAreaID,
				Name:     "test-name-same",
			})
		if err != nil {
			t.Errorf("prepareData() upsert system tenant error = %v", err)
		}
	})
}

// Test_handler_Get get network area.
func Test_handler_Get(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	ctx, _ := tenant.SetID(context.Background(), "test")

	prepareData(t, ctx)

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "normal",
			id:      90001,
			wantErr: false,
		},
		{
			name:    "invalid id",
			id:      -1,
			wantErr: true,
		},
		{
			name:    "not exist",
			id:      10000000,
			wantErr: true,
		},
		{
			name:    "get global",
			id:      base.GlobalNetworkAreaID,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.Get(ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Get() got = %v", got)
		})
	}
}

// Test_handler_Count count network area.
func Test_handler_Count(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	ctx, _ := tenant.SetID(context.Background(), "test")

	prepareData(t, ctx)

	tests := []struct {
		name      string
		optFn     []OptFn
		wantTotal int64
		wantErr   bool
	}{
		{
			name:      "normal",
			optFn:     nil,
			wantTotal: -1,
			wantErr:   false,
		},
		{
			name:      "filter by networkarea id",
			optFn:     []OptFn{WithNetworkAreaID(90001, 90002)},
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:      "filter by fuzzy networkarea name",
			optFn:     []OptFn{WithFuzzyNetworkAreaName("name-same")},
			wantTotal: 3,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.Count(ctx, tt.optFn...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Count() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantTotal > 0 && got != tt.wantTotal {
				t.Errorf("Count() got = %d, wantTotal %d", got, tt.wantTotal)
				return
			}
			t.Logf("Count() got = %d", got)
		})
	}
}

// Test_handler_List list network area.
func Test_handler_List(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	ctx, _ := tenant.SetID(context.Background(), "test")

	prepareData(t, ctx)

	tests := []struct {
		name      string
		page      types.Page
		optFn     []OptFn
		wantTotal int64
		wantNum   int64
		wantErr   bool
	}{
		{
			name: "normal",
			page: types.Page{
				Offset: 0,
				Limit:  0,
			},
			optFn:     nil,
			wantTotal: -1,
			wantNum:   -1,
			wantErr:   false,
		},
		{
			name: "filter by networkarea id",
			page: types.Page{
				Offset: 0,
				Limit:  1,
			},
			optFn:     []OptFn{WithNetworkAreaID(90001, 90002, 0)},
			wantTotal: 3,
			wantNum:   1,
			wantErr:   false,
		},
		{
			name: "filter by biz name",
			page: types.Page{
				Offset: 1,
				Limit:  1,
			},
			optFn:     []OptFn{WithFuzzyNetworkAreaName("name-same")},
			wantTotal: 3,
			wantNum:   1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, total, err := h.List(ctx, tt.page, tt.optFn...)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantTotal > 0 && tt.wantTotal != total {
				t.Errorf("List() total = %d, wantTotal %d", total, tt.wantTotal)
				return
			}
			t.Logf("List() total = %d", total)

			if tt.wantNum > 0 && tt.wantNum != int64(len(got)) {
				t.Errorf("List() num = %d, wantNum %d", len(got), tt.wantNum)
				return
			}
			t.Logf("List() num = %d", len(got))

			for _, v := range got {
				t.Logf("List() got = %v", v)
			}
		})
	}
}

// Test_handler_UpsertMany upserts many networkareas.
func Test_handler_UpsertMany(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	ctx, _ := tenant.SetID(context.Background(), "test")

	type args struct {
		ctx          context.Context
		networkAreas []*types.NetworkArea
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nil ctx",
			args: args{
				ctx:          nil,
				networkAreas: nil,
			},
			wantErr: true,
		},
		{
			name: "nil networkareas",
			args: args{
				ctx:          ctx,
				networkAreas: nil,
			},
			wantErr: true,
		},
		{
			name: "empty networkareas",
			args: args{
				ctx:          ctx,
				networkAreas: []*types.NetworkArea{},
			},
			wantErr: true,
		},
		{
			name: "forbid cross tenant",
			args: args{
				ctx: ctx,
				networkAreas: []*types.NetworkArea{
					{
						TenantID: "test",
						ID:       0,
						Name:     "new-name-90001",
					}},
			},
			wantErr: true,
		},
		{
			name: "normal",
			args: args{
				ctx: ctx,
				networkAreas: []*types.NetworkArea{
					{
						TenantID: "test",
						ID:       90001,
						Name:     "new-name-90001",
					},
					{
						TenantID: "test",
						ID:       90002,
						Name:     "new-name-90002",
					},
					{
						TenantID: "test",
						ID:       90003,
						Name:     "new-name-90003",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			err := h.UpsertMany(tt.args.ctx, tt.args.networkAreas...)
			if err != nil {
				t.Logf("UpsertMany() error = %v", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("UpsertMany() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_handler_DeleteMany deletes many networkareas.
func Test_handler_DeleteMany(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	ctx, _ := tenant.SetID(context.Background(), "test")

	prepareData(t, ctx)

	type args struct {
		ctx            context.Context
		networkAreaIDs []int64
	}

	tests := []struct {
		name            string
		args            args
		wantErr         bool
		ensureExists    []int64
		ensureNotExists []int64
	}{
		{
			name: "nil ctx",
			args: args{
				ctx:            nil,
				networkAreaIDs: nil,
			},
			wantErr:         true,
			ensureExists:    nil,
			ensureNotExists: nil,
		},
		{
			name: "empty ids",
			args: args{
				ctx:            ctx,
				networkAreaIDs: []int64{},
			},
			wantErr:         true,
			ensureExists:    nil,
			ensureNotExists: nil,
		},
		{
			name: "forbid cross tenant",
			args: args{
				ctx:            ctx,
				networkAreaIDs: []int64{0},
			},
			wantErr:         false,
			ensureExists:    []int64{0},
			ensureNotExists: nil,
		},
		{
			name: "normal",
			args: args{
				ctx:            ctx,
				networkAreaIDs: []int64{90001, 90002, 90003},
			},
			wantErr:         false,
			ensureExists:    nil,
			ensureNotExists: []int64{90001, 90002, 90003},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			err := h.DeleteMany(tt.args.ctx, tt.args.networkAreaIDs...)
			if err != nil {
				t.Logf("DeleteMany() error = %v", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteMany() error = %v, wantErr %v", err, tt.wantErr)
			}

			for _, networkAreaID := range tt.ensureExists {
				if _, err := h.Get(tt.args.ctx, networkAreaID); err != nil {
					t.Errorf("DeleteMany() ensureExists %d, error = %v", networkAreaID, err)
				}
			}

			for _, networkAreaID := range tt.ensureNotExists {
				if _, err := h.Get(tt.args.ctx, networkAreaID); err == nil {
					t.Errorf("DeleteMany() ensureNotExists %d, error = %v", networkAreaID, err)
				}
			}
		})
	}
}
