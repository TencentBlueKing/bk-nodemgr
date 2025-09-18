/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package tenant ...
package tenant

import (
	"context"
	"reflect"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/contextvalues"
)

// TestGetMode ...
func TestGetMode(t *testing.T) {
	tests := []struct {
		name string
		want Mode
	}{
		{
			name: "default",
			want: ModeSingle,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetMode(); got != tt.want {
				t.Errorf("GetMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSetMode ...
func TestSetMode(t *testing.T) {
	type args struct {
		modes []Mode
	}
	tests := []struct {
		name string
		args args
		want Mode
	}{
		{
			name: "normal",
			args: args{
				modes: []Mode{ModeMultiple},
			},
			want: ModeMultiple,
		},
		{
			name: "multiple mode changes",
			args: args{
				modes: []Mode{ModeMultiple, ModeSingle},
			},
			want: ModeMultiple,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, mode := range tt.args.modes {
				SetMode(mode)
			}

			if GetMode() != tt.want {
				t.Errorf("SetMode() = %v, want %v", GetMode(), tt.want)
			}
		})
	}
}

// TestGetID ...
func TestGetID(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "nil context",
			args: args{
				ctx: nil,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "without tenant id",
			args: args{
				ctx: context.Background(),
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "with tenant id",
			args: args{
				ctx: context.WithValue(context.Background(), contextvalues.KeyTenantID, "test"),
			},
			want:    "test",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetID(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSetID ...
func TestSetID(t *testing.T) {
	type args struct {
		ctx      context.Context
		tenantID string
	}
	tests := []struct {
		name    string
		args    args
		want    context.Context
		wantErr bool
	}{
		{
			name: "nil context",
			args: args{
				ctx:      nil,
				tenantID: "",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "normal",
			args: args{
				ctx:      context.Background(),
				tenantID: "test",
			},
			want:    context.WithValue(context.Background(), contextvalues.KeyTenantID, "test"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SetID(tt.args.ctx, tt.args.tenantID)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetID() got = %v, want %v", got, tt.want)
			}

			t.Logf("SetID() got = %#v", got)
		})
	}
}
