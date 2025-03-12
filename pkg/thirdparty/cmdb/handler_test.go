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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/ssl"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
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
				t.Logf("SearchNetworkArea got: %+v", *netArea)
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

// Test_handler_SearchBizInstTopo...
func Test_handler_SearchBizInstTopo(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx   context.Context
		bizID int64
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
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.SearchBizInstTopo(tt.args.ctx, tt.args.bizID)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchBizInstTopo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, topo := range got {
				t.Logf("topo: %v", topo)
			}
		})
	}
}

// Test_handler_GetBizInternalModule...
func Test_handler_GetBizInternalModule(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx   context.Context
		bizID int64
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
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.GetBizInternalModule(tt.args.ctx, tt.args.bizID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBizInternalModule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("got: %+v", *got)
		})
	}
}

// Test_handler_FindTopoNodePaths...
func Test_handler_FindTopoNodePaths(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx   context.Context
		bizID int64
		node  []*types.TopoNode
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
				node:  []*types.TopoNode{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.FindTopoNodePaths(tt.args.ctx, tt.args.bizID, tt.args.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindTopoNodePaths() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, path := range got {
				t.Logf("index: %d, path: %+v", index, path)
			}
		})
	}
}

// Test_handler_FindModuleBatch...
func Test_handler_FindModuleBatch(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx    context.Context
		bizID  int64
		ids    []int64
		fields []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:    ctx,
				bizID:  2,
				ids:    []int64{1, 2, 3},
				fields: []string{"bk_module_id", "bk_module_name"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.FindModuleBatch(tt.args.ctx, tt.args.bizID, tt.args.ids, tt.args.fields)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindModuleBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, module := range got {
				t.Logf("index: %d, module: %+v", index, module)
			}
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
				t.Logf("index: %d, attr: %+v", index, *attr)
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
				t.Logf("index: %d, attr: %+v", index, *attr)
			}
		})
	}
}

// Test_handler_ListServiceTemplate...
func Test_handler_ListServiceTemplate(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx                 context.Context
		bizID               int64
		serviceCategoryID   int64
		serviceTemplateName string
		serviceTemplateIDs  []int64
		isExact             bool
		page                types.Page
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:                 ctx,
				bizID:               2,
				serviceCategoryID:   0,
				serviceTemplateName: "",
				serviceTemplateIDs:  []int64{},
				isExact:             false,
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
			got, err := h.ListServiceTemplate(tt.args.ctx, tt.args.bizID, tt.args.serviceCategoryID,
				tt.args.serviceTemplateName, tt.args.serviceTemplateIDs, tt.args.isExact, tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListServiceTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, serviceTemplate := range got {
				t.Logf("index: %d, serviceTemplate: %+v", index, *serviceTemplate)
			}
		})
	}
}

// Test_handler_ListServiceInstance...
func Test_handler_ListServiceInstance(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx              context.Context
		bizID            int64
		moduleID         int64
		hostIDs          []int64
		svcInstFuzzyName string
		page             types.Page
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:              ctx,
				bizID:            2,
				moduleID:         0,
				hostIDs:          []int64{1},
				svcInstFuzzyName: "",
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
			got, err := h.ListServiceInstance(tt.args.ctx, tt.args.bizID, tt.args.moduleID, tt.args.hostIDs,
				tt.args.svcInstFuzzyName, tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListServiceInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, serviceInstance := range got {
				t.Logf("index: %d, serviceInstance: %+v", index, *serviceInstance)
			}
		})
	}
}

func Test_handler_ListProcessInstance(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx               context.Context
		bizID             int64
		serviceInstanceID int64
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:               ctx,
				bizID:             2,
				serviceInstanceID: 1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.ListProcessInstance(tt.args.ctx, tt.args.bizID, tt.args.serviceInstanceID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListProcessInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, processInstance := range got {
				t.Logf("index: %d, processInstance: %+v", index, *processInstance)
			}
		})
	}
}

// Test_handler_ListProcTemplate...
func Test_handler_ListProcTemplate(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx                context.Context
		bizID              int64
		serviceTemplateID  int64
		processTemplateIDs []int64
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
				serviceTemplateID:  1,
				processTemplateIDs: []int64{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.ListProcTemplate(tt.args.ctx, tt.args.bizID, tt.args.serviceTemplateID, tt.args.processTemplateIDs)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListProcTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, procTemplate := range got {
				t.Logf("index: %d, procTemplate: %+v", index, *procTemplate)
			}
		})
	}
}

