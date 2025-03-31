/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package host ...
package host

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// testClient ...
func testClient(t *testing.T) Handler {
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
			&types.Host{
				TenantID: tenantID,
				HostID:   90001,
				Static: &types.HostStatic{
					HostName: "hostname-1",
				},
				Dynamic: &types.HostDynamic{
					NodeRole: types.NodeRoleAgent,
				},
			},
			&types.Host{
				TenantID: tenantID,
				HostID:   90002,
				Static: &types.HostStatic{
					HostName: "hostname-2",
				},
			},
			&types.Host{
				TenantID: tenantID,
				HostID:   90003,
				Static: &types.HostStatic{
					HostName: "unknown-name",
				},
				Dynamic: &types.HostDynamic{
					NodeRole:         types.NodeRoleAgent,
					NodeStatus:       types.NodeStatusRunning,
					NodeVersion:      "v2.0.0",
					NodeGeneration:   2,
					AgentID:          "00011113330003",
					NetworkUnitID:    1,
					Tag:              "deprecated",
					ProxyClusterPort: 33066,
					ProxyDataPort:    33067,
					ProxyFilePort:    33068,
				},
			},
			&types.Host{
				TenantID: tenantID,
				HostID:   90003,
				Static: &types.HostStatic{
					BizID:         0,
					NetworkAreaID: 0,
					HostName:      "unknown-name",
					DeptName:      "",
					InnerIP:       "192.168.132.114,172.133.122.143",
					InnerIPV6:     "",
					OuterIP:       "",
					OuterIPV6:     "",
					Mac:           "",
					OSType:        "",
					Arch:          "",
					Addressing:    types.AddressingDynamic,
					SyncedAgentID: "",
				},
			},
		)
		if err != nil {
			t.Errorf("prepareData() error = %v", err)
		}
	})
}

// Test_handler_ListAll ...
func Test_handler_ListAll(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "test")

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "test",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.ListAll(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, v := range got {
				t.Logf("ListAll() got = %v", v)
			}
		})
	}
}

