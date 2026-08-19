//go:build integration

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

package cmdb

import (
	"context"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
)

func testPrivateCli(t *testing.T) *cli {
	t.Helper()

	return newIntegrationClient(t, newIntegrationTarget(t))
}

// Test_enumOSTypeKeeper_getValue test get value.
func Test_enumOSTypeKeeper_getValue(t *testing.T) {
	ctx := contextx.New(context.Background(), contextx.WithTenantID("0"), contextx.WithBKUsername("test"))

	type args struct {
		ctx contextx.IContext
		key string
	}
	tests := []struct {
		name       string
		args       args
		wantResult string
	}{
		{
			name: "test1",
			args: args{
				ctx: ctx,
				key: "1",
			},
			wantResult: string(criteria.OSLinux),
		},
		{
			name: "test2",
			args: args{
				ctx: ctx,
				key: "2",
			},
			wantResult: string(criteria.OSWindows),
		},
		{
			name: "test3",
			args: args{
				ctx: ctx,
				key: "3",
			},
			wantResult: string(criteria.OSAix),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper := newOSTypeKeeper(testPrivateCli(t))
			if err := keeper.update(tt.args.ctx); err != nil {
				t.Errorf("update() error = %v", err)
			}

			got := keeper.getValue(tt.args.key)
			if got != tt.wantResult {
				t.Errorf("getValue() got = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

// Test_enumOSTypeKeeper_getKey test get key.
func Test_enumOSTypeKeeper_getKey(t *testing.T) {
	ctx := contextx.New(context.Background(), contextx.WithTenantID("0"), contextx.WithBKUsername("test"))

	type args struct {
		ctx   contextx.IContext
		cache map[string]string
		value string
	}
	tests := []struct {
		name       string
		args       args
		wantResult string
	}{
		{
			name: "test1",
			args: args{
				ctx:   ctx,
				value: string(criteria.OSLinux),
			},
			wantResult: "1",
		},
		{
			name: "test2",
			args: args{
				ctx:   ctx,
				value: string(criteria.OSWindows),
			},
			wantResult: "2",
		},
		{
			name: "test3",
			args: args{
				ctx:   ctx,
				value: string(criteria.OSAix),
			},
			wantResult: "3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper := newOSTypeKeeper(testPrivateCli(t))
			if err := keeper.update(tt.args.ctx); err != nil {
				t.Errorf("update() error = %v", err)
			}

			got := keeper.getKey(tt.args.value)
			if got != tt.wantResult {
				t.Errorf("getKey() got = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

// Test_enumCloudVendorKeeper_getValue test get value.
func Test_enumCloudVendorKeeper_getValue(t *testing.T) {
	ctx := contextx.New(context.Background(), contextx.WithTenantID("0"), contextx.WithBKUsername("test"))

	type args struct {
		ctx contextx.IContext
		key string
	}
	tests := []struct {
		name       string
		args       args
		wantResult string
	}{
		{
			name: "test1",
			args: args{
				ctx: ctx,
				key: "1",
			},
			wantResult: "AWS",
		},
		{
			name: "test2",
			args: args{
				ctx: ctx,
				key: "2",
			},
			wantResult: "Tencent Cloud",
		},
		{
			name: "test3",
			args: args{
				ctx: ctx,
				key: "3",
			},
			wantResult: "Google Cloud",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper := newCloudVendorKeeper(testPrivateCli(t))
			if err := keeper.update(tt.args.ctx); err != nil {
				t.Errorf("update() error = %v", err)
			}

			got := keeper.getValue(tt.args.key)
			if got != tt.wantResult {
				t.Errorf("getValue() got = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

// Test_enumCloudVendorKeeper_getKey test get key.
func Test_enumCloudVendorKeeper_getKey(t *testing.T) {
	ctx := contextx.New(context.Background(), contextx.WithTenantID("0"), contextx.WithBKUsername("test"))

	type args struct {
		ctx   contextx.IContext
		cache map[string]string
		value string
	}
	tests := []struct {
		name       string
		args       args
		wantResult string
	}{
		{
			name: "test1",
			args: args{
				ctx:   ctx,
				value: "AWS",
			},
			wantResult: "1",
		},
		{
			name: "test2",
			args: args{
				ctx:   ctx,
				value: "Tencent Cloud",
			},
			wantResult: "2",
		},
		{
			name: "test3",
			args: args{
				ctx:   ctx,
				value: "Google Cloud",
			},
			wantResult: "3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper := newCloudVendorKeeper(testPrivateCli(t))
			if err := keeper.update(tt.args.ctx); err != nil {
				t.Errorf("update() error = %v", err)
			}

			got := keeper.getKey(tt.args.value)
			if got != tt.wantResult {
				t.Errorf("getKey() got = %v, want %v", got, tt.wantResult)
			}
		})
	}
}
