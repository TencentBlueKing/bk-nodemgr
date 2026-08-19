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

package configpolicyevent

import (
	"context"
	"os"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var nCtx = contextx.New(context.Background(), contextx.WithTenantID("test"))

// testClient ...
func testClient(t *testing.T) IHandler {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

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

var once = sync.Once{}

// prepareData for all tests.
func prepareData(t *testing.T, nCtx contextx.IContext) {
	once.Do(func() {
		tenantID, _ := tenant.GetID(nCtx)

		// pre insert.
		h := testClient(t)
		err := h.CreateMany(nCtx,
			&types.PolicyEvent{
				TenantID:         tenantID,
				EventType:        types.PolicyEventTypeCreate,
				NodeRole:         types.NodeRoleAgent,
				ConfigPolicyID:   1,
				ConfigPolicyName: "test",
				Version:          1,
				Operator:         "admin",
				OperateTime:      time.Date(2024, 3, 1, 0, 0, 0, 0, time.Local),
			},
			&types.PolicyEvent{
				TenantID:         tenantID,
				EventType:        types.PolicyEventTypeCreate,
				NodeRole:         types.NodeRoleAgent,
				ConfigPolicyID:   2,
				ConfigPolicyName: "test-2",
				Version:          2,
				Operator:         "admin",
				OperateTime:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local),
			},
			&types.PolicyEvent{
				TenantID:         tenantID,
				EventType:        types.PolicyEventTypeCreate,
				NodeRole:         types.NodeRoleProxy,
				ConfigPolicyID:   3,
				ConfigPolicyName: "test-3",
				Version:          3,
				Operator:         "admin",
				OperateTime:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local),
			},
			&types.PolicyEvent{
				TenantID:         tenantID,
				EventType:        types.PolicyEventTypeUpdate,
				NodeRole:         types.NodeRoleAgent,
				ConfigPolicyID:   4,
				ConfigPolicyName: "test-4",
				Version:          4,
				Operator:         "admin",
				OperateTime:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local),
			},
		)
		if err != nil {
			t.Errorf("prepareData() error = %v", err)
		}
	})
}

// Test_handler_List list event by page and conditions.
func Test_handler_List(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("test"))

	prepareData(t, nCtx)

	tests := []struct {
		name      string
		nCtx      contextx.IContext
		page      types.Page
		optFn     []OptFn
		wantTotal int64
		wantNum   int64
		wantErr   bool
	}{
		{
			name: "nil nCtx",
			nCtx: nil,
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
			nCtx: nCtx,
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
			name: "filter by node role",
			nCtx: nCtx,
			page: types.Page{
				Offset: 0,
				Limit:  1,
			},
			optFn:     []OptFn{WithNodeRole(types.NodeRoleAgent)},
			wantTotal: -1,
			wantNum:   1,
			wantErr:   false,
		},
		{
			name: "filter by event type",
			nCtx: nCtx,
			page: types.Page{
				Offset: 0,
				Limit:  4,
			},
			optFn:     []OptFn{WithEventType(types.PolicyEventTypeCreate)},
			wantTotal: -1,
			wantNum:   3,
			wantErr:   false,
		},
		{
			name: "filter with time range",
			nCtx: nCtx,
			page: types.UnlimitedPage(),
			optFn: []OptFn{WithOperateTimeRange(types.TimeRange{
				StartTime: time.Date(2022, 1, 1, 0, 0, 0, 0, time.Local),
				EndTime:   time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local),
			})},
			wantTotal: -1,
			wantNum:   4,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, total, err := h.List(tt.nCtx, tt.page, tt.optFn...)
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

	tests := []struct {
		name    string
		nCtx    contextx.IContext
		optFn   []OptFn
		wantErr bool
	}{
		{
			name:    "nil nCtx",
			nCtx:    nil,
			optFn:   nil,
			wantErr: true,
		},
		{
			name:    "normal",
			nCtx:    nCtx,
			optFn:   nil,
			wantErr: false,
		},
		{
			name:    "filter by node role",
			nCtx:    nCtx,
			optFn:   []OptFn{WithNodeRole(types.NodeRoleAgent)},
			wantErr: false,
		},
		{
			name:    "filter by event type",
			nCtx:    nCtx,
			optFn:   []OptFn{WithEventType(types.PolicyEventTypeCreate)},
			wantErr: false,
		},
		{
			name:    "filter by operator",
			nCtx:    nCtx,
			optFn:   []OptFn{WithOperator("admin")},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.Count(tt.nCtx, tt.optFn...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Count() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Count() total = %d", got)
		})
	}
}

// Test_handler_DistinctEventType tests the distinct with type field.
func Test_handler_DistinctEventType(t *testing.T) {
	type args struct {
		nCtx  contextx.IContext
		optFn []OptFn
	}
	tests := []struct {
		name       string
		args       args
		wantResult []types.PolicyEventType
		wantErr    bool
	}{
		{
			name: "invalid nCtx",
			args: args{
				nCtx:  nil,
				optFn: []OptFn{},
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name: "normal",
			args: args{
				nCtx:  nCtx,
				optFn: []OptFn{},
			},
			wantResult: []types.PolicyEventType{types.PolicyEventTypeCreate, types.PolicyEventTypeUpdate},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			gotResult, err := h.DistinctEventType(tt.args.nCtx, tt.args.optFn...)
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

// Test_handler_DistinctNodeRole tests the distinct with node role type field.
func Test_handler_DistinctNodeRole(t *testing.T) {
	type args struct {
		nCtx  contextx.IContext
		optFn []OptFn
	}
	tests := []struct {
		name       string
		args       args
		wantResult []types.NodeRole
		wantErr    bool
	}{
		{
			name: "invalid nCtx",
			args: args{
				nCtx:  nil,
				optFn: []OptFn{},
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name: "normal",
			args: args{
				nCtx:  nCtx,
				optFn: []OptFn{},
			},
			wantResult: []types.NodeRole{types.NodeRoleAgent, types.NodeRoleProxy},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			gotResult, err := h.DistinctNodeRole(tt.args.nCtx, tt.args.optFn...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DistinctNodeRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			sort.Slice(gotResult, func(i, j int) bool { return gotResult[i] < gotResult[j] })
			sort.Slice(tt.wantResult, func(i, j int) bool { return tt.wantResult[i] < tt.wantResult[j] })

			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("DistinctNodeRole() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

// Test_handler_DistinctOperator tests the distinct with operator field.
func Test_handler_DistinctOperator(t *testing.T) {
	type args struct {
		nCtx  contextx.IContext
		optFn []OptFn
	}
	tests := []struct {
		name       string
		args       args
		wantResult []string
		wantErr    bool
	}{
		{
			name: "invalid nCtx",
			args: args{
				nCtx:  nil,
				optFn: []OptFn{},
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name: "normal",
			args: args{
				nCtx:  nCtx,
				optFn: []OptFn{},
			},
			wantResult: []string{"admin"},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			gotResult, err := h.DistinctOperator(tt.args.nCtx, tt.args.optFn...)
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
