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
		TenantID:     os.Getenv("BK_APIGW_TENANT_ID"),
		HeaderSetter: testHeaderSetter{},
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
func Test_handler_SearchNetArea(t *testing.T) {
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
