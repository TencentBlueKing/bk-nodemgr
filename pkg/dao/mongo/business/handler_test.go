//go:build integration

/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package business ...
package business

import (
	"context"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/testsuite/support"
)

// testClient ...
func testClient(t *testing.T) IHandler {
	t.Helper()

	_, db := support.RequireMongoDatabase(t)
	return New(db)
}

// prepareData for all tests.
func prepareData(t *testing.T, nCtx contextx.IContext) IHandler {
	t.Helper()

	tenantID := nCtx.TenantID()
	h := testClient(t)
	err := h.UpsertMany(nCtx,
		&types.Business{
			TenantID: tenantID,
			BizID:    90001,
			BizName:  "test-name-90001",
		},
		&types.Business{
			TenantID: tenantID,
			BizID:    90002,
			BizName:  "test-name-same",
		},
		&types.Business{
			TenantID: tenantID,
			BizID:    90003,
			BizName:  "test-name-same",
		},
	)
	if err != nil {
		t.Fatalf("prepareData() error = %v", err)
	}

	return h
}

func Test_handler_List_without_limit(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("test"))

	h := prepareData(t, nCtx)

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "test",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := h.List(nCtx, types.Page{})
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, v := range got {
				t.Logf("List() got = %v", v)
			}
		})
	}
}

// Test_handler_Count covers count method.
func Test_handler_Count(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("test"))

	h := prepareData(t, nCtx)

	tests := []struct {
		name      string
		optFn     []OptFn
		wantTotal int64
		wantErr   bool
	}{
		{
			name:      "normal",
			optFn:     nil,
			wantTotal: -1,
			wantErr:   false,
		},
		{
			name:      "filter by biz id",
			optFn:     []OptFn{WithBizID(90001, 90002)},
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:      "filter by biz name",
			optFn:     []OptFn{WithFuzzyBizName("test-name-90001", "test-name-same")},
			wantTotal: 3,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.Count(nCtx, tt.optFn...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Count() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantTotal > 0 && got != tt.wantTotal {
				t.Errorf("Count() got = %d, wantTotal %d", got, tt.wantTotal)
				return
			}
			t.Logf("Count() got = %d", got)
		})
	}
}

// Test_handler_List covers list method.
func Test_handler_List(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("test"))

	h := prepareData(t, nCtx)

	tests := []struct {
		name      string
		page      types.Page
		optFn     []OptFn
		wantTotal int64
		wantNum   int64
		wantErr   bool
	}{
		{
			name: "normal",
			page: types.Page{
				Offset: 0,
				Limit:  0,
			},
			optFn:     nil,
			wantTotal: -1,
			wantNum:   -1,
			wantErr:   false,
		},
		{
			name: "filter by biz id",
			page: types.Page{
				Offset: 0,
				Limit:  1,
			},
			optFn:     []OptFn{WithBizID(90001, 90002)},
			wantTotal: 2,
			wantNum:   1,
			wantErr:   false,
		},
		{
			name: "filter by biz name",
			page: types.Page{
				Offset: 1,
				Limit:  1,
			},
			optFn:     []OptFn{WithFuzzyBizName("test-name-same")},
			wantTotal: 2,
			wantNum:   1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, total, err := h.List(nCtx, tt.page, tt.optFn...)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantTotal > 0 && tt.wantTotal != total {
				t.Errorf("List() total = %d, wantTotal %d", total, tt.wantTotal)
				return
			}
			t.Logf("List() total = %d", total)

			if tt.wantNum > 0 && tt.wantNum != int64(len(got)) {
				t.Errorf("List() num = %d, wantNum %d", len(got), tt.wantNum)
				return
			}
			t.Logf("List() num = %d", len(got))

			for _, v := range got {
				t.Logf("List() got = %v", v)
			}
		})
	}
}

// Test_handler_UpsertMany ...
func Test_handler_UpsertMany(t *testing.T) {
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("test"))

	type args struct {
		nCtx contextx.IContext
		bizs []*types.Business
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				nCtx: nCtx,
				bizs: []*types.Business{
					{
						TenantID: "test",
						BizID:    1,
						BizName:  "test",
					},
					{
						TenantID: "test",
						BizID:    2,
						BizName:  "test2",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "tenant no match",
			args: args{
				nCtx: nCtx,
				bizs: []*types.Business{
					{
						TenantID: "single",
						BizID:    1,
						BizName:  "test",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			err := h.UpsertMany(tt.args.nCtx, tt.args.bizs...)
			if err != nil {
				t.Logf("UpsertMany() error = %v", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("UpsertMany() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
