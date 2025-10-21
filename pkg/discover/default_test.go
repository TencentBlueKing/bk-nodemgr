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
	"sync"
	"testing"
)

var once sync.Once
var testProvider *ProviderDefault

func testProviderDefault(t *testing.T) *ProviderDefault {
	once.Do(func() {
		testProvider = NewProviderDefault()

		err := testProvider.Register(ServiceNameBackend, []Instance{
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
	})

	return testProvider
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
		{
			name: "invalid service name",
			args: args{
				serviceName: "invalid",
			},
			want:    nil,
			wantErr: true,
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
		{
			name: "invalid service name",
			args: args{
				ctx:          context.Background(),
				serviceName:  "invalid",
				endpointName: EndpointNameBackendBasic,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid endpoint name",
			args: args{
				ctx:          context.Background(),
				serviceName:  ServiceNameBackend,
				endpointName: "invalid",
			},
			want:    nil,
			wantErr: true,
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
		{
			name: "invalid service name",
			args: args{
				serviceName: "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid endpoint name",
			args: args{
				serviceName:  ServiceNameBackend,
				endpointName: "invalid",
			},
			wantErr: true,
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

// TestProviderDefault_Update test Update.
func TestProviderDefault_Update(t *testing.T) {
	type args struct {
		serviceName ServiceName
		instance    Instance
	}
	tests := []struct {
		name    string
		args    args
		want    []Instance
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				serviceName: ServiceNameBackend,
				instance: Instance{
					ID:   "1",
					Name: "test1",
					Endpoints: map[EndpointName]Endpoint{
						EndpointNameBackendBasic: {
							IPV4: "192.168.186.2",
							IPV6: "::1",
							Port: 9000,
						},
						EndpointNameBackendCallback: {
							IPV4: "192.168.186.2",
							IPV6: "::1",
							Port: 9001,
						},
					},
					Meta: nil,
				},
			},
			want: []Instance{
				{
					ID:   "1",
					Name: "test1",
					Endpoints: map[EndpointName]Endpoint{
						EndpointNameBackendBasic: {
							IPV4: "192.168.186.2",
							IPV6: "::1",
							Port: 9000,
						},
						EndpointNameBackendCallback: {
							IPV4: "192.168.186.2",
							IPV6: "::1",
							Port: 9001,
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
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testProviderDefault(t)
			err := p.Update(tt.args.serviceName, tt.args.instance)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}

			got, _ := p.GetAllService(tt.args.serviceName)

			for _, instance := range got {
				found := false
				for _, want := range tt.want {
					if instance.ID == want.ID {
						if instance.Name != want.Name {
							t.Errorf("instance.Name = %v, want.Name = %v", instance.Name, want.Name)
						}

						for name, endpoint := range instance.Endpoints {
							wantEndpoint, ok := want.Endpoints[name]
							if !ok {
								t.Errorf("want.Endpoint not found: %s", name)
							}

							if endpoint.IPV4 != wantEndpoint.IPV4 {
								t.Errorf("instance.Endpoint.IPV4 = %v, want.Endpoint.IPV4 = %v", endpoint.IPV4, wantEndpoint.IPV4)
							}

							if endpoint.IPV6 != wantEndpoint.IPV6 {
								t.Errorf("instance.Endpoint.IPV6 = %v, want.Endpoint.IPV6 = %v", endpoint.IPV4, wantEndpoint.IPV4)
							}

							if endpoint.Port != wantEndpoint.Port {
								t.Errorf("instance.Endpoint.Port = %v, want.Endpoint.Port = %v", endpoint.Port, wantEndpoint.Port)
							}

							if endpoint.Nice != wantEndpoint.Nice {
								t.Errorf("instance.Endpoint.Nice = %v, want.Endpoint.Nice = %v", endpoint.Nice, wantEndpoint.Nice)
							}
						}

						found = true
						break
					}
				}

				if !found {
					t.Errorf("instance not found: %s", instance.ID)
				}
			}
		})
	}
}

// TestProviderDefault_Deregister test Deregister.
func TestProviderDefault_Deregister(t *testing.T) {
	type args struct {
		serviceName ServiceName
		instanceID  string
	}
	tests := []struct {
		name    string
		args    args
		want    []Instance
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				serviceName: ServiceNameBackend,
				instanceID:  "1",
			},
			want: []Instance{
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
			wantErr: false,
		},
		{
			name: "invalid-id",
			args: args{
				serviceName: ServiceNameBackend,
				instanceID:  "999",
			},
			want: []Instance{
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
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testProviderDefault(t)
			err := p.Deregister(tt.args.serviceName, tt.args.instanceID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Deregister() error = %v, wantErr %v", err, tt.wantErr)
			}

			got, _ := p.GetAllService(tt.args.serviceName)

			for _, instance := range got {
				found := false
				for _, want := range tt.want {
					if instance.ID == want.ID {
						if instance.Name != want.Name {
							t.Errorf("instance.Name = %v, want.Name = %v", instance.Name, want.Name)
						}

						for name, endpoint := range instance.Endpoints {
							wantEndpoint, ok := want.Endpoints[name]
							if !ok {
								t.Errorf("want.Endpoint not found: %s", name)
							}

							if endpoint.IPV4 != wantEndpoint.IPV4 {
								t.Errorf("instance.Endpoint.IPV4 = %v, want.Endpoint.IPV4 = %v", endpoint.IPV4, wantEndpoint.IPV4)
							}

							if endpoint.IPV6 != wantEndpoint.IPV6 {
								t.Errorf("instance.Endpoint.IPV6 = %v, want.Endpoint.IPV6 = %v", endpoint.IPV4, wantEndpoint.IPV4)
							}

							if endpoint.Port != wantEndpoint.Port {
								t.Errorf("instance.Endpoint.Port = %v, want.Endpoint.Port = %v", endpoint.Port, wantEndpoint.Port)
							}

							if endpoint.Nice != wantEndpoint.Nice {
								t.Errorf("instance.Endpoint.Nice = %v, want.Endpoint.Nice = %v", endpoint.Nice, wantEndpoint.Nice)
							}
						}

						found = true
						break
					}
				}

				if !found {
					t.Errorf("instance not found: %s", instance.ID)
				}
			}
		})
	}
}

// TestProviderDefault_Register test register.
func TestProviderDefault_Register(t *testing.T) {
	type args struct {
		serviceName ServiceName
		instance    Instance
	}

	// test failing case. the success case is tested in testProviderDefault.
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid id",
			args: args{
				serviceName: ServiceNameBackend,
				instance: Instance{
					ID:   "",
					Name: "test",
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
			},
			wantErr: true,
		},
		{
			name: "invalid name",
			args: args{
				serviceName: ServiceNameBackend,
				instance: Instance{
					ID:   "1",
					Name: "",
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
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testProvider.Register(tt.args.serviceName, tt.args.instance)
			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestProviderDefault_Stop test stop.
func TestProviderDefault_Stop(t *testing.T) {
	p := testProviderDefault(t)
	if err := p.Stop(); err != nil {
		t.Errorf("stop failed: %v", err)
	}
}

// TestAllErrors test all errors.
func TestAllErrors(t *testing.T) {
	if err := ErrServiceNotFound(); err == nil {
		t.Errorf("ErrServiceNotFound() should return an error")
	}

	if err := ErrEndpointNotFound(); err == nil {
		t.Errorf("ErrEndpointNotFound() should return an error")
	}

	if err := ErrInvalidInstance(); err == nil {
		t.Errorf("ErrInvalidInstance() should return an error")
	}

	if err := ErrNotRegistered(); err == nil {
		t.Errorf("ErrNotRegistered() should return an error")
	}

	if err := ErrInvalidInstanceID(); err == nil {
		t.Errorf("ErrInvalidInstanceID() should return an error")
	}

	if err := ErrInvalidInstanceName(); err == nil {
		t.Errorf("ErrInvalidInstanceName() should return an error")
	}

	if err := ErrInvalidServiceName(); err == nil {
		t.Errorf("ErrInvalidServiceName() should return an error")
	}

	if err := ErrInvalidServiceIP(); err == nil {
		t.Errorf("ErrInvalidServiceIP() should return an error")
	}

	if err := ErrInvalidServicePort(); err == nil {
		t.Errorf("ErrInvalidServicePort() should return an error")
	}

	if err := ErrInvalidSelector(); err == nil {
		t.Errorf("ErrInvalidSelector() should return an error")
	}

	if err := ErrDiscoverNotStarted(); err == nil {
		t.Errorf("ErrDiscoverNotStarted() should return an error")
	}

	if err := ErrDiscoverInternalError(); err == nil {
		t.Errorf("ErrDiscoverInternalError() should return an error")
	}
}
