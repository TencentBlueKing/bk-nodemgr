/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package plugindeployment

import (
	"context"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
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
						Plugin: types.Plugin{
							HostID:     0,
							Type:       "",
							Generation: 0,
							Platform: Platform.Platform{
								OS:   "",
								Arch: "",
							},
							Version: "",
						},
						InstallerWorkDir: "",
						InstallOptions:   types.PluginDeploymentInstallOptions{},
						UpgradeOptions:   types.PluginDeploymentUpgradeOptions{},
						RestartOptions:   types.PluginDeploymentRestartOptions{},
						TransferOptions: types.PluginDeploymentTransferOptions{
							SelectDownloads:      false,
							EnableReleasePackage: false,
							EnableInstaller:      false,
						},
						TargetVersion: []types.TargetPluginVersion{
							{
								Platform: Platform.Platform{
									OS:   criteria.OSLinux,
									Arch: criteria.CPUArchAmd64,
								},
								Version: "1.0.0",
							},
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
				Plugin: types.Plugin{
					HostID:     0,
					Type:       "",
					Generation: 0,
					Platform: Platform.Platform{
						OS:   "",
						Arch: "",
					},
					Version: "",
				},
				InstallerWorkDir: "",
				InstallOptions:   types.PluginDeploymentInstallOptions{},
				UpgradeOptions:   types.PluginDeploymentUpgradeOptions{},
				RestartOptions:   types.PluginDeploymentRestartOptions{},
				TransferOptions: types.PluginDeploymentTransferOptions{
					SelectDownloads:      false,
					EnableReleasePackage: false,
					EnableInstaller:      false,
				},
				TargetVersion: []types.TargetPluginVersion{
					{
						Platform: Platform.Platform{
							OS:   criteria.OSLinux,
							Arch: criteria.CPUArchAmd64,
						},
						Version: "1.0.0",
					},
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
					Plugin: types.Plugin{
						HostID:     0,
						Type:       "",
						Generation: 0,
						Platform: Platform.Platform{
							OS:   "",
							Arch: "",
						},
						Version: "",
					},
					InstallerWorkDir: "",
					InstallOptions:   types.PluginDeploymentInstallOptions{},
					UpgradeOptions:   types.PluginDeploymentUpgradeOptions{},
					RestartOptions:   types.PluginDeploymentRestartOptions{},
					TransferOptions: types.PluginDeploymentTransferOptions{
						SelectDownloads:      true,
						EnableReleasePackage: true,
						EnableInstaller:      true,
					},
					TargetVersion: []types.TargetPluginVersion{
						{
							Platform: Platform.Platform{
								OS:   criteria.OSLinux,
								Arch: criteria.CPUArchAmd64,
							},
							Version: "2.0.0",
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
			if err := h.UpdateInfo(tt.args.nCtx, tt.args.token, tt.args.info); (err != nil) != tt.wantErr {
				t.Errorf("UpdateNodeDeploymentInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
