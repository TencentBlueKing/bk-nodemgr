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

// Package notice provides handlers to operate notice api.
package notice

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tracing"
	"github.com/joho/godotenv"
)

// LoadAuthHeader load auth header from environment variables.
func LoadAuthHeader() (apigwclient.UserConfig, error) {
	apigwAuthHeader := os.Getenv("BK_APIGW_AUTHHEADER")
	header := make(map[string]string)
	if err := json.Unmarshal([]byte(apigwAuthHeader), &header); err != nil {
		return apigwclient.UserConfig{}, err
	}

	apigwUserConfig := apigwclient.UserConfig{
		AppConfig: apigwclient.NewAppConfig(
			[]string{os.Getenv("BK_APIGW_ENDPOINT")},
			header["bk_app_code"],
			header["bk_app_secret"]),
		AuthMode:  "un",
		LoginName: header["bk_username"],
	}

	return apigwUserConfig, nil
}

// testClient initialize a test notice client.
func testClient(t *testing.T) *cli {
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
		Name:                 "notice",
		HTTPClient:           httpClient,
		Discover:             restdiscovery.NewDiscovery("apigateway", []string{os.Getenv("BK_APIGW_ENDPOINT")}),
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
		MetricOpts:           restclient.MetricOption{},
		TraceSvc: func() tracing.IService {
			traceSvc, err := tracing.G().NewService(tracing.ServiceConfig{
				ServiceName: "notice",
				SampleRate:  0,
			})
			if err != nil {
				t.Fatal(err)
			}
			return traceSvc
		}(),
	}

	apigwUserConfig, err := LoadAuthHeader()
	if err != nil {
		t.Fatal(err)
	}

	cli, err := newClient(clientCap, &Config{
		APIGWUserConfig: apigwUserConfig,
	})
	if err != nil {
		t.Fatal(err)
	}

	return cli
}

// Test_cli_getCurrentAnnouncements test getCurrentAnnouncements method.
func Test_cli_getCurrentAnnouncements(t *testing.T) {
	ctx := contextx.New(context.Background(), contextx.WithTenantID("0"), contextx.WithBKUsername("test"))

	type args struct {
		ctx    contextx.IContext
		params *getCurrentAnnouncementsParams
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
				params: &getCurrentAnnouncementsParams{
					Platform: "bk-nodemgr",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := testClient(t)
			got, err := c.getCurrentAnnouncements(tt.args.ctx, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCurrentAnnouncements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, ann := range got {
				t.Logf("announcement: %#v", ann)
			}
		})
	}
}

// Test_cli_registerApplication test registerApplication method.
func Test_cli_registerApplication(t *testing.T) {
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
			name: "base",
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := testClient(t)
			got, err := c.registerApplication(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("registerApplication() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("registration: %#v", got)
		})
	}
}

// Test_cli_getHeader test getHeader method.
func Test_cli_getHeader(t *testing.T) {
	ctx := contextx.New(context.Background(), contextx.WithTenantID("0"), contextx.WithBKUsername("test"))

	type args struct {
		ctx contextx.IContext
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "base",
			args: args{
				ctx: ctx,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := testClient(t)
			got := c.getHeader(tt.args.ctx)

			// Verify required headers are set
			if got.Get("X-Bk-Tenant-Id") == "" {
				t.Error("getHeader() missing X-Bk-Tenant-Id header")
			}
			if got.Get("X-Bkapi-Request-Id") == "" {
				t.Error("getHeader() missing X-Bkapi-Request-Id header")
			}
			if got.Get("X-Bkapi-Authorization") == "" {
				t.Error("getHeader() missing X-Bkapi-Authorization header")
			}

			t.Logf("headers: %#v", got)
		})
	}
}
