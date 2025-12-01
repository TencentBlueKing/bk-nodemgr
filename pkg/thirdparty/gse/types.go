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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// runMode represents the gse agent run mode.
type runMode int

const (
	// RunModeAgent means agent run mode.
	RunModeAgent = 1
	// RunModeProxy means proxy run mode.
	RunModeProxy = 0
)

// agentStatusCode represents the gse agent status code.
type agentStatusCode int

const (
	// agentStatusCodeUnknown means status unknown.
	agentStatusCodeUnknown agentStatusCode = -1

	// agentStatusCodeInit means status init.
	agentStatusCodeInit agentStatusCode = 0

	// agentStatusCodeStarting means status starting.
	agentStatusCodeStarting agentStatusCode = 1

	// agentStatusCodeRunning means agent is OK.
	agentStatusCodeRunning agentStatusCode = 2

	// agentStatusCodeDamaged means status damaged.
	agentStatusCodeDamaged agentStatusCode = 3

	// agentStatusCodeBusy means status busy.
	agentStatusCodeBusy agentStatusCode = 4

	// agentStatusCodeUpgrade means agent is upgrading.
	agentStatusCodeUpgrade agentStatusCode = 5

	// agentStatusCodeStopping means status stopping.
	agentStatusCodeStopping agentStatusCode = 6

	// agentStatusCodeUninit means status uninit.
	agentStatusCodeUninit agentStatusCode = 7
)

// ToNodeStatus returns the node status.
func (code agentStatusCode) ToNodeStatus() types.NodeStatus {
	switch code {
	case agentStatusCodeInit:
		return types.NodeStatusInit
	case agentStatusCodeStarting:
		return types.NodeStatusStarting
	case agentStatusCodeRunning:
		return types.NodeStatusRunning
	case agentStatusCodeDamaged:
		return types.NodeStatusDamaged
	case agentStatusCodeBusy:
		return types.NodeStatusBusy
	case agentStatusCodeUpgrade:
		return types.NodeStatusUpgrade
	case agentStatusCodeStopping:
		return types.NodeStatusStopping
	case agentStatusCodeUninit:
		return types.NodeStatusUninit
	case agentStatusCodeUnknown:
		return types.NodeStatusUnknown
	default:
		return types.NodeStatusUnknown
	}
}

// scriptStatusCode represents the gse script status.
type scriptStatusCode int

const (
	// scriptStatusCodeReceived means script received. It is the initial state.
	scriptStatusCodeReceived scriptStatusCode = 0

	// scriptStatusCodeRunning means script running.
	scriptStatusCodeRunning scriptStatusCode = 1

	// scriptStatusCodeFinished means script finished and has an exit code 0.
	scriptStatusCodeFinished scriptStatusCode = 2

	// scriptStatusCodeTimeout means script timeout while execution.
	scriptStatusCodeTimeout scriptStatusCode = 3

	// scriptStatusCodeStopped means script is force stopped by others.
	scriptStatusCodeStopped scriptStatusCode = 4

	// scriptStatusCodeFailed means script execution failed and exit with a non-zero code.
	scriptStatusCodeFailed scriptStatusCode = 5

	// scriptStatusCodeAgentRestarted means agent was restarted during script execution and the script is aborted.
	scriptStatusCodeAgentRestarted scriptStatusCode = 6
)

func (code scriptStatusCode) ToScriptStatus() types.ScriptStatus {
	switch code {
	case scriptStatusCodeReceived:
		return types.ScriptStatusReceived
	case scriptStatusCodeRunning:
		return types.ScriptStatusRunning
	case scriptStatusCodeFinished:
		return types.ScriptStatusFinished
	case scriptStatusCodeTimeout:
		return types.ScriptStatusTimeout
	case scriptStatusCodeStopped:
		return types.ScriptStatusStopped
	case scriptStatusCodeFailed:
		return types.ScriptStatusFailed
	case scriptStatusCodeAgentRestarted:
		return types.ScriptStatusAgentRestarted
	default:
		return types.ScriptStatusUnknown
	}
}

