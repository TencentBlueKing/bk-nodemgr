/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package operinstdata

import (
	"context"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
)

// testActionInstData ...
func testActionInstData(t *testing.T) IActionInstData {
	return testHandler(t)
}

// Test_handler_AddActInstPrivateData ...
func Test_handler_UpdateActionInstContent(t *testing.T) {
	type args struct {
		ctx        context.Context
		operInstID string
		actionName string
		content    map[string]any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:        context.Background(),
				operInstID: "operation-inst-7bd49883-bcc9-4776-80ff-d3d37ca4143f",
				actionName: "sync_host_from_cmdb",
				content: map[string]any{
					"test": "test1",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testActionInstData(t)
			if err := h.UpdateActionInstContent(tt.args.ctx, tt.args.operInstID, tt.args.actionName, tt.args.content); (err != nil) != tt.wantErr {
				t.Errorf("UpdateActionInstContent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_handler_FindOneActionInstData ...
func Test_handler_FindOneActionInstData(t *testing.T) {
	type args struct {
		ctx        context.Context
		operInstID string
		actionName string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:        context.Background(),
				operInstID: "operation-inst-7bd49883-bcc9-4776-80ff-d3d37ca4143f",
				actionName: "action-1",
			},
			wantErr: false,
		},
		{
			name: "nil context",
			args: args{
				ctx:        nil,
				operInstID: "operation-inst-7bd49883-bcc9-4776-80ff-d3d37ca4143f",
				actionName: "sync_biz_from_cmdb",
			},
			wantErr: true,
		},
		{
			name: "empty operInstID",
			args: args{
				ctx:        context.Background(),
				operInstID: "",
				actionName: "sync_biz_from_cmdb",
			},
			wantErr: true,
		},
		{
			name: "empty actionName",
			args: args{
				ctx:        context.Background(),
				operInstID: "oper-inst-4b92daa2-6294-430f-a3ff-aa20a7c664ba",
				actionName: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testActionInstData(t)
			got, err := h.GetActionInstData(tt.args.ctx, tt.args.operInstID, tt.args.actionName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetActionInstData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("GetActionInstData() got = %#v", got)
		})
	}
}

// Test_handler_UpdateActionInstData ...
func Test_handler_UpdateActionInstData(t *testing.T) {
	type args struct {
		ctx        context.Context
		operInstID string
		data       *operengine.ActionInstData
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "",
			args: args{
				ctx:        context.Background(),
				operInstID: "operation-inst-7bd49883-bcc9-4776-80ff-d3d37ca4143f",
				data: &operengine.ActionInstData{
					TriggerID:  "trigger-2",
					OperInstID: "operation-inst-7bd49883-bcc9-4776-80ff-d3d37ca4143f",
					Name:       "sync_biz_from_cmdb",
					Index:      0,
					Lifecycle: &operengine.ActInstLifeCycle{
						State:     operengine.ActionInstStateRunning,
						StartedAt: time.Now(),
						EndedAt:   time.Now(),
						StoppedAt: time.Now(),
					},
					Messages: nil,
					Content:  nil,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testActionInstData(t)
			if err := h.UpdateActionInstData(tt.args.ctx, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("UpdateActInstLifecycle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_handler_UpdateActionInstStatus ...
func Test_handler_UpdateActionInstStatus(t *testing.T) {
	type args struct {
		ctx        context.Context
		operInstID string
		actionName string
		status     operengine.ActionInstState
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:        context.Background(),
				operInstID: "operation-inst-7bd49883-bcc9-4776-80ff-d3d37ca4143f",
				actionName: "sync_biz_from_cmdb",
				status:     "success",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testActionInstData(t)
			if err := h.UpdateActionInstStatus(tt.args.ctx, tt.args.operInstID, tt.args.actionName, tt.args.status); (err != nil) != tt.wantErr {
				t.Errorf("UpdateActionInstStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_handler_GetActInstLifecycle ...
func Test_handler_GetActInstLifecycle(t *testing.T) {
	type args struct {
		ctx        context.Context
		operInstID string
		actionName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:        context.Background(),
				operInstID: "operation-inst-7bd49883-bcc9-4776-80ff-d3d37ca4143f",
				actionName: "sync_biz_from_cmdb",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testActionInstData(t)
			got, err := h.GetActInstLifecycle(tt.args.ctx, tt.args.operInstID, tt.args.actionName)
			if err != nil {
				t.Logf("GetActInstLifecycle() error = %v", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetActInstLifecycle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("got: %v", got)
		})
	}
}

// Test_handler_AddActInstPrivateData ...
func Test_handler_AddActInstPrivateData(t *testing.T) {
	type args struct {
		ctx        context.Context
		operInstID string
		actionName string
		data       map[string]any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:        context.Background(),
				operInstID: "oper-inst-f30540ee-19b9-438a-912f-6d538bd18508",
				actionName: "sync_host_from_cmdb",
				data:       map[string]any{"test": "test1"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testActionInstData(t)
			err := h.AddActInstPrivateData(tt.args.ctx, tt.args.operInstID, tt.args.actionName, tt.args.data)
			if err != nil {
				t.Logf("AddActInstPrivateData() error = %v", err)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("AddActInstPrivateData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_handler_PushActInstMsgs ...
func Test_handler_PushActInstMsgs(t *testing.T) {
	type args struct {
		ctx        context.Context
		operInstID string
		actionName string
		msgs       []operengine.Message
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:        context.Background(),
				operInstID: "operation-inst-7bd49883-bcc9-4776-80ff-d3d37ca4143f",
				actionName: "sync_host_from_cmdb",
				msgs: []operengine.Message{
					{
						Time: time.Now(),
						Text: "test1 ",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testActionInstData(t)
			if err := h.PushActInstMsgs(tt.args.ctx, tt.args.operInstID, tt.args.actionName, tt.args.msgs...); (err != nil) != tt.wantErr {
				t.Errorf("PushActInstMsgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_handler_UpdateActInstLifecycle ...
func Test_handler_UpdateActInstLifecycle(t *testing.T) {
	type args struct {
		ctx        context.Context
		operInstID string
		actionName string
		lifecycle  *operengine.ActInstLifeCycle
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:        context.Background(),
				operInstID: "oper-inst-db0b07dc-4d1e-41bd-99e9-b4059d1a0d84",
				actionName: "sync_biz_from_cmdb",
				lifecycle: &operengine.ActInstLifeCycle{
					State:     operengine.ActionInstStateRunning,
					StartedAt: time.Now(),
					EndedAt:   time.Now(),
					StoppedAt: time.Now(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testActionInstData(t)
			err := h.UpdateActInstLifecycle(tt.args.ctx, tt.args.operInstID, tt.args.actionName, tt.args.lifecycle)
			if err != nil {
				t.Logf("UpdateActInstLifecycle() error = %v", err)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateActInstLifecycle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
