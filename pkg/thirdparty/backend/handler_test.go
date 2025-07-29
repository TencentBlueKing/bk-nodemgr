/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package backend provides handlers to operate nodemgr backend api.
package backend

import (
	"context"
	"os"
	"testing"

	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/joho/godotenv"
)

type testHeaderSetter struct{}

// GetAuthHeader ...
func (testHeaderSetter) GetAuthHeader() string {
	return os.Getenv("BK_APIGW_AUTHHEADER")
}

// testClient ...
func testClient(t *testing.T) Handler {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	httpClient, err := restclient.NewHTTPClient(&ssl.TLSConfig{
		InsecureSkipVerify: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	clientCap := &restclient.Capability{
		HTTPClient:           httpClient,
		Discover:             restdiscovery.NewDiscovery("apigateway", []string{os.Getenv("BK_APIGW_ENDPOINT")}),
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
		MetricOpts:           restclient.MetricOption{},
		Logger:               logger.LoggerDefault{},
	}

	h, err := New(clientCap, &Config{
		HeaderSetter: testHeaderSetter{},
	})
	if err != nil {
		t.Fatal(err)
	}

	return h
}

// Test_handler_ListBusiness list business.
func Test_handler_ListBusiness(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "single")

	type args struct {
		ctx       context.Context
		page      types.Page
		condition *types.BusinessCondition
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
				condition: nil,
			},
			wantErr: false,
		},
		{
			name: "page",
			args: args{
				ctx: ctx,
				page: types.Page{
					Offset: 0,
					Limit:  1,
				},
				condition: nil,
			},
			wantErr: false,
		},
		{
			name: "condition",
			args: args{
				ctx: ctx,
				page: types.Page{
					Offset: 0,
					Limit:  500,
				},
				condition: &types.BusinessCondition{
					ExactInclude: &types.BusinessExactFields{
						BizID: []int64{0, 1, 2},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, total, err := h.ListBusiness(tt.args.ctx, tt.args.page, tt.args.condition)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListBizHosts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, item := range got {
				t.Logf("biz: %v", item)
			}

			t.Logf("total: %d", total)
		})
	}
}

// Test_hanlder_ListHost list host.
func Test_hanlder_ListHost(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "single")

	type args struct {
		ctx       context.Context
		page      types.Page
		condition *types.HostCondition
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
				condition: nil,
			},
			wantErr: false,
		},
		{
			name: "page",
			args: args{
				ctx: ctx,
				page: types.Page{
					Offset: 0,
					Limit:  1,
				},
				condition: nil,
			},
			wantErr: false,
		},
		{
			name: "condition",
			args: args{
				ctx: ctx,
				page: types.Page{
					Offset: 0,
					Limit:  500,
				},
				condition: &types.HostCondition{
					ExactInclude: &types.HostExactFields{
						BizID: []int64{0, 1, 2},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, total, err := h.ListHost(tt.args.ctx, tt.args.page, tt.args.condition)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, item := range got {
				t.Logf("host: %v", item)
			}

			t.Logf("total: %d", total)
		})
	}
}

// Test_hanlder_ListNetworkArea list networkarea.
func Test_hanlder_ListNetworkArea(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "single")

	type args struct {
		ctx       context.Context
		page      types.Page
		condition *types.NetworkAreaCondition
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
				condition: nil,
			},
			wantErr: false,
		},
		{
			name: "page",
			args: args{
				ctx: ctx,
				page: types.Page{
					Offset: 0,
					Limit:  1,
				},
				condition: nil,
			},
			wantErr: false,
		},
		{
			name: "condition",
			args: args{
				ctx: ctx,
				page: types.Page{
					Offset: 0,
					Limit:  500,
				},
				condition: &types.NetworkAreaCondition{
					ExactInclude: &types.NetworkAreaExactFields{
						NetworkAreaID: []int64{0, 1, 2},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, total, err := h.ListNetworkArea(tt.args.ctx, tt.args.page, tt.args.condition)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListNetworkArea() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, item := range got {
				t.Logf("networkarea: %v", item)
			}

			t.Logf("total: %d", total)
		})
	}
}

// Test_hanlder_ListNetworkUnit list networkunit.
func Test_hanlder_ListNetworkUnit(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "single")

	type args struct {
		ctx       context.Context
		page      types.Page
		condition *types.NetworkUnitCondition
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
				condition: nil,
			},
			wantErr: false,
		},
		{
			name: "page",
			args: args{
				ctx: ctx,
				page: types.Page{
					Offset: 0,
					Limit:  1,
				},
				condition: nil,
			},
			wantErr: false,
		},
		{
			name: "condition",
			args: args{
				ctx: ctx,
				page: types.Page{
					Offset: 0,
					Limit:  500,
				},
				condition: &types.NetworkUnitCondition{
					ExactInclude: &types.NetworkUnitExactFields{
						NetworkAreaID: []int64{0, 1, 2},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, total, err := h.ListNetworkUnit(tt.args.ctx, tt.args.page, tt.args.condition)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListNetworkUnit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, item := range got {
				t.Logf("networkunit: %v", item)
			}

			t.Logf("total: %d", total)
		})
	}
}

// Test_hanlder_ListNetworkArea get network area.
func Test_hanlder_GetNetworkArea(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "single")

	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid id",
			args: args{
				ctx: ctx,
				id:  -1,
			},
			wantErr: true,
		},
		{
			name: "normal",
			args: args{
				ctx: ctx,
				id:  0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.GetNetworkArea(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNetworkArea() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("networkarea: %v", got)
		})
	}
}

