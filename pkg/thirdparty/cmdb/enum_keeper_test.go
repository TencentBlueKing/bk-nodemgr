/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cmdb

import (
	"context"
	"os"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	"github.com/joho/godotenv"
)

func testPrivateCli(t *testing.T) *cli {
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

	client, err := newClient(clientCap, &Config{
		SupplierAccount: os.Getenv("BK_SUPPLIER_ACCOUNT"),
		HeaderSetter:    testHeaderSetter{},
	})
	if err != nil {
		t.Fatal(err)
	}

	return client
}

// Test_enumOSTypeKeeper_getValue test get value.
func Test_enumOSTypeKeeper_getValue(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name       string
		args       args
		wantResult string
	}{
		{
			name: "test1",
			args: args{
				key: "1",
			},
			wantResult: criteria.OSLinux,
		},
		{
			name: "test2",
			args: args{
				key: "2",
			},
			wantResult: criteria.OSWindows,
		},
		{
			name: "test3",
			args: args{
				key: "3",
			},
			wantResult: criteria.OSAix,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper := newOSTypeKeeper(testPrivateCli(t))
			if err := keeper.update(context.Background()); err != nil {
				t.Errorf("update() error = %v", err)
			}

			got := keeper.getValue(tt.args.key)
			if got != tt.wantResult {
				t.Errorf("getValue() got = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

// Test_enumOSTypeKeeper_getKey test get key.
func Test_enumOSTypeKeeper_getKey(t *testing.T) {
	type args struct {
		cache map[string]string
		value string
	}
	tests := []struct {
		name       string
		args       args
		wantResult string
	}{
		{
			name: "test1",
			args: args{
				value: criteria.OSLinux,
			},
			wantResult: "1",
		},
		{
			name: "test2",
			args: args{
				value: criteria.OSWindows,
			},
			wantResult: "2",
		},
		{
			name: "test3",
			args: args{
				value: criteria.OSAix,
			},
			wantResult: "3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper := newOSTypeKeeper(testPrivateCli(t))
			if err := keeper.update(context.Background()); err != nil {
				t.Errorf("update() error = %v", err)
			}

			got := keeper.getKey(tt.args.value)
			if got != tt.wantResult {
				t.Errorf("getKey() got = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

// Test_enumCloudVendorKeeper_getValue test get value.
func Test_enumCloudVendorKeeper_getValue(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name       string
		args       args
		wantResult string
	}{
		{
			name: "test1",
			args: args{
				key: "1",
			},
			wantResult: "AWS",
		},
		{
			name: "test2",
			args: args{
				key: "2",
			},
			wantResult: "Tencent Cloud",
		},
		{
			name: "test3",
			args: args{
				key: "3",
			},
			wantResult: "Google Cloud",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper := newCloudVendorKeeper(testPrivateCli(t))
			if err := keeper.update(context.Background()); err != nil {
				t.Errorf("update() error = %v", err)
			}

			got := keeper.getValue(tt.args.key)
			if got != tt.wantResult {
				t.Errorf("getValue() got = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

// Test_enumCloudVendorKeeper_getKey test get key.
func Test_enumCloudVendorKeeper_getKey(t *testing.T) {
	type args struct {
		cache map[string]string
		value string
	}
	tests := []struct {
		name       string
		args       args
		wantResult string
	}{
		{
			name: "test1",
			args: args{
				value: "AWS",
			},
			wantResult: "1",
		},
		{
			name: "test2",
			args: args{
				value: "Tencent Cloud",
			},
			wantResult: "2",
		},
		{
			name: "test3",
			args: args{
				value: "Google Cloud",
			},
			wantResult: "3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper := newCloudVendorKeeper(testPrivateCli(t))
			if err := keeper.update(context.Background()); err != nil {
				t.Errorf("update() error = %v", err)
			}

			got := keeper.getKey(tt.args.value)
			if got != tt.wantResult {
				t.Errorf("getKey() got = %v, want %v", got, tt.wantResult)
			}
		})
	}
}
