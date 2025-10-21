/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package configpolicy

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

var once = sync.Once{}
var preparedConfigPolicyIDs []int64
var otherTenantConfigPolicyIDs []int64

// prepareData for all tests.
func prepareData(t *testing.T, nCtx contextx.IContext) {
	once.Do(func() {
		tenantID := nCtx.TenantID()

		// pre insert.
		h := testClient(t)

		systemCtx := contextx.New(context.Background(), contextx.WithTenantID("system_tenant"))
		tests := []struct {
			nCtx         contextx.IContext
			configPolicy *types.ConfigPolicy
			otherTenant  bool
		}{
			{
				nCtx: nCtx,
				configPolicy: &types.ConfigPolicy{
					TenantID: tenantID,
					Name:     "test-name-1",
					NodeRole: types.NodeRoleAgent,
					BizID:    []int64{0},
					Scopes: []types.ConfigPolicyScope{
						{
							NetworkAreaID: 1,
							NetworkUnitID: 2,
							NodeOsType:    criteria.OSLinux,
							NodeCPUArch:   criteria.CPUArchAmd64,
						},
					},
					Configs: map[string]any{
						"agent.thread_num": 10,
						"file.disable_bt":  true,
					},
					Enabled:   true,
					UpdatedAt: time.Now(),
					Operator:  "admin",
				},
			},
			{
				nCtx: nCtx,
				configPolicy: &types.ConfigPolicy{
					TenantID: tenantID,
					Name:     "test-name-2",
					NodeRole: types.NodeRoleAgent,
					BizID:    []int64{0},
					Scopes: []types.ConfigPolicyScope{
						{
							NetworkAreaID: 3,
							NetworkUnitID: 4,
							NodeOsType:    criteria.OSWindows,
							NodeCPUArch:   criteria.CPUArchAmd64,
						},
					},
					Configs: map[string]any{
						"data.compressed": true,
						"logger.level":    "DEBUG",
					},
					Enabled:   true,
					UpdatedAt: time.Now(),
					Operator:  "admin",
				},
			},
			{
				nCtx: systemCtx,
				configPolicy: &types.ConfigPolicy{
					TenantID: "system_tenant",
					Name:     "test-name-system",
					NodeRole: types.NodeRoleAgent,
					BizID:    []int64{0},
					Scopes: []types.ConfigPolicyScope{
						{
							NetworkAreaID: 5,
							NetworkUnitID: 6,
							NodeOsType:    criteria.OSAix,
							NodeCPUArch:   criteria.CPUArchArm64,
						},
					},
					Configs:   map[string]any{},
					Enabled:   false,
					UpdatedAt: time.Now(),
					Operator:  "admin",
				},
				otherTenant: true,
			},
		}

		preparedConfigPolicyIDs = make([]int64, 0)
		for _, tt := range tests {
			configPolicyID, err := h.Create(tt.nCtx, tt.configPolicy)
			if err != nil {
				t.Errorf("prepareData() error = %v", err)
			}
			if tt.otherTenant {
				otherTenantConfigPolicyIDs = append(otherTenantConfigPolicyIDs, configPolicyID)
				continue
			}

			preparedConfigPolicyIDs = append(preparedConfigPolicyIDs, configPolicyID)
		}
	})
}

// Test_Count tests the Count method.
func Test_Count(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("test"))

	prepareData(t, nCtx)

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
			name:      "filter by configpolicy id",
			optFn:     []OptFn{WithConfigPolicyID(preparedConfigPolicyIDs...)},
			wantTotal: int64(len(preparedConfigPolicyIDs)),
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.Count(nCtx, tt.optFn...)
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

