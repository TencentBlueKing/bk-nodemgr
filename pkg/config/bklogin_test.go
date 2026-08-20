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

package config

import (
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/stretchr/testify/require"
)

func TestApplicationService_validateBKLogin(t *testing.T) {
	validAPIGatewayClient := APIGatewayClient{
		Endpoints: []string{"https://bklogin.example.com/api/bk-login/prod"},
		AppCode:   "bk-nodemgr",
		AppSecret: "app-secret",
		User:      "admin",
		AuthMode:  "un",
	}

	tests := []struct {
		name       string
		tenantMode tenant.Mode
		authType   LoginAuthType
		client     APIGatewayClient
		wantErr    string
	}{
		{
			name:       "single bk_token does not require gateway credentials",
			tenantMode: tenant.ModeSingle,
			authType:   LoginAuthTypeBKToken,
			client: APIGatewayClient{
				Endpoints: []string{"https://bklogin.example.com"},
			},
		},
		{
			name:       "multiple bk_ticket does not require gateway credentials",
			tenantMode: tenant.ModeMultiple,
			authType:   LoginAuthTypeBKTicket,
			client: APIGatewayClient{
				Endpoints: []string{"https://bklogin.example.com"},
			},
		},
		{
			name:       "multiple bk_token requires gateway credentials",
			tenantMode: tenant.ModeMultiple,
			authType:   LoginAuthTypeBKToken,
			client: APIGatewayClient{
				Endpoints: []string{"https://bklogin.example.com/api/bk-login/prod"},
			},
			wantErr: "app code of api-gateway is empty",
		},
		{
			name:       "multiple bk_token requires gateway user",
			tenantMode: tenant.ModeMultiple,
			authType:   LoginAuthTypeBKToken,
			client: APIGatewayClient{
				Endpoints: []string{"https://bklogin.example.com/api/bk-login/prod"},
				AppCode:   "bk-nodemgr",
				AppSecret: "app-secret",
				AuthMode:  "un",
			},
			wantErr: "user is empty in un auth mode",
		},
		{
			name:       "multiple bk_token accepts full gateway config",
			tenantMode: tenant.ModeMultiple,
			authType:   LoginAuthTypeBKToken,
			client:     validAPIGatewayClient,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &ApplicationService{
				TenantMode: tt.tenantMode,
				BKSaas: BKSaas{BKLogin: BKLogin{
					LoginURL:         "https://bklogin.example.com/login",
					AuthType:         tt.authType,
					APIGatewayClient: tt.client,
				}},
			}

			err := svc.validateBKLogin()
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}