// agentInfo describes the agent information from gse.
type agentInfo struct {
	BKAgentID      string          `json:"bk_agent_id"`
	BKCloudID      int             `json:"bk_cloud_id"`
	BKHostIP       string          `json:"bk_host_ip"`
	BKOSType       string          `json:"bk_os_type"`
	BKCPUArch      string          `json:"bk_cpu_architecture"`
	ParentIP       string          `json:"parent_ip"`
	ParentPort     uint            `json:"parent_port"`
	ReportTime     uint64          `json:"report_time"`
	Version        string          `json:"version"`
	CPURate        float32         `json:"cpu_rate"`
	MemRate        float32         `json:"mem_rate"`
	CPUNum         uint            `json:"cpu_num"`
	MemSize        uint64          `json:"mem_size"`
	StartTime      uint64          `json:"start_time"`
	LastWorkTime   uint64          `json:"last_work_time"`
	ConnCycleTime  string          `json:"conn_cycle_time"`
	RunMode        runMode         `json:"run_mode"`
	StatusCode     agentStatusCode `json:"status_code"`
	Status         string          `json:"status"`
	LastStatusCode agentStatusCode `json:"last_status_code"`
	LastStatus     string          `json:"last_status"`
	Remark         string          `json:"remark"`
}

// agentState describes the agent state. It is a subset of agentInfo.
type agentState struct {
	BKAgentID  string          `json:"bk_agent_id"`
	BKCloudID  int             `json:"bk_cloud_id"`
	Version    string          `json:"version"`
	RunMode    runMode         `json:"run_mode"`
	StatusCode agentStatusCode `json:"status_code"`
	ReportTime uint64          `json:"report_time"`
}

// Endpoint describes the basic endpoint information.
type Endpoint struct {
	BKAgentID     string `json:"bk_agent_id"`
	BKContainerID string `json:"bk_container_id"`
}

// EndpointWithAuth describes the endpoint with auth.
type EndpointWithAuth struct {
	Endpoint

	User     string `json:"user"`
	Password string `json:"pwd"`
}

// RespCommon describes the common response.
type RespCommon struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// BaseBroker describe the base broker.
type BaseBroker[T any] struct {
	RespCommon
	Data T `json:"data"`
}

// CodeOK define the success code.
const CodeOK = 0

// IsFailed returns the code of response.
func (resp *BaseBroker[T]) IsFailed() error {
	if resp.Code != CodeOK {
		return fmt.Errorf("code(%d), msg(%s)", resp.Code, resp.Message)
	}

	return nil
}

// ListAgentInfoReq describes the request data of list_agent_info.
type ListAgentInfoReq struct {
	AgentIDList []string `json:"agent_id_list"`
}

// ListAgentInfoResp describes the response data of list_agent_info.
type ListAgentInfoResp []*agentInfo

// ListAgentStateReq describes the request data of list_agent_state.
type ListAgentStateReq struct {
	AgentIDList []string `json:"agent_id_list"`
}

// ListAgentStateResp describe the response data of list_agent_state.
type ListAgentStateResp []*agentState

// ScriptDetail describes the script detail.
type ScriptDetail struct {
	Name      string `json:"script_name"`
	StoredDir string `json:"script_store_dir"`
	Content   string `json:"script_content"`
}

// ScriptAtomicTask describes the atomic task of script.
type ScriptAtomicTask struct {
	ID         int    `json:"atomic_task_id"`
	Command    string `json:"command"`
	TimeoutSec int    `json:"timeout_seconds"`
}

// ScriptAtomicRelation describes the atomic task relation.
// Relations defines the order of atomic tasks execution.
type ScriptAtomicRelation struct {
	AtomicTaskID int   `json:"atomic_task_id"`
	Dependencies []int `json:"atomic_task_id_idx"`
}

// AsyncExecuteScriptReq describes the request data of execute_script.
type AsyncExecuteScriptReq struct {
	Endpoints []*EndpointWithAuth     `json:"agents"`
	Scripts   []*ScriptDetail         `json:"scripts"`
	Atomics   []*ScriptAtomicTask     `json:"atomic_tasks"`
	Relations []*ScriptAtomicRelation `json:"atomic_tasks_relations"`
}

// ScriptResult describes the script result.
type ScriptResult struct {
	Endpoint

	AtomicTaskID int              `json:"atomic_task_id"`
	Status       scriptStatusCode `json:"status"`
	ErrorCode    int              `json:"error_code"`
	ErrorMessage string           `json:"error_msg"`
	StartTime    int64            `json:"start_time"`
	EndTime      int64            `json:"end_time"`
	ExitCode     int              `json:"script_exit_code"`
	Tag          string           `json:"tag"`
	ScreenLog    string           `json:"screen"`
}

// AsyncExecuteScriptResp describes the response data of execute_script.
type AsyncExecuteScriptResp struct {
	Result struct {
		TaskID string `json:"task_id"`
	} `json:"result"`
}

// ScriptAtomicTaskCondition describes the query condition of script atomic task.
type ScriptAtomicTaskCondition struct {
	ID     int  `json:"atomic_task_id"`
	Offset uint `json:"offset"`
	Limit  uint `json:"limit"`
}

