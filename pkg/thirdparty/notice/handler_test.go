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

package notice

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tracing"
	"github.com/joho/godotenv"
)

// testHandler initialize a test notice handler.
func testHandler(t *testing.T) IHandler {
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

	h, err := New(clientCap, &Config{
		APIGWUserConfig: apigwUserConfig,
	})
	if err != nil {
		t.Fatal(err)
	}

	return h
}

// Test_Handler_GetCurrentAnnouncements test GetCurrentAnnouncements method.
func Test_Handler_GetCurrentAnnouncements(t *testing.T) {
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
			h := testHandler(t)
			got, err := h.GetCurrentAnnouncements(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCurrentAnnouncements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, ann := range got {
				t.Logf("announcement: ID=%d, Title=%s, Type=%s, StartTime=%s, EndTime=%s",
					ann.ID, ann.Title, ann.AnnounceType, ann.StartTime, ann.EndTime)
			}
		})
	}
}

// Test_parseTime test parseTime function with different time formats.
func Test_parseTime(t *testing.T) {
	tests := []struct {
		name    string
		timeStr string
		wantErr bool
	}{
		{
			name:    "RFC3339 format",
			timeStr: "2024-01-15T10:30:00Z",
			wantErr: false,
		},
		{
			name:    "datetime format",
			timeStr: "2024-01-15 10:30:00",
			wantErr: false,
		},
		{
			name:    "date format",
			timeStr: "2024-01-15",
			wantErr: false,
		},
		{
			name:    "empty string",
			timeStr: "",
			wantErr: false,
		},
		{
			name:    "invalid format",
			timeStr: "invalid-time",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTime(tt.timeStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.timeStr != "" {
				if got.IsZero() {
					t.Errorf("parseTime() returned zero time for valid input: %s", tt.timeStr)
				}
				t.Logf("parseTime(%s) = %s", tt.timeStr, got.Format(time.RFC3339))
			}
		})
	}
}

// Test_Handler_registerApplication test registerApplication method.
func Test_Handler_registerApplication(t *testing.T) {
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
		{
			name: "nil context",
			args: args{
				ctx: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testHandler(t).(*Handler)
			got, err := h.registerApplication(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("registerApplication() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				t.Logf("registration: ID=%d, Code=%s, Name=%s", got.ID, got.Code, got.Name)
			}
		})
	}
}