// Test_handler_FindSetBatch...
func Test_handler_FindSetBatch(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx    context.Context
		bizID  int64
		setIDs []int64
		fields []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:    ctx,
				bizID:  2,
				setIDs: []int64{1, 2, 3},
				fields: []string{"bk_set_id", "bk_set_name"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.FindSetBatch(tt.args.ctx, tt.args.bizID, tt.args.setIDs, tt.args.fields)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindSetBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, set := range got {
				t.Logf("index: %d, set: %+v", index, *set)
			}
		})
	}
}

// Test_handler_SearchSet...
func Test_handler_SearchSet(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx    context.Context
		bizID  int64
		fields []string
		page   types.Page
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:    ctx,
				bizID:  2,
				fields: []string{"bk_set_id", "bk_set_name"},
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
			got, err := h.SearchSet(tt.args.ctx, tt.args.bizID, tt.args.fields, tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, set := range got {
				t.Logf("index: %d, set: %+v", index, *set)
			}
		})
	}
}

// Test_handler_SearchModule...
func Test_handler_SearchModule(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx    context.Context
		bizID  int64
		setID  int64
		fields []string
		page   types.Page
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:    ctx,
				bizID:  2,
				setID:  1,
				fields: []string{"bk_module_id", "bk_module_name"},
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
			got, err := h.SearchModule(tt.args.ctx, tt.args.bizID, tt.args.setID, tt.args.fields, tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchModule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, module := range got {
				t.Logf("index: %d, module: %+v", index, *module)
			}
		})
	}
}

// Test_handler_FindHostTopoRelation...
func Test_handler_FindHostTopoRelation(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx       context.Context
		bizID     int64
		setIDs    []int64
		moduleIDs []int64
		hostIDs   []int64
		page      types.Page
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:       ctx,
				bizID:     2,
				setIDs:    []int64{},
				moduleIDs: []int64{},
				hostIDs:   []int64{},
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
			got, err := h.FindHostTopoRelation(tt.args.ctx, tt.args.bizID, tt.args.setIDs, tt.args.moduleIDs,
				tt.args.hostIDs, tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindHostTopoRelation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, relation := range got {
				t.Logf("index: %d, relation: %+v", index, *relation)
			}
		})
	}
}

// Test_handler_FindHostBizRelation...
func Test_handler_FindHostBizRelations(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx    context.Context
		bizID  int64
		hostID []int64
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:    ctx,
				bizID:  2,
				hostID: []int64{1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.FindHostBizRelations(tt.args.ctx, tt.args.bizID, tt.args.hostID)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindHostBizRelations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, relation := range got {
				t.Logf("index: %d, relation: %+v", index, *relation)
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
		moduleIDs          []int64
		fields             []string
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
				fields: []string{
					"bk_host_id",
					"bk_cloud_id",
				},
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
			got, err := h.FindHostByServiceTemplate(tt.args.ctx, tt.args.bizID, tt.args.serviceTemplateIDs,
				tt.args.moduleIDs, tt.args.fields, tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindHostByServiceTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, host := range got {
				t.Logf("index: %d, host: %+v", index, *host)
			}
		})
	}
}

// Test_handler_FindHostBySetTemplate...
func Test_handler_FindHostBySetTemplate(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx            context.Context
		bizID          int64
		setTemplateIDs []int64
		setIDs         []int64
		fields         []string
		page           types.Page
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:            ctx,
				bizID:          2,
				setTemplateIDs: []int64{1},
				setIDs:         []int64{1},
				fields: []string{
					"bk_host_id",
					"bk_cloud_id",
				},
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
			got, err := h.FindHostBySetTemplate(tt.args.ctx, tt.args.bizID, tt.args.setTemplateIDs,
				tt.args.setIDs, tt.args.fields, tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindHostBySetTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, host := range got {
				t.Logf("index: %d, host: %+v", index, *host)
			}
		})
	}
}