// ScriptEndpointCondition describes the query endpoint condition of script atomic task.
type ScriptEndpointCondition struct {
	Endpoint

	Atomics []*ScriptAtomicTaskCondition `json:"atomic_tasks"`
}

// GetExecuteScriptResultReq describes the request data of query_script_execution_result.
type GetExecuteScriptResultReq struct {
	TaskID string `json:"task_id"`

	AgentTasks []*ScriptEndpointCondition `json:"agent_tasks"`
}

// GetExecuteScriptResultResp describes the response data of query_script_execution_result.
type GetExecuteScriptResultResp struct {
	Result []*ScriptResult `json:"result"`
}

// AsyncTerminateExecuteScriptReq describes the request data of terminate_script_execution.
type AsyncTerminateExecuteScriptReq struct {
	TaskID    string      `json:"task_id"`
	Endpoints []*Endpoint `json:"agents"`
}

// AsyncTerminateExecuteScriptResp describes the response data of terminate_script_execution.
type AsyncTerminateExecuteScriptResp struct {
	Result struct {
		TaskID string `json:"task_id"`
	} `json:"result"`
}

const (
	maxPushFileContentSize = 1 * 1024 * 100 // 100kb
)

// PushFileTask describes the push file task.
type PushFileTask struct {
	FileName    string      `json:"file_name"`
	StoreDir    string      `json:"store_dir"`
	FileContent string      `json:"file_content"`
	Md5         string      `json:"md5"`
	Owner       string      `json:"owner"`
	Permission  int64       `json:"right"`
	Endpoints   []*Endpoint `json:"agents"`
}

// AsyncPushFileReq describes the request data of push_file.
type AsyncPushFileReq struct {
	TimeoutSec uint            `json:"timeout_seconds,omitempty"`
	AutoMkdir  bool            `json:"auto_mkdir,omitempty"`
	Tasks      []*PushFileTask `json:"tasks"`
}

// AsyncPushFileResp describes the response data of push_file.
type AsyncPushFileResp struct {
	Result struct {
		TaskID string `json:"task_id"`
	} `json:"result"`
}

// TransferSource describes the transfer source.
type TransferSource struct {
	FileName  string           `json:"file_name"`
	StoredDir string           `json:"store_dir"`
	MD5       string           `json:"md5"`
	Endpoint  EndpointWithAuth `json:"agent"`
}

// TransferTarget describes the transfer target.
type TransferTarget struct {
	FileName   string              `json:"file_name"`
	StoredDir  string              `json:"store_dir"`
	Owner      string              `json:"owner"`
	Permission int                 `json:"permission"`
	Endpoints  []*EndpointWithAuth `json:"agents"`
}

// TransferDetail describes the transfer detail.
type TransferDetail struct {
	Source *TransferSource `json:"source"`
	Target *TransferTarget `json:"target"`
}

// AsyncTransferFileReq describes the request data of transfer_file.
type AsyncTransferFileReq struct {
	TimeoutSec        uint              `json:"timeout_seconds"`
	AutoMkdir         bool              `json:"auto_mkdir"`
	UploadSpeed       uint              `json:"upload_speed"`
	DownloadSpeed     uint              `json:"download_speed"`
	KeepSourceSeeding bool              `json:"keep_source_seeding"`
	Tasks             []*TransferDetail `json:"tasks"`
}

// AsyncTransferFileResp describes the response data of transfer_file.
type AsyncTransferFileResp struct {
	Result struct {
		TaskID string `json:"task_id"`
	} `json:"result"`
}

// FileResultContent describes the file result content.
type FileResultContent struct {
	DestinationAgentID     string `json:"dest_agent_id"`
	DestinationContainerID string `json:"dest_container_id"`
	DestFileDir            string `json:"dest_file_dir"`
	DestFileName           string `json:"dest_file_name"`
	Mode                   int    `json:"mode"`
	Progress               uint   `json:"progress"`
	SourceAgentID          string `json:"source_agent_id"`
	SourceContainerID      string `json:"source_container_id"`
	SourceFileDir          string `json:"source_file_dir"`
	SourceFileName         string `json:"source_file_name"`
	Speed                  uint64 `json:"speed"`
	Status                 int    `json:"status"`
	StatusInfo             string `json:"status_info"`
	Type                   string `json:"type"`
	StartTime              int64  `json:"start_time"`
	EndTime                int64  `json:"end_time"`
	Size                   uint64 `json:"size"`
}

