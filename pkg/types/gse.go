/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types

import (
	"context"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/filex/iface"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
)

// AgentStatusCode represents the gse agent status code.
type AgentStatusCode int

const (
	// AgentStatusCodeUnknown means status unknown.
	AgentStatusCodeUnknown AgentStatusCode = iota - 1

	// AgentStatusCodeInit means status init.
	AgentStatusCodeInit

	// AgentStatusCodeStarting means status starting.
	AgentStatusCodeStarting

	// AgentStatusCodeRunning means agent is OK.
	AgentStatusCodeRunning

	// AgentStatusCodeDamaged means status damaged.
	AgentStatusCodeDamaged

	// AgentStatusCodeBusy means status busy.
	AgentStatusCodeBusy

	// AgentStatusCodeUpgrade means agent is upgrading.
	AgentStatusCodeUpgrade

	// AgentStatusCodeStopping means status stopping.
	AgentStatusCodeStopping

	// AgentStatusCodeUninit means status uninit.
	AgentStatusCodeUninit
)

// String returns the string representation of the AgentStatusCode.
func (code AgentStatusCode) String() string {
	switch code {
	case AgentStatusCodeUnknown:
		return "Unknown"
	case AgentStatusCodeInit:
		return "Init"
	case AgentStatusCodeStarting:
		return "Starting"
	case AgentStatusCodeRunning:
		return "Running"
	case AgentStatusCodeDamaged:
		return "Damaged"
	case AgentStatusCodeBusy:
		return "Busy"
	case AgentStatusCodeUpgrade:
		return "Upgrade"
	case AgentStatusCodeStopping:
		return "Stopping"
	case AgentStatusCodeUninit:
		return "Uninit"
	}

	return "Unknown"
}

// ToNodeStatus returns the node status.
func (code AgentStatusCode) ToNodeStatus() NodeStatus {
	switch code {
	case AgentStatusCodeInit:
		return NodeStatusInit
	case AgentStatusCodeStarting:
		return NodeStatusStarting
	case AgentStatusCodeRunning:
		return NodeStatusRunning
	case AgentStatusCodeDamaged:
		return NodeStatusDamaged
	case AgentStatusCodeBusy:
		return NodeStatusBusy
	case AgentStatusCodeUpgrade:
		return NodeStatusUpgrade
	case AgentStatusCodeStopping:
		return NodeStatusStopping
	case AgentStatusCodeUninit:
		return NodeStatusUninit
	default:
		return NodeStatusUnknown
	}
}

// AgentState describes the agent state. It is a subset of AgentInfo.
type AgentState struct {
	AgentID        string
	CloudID        int
	Version        string
	ReportTime     uint64
	StatusCode     AgentStatusCode
	LastStatusCode AgentStatusCode
}

// AgentInfo represents the gse agent information.
type AgentInfo struct {
	AgentState

	HostIP        string
	OSType        criteria.OSType
	Arch          criteria.CPUArch
	ParentIP      string
	ParentPort    uint
	CPURate       float32
	MemRate       float32
	StartTime     uint64
	LastWorkTime  uint64
	ConnCycleTime string
	Status        string
	LastStatus    string
	RunMode       int
	Remark        string
}

// Endpoint represents the gse execution/transmission target.
type Endpoint struct {
	AgentID     string
	ContainerID string
}

// EndpointWithAuth represents the gse execution/transmission target with auth.
type EndpointWithAuth struct {
	Endpoint

	User     string
	Password string
}

// EndpointWithRestrict represents the gse execution/transmission target with restrict.
type EndpointWithRestrict struct {
	Endpoint

	Offset uint
	Limit  uint
}

// ScriptStatus represents the gse script status.
type ScriptStatus int

const (
	// ScriptStatusReceived means script received. It is the initial state.
	ScriptStatusReceived ScriptStatus = iota

	// ScriptStatusRunning means script running.
	ScriptStatusRunning

	// ScriptStatusFinished means script finished and has an exit code 0.
	ScriptStatusFinished

	// ScriptStatusTimeout means script timeout while execution.
	ScriptStatusTimeout

	// ScriptStatusStopped means script is force stopped by others.
	ScriptStatusStopped

	// ScriptStatusFailed means script execution failed and exit with a non-zero code.
	ScriptStatusFailed

	// ScriptStatusAgentRestarted means agent was restarted during script execution and the script is aborted.
	ScriptStatusAgentRestarted
)

