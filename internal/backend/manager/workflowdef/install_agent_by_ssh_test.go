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
	"os"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/local"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
	"github.com/joho/godotenv"
)

func testToolGroup(t *testing.T) iface.FileGroup {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	toolGroup, err := local.NewLocalDir(os.Getenv("TOOL_GROUP_DIR"), logger.LoggerDefault{})
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	return toolGroup
}

// TestInstallAgentBySSH_Do ...
func TestInstallAgentBySSH_Do(t *testing.T) {
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
			name: "test",
			args: args{
				ctx: &operengine.ActionInstContext{
					Ctx: context.Background(),
					Data: &operengine.ActionInstData{
						TriggerID:   "",
						OperInstID:  "asdasd",
						OperationID: "",
						Name:        "install_agent_by_ssh",
						Index:       0,
						Messages:    nil,
						Content: map[string]any{
							"ip":        os.Getenv("SSH_IP"),
							"port":      36000,
							"node_type": "agent",
							"user":      os.Getenv("SSH_USER"),
							"passwd": func() []byte {
								ciphertext, err := capability.Crypter.Encrypt([]byte(os.Getenv("SSH_PASSWD")))
								if err != nil {
									t.Fatal(err)
								}

								return ciphertext
							}(),
							"callback_endpoint": os.Getenv("CALLBACK_ENDPOINT"),
							"download_endpoint": os.Getenv("DOWNLOAD_ENDPOINT"),
							"pkg_version":       "v1",
							"pkg_generation":    1,
							"install_env":       "gse2_opbk",
							"token":             "111",
							"addition_args":     []string{"--reinstall", "--debug"},
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
			action := NewActionInstallAgentBySSH(testToolGroup(t), capability.Crypter, capability.Logger)
			err := action.Do(tt.args.ctx)
			if err != nil {
				t.Logf("err: %v", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("Do() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("data: %+v", tt.args.ctx.Data)
		})
	}
}
