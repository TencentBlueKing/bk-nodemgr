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

package etcddiscover

import (
	"context"
	"os"
	"reflect"
	"sort"
	"sync"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/joho/godotenv"
)

var once sync.Once
var testProvider *ProviderEtcd

func testProviderEtcd(t *testing.T) *ProviderEtcd {
	once.Do(func() {
		err := godotenv.Load(".env")
		if err != nil {
			t.Fatal(err)
		}

		testProvider = NewProviderEtcd(&config.Etcd{
			Endpoints: []string{os.Getenv("ETCD_ENDPOINT")},
		})

		if err := testProvider.Start(context.Background()); err != nil {
			t.Fatal(err)
		}

		if err := testProvider.Register(discover.ServiceNameBackend, discover.Instance{
			ID:   "1",
			Name: "test1",
			Endpoints: map[discover.EndpointName]discover.Endpoint{
				discover.EndpointNameBackendBasic: {
					IPV4: "127.0.0.1",
					Port: 8000,
				},
				discover.EndpointNameBackendCallback: {
					IPV4: "127.0.0.1",
					Port: 8001,
				},
			},
			Meta: nil,
		}); err != nil {
			t.Fatal(err)
		}

		if err := testProvider.Register(discover.ServiceNameBackend, discover.Instance{
			ID:   "2",
			Name: "test2",
			Endpoints: map[discover.EndpointName]discover.Endpoint{
				discover.EndpointNameBackendBasic: {
					IPV4: "127.0.0.2",
					Port: 8000,
				},
				discover.EndpointNameBackendCallback: {
					IPV4: "127.0.0.2",
					Port: 8001,
				},
			},
			Meta: nil,
		}); err != nil {
			t.Fatal(err)
		}

		if err := testProvider.Register(discover.ServiceNameBackend, discover.Instance{
			ID:   "3",
			Name: "test3",
			Endpoints: map[discover.EndpointName]discover.Endpoint{
				discover.EndpointNameBackendBasic: {
					IPV4: "127.0.0.3",
					Port: 8000,
				},
				discover.EndpointNameBackendCallback: {
					IPV4: "127.0.0.3",
					Port: 8001,
				},
			},
			Meta: nil,
		}); err != nil {
			t.Fatal(err)
		}
	})

	return testProvider
}

