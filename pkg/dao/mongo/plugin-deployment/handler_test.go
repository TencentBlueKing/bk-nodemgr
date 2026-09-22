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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
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

// TestPluginDeploymentProcessLastSyncAt verifies timestamps survive persistence between actions.
func TestPluginDeploymentProcessLastSyncAt(t *testing.T) {
	lastSyncAt := time.Date(2026, time.January, 2, 3, 4, 5, 123000000, time.UTC)
	info := &types.PluginDeploymentInfo{
		Process: types.Process{Info: types.ProcessInfo{LastSyncAt: lastSyncAt}},
	}

	stored, err := convertPluginDeploymentInfoFromTypes(info)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := bson.Marshal(stored)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Info
	if err := bson.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := convertPluginDeploymentInfoToTypes(&decoded)
	if err != nil {
		t.Fatal(err)
	}
	if !restored.Process.Info.LastSyncAt.Equal(lastSyncAt) {
		t.Fatalf("LastSyncAt = %s, want %s", restored.Process.Info.LastSyncAt, lastSyncAt)
	}
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

func TestConvertPluginDeploymentPluginConfRemoveAllConfigs(t *testing.T) {
	tests := []struct {
		name             string
		conf             *types.PluginDeploymentPluginConf
		wantRemoveAll    bool
		wantRemoveByName []string
	}{
		{
			name: "remove all configs",
			conf: &types.PluginDeploymentPluginConf{
				RemoveAllConfigs: true,
			},
			wantRemoveAll: true,
		},
		{
			name: "remove named configs",
			conf: &types.PluginDeploymentPluginConf{
				RemoveConfigFileName: []string{"config.yaml"},
			},
			wantRemoveByName: []string{"config.yaml"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := convertPluginDeploymentPluginConfFromTypes(tt.conf)
			if data.RemoveAllConfigs != tt.wantRemoveAll {
				t.Fatalf(
					"convertPluginDeploymentPluginConfFromTypes() RemoveAllConfigs = %t, want %t",
					data.RemoveAllConfigs,
					tt.wantRemoveAll,
				)
			}

			raw, err := bson.Marshal(data)
			if err != nil {
				t.Fatalf("bson.Marshal() error = %v", err)
			}
			value := bson.Raw(raw).Lookup("remove_all_configs")
			if value.Type != bson.TypeBoolean || value.Boolean() != tt.wantRemoveAll {
				t.Fatalf("BSON remove_all_configs = %v, want %t", value, tt.wantRemoveAll)
			}

			var persisted PluginConf
			if err := bson.Unmarshal(raw, &persisted); err != nil {
				t.Fatalf("bson.Unmarshal() error = %v", err)
			}

			got := convertPluginDeploymentPluginConfToTypes(&persisted)
			if got.RemoveAllConfigs != tt.wantRemoveAll {
				t.Fatalf(
					"convertPluginDeploymentPluginConfToTypes() RemoveAllConfigs = %t, want %t",
					got.RemoveAllConfigs,
					tt.wantRemoveAll,
				)
			}
			if !reflect.DeepEqual(got.RemoveConfigFileName, tt.wantRemoveByName) {
				t.Fatalf(
					"convertPluginDeploymentPluginConfToTypes() RemoveConfigFileName = %v, want %v",
					got.RemoveConfigFileName,
					tt.wantRemoveByName,
				)
			}
		})
	}
}

func TestConvertPluginDeploymentPluginConfLegacyBSONDefaultsRemoveAllConfigs(t *testing.T) {
	raw, err := bson.Marshal(bson.M{
		"remove_config_file_name": []string{"config.yaml"},
	})
	if err != nil {
		t.Fatalf("bson.Marshal() error = %v", err)
	}

	var persisted PluginConf
	if err := bson.Unmarshal(raw, &persisted); err != nil {
		t.Fatalf("bson.Unmarshal() error = %v", err)
	}

	got := convertPluginDeploymentPluginConfToTypes(&persisted)
	if got.RemoveAllConfigs {
		t.Fatal("convertPluginDeploymentPluginConfToTypes() RemoveAllConfigs = true, want false")
	}
	if !reflect.DeepEqual(got.RemoveConfigFileName, []string{"config.yaml"}) {
		t.Fatalf(
			"convertPluginDeploymentPluginConfToTypes() RemoveConfigFileName = %v, want [config.yaml]",
			got.RemoveConfigFileName,
		)
	}
}

func TestConvertPluginDeploymentInfoKeepsDebugCmd(t *testing.T) {
	info := &types.PluginDeploymentInfo{
		Process: types.Process{
			BindIP:   "127.0.0.1",
			BindPort: 10000,
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
	if data.Process.BindIP != "127.0.0.1" || data.Process.BindPort != 10000 {
		t.Fatalf("convertPluginDeploymentInfoFromTypes() bind address = %s:%d, want %s:%d",
			data.Process.BindIP, data.Process.BindPort, "127.0.0.1", 10000)
	}

	got, err := convertPluginDeploymentInfoToTypes(data)
	if err != nil {
		t.Fatalf("convertPluginDeploymentInfoToTypes() error = %v", err)
	}
	if got.Process.Controller.DebugCmd != "debug.sh" {
		t.Fatalf("convertPluginDeploymentInfoToTypes() DebugCmd = %q, want %q", got.Process.Controller.DebugCmd, "debug.sh")
	}
	if got.Process.BindIP != "127.0.0.1" || got.Process.BindPort != 10000 {
		t.Fatalf("convertPluginDeploymentInfoToTypes() bind address = %s:%d, want %s:%d",
			got.Process.BindIP, got.Process.BindPort, "127.0.0.1", 10000)
	}
}

func TestConvertPluginDeploymentInfoKeepsConfigSource(t *testing.T) {
	processCreateTime := time.Date(2026, 9, 3, 10, 0, 0, 0, time.UTC)
	processLastTime := time.Date(2026, 9, 3, 10, 1, 0, 0, time.UTC)
	info := &types.PluginDeploymentInfo{
		ConfigSource: types.Target{
			Host: types.Host{
				HostID:   1001,
				TenantID: "tenant-1",
				Static: &types.HostStatic{
					BizID:       2001,
					InnerIPList: []string{"127.0.0.1"},
				},
				Dynamic: &types.HostDynamic{
					LoginIP: "127.0.0.2",
				},
			},
			ServiceInstance: types.ServiceInstance{
				ID:     3001,
				Name:   "svc-1",
				Labels: map[string]string{"CMDB_LABEL_1": "something"},
				Processes: map[string]types.ServiceInstanceProcess{
					"process-1": {
						AutoStart:       true,
						BizID:           2001,
						FuncName:        "func-1",
						ProcessID:       7001,
						ProcessName:     "process-1",
						StartParamRegex: "--config config.yaml",
						SupplierAccount: "0",
						CreateTime:      processCreateTime,
						LastTime:        processLastTime,
						Description:     "process description",
						FaceStopCmd:     "kill",
						PidFile:         "/var/run/process.pid",
						Priority:        1,
						ProcNum:         2,
						ReloadCmd:       "reload",
						RestartCmd:      "restart",
						StartCmd:        "start",
						StopCmd:         "stop",
						Timeout:         3,
						User:            "root",
						WorkPath:        "/data/process",
						CreateAt:        "2026-09-03T10:00:00Z",
						CreateBy:        "creator",
						UpdateAt:        "2026-09-03T10:01:00Z",
						UpdateBy:        "updater",
						BindInfo: []types.ServiceInstanceProcessBindInfo{
							{Enable: true, IP: "127.0.0.1", Port: "8080", Protocol: "tcp", TemplateRowID: 8001},
						},
					},
				},
				BizID:             2001,
				HostID:            1001,
				ModuleID:          4001,
				ServiceTemplateID: 5001,
				ServiceCategoryID: 6001,
			},
			MatchedTopoRelations: []types.TargetMatchedTopoRelation{
				{TopoObjID: "module", TopoInstID: 4001},
			},
		},
	}

	data, err := convertPluginDeploymentInfoFromTypes(info)
	if err != nil {
		t.Fatalf("convertPluginDeploymentInfoFromTypes() error = %v", err)
	}
	if data.ConfigSource == nil {
		t.Fatal("convertPluginDeploymentInfoFromTypes() ConfigSource is nil")
	}
	if data.ConfigSource.Host.HostID != 1001 {
		t.Fatalf("convertPluginDeploymentInfoFromTypes() ConfigSource.Host.HostID = %d, want %d",
			data.ConfigSource.Host.HostID, 1001)
	}
	if data.ConfigSource.ServiceInstance.ID != 3001 {
		t.Fatalf("convertPluginDeploymentInfoFromTypes() ConfigSource.ServiceInstance.ID = %d, want %d",
			data.ConfigSource.ServiceInstance.ID, 3001)
	}
	if data.ConfigSource.ServiceInstance.Labels["CMDB_LABEL_1"] != "something" {
		t.Fatalf("convertPluginDeploymentInfoFromTypes() ConfigSource.ServiceInstance.Labels[CMDB_LABEL_1] = %q, want %q",
			data.ConfigSource.ServiceInstance.Labels["CMDB_LABEL_1"], "something")
	}
	if data.ConfigSource.ServiceInstance.Processes["process-1"].ProcessID != 7001 {
		t.Fatalf("convertPluginDeploymentInfoFromTypes() ConfigSource.ServiceInstance.Processes[process-1].ProcessID = %d, want %d",
			data.ConfigSource.ServiceInstance.Processes["process-1"].ProcessID, 7001)
	}
	if len(data.ConfigSource.MatchedTopoRelations) != 1 {
		t.Fatalf("convertPluginDeploymentInfoFromTypes() ConfigSource.MatchedTopoRelations length = %d, want %d",
			len(data.ConfigSource.MatchedTopoRelations), 1)
	}

	got, err := convertPluginDeploymentInfoToTypes(data)
	if err != nil {
		t.Fatalf("convertPluginDeploymentInfoToTypes() error = %v", err)
	}
	if got.ConfigSource.Host.HostID != 1001 {
		t.Fatalf("convertPluginDeploymentInfoToTypes() ConfigSource.Host.HostID = %d, want %d",
			got.ConfigSource.Host.HostID, 1001)
	}
	if got.ConfigSource.Host.Static.BizID != 2001 {
		t.Fatalf("convertPluginDeploymentInfoToTypes() ConfigSource.Host.Static.BizID = %d, want %d",
			got.ConfigSource.Host.Static.BizID, 2001)
	}
	if got.ConfigSource.ServiceInstance.ModuleID != 4001 {
		t.Fatalf("convertPluginDeploymentInfoToTypes() ConfigSource.ServiceInstance.ModuleID = %d, want %d",
			got.ConfigSource.ServiceInstance.ModuleID, 4001)
	}
	if got.ConfigSource.ServiceInstance.Labels["CMDB_LABEL_1"] != "something" {
		t.Fatalf("convertPluginDeploymentInfoToTypes() ConfigSource.ServiceInstance.Labels[CMDB_LABEL_1] = %q, want %q",
			got.ConfigSource.ServiceInstance.Labels["CMDB_LABEL_1"], "something")
	}
	process := got.ConfigSource.ServiceInstance.Processes["process-1"]
	if process.ProcessName != "process-1" {
		t.Fatalf("convertPluginDeploymentInfoToTypes() ConfigSource.ServiceInstance.Processes[process-1].ProcessName = %q, want %q",
			process.ProcessName, "process-1")
	}
	if process.BindInfo[0].Port != "8080" {
		t.Fatalf("convertPluginDeploymentInfoToTypes() ConfigSource.ServiceInstance.Processes[process-1].BindInfo[0].Port = %q, want %q",
			process.BindInfo[0].Port, "8080")
	}
	if len(got.ConfigSource.MatchedTopoRelations) != 1 {
		t.Fatalf("convertPluginDeploymentInfoToTypes() ConfigSource.MatchedTopoRelations length = %d, want %d",
			len(got.ConfigSource.MatchedTopoRelations), 1)
	}
	if got.ConfigSource.MatchedTopoRelations[0].TopoObjID != "module" {
		t.Fatalf("convertPluginDeploymentInfoToTypes() ConfigSource.MatchedTopoRelations[0].TopoObjID = %q, want %q",
			got.ConfigSource.MatchedTopoRelations[0].TopoObjID, "module")
	}
}

func TestConvertPluginDeploymentInfoOmitsEmptyConfigSource(t *testing.T) {
	data, err := convertPluginDeploymentInfoFromTypes(&types.PluginDeploymentInfo{})
	if err != nil {
		t.Fatalf("convertPluginDeploymentInfoFromTypes() error = %v", err)
	}
	if data.ConfigSource != nil {
		t.Fatalf("convertPluginDeploymentInfoFromTypes() ConfigSource = %+v, want nil", data.ConfigSource)
	}

	got, err := convertPluginDeploymentInfoToTypes(data)
	if err != nil {
		t.Fatalf("convertPluginDeploymentInfoToTypes() error = %v", err)
	}
	if got.ConfigSource.Host.HostID != 0 {
		t.Fatalf("convertPluginDeploymentInfoToTypes() ConfigSource.Host.HostID = %d, want %d",
			got.ConfigSource.Host.HostID, 0)
	}
}
