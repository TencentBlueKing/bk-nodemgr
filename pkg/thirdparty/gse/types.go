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
)

// AgentInfo describes the agent information from gse.
type AgentInfo struct {
	BKAgentID      string  `json:"bk_agent_id"`
	BKCloudID      int     `json:"bk_cloud_id"`
	BKHostIP       string  `json:"bk_host_ip"`
	BKOSType       string  `json:"bk_os_type"`
	ReportTime     uint64  `json:"report_time"`
	ParentIP       string  `json:"parent_ip"`
	ParentPort     uint    `json:"parent_port"`
	Version        string  `json:"version"`
	CPURate        float32 `json:"cpu_rate"`
	MemRate        float32 `json:"mem_rate"`
	StartTime      uint64  `json:"start_time"`
	LastWorkTime   uint64  `json:"last_work_time"`
	ConnCycleTime  string  `json:"conn_cycle_time"`
	RunMode        int     `json:"run_mode"`
	StatusCode     int     `json:"status_code"`
	Status         string  `json:"status"`
	LastStatusCode int     `json:"last_status_code"`
	LastStatus     string  `json:"last_status"`
	Remark         string  `json:"remark"`
}

// AgentState describes the agent state. It is a subset of AgentInfo.
type AgentState struct {
	BKAgentID      string `json:"bk_agent_id"`
	BKCloudID      int    `json:"bk_cloud_id"`
	Version        string `json:"version"`
	ReportTime     uint64 `json:"report_time"`
	StatusCode     int    `json:"status_code"`
	Status         string `json:"status"`
	LastStatusCode int    `json:"last_status_code"`
	LastStatus     string `json:"last_status"`
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
	Message string `json:"messaage"`
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
		return fmt.Errorf("code(%d) , msg(%s) ", resp.Code, resp.Message)
	}

	return nil
}

// ListAgentInfoReq describes the request data of list_agent_info.
type ListAgentInfoReq struct {
	AgentIDList []string `json:"agent_id_list"`
}

// ListAgentInfoResp describes the response data of list_agent_info.
type ListAgentInfoResp []*AgentInfo

// ListAgentStateReq describes the request data of list_agent_state.
type ListAgentStateReq struct {
	AgentIDList []string `json:"agent_id_list"`
}

// ListAgentStateResp describe the response data of list_agent_state.
type ListAgentStateResp []*AgentState

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

	AtomicTaskID int    `json:"atomic_task_id"`
	Status       int    `json:"status"`
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_msg"`
	StartTime    int64  `json:"start_time"`
	EndTime      int64  `json:"end_time"`
	ExitCode     int    `json:"script_exit_code"`
	Tag          string `json:"tag"`
	ScreenLog    string `json:"screen"`
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
	TimeoutSec    uint              `json:"timeout_seconds"`
	AutoMkdir     bool              `json:"auto_mkdir"`
	UploadSpeed   uint              `json:"upload_speed"`
	DownloadSpeed uint              `json:"download_speed"`
	Tasks         []*TransferDetail `json:"tasks"`
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

// SystemArchitectureInfo describes the special system architecture information.
type SystemArchitectureInfo struct {
	Signature string `json:"signature"`
	StoredDir string `json:"store_dir"`
	Filename  string `json:"filename"`
	User      string `json:"user"`
	BKAgentID string `json:"bk_agent_id"`
}

// TargetVersionAgentInfo describes the target version gse agent information.
type TargetVersionAgentInfo struct {
	Generation   uint                   `json:"generation"`
	Version      string                 `json:"version"`
	LinuxX8664   SystemArchitectureInfo `json:"linux_x86_64"`
	LinuxAarch64 SystemArchitectureInfo `json:"linux_aarch64"`
	WindowsX8664 SystemArchitectureInfo `json:"windows_x86_64"`
	DarwinX8664  SystemArchitectureInfo `json:"darwin_x86_64"`
}

// TargetVersionProxyInfo describes the target version gse proxy information.
type TargetVersionProxyInfo struct {
	Generation uint                    `json:"generation"`
	Version    string                  `json:"version"`
	Agent      UpgradeProxyVersionInfo `json:"agent"`
	Data       UpgradeProxyVersionInfo `json:"data"`
	File       UpgradeProxyVersionInfo `json:"file"`
}

// OperateDirectory describes the gse operate directory.
type OperateDirectory struct {
	LinuxX8664Dir   string `json:"linux_x86_64_dir"`
	LinuxAarch64Dir string `json:"linux_aarch64_dir"`
	WindowsX8664Dir string `json:"windows_x86_64_dir"`
	DarwinX8664Dir  string `json:"darwin_x86_64_dir"`
}

// UpgradeProxyAgentVersionInfo describes the gse upgrade proxy agent version information.
type UpgradeProxyVersionInfo struct {
	LinuxX8664   SystemArchitectureInfo `json:"linux_x86_64"`
	LinuxAarch64 SystemArchitectureInfo `json:"linux_aarch64"`
	WindowsX8664 SystemArchitectureInfo `json:"windows_x86_64"`
	DarwinX8664  SystemArchitectureInfo `json:"darwin_x86_64"`
}

// BasicCountInfo describes the basic count information.
type BasicCountInfo struct {
	LinuxX8664AgentCount   uint `json:"linux_x86_64_agent_count"`
	LinuxAarch64AgentCount uint `json:"linux_aarch64_agent_count"`
	WindowsX8664AgentCount uint `json:"windows_x86_64_agent_count"`
	DarwinX8664AgentCount  uint `json:"darwin_x86_64_agent_count"`
	LinuxX8664BatchCount   uint `json:"linux_x86_64_batch_count"`
	LinuxAarch64BatchCount uint `json:"linux_aarch64_batch_count"`
	WindowsX8664BatchCount uint `json:"windows_x86_64_batch_count"`
	DarwinX8664BatchCount  uint `json:"darwin_x86_64_batch_count"`
}