// Test_hanlder_GetNetworkUnit get network unit.
func Test_hanlder_GetNetworkUnit(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "single")

	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid id",
			args: args{
				ctx: ctx,
				id:  -1,
			},
			wantErr: true,
		},
		{
			name: "normal",
			args: args{
				ctx: ctx,
				id:  0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, aps, err := h.GetNetworkUnit(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNetworkUnit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if len(aps) != len(got.AccessPoints) {
					t.Errorf("GetNetworkUnit() accesspoints length not match")
					return
				}

				t.Logf("networkunit: %v", got)
				t.Logf("accesspoints: %v", aps)
			}
		})
	}
}

// Test_hanlder_CreateNetworkUnit create network unit.
func Test_handler_CreateNetworkUnit(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "single")

	type args struct {
		ctx          context.Context
		networkunit  *types.NetworkUnit
		accesspoints []*types.AccessPoint
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid name",
			args: args{
				ctx: ctx,
				networkunit: &types.NetworkUnit{
					TenantID:      "single",
					Name:          "",
					NetworkAreaID: 0,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid networkarea id",
			args: args{
				ctx: ctx,
				networkunit: &types.NetworkUnit{
					TenantID:      "single",
					Name:          "default",
					NetworkAreaID: -1,
				},
			},
			wantErr: true,
		},
		{
			name: "normal",
			args: args{
				ctx: ctx,
				networkunit: &types.NetworkUnit{
					TenantID:      "single",
					Name:          "default",
					NetworkAreaID: 0,
				},
				accesspoints: []*types.AccessPoint{
					{
						TenantID:      "single",
						Name:          "default ap",
						NetworkAreaID: 0,
						Endpoints: types.Endpoints{
							Cluster: []string{"1.1.1.1"},
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
			got, err := h.CreateNetworkUnit(tt.args.ctx, tt.args.networkunit)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateNetworkUnit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got < 0 {
				t.Errorf("CreateNetworkUnit() got invalid id: %v", got)
			}

			t.Logf("got: %v", got)
		})
	}
}

// Test_handler_UpdateNetworkUnit update network unit.
func Test_handler_UpdateNetworkUnit(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "single")

	type args struct {
		ctx          context.Context
		networkunit  *types.NetworkUnit
		accesspoints []*types.AccessPoint
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid name",
			args: args{
				ctx: ctx,
				networkunit: &types.NetworkUnit{
					TenantID:      "single",
					Name:          "",
					NetworkAreaID: 0,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid networkarea id",
			args: args{
				ctx: ctx,
				networkunit: &types.NetworkUnit{
					TenantID:      "single",
					Name:          "default",
					NetworkAreaID: -1,
				},
			},
			wantErr: true,
		},
		{
			name: "normal",
			args: args{
				ctx: ctx,
				networkunit: &types.NetworkUnit{
					TenantID:      "single",
					Name:          "default-new",
					NetworkAreaID: 0,
				},
				accesspoints: []*types.AccessPoint{
					{
						TenantID:      "single",
						Name:          "default ap",
						NetworkAreaID: 0,
						Endpoints: types.Endpoints{
							Cluster: []string{"1.1.1.1"},
							File:    []string{"2.2.2.2", "3.3.3.3"},
						},
					},
					{
						TenantID:      "single",
						Name:          "default ap2",
						NetworkAreaID: 0,
						Endpoints: types.Endpoints{
							Cluster: []string{"1.1.1.1"},
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
			err := h.UpdateNetworkUnit(tt.args.ctx, tt.args.networkunit)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateNetworkUnit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

// Test_handler_UpdateNetworkUnit
func Test_handler_DeleteNetworkUnit(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "single")

	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid ctx",
			args: args{
				ctx: nil,
				id:  -1,
			},
			wantErr: true,
		},
		{
			name: "invalid id",
			args: args{
				ctx: ctx,
				id:  -1,
			},
			wantErr: true,
		},
		{
			name: "normal",
			args: args{
				ctx: ctx,
				id:  0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			err := h.DeleteNetworkUnit(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteNetworkUnit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

// Test_handler_ListAccessPoint list access point.
func Test_handler_ListAccessPoint(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "single")

	type args struct {
		ctx       context.Context
		page      types.Page
		condition *types.AccessPointCondition
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
				condition: nil,
			},
			wantErr: false,
		},
		{
			name: "page",
			args: args{
				ctx: ctx,
				page: types.Page{
					Offset: 0,
					Limit:  1,
				},
				condition: nil,
			},
			wantErr: false,
		},
		{
			name: "condition",
			args: args{
				ctx: ctx,
				page: types.Page{
					Offset: 0,
					Limit:  500,
				},
				condition: &types.AccessPointCondition{
					ExactInclude: &types.AccessPointExactFields{
						NetworkAreaID: []int64{0, 1, 2},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, total, err := h.ListAccessPoint(tt.args.ctx, tt.args.page, tt.args.condition)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListAccessPoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, item := range got {
				t.Logf("accesspoints: %v", item)
			}

			t.Logf("total: %d", total)
		})
	}
}

// Test_hander_GetConstant get constant
func Test_hander_GetConstant(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "single")

	type args struct {
		ctx    context.Context
		fields types.TopoConstantFields
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid ctx",
			args: args{
				ctx:    nil,
				fields: types.TopoConstantFields{},
			},
			wantErr: true,
		},
		{
			name: "base",
			args: args{
				ctx: ctx,
				fields: types.TopoConstantFields{
					CloudVendor: true,
					OSType:      true,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.GetConstant(tt.args.ctx, tt.args.fields)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConstant() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("got: %v", got)
		})
	}
}
