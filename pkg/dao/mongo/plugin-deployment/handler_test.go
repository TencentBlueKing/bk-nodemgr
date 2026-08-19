/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package plugindeployment

import (
	"context"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var instance struct {
	Token string
	sync.Once
}

func token() string {
	instance.Once.Do(func() {
		instance.Token = uuid.NewString()
	})

	return instance.Token
}

// testClient ...
func testClient(t *testing.T) IHandler {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	nCtx := context.Background()
	mongoClient, err := mongo.Connect(
		nCtx,
		&options.ClientOptions{
			Hosts: []string{
				os.Getenv("MONGO_ADDRESS"),
			},
			Auth: &options.Credential{
				Username:      os.Getenv("MONGO_USER"),
				Password:      os.Getenv("MONGO_PASSWORD"),
				AuthSource:    os.Getenv("MONGO_AUTH_SOURCE"),
				AuthMechanism: os.Getenv("MONGO_AUTH_MECHANISM"),
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	return New(mongoClient.Database(os.Getenv("MONGO_DATABASE")))
}

// TestHandler_Create test handler Create
func TestHandler_Create(t *testing.T) {
	type args struct {
		nCtx             contextx.IContext
		pluginDeployment *types.PluginDeployment
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal test",
			args: args{
				nCtx: contextx.New(context.Background()),
				pluginDeployment: &types.PluginDeployment{
					Token: token(),
					Info: &types.PluginDeploymentInfo{
						BlockingActionName: "wait_plugin_installer_complete",
						Process: types.Process{
							TenantID:   "",
							HostID:     0,
							PluginName: "",
							Info: types.ProcessInfo{
								Pid:       233,
								Version:   "1.0.0",
								AgentID:   "123456",
								AutoStart: false,
								Status:    "running",
							},
							Identity: types.ProcessIdentity{
								Name:       "proc-test",
								SetupPath:  "/usr/local/bin",
								PidPath:    "/var/run/proc-test.pid",
								ConfigPath: "/etc/proc-test/proc-test.conf",
								LogPath:    "/var/log/proc-test.log",
								User:       "root",
							},
							Controller: types.ProcessController{
								StartCmd:   "./start.sh",
								StopCmd:    "./stop.sh",
								RestartCmd: "./restart.sh",
								ReloadCmd:  "./reload.sh",
								KillCmd:    "./kill.sh",
								VersionCmd: "./version.sh",
								HealthCmd:  "./health.sh",
							},
							Resource: types.ProcessResource{
								CPULimitPercent: 10,
								MemLimitPercent: 10,
							},
							MonitorPolicy: types.ProcessMonitorPolicy{
								RestartType:    "trusteeship",
								StartCheckSecs: 5,
								StopCheckSecs:  5,
								OpTimeoutSecs:  5,
							},
						},
						InstallerRuntime: types.PluginDeploymentInstallerRuntime{
							BaseWorkDir: "/usr/local/bin",
						},
						InstallOptions: types.PluginDeploymentInstallOptions{
							Version: "1.0.0",
						},
						TransferOptions: types.PluginDeploymentTransferOptions{
							DisableReleasePackage: false,
							DisableInstaller:      false,
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			if err := h.Create(tt.args.nCtx, tt.args.pluginDeployment); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandler_GetInfo(t *testing.T) {
	type args struct {
		nCtx  contextx.IContext
		token string
	}
	tests := []struct {
		name    string
		args    args
		want    *types.PluginDeploymentInfo
		wantErr bool
	}{
		{
			name: "normal test",
			args: args{
				nCtx:  contextx.New(context.Background()),
				token: token(),
			},
			want: &types.PluginDeploymentInfo{
				BlockingActionName: "wait_plugin_installer_complete",
				Process:            types.Process{},
				InstallerRuntime: types.PluginDeploymentInstallerRuntime{
					BaseWorkDir: "",
				},
				InstallOptions: types.PluginDeploymentInstallOptions{},
				TransferOptions: types.PluginDeploymentTransferOptions{
					DisableReleasePackage: false,
					DisableInstaller:      false,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.GetInfo(tt.args.nCtx, tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNodeDeploymentInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNodeDeploymentInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestHandler_UpdateInfo ...
func TestHandler_UpdateInfo(t *testing.T) {
	type args struct {
		nCtx  contextx.IContext
		token string
		info  *types.PluginDeploymentInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal test",
			args: args{
				nCtx:  contextx.New(context.Background()),
				token: token(),
				info: &types.PluginDeploymentInfo{
					BlockingActionName: "wait_plugin_installer_complete",
					Process: types.Process{
						TenantID:   "",
						HostID:     0,
						PluginName: "",
						Info: types.ProcessInfo{
							Pid:       233,
							Version:   "1.0.0",
							AgentID:   "123456",
							AutoStart: false,
							Status:    "running",
						},
						Identity: types.ProcessIdentity{
							Name:       "proc-test",
							SetupPath:  "/usr/local/bin",
							PidPath:    "/var/run/proc-test.pid",
							ConfigPath: "/etc/proc-test/proc-test.conf",
							LogPath:    "/var/log/proc-test.log",
							User:       "root",
						},
						Controller: types.ProcessController{
							StartCmd:   "./start.sh",
							StopCmd:    "./stop.sh",
							RestartCmd: "./restart.sh",
							ReloadCmd:  "./reload.sh",
							KillCmd:    "./kill.sh",
							VersionCmd: "./version.sh",
							HealthCmd:  "./health.sh",
						},
						Resource: types.ProcessResource{
							CPULimitPercent: 10,
							MemLimitPercent: 10,
						},
						MonitorPolicy: types.ProcessMonitorPolicy{
							RestartType:    "trusteeship",
							StartCheckSecs: 5,
							StopCheckSecs:  5,
							OpTimeoutSecs:  5,
						},
					},
					InstallerRuntime: types.PluginDeploymentInstallerRuntime{
						BaseWorkDir: "/usr/local/bin",
					},
					InstallOptions: types.PluginDeploymentInstallOptions{
						Version: "1.0.0",
					},
					TransferOptions: types.PluginDeploymentTransferOptions{
						DisableReleasePackage: false,
						DisableInstaller:      false,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			if err := h.UpdateInfo(tt.args.nCtx, tt.args.token, tt.args.info); (err != nil) != tt.wantErr {
				t.Errorf("UpdateNodeDeploymentInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConvertPluginConfigDetailsTemplateName(t *testing.T) {
	configDetails := convertPluginConfigDetailsFromTypes(&types.PluginConfigDetail{
		Name:         "test-config-name",
		TemplateName: "test-template-name",
		Content:      "test-content",
		IsMainConfig: true,
		FilePath:     "etc/test.conf",
	})
	if len(configDetails) != 1 {
		t.Fatalf("convertPluginConfigDetailsFromTypes() length = %d, want %d", len(configDetails), 1)
	}
	if configDetails[0].TemplateName != "test-template-name" {
		t.Fatalf("convertPluginConfigDetailsFromTypes() TemplateName = %q, want %q", configDetails[0].TemplateName, "test-template-name")
	}

	got := convertPluginConfigDetailsToTypes(configDetails...)
	if len(got) != 1 {
		t.Fatalf("convertPluginConfigDetailsToTypes() length = %d, want %d", len(got), 1)
	}
	if got[0].TemplateName != "test-template-name" {
		t.Fatalf("convertPluginConfigDetailsToTypes() TemplateName = %q, want %q", got[0].TemplateName, "test-template-name")
	}
}

func TestConvertPluginDeploymentInfoKeepsDebugCmd(t *testing.T) {
	info := &types.PluginDeploymentInfo{
		Process: types.Process{
			Controller: types.ProcessController{
				DebugCmd: "debug.sh",
			},
		},
	}

	data, err := convertPluginDeploymentInfoFromTypes(info)
	if err != nil {
		t.Fatalf("convertPluginDeploymentInfoFromTypes() error = %v", err)
	}
	if data.Process.Controller.DebugCmd != "debug.sh" {
		t.Fatalf("convertPluginDeploymentInfoFromTypes() DebugCmd = %q, want %q", data.Process.Controller.DebugCmd, "debug.sh")
	}

	got, err := convertPluginDeploymentInfoToTypes(data)
	if err != nil {
		t.Fatalf("convertPluginDeploymentInfoToTypes() error = %v", err)
	}
	if got.Process.Controller.DebugCmd != "debug.sh" {
		t.Fatalf("convertPluginDeploymentInfoToTypes() DebugCmd = %q, want %q", got.Process.Controller.DebugCmd, "debug.sh")
	}
}
