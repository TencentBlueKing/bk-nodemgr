/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cmdb

import (
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// TestHandler_GetTargetByScopeSetTemplate
func TestHandler_GetTargetByScopeSetTemplate(t *testing.T) {
	type args struct {
		nCtx  contextx.IContext
		scope *types.ScopeSetTemplate
	}
	var tests = []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal_test_host",
			args: args{
				nCtx: contextx.New(contextx.Background(), contextx.WithBKUsername("admin")),
				scope: &types.ScopeSetTemplate{
					Granularity:    types.TargetGranularityHost,
					TenantID:       "default",
					BizID:          2,
					SetTemplateIDs: []int64{1},
					SetIDs:         []int64{},
					Filter:         types.TargetFilter{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.GetTargetByScopeSetTemplate(tt.args.nCtx, tt.args.scope)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTargetByScopeSetTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for idx, target := range got {
				t.Logf("%d: %+v", idx, target)
			}
		})
	}
}

// TestHandler_GetTargetByScopeServiceTemplate ...
func TestHandler_GetTargetByScopeServiceTemplate(t *testing.T) {
	type args struct {
		nCtx  contextx.IContext
		scope *types.ScopeServiceTemplate
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal_test_host",
			args: args{
				nCtx: contextx.New(contextx.Background(), contextx.WithBKUsername("admin")),
				scope: &types.ScopeServiceTemplate{
					Granularity:        types.TargetGranularityHost,
					BizID:              2,
					TenantID:           "default",
					Filter:             types.TargetFilter{},
					ServiceTemplateIDs: []int64{1},
					ModuleIDs:          []int64{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.GetTargetByScopeServiceTemplate(tt.args.nCtx, tt.args.scope)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTargetByScopeServiceTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for idx, target := range got {
				t.Logf("%d: %+v", idx, target)
			}
		})
	}
}

// TestHandler_GetTargetByScopeInstance ...
func TestHandler_GetTargetByScopeInstance(t *testing.T) {
	nCtx := contextx.New(contextx.Background(), contextx.WithBKUsername("admin"))
	type args struct {
		nCtx  contextx.IContext
		scope *types.ScopeInstance
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal_test_host",
			args: args{
				nCtx: nCtx,
				scope: &types.ScopeInstance{
					Granularity: types.TargetGranularityHost,
					BizID:       2,
					TenantID:    "default",
					InstanceIDs: []int64{2},
					Filter:      types.TargetFilter{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.GetTargetByScopeInstance(tt.args.nCtx, tt.args.scope)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTargetByScopeInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for idx, target := range got {
				t.Logf("%d: %+v", idx, target)
			}
		})
	}
}

func TestHandler_GetTargetByScopeDynamicGroup(t *testing.T) {
	nCtx := contextx.New(contextx.Background(), contextx.WithBKUsername("admin"))
	type args struct {
		nCtx  contextx.IContext
		scope *types.ScopeDynamicGroup
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal_test_host",
			args: args{
				nCtx: nCtx,
				scope: &types.ScopeDynamicGroup{
					Granularity: types.TargetGranularityHost,
					BizID:       12,
					TenantID:    "default",
					DynamicGroupIDs: []string{
						"8e2a1084-d593-11f0-b7cd-0692be2071e4",
					},
					Filter: types.TargetFilter{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.GetTargetByScopeDynamicGroup(tt.args.nCtx, tt.args.scope)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTargetByScopeDynamicGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for idx, target := range got {
				t.Logf("%d: %+v", idx, target)
			}
		})
	}
}
