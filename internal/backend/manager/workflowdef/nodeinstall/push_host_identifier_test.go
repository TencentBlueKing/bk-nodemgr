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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/keys"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
)

// TestPushHostIdentifier_Name ...
func TestPushHostIdentifier_Do(t *testing.T) {
	capability := testCapability(t)

	type args struct {
		ctx *operengine.ActionInstContext
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx: &operengine.ActionInstContext{
					Ctx: context.Background(),
					Data: &operengine.ActionInstData{
						TriggerID:   "",
						OperInstID:  "",
						OperationID: "",
						Name:        "",
						Index:       0,
						Messages:    nil,
						Content: map[string]any{
							keys.CKeyToken: "123",
						},
						PrivateData: nil,
						Lifecycle:   nil,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &PushHostIdentifier{
				gseClient:  capability.GseHandler,
				cmdbClient: capability.CmdbHandler,
				storage:    capability.NodeDeploymentStorage,
				logger:     capability.Logger,
			}
			if err := action.Do(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Do() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