// TestProviderEtcd_GetAllService
func TestProviderEtcd_GetAllService(t *testing.T) {
	type args struct {
		serviceName discover.ServiceName
	}
	tests := []struct {
		name string
		args args
		want []discover.Instance
	}{
		{
			name: "normal",
			args: args{
				serviceName: discover.ServiceNameBackend,
			},
			want: []discover.Instance{
				{
					ID:   "1",
					Name: "test1",
					Endpoints: map[discover.EndpointName]discover.Endpoint{
						discover.EndpointNameBackendBasic: {
							IPV4: "127.0.0.1",
							Port: 8000,
						},
						discover.EndpointNameBackendCallback: {
							IPV4: "127.0.0.1",
							Port: 8001,
						},
					},
					Meta: nil,
				},
				{
					ID:   "2",
					Name: "test2",
					Endpoints: map[discover.EndpointName]discover.Endpoint{
						discover.EndpointNameBackendBasic: {
							IPV4: "127.0.0.2",
							Port: 8000,
						},
						discover.EndpointNameBackendCallback: {
							IPV4: "127.0.0.2",
							Port: 8001,
						},
					},
					Meta: nil,
				},
				{
					ID:   "3",
					Name: "test3",
					Endpoints: map[discover.EndpointName]discover.Endpoint{
						discover.EndpointNameBackendBasic: {
							IPV4: "127.0.0.3",
							Port: 8000,
						},
						discover.EndpointNameBackendCallback: {
							IPV4: "127.0.0.3",
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
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testProviderEtcd(t)
			p.list(tt.args.serviceName)
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

// TestProviderEtcd_GetAllEndpoint tests GetAllEndpoint.
func TestProviderEtcd_GetAllEndpoint(t *testing.T) {
	type args struct {
		serviceName  discover.ServiceName
		endpointName discover.EndpointName
	}
	tests := []struct {
		name string
		args args
		want []discover.Endpoint
	}{
		{
			name: "backend-basic",
			args: args{
				serviceName:  discover.ServiceNameBackend,
				endpointName: discover.EndpointNameBackendBasic,
			},
			want: []discover.Endpoint{
				{
					IPV4: "127.0.0.1",
					Port: 8000,
				},
				{
					IPV4: "127.0.0.2",
					Port: 8000,
				},
				{
					IPV4: "127.0.0.3",
					Port: 8000,
				},
			},
		},
		{
			name: "backend-callback",
			args: args{
				serviceName:  discover.ServiceNameBackend,
				endpointName: discover.EndpointNameBackendCallback,
			},
			want: []discover.Endpoint{
				{
					IPV4: "127.0.0.1",
					Port: 8001,
				},
				{
					IPV4: "127.0.0.2",
					Port: 8001,
				},
				{
					IPV4: "127.0.0.3",
					Port: 8001,
				},
			},
		},
		{
			name: "invalid service name",
			args: args{
				serviceName:  "invalid",
				endpointName: discover.EndpointNameBackendBasic,
			},
			want: nil,
		},
		{
			name: "invalid endpoint name",
			args: args{
				serviceName:  discover.ServiceNameBackend,
				endpointName: "invalid",
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testProviderEtcd(t)
			p.list(tt.args.serviceName)
			got, _ := p.GetAllEndpoint(tt.args.serviceName, tt.args.endpointName)

			if len(got) == 0 && len(tt.want) == 0 {
				return
			}

			sort.Slice(got, func(i, j int) bool {
				return got[i].IPV4 < got[j].IPV4
			})
			sort.Slice(tt.want, func(i, j int) bool {
				if tt.want[i].IPV4 != tt.want[j].IPV4 {
					return tt.want[i].IPV4 < tt.want[j].IPV4
				}
				return tt.want[i].Port < tt.want[j].Port
			})

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllEndpoint(), name(%s) got = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// TestProviderEtcd_GetEndpoint get endpoint.
func TestProviderEtcd_GetEndpoint(t *testing.T) {
	type args struct {
		serviceName  discover.ServiceName
		endpointName discover.EndpointName
	}
	tests := []struct {
		name    string
		args    args
		want    []discover.Endpoint
		wantErr bool
	}{
		{
			name: "backend-basic",
			args: args{
				serviceName:  discover.ServiceNameBackend,
				endpointName: discover.EndpointNameBackendBasic,
			},
			want: []discover.Endpoint{
				{
					IPV4: "127.0.0.1",
					Port: 8000,
				},
				{
					IPV4: "127.0.0.2",
					Port: 8000,
				},
				{
					IPV4: "127.0.0.3",
					Port: 8000,
				},
			},
			wantErr: false,
		},
		{
			name: "backend-callback",
			args: args{
				serviceName:  discover.ServiceNameBackend,
				endpointName: discover.EndpointNameBackendCallback,
			},
			want: []discover.Endpoint{
				{
					IPV4: "127.0.0.1",
					Port: 8001,
				},
				{
					IPV4: "127.0.0.2",
					Port: 8001,
				},
				{
					IPV4: "127.0.0.3",
					Port: 8001,
				},
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
				serviceName:  discover.ServiceNameBackend,
				endpointName: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testProviderEtcd(t)
			p.list(tt.args.serviceName)
			got, err := p.GetEndpoint(tt.args.serviceName, tt.args.endpointName, discover.NewRandomSelector())
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			found := false
			for _, want := range tt.want {
				if reflect.DeepEqual(got, want) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("GetEndpoint() got = %v, want %v", got, tt.want)
			}

		})
	}
}

// TestProviderEtcd_SelectEndpoints test SelectEndpoints.
func TestProviderEtcd_SelectEndpoints(t *testing.T) {
	type args struct {
		serviceName  discover.ServiceName
		endpointName discover.EndpointName
		count        int
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
		wantErr bool
	}{
		{
			name: "normal - select 3 from 3",
			args: args{
				serviceName:  discover.ServiceNameBackend,
				endpointName: discover.EndpointNameBackendCallback,
				count:        3,
			},
			wantLen: 3,
			wantErr: false,
		},
		{
			name: "normal - select 2 from 3",
			args: args{
				serviceName:  discover.ServiceNameBackend,
				endpointName: discover.EndpointNameBackendCallback,
				count:        2,
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "boundary - count equals endpoints count",
			args: args{
				serviceName:  discover.ServiceNameBackend,
				endpointName: discover.EndpointNameBackendCallback,
				count:        3,
			},
			wantLen: 3,
			wantErr: false,
		},
		{
			name: "boundary - count greater than endpoints count",
			args: args{
				serviceName:  discover.ServiceNameBackend,
				endpointName: discover.EndpointNameBackendCallback,
				count:        5,
			},
			wantLen: 3,
			wantErr: false,
		},
		{
			name: "boundary - count <= 0",
			args: args{
				serviceName:  discover.ServiceNameBackend,
				endpointName: discover.EndpointNameBackendCallback,
				count:        0,
			},
			wantLen: 3,
			wantErr: false,
		},
		{
			name: "invalid service name",
			args: args{
				serviceName:  "invalid",
				endpointName: discover.EndpointNameBackendCallback,
				count:        3,
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "invalid endpoint name",
			args: args{
				serviceName:  discover.ServiceNameBackend,
				endpointName: "invalid",
				count:        3,
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testProviderEtcd(t)
			p.list(tt.args.serviceName)
			got, err := p.SelectEndpoints(tt.args.serviceName, tt.args.endpointName, tt.args.count, discover.NewRoundRobinSelector())
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectEndpoints() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("SelectEndpoints() len = %v, wantLen %v", len(got), tt.wantLen)
			}
			// 验证去重：确保没有重复的地址。
			addrMap := make(map[string]bool)
			for _, ep := range got {
				addr := ep.GetIPV4Address()
				if addrMap[addr] {
					t.Errorf("SelectEndpoints() found duplicate address: %v", addr)
				}
				addrMap[addr] = true
			}
			t.Logf("SelectEndpoints() got = %v", got)
		})
	}
}

// TestProviderEtcd_Update update instance.
func TestProviderEtcd_Update(t *testing.T) {
	type args struct {
		serviceName discover.ServiceName
		instance    discover.Instance
	}
	tests := []struct {
		name    string
		args    args
		want    []discover.Instance
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				serviceName: discover.ServiceNameBackend,
				instance: discover.Instance{
					ID:   "1",
					Name: "test1",
					Endpoints: map[discover.EndpointName]discover.Endpoint{
						discover.EndpointNameBackendBasic: {
							IPV4: "127.0.0.1",
							IPV6: "::1",
							Port: 9000,
						},
						discover.EndpointNameBackendCallback: {
							IPV4: "127.0.0.1",
							IPV6: "::1",
							Port: 9001,
						},
					},
					Meta: nil,
				},
			},
			want: []discover.Instance{
				{
					ID:   "1",
					Name: "test1",
					Endpoints: map[discover.EndpointName]discover.Endpoint{
						discover.EndpointNameBackendBasic: {
							IPV4: "127.0.0.1",
							IPV6: "::1",
							Port: 9000,
						},
						discover.EndpointNameBackendCallback: {
							IPV4: "127.0.0.1",
							IPV6: "::1",
							Port: 9001,
						},
					},
					Meta: nil,
				},
				{
					ID:   "2",
					Name: "test2",
					Endpoints: map[discover.EndpointName]discover.Endpoint{
						discover.EndpointNameBackendBasic: {
							IPV4: "127.0.0.2",
							Port: 8000,
						},
						discover.EndpointNameBackendCallback: {
							IPV4: "127.0.0.2",
							Port: 8001,
						},
					},
					Meta: nil,
				},
				{
					ID:   "3",
					Name: "test3",
					Endpoints: map[discover.EndpointName]discover.Endpoint{
						discover.EndpointNameBackendBasic: {
							IPV4: "127.0.0.3",
							Port: 8000,
						},
						discover.EndpointNameBackendCallback: {
							IPV4: "127.0.0.3",
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
			p := testProviderEtcd(t)
			err := p.Update(tt.args.serviceName, tt.args.instance)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}

			p.list(tt.args.serviceName)
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

// TestProviderEtcd_Deregister deregister test.
func TestProviderEtcd_Deregister(t *testing.T) {
	type args struct {
		serviceName discover.ServiceName
		instanceID  string
	}
	tests := []struct {
		name    string
		args    args
		want    []discover.Instance
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				serviceName: discover.ServiceNameBackend,
				instanceID:  "1",
			},
			want: []discover.Instance{
				{
					ID:   "2",
					Name: "test2",
					Endpoints: map[discover.EndpointName]discover.Endpoint{
						discover.EndpointNameBackendBasic: {
							IPV4: "127.0.0.2",
							Port: 8000,
						},
						discover.EndpointNameBackendCallback: {
							IPV4: "127.0.0.2",
							Port: 8001,
						},
					},
					Meta: nil,
				},
				{
					ID:   "3",
					Name: "test3",
					Endpoints: map[discover.EndpointName]discover.Endpoint{
						discover.EndpointNameBackendBasic: {
							IPV4: "127.0.0.3",
							Port: 8000,
						},
						discover.EndpointNameBackendCallback: {
							IPV4: "127.0.0.3",
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
				serviceName: discover.ServiceNameBackend,
				instanceID:  "999",
			},
			want: []discover.Instance{
				{
					ID:   "2",
					Name: "test2",
					Endpoints: map[discover.EndpointName]discover.Endpoint{
						discover.EndpointNameBackendBasic: {
							IPV4: "127.0.0.2",
							Port: 8000,
						},
						discover.EndpointNameBackendCallback: {
							IPV4: "127.0.0.2",
							Port: 8001,
						},
					},
					Meta: nil,
				},
				{
					ID:   "3",
					Name: "test3",
					Endpoints: map[discover.EndpointName]discover.Endpoint{
						discover.EndpointNameBackendBasic: {
							IPV4: "127.0.0.3",
							Port: 8000,
						},
						discover.EndpointNameBackendCallback: {
							IPV4: "127.0.0.3",
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
			p := testProviderEtcd(t)
			err := p.Deregister(tt.args.serviceName, tt.args.instanceID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Deregister() error = %v, wantErr %v", err, tt.wantErr)
			}

			p.list(tt.args.serviceName)
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
		serviceName discover.ServiceName
		instance    discover.Instance
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
				serviceName: discover.ServiceNameBackend,
				instance: discover.Instance{
					ID:   "",
					Name: "test",
					Endpoints: map[discover.EndpointName]discover.Endpoint{
						discover.EndpointNameBackendBasic: {
							IPV4: "127.0.0.2",
							Port: 8000,
						},
						discover.EndpointNameBackendCallback: {
							IPV4: "127.0.0.2",
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
				serviceName: discover.ServiceNameBackend,
				instance: discover.Instance{
					ID:   "1",
					Name: "",
					Endpoints: map[discover.EndpointName]discover.Endpoint{
						discover.EndpointNameBackendBasic: {
							IPV4: "127.0.0.2",
							Port: 8000,
						},
						discover.EndpointNameBackendCallback: {
							IPV4: "127.0.0.2",
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

// TestProviderEtcd_GracefulShutdown test graceful shutdown.
func TestProviderEtcd_GracefulShutdown(t *testing.T) {
	p := testProviderEtcd(t)
	if err := p.GracefulShutdown(); err != nil {
		t.Errorf("graceful shutdown failed: %v", err)
	}
}
