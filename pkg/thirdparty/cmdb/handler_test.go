/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package cmdb ...
package cmdb

import (
	"context"
	"os"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/joho/godotenv"
)

type testHeaderSetter struct{}

// GetAuthHeader ...
func (testHeaderSetter) GetAuthHeader() (string, error) {
	return os.Getenv("BK_APIGW_AUTHHEADER"), nil
}

// testClient ...
func testClient(t *testing.T) Handler {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	httpClient, err := client.NewClient(&ssl.TLSConfig{
		InsecureSkipVerify: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	clientCap := &client.Capability{
		Client:               httpClient,
		Discover:             discovery.NewDiscovery("apigateway", []string{os.Getenv("BK_APIGW_ENDPOINT")}),
		ToleranceLatencyTime: client.ToleranceLatencyTimeDefault,
		MetricOpts:           client.MetricOption{},
		Logger:               logger.LoggerDefault{},
	}

	h, err := New(clientCap, &Config{
		SupplierAccount: os.Getenv("BK_SUPPLIER_ACCOUNT"),
		HeaderSetter:    testHeaderSetter{},
	})
	if err != nil {
		t.Fatal(err)
	}

	return h
}

// Test_handler_ListBizHosts ...
func Test_handler_ListBizHosts(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx  context.Context
		biz  types.Business
		page types.Page
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "base",
			args: args{
				ctx: ctx,
				biz: types.Business{
					TenantID: "",
					BizID:    2,
					BizName:  "",
				},
				page: types.Page{
					Offset: 0,
					Limit:  500,
					Sort:   "",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.ListBizHosts(tt.args.ctx, tt.args.biz.BizID, tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListBizHosts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, host := range got {
				t.Logf("host: %v", host)
			}
		})
	}
}

// Test_handler_SearchBusiness ...
func Test_handler_SearchBusiness(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx  context.Context
		page types.Page
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "base",
			args: args{
				ctx: ctx,
				page: types.Page{
					Offset: 0,
					Limit:  500,
					Sort:   "",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.SearchBusiness(tt.args.ctx, tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchBusiness() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, biz := range got {
				t.Logf("biz: %v", biz)
			}
		})
	}
}

// Test_handler_SearchNetArea ...
func Test_handler_SearchNetworkArea(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx  context.Context
		page types.Page
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx: ctx,
				page: types.Page{
					Offset: 0,
					Limit:  500,
					Sort:   "",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.SearchNetworkArea(tt.args.ctx, tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchNetArea() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, netArea := range got {
				t.Logf("netArea: %v", netArea)
			}
		})
	}
}

// Test_handler_NetworkArea... handle network area curd test
func Test_handler_NetworkArea(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx             context.Context
		networkAreaName string
		cloudVendor     string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:             ctx,
				networkAreaName: "test_nodemgr_cloud",
				cloudVendor:     "5",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			created, err := h.CreateNetworkArea(tt.args.ctx, tt.args.networkAreaName, tt.args.cloudVendor)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateNetworkArea() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("CreateNetworkArea got: %v", created)

			changeName := tt.args.networkAreaName + "_1"
			changeVendor := "1"
			err = h.UpdateNetworkArea(tt.args.ctx, created.ID, changeName, changeVendor)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateNetworkArea() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("UpdateNetworkArea success")

			page := types.Page{
				Offset: 0,
				Limit:  500,
				Sort:   "",
			}
			search, err := h.SearchNetworkArea(tt.args.ctx, page)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchNetworkArea() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, netArea := range search {
				t.Logf("SearchNetworkArea got: %#v", *netArea)
			}

			err = h.DeleteNetworkArea(tt.args.ctx, created.ID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteNetworkArea() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("DeleteNetworkArea success")
		})
	}
}

// Test_handler_SearchCloudVendor...
func Test_handler_SearchCloudVendor(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.SearchCloudVendor(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchCloudVendor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, attr := range got {
				t.Logf("index: %d, attr: %#v", index, *attr)
			}
		})
	}
}

// Test_handler_SearchOsType...
func Test_handler_SearchOsType(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.SearchOsType(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchOsType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, attr := range got {
				t.Logf("index: %d, attr: %#v", index, *attr)
			}
		})
	}
}

