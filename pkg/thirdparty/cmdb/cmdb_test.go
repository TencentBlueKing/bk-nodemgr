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

// testClient ...
func testCCClient(t *testing.T) *cli {
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

	h, err := newClient(clientCap, &Config{
		SupplierAccount: os.Getenv("BK_SUPPLIER_ACCOUNT"),
		HeaderSetter:    testHeaderSetter{},
	})
	if err != nil {
		t.Fatal(err)
	}

	return h
}

// Test_cmdb_listBizHosts ...
func Test_handler_listBizHosts(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx context.Context
		req *ListBizHostsReq
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
				req: &ListBizHostsReq{
					BKBizID: 2,
					Page: Page{
						Start: 0,
						Limit: 500,
						Sort:  "",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)
			got, err := h.listBizHosts(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("listBizHosts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, host := range got.Info {
				t.Logf("host: %#v", host)
			}
		})
	}
}

// Test_cmdb_searchBusiness ...
func Test_handler_searchBusiness(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx context.Context
		req *SearchBusinessReq
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
				req: &SearchBusinessReq{
					Fields: []string{"bk_biz_id", "bk_biz_name"},
					Page: Page{
						Start: 0,
						Limit: 500,
						Sort:  "",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)
			got, err := h.searchBusiness(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("searchBusiness() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, biz := range got.Info {
				t.Logf("biz: %#v", biz)
			}
		})
	}
}

// Test_cmdb_cloudArea...
func Test_cmdb_cloudArea(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx           context.Context
		cloudAreaName string
		cloudVendor   string
		page          Page
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
				cloudAreaName: "test_nodemgr_cloud",
				cloudVendor:   "5",
				page: Page{
					Start: 0,
					Limit: 500,
					Sort:  "",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)

			createCloudAreaReq := &CreateCloudAreaReq{
				TenantID:      "0",
				BKCloudName:   tt.args.cloudAreaName,
				BKCloudVendor: tt.args.cloudVendor,
			}
			created, err := h.createCloudArea(tt.args.ctx, createCloudAreaReq)
			if (err != nil) != tt.wantErr {
				t.Errorf("createCloudArea() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("createCloudArea got: %#v", created)

			searchCloudReq := &SearchCloudAreaReq{
				TenantID: "0",
				Page:     tt.args.page,
			}
			search, err := h.searchCloudArea(tt.args.ctx, searchCloudReq)
			if (err != nil) != tt.wantErr {
				t.Errorf("searchCloudArea() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, cloud := range search.Info {
				t.Logf("searchCloudArea got: %#v", cloud)
			}

			updateCloudReq := &UpdateCloudAreaReq{
				TenantID:      "0",
				BKCloudID:     created.Created.ID,
				BKCloudName:   tt.args.cloudAreaName + "_1",
				BKCloudVendor: "1",
			}
			err = h.updateCloudArea(tt.args.ctx, updateCloudReq)
			if (err != nil) != tt.wantErr {
				t.Errorf("updateCloudArea() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("updateCloudArea success")

			search, err = h.searchCloudArea(tt.args.ctx, searchCloudReq)
			if (err != nil) != tt.wantErr {
				t.Errorf("searchCloudArea() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, cloud := range search.Info {
				t.Logf("searchCloudArea got: %#v", cloud)
			}

			deleteCloudReq := &DeleteCloudAreaReq{
				TenantID:  "0",
				BKCloudID: created.Created.ID,
			}
			err = h.deleteCloudArea(tt.args.ctx, deleteCloudReq)
			if (err != nil) != tt.wantErr {
				t.Errorf("deleteCloudArea() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("deleteNetworkArea success")

			search, err = h.searchCloudArea(tt.args.ctx, searchCloudReq)
			if (err != nil) != tt.wantErr {
				t.Errorf("searchCloudArea() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, cloud := range search.Info {
				t.Logf("searchCloudArea got: %#v", cloud)
			}
		})
	}
}

// Test_cmdb_searchBizInstTopo...
func Test_cmdb_searchBizInstTopo(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx context.Context
		req *SearchBizInstTopoReq
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
				req: &SearchBizInstTopoReq{
					TenantID: "0",
					BKBizID:  2,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)

			got, err := h.searchBizInstTopo(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("searchBizInstTopo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, topo := range *got {
				t.Logf("topo: %#v", topo)
			}
		})
	}
}

// Test_cmdb_getBizInternalModule...
func Test_cmdb_getBizInternalModule(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx context.Context
		req *GetBizInternalModuleReq
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
				req: &GetBizInternalModuleReq{
					TenantID: "0",
					BKBizID:  2,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)
			got, err := h.getBizInternalModule(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("getBizInternalModule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("got: %#v", *got)
		})
	}
}

// Test_cmdb_findTopoNodePaths...
func Test_cmdb_findTopoNodePaths(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx context.Context
		req *FindTopoNodePathsReq
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
				req: &FindTopoNodePathsReq{
					TenantID: "0",
					BKBizID:  2,
					BKNodes: []*Node{
						{
							BKObjID:  "module",
							BKInstID: 190,
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)
			got, err := h.findTopoNodePaths(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("findTopoNodePaths() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, path := range *got {
				t.Logf("index: %d, path: %#v", index, path)
			}
		})
	}
}

// Test_cmdb_findModuleBatch...
func Test_cmdb_findModuleBatch(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx context.Context
		req *FindModuleBatchReq
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
				req: &FindModuleBatchReq{
					TenantID: "0",
					BKBizID:  2,
					BKIDs:    []int64{1},
					Fields: []string{
						"bk_module_id",
						"bk_module_name",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)
			got, err := h.findModuleBatch(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("findModuleBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, module := range *got {
				t.Logf("index: %d, module: %#v", index, module)
			}
		})
	}
}

// Test_cmdb_searchObjectAttribute...
func Test_cmdb_searchObjectAttribute(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx context.Context
		req *SearchObjectAttributeReq
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
				req: &SearchObjectAttributeReq{
					TenantID: "0",
					BKBizID:  DefaultBusinessID,
					BKObjID:  "plat",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)
			got, err := h.searchObjectAttribute(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("searchObjectAttribute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, attr := range *got {
				t.Logf("index: %d, attr: %#v", index, *attr)
			}
		})
	}
}