// UpsertMany upsert many hosts.
func Test_handler_UpsertMany(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "test")

	type args struct {
		ctx   context.Context
		hosts []*types.Host
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nil ctx",
			args: args{
				ctx:   nil,
				hosts: nil,
			},
			wantErr: true,
		},
		{
			name: "nil hosts",
			args: args{
				ctx:   ctx,
				hosts: nil,
			},
			wantErr: true,
		},
		{
			name: "empty hosts",
			args: args{
				ctx:   ctx,
				hosts: []*types.Host{},
			},
			wantErr: true,
		},
		{
			name: "normal",
			args: args{
				ctx: ctx,
				hosts: []*types.Host{
					{
						HostID:   1,
						TenantID: "single",
						Static: &types.HostStatic{
							NetworkAreaID: 1,
							BizID:         1,
							InnerIP:       "127.0.0.1",
							Mac:           "123",
							OSType:        "centos",
						},
						Dynamic: &types.HostDynamic{
							NodeRole:       types.NodeRoleAgent,
							NodeVersion:    "1.1.1",
							NodeGeneration: 2,
						},
					},
					{
						HostID:   2,
						TenantID: "single",
						Static: &types.HostStatic{
							NetworkAreaID: 1,
							BizID:         2,
							InnerIP:       "127.0.0.21",
							Mac:           "123",
							OSType:        "centos",
						},
						Dynamic: &types.HostDynamic{
							NodeRole:       types.NodeRoleAgent,
							NodeVersion:    "1.1.1",
							NodeGeneration: 2,
						},
					},
					{
						HostID:   3,
						TenantID: "single",
						Static: &types.HostStatic{
							NetworkAreaID: 1,
							BizID:         3,
							InnerIP:       "127.0.0.3",
							Mac:           "123",
							OSType:        "centos",
						},
						Dynamic: &types.HostDynamic{
							NodeRole:       types.NodeRoleProxy,
							NodeVersion:    "1.1.2",
							NodeGeneration: 2,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "tenantID not match",
			args: args{
				ctx: ctx,
				hosts: []*types.Host{
					{
						HostID:   1,
						TenantID: "test",
						Static: &types.HostStatic{
							NetworkAreaID: 1,
							BizID:         1,
							InnerIP:       "127.0.0.1",
							Mac:           "123",
							OSType:        "centos",
						},
						Dynamic: &types.HostDynamic{
							NodeRole:       types.NodeRoleAgent,
							NodeVersion:    "1.1.1",
							NodeGeneration: 2,
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			err := h.UpsertMany(tt.args.ctx, tt.args.hosts...)
			if err != nil {
				t.Logf("UpsertMany() error = %v", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("UpsertMany() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_handler_UpsertStaticMany tests upsert host statics.
func Test_handler_UpsertStaticMany(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "test")

	type args struct {
		ctx   context.Context
		hosts []*types.Host
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nil ctx",
			args: args{
				ctx:   nil,
				hosts: nil,
			},
			wantErr: true,
		},
		{
			name: "nil hosts",
			args: args{
				ctx:   ctx,
				hosts: nil,
			},
			wantErr: true,
		},
		{
			name: "empty hosts",
			args: args{
				ctx:   ctx,
				hosts: []*types.Host{},
			},
			wantErr: true,
		},
		{
			name: "nil dynamic",
			args: args{
				ctx: ctx,
				hosts: []*types.Host{
					{
						HostID:   1,
						TenantID: "single",
						Static: &types.HostStatic{
							NetworkAreaID: 2,
							BizID:         2,
							InnerIP:       "127.0.0.2",
							Mac:           "456",
							OSType:        "macos",
						},
						Dynamic: nil,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "normal",
			args: args{
				ctx: ctx,
				hosts: []*types.Host{
					{
						HostID:   2,
						TenantID: "single",
						Static: &types.HostStatic{
							NetworkAreaID: 2,
							BizID:         2,
							InnerIP:       "127.0.0.2",
							Mac:           "456",
							OSType:        "macos",
						},
						Dynamic: &types.HostDynamic{
							NodeRole:       types.NodeRoleAgent,
							NodeVersion:    "wrong value",
							NodeGeneration: 2,
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			err := h.UpsertStaticMany(tt.args.ctx, tt.args.hosts...)
			if err != nil {
				t.Logf("UpsertStaticMany() %s error = %v", tt.name, err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("UpsertStaticMany() %s error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

// Test_handler_UpdateDynamicMany tests update host dynamics.
func Test_handler_UpdateDynamicMany(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "test")

	type args struct {
		ctx   context.Context
		hosts []*types.Host
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nil ctx",
			args: args{
				ctx:   nil,
				hosts: nil,
			},
			wantErr: true,
		},
		{
			name: "nil hosts",
			args: args{
				ctx:   ctx,
				hosts: nil,
			},
			wantErr: true,
		},
		{
			name: "empty hosts",
			args: args{
				ctx:   ctx,
				hosts: []*types.Host{},
			},
			wantErr: true,
		},
		{
			name: "nil static",
			args: args{
				ctx: ctx,
				hosts: []*types.Host{
					{
						HostID:   1,
						TenantID: "single",
						Static:   nil,
						Dynamic: &types.HostDynamic{
							NodeRole:       types.NodeRoleAgent,
							NodeVersion:    "1.1.1",
							NodeGeneration: 2,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "normal",
			args: args{
				ctx: ctx,
				hosts: []*types.Host{
					{
						HostID:   2,
						TenantID: "single",
						Static: &types.HostStatic{
							BizID:         1,
							NetworkAreaID: 1,
							HostName:      "test",
							DeptName:      "test",
							InnerIP:       "127.0.0.1",
							OSType:        "test",
						},
						Dynamic: &types.HostDynamic{
							NodeRole:       types.NodeRoleAgent,
							NodeVersion:    "1.1.2",
							NodeGeneration: 2,
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			err := h.UpdateDynamicMany(tt.args.ctx, tt.args.hosts...)
			if err != nil {
				t.Logf("UpdateDynamicMany() %s error = %v", tt.name, err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateDynamicMany() %s error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

// Test_handler_Count covers count method.
func Test_handler_Count(t *testing.T) {
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
			name:      "filter by host id",
			optFn:     []OptFn{WithHostID(90001, 90002)},
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:      "filter by fuzzy host name",
			optFn:     []OptFn{WithFuzzyHostName("hostname")},
			wantTotal: 2,
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

// Test_handler_List covers list method.
func Test_handler_List(t *testing.T) {
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
			name: "filter by host id",
			page: types.Page{
				Offset: 0,
				Limit:  1,
			},
			optFn:     []OptFn{WithHostID(90001, 90002)},
			wantTotal: 2,
			wantNum:   1,
			wantErr:   false,
		},
		{
			name: "filter by biz name",
			page: types.Page{
				Offset: 1,
				Limit:  1,
			},
			optFn:     []OptFn{WithFuzzyHostName("hostname")},
			wantTotal: 2,
			wantNum:   1,
			wantErr:   false,
		},
		{
			name: "filter by contains static inner ip",
			page: types.Page{
				Offset: 0,
				Limit:  1,
			},
			optFn:     []OptFn{WithStaticInnerIP("172.133.122.143")},
			wantTotal: 1,
			wantNum:   1,
			wantErr:   false,
		},
		{
			name: "filter by contains static addressing",
			page: types.Page{
				Offset: 0,
				Limit:  1,
				Sort:   "",
			},
			optFn:     []OptFn{WithStaticAddressing(types.AddressingDynamic)},
			wantTotal: 1,
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
