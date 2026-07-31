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

// Package stopoperinst ...
package stopoperinst

import (
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/testsuite/support"
)

// testClient creates a stopping operation instance handler backed by an isolated integration database.
func testClient(t *testing.T) IHandler {
	t.Helper()

	_, db := support.RequireMongoDatabase(t)
	return New(db)
}

// Test_handler_Upsert ...
func Test_handler_Upsert(t *testing.T) {
	type args struct {
		nCtx       contextx.IContext
		operInstID string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				nCtx:       contextx.Background(),
				operInstID: "temp-test",
			},
			wantErr: false,
		},
		{
			name: "nil nCtx",
			args: args{
				nCtx:       nil,
				operInstID: "temp-test",
			},
			wantErr: true,
		},
		{
			name: "empty operInstID",
			args: args{
				nCtx:       contextx.Background(),
				operInstID: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			if err := h.Upsert(tt.args.nCtx, tt.args.operInstID); (err != nil) != tt.wantErr {
				t.Errorf("Upsert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_handler_FindAll ...
func Test_handler_FindAll(t *testing.T) {
	type args struct {
		nCtx contextx.IContext
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				nCtx: contextx.Background(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.FindAll(tt.args.nCtx)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, v := range got {
				t.Logf("FindAll() got = %v", v)
			}
		})
	}
}

// Test_handler_FindByIDs ...
func Test_handler_FindByIDs(t *testing.T) {
	nCtx := contextx.Background()
	h := testClient(t)

	if err := h.Upsert(nCtx, "included-oper-inst"); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	if err := h.Upsert(nCtx, "excluded-oper-inst"); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	got, err := h.FindByIDs(nCtx, "included-oper-inst")
	if err != nil {
		t.Fatalf("FindByIDs() error = %v", err)
	}
	if len(got) != 1 || got[0] != "included-oper-inst" {
		t.Fatalf("FindByIDs() got = %v, want [included-oper-inst]", got)
	}
}