// Test_cmdb_listServiceTemplate...
func Test_cmdb_listServiceTemplate(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx context.Context
		req *ListServiceTemplateReq
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
				req: &ListServiceTemplateReq{
					TenantID:            "0",
					BKBizID:             2,
					ServiceCategoryID:   0,
					ServiceTemplateName: "",
					ServiceTemplateIDs:  []int64{},
					IsExact:             false,
					Page: Page{
						Start: 0,
						Limit: 500,
						Sort:  "",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)
			got, err := h.listServiceTemplate(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("listServiceTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, serviceTemplate := range got.Info {
				t.Logf("index: %d, serviceTemplate: %#v", index, *serviceTemplate)
			}
		})
	}
}

// Test_cmdb_listServiceInstance...
func Test_cmdb_listServiceInstance(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx context.Context
		req *ListServiceInstanceReq
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
				req: &ListServiceInstanceReq{
					TenantID:                 "0",
					BKBizID:                  2,
					BKModuleID:               0,
					BKHostIDs:                []int64{1},
					ServiceInstanceFuzzyName: "",
					Page: Page{
						Start: 0,
						Limit: 500,
						Sort:  "",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)
			got, err := h.listServiceInstance(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("listServiceInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, serviceInstance := range got.Info {
				t.Logf("index: %d, serviceInstance: %#v", index, *serviceInstance)
			}
		})
	}
}