// AgentResult describes the agent result.
type AgentResult struct {
	State   int    `json:"state"`
	Message string `json:"message"`
}

// StepResult describes the step result.
type StepResult struct {
	Index          int                    `json:"index"`
	State          string                 `json:"state"`
	Type           string                 `json:"type"`
	StartTimestamp uint64                 `json:"start_timestamp"`
	EndTimeStamp   uint64                 `json:"end_timestamp"`
	TaskID         string                 `json:"task_id"`
	AgentResult    map[string]AgentResult `json:"agent_result"`
}

// BatchResult describes the batch result.
type BatchResult struct {
	Index          int                    `json:"index"`
	State          string                 `json:"state"`
	Type           string                 `json:"type"`
	StartTimestamp uint64                 `json:"start_timestamp"`
	EndTimeStamp   uint64                 `json:"end_timestamp"`
	AgentCount     uint                   `json:"agent_count"`
	StepCount      uint                   `json:"step_count"`
	AgentResult    map[string]AgentResult `json:"agent_result"`
	StepResult     []StepResult           `json:"step_result"`
}

// AsyncUpgradeAgentReq describes the request data of upgrade_agent.
type AsyncUpgradeAgentReq struct {
	TargetVersionAgentInfo TargetVersionAgentInfo `json:"target_version_agent_info"`
	AgentOperateDirectory  OperateDirectory       `json:"operate_directory"`
	CurrentVersion         string                 `json:"current_version"`
	Timeout                uint                   `json:"timeout_seconds"`
	AgentIDList            []string               `json:"agent_id_list"`
}

// AsyncUpgradeAgentResp describes the response data of upgrade_agent.
type AsyncUpgradeAgentResp struct {
	TaskID string `json:"task_id"`
}

// AsyncTerminateUpgradeAgentReq describes the request data of terminate_upgrade_agent.
type AsyncTerminateUpgradeAgentReq struct {
	TaskID      string   `json:"task_id"`
	AgentIDList []string `json:"agent_id_list"`
}

// AsyncTerminateUpgradeAgentResp describes the response data of terminate_upgrade_agent.
type AsyncTerminateUpgradeAgentResp struct{}

// AsyncUpgradeProxyReq describes the request data of get_upgrade_agent_result.
type AsyncGetUpgradeAgentResultReq struct {
	TaskID string `json:"task_id"`
}

// AsyncUpgradeProxyResp describes the response data of get_upgrade_agent_result.
type AsyncGetUpgradeAgentResultResp struct {
	State          string                 `json:"state"`
	ServiceID      string                 `json:"service_id"`
	CurrentVersion string                 `json:"current_version"`
	TargetVersion  string                 `json:"target_version"`
	Timeout        uint                   `json:"timeout_seconds"`
	StartTimestamp uint64                 `json:"start_timestamp"`
	EndTimeStamp   uint64                 `json:"end_timestamp"`
	AgentCount     uint                   `json:"agent_count"`
	BatchCount     uint                   `json:"batch_count"`
	Info           BasicCountInfo         `json:"info"`
	AgentResult    map[string]AgentResult `json:"agent_result"`
	BatchResult    []BatchResult          `json:"batch_result"`
}

// AsyncUpgradeProxyReq describes the request data of upgrade_proxy.
type AsyncUpgradeProxyReq struct {
	TargetVersionProxyInfo TargetVersionProxyInfo `json:"target_version_proxy_info"`
	ProxyOperateDirectory  OperateDirectory       `json:"operate_directory"`
	CurrentVersion         string                 `json:"current_version"`
	Timeout                uint                   `json:"timeout_seconds"`
	AgentIDList            []string               `json:"agent_id_list"`
}

// AsyncUpgradeProxyResp describes the response data of upgrade_proxy.
type AsyncUpgradeProxyResp struct {
	TaskID string `json:"task_id"`
}

// AsyncTerminateUpgradeProxyReq describes the request data of terminate_upgrade_proxy.
type AsyncTerminateUpgradeProxyReq struct {
	TaskID      string   `json:"task_id"`
	AgentIDList []string `json:"agent_id_list"`
}

// AsyncTerminateUpgradeProxyResp describes the response data of terminate_upgrade_proxy.
type AsyncTerminateUpgradeProxyResp struct{}

// AsyncGetUpgradeProxyResultReq describes the request data of get_upgrade_proxy_result.
type AsyncGetUpgradeProxyResultReq struct {
	TaskID string `json:"task_id"`
}

// AsyncGetUpgradeProxyResultResp describes the response data of get_upgrade_proxy_result.
type AsyncGetUpgradeProxyResultResp struct {
	State          string                 `json:"state"`
	ServiceID      string                 `json:"service_id"`
	CurrentVersion string                 `json:"current_version"`
	TargetVersion  string                 `json:"target_version"`
	Timeout        uint                   `json:"timeout_seconds"`
	StartTimestamp uint64                 `json:"start_timestamp"`
	EndTimeStamp   uint64                 `json:"end_timestamp"`
	AgentCount     uint                   `json:"agent_count"`
	BatchCount     uint                   `json:"batch_count"`
	Info           BasicCountInfo         `json:"info"`
	AgentResult    map[string]AgentResult `json:"agent_result"`
	BatchResult    []BatchResult          `json:"batch_result"`
}
