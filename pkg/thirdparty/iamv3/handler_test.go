/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package iamv3

import (
	"testing"

	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
)

func TestNew(t *testing.T) {
	validAPIGWUserConfig := apigwclient.UserConfig{
		AppConfig: apigwclient.NewAppConfig(
			[]string{"https://example.com/api/bk-iam/prod"},
			"test-app",
			"test-secret",
		),
		AuthMode:   apigwclient.AuthModeUn,
		BKUsername: "admin",
	}

	httpClient, err := restclient.NewHTTPClient(&ssl.TLSConfig{
		InsecureSkipVerify: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	clientCap := &restclient.Capability{
		HTTPClient:           httpClient,
		Discover:             restdiscovery.NewDiscovery("iam-v3", []string{"https://example.com/api/bk-iam/prod"}),
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
		MetricOpts:           restclient.MetricOption{},
	}

	tests := []struct {
		name    string
		cap     *restclient.Capability
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cap:  clientCap,
			config: &Config{
				APIGWUserConfig: validAPIGWUserConfig,
				SystemID:        "bk_nodemgr",
				CallbackPath:    "/api/v3/iam/callback",
			},
			wantErr: false,
		},
		{
			name: "missing system ID",
			cap:  clientCap,
			config: &Config{
				APIGWUserConfig: validAPIGWUserConfig,
				SystemID:        "",
				CallbackPath:    "/api/v3/iam/callback",
			},
			wantErr: true,
		},
		{
			name: "missing callback path",
			cap:  clientCap,
			config: &Config{
				APIGWUserConfig: validAPIGWUserConfig,
				SystemID:        "bk_nodemgr",
				CallbackPath:    "",
			},
			wantErr: true,
		},
		{
			name: "invalid APIGWUserConfig",
			cap:  clientCap,
			config: &Config{
				APIGWUserConfig: apigwclient.UserConfig{
					AppConfig: apigwclient.NewAppConfig(
						[]string{"https://example.com/api/bk-iam/prod"},
						"", // empty app code
						"test-secret",
					),
					AuthMode:   apigwclient.AuthModeUn,
					BKUsername: "admin",
				},
				SystemID:     "bk_nodemgr",
				CallbackPath: "/api/v3/iam/callback",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := New(tt.cap, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && handler == nil {
				t.Errorf("New() returned nil handler for valid config")
			}
			if !tt.wantErr && handler.cli == nil {
				t.Errorf("New() handler.cli is nil for valid config")
			}
		})
	}
}