// FileResult describes the file result.
type FileResult struct {
	ErrorCode    int               `json:"error_code"`
	ErrorMessage string            `json:"error_msg"`
	Content      FileResultContent `json:"content"`
}

// GetTransferFileResultReq describes the request data of query_file_transmission_result.
type GetTransferFileResultReq struct {
	TaskID    string      `json:"task_id"`
	Endpoints []*Endpoint `json:"agents"`
}

// GetTransferFileResultResp describes the response data of query_file_transmission_result.
type GetTransferFileResultResp struct {
	Result []*FileResult `json:"result"`
}

// AsyncTerminateTransferFileReq describes the request data of terminate_file_transmission.
type AsyncTerminateTransferFileReq struct {
	TaskID    string      `json:"task_id"`
	Endpoints []*Endpoint `json:"agents"`
}

// AsyncTerminateTransferFileResp describes the response data of terminate_file_transmission.
type AsyncTerminateTransferFileResp struct {
	Result struct {
		TaskID string `json:"task_id"`
	} `json:"result"`
}

// OperateAgentReq describes the request data of operate_agent.
type OperateAgentReq struct {
	Type              int      `json:"type"`
	CurrentVersion    string   `json:"current_version"`
	TargetVersionSign string   `json:"target_version_sign"`
	Timeout           int      `json:"timeout"`
	Force             bool     `json:"force"`
	Remark            string   `json:"remark"`
	AgentIDList       []string `json:"agent_id_list"`
}

// OperateAgentResp describes the response data of operate_agent.
type OperateAgentResp struct {
	MissingAgentIDList []string `json:"results"`
}

// GSE operate agent type.
const (
	operateAgentTypeUnknown = 0
	operateAgentTypeRestart = 1
)

const (
	scriptExtBash       = "sh"
	scriptExtBat        = "bat"
	scriptExtPowershell = "ps1"
	scriptExtPython     = "py"
	scriptExtKsh        = "ksh"
	scriptExtZsh        = "zsh"
)

func getScriptExt(scriptType types.ScriptType) (string, error) {
	switch scriptType {
	case types.ScriptTypeBash:
		return scriptExtBash, nil
	case types.ScriptTypeKsh:
		return scriptExtKsh, nil
	case types.ScriptTypeZsh:
		return scriptExtZsh, nil
	case types.ScriptTypeBat:
		return scriptExtBat, nil
	case types.ScriptTypePowershell:
		return scriptExtPowershell, nil
	case types.ScriptTypePython:
		return scriptExtPython, nil
	default:
		return "", fmt.Errorf("unsupported script type: %s", scriptType)
	}
}

// procMeta describes the meta data.
type procMeta struct {
	Namespace string             `json:"namespace"`
	Name      string             `json:"name"`
	Labels    procInfoMetaLabels `json:"labels"`
}

// procInfoMetaLabels describes the labels.
type procInfoMetaLabels struct {
	ProcName string `json:"proc_name"`
}

// procSpec describes the spec data of gse proc.
type procSpec struct {
	Identity      procSpecIdentity      `json:"identity"`
	Control       procSpecControl       `json:"control"`
	Resource      procSpecResource      `json:"resource"`
	MonitorPolicy procSpecMonitorPolicy `json:"monitor_policy"`
}

// procSpecIdentity describes the identity data of gse proc spec.
type procSpecIdentity struct {
	ProcName   string `json:"proc_name"`
	SetupPath  string `json:"setup_path"`
	PidPath    string `json:"pid_path"`
	ConfigPath string `json:"config_path"`
	LogPath    string `json:"log_path"`
	User       string `json:"user"`
}

// procSpecControl describes the control data of gse proc spec.
type procSpecControl struct {
	StartCmd   string `json:"start_cmd"`
	StopCmd    string `json:"stop_cmd"`
	RestartCmd string `json:"restart_cmd"`
	ReloadCmd  string `json:"reload_cmd"`
	KillCmd    string `json:"kill_cmd"`
	VersionCmd string `json:"version_cmd"`
	HealthCmd  string `json:"health_cmd"`
}

type procSpecMonitorPolicyAutoType int

const (
	// procSpecMonitorPolicyAutoTypeTrusteeship this means that this is a resident process, and will be restarted automatically.
	procSpecMonitorPolicyAutoTypeTrusteeship procSpecMonitorPolicyAutoType = 1

	// procSpecMonitorPolicyAutoTypeOnce this means that this is a one-time process, executed only once, and then exited.
	procSpecMonitorPolicyAutoTypeOnce procSpecMonitorPolicyAutoType = 2
)