// ScriptResult represents the gse script result.
type ScriptResult struct {
	Endpoint

	Status       ScriptStatus
	ErrorCode    int
	ErrorMessage string
	StartTime    time.Time
	EndTime      time.Time
	ExitCode     int
	ScreenLog    string
}

// TransferOptions represents the gse transfer option.
type TransferOptions struct {
	Timeout               time.Duration
	AutoMkdir             bool
	UploadSpeedMBPerSec   uint
	DownloadSpeedMBPerSec uint
}

// TransferSource represents the gse transfer source.
type TransferSource struct {
	FileName  string
	StoredDir string
	Endpoint  EndpointWithAuth
}

// TransferTarget represents the gse transfer target.
type TransferTarget struct {
	StoredDir string
	Endpoints []*EndpointWithAuth
}

// TransferDetail represents the single gse transfer detail.
type TransferDetail struct {
	Source TransferSource
	Target TransferTarget
}

// TransferMode represents the gse transfer mode.
type TransferMode int

const (
	// TransferModeUpload means upload mode.
	TransferModeUpload TransferMode = iota

	// TransferModeDownload means download mode.
	TransferModeDownload
)

// TransferStatus represents the gse transfer status.
type TransferStatus int

const (
	// TransferStatusRunning means running.
	TransferStatusRunning TransferStatus = 500

	// TransferStatusStopped means stopped.
	TransferStatusStopped TransferStatus = 506

	// TransferStatusChecking means checking files' checksum.
	TransferStatusChecking TransferStatus = 508

	// TransferStatusTimeout means timeout.
	TransferStatusTimeout TransferStatus = 512

	// TransferStatusEndUploading means uploading ended, including success and failure. Check with error code.
	TransferStatusEndUploading TransferStatus = 590

	// TransferStatusEndDownloading means downloading ended, including success and failure. Check with error code.
	TransferStatusEndDownloading TransferStatus = 592
)

// TransferResult represents the gse transfer result.
type TransferResult struct {
	Source Endpoint
	Target Endpoint

	Mode          TransferMode
	Progress      uint
	SpeedKBPerSec uint64
	SizeBytes     uint64

	SourceDir      string
	SourceFileName string
	TargetDir      string
	TargetFileName string

	ErrorCode     int
	ErrorMessage  string
	StatusCode    TransferStatus
	StatusMessage string

	StartTime time.Time
	EndTime   time.Time
}

// ISimpleTransferHandler defines single transfer handler interface.
// in simple transfer, there are only one file from one source to one target.
type ISimpleTransferHandler interface {
	// GetTaskID get task id.
	GetTaskID() string

	// GetFileInfo get file info.
	GetFileInfo() iface.FileInfo

	// WaitUntilDone wait until done.
	WaitUntilDone(ctx context.Context) (*SimpleTransferResult, error)
}

// SimpleTransferResult represents the simple transfer result.
type SimpleTransferResult struct {
	Mode          TransferMode
	ErrorCode     int
	ErrorMessage  string
	Terminated    bool
	Progress      uint
	SpeedKBPerSec uint64
	StartTime     time.Time
	EndTime       time.Time
}

// ConvertTransferResultToSimple convert transfer result to simple transfer result.
func ConvertTransferResultToSimple(result *TransferResult) *SimpleTransferResult {
	return &SimpleTransferResult{
		Mode:         result.Mode,
		ErrorCode:    result.ErrorCode,
		ErrorMessage: result.ErrorMessage,
		Terminated: result.Mode == TransferModeUpload && result.StatusCode == TransferStatusEndUploading ||
			result.Mode == TransferModeDownload && result.StatusCode == TransferStatusEndDownloading,
		Progress:      result.Progress,
		SpeedKBPerSec: result.SpeedKBPerSec,
		StartTime:     result.StartTime,
		EndTime:       result.EndTime,
	}
}
