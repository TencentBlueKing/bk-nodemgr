/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodeinstall

import (
	"context"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
)

// Test_waitComplete_Do ...
func Test_waitComplete_Do(t *testing.T) {
	capability := testCapability(t)

	type fields struct {
		storage operengine.OperInstStorage
		logger  logger.Logger
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
			name: "test",
			fields: fields{
				storage: capability.OperInstStorage,
				logger:  capability.Logger,
			},
			args: args{
				ctx: &operengine.ActionInstContext{
					Ctx: context.Background(),
					Data: &operengine.ActionInstData{
						TriggerID:   "",
						OperInstID:  "oper-inst-efb1362d-73af-434b-aae2-e40aed1ed2e0",
						OperationID: "",
						Name:        "sync_biz_from_cmdb",
						Index:       0,
						Messages:    nil,
						Content: map[string]interface{}{
							"wait_time": 10,
						},
						PrivateData: nil,
						Lifecycle:   nil,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewActionWaitComplete(capability.OperInstStorage, capability.Logger)
			if err := s.Do(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Do() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
