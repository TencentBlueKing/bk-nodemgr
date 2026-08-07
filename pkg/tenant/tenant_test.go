/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package tenant

import (
	"errors"
	"sync"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

type tenantIDProviderFn func(contextx.IContext) ([]string, error)

func (fn tenantIDProviderFn) ListTenantIDs(nCtx contextx.IContext) ([]string, error) {
	return fn(nCtx)
}

func resetTenantIDProvider() {
	tenantStorage.provider = new(singleModeTenantIDStorage)
	tenantStorage.Once = sync.Once{}
}

func TestListTenantIDs(t *testing.T) {
	resetTenantIDProvider()

	tenantIDs, err := ListTenantIDs(contextx.Background())
	if err != nil {
		t.Fatalf("ListTenantIDs() error = %v, want nil", err)
	}

	if len(tenantIDs) != 1 || tenantIDs[0] != SingleModeTenantID {
		t.Fatalf("ListTenantIDs() = %v, want [%s]", tenantIDs, SingleModeTenantID)
	}
}

func TestSetTenantIDProvider(t *testing.T) {
	tests := []struct {
		name    string
		prepare func()
		set     ITenantIDProvider
		wantErr bool
	}{
		{
			name:    "nil provider",
			set:     nil,
			wantErr: true,
		},
		{
			name: "set custom provider",
			set: tenantIDProviderFn(func(_ contextx.IContext) ([]string, error) {
				return []string{SystemTenantID}, nil
			}),
		},
		{
			name: "set provider twice",
			prepare: func() {
				_ = SetTenantIDProvider(tenantIDProviderFn(func(_ contextx.IContext) ([]string, error) {
					return []string{SystemTenantID}, nil
				}))
			},
			set: tenantIDProviderFn(func(_ contextx.IContext) ([]string, error) {
				return []string{"other"}, nil
			}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetTenantIDProvider()
			if tt.prepare != nil {
				tt.prepare()
			}

			err := SetTenantIDProvider(tt.set)
			if (err != nil) != tt.wantErr {
				t.Fatalf("SetTenantIDProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetAllTenantIDs(t *testing.T) {
	tests := []struct {
		name string
		set  ITenantIDProvider
		want []string
	}{
		{
			name: "provider returns tenant ids",
			set: tenantIDProviderFn(func(_ contextx.IContext) ([]string, error) {
				return []string{SystemTenantID}, nil
			}),
			want: []string{SystemTenantID},
		},
		{
			name: "provider returns error",
			set: tenantIDProviderFn(func(_ contextx.IContext) ([]string, error) {
				return nil, errors.New("list tenant ids failed")
			}),
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetTenantIDProvider()
			if err := SetTenantIDProvider(tt.set); err != nil {
				t.Fatalf("SetTenantIDProvider() error = %v, want nil", err)
			}

			got := GetAllTenantIDs()
			if len(got) != len(tt.want) {
				t.Fatalf("GetAllTenantIDs() = %v, want %v", got, tt.want)
			}
			for idx := range got {
				if got[idx] != tt.want[idx] {
					t.Fatalf("GetAllTenantIDs() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
