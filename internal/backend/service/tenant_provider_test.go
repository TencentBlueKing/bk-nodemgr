/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations
 * under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 * to the current version of the project delivered to anyone in the future.
 */

package service

import (
	"errors"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/stretchr/testify/require"
)

type fakeUserManagerHandler struct {
	tenants []*types.Tenant
	err     error
}

func (handler fakeUserManagerHandler) GetBKUsernameByLoginName(contextx.IContext, string) (string, error) {
	return "", nil
}

func (handler fakeUserManagerHandler) ListALLTenants(contextx.IContext) ([]*types.Tenant, error) {
	return handler.tenants, handler.err
}

func TestUserManagerTenantIDProviderListTenantIDs(t *testing.T) {
	testErr := errors.New("list tenants failed")
	testCases := []struct {
		name    string
		handler fakeUserManagerHandler
		wantIDs []string
		wantErr error
	}{
		{
			name: "maps business tenants",
			handler: fakeUserManagerHandler{
				tenants: []*types.Tenant{
					{ID: "tenant-a", Enabled: true},
					{ID: "tenant-b", Enabled: false},
				},
			},
			wantIDs: []string{"tenant-a", "tenant-b"},
		},
		{
			name: "wraps tenant list error",
			handler: fakeUserManagerHandler{
				err: testErr,
			},
			wantErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			provider := userManagerTenantIDProvider{handler: testCase.handler}

			tenantIDs, err := provider.ListTenantIDs(contextx.Background())

			if testCase.wantErr != nil {
				require.ErrorIs(t, err, testCase.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, testCase.wantIDs, tenantIDs)
		})
	}
}
