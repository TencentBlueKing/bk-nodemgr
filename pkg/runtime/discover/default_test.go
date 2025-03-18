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
	"sync"
	"testing"
)

func testProviderDefault() *ProviderDefault {
	provider := NewProviderDefault(&RandomSelector{})

	provider.Register(context.Background(), "test", Instance{
		ID: "1",
	})

	return provider
}

// TestProviderDefault_GetAllService ...
func TestProviderDefault_GetAllService(t *testing.T) {
	type fields struct {
		mutex      sync.RWMutex
		services   map[string]map[string]Instance
		watchChans map[string][]chan []Instance
		done       chan struct{}
		selector   Selector
	}
	type args struct {
		in0         context.Context
		serviceName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Instance
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testProviderDefault()
			got, err := p.GetAllService(tt.args.in0, tt.args.serviceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllService() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProviderDefault_GetAllService1(t *testing.T) {
	type fields struct {
		mutex      sync.RWMutex
		services   map[string]map[string]Instance
		watchChans map[string][]chan []Instance
		done       chan struct{}
		selector   Selector
	}
	type args struct {
		in0         context.Context
		serviceName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Instance
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProviderDefault{
				mutex:      tt.fields.mutex,
				services:   tt.fields.services,
				watchChans: tt.fields.watchChans,
				done:       tt.fields.done,
				selector:   tt.fields.selector,
			}
			got, err := p.GetAllService(tt.args.in0, tt.args.serviceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllService() got = %v, want %v", got, tt.want)
			}
		})
	}
}