// Test_cmdb_listProcessInstance...
func Test_cmdb_listProcessInstance(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx context.Context
		req *ListProcessInstanceReq
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
				req: &ListProcessInstanceReq{
					TenantID:          "0",
					BKBizID:           2,
					ServiceInstanceID: 1,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)
			got, err := h.listProcessInstance(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("listProcessInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, processInstance := range *got {
				t.Logf("index: %d, processInstance: %#v", index, *processInstance)
			}
		})
	}
}

// Test_cmdb_listProcTemplate...
func Test_cmdb_listProcTemplate(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx context.Context
		req *ListProcTemplateReq
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
				req: &ListProcTemplateReq{
					BKBizID:            2,
					ServiceTemplateID:  1,
					ProcessTemplateIDs: []int64{},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)
			got, err := h.listProcTemplate(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("listProcTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, procTemplate := range got.Info {
				t.Logf("index: %d, procTemplate: %#v", index, *procTemplate)
			}
		})
	}
}

// Test_cmdb_findSetBatch...
func Test_cmdb_findSetBatch(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx context.Context
		req *FindSetBatchReq
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
				req: &FindSetBatchReq{
					TenantID: "0",
					BKBizID:  2,
					BKIDs:    []int64{1, 2, 3},
					Fields: []string{
						"bk_set_id",
						"bk_set_name",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)
			got, err := h.findSetBatch(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("findSetBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, set := range *got {
				t.Logf("index: %d, set: %#v", index, *set)
			}
		})
	}
}

// Test_cmdb_searchSet...
func Test_cmdb_searchSet(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx context.Context
		req *SearchSetReq
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
				req: &SearchSetReq{
					TenantID: "0",
					BKBizID:  2,
					Fields: []string{
						"bk_set_id",
						"bk_set_name",
					},
					Page: Page{
						Start: 0,
						Limit: 500,
						Sort:  "",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)
			got, err := h.searchSet(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("searchSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, set := range got.Info {
				t.Logf("index: %d, set: %#v", index, *set)
			}
		})
	}
}

// Test_cmdb_searchModule...
func Test_cmdb_searchModule(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx context.Context
		req *SearchModuleReq
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
				req: &SearchModuleReq{
					TenantID: "0",
					BKBizID:  2,
					BKSetID:  1,
					Fields: []string{
						"bk_module_id",
						"bk_module_name",
					},
					Page: Page{
						Start: 0,
						Limit: 500,
						Sort:  "",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)
			got, err := h.searchModule(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("searchModule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, module := range got.Info {
				t.Logf("index: %d, module: %#v", index, *module)
			}
		})
	}
}

// Test_cmdb_findHostTopoRelation...
func Test_cmdb_findHostTopoRelation(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx context.Context
		req *FindHostTopoRelationReq
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
				req: &FindHostTopoRelationReq{
					TenantID:    "0",
					BKBizID:     2,
					BKSetIDs:    []int64{},
					BKModuleIDs: []int64{},
					BKHostIDs:   []int64{},
					Page: Page{
						Start: 0,
						Limit: 500,
						Sort:  "",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)
			got, err := h.findHostTopoRelation(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("findHostTopoRelation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, relation := range got.Data {
				t.Logf("index: %d, relation: %#v", index, *relation)
			}
		})
	}
}

// Test_cmdb_findHostBizRelation...
func Test_cmdb_findHostBizRelations(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx context.Context
		req *FindHostBizRelationsReq
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
				req: &FindHostBizRelationsReq{
					TenantID: "0",
					BKBizID:  2,
					BKHostID: []int64{1},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)
			got, err := h.findHostBizRelations(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("findHostBizRelations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, relation := range *got {
				t.Logf("index: %d, relation: %#v", index, *relation)
			}
		})
	}
}

// Test_cmdb_findHostByServiceTemplate...
func Test_cmdb_findHostByServiceTemplate(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx context.Context
		req *FindHostByServiceTemplateReq
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
				req: &FindHostByServiceTemplateReq{
					TenantID:             "0",
					BKBizID:              2,
					BKServiceTemplateIDs: []int64{1},
					BKModuleIDs:          []int64{},
					Fields: []string{
						"bk_host_id",
						"bk_cloud_id",
					},
					Page: Page{
						Start: 0,
						Limit: 500,
						Sort:  "",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)
			got, err := h.findHostByServiceTemplate(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("findHostByServiceTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, host := range got.Info {
				t.Logf("index: %d, host: %#v", index, *host)
			}
		})
	}
}

