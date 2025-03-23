/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package discover ...
package discover

import (
	"context"
	"reflect"
	"sort"
	"testing"
)

func testProviderDefault(t *testing.T) *ProviderDefault {
	provider := NewProviderDefault(&RandomSelector{})

	err := provider.Register(ServiceNameBackend, []Instance{
		{
			ID:   "1",
			Name: "test1",
			Endpoints: map[EndpointName]Endpoint{
				EndpointNameBackendBasic: {
					IPV4: "192.168.186.2",
					Port: 8000,
				},
				EndpointNameBackendCallback: {
					IPV4: "192.168.186.2",
					Port: 8001,
				},
			},
			Meta: nil,
		},
		{
			ID:   "2",
			Name: "test2",
			Endpoints: map[EndpointName]Endpoint{
				EndpointNameBackendBasic: {
					IPV4: "192.168.186.3",
					Port: 8000,
				},
				EndpointNameBackendCallback: {
					IPV4: "192.168.186.3",
					Port: 8001,
				},
			},
			Meta: nil,
		},
		{
			ID:   "3",
			Name: "test3",
			Endpoints: map[EndpointName]Endpoint{
				EndpointNameBackendBasic: {
					IPV4: "192.168.186.4",
					Port: 8000,
				},
				EndpointNameBackendCallback: {
					IPV4: "192.168.186.4",
					Port: 8001,
				},
			},
			Meta: nil,
		},
	}...)

	if err != nil {
		t.Fatal(err)
	}

	return provider
}

// TestProviderDefault_GetAllService test GetAllService.
func TestProviderDefault_GetAllService(t *testing.T) {
	type args struct {
		serviceName ServiceName
	}
	tests := []struct {
		name    string
		args    args
		want    []Instance
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				serviceName: ServiceNameBackend,
			},
			want: []Instance{
				{
					ID:   "1",
					Name: "test1",
					Endpoints: map[EndpointName]Endpoint{
						EndpointNameBackendBasic: {
							IPV4: "192.168.186.2",
							Port: 8000,
						},
						EndpointNameBackendCallback: {
							IPV4: "192.168.186.2",
							Port: 8001,
						},
					},
					Meta: nil,
				},
				{
					ID:   "2",
					Name: "test2",
					Endpoints: map[EndpointName]Endpoint{
						EndpointNameBackendBasic: {
							IPV4: "192.168.186.3",
							Port: 8000,
						},
						EndpointNameBackendCallback: {
							IPV4: "192.168.186.3",
							Port: 8001,
						},
					},
					Meta: nil,
				},
				{
					ID:   "3",
					Name: "test3",
					Endpoints: map[EndpointName]Endpoint{
						EndpointNameBackendBasic: {
							IPV4: "192.168.186.4",
							Port: 8000,
						},
						EndpointNameBackendCallback: {
							IPV4: "192.168.186.4",
							Port: 8001,
						},
					},
					Meta: nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testProviderDefault(t)
			got, err := p.GetAllService(tt.args.serviceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			sort.Slice(got, func(i, j int) bool {
				return got[i].ID < got[j].ID
			})
			sort.Slice(tt.want, func(i, j int) bool {
				return tt.want[i].ID < tt.want[j].ID
			})

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllService() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestProviderDefault_GetAllEndpoint test GetAllEndpoint.
func TestProviderDefault_GetAllEndpoint(t *testing.T) {
	type args struct {
		ctx          context.Context
		serviceName  ServiceName
		endpointName EndpointName
	}
	tests := []struct {
		name    string
		args    args
		want    []Endpoint
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:          context.Background(),
				serviceName:  ServiceNameBackend,
				endpointName: EndpointNameBackendBasic,
			},
			want: []Endpoint{
				{
					IPV4: "192.168.186.2",
					Port: 8000,
				},
				{
					IPV4: "192.168.186.3",
					Port: 8000,
				},
				{
					IPV4: "192.168.186.4",
					Port: 8000,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testProviderDefault(t)
			got, err := p.GetAllEndpoint(tt.args.serviceName, tt.args.endpointName)
			if err != nil {
				t.Logf("GetAllEndpoint() error = %v", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, wantEndpoint := range tt.want {
				found := false
				for _, gotEndpoint := range got {
					if reflect.DeepEqual(wantEndpoint, gotEndpoint) {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("GetAllEndpoint() want endpoint = %v, not found", wantEndpoint)
				}
			}
			t.Logf("GetAllEndpoint() got = %v", got)
		})
	}
}

// TestProviderDefault_GetEndpoint test GetEndpoint.
func TestProviderDefault_GetEndpoint(t *testing.T) {
	type args struct {
		ctx          context.Context
		serviceName  ServiceName
		endpointName EndpointName
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:          context.Background(),
				serviceName:  ServiceNameBackend,
				endpointName: EndpointNameBackendCallback,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testProviderDefault(t)
			got, err := p.GetEndpoint(tt.args.serviceName, tt.args.endpointName, NewRandomSelector())
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("GetEndpoint() got = %v", got)
		})
	}
}