// procSpecMonitorPolicy describes the monitor policy of gse proc spec.
type procSpecMonitorPolicy struct {
	AutoType       procSpecMonitorPolicyAutoType `json:"auto_type"`
	StartCheckSecs int64                         `json:"start_check_secs"`
	StopCheckSecs  int64                         `json:"stop_check_secs"`
	OpTimeoutSecs  int64                         `json:"op_timeout"`
}

// procSpecResource describes the resource data of gse proc spec.
type procSpecResource struct {
	CPU float64 `json:"cpu"`
	Mem float64 `json:"mem"`
}

// procOperateResultCode describes the code of operate proc.
type procOperateResultCode int64

const (
	// procOperateResultCodeOK means success.
	procOperateResultCodeOK procOperateResultCode = 0
	// procOperateResultCodeRunning means running.
	procOperateResultCodeRunning procOperateResultCode = 115
)

// getProcOperateResultV2Req describes the request data of get_proc_operate_result_v2.
type getProcOperateResultV2Req struct {
	TaskID string `json:"task_id"`
}

// getProcOperateResultV2Resp describes the response data of get_proc_operate_result_v2.
type getProcOperateResultV2Resp = map[string]procOperateResult

// procOperateResult describes the result of update_proc_info.
type procOperateResult struct {
	ErrorCode procOperateResultCode `json:"error_code"`
	ErrorMsg  string                `json:"error_msg"`
	Content   string                `json:"content"`
}

// queryProcessContent describes the content when do query operate.
// This struct has no valuable data and is only used for protocol conversion.
type queryProcessContent struct {
	BkAgentID string `json:"bk_agent_id"`
	Utctime   string `json:"utctime"`
	Utctime2  string `json:"utctime2"`
	Timezone  int    `json:"timezone"`
	Process   []struct {
		Procname string `json:"procname"`
		// Notice: this instance list always is a single list.
		// It's just a historical legacy, don't modify it.
		Instance []processInfo `json:"instance"`
	} `json:"process"`
}

// processInfo describes the instance info.
type processInfo struct {
	ProcessName string  `json:"processName"`
	Version     string  `json:"version"`
	IsAuto      bool    `json:"isAuto"`
	CPUUsage    float64 `json:"cpuUsage"`
	PhyMemUsage float64 `json:"phyMemUsage"`
	Pid         int     `json:"pid"`
	ReportTime  int     `json:"report_time"`
}

// controlProcessContent describes the content when do start, stop, trusteeship, untrusteeship, kill, reload,
// This struct has no valuable data and is only used for protocol conversion.
type controlProcessContent struct {
	Value []struct {
		BkAgentID string `json:"bk_agent_id"`
		ProcName  string `json:"procName"`
		SetupPath string `json:"setupPath"`
		Result    string `json:"result"`
		IsAuto    bool   `json:"isAuto"`
	} `json:"value"`
}

type procOperateCode int

const (
	// procOperateCodeStart proc operate code start.
	procOperateCodeStart procOperateCode = 0
	// procOperateCodeStop proc operate code stop.
	procOperateCodeStop procOperateCode = 1
	// procOperateCodeStatus proc operate code status.
	procOperateCodeStatus procOperateCode = 2
	// procOperateCodeTrusteeship proc operate code trusteeship.
	procOperateCodeTrusteeship procOperateCode = 3
	// procOperateCodeUnTrusteeship proc operate code untrusteeship.
	procOperateCodeUnTrusteeship procOperateCode = 4
	// procOperateCodeRestart proc operate code restart.
	procOperateCodeRestart procOperateCode = 7
	// procOperateCodeReload proc operate code reload.
	procOperateCodeReload procOperateCode = 8
)

// procOperateReq describes the request data of operate_proc.
type procOperateReq struct {
	Meta        procMeta        `json:"meta"`
	OpType      procOperateCode `json:"op_type"`
	AgentIDList []string        `json:"agent_id_list"`
	Spec        procSpec        `json:"spec"`
}

// operateProcV2Req describes the request data of operate_proc_v2.
type operateProcV2Req procOperateReq

// operateProcV2Resp describes the response data of operate_proc_v2.
type operateProcV2Resp struct {
	TaskID string `json:"task_id"`
}

type operateProcV2MultiReq []*procOperateReq

// operateProcMultiReq describes the request data of operate_proc_multi.
type operateProcMultiReq struct {
	ProcOperateReq operateProcV2MultiReq `json:"proc_operate_req"`
}

// operateProcV2Resp describes the response data of operate_proc_v2.
type operateProcMultiResp struct {
	TaskID string `json:"task_id"`
}