// Test_handler_CreateAndUpdateHost...
func Test_handler_CreateAndUpdateHost(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx   context.Context
		bizID int64
		hosts []*types.Host
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:   ctx,
				bizID: 2,
				hosts: []*types.Host{
					{
						Static: &types.HostStatic{
							InnerIP:       "1.1.1.3",
							NetworkAreaID: 0,
							OSType:        "1",
							Arch:          "x86",
							Addressing:    "static",
						},
					},
					{
						Static: &types.HostStatic{
							InnerIP:       "1.1.1.4",
							NetworkAreaID: 0,
							OSType:        "1",
							Arch:          "x86",
							Addressing:    "static",
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

			created, err := h.AddHostToBusinessIdle(tt.args.ctx, tt.args.bizID, tt.args.hosts[:1])
			if (err != nil) != tt.wantErr {
				t.Errorf("AddHostToBusinessIdle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, hostID := range created {
				t.Logf("index: %d, host: %#v", index, hostID)
			}

			success, errmsg, err := h.AddHostToResourcePool(tt.args.ctx, tt.args.hosts[1:])
			if (err != nil) != tt.wantErr {
				t.Errorf("AddHostToResourcePool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, host := range success {
				t.Logf("index: %d, host: %#v", index, host)
			}

			for index, msg := range errmsg {
				t.Logf("index: %d, msg: %#v", index, msg)
			}

			var changeNetworkAreaID int64 = 3
			err = h.UpdateHostNetworkAreaField(tt.args.ctx, tt.args.bizID, changeNetworkAreaID, created...)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateHostNetworkAreaField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			hostAgentIDs := []*types.Host{
				{
					HostID: created[0],
					Dynamic: &types.HostDynamic{
						AgentID: "xxxxxxxxxxxxxxxxxxxxxxxxx",
					},
				},
			}
			err = h.BindHostAgent(tt.args.ctx, hostAgentIDs...)
			if (err != nil) != tt.wantErr {
				t.Errorf("BindHostAgent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			err = h.UnbindHostAgent(tt.args.ctx, hostAgentIDs)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnbindHostAgent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

// Test_handler_ListResourcePoolHosts...
func Test_handler_ListResourcePoolHosts(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx  context.Context
		page types.Page
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "base",
			args: args{
				ctx: ctx,
				page: types.Page{
					Offset: 0,
					Limit:  500,
					Sort:   "",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.ListResourcePoolHosts(tt.args.ctx, tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListResourcePoolHosts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, host := range got {
				t.Logf("host: %#v", host)
			}
		})
	}
}

func Test_handler_ListHostsWithoutBusiness(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx  context.Context
		page types.Page
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "base",
			args: args{
				ctx: ctx,
				page: types.Page{
					Offset: 0,
					Limit:  500,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.ListHostsWithoutBusiness(tt.args.ctx, tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListHostsWithoutBusiness() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, host := range got {
				t.Logf("host: %#v", host)
			}
		})
	}
}

// Test_handler_DynamicGroup...
func Test_handler_DynamicGroup(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx   context.Context
		bizID int64
		group *types.DynamicGroup
		page  types.Page
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:   ctx,
				bizID: 2,
				group: &types.DynamicGroup{
					BizID: 2,
					ObjID: "host",
					Name:  "nodemgr_test",
				},
				page: types.Page{
					Offset: 0,
					Limit:  500,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.SearchDynamicGroup(tt.args.ctx, tt.args.bizID, tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchDynamicGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, group := range got {
				t.Logf("index: %d, group: %#v", index, *group)

				hosts, err := h.ExecuteHostDynamicGroup(tt.args.ctx, tt.args.bizID, group.ID, tt.args.page)
				if (err != nil) != tt.wantErr {
					t.Errorf("ExecuteDynamicGroup() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				for index, host := range hosts {
					t.Logf("index: %d, host: %#v", index, *host)
				}
			}
		})
	}
}

// Test_handler_ListServiceTemplate...
func Test_handler_ListServiceTemplate(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx   context.Context
		bizID int64
		page  types.Page
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:   ctx,
				bizID: 2,
				page: types.Page{
					Offset: 0,
					Limit:  500,
					Sort:   "",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.ListServiceTemplate(tt.args.ctx, tt.args.bizID, tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListServiceTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, serviceTemplate := range got {
				t.Logf("index: %d, serviceTemplate: %#v", index, *serviceTemplate)
			}
		})
	}
}

// Test_handler_FindHostByServiceTemplate...
func Test_handler_FindHostByServiceTemplate(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx                context.Context
		bizID              int64
		serviceTemplateIDs []int64
		page               types.Page
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:                ctx,
				bizID:              2,
				serviceTemplateIDs: []int64{1},
				page: types.Page{
					Offset: 0,
					Limit:  500,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.FindHostByServiceTemplate(tt.args.ctx, tt.args.bizID, tt.args.page,
				tt.args.serviceTemplateIDs...)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindHostByServiceTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, host := range got {
				t.Logf("index: %d, host: %#v", index, *host)
			}
		})
	}
}
