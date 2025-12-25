/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package gse

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	restdiscovery "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/discovery"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	apigwclient "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/apigw/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/joho/godotenv"
)

// LoadAuthHeader load auth header from environment variables.
func LoadAuthHeader() (apigwclient.UserConfig, error) {
	apigwAuthHeader := os.Getenv("BK_APIGW_AUTHHEADER")
	header := make(map[string]string)
	if err := json.Unmarshal([]byte(apigwAuthHeader), &header); err != nil {
		return apigwclient.UserConfig{}, err
	}

	apigwClientConfig := apigwclient.UserConfig{
		AppConfig: apigwclient.NewAppConfig(
			[]string{os.Getenv("BK_APIGW_ENDPOINT")},
			header["bk_app_code"],
			header["bk_app_secret"]),
		BKUsername: header["bk_username"],
		AuthMode:   apigwclient.AuthModeUn,
	}

	return apigwClientConfig, nil
}

type testContext struct {
	tasks map[string]string

	listAgents          []string
	scriptAgents        []string
	fileSourceAgent     string
	fileTargetAgentBase string
	fileTargetAgentTerm string
}

// nolint:gochecknoglobals
var (
	once sync.Once
	tc   *testContext
)

// getGlobalContext maintains a global context for testing.
// it will load agent-ids from environment variables and manage the task-id from gse API.
func getGlobalContext(t *testing.T) *testContext {
	once.Do(func() {
		err := godotenv.Load(".env")
		if err != nil {
			t.Fatal(err)
		}

		tc = &testContext{
			tasks: make(map[string]string),
		}

		tc.listAgents = strings.Split(os.Getenv("GSE_LIST_AGENTS"), ",")
		tc.scriptAgents = strings.Split(os.Getenv("GSE_SCRIPTS_AGENTS"), ",")
		tc.fileSourceAgent = os.Getenv("GSE_FILE_SOURCE_AGENT")
		tc.fileTargetAgentBase = os.Getenv("GSE_FILE_TARGET_AGENT_BASE")
		tc.fileTargetAgentTerm = os.Getenv("GSE_FILE_TARGET_AGENT_TERM")
	})

	return tc
}

