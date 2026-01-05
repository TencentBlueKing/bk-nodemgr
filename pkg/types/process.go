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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
)

// ProcessStatus represents the status of process.
type ProcessStatus string

const (
	// ProcessStatusInit represents process init status.
	ProcessStatusInit ProcessStatus = "init"

	// ProcessStatusRunning represents process running status.
	ProcessStatusRunning ProcessStatus = "running"

	// ProcessStatusStopped represents process stopped status.
	ProcessStatusStopped ProcessStatus = "stopped"

	// ProcessStatusUnRegister represents process unregistered status.
	ProcessStatusUnRegister ProcessStatus = "unregister"

	// ProcessStatusUnknown represents unknown process status.
	ProcessStatusUnknown ProcessStatus = "unknown"
)

// Validate validate the process status.
func (status ProcessStatus) Validate() error {
	switch status {
	case ProcessStatusInit, ProcessStatusRunning, ProcessStatusStopped, ProcessStatusUnRegister, ProcessStatusUnknown:
		return nil
	default:
		return fmt.Errorf("invalid process status, process-status(%s)", status)
	}
}

// String convert process status to string.
func (status ProcessStatus) String() string {
	return string(status)
}

// ProcessInfo represents the process info.
type ProcessInfo struct {
	Pid       int
	Version   string
	AgentID   string
	AutoStart bool
	Status    ProcessStatus
}

// ProcessIdentity defines the identity of process.
type ProcessIdentity struct {
	// Name is the actual name of process binary file.
	Name       string
	SetupPath  string
	PidPath    string
	ConfigPath string
	LogPath    string
	User       string
}

// Validate validate the process identity.
func (identity ProcessIdentity) Validate() error {
	if identity.Name == "" {
		return fmt.Errorf("invalid process identity, name(%s)", identity.Name)
	}

	if identity.ConfigPath == "" {
		return fmt.Errorf("invalid process identity, config-path(%s)", identity.ConfigPath)
	}

	if identity.User == "" {
		return fmt.Errorf("invalid process identity, user(%s)", identity.User)
	}

	if identity.SetupPath == "" {
		return fmt.Errorf("invalid process identity, setup-path(%s)", identity.SetupPath)
	}

	if identity.PidPath == "" {
		return fmt.Errorf("invalid process identity, pid-path(%s)", identity.PidPath)
	}

	if identity.LogPath == "" {
		return fmt.Errorf("invalid process identity, log-path(%s)", identity.LogPath)
	}

	return nil
}

// ProcessResource defines the resource of process.
type ProcessResource struct {
	CPULimitPercent float64
	MemLimitPercent float64
}

// Validate validate the process resource.
func (resource ProcessResource) Validate() error {
	if resource.CPULimitPercent < 0 || resource.MemLimitPercent < 0 {
		return fmt.Errorf("invalid process resource, cpu(%f) mem(%f)", resource.CPULimitPercent, resource.MemLimitPercent)
	}

	return nil
}

// ProcessRestartType defines the restart type of process.
type ProcessRestartType string

const (
	// ProcessRestartTypeAuto means this process is auto restart.
	ProcessRestartTypeAuto ProcessRestartType = "auto"

	// ProcessRestartTypeManual means this process is manual restart.
	ProcessRestartTypeManual ProcessRestartType = "manual"
)

// Validate validate the process restart type.
func (rt ProcessRestartType) Validate() error {
	switch rt {
	case ProcessRestartTypeAuto, ProcessRestartTypeManual:
		return nil
	default:
		return fmt.Errorf("invalid process restart type(%s)", rt)
	}
}

// String convert process restart type to string.
func (rt ProcessRestartType) String() string {
	return string(rt)
}

