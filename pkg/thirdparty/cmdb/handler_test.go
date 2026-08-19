//go:build integration

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

// Package cmdb ...
package cmdb

import (
	"context"
	"strings"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// testClient ...
func testClient(t *testing.T) IHandler {
	t.Helper()

	return newIntegrationHandler(t, newIntegrationTarget(t))
}

// Test_handler_ListBizHosts ...
func Test_handler_ListBizHosts(t *testing.T) {
	ctx := contextx.New(context.Background(), contextx.WithTenantID("0"), contextx.WithBKUsername("test"))

	type args struct {
		ctx  contextx.IContext
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
		{
			name: "base",
			args: args{
				ctx: ctx,
				biz: types.Business{
					TenantID: "",
					BizID:    12,
					BizName:  "",
				},
				page: types.UnlimitedPage(),
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
	ctx := contextx.New(context.Background(), contextx.WithTenantID("0"), contextx.WithBKUsername("test"))

	type args struct {
		ctx  contextx.IContext
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
	ctx := contextx.New(context.Background(), contextx.WithTenantID("0"), contextx.WithBKUsername("test"))

	type args struct {
		ctx  contextx.IContext
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
	ctx := contextx.New(context.Background(), contextx.WithTenantID("0"), contextx.WithBKUsername("test"))
	type args struct {
		ctx             contextx.IContext
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

// Test_handler_CreateAndUpdateHost...
func Test_handler_CreateAndUpdateHost(t *testing.T) {
	ctx := contextx.New(context.Background(), contextx.WithTenantID("0"), contextx.WithBKUsername("test"))
	type args struct {
		ctx   contextx.IContext
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
							InnerIPList:   []string{"1.1.1.3"},
							NetworkAreaID: 0,
							OSType:        "1",
							Arch:          "x86",
							Addressing:    "static",
						},
					},
					{
						Static: &types.HostStatic{
							InnerIPList:   []string{"1.1.1.4"},
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

			created, err := h.AddHostToBusinessIdle(tt.args.ctx, tt.args.bizID, tt.args.hosts[:1]...)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddHostToBusinessIdle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, hostID := range created {
				t.Logf("index: %d, host: %#v", index, hostID)
			}

			exist, err := h.CheckBizHostByIP(tt.args.ctx, tt.args.bizID, tt.args.hosts[0].Static.NetworkAreaID,
				strings.Join(tt.args.hosts[0].Static.InnerIPList, ipSeparator))
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckBizHostByIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if exist {
				t.Logf("CheckBizHostByIP success")
			}

			success, errmsg, err := h.AddHostToResourcePool(tt.args.ctx, tt.args.hosts[1:]...)
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

			hosts := []*types.Host{
				{
					HostID: created[0],
					Dynamic: &types.HostDynamic{
						AgentID: "xxxxxxxxxxxxxxxxxxxxxxxxx",
					},
				},
			}
			err = h.BindHostAgent(tt.args.ctx, hosts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("IBindHostAgent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			err = h.UnbindHostAgent(tt.args.ctx, hosts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("IUnbindHostAgent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

// Test_handler_ListResourcePoolHosts...
func Test_handler_ListResourcePoolHosts(t *testing.T) {
	ctx := contextx.New(context.Background(), contextx.WithTenantID("0"), contextx.WithBKUsername("test"))

	type args struct {
		ctx  contextx.IContext
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
	ctx := contextx.New(context.Background(), contextx.WithTenantID("0"), contextx.WithBKUsername("test"))

	type args struct {
		ctx  contextx.IContext
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

// Test_handler_ListServiceTemplate...
func Test_handler_ListServiceTemplate(t *testing.T) {
	ctx := contextx.New(context.Background(), contextx.WithTenantID("0"), contextx.WithBKUsername("test"))

	type args struct {
		ctx   contextx.IContext
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
	ctx := contextx.New(context.Background(), contextx.WithTenantID("0"), contextx.WithBKUsername("test"))
	type args struct {
		ctx                contextx.IContext
		bizID              int64
		serviceTemplateIDs []int64
		moduleIDs          []int64
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
				moduleIDs:          []int64{},
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
			got, err := h.FindHostByServiceTemplate(tt.args.ctx, tt.args.bizID, tt.args.page, tt.args.serviceTemplateIDs, tt.args.moduleIDs)
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

// Test_handler_WatchResourceEvent...
func Test_handler_WatchResourceEvent(t *testing.T) {
	ctx := contextx.New(context.Background(), contextx.WithTenantID("0"), contextx.WithBKUsername("test"))
	type args struct {
		ctx contextx.IContext
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
			hostCursor := ""
			hostRelationCursor := ""
			for i := 0; i < 2; i++ {
				host, err := h.WatchHostResourceEvent(tt.args.ctx, hostCursor)
				if (err != nil) != tt.wantErr {
					t.Errorf("WatchHostResourceEvent() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				t.Logf("host resource event: %#v", host)
				for _, event := range host {
					hostCursor = event.Cursor
				}

				hostRelation, err := h.WatchHostRelationResourceEvent(tt.args.ctx, hostRelationCursor)
				if (err != nil) != tt.wantErr {
					t.Errorf("WatchHostRelation() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				t.Logf("host relation resource event: %#v", hostRelation)
				for _, event := range hostRelation {
					hostRelationCursor = event.Cursor
				}
			}
		})
	}
}

// Test_handler_FindHostWithCondition...
func TestHandler_FindHostWithCondition(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("0"), contextx.WithBKUsername("admin"))

	type args struct {
		nCtx contextx.IContext
		page types.Page
		cond *types.HostStaticExactCondition
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal-filter-by-addressing",
			args: args{
				nCtx: nCtx,
				page: types.Page{
					Offset: 0,
					Limit:  50,
					Sort:   "",
				},
				cond: &types.HostStaticExactCondition{
					StaticExactInclude: &types.HostStaticExactFields{
						Addressing: []types.Addressing{
							types.AddressingStatic,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			// 实际的主机ip 是 127.0.0.1,127.0.0.2
			name: "normal-filter-by-inner-ip",
			args: args{
				nCtx: nCtx,
				page: types.Page{
					Offset: 0,
					Limit:  50,
					Sort:   "",
				},
				cond: &types.HostStaticExactCondition{
					StaticExactInclude: &types.HostStaticExactFields{
						InnerIP: []string{"127.0.0.2"},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.FindHostWithCondition(tt.args.nCtx, tt.args.page, tt.args.cond)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindHostWithCondition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, host := range got {
				t.Logf("index: %d, host: %#v", index, *host)
			}
		})
	}
}
