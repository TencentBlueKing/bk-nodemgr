/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package etcddiscover

import (
	"os"
	"reflect"
	"sort"
	"sync"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/discover"
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

		if err := testProvider.Start(); err != nil {
			t.Fatal(err)
		}

		if err := testProvider.Register(discover.ServiceNameBackend, discover.Instance{
			ID:   "1",
			Name: "test1",
			Endpoints: map[discover.EndpointName]discover.Endpoint{
				discover.EndpointNameBackendBasic: {
					IPV4: "192.168.186.2",
					Port: 8000,
				},
				discover.EndpointNameBackendCallback: {
					IPV4: "192.168.186.2",
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
					IPV4: "192.168.186.3",
					Port: 8000,
				},
				discover.EndpointNameBackendCallback: {
					IPV4: "192.168.186.3",
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
					IPV4: "192.168.186.4",
					Port: 8000,
				},
				discover.EndpointNameBackendCallback: {
					IPV4: "192.168.186.4",
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
							IPV4: "192.168.186.2",
							Port: 8000,
						},
						discover.EndpointNameBackendCallback: {
							IPV4: "192.168.186.2",
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
							IPV4: "192.168.186.3",
							Port: 8000,
						},
						discover.EndpointNameBackendCallback: {
							IPV4: "192.168.186.3",
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
							IPV4: "192.168.186.4",
							Port: 8000,
						},
						discover.EndpointNameBackendCallback: {
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
			p := testProviderEtcd(t)
			p.list(tt.args.serviceName)
			got, _ := p.GetAllService(tt.args.serviceName)

			for _, instance := range got {
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
						break
					}
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
		},
		{
			name: "backend-callback",
			args: args{
				serviceName:  discover.ServiceNameBackend,
				endpointName: discover.EndpointNameBackendCallback,
			},
			want: []discover.Endpoint{
				{
					IPV4: "192.168.186.2",
					Port: 8001,
				},
				{
					IPV4: "192.168.186.3",
					Port: 8001,
				},
				{
					IPV4: "192.168.186.4",
					Port: 8001,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testProviderEtcd(t)
			p.list(tt.args.serviceName)
			got, _ := p.GetAllEndpoint(tt.args.serviceName, tt.args.endpointName)

			sort.Slice(got, func(i, j int) bool {
				return got[i].IPV4 < got[j].IPV4
			})
			sort.Slice(tt.want, func(i, j int) bool {
				return tt.want[i].IPV6 < tt.want[j].IPV6
			})
			sort.Slice(tt.want, func(i, j int) bool {
				return tt.want[i].Port < tt.want[j].Port
			})

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllEndpoint() got = %v, want %v", got, tt.want)
			}
		})
	}
}
