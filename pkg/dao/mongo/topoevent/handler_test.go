/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package topoevent

import (
	"context"
	"os"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"

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
		err := h.CreateMany(ctx,
			&types.TopoEvent{
				TenantID:        tenantID,
				Type:            types.TopoEventNetworkAreaCreate,
				NetworkAreaID:   10001,
				NetworkAreaName: "default",
				Operator:        "admin",
				OperateTime:     time.Date(2024, 3, 1, 0, 0, 0, 0, time.Local),
			},
			&types.TopoEvent{
				TenantID:        tenantID,
				Type:            types.TopoEventNetworkUnitUpdate,
				NetworkAreaID:   10001,
				NetworkAreaName: "default",
				NetworkUnitID:   10001,
				NetworkUnitName: "test-network-unit",
				Operator:        "admin",
				OperateTime:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local),
			},
			&types.TopoEvent{
				TenantID:        tenantID,
				Type:            types.TopoEventAccessPointDelete,
				NetworkAreaID:   10001,
				NetworkAreaName: "default",
				NetworkUnitID:   10001,
				NetworkUnitName: "test-network-unit",
				AccessPointID:   10001,
				AccessPointName: "test-access-point",
				Operator:        "admin",
				OperateTime:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local),
			},
			&types.TopoEvent{
				TenantID:        tenantID,
				Type:            types.TopoEventAccessPointDelete,
				NetworkAreaID:   10001,
				NetworkAreaName: "default",
				NetworkUnitID:   10001,
				NetworkUnitName: "test-network-unit",
				AccessPointID:   10002,
				AccessPointName: "test-access-point-2",
				Operator:        "user",
				OperateTime:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local),
			},
		)
		if err != nil {
			t.Errorf("prepareData() error = %v", err)
		}
	})
}

