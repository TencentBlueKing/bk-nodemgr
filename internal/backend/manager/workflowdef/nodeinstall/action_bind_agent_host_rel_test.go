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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

// NewActionBindAgentHostRel ...
func TestBindAgentHostRel_Do(t *testing.T) {
	capability := testCapability(t)

	type args struct {
		ctx *action.InstanceContext
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx: &action.InstanceContext{
					Ctx: context.Background(),
					Data: &action.InstanceData{
						TriggerID:           "",
						OperationInstanceID: "",
						OperationID:         "",
						Name:                "",
						Index:               0,
						Messages:            nil,
						Content: map[string]any{
							keys.CKeyToken: "666",
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
			action := NewActionBindAgentHostRel(
				capability.CmdbHandler,
				capability.StorageTopo,
				capability.NodeDeploymentStorage,
				capability.Logger)
			if err := action.Do(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Do() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