// Test_cmdb_findHostBySetTemplate...
func Test_cmdb_findHostBySetTemplate(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx context.Context
		req *FindHostBySetTemplateReq
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
				req: &FindHostBySetTemplateReq{
					TenantID:         "0",
					BKBizID:          2,
					BKSetTemplateIDs: []int64{1},
					BKSetIDs:         []int64{},
					Fields: []string{
						"bk_host_id",
						"bk_cloud_id",
					},
					Page: Page{
						Start: 0,
						Limit: 500,
						Sort:  "",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)
			got, err := h.findHostBySetTemplate(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("findHostBySetTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, host := range got.Info {
				t.Logf("index: %d, host: %#v", index, *host)
			}
		})
	}
}

// Test_cmdb_findHostByTopo...
func Test_cmdb_findHostByTopo(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx context.Context
		req *FindHostByTopoReq
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
				req: &FindHostByTopoReq{
					TenantID: "0",
					BKBizID:  2,
					BKObjID:  "set",
					BKInstID: 1,
					Fields: []string{
						"bk_host_id",
						"bk_cloud_id",
					},
					Page: Page{
						Start: 0,
						Limit: 500,
						Sort:  "",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)
			got, err := h.findHostByTopo(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("findHostByTopo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, host := range got.Info {
				t.Logf("index: %d, host: %#v", index, *host)
			}
		})
	}
}

// Test_cmdb_findHostRelationsWithTopo...
func Test_cmdb_findHostRelationsWithTopo(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx context.Context
		req *FindHostRelationsWithTopoReq
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
				req: &FindHostRelationsWithTopoReq{
					TenantID:  "0",
					BKBizID:   2,
					BKObjID:   "set",
					BKInstIDs: []int64{1},
					Fields: []string{
						"bk_host_id",
						"bk_set_id",
					},
					Page: Page{
						Start: 0,
						Limit: 500,
						Sort:  "",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)
			got, err := h.findHostRelationsWithTopo(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("findHostRelationsWithTopo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, relation := range got.Info {
				t.Logf("index: %d, relation: %#v", index, relation)
			}
		})
	}
}

// Test_cmdb_listServiceInstanceDetail...
func Test_cmdb_listServiceInstanceDetail(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx context.Context
		req *ListServiceInstanceDetailReq
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
				req: &ListServiceInstanceDetailReq{
					TenantID: "0",
					BKBizID:  2,
					Page: Page{
						Start: 0,
						Limit: 500,
						Sort:  "",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)
			got, err := h.listServiceInstanceDetail(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("listServiceInstanceDetail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, svcInst := range got.Info {
				t.Logf("index: %d, host: %#v", index, *svcInst)
			}
		})
	}
}

// Test_cmdb_getMainlineObjectTopo...
func Test_cmdb_getMainlineObjectTopo(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx context.Context
		req *GetMainlineObjectTopoReq
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
				req: &GetMainlineObjectTopoReq{
					TenantID: "0",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)
			got, err := h.getMainlineObjectTopo(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("getMainlineObjectTopo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, topo := range *got {
				t.Logf("index: %d, topo: %#v", index, topo)
			}
		})
	}
}

// Test_cmdb_listBizHostsTopo...
func Test_cmdb_listBizHostsTopo(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx context.Context
		req *ListBizHostsTopoReq
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
				req: &ListBizHostsTopoReq{
					TenantID: "0",
					BKBizID:  2,
					Fields:   []string{},
					Page: Page{
						Start: 0,
						Limit: 500,
						Sort:  "",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)
			got, err := h.listBizHostsTopo(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("listBizHostsTopo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, host := range got.Info {
				t.Logf("index: %d, host: %#v, topo: %#v", index, host.Host, host.Topo)
			}
		})
	}
}

