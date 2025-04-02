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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/keys"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
)

// TestUpsertHost is a test suite for the UpsertHost action.
func (suite *TestSuite) TestUpsertHost() {
	type args struct {
		ctx *operengine.ActionInstContext
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "normal",
			args: args{
				ctx: &operengine.ActionInstContext{
					Ctx: context.Background(),
					Data: &operengine.ActionInstData{
						Content: map[string]interface{}{
							keys.CKeyToken: "123",
						},
					},
				},
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			action := NewActionUpsertHost(
				suite.capability.CmdbHandler,
				suite.capability.TopoStorage,
				suite.capability.NodeDeploymentStorage,
			)
			err := action.Do(tt.args.ctx)
			suite.Require().NoError(err, "failed to execute action")
			suite.T().Logf("action result: %v", tt.args.ctx.Data)
		})
	}
}