// Test_handler_FindHostByTopo...
func Test_handler_FindHostByTopo(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx    context.Context
		biz    int64
		objID  string
		instID int64
		fields []string
		page   types.Page
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:    ctx,
				biz:    2,
				objID:  "set",
				instID: 1,
				fields: []string{
					"bk_host_id",
					"bk_cloud_id",
				},
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
			got, err := h.FindHostByTopo(tt.args.ctx, tt.args.biz, tt.args.objID, tt.args.instID, tt.args.fields,
				tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindHostByTopo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, host := range got {
				t.Logf("index: %d, host: %+v", index, *host)
			}
		})
	}
}

// Test_handler_FindHostRelationsWithTopo...
func Test_handler_FindHostRelationsWithTopo(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx     context.Context
		bizID   int64
		objID   string
		instIDs []int64
		fields  []string
		page    types.Page
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{

				ctx:     ctx,
				bizID:   2,
				objID:   "set",
				instIDs: []int64{1},
				fields: []string{
					"bk_host_id",
					"bk_set_id",
				},
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
			got, err := h.FindHostRelationsWithTopo(tt.args.ctx, tt.args.bizID, tt.args.objID, tt.args.instIDs,
				tt.args.fields, tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindHostRelationsWithTopo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, relation := range got {
				t.Logf("index: %d, relation: %+v", index, relation)
			}
		})
	}
}

func Test_handler_ListServiceInstanceDetail(t *testing.T) {
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
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.ListServiceInstanceDetail(tt.args.ctx, tt.args.bizID, tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListServiceInstanceDetail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, svcInst := range got {
				t.Logf("index: %d, host: %+v", index, *svcInst)
			}
		})
	}
}

// Test_handler_GetMainlineObjectTopo...
func Test_handler_CreateServiceInstance(t *testing.T) {
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
			got, err := h.GetMainlineObjectTopo(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMainlineObjectTopo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, topo := range got {
				t.Logf("index: %d, topo: %+v", index, topo)
			}
		})
	}
}

// Test_handler_ListBizHostsTopo...
func Test_handler_ListBizHostsTopo(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx    context.Context
		bizID  int64
		fields []string
		page   types.Page
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:    ctx,
				bizID:  2,
				fields: []string{},
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
			got, err := h.ListBizHostsTopo(tt.args.ctx, tt.args.bizID, tt.args.fields, tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListBizHostsTopo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, host := range got {
				t.Logf("index: %d, host: %+v", index, *host)
			}
		})
	}
}

// Test_handler_ListServiceInstanceByHost...
func Test_handler_ListServiceInstanceByHost(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx    context.Context
		bizID  int64
		hostID int64
		page   types.Page
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:    ctx,
				bizID:  2,
				hostID: 1,
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
			got, err := h.ListServiceInstanceByHost(tt.args.ctx, tt.args.bizID, tt.args.hostID, tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListServiceInstanceByHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, host := range got {
				t.Logf("index: %d, host: %+v", index, *host)
			}
		})
	}
}

// Test_handler_ListServiceInstanceBySetTemplate...
func Test_handler_ListServiceInstanceBySetTemplate(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx           context.Context
		bizID         int64
		setTemplateID int64
		page          types.Page
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:           ctx,
				bizID:         2,
				setTemplateID: 1,
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
			got, err := h.ListServiceInstanceBySetTemplate(tt.args.ctx, tt.args.bizID, tt.args.setTemplateID,
				tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListServiceInstanceBySetTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, host := range got {
				t.Logf("index: %d, host: %+v", index, *host)
			}
		})
	}
}

