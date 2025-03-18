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

	err := provider.Register(context.Background(), "test", []Instance{
		{
			ID:      "1",
			Name:    "test1",
			Address: "192.168.186.2",
			Meta:    nil,
		},
		{
			ID:      "2",
			Name:    "test2",
			Address: "192.168.186.3",
			Meta:    nil,
		},
		{
			ID:      "3",
			Name:    "test3",
			Address: "192.168.186.6",
			Meta:    nil,
		},
	}...)

	if err != nil {
		t.Fatal(err)
	}

	return provider
}

// TestProviderDefault_GetAllService ...
func TestProviderDefault_GetAllService(t *testing.T) {
	type args struct {
		in0         context.Context
		serviceName string
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
				in0:         context.Background(),
				serviceName: "test",
			},
			want: []Instance{
				{
					ID:      "1",
					Name:    "test1",
					Address: "192.168.186.2",
					Meta:    nil,
				},
				{
					ID:      "2",
					Name:    "test2",
					Address: "192.168.186.3",
					Meta:    nil,
				},
				{
					ID:      "3",
					Name:    "test3",
					Address: "192.168.186.6",
					Meta:    nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testProviderDefault(t)
			got, err := p.GetAllService(tt.args.in0, tt.args.serviceName)
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

// TestProviderDefault_GetService ...
func TestProviderDefault_GetService(t *testing.T) {
	type args struct {
		ctx         context.Context
		serviceName string
		selector    Selector
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:         context.Background(),
				serviceName: "test",
				selector:    &RandomSelector{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testProviderDefault(t)
			got, err := p.GetService(tt.args.ctx, tt.args.serviceName, tt.args.selector)
			if err != nil {
				t.Logf("GetService() error = %v", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("GetService() got = %v", got)
		})
	}
}
