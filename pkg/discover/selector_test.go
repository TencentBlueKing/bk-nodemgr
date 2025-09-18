/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package discover

import (
	"reflect"
	"testing"
)

// TestRandomSelector_Select test the random selector.
func TestRandomSelector_Select(t *testing.T) {
	type args struct {
		endpoints []Endpoint
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
				endpoints: []Endpoint{
					{
						IPV4: "127.0.0.1",
						IPV6: "::1",
						Port: 1001,
					},
					{
						IPV4: "127.0.0.2",
						IPV6: "::1",
						Port: 1002,
					},
					{
						IPV4: "127.0.0.3",
						IPV6: "::1",
						Port: 1003,
					},
				},
			},
			want: []Endpoint{
				{
					IPV4: "127.0.0.1",
					IPV6: "::1",
					Port: 1001,
				},
				{
					IPV4: "127.0.0.2",
					IPV6: "::1",
					Port: 1002,
				},
				{
					IPV4: "127.0.0.3",
					IPV6: "::1",
					Port: 1003,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := NewRandomSelector()
			got, err := selector.Select(tt.args.endpoints)
			if (err != nil) != tt.wantErr {
				t.Errorf("RandomSelector.Select() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			found := false
			for _, endpoint := range tt.want {
				if reflect.DeepEqual(got, endpoint) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("RandomSelector.Select() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestRoundRobinSelector_Select test round robin selector.
func TestRoundRobinSelector_Select(t *testing.T) {
	type args struct {
		endpoints []Endpoint
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
				endpoints: []Endpoint{
					{
						IPV4: "127.0.0.1",
						IPV6: "::1",
						Port: 1001,
					},
					{
						IPV4: "127.0.0.2",
						IPV6: "::1",
						Port: 1002,
					},
					{
						IPV4: "127.0.0.3",
						IPV6: "::1",
						Port: 1003,
					},
				},
			},
			want: []Endpoint{
				{
					IPV4: "127.0.0.1",
					IPV6: "::1",
					Port: 1001,
				},
				{
					IPV4: "127.0.0.2",
					IPV6: "::1",
					Port: 1002,
				},
				{
					IPV4: "127.0.0.3",
					IPV6: "::1",
					Port: 1003,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := NewRoundRobinSelector()

			for _, want := range tt.want {
				got, err := selector.Select(tt.args.endpoints)
				if (err != nil) != tt.wantErr {
					t.Errorf("RandomSelector.Select() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if !reflect.DeepEqual(got, want) {
					t.Errorf("RandomSelector.Select() got = %v, want %v", got, want)
				}
			}
		})
	}
}