// Test_handler_ListServiceInstanceByModule...
func Test_handler_ListSetTemplate(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx            context.Context
		bizID          int64
		setTemplateIDs []int64
		page           types.Page
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:            ctx,
				bizID:          2,
				setTemplateIDs: []int64{1},
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

			got, err := h.ListSetTemplate(tt.args.ctx, tt.args.bizID, tt.args.setTemplateIDs, tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListSetTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, set := range got {
				t.Logf("index: %d, set: %+v", index, *set)
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
		hosts []*types.CreateHostInfo
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
				hosts: []*types.CreateHostInfo{
					{
						InnerIP:       "1.1.1.1",
						NetworkAreaID: 0,
						OSType:        "1",
						Arch:          "x86",
						Addressing:    "static",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)

			// create host
			created, err := h.AddHostToBusinessIdle(tt.args.ctx, tt.args.bizID, tt.args.hosts)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddHostToBusinessIdle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, hostID := range created {
				t.Logf("index: %d, host: %+v", index, hostID)
			}

			updatePropertys := []*types.UpdateHostProperties{
				{
					HostID:   created[0],
					HostName: "test",
				},
			}

			// update host
			err = h.BatchUpdateHost(tt.args.ctx, updatePropertys)
			if (err != nil) != tt.wantErr {
				t.Errorf("BatchUpdateHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var changeNetworkAreaID int64 = 1
			err = h.UpdateHostNetworkAreaField(tt.args.ctx, created, tt.args.bizID, changeNetworkAreaID)

			hostAgentIDs := []*types.HostAgentID{
				{
					HostID:  created[0],
					AgentID: "xxxxxxxxxxxxxxxxxxxxxxxxx",
				},
			}
			err = h.BindHostAgent(tt.args.ctx, hostAgentIDs)
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

// Test_handler_HostResourceWatch...
func Test_handler_HostResourceWatch(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx       context.Context
		hostEvent *types.ResourceWatchEvent
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
				hostEvent: &types.ResourceWatchEvent{
					Resource: types.ResourceWatchEventResourceHost,
					Fields:   []string{"bk_host_id", "bk_host_name"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)

			hostEvent, err := h.HostResourceWatch(tt.args.ctx, *tt.args.hostEvent)
			if (err != nil) != tt.wantErr {
				t.Errorf("HostResourceWatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, event := range hostEvent {
				t.Logf("index: %d, event: %+v", index, event)
			}
		})
	}
}

// Test_handler_HostRelationResourceWatch...
func Test_handler_HostRelationResourceWatch(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx               context.Context
		hostRelationEvent *types.ResourceWatchEvent
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
				hostRelationEvent: &types.ResourceWatchEvent{
					Resource: types.ResourceWatchEventResourceHostRelation,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)

			hostRelationEvent, err := h.HostRelationResourceWatch(tt.args.ctx, *tt.args.hostRelationEvent)
			if (err != nil) != tt.wantErr {
				t.Errorf("HostRelationResourceWatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, event := range hostRelationEvent {
				t.Logf("index: %d, event: %+v", index, event)
			}
		})
	}
}

// Test_handler_ProcessResourceWatch...
func Test_handler_ProcessResourceWatch(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx          context.Context
		processEvent *types.ResourceWatchEvent
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
				processEvent: &types.ResourceWatchEvent{
					Resource: types.ResourceWatchEventResourceProcess,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)

			processEvent, err := h.ProcessResourceWatch(tt.args.ctx, *tt.args.processEvent)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessResourceWatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, event := range processEvent {
				t.Logf("index: %d, event: %+v", index, event)
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
					BizID:             2,
					ObjID:             "host",
					Name:              "nodemgr_test",
					Condition:         make([]*types.DynamicGroupCondition, 0),
					VariableCondition: make([]*types.DynamicGroupCondition, 0),
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

			varCondition := &types.DynamicGroupCondition{ObjID: "host"}
			varCondition.Condition = make([]*types.FieldCondition, 1)
			varCondition.Condition[0] = &types.FieldCondition{
				Field:    "bk_host_innerip",
				Operator: "$in",
				Value:    []string{"1.1.1.1"},
			}
			tt.args.group.VariableCondition = append(tt.args.group.VariableCondition, varCondition)
			group, err := h.CreateDynamicGroup(tt.args.ctx, tt.args.group)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDynamicGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("groupID: %+v", group)

			hosts, err := h.ExecuteDynamicGroup(tt.args.ctx, tt.args.bizID, group, []string{}, tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteDynamicGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, host := range hosts {
				t.Logf("index: %d, host: %+v", index, *host)
			}

			got, err := h.GetDynamicGroup(tt.args.ctx, tt.args.bizID, group)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDynamicGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("group: %+v", got)

			got.Name = "nodemgr_test_update"
			err = h.UpdateDynamicGroup(tt.args.ctx, got)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateDynamicGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got, err = h.GetDynamicGroup(tt.args.ctx, tt.args.bizID, group)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDynamicGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("group: %+v", got)

			err = h.DeleteDynamicGroup(tt.args.ctx, tt.args.bizID, group)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteDynamicGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
