/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package workflowdef ...
package workflowdef

import (
	"context"
	"testing"
	"time"

	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
)

// Test_syncHostFromCMDB_Do ...
func Test_syncHostFromCMDB_Do(t *testing.T) {
	ctx, _ := tenant.SetID(context.Background(), "bk_nodeman")

	type fields struct {
		cmdbHandler cmdb.IHandler
		topoStorage topoStg.IStorage
	}
	type args struct {
		ctx *operengine.ActionInstContext
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "normal",
			fields: fields{},
			args: args{
				ctx: &operengine.ActionInstContext{
					Ctx: ctx,
					Data: &operengine.ActionInstData{
						TriggerID:  "",
						OperInstID: "",
						Name:       "",
						Index:      0,
						Lifecycle: &operengine.ActInstLifeCycle{
							State:     operengine.ActionInstStateRunning,
							StartedAt: time.Time{},
							EndedAt:   time.Time{},
							StoppedAt: time.Time{},
						},
						Messages: nil,
						Content: map[string]any{
							"biz_id":   2,
							"biz_name": "test",
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewActionSyncHostFromCMDB(tt.fields.cmdbHandler, tt.fields.topoStorage)
			if err := s.Do(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Do() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("Data %#v", tt.args.ctx.Data)
		})
	}
}
