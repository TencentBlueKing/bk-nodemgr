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

package iamv3

import (
	"testing"

	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
)

func TestConfig_Validate(t *testing.T) {
	validAPIGWUserConfig := apigwclient.UserConfig{
		AppConfig: apigwclient.NewAppConfig(
			[]string{"https://example.com/api/bk-iam/prod"},
			"test-app",
			"test-secret",
		),
		AuthMode:  apigwclient.AuthModeUn,
		LoginName: "admin",
	}

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				APIGWUserConfig: validAPIGWUserConfig,
				SystemID:        "bk_nodemgr",
				CallbackPath:    "/api/v3/iam/callback",
			},
			wantErr: false,
		},
		{
			name: "missing system ID",
			config: Config{
				APIGWUserConfig: validAPIGWUserConfig,
				SystemID:        "",
				CallbackPath:    "/api/v3/iam/callback",
			},
			wantErr: true,
		},
		{
			name: "missing callback path",
			config: Config{
				APIGWUserConfig: validAPIGWUserConfig,
				SystemID:        "bk_nodemgr",
				CallbackPath:    "",
			},
			wantErr: true,
		},
		{
			name: "invalid APIGWUserConfig - empty app code",
			config: Config{
				APIGWUserConfig: apigwclient.UserConfig{
					AppConfig: apigwclient.NewAppConfig(
						[]string{"https://example.com/api/bk-iam/prod"},
						"", // empty app code
						"test-secret",
					),
					AuthMode:  apigwclient.AuthModeUn,
					LoginName: "admin",
				},
				SystemID:     "bk_nodemgr",
				CallbackPath: "/api/v3/iam/callback",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