func testClient(t *testing.T) IHandler {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	httpClient, err := restclient.NewHTTPClient(&ssl.TLSConfig{
		InsecureSkipVerify: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	clientCap := &restclient.Capability{
		HTTPClient:           httpClient,
		Discover:             restdiscovery.NewDiscovery("apigateway", []string{os.Getenv("BK_APIGW_ENDPOINT")}),
		ToleranceLatencyTime: restclient.ToleranceLatencyTimeDefault,
		MetricOpts:           restclient.MetricOption{},
	}

	apigwClientConfig, err := LoadAuthHeader()
	if err != nil {
		t.Fatal(err)
	}

	h, err := New(clientCap, &Config{
		APIGWUserConfig: apigwClientConfig,
	})
	if err != nil {
		t.Fatal(err)
	}

	return h
}

// Test_Handler_ListAgentInfo tests Handler.ListAgentInfo.
func Test_Handler_ListAgentInfo(t *testing.T) {
	nCtx := contextx.New(context.Background())

	type args struct {
		agentIDList []string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "base",
			args: args{
				agentIDList: getGlobalContext(t).listAgents,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.ListAgentInfo(nCtx, tt.args.agentIDList...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListAgentInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != len(tt.args.agentIDList) {
				t.Errorf("ListAgentInfo() result size = %v, want %v", len(got), len(tt.args.agentIDList))
				return
			}

			for _, info := range got {
				t.Logf("agent-info: %#v", info)
			}
		})
	}
}

// Test_Handler_ListAgentState tests Handler.ListAgentState.
func Test_Handler_ListAgentState(t *testing.T) {
	nCtx := contextx.New(context.Background())

	type args struct {
		agentIDList []string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "base",
			args: args{
				agentIDList: getGlobalContext(t).listAgents,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.ListAgentState(nCtx, tt.args.agentIDList...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListAgentState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != len(tt.args.agentIDList) {
				t.Errorf("ListAgentState() result size = %v, want %v", len(got), len(tt.args.agentIDList))
				return
			}

			for _, state := range got {
				t.Logf("agent-state: %#v", state)
			}
		})
	}
}

// Test_Handler_ExecuteScript tests Handler.ExecuteScript.
func Test_Handler_ExecuteScript(t *testing.T) {
	nCtx := contextx.New(context.Background())

	type args struct {
		endpoints     []*types.EndpointWithAuth
		scriptContent string
		timeout       time.Duration
	}

	var endpoints []*types.EndpointWithAuth
	for _, agentID := range getGlobalContext(t).scriptAgents {
		endpoints = append(endpoints, &types.EndpointWithAuth{
			Endpoint: types.Endpoint{
				AgentID: agentID,
			},
			User: "root",
		})
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "script_base",
			args: args{
				endpoints:     endpoints,
				scriptContent: "truncate -s 10K /tmp/nodeman_gsetest",
				timeout:       10 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "script_wait_for_termination",
			args: args{
				endpoints:     endpoints,
				scriptContent: "sleep 100",
				timeout:       120 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			taskID, err := h.ExecuteScript(
				nCtx, types.ScriptTypeBash, tt.args.scriptContent, tt.args.timeout, tt.args.endpoints...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("execute-script: %v", taskID)

			getGlobalContext(t).tasks[tt.name] = taskID
		})
	}
}

// Test_Handler_TerminateScriptExecution tests Handler.TerminateScriptExecution.
func Test_Handler_TerminateScriptExecution(t *testing.T) {
	nCtx := contextx.New(context.Background())

	type args struct {
		taskID    string
		endpoints []*types.Endpoint
	}

	var endpoints []*types.Endpoint
	for _, agentID := range getGlobalContext(t).scriptAgents {
		endpoints = append(endpoints, &types.Endpoint{
			AgentID: agentID,
		})
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "script_wait_for_termination",
			args: args{
				taskID:    getGlobalContext(t).tasks["script_wait_for_termination"],
				endpoints: endpoints,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			termTaskID, err := h.TerminateScriptExecution(nCtx, tt.args.taskID, tt.args.endpoints...)
			if (err != nil) != tt.wantErr {
				t.Errorf("TerminateScriptExecution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("terminate-script: %v", termTaskID)
		})
	}
}

// Test_Handler_QueryScriptExecutionResult tests Handler.QueryScriptExecutionResult.
func Test_Handler_QueryScriptExecutionResult(t *testing.T) {
	nCtx := contextx.New(context.Background())

	type args struct {
		taskID    string
		endpoints []*types.EndpointWithRestrict
	}
	var endpoints []*types.EndpointWithRestrict
	for _, agentID := range getGlobalContext(t).scriptAgents {
		endpoints = append(endpoints, &types.EndpointWithRestrict{
			Endpoint: types.Endpoint{
				AgentID: agentID,
			},
			Offset: 0,
			Limit:  0,
		})
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantStatus types.ScriptStatus
	}{
		{
			name: "script_base",
			args: args{
				taskID:    getGlobalContext(t).tasks["script_base"],
				endpoints: endpoints,
			},
			wantErr:    false,
			wantStatus: types.ScriptStatusFinished,
		},
		{
			name: "script_wait_for_termination",
			args: args{
				taskID:    getGlobalContext(t).tasks["script_wait_for_termination"],
				endpoints: endpoints,
			},
			wantErr:    false,
			wantStatus: types.ScriptStatusStopped,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)

			var resp []*types.ScriptResult
			var err error

			for attempt := 0; attempt < 5; attempt++ {
				resp, err = h.QueryScriptExecutionResult(nCtx, tt.args.taskID, tt.args.endpoints...)
				if (err != nil) != tt.wantErr {
					t.Errorf("QueryScriptExecutionResult() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				stillRunning := false
				for _, v := range resp {
					if v.Status == types.ScriptStatusRunning {
						stillRunning = true
						break
					}
				}

				if !stillRunning && len(resp) == len(tt.args.endpoints) {
					break
				}

				time.Sleep(1 * time.Second)
			}

			if len(resp) != len(tt.args.endpoints) {
				t.Errorf("QueryScriptExecutionResult() result size = %v, wantLen %v", len(resp), len(tt.args.endpoints))
				return
			}

			for _, v := range resp {
				if v.Status != tt.wantStatus {
					t.Errorf("QueryScriptExecutionResult() status = %v, wantStatus %v", v.Status, tt.wantStatus)
					return
				}

				t.Logf("query-script: %#v", v)
			}
		})
	}
}

// Test_Handler_TransferFile tests Handler.TransferFile.
func Test_Handler_TransferFile(t *testing.T) {
	nCtx := contextx.New(context.Background())

	type args struct {
		options *types.TransferOptions
		details []*types.TransferDetail
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "file_base",
			args: args{
				options: &types.TransferOptions{
					Timeout:   20 * time.Second,
					AutoMkdir: true,
				},
				details: []*types.TransferDetail{{
					Source: types.TransferSource{
						Endpoint: types.EndpointWithAuth{
							Endpoint: types.Endpoint{
								AgentID: getGlobalContext(t).fileSourceAgent,
							},
							User: "root",
						},
						StoredDir: "/tmp",
						FileName:  "nodeman_gsetest",
					},
					Target: types.TransferTarget{
						Endpoints: []*types.EndpointWithAuth{
							{
								Endpoint: types.Endpoint{
									AgentID: getGlobalContext(t).fileTargetAgentBase,
								},
								User: "root",
							},
						},
						StoredDir: "/tmp",
					},
				}},
			},
			wantErr: false,
		},
		{
			name: "file_wait_for_termination",
			args: args{
				options: &types.TransferOptions{
					Timeout:             20 * time.Second,
					AutoMkdir:           true,
					UploadSpeedMBPerSec: 1,
				},
				details: []*types.TransferDetail{{
					Source: types.TransferSource{
						Endpoint: types.EndpointWithAuth{
							Endpoint: types.Endpoint{
								AgentID: getGlobalContext(t).fileSourceAgent,
							},
							User: "root",
						},
						StoredDir: "/tmp",
						FileName:  "nodeman_gsetest",
					},
					Target: types.TransferTarget{
						Endpoints: []*types.EndpointWithAuth{
							{
								Endpoint: types.Endpoint{
									AgentID: getGlobalContext(t).fileTargetAgentTerm,
								},
								User: "root",
							},
						},
						StoredDir: "/tmp",
					},
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			taskID, err := h.TransferFile(nCtx, tt.args.options, tt.args.details...)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransferFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("transfer-file: %v", taskID)

			getGlobalContext(t).tasks[tt.name] = taskID
		})
	}
}

// Test_Handler_TerminateFileTransmission tests Handler.TerminateFileTransmission.
func Test_Handler_TerminateFileTransmission(t *testing.T) {
	nCtx := contextx.New(context.Background())

	type args struct {
		taskID    string
		endpoints []*types.Endpoint
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "file_wait_for_termination",
			args: args{
				taskID: getGlobalContext(t).tasks["file_wait_for_termination"],
				endpoints: []*types.Endpoint{
					{AgentID: getGlobalContext(t).fileSourceAgent},
					{AgentID: getGlobalContext(t).fileTargetAgentTerm},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			_, err := h.TerminateFileTransmission(nCtx, tt.args.taskID, tt.args.endpoints...)
			if (err != nil) != tt.wantErr {
				t.Errorf("TerminateFileTransmission() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("terminate-file: %v", tt.args.taskID)
		})
	}
}

// Test_Handler_QueryFileTransmissionResult tests Handler.QueryFileTransmissionResult.
func Test_Handler_QueryFileTransmissionResult(t *testing.T) {
	nCtx := contextx.New(context.Background())

	type args struct {
		taskID    string
		endpoints []*types.Endpoint
	}
	tests := []struct {
		name         string
		args         args
		wantErr      bool
		wantStatuses []types.TransferStatus
	}{
		{
			name: "file_base",
			args: args{
				taskID: getGlobalContext(t).tasks["file_base"],
				endpoints: []*types.Endpoint{
					{AgentID: getGlobalContext(t).fileSourceAgent},
					{AgentID: getGlobalContext(t).fileTargetAgentBase},
				},
			},
			wantErr: false,
			wantStatuses: []types.TransferStatus{
				types.TransferStatusEndDownloading,
				types.TransferStatusEndUploading,
			},
		},
		{
			name: "file_wait_for_termination",
			args: args{
				taskID: getGlobalContext(t).tasks["file_wait_for_termination"],
				endpoints: []*types.Endpoint{
					{AgentID: getGlobalContext(t).fileSourceAgent},
					{AgentID: getGlobalContext(t).fileTargetAgentTerm},
				},
			},
			wantErr: false,
			wantStatuses: []types.TransferStatus{
				types.TransferStatusStopped,
				types.TransferStatusEndDownloading,
				types.TransferStatusEndUploading,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			var resp []*types.TransferResult
			var err error

			for attempt := 0; attempt < 20; attempt++ {
				resp, err = h.QueryFileTransmissionResult(nCtx, tt.args.taskID, tt.args.endpoints...)
				if (err != nil) != tt.wantErr {
					t.Errorf("QueryFileTransmissionResult() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				stillRunning := false
				for _, v := range resp {
					if v.StatusCode == types.TransferStatusChecking || v.StatusCode == types.TransferStatusRunning {
						stillRunning = true
						break
					}
				}

				if !stillRunning && len(resp) == len(tt.args.endpoints) {
					break
				}

				time.Sleep(1 * time.Second)
			}

			if len(resp) != len(tt.args.endpoints) {
				t.Errorf("QueryFileTransmissionResult() result size = %v, wantLen %v", len(resp), len(tt.args.endpoints))
				return
			}

			for _, v := range resp {
				statusMatched := false
				for _, status := range tt.wantStatuses {
					if v.StatusCode == status {
						statusMatched = true
						break
					}
				}

				if !statusMatched {
					t.Errorf("QueryFileTransmissionResult() status = %v, wantStatus %v", v.StatusCode, tt.wantStatuses)
					return
				}

				t.Logf("query-file: %#v", v)
			}
		})
	}
}

// Test_Handler_PushFile tests Handler.PushFile.
func Test_Handler_PushFile(t *testing.T) {
	nCtx := contextx.New(context.Background())

	type args struct {
		details []*types.PushFileDetail
	}
	tests := []struct {
		name         string
		args         args
		wantStatuses []types.TransferStatus
		wantErr      bool
	}{
		{
			name: "file_base",
			args: args{
				details: []*types.PushFileDetail{
					{
						FileName:    "nodeman_gsetest",
						StoreDir:    "/tmp",
						FileContent: "This is a test file pushed by bk-nodemgr GSE client.\n",
						Owner:       "root",
						Endpoints: []*types.Endpoint{
							{
								AgentID: getGlobalContext(t).fileTargetAgentBase,
							},
						},
					},
				},
			},
			wantStatuses: []types.TransferStatus{
				types.TransferStatusStopped,
				types.TransferStatusEndDownloading,
				types.TransferStatusEndUploading,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			taskID, err := h.PushFile(nCtx, tt.args.details...)
			if (err != nil) != tt.wantErr {
				t.Errorf("PushFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var resp []*types.TransferResult
			for attempt := 0; attempt < 20; attempt++ {
				resp, err = h.QueryFileTransmissionResult(nCtx, taskID, tt.args.details[0].Endpoints...)
				if (err != nil) != tt.wantErr {
					t.Errorf("QueryFileTransmissionResult() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				stillRunning := false
				for _, v := range resp {
					if v.StatusCode == types.TransferStatusChecking || v.StatusCode == types.TransferStatusRunning {
						stillRunning = true
						break
					}
				}

				if !stillRunning && len(resp) == len(tt.args.details[0].Endpoints) {
					break
				}

				time.Sleep(1 * time.Second)
			}

			if len(resp) != len(tt.args.details[0].Endpoints) {
				t.Errorf("QueryFileTransmissionResult() result size = %v, wantLen %v", len(resp), len(tt.args.details[0].Endpoints))
				return
			}

			for _, v := range resp {
				statusMatched := false
				for _, status := range tt.wantStatuses {
					if v.StatusCode == status {
						statusMatched = true
						break
					}
				}

				if !statusMatched {
					t.Errorf("QueryFileTransmissionResult() status = %v, wantStatus %v", v.StatusCode, tt.wantStatuses)
					return
				}

				t.Logf("query-file: %#v", v)
			}

			t.Logf("push-file: %v", taskID)
		})
	}
}

// Test_Handler_ReloadProcess test Handler.ReloadProcess.
func Test_Handler_ReloadProcess(t *testing.T) {
	nCtx := contextx.New(context.Background())

	type args struct {
		nCtx        contextx.IContext
		processSpec types.ProcessSpec
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				nCtx: nCtx,
				processSpec: types.ProcessSpec{
					AgentID: "020000000052540025157e1753363692251c",
					Identity: types.ProcessIdentity{
						Name:       "bkmonitorbeat",
						SetupPath:  "/usr/local/gse2/plugins/bin",
						PidPath:    "/var/run/gse2/bkmonitorbeat.pid",
						ConfigPath: "/usr/local/gse2/plugins/etc/bkmonitorbeat.conf",
						LogPath:    "/var/log/gse2",
						User:       "root",
					},
					Controller: types.ProcessController{
						StartCmd:   "./start.sh bkmonitorbeat",
						StopCmd:    "./stop.sh bkmonitorbeat",
						RestartCmd: "./restart.sh bkmonitorbeat",
						ReloadCmd:  "./reload.sh bkmonitorbeat",
						KillCmd:    "",
						VersionCmd: "./bkmonitorbeat -v",
						HealthCmd:  "",
					},
					Resource: types.ProcessResource{
						CPULimitPercent: 10,
						MemLimitPercent: 10,
					},
					MonitorPolicy: types.ProcessMonitorPolicy{
						AutoType:       types.ProcessAutoTypeTrusteeship,
						StartCheckSecs: 5,
						StopCheckSecs:  0,
						OpTimeoutSecs:  5,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.TrusteeshipAndReloadProcess(tt.args.nCtx, tt.args.processSpec)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReloadProcess() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("got: %v", got)
		})
	}
}

// Test_Handler_RestartProcess test Handler.RestartProcess.
func Test_Handler_RestartProcess(t *testing.T) {
	nCtx := contextx.New(context.Background())

	type args struct {
		nCtx        contextx.IContext
		processSpec types.ProcessSpec
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				nCtx: nCtx,
				processSpec: types.ProcessSpec{
					AgentID: "020000000052540025157e1753363692251c",
					Identity: types.ProcessIdentity{
						Name:       "bkmonitorbeat",
						SetupPath:  "/usr/local/gse2/plugins/bin",
						PidPath:    "/var/run/gse2/bkmonitorbeat.pid",
						ConfigPath: "/usr/local/gse2/plugins/etc/bkmonitorbeat.conf",
						LogPath:    "/var/log/gse2",
						User:       "root",
					},
					Controller: types.ProcessController{
						StartCmd:   "./start.sh bkmonitorbeat",
						StopCmd:    "./stop.sh bkmonitorbeat",
						RestartCmd: "./restart.sh bkmonitorbeat",
						ReloadCmd:  "./reload.sh bkmonitorbeat",
						KillCmd:    "",
						VersionCmd: "./bkmonitorbeat -v",
						HealthCmd:  "",
					},
					Resource: types.ProcessResource{
						CPULimitPercent: 10,
						MemLimitPercent: 10,
					},
					MonitorPolicy: types.ProcessMonitorPolicy{
						AutoType:       types.ProcessAutoTypeTrusteeship,
						StartCheckSecs: 5,
						StopCheckSecs:  0,
						OpTimeoutSecs:  5,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.TrusteeshipAndRestartProcess(tt.args.nCtx, tt.args.processSpec)
			if (err != nil) != tt.wantErr {
				t.Errorf("RestartProcess() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("got: %v", got)
		})
	}
}

// Test_Handler_StartProcess test Handler.StartProcess.
func Test_Handler_StartProcess(t *testing.T) {
	nCtx := contextx.New(context.Background())

	type args struct {
		nCtx        contextx.IContext
		processSpec types.ProcessSpec
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				nCtx: nCtx,
				processSpec: types.ProcessSpec{
					AgentID: "020000000052540025157e1753363692251c",
					Identity: types.ProcessIdentity{
						Name:       "bkmonitorbeat",
						SetupPath:  "/usr/local/gse2/plugins/bin",
						PidPath:    "/var/run/gse2/bkmonitorbeat.pid",
						ConfigPath: "/usr/local/gse2/plugins/etc/bkmonitorbeat.conf",
						LogPath:    "/var/log/gse2",
						User:       "root",
					},
					Controller: types.ProcessController{
						StartCmd:   "./start.sh bkmonitorbeat",
						StopCmd:    "./stop.sh bkmonitorbeat",
						RestartCmd: "./restart.sh bkmonitorbeat",
						ReloadCmd:  "./reload.sh bkmonitorbeat",
						KillCmd:    "",
						VersionCmd: "./bkmonitorbeat -v",
						HealthCmd:  "",
					},
					Resource: types.ProcessResource{
						CPULimitPercent: 10,
						MemLimitPercent: 10,
					},
					MonitorPolicy: types.ProcessMonitorPolicy{
						AutoType:       types.ProcessAutoTypeTrusteeship,
						StartCheckSecs: 5,
						StopCheckSecs:  0,
						OpTimeoutSecs:  5,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.TrusteeshipAndStartProcess(tt.args.nCtx, tt.args.processSpec)
			if (err != nil) != tt.wantErr {
				t.Errorf("StartProcess() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("got: %v", got)
		})
	}
}

// Test_Handler_StopProcess test Handler.StopProcess.
func Test_Handler_StopProcess(t *testing.T) {
	nCtx := contextx.New(context.Background())

	type args struct {
		nCtx        contextx.IContext
		processSpec types.ProcessSpec
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				nCtx: nCtx,
				processSpec: types.ProcessSpec{
					AgentID: "020000000052540025157e1753363692251c",
					Identity: types.ProcessIdentity{
						Name:       "bkmonitorbeat",
						SetupPath:  "/usr/local/gse2/plugins/bin",
						PidPath:    "/var/run/gse2/bkmonitorbeat.pid",
						ConfigPath: "/usr/local/gse2/plugins/etc/bkmonitorbeat.conf",
						LogPath:    "/var/log/gse2",
						User:       "root",
					},
					Controller: types.ProcessController{
						StartCmd:   "./start.sh bkmonitorbeat",
						StopCmd:    "./stop.sh bkmonitorbeat",
						RestartCmd: "./restart.sh bkmonitorbeat",
						ReloadCmd:  "./reload.sh bkmonitorbeat",
						KillCmd:    "",
						VersionCmd: "./bkmonitorbeat -v",
						HealthCmd:  "",
					},
					Resource: types.ProcessResource{
						CPULimitPercent: 10,
						MemLimitPercent: 10,
					},
					MonitorPolicy: types.ProcessMonitorPolicy{
						AutoType:       types.ProcessAutoTypeTrusteeship,
						StartCheckSecs: 5,
						StopCheckSecs:  0,
						OpTimeoutSecs:  5,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.UnTrusteeshipAndStopProcess(tt.args.nCtx, tt.args.processSpec)
			if (err != nil) != tt.wantErr {
				t.Errorf("StopProcess() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("got: %v", got)
		})
	}
}

// Test_Handler_TrusteeshipProcess test Handler.TrusteeshipProcess.
func Test_Handler_TrusteeshipProcess(t *testing.T) {
	nCtx := contextx.New(context.Background())

	type args struct {
		nCtx        contextx.IContext
		processSpec types.ProcessSpec
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				nCtx: nCtx,
				processSpec: types.ProcessSpec{
					AgentID: "020000000052540025157e1753363692251c",
					Identity: types.ProcessIdentity{
						Name:       "bkmonitorbeat",
						SetupPath:  "/usr/local/gse2/plugins/bin",
						PidPath:    "/var/run/gse2/bkmonitorbeat.pid",
						ConfigPath: "/usr/local/gse2/plugins/etc/bkmonitorbeat.conf",
						LogPath:    "/var/log/gse2",
						User:       "root",
					},
					Controller: types.ProcessController{
						StartCmd:   "./start.sh bkmonitorbeat",
						StopCmd:    "./stop.sh bkmonitorbeat",
						RestartCmd: "./restart.sh bkmonitorbeat",
						ReloadCmd:  "./reload.sh bkmonitorbeat",
						KillCmd:    "",
						VersionCmd: "./bkmonitorbeat -v",
						HealthCmd:  "",
					},
					Resource: types.ProcessResource{
						CPULimitPercent: 10,
						MemLimitPercent: 10,
					},
					MonitorPolicy: types.ProcessMonitorPolicy{
						AutoType:       types.ProcessAutoTypeTrusteeship,
						StartCheckSecs: 5,
						StopCheckSecs:  0,
						OpTimeoutSecs:  5,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.TrusteeshipProcess(tt.args.nCtx, tt.args.processSpec)
			if (err != nil) != tt.wantErr {
				t.Errorf("TrusteeshipProcess() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("got: %v", got)
		})
	}
}

// Test_Handler_UnTrusteeshipProcess test Handler.UnTrusteeshipProcess.
func Test_Handler_UnTrusteeshipProcess(t *testing.T) {
	nCtx := contextx.New(context.Background())

	type args struct {
		nCtx        contextx.IContext
		processSpec types.ProcessSpec
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				nCtx: nCtx,
				processSpec: types.ProcessSpec{
					AgentID: "020000000052540025157e1753363692251c",
					Identity: types.ProcessIdentity{
						Name:       "bkmonitorbeat",
						SetupPath:  "/usr/local/gse2/plugins/bin",
						PidPath:    "/var/run/gse2/bkmonitorbeat.pid",
						ConfigPath: "/usr/local/gse2/plugins/etc/bkmonitorbeat.conf",
						LogPath:    "/var/log/gse2",
						User:       "root",
					},
					Controller: types.ProcessController{
						StartCmd:   "./start.sh bkmonitorbeat",
						StopCmd:    "./stop.sh bkmonitorbeat",
						RestartCmd: "./restart.sh bkmonitorbeat",
						ReloadCmd:  "./reload.sh bkmonitorbeat",
						KillCmd:    "",
						VersionCmd: "./bkmonitorbeat -v",
						HealthCmd:  "",
					},
					Resource: types.ProcessResource{
						CPULimitPercent: 10,
						MemLimitPercent: 10,
					},
					MonitorPolicy: types.ProcessMonitorPolicy{
						AutoType:       types.ProcessAutoTypeTrusteeship,
						StartCheckSecs: 5,
						StopCheckSecs:  0,
						OpTimeoutSecs:  5,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.UnTrusteeshipProcess(tt.args.nCtx, tt.args.processSpec)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnTrusteeshipProcess() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("got: %v", got)
		})
	}
}

// TestHandler_QueryMultiProcessInfoMany test.
func TestHandler_QueryMultiProcessInfoMany(t *testing.T) {
	nCtx := contextx.New(context.Background())

	type args struct {
		nCtx               contextx.IContext
		procNameAgentIDMap []*types.ProcessAgentGroup
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				nCtx: nCtx,
				procNameAgentIDMap: []*types.ProcessAgentGroup{
					{
						PluginName: "bkmonitorbeat",
						AgentIDList: []string{
							"02000000000050568c81cb1756717277188u",
							"02000000000050568c468217552397004082",
							"02000000000050568c22a41760081249022o",
						},
					},
					{
						PluginName: "bkunifylogbeat",
						AgentIDList: []string{
							"02000000000050568c81cb1756717277188u",
							"02000000000050568c468217552397004082",
							"02000000000050568c22a41760081249022o",
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
			got, err := h.QueryMultiProcessInfoMany(tt.args.nCtx, tt.args.procNameAgentIDMap...)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueryMultiProcessInfoMany() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("got: %v", got)
		})
	}
}
