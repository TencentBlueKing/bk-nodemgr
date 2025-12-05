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
	"reflect"
	"sort"
	"sync"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
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

// prepareData for all tests.
func prepareData(t *testing.T, nCtx contextx.IContext) {
	once.Do(func() {
		tenantID := nCtx.TenantID()

		// pre insert.
		h := testClient(t)
		err := h.UpsertMany(nCtx,
			&types.Host{
				TenantID: tenantID,
				HostID:   90001,
				Static: &types.HostStatic{
					HostName: "hostname-1",
				},
				Dynamic: &types.HostDynamic{
					NodeStatus:  types.NodeStatusDamaged,
					NodeRole:    types.NodeRoleAgent,
					NodeVersion: "v2.0.0",
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
					HostName:      "unknown-name",
					NetworkAreaID: 1,
				},
				Dynamic: &types.HostDynamic{
					NodeRole:            types.NodeRoleAgent,
					NodeStatus:          types.NodeStatusRunning,
					NodeVersion:         "v2.0.0",
					NodeGeneration:      2,
					AgentID:             "00011113330003",
					NetworkUnitID:       1,
					ProxyTags:           []types.ProxyTag{types.ProxyTagDedicatedInstaller},
					ProxyAccessDisabled: true,
					ProxyClusterPort:    33066,
					ProxyDataPort:       33067,
					ProxyFilePort:       33068,
				},
			},
			&types.Host{
				TenantID: tenantID,
				HostID:   90004,
				Static: &types.HostStatic{
					BizID:         0,
					NetworkAreaID: 0,
					HostName:      "unknown-name",
					DeptName:      "",
					InnerIPList:   []string{"127.0.0.1", "127.0.0.2"},
					InnerIPV6List: []string{""},
					OuterIPList:   []string{""},
					OuterIPV6List: []string{""},
					Mac:           "",
					OSType:        "",
					Arch:          "",
					Addressing:    types.AddressingDynamic,
					SyncedAgentID: "",
				},
				Dynamic: &types.HostDynamic{
					NodeVersion: "v2.0.1",
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
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("single"))

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
			got, err := h.ListAll(nCtx)
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
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("single"))

	type args struct {
		nCtx  contextx.IContext
		hosts []*types.Host
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nil nCtx",
			args: args{
				nCtx:  nil,
				hosts: nil,
			},
			wantErr: true,
		},
		{
			name: "nil hosts",
			args: args{
				nCtx:  nCtx,
				hosts: nil,
			},
			wantErr: true,
		},
		{
			name: "empty hosts",
			args: args{
				nCtx:  nCtx,
				hosts: []*types.Host{},
			},
			wantErr: true,
		},
		{
			name: "normal",
			args: args{
				nCtx: nCtx,
				hosts: []*types.Host{
					{
						HostID:   1,
						TenantID: "single",
						Static: &types.HostStatic{
							NetworkAreaID: 1,
							BizID:         1,
							InnerIPList:   []string{"127.0.0.1"},
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
							InnerIPList:   []string{"127.0.0.2"},
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
							InnerIPList:   []string{"127.0.0.3"},
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
				nCtx: nCtx,
				hosts: []*types.Host{
					{
						HostID:   1,
						TenantID: "test",
						Static: &types.HostStatic{
							NetworkAreaID: 1,
							BizID:         1,
							InnerIPList:   []string{"127.0.0.1"},
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
			err := h.UpsertMany(tt.args.nCtx, tt.args.hosts...)
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
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("single"))

	type args struct {
		nCtx  contextx.IContext
		hosts []*types.Host
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nil nCtx",
			args: args{
				nCtx:  nil,
				hosts: nil,
			},
			wantErr: true,
		},
		{
			name: "nil hosts",
			args: args{
				nCtx:  nCtx,
				hosts: nil,
			},
			wantErr: true,
		},
		{
			name: "empty hosts",
			args: args{
				nCtx:  nCtx,
				hosts: []*types.Host{},
			},
			wantErr: true,
		},
		{
			name: "nil dynamic",
			args: args{
				nCtx: nCtx,
				hosts: []*types.Host{
					{
						HostID:   1,
						TenantID: "single",
						Static: &types.HostStatic{
							NetworkAreaID: 2,
							BizID:         2,
							InnerIPList:   []string{"127.0.0.2"},
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
				nCtx: nCtx,
				hosts: []*types.Host{
					{
						HostID:   2,
						TenantID: "single",
						Static: &types.HostStatic{
							NetworkAreaID: 2,
							BizID:         2,
							InnerIPList:   []string{"127.0.0.2"},
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
			err := h.UpsertStaticMany(tt.args.nCtx, tt.args.hosts...)
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
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("single"))

	type args struct {
		nCtx  contextx.IContext
		hosts []*types.Host
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nil nCtx",
			args: args{
				nCtx:  nil,
				hosts: nil,
			},
			wantErr: true,
		},
		{
			name: "nil hosts",
			args: args{
				nCtx:  nCtx,
				hosts: nil,
			},
			wantErr: true,
		},
		{
			name: "empty hosts",
			args: args{
				nCtx:  nCtx,
				hosts: []*types.Host{},
			},
			wantErr: true,
		},
		{
			name: "nil static",
			args: args{
				nCtx: nCtx,
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
				nCtx: nCtx,
				hosts: []*types.Host{
					{
						HostID:   2,
						TenantID: "single",
						Static: &types.HostStatic{
							BizID:         1,
							NetworkAreaID: 1,
							HostName:      "test",
							DeptName:      "test",
							InnerIPList:   []string{"127.0.0.1"},
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
			err := h.UpdateDynamicMany(tt.args.nCtx, tt.args.hosts...)
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
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("single"))

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
			name:      "filter by host id",
			optFn:     []OptFn{WithHostID(90001, 90002)},
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:      "filter by fuzzy host name",
			optFn:     []OptFn{WithFuzzyStaticHostName("hostname")},
			wantTotal: 2,
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

// Test_handler_List covers list method.
func Test_handler_List(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("single"))

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
			optFn:     []OptFn{WithFuzzyStaticHostName("hostname")},
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
			optFn:     []OptFn{WithStaticInnerIPList("127.0.0.1")},
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
		{
			name: "filter by dynamic proxy tag",
			page: types.Page{
				Offset: 0,
				Limit:  1,
			},
			optFn:     []OptFn{WithDynamicProxyTags(types.ProxyTagDedicatedInstaller)},
			wantTotal: 1,
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

// Test_handler_DistinctNodeVersion distinct node role fields.
func Test_handler_DistinctNodeVersion(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("single"))

	prepareData(t, nCtx)

	type args struct {
		optFn []OptFn
	}
	tests := []struct {
		name       string
		args       args
		wantResult []string
		wantErr    bool
	}{
		{
			name: "normal",
			args: args{
				optFn: []OptFn{WithHostID(90001, 90002, 90003, 90004)},
			},
			wantResult: []string{"", "v2.0.0", "v2.0.1"},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			gotResult, err := h.DistinctNodeVersion(nCtx, tt.args.optFn...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DistinctNodeVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			sort.Strings(gotResult)
			sort.Strings(tt.wantResult)

			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("DistinctNodeVersion() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

// Test_handler_DistinctNodeStatus distinct node status fields.
func Test_handler_DistinctNodeStatus(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("single"))

	prepareData(t, nCtx)

	type args struct {
		optFn []OptFn
	}
	tests := []struct {
		name       string
		args       args
		wantResult []types.NodeStatus
		wantErr    bool
	}{
		{
			name: "normal",
			args: args{
				optFn: []OptFn{WithHostID(90001, 90002, 90003, 90004)},
			},
			wantResult: []types.NodeStatus{"", types.NodeStatusDamaged, types.NodeStatusRunning},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			gotResult, err := h.DistinctNodeStatus(nCtx, tt.args.optFn...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DistinctNodeStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			sort.Slice(gotResult, func(i, j int) bool { return gotResult[i] < gotResult[j] })
			sort.Slice(tt.wantResult, func(i, j int) bool { return tt.wantResult[i] < tt.wantResult[j] })

			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("DistinctNodeStatus() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

// Test_handler_DistinctNetworkAreaID distinct networkarea fields.
func Test_handler_DistinctNetworkAreaID(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("single"))

	prepareData(t, nCtx)

	type args struct {
		optFn []OptFn
	}
	tests := []struct {
		name       string
		args       args
		wantResult []int64
		wantErr    bool
	}{
		{
			name: "normal",
			args: args{
				optFn: []OptFn{WithHostID(90001, 90002, 90003, 90004)},
			},
			wantResult: []int64{0, 1},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			gotResult, err := h.DistinctNetworkAreaID(nCtx, tt.args.optFn...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DistinctNetworkAreaID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			sort.Slice(gotResult, func(i, j int) bool { return gotResult[i] < gotResult[j] })
			sort.Slice(tt.wantResult, func(i, j int) bool { return tt.wantResult[i] < tt.wantResult[j] })

			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("DistinctNetworkAreaID() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

// Test_handler_DeleteMany delete many hosts.
func Test_handler_DeleteMany(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("single"))

	prepareData(t, nCtx)

	type args struct {
		nCtx    contextx.IContext
		hostIDs []int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				nCtx:    nCtx,
				hostIDs: []int64{90001, 90002, 90003, 90004},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			if err := h.DeleteMany(tt.args.nCtx, tt.args.hostIDs...); (err != nil) != tt.wantErr {
				t.Errorf("DeleteMany() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