// Test_cmdb_listServiceInstanceByHost...
func Test_cmdb_listServiceInstanceByHost(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx context.Context
		req *ListServiceInstanceByHostReq
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
				req: &ListServiceInstanceByHostReq{
					TenantID: "0",
					BKBizID:  2,
					BKHostID: 1,
					Page: Page{
						Start: 0,
						Limit: 500,
						Sort:  "",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)
			got, err := h.listServiceInstanceByHost(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("listServiceInstanceByHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, svcInst := range got.Info {
				t.Logf("index: %d, service instance: %#v", index, *svcInst)
			}
		})
	}
}

// Test_cmdb_listServiceInstanceBySetTemplate...
func Test_cmdb_listServiceInstanceBySetTemplate(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx context.Context
		req *ListServiceInstanceBySetTemplateReq
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
				req: &ListServiceInstanceBySetTemplateReq{
					TenantID:      "0",
					BKBizID:       2,
					SetTemplateID: 1,
					Page: Page{
						Start: 0,
						Limit: 500,
						Sort:  "",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)
			got, err := h.listServiceInstanceBySetTemplate(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("listServiceInstanceBySetTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, svcInst := range got.Info {
				t.Logf("index: %d, service instance: %#v", index, *svcInst)
			}
		})
	}
}

// Test_cmdb_listSetTemplate...
func Test_cmdb_listSetTemplate(t *testing.T) {
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
			name: "base",
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

			got, err := h.ListSetTemplate(tt.args.ctx, tt.args.bizID, tt.args.page)
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

// Test_cmdb_createAndUpdateHost...
func Test_cmdb_createAndUpdateHost(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx   context.Context
		bizID int64
		hosts []*CreateHostInfo
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "base",
			args: args{
				ctx:   ctx,
				bizID: 2,
				hosts: []*CreateHostInfo{
					{
						BKHostInnerIP:     "1.1.1.1",
						BKCloudID:         0,
						BKOSType:          "1",
						BKCpuArchitecture: "x86",
						BKAddressing:      "static",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)

			createReq := &AddHostToBusinessIdleReq{
				TenantID:   "0",
				BKBizID:    tt.args.bizID,
				BKHostList: tt.args.hosts,
			}
			created, err := h.addHostToBusinessIdle(tt.args.ctx, createReq)
			if (err != nil) != tt.wantErr {
				t.Errorf("addHostToBusinessIdle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, hostID := range created.BKHostIDs {
				t.Logf("index: %d, host id: %#v", index, hostID)
			}

			updateHostInfo := &UpdateHostProperties{
				BKHostID: created.BKHostIDs[0],
			}
			updateHostInfo.Properties.BKComment = "test"
			updateHostInfo.Properties.BKHostName = "test"
			updatePropertys := &BatchUpdateHostReq{
				TenantID: "0",
				Update:   []*UpdateHostProperties{updateHostInfo},
			}

			// update host
			err = h.batchUpdateHost(tt.args.ctx, updatePropertys)
			if (err != nil) != tt.wantErr {
				t.Errorf("batchUpdateHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var changeNetworkAreaID int64 = 3
			updateHostCloud := &UpdateHostCloudAreaFieldReq{
				TenantID:  "0",
				BKHostIDs: created.BKHostIDs,
				BKCloudID: changeNetworkAreaID,
				BKBizID:   tt.args.bizID,
			}
			err = h.updateHostCloudAreaField(tt.args.ctx, updateHostCloud)
			if (err != nil) != tt.wantErr {
				t.Errorf("updateHostCloudAreaField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			hostAgentID := &HostAgentIDInfo{
				BKHostID:  created.BKHostIDs[0],
				BKAgentID: "xxxxxxxxxxxxxxxxxxxxxxxxx",
			}

			bindHostAgentIDs := &BindHostAgentReq{
				TenantID: "0",
				List:     []*HostAgentIDInfo{hostAgentID},
			}
			err = h.bindHostAgent(tt.args.ctx, bindHostAgentIDs)
			if (err != nil) != tt.wantErr {
				t.Errorf("bindHostAgent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			unbindHostAgentIDs := &UnbindHostAgentReq{
				TenantID: "0",
				List:     []*HostAgentIDInfo{hostAgentID},
			}
			err = h.unbindHostAgent(tt.args.ctx, unbindHostAgentIDs)
			if (err != nil) != tt.wantErr {
				t.Errorf("unbindHostAgent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

// Test_handler_dynamicGroup...
func Test_cmdb_dynamicGroup(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")

	type args struct {
		ctx       context.Context
		bizID     int64
		objID     string
		groupName string
		page      Page
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "base",
			args: args{
				ctx:       ctx,
				bizID:     2,
				objID:     "host",
				groupName: "nodemgr_test",
				page: Page{
					Start: 0,
					Limit: 500,
					Sort:  "",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)

			createReq := &CreateDynamicGroupReq{TenantID: "0"}
			createReq.BKBizID = tt.args.bizID
			createReq.BKObjID = tt.args.objID
			createReq.Name = tt.args.groupName
			createReq.Info.Condition = make([]*DynamicGroupCondition, 0)
			createReq.Info.VariableCondition = make([]*DynamicGroupCondition, 1)
			createReq.Info.VariableCondition[0] = &DynamicGroupCondition{
				BKObjID:   "host",
				Condition: make([]*FieldCondition, 1),
			}
			createReq.Info.VariableCondition[0].Condition[0] = &FieldCondition{
				Field:    "bk_host_innerip",
				Operator: "$in",
				Value:    []string{"1.1.1.1"},
			}
			group, err := h.createDynamicGroup(tt.args.ctx, createReq)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDynamicGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("groupID: %#v", group)

			executeReq := &ExecuteDynamicGroupReq{
				TenantID: "0",
				BKBizID:  tt.args.bizID,
				ID:       group.ID,
				Fields:   []string{"bk_host_id", "bk_host_name"},
				Page:     tt.args.page,
			}
			got, err := h.executeDynamicGroup(tt.args.ctx, executeReq)
			if (err != nil) != tt.wantErr {
				t.Errorf("executeDynamicGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for index, item := range got.Info {
				t.Logf("index: %d, dynamic group info: %#v", index, *item)
			}

			getReq := &GetDynamicGroupReq{
				TenantID: "0",
				BKBizID:  tt.args.bizID,
				ID:       group.ID,
			}
			dynamicGroup, err := h.getDynamicGroup(tt.args.ctx, getReq)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDynamicGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("group: %#v", dynamicGroup)

			updateReq := &UpdateDynamicGroupReq{
				TenantID: "0",
			}
			updateReq.Name = "nodemgr_test_update"
			updateReq.ID = dynamicGroup.ID
			updateReq.BKBizID = dynamicGroup.BKBizID
			err = h.updateDynamicGroup(tt.args.ctx, updateReq)
			if (err != nil) != tt.wantErr {
				t.Errorf("updateDynamicGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			dynamicGroup, err = h.getDynamicGroup(tt.args.ctx, getReq)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDynamicGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("group: %#v", got)

			deleteReq := &DeleteDynamicGroupReq{
				TenantID: "0",
				BKBizID:  tt.args.bizID,
				ID:       group.ID,
			}
			err = h.deleteDynamicGroup(tt.args.ctx, deleteReq)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteDynamicGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

// Test_cmdb_resourceWatch...
func Test_cmdb_resourceWatch(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "0")
	type args struct {
		ctx context.Context
		req *ResourceWatchReq
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
				req: &ResourceWatchReq{
					TenantID:   "0",
					BKResource: "host",
					BKFields:   []string{"bk_host_id", "bk_host_name"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testCCClient(t)

			event, err := h.resourceWatch(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("resourceWatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("event: %#v", event)
		})
	}
}