// Test_handler_List list event by page and conditions.
func Test_handler_List(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	ctx, _ := tenant.SetID(context.Background(), "test")

	prepareData(t, ctx)

	tests := []struct {
		name      string
		ctx       context.Context
		page      types.Page
		optFn     []OptFn
		wantTotal int64
		wantNum   int64
		wantErr   bool
	}{
		{
			name: "nil ctx",
			ctx:  nil,
			page: types.Page{
				Offset: 0,
				Limit:  0,
			},
			optFn:     nil,
			wantTotal: -1,
			wantNum:   -1,
			wantErr:   true,
		},
		{
			name: "normal",
			ctx:  ctx,
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
			ctx:  ctx,
			page: types.Page{
				Offset: 0,
				Limit:  1,
			},
			optFn:     []OptFn{WithNetworkAreaID(10001)},
			wantTotal: -1,
			wantNum:   1,
			wantErr:   false,
		},
		{
			name: "filter by networkunit id",
			ctx:  ctx,
			page: types.Page{
				Offset: 0,
				Limit:  1,
			},
			optFn:     []OptFn{WithNetworkUnitID(10001)},
			wantTotal: -1,
			wantNum:   1,
			wantErr:   false,
		},
		{
			name: "filter by accesspoint id",
			ctx:  ctx,
			page: types.Page{
				Offset: 0,
				Limit:  1,
			},
			optFn:     []OptFn{WithAccessPointID(10001)},
			wantTotal: -1,
			wantNum:   1,
			wantErr:   false,
		},
		{
			name: "filter with time range",
			ctx:  ctx,
			page: types.Page{
				Offset: 0,
				Limit:  1,
			},
			optFn: []OptFn{WithOperateTimeRange(types.TimeRange{
				StartTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
				EndTime:   time.Date(2024, 5, 1, 0, 0, 0, 0, time.Local),
			})},
			wantTotal: -1,
			wantNum:   1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, total, err := h.List(tt.ctx, tt.page, tt.optFn...)
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

// Test_handler_Count tests the count with filter.
func Test_handler_Count(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	ctx, _ := tenant.SetID(context.Background(), "test")

	prepareData(t, ctx)

	tests := []struct {
		name    string
		ctx     context.Context
		optFn   []OptFn
		wantErr bool
	}{
		{
			name:    "nil ctx",
			ctx:     nil,
			optFn:   nil,
			wantErr: true,
		},
		{
			name:    "normal",
			ctx:     ctx,
			optFn:   nil,
			wantErr: false,
		},
		{
			name:    "filter by networkarea id",
			ctx:     ctx,
			optFn:   []OptFn{WithNetworkAreaID(10001)},
			wantErr: false,
		},
		{
			name:    "filter by networkunit id",
			ctx:     ctx,
			optFn:   []OptFn{WithNetworkUnitID(10001)},
			wantErr: false,
		},
		{
			name:    "filter by accesspoint id",
			ctx:     ctx,
			optFn:   []OptFn{WithAccessPointID(10001)},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.Count(tt.ctx, tt.optFn...)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("List() total = %d", got)
		})
	}
}

// Test_handler_DistinctType tests the distinct with type field.
func Test_handler_DistinctType(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	ctx, _ := tenant.SetID(context.Background(), "test")

	prepareData(t, ctx)

	type args struct {
		ctx   context.Context
		optFn []OptFn
	}
	tests := []struct {
		name       string
		args       args
		wantResult []types.TopoEventType
		wantErr    bool
	}{
		{
			name: "invalid ctx",
			args: args{
				ctx:   nil,
				optFn: []OptFn{},
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name: "normal",
			args: args{
				ctx:   ctx,
				optFn: []OptFn{},
			},
			wantResult: []types.TopoEventType{types.TopoEventNetworkAreaCreate, types.TopoEventNetworkUnitUpdate, types.TopoEventAccessPointDelete},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			gotResult, err := h.DistinctType(tt.args.ctx, tt.args.optFn...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DistinctType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			sort.Slice(gotResult, func(i, j int) bool { return gotResult[i] < gotResult[j] })
			sort.Slice(tt.wantResult, func(i, j int) bool { return tt.wantResult[i] < tt.wantResult[j] })

			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("DistinctType() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

// Test_handler_DistinctAccessPoint tests the distinct with accesspoint-id field.
func Test_handler_DistinctAccessPointID(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	ctx, _ := tenant.SetID(context.Background(), "test")

	prepareData(t, ctx)

	type args struct {
		ctx   context.Context
		optFn []OptFn
	}
	tests := []struct {
		name       string
		args       args
		wantResult []int64
		wantErr    bool
	}{
		{
			name: "invalid ctx",
			args: args{
				ctx:   nil,
				optFn: []OptFn{},
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name: "normal",
			args: args{
				ctx:   ctx,
				optFn: []OptFn{WithType(types.TopoEventAccessPointDelete)},
			},
			wantResult: []int64{10001, 10002},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			gotResult, err := h.DistinctAccessPointID(tt.args.ctx, tt.args.optFn...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DistinctAccessPointID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			sort.Slice(gotResult, func(i, j int) bool { return gotResult[i] < gotResult[j] })
			sort.Slice(tt.wantResult, func(i, j int) bool { return tt.wantResult[i] < tt.wantResult[j] })

			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("DistinctAccessPointID() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

// Test_handler_DistinctOperator tests the distinct with operator field.
func Test_handler_DistinctOperator(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	ctx, _ := tenant.SetID(context.Background(), "test")

	prepareData(t, ctx)

	type args struct {
		ctx   context.Context
		optFn []OptFn
	}
	tests := []struct {
		name       string
		args       args
		wantResult []string
		wantErr    bool
	}{
		{
			name: "invalid ctx",
			args: args{
				ctx:   nil,
				optFn: []OptFn{},
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name: "normal",
			args: args{
				ctx:   ctx,
				optFn: []OptFn{},
			},
			wantResult: []string{"admin", "user"},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			gotResult, err := h.DistinctOperator(tt.args.ctx, tt.args.optFn...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DistinctOperator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			sort.Slice(gotResult, func(i, j int) bool { return gotResult[i] < gotResult[j] })
			sort.Slice(tt.wantResult, func(i, j int) bool { return tt.wantResult[i] < tt.wantResult[j] })

			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("DistinctOperator() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
