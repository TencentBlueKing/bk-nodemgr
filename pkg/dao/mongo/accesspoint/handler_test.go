/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package accesspoint

import (
	"context"
	"fmt"
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
var preparedAccessPointIDs []int64
var preparedGlobalAccessPointID int64

// prepareData for all tests.
func prepareData(t *testing.T, ctx context.Context) {
	once.Do(func() {
		tenantID, _ := tenant.GetID(ctx)

		// pre insert.
		h := testClient(t)

		systemCtx, _ := tenant.SetID(context.Background(), "system_tenant")
		tests := []struct {
			ctx         context.Context
			accessPoint *types.AccessPoint
			isGlobal    bool
		}{
			{
				ctx: ctx,
				accessPoint: &types.AccessPoint{
					TenantID:      tenantID,
					NetworkAreaID: base.GlobalNetworkAreaID,
					Name:          "test-name-1",
					Endpoints: types.Endpoints{
						Cluster: []string{"1.1.1.1"},
						File:    []string{"2.2.2.2"},
						Data:    []string{"3.3.3.3"},
					},
				},
				isGlobal: false,
			},
			{
				ctx: ctx,
				accessPoint: &types.AccessPoint{
					TenantID:      tenantID,
					NetworkAreaID: 1,
					Name:          "test-name-2",
					Endpoints: types.Endpoints{
						Cluster: []string{"1.1.1.1"},
						File:    []string{"2.2.2.2"},
						Data:    []string{"3.3.3.3"},
					},
				},
				isGlobal: false,
			},
			{
				ctx: systemCtx,
				accessPoint: &types.AccessPoint{
					TenantID:      "system_tenant",
					NetworkAreaID: base.GlobalNetworkAreaID,
					Name:          "test-name-system",
					Endpoints: types.Endpoints{
						Cluster: []string{"1.1.1.1"},
						File:    []string{"2.2.2.2"},
						Data:    []string{"3.3.3.3"},
					},
				},
				isGlobal: true,
			},
		}

		preparedAccessPointIDs = make([]int64, 0)
		for _, tt := range tests {
			accesspointID, err := h.Create(tt.ctx, tt.accessPoint)
			if err != nil {
				t.Errorf("prepareData() error = %v", err)
			}

			preparedAccessPointIDs = append(preparedAccessPointIDs, accesspointID)

			if tt.isGlobal {
				preparedGlobalAccessPointID = accesspointID
			}
		}
	})
}

// Test_handler_Get get access point.
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
			name:    "invalid id",
			id:      -1,
			wantErr: true,
		},
		{
			name:    "not exist",
			id:      10000000,
			wantErr: true,
		},
	}
	for idx, id := range preparedAccessPointIDs {
		tests = append(tests, struct {
			name    string
			id      int64
			wantErr bool
		}{
			name:    fmt.Sprintf("exist-%d", idx),
			id:      id,
			wantErr: false,
		})
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

// Test_handler_Count count access point.
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
			optFn:     []OptFn{WithNetworkAreaID(base.GlobalNetworkAreaID)},
			wantTotal: -1,
			wantErr:   false,
		},
		{
			name:      "filter by accesspoint id",
			optFn:     []OptFn{WithAccessPointID(preparedAccessPointIDs...)},
			wantTotal: int64(len(preparedAccessPointIDs)),
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

// Test_handler_List list access point.
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
			optFn:     []OptFn{WithNetworkAreaID(base.GlobalNetworkAreaID)},
			wantTotal: -1,
			wantNum:   1,
			wantErr:   false,
		},
		{
			name: "filter by accesspoint id",
			page: types.Page{
				Offset: 1,
				Limit:  1,
			},
			optFn:     []OptFn{WithAccessPointID(preparedAccessPointIDs...)},
			wantTotal: int64(len(preparedAccessPointIDs)),
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

// Test_handler_Create creates accesspoint.
func Test_handler_Create(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	ctx, _ := tenant.SetID(context.Background(), "test")

	type args struct {
		ctx         context.Context
		accessPoint *types.AccessPoint
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nil ctx",
			args: args{
				ctx:         nil,
				accessPoint: nil,
			},
			wantErr: true,
		},
		{
			name: "nil accesspoint",
			args: args{
				ctx:         ctx,
				accessPoint: nil,
			},
			wantErr: true,
		},
		{
			name: "forbid cross tenant",
			args: args{
				ctx: ctx,
				accessPoint: &types.AccessPoint{
					TenantID: "test-other",
					Name:     "new-name-90001",
				},
			},
			wantErr: true,
		},
		{
			name: "normal",
			args: args{
				ctx: ctx,
				accessPoint: &types.AccessPoint{
					TenantID: "test",
					Name:     "new-name-90002",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			id, err := h.Create(tt.args.ctx, tt.args.accessPoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && id < 0 {
				t.Errorf("Create() return invalid accesspoint id: %d", id)
			}
		})
	}
}

// Test_handler_CreateMany creates many accesspoint.
func Test_handler_CreateMany(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	ctx, _ := tenant.SetID(context.Background(), "test")

	type args struct {
		ctx          context.Context
		accessPoints []*types.AccessPoint
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
				accessPoints: nil,
			},
			wantErr: true,
		},
		{
			name: "nil accesspoint",
			args: args{
				ctx:          ctx,
				accessPoints: nil,
			},
			wantErr: true,
		},
		{
			name: "forbid cross tenant",
			args: args{
				ctx: ctx,
				accessPoints: []*types.AccessPoint{
					{
						TenantID: "test-other",
						Name:     "new-name-90001",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "normal",
			args: args{
				ctx: ctx,
				accessPoints: []*types.AccessPoint{
					{
						TenantID: "test",
						Name:     "new-name-90002",
					},
					{
						TenantID: "test",
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
			ids, err := h.CreateMany(tt.args.ctx, tt.args.accessPoints...)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateMany() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && len(ids) != len(tt.args.accessPoints) {
				t.Errorf("CreateMany() return invalid accesspoint id num: %d, expected %d",
					len(ids), len(tt.args.accessPoints))
			}
		})
	}
}

// Test_handler_UpdateMany updates many accesspoints.
func Test_handler_UpdateMany(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	ctx, _ := tenant.SetID(context.Background(), "test")

	type args struct {
		ctx          context.Context
		accessPoints []*types.AccessPoint
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
				accessPoints: nil,
			},
			wantErr: true,
		},
		{
			name: "nil accesspoint",
			args: args{
				ctx:          ctx,
				accessPoints: nil,
			},
			wantErr: true,
		},
		{
			name: "empty accesspoints",
			args: args{
				ctx:          ctx,
				accessPoints: []*types.AccessPoint{},
			},
			wantErr: true,
		},
		{
			name: "forbid cross tenant",
			args: args{
				ctx: ctx,
				accessPoints: []*types.AccessPoint{
					{
						TenantID:      "system_tenant",
						ID:            preparedGlobalAccessPointID,
						NetworkAreaID: base.GlobalNetworkAreaID,
						Name:          "new-name-90001",
					}},
			},
			wantErr: true,
		},
		{
			name: "normal",
			args: args{
				ctx: ctx,
				accessPoints: []*types.AccessPoint{
					{
						TenantID:      "test",
						ID:            preparedAccessPointIDs[0],
						NetworkAreaID: base.GlobalNetworkAreaID,
						Name:          "new-name-90002",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			err := h.UpdateMany(tt.args.ctx, tt.args.accessPoints...)
			if err != nil {
				t.Logf("UpdateMany() error = %v", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateMany() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_handler_DeleteMany deletes many accesspoints.
func Test_handler_DeleteMany(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	ctx, _ := tenant.SetID(context.Background(), "test")

	prepareData(t, ctx)

	type args struct {
		ctx            context.Context
		accessPointIDs []int64
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
				accessPointIDs: nil,
			},
			wantErr:         true,
			ensureExists:    nil,
			ensureNotExists: nil,
		},
		{
			name: "empty ids",
			args: args{
				ctx:            ctx,
				accessPointIDs: []int64{},
			},
			wantErr:         true,
			ensureExists:    nil,
			ensureNotExists: nil,
		},
		{
			name: "forbid cross tenant",
			args: args{
				ctx:            ctx,
				accessPointIDs: []int64{preparedGlobalAccessPointID},
			},
			wantErr:         false,
			ensureExists:    []int64{preparedGlobalAccessPointID},
			ensureNotExists: nil,
		},
		{
			name: "normal",
			args: args{
				ctx:            ctx,
				accessPointIDs: []int64{preparedAccessPointIDs[0]},
			},
			wantErr:         false,
			ensureExists:    nil,
			ensureNotExists: []int64{preparedAccessPointIDs[0]},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			err := h.DeleteMany(tt.args.ctx, tt.args.accessPointIDs...)
			if err != nil {
				t.Logf("DeleteMany() error = %v", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteMany() error = %v, wantErr %v", err, tt.wantErr)
			}

			for _, accessPointID := range tt.ensureExists {
				if _, err := h.Get(tt.args.ctx, accessPointID); err != nil {
					t.Errorf("DeleteMany() ensureExists %d, error = %v", accessPointID, err)
				}
			}

			for _, accessPointID := range tt.ensureNotExists {
				if _, err := h.Get(tt.args.ctx, accessPointID); err == nil {
					t.Errorf("DeleteMany() ensureNotExists %d, error = %v", accessPointID, err)
				}
			}
		})
	}
}