// ProcessMonitorPolicy defines the monitor policy of process.
type ProcessMonitorPolicy struct {
	// RestartType defines the restart type of process.
	RestartType ProcessRestartType

	// StartCheckSecs start checking the time the process survives after the execution of the command, in seconds, with a default value of 5.
	StartCheckSecs int64

	// StopCheckSecs stop checking the time the process survives after the execution of the command, in seconds.
	StopCheckSecs int64

	// OpTimeoutSecs the timeout period of the operation, in seconds.
	OpTimeoutSecs int64
}

// Validate validate the process monitor policy.
func (monitorPolicy ProcessMonitorPolicy) Validate() error {
	if monitorPolicy.StartCheckSecs < 0 || monitorPolicy.StopCheckSecs < 0 {
		return fmt.Errorf("invalid process monitor policy, start-check-secs(%d) stop-check-secs(%d)",
			monitorPolicy.StartCheckSecs, monitorPolicy.StopCheckSecs)
	}

	return nil
}

// ProcessSpec defines the spec of process.
type ProcessSpec struct {
	// PluginName is the name for GSE metadata.
	PluginName    string
	AgentID       string
	Identity      ProcessIdentity
	Controller    ProcessController
	Resource      ProcessResource
	MonitorPolicy ProcessMonitorPolicy
}

// Validate validate the process spec.
func (spec ProcessSpec) Validate() error {
	if spec.PluginName == "" {
		return fmt.Errorf("failed to validate process spec: plugin-name is empty")
	}

	if spec.AgentID == "" {
		return fmt.Errorf("failed to validate process spec: agent-id is empty")
	}

	if err := spec.Identity.Validate(); err != nil {
		return fmt.Errorf("failed to validate process identity: %w", err)
	}

	if err := spec.Controller.Validate(); err != nil {
		return fmt.Errorf("failed to validate process controller: %w", err)
	}

	if err := spec.Resource.Validate(); err != nil {
		return fmt.Errorf("failed to validate process resource: %w", err)
	}

	if err := spec.MonitorPolicy.Validate(); err != nil {
		return fmt.Errorf("failed to validate process monitor policy: %w", err)
	}

	return nil
}

// ProcessController defines the control of plugin.
type ProcessController struct {
	StartCmd   string
	StopCmd    string
	RestartCmd string
	ReloadCmd  string
	KillCmd    string
	VersionCmd string
	HealthCmd  string
}

// Validate validates the release type.
func (controller ProcessController) Validate() error {
	if controller.StartCmd == "" {
		return fmt.Errorf("start cmd is empty")
	}
	if controller.StopCmd == "" {
		return fmt.Errorf("stop cmd is empty")
	}
	if controller.RestartCmd == "" {
		return fmt.Errorf("restart cmd is empty")
	}
	if controller.ReloadCmd == "" {
		return fmt.Errorf("reload cmd is empty")
	}

	return nil
}

// Process define the all info of plugin.
type Process struct {
	TenantID string
	HostID   int64

	// PluginName is the name of plugin, no actually the process name.
	PluginName string
	// PluginPkgName is the name of plugin release package.
	PluginPkgName string
	// PluginGroup is the group of plugin.
	PluginGroup string

	Platform   platform.Platform
	Generation Generation

	Info          ProcessInfo
	Identity      ProcessIdentity
	Controller    ProcessController
	Resource      ProcessResource
	MonitorPolicy ProcessMonitorPolicy
}

// ToProcessSpec converts the process to a process spec.
func (proc *Process) ToProcessSpec() ProcessSpec {
	return ProcessSpec{
		PluginName:    proc.PluginName,
		AgentID:       proc.Info.AgentID,
		Identity:      proc.Identity,
		Controller:    proc.Controller,
		Resource:      proc.Resource,
		MonitorPolicy: proc.MonitorPolicy,
	}
}

// ProcessInfoDelta is a map of process ID to process info.
type ProcessInfoDelta struct {
	HostID      int64
	PluginName  string
	ProcessInfo ProcessInfo
}

// ProcessAgentGroup process agent group.
type ProcessAgentGroup struct {
	PluginName  string
	ProcessName string
	AgentIDList []string
}
