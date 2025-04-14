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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
	"github.com/TencentBlueKing/bk-nodemgr/test/basetest"
)

// TestRenderNodeInstallConfig is a test suite for the RenderNodeDeployment action
type TestSuite struct {
	basetest.TestSuit
	capability *Capability
}

// testCapability initializes the capability for the test suite
func TestAll(t *testing.T) {
	suit := new(TestSuite)
	suit.AddSetupSuiteFunc(func() {
		suit.capability = testCapability(t)

		err := deployconstant.SetDeployConf(deployconstant.DeployConf{
			Generation:    2,
			OsType:        criteria.OSLinux,
			HostIDPath:    "/var/lib/gse/host",
			GseDataIPC:    "/usr/local/gse2/agent/data/ipc.state.report",
			GsePluginIPC:  "/usr/local/gse2/agent/lib/ipc.state.message",
			GseHomeDir:    "/usr/local/gse2",
			GseDataDir:    "/var/lib/gse2",
			GseRunDir:     "/var/run/gse2",
			GseLogDir:     "/var/log/gse2",
			GseEnvironDir: "/etc/sysconfig/gse/gse2",
		})
		suit.Require().NoError(err, "failed to set deploy conf")
	})

	basetest.RunTests(t, suit)
}

func (suite *TestSuite) TestRenderNodeInstallConfig_Do() {
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
		suite.Run(tt.name, func() {
			action := NewActionRenderNodeDeployment(
				suite.capability.NodeDeploymentStorage,
				suite.capability.TopoStorage,
				suite.capability.TopoStorage,
				suite.capability.Logger,
			)
			err := action.Do(tt.args.ctx)
			suite.Require().NoError(err, "failed to execute action")

			suite.T().Logf("action result: %v", tt.args.ctx.Data.Content)
		})
	}
}