// Test_List tests the List method.
func Test_List(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("test"))

	prepareData(t, nCtx)

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
			name: "filter by configpolicy id",
			page: types.Page{
				Offset: 1,
				Limit:  1,
			},
			optFn:     []OptFn{WithConfigPolicyID(preparedConfigPolicyIDs...)},
			wantTotal: int64(len(preparedConfigPolicyIDs)),
			wantNum:   1,
			wantErr:   false,
		},
		{
			name: "filter by scope",
			page: types.Page{
				Offset: 0,
				Limit:  1,
			},
			optFn:     []OptFn{WithEnabledScope(0, 1, 2, criteria.OSLinux, criteria.CPUArchAmd64)},
			wantTotal: -1,
			wantNum:   1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, total, err := h.List(nCtx, tt.page, tt.optFn...)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantTotal > 0 && tt.wantTotal != total {
				t.Errorf("List() total = %d, wantTotal %d", total, tt.wantTotal)
				return
			}
			t.Logf("List() total = %d, got = %d", total, len(got))

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

// Test_Get tests the Get method.
func Test_Get(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("test"))

	prepareData(t, nCtx)

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

	for idx, id := range preparedConfigPolicyIDs {
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

	for idx, id := range otherTenantConfigPolicyIDs {
		tests = append(tests, struct {
			name    string
			id      int64
			wantErr bool
		}{
			name:    fmt.Sprintf("other-tenant-%d", idx),
			id:      id,
			wantErr: true,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.Get(nCtx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Get() got = %v", got)
		})
	}
}

// Test_Create tests the Create method.
func Test_Create(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("test"))

	prepareData(t, nCtx)

	type args struct {
		nCtx         contextx.IContext
		configPolicy *types.ConfigPolicy
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nil nCtx",
			args: args{
				nCtx:         nil,
				configPolicy: nil,
			},
			wantErr: true,
		},
		{
			name: "nil configpolicy",
			args: args{
				nCtx:         nCtx,
				configPolicy: nil,
			},
			wantErr: true,
		},
		{
			name: "forbid cross tenant",
			args: args{
				nCtx: nCtx,
				configPolicy: &types.ConfigPolicy{
					TenantID: "test-other",
					Name:     "new-name-90001",
				},
			},
			wantErr: true,
		},
		{
			name: "normal",
			args: args{
				nCtx: nCtx,
				configPolicy: &types.ConfigPolicy{
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
			id, err := h.Create(tt.args.nCtx, tt.args.configPolicy)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && id < 0 {
				t.Errorf("Create() return invalid configpolicy id: %d", id)
			}
		})
	}
}

// Test_UpdateMany tests the UpdateMany method.
func Test_UpdateMany(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("test"))

	prepareData(t, nCtx)

	type args struct {
		nCtx          contextx.IContext
		configPolicys []*types.ConfigPolicy
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nil nCtx",
			args: args{
				nCtx:          nil,
				configPolicys: nil,
			},
			wantErr: true,
		},
		{
			name: "nil configpolicy",
			args: args{
				nCtx:          nCtx,
				configPolicys: nil,
			},
			wantErr: true,
		},
		{
			name: "empty configpolicys",
			args: args{
				nCtx:          nCtx,
				configPolicys: []*types.ConfigPolicy{},
			},
			wantErr: true,
		},
	}

	for idx, id := range preparedConfigPolicyIDs {
		tests = append(tests, struct {
			name    string
			args    args
			wantErr bool
		}{
			name: fmt.Sprintf("exist-%d", idx),
			args: args{
				nCtx: nCtx,
				configPolicys: []*types.ConfigPolicy{
					{
						ID:       id,
						TenantID: "test",
						Name:     fmt.Sprintf("updated-%d", idx),
						BizID:    []int64{1, 2, 3, 4},
						Version:  999,
					},
				},
			},
			wantErr: false,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			err := h.UpdateMany(tt.args.nCtx, tt.args.configPolicys...)
			if err != nil {
				t.Logf("UpdateMany() error = %v", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateMany() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

// Test_EnableMany tests the EnableMany method.
func Test_EnableMany(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("test"))

	prepareData(t, nCtx)

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{}

	for idx, id := range preparedConfigPolicyIDs {
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
			err := h.EnableMany(nCtx, tt.id)
			if err != nil {
				t.Logf("EnableMany() error = %v", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("EnableMany() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			configPolicy, err := h.Get(nCtx, tt.id)
			if err != nil {
				t.Errorf("EnableMany() check Get() error = %v", err)
			}
			if !configPolicy.Enabled {
				t.Errorf("EnableMany() failed to enable")
			}
		})
	}
}

// Test_DisableMany tests the DisableMany method.
func Test_DisableMany(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("test"))

	prepareData(t, nCtx)

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{}

	for idx, id := range preparedConfigPolicyIDs {
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
			err := h.DisableMany(nCtx, tt.id)
			if err != nil {
				t.Logf("DisableMany() error = %v", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("DisableMany() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			configPolicy, err := h.Get(nCtx, tt.id)
			if err != nil {
				t.Errorf("DisableMany() check Get() error = %v", err)
			}
			if configPolicy.Enabled {
				t.Errorf("DisableMany() failed to enable")
			}
		})
	}
}

// Test_DeleteMany tests the DeleteMany method.
func Test_DeleteMany(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("test"))

	prepareData(t, nCtx)

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{}

	for idx, id := range preparedConfigPolicyIDs {
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
			err := h.DeleteMany(nCtx, tt.id)
			if err != nil {
				t.Logf("DeleteMany() error = %v", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteMany() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if _, err = h.Get(nCtx, tt.id); err == nil {
				t.Error("DeleteMany() found not deleted")
			}
		})
	}
}
