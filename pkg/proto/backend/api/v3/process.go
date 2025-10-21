/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package v3

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate check body.
func (x *ProcessListReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *ProcessListReq) AutoConvert() {}

// ConvertConditionsToTypes convert conditions to types.
func (x *ProcessListReq) ConvertConditionsToTypes() *types.ProcessCondition {
	return convertProcessConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions())
}

func convertProcessConditionsToTypes(
	exactCond *ProcessListReq_ExactConditions, fuzzyCond *ProcessListReq_FuzzyConditions) *types.ProcessCondition {

	condition := &types.ProcessCondition{}

	// exact conditions.
	if exactCond != nil {
		condition.ExactInclude = &types.ProcessExactFields{
			HostID:         exactCond.GetBkHostId(),
			PluginGroup:    exactCond.GetPluginGroup(),
			NodeGeneration: exactCond.GetNodeGeneration(),
			PlatformOS:     exactCond.GetPlatformOs(),
			PlatformArch:   exactCond.GetPlatformArch(),
			InfoStatus:     exactCond.GetStatus(),
			InfoAgentID:    exactCond.GetAgentId(),
			InfoVersion:    exactCond.GetVersion(),
			PluginName:     exactCond.GetPluginName(),
			PluginPkgName:  exactCond.GetPluginPkgName(),
		}
	}

	// fuzzy conditions.
	if fuzzyCond != nil {
		condition.FuzzyInclude = &types.ProcessFuzzyFields{
			Name:    fuzzyCond.GetName(),
			PkgName: fuzzyCond.GetPluginPkgName(),
		}
	}

	return condition
}

// ConvertPageToTypes converts page to types.
func (x *ProcessListReq) ConvertPageToTypes(maxLimit int) types.Page {
	return generatePage(x.GetPage(), maxLimit)
}

// ConvertConditionFromTypes converts condition from types.
func (x *ProcessListReq) ConvertConditionFromTypes(condition *types.ProcessCondition) {
	if condition == nil {
		return
	}

	// exact conditions.
	if condition.ExactInclude != nil {
		x.ExactIncludeConditions = &ProcessListReq_ExactConditions{
			BkHostId:       condition.ExactInclude.HostID,
			PluginGroup:    condition.ExactInclude.PluginGroup,
			NodeGeneration: condition.ExactInclude.NodeGeneration,
			PlatformOs:     condition.ExactInclude.PlatformOS,
			PlatformArch:   condition.ExactInclude.PlatformArch,
			Status:         condition.ExactInclude.InfoStatus,
			AgentId:        condition.ExactInclude.InfoAgentID,
			Version:        condition.ExactInclude.InfoVersion,
			PluginName:     condition.ExactInclude.PluginName,
			PluginPkgName:  condition.ExactInclude.PluginPkgName,
		}
	}

	// fuzzy conditions.
	if condition.FuzzyInclude != nil {
		x.FuzzyIncludeConditions = &ProcessListReq_FuzzyConditions{
			Name:          condition.FuzzyInclude.Name,
			PluginPkgName: condition.FuzzyInclude.PkgName,
		}
	}
}

// ConvertProcessFromTypes converts process from types.
func (x *ProcessListResp) ConvertProcessFromTypes(total int64, process []*types.Process) {
	items := make([]*Process, len(process))
	for idx, proc := range process {
		item := newEmptyProcess()
		*item.TenantId = proc.TenantID
		*item.BkHostId = proc.HostID
		*item.PluginName = proc.PluginName
		*item.PluginPkgName = proc.PluginPkgName
		*item.PluginGroup = proc.PluginGroup
		item.Platform.OsType = proc.Platform.OS.String()
		item.Platform.CpuArch = proc.Platform.Arch.String()
		*item.Generation = int64(proc.Generation)
		*item.ProcessInfo.Pid = int32(proc.Info.Pid)
		*item.ProcessInfo.Version = proc.Info.Version
		*item.ProcessInfo.AgentId = proc.Info.AgentID
		*item.ProcessInfo.Trusteeship = proc.Info.Trusteeship
		*item.ProcessInfo.Status = proc.Info.Status.String()
		*item.ProcessIdentity.Name = proc.Identity.Name
		*item.ProcessIdentity.SetupPath = proc.Identity.SetupPath
		*item.ProcessIdentity.PidPath = proc.Identity.PidPath
		*item.ProcessIdentity.ConfigPath = proc.Identity.ConfigPath
		*item.ProcessIdentity.LogPath = proc.Identity.LogPath
		*item.ProcessIdentity.User = proc.Identity.User
		*item.ProcessController.StartCmd = proc.Controller.StartCmd
		*item.ProcessController.StopCmd = proc.Controller.StopCmd
		*item.ProcessController.RestartCmd = proc.Controller.RestartCmd
		*item.ProcessController.ReloadCmd = proc.Controller.ReloadCmd
		*item.ProcessController.KillCmd = proc.Controller.KillCmd
		*item.ProcessController.VersionCmd = proc.Controller.VersionCmd
		*item.ProcessController.HealthCmd = proc.Controller.HealthCmd
		*item.ProcessResource.CpuLimitPercent = proc.Resource.CPULimitPercent
		*item.ProcessResource.MemLimitPercent = proc.Resource.MemLimitPercent
		*item.ProcessMonitorPolicy.AutoType = proc.MonitorPolicy.AutoType.String()
		*item.ProcessMonitorPolicy.StartCheckSeconds = proc.MonitorPolicy.StartCheckSecs
		*item.ProcessMonitorPolicy.StopCheckSeconds = proc.MonitorPolicy.StopCheckSecs
		*item.ProcessMonitorPolicy.OperateTimeoutSeconds = proc.MonitorPolicy.OpTimeoutSecs

		items[idx] = item
	}

	x.Data = &ProcessListResp_Data{
		Total: total,
		Items: items,
	}
}

func newEmptyProcess() *Process {
	return &Process{
		TenantId:      new(string),
		BkHostId:      new(int64),
		PluginName:    new(string),
		PluginPkgName: new(string),
		PluginGroup:   new(string),
		Platform: &Platform{
			OsType:  "",
			CpuArch: "",
		},
		Generation: new(int64),
		ProcessInfo: &ProcessInfo{
			Pid:         new(int32),
			Version:     new(string),
			AgentId:     new(string),
			Trusteeship: new(bool),
			Status:      new(string),
		},
		ProcessIdentity: &ProcessIdentity{
			Name:       new(string),
			SetupPath:  new(string),
			PidPath:    new(string),
			ConfigPath: new(string),
			LogPath:    new(string),
			User:       new(string),
		},
		ProcessController: &ProcessController{
			StartCmd:   new(string),
			StopCmd:    new(string),
			RestartCmd: new(string),
			ReloadCmd:  new(string),
			KillCmd:    new(string),
			VersionCmd: new(string),
			HealthCmd:  new(string),
		},
		ProcessResource: &ProcessResource{
			CpuLimitPercent: new(float64),
			MemLimitPercent: new(float64),
		},
		ProcessMonitorPolicy: &ProcessMonitorPolicy{
			AutoType:              new(string),
			StartCheckSeconds:     new(int64),
			StopCheckSeconds:      new(int64),
			OperateTimeoutSeconds: new(int64),
		},
	}
}

// ConvertProcessToTypes converts process to types.
func (x *ProcessListResp) ConvertProcessToTypes() ([]*types.Process, int64) {
	data := x.GetData()
	total := data.GetTotal()
	process := make([]*types.Process, len(data.GetItems()))
	for idx, proc := range data.GetItems() {
		item := &types.Process{
			TenantID:      proc.GetTenantId(),
			HostID:        proc.GetBkHostId(),
			PluginName:    proc.GetPluginName(),
			PluginPkgName: proc.GetPluginPkgName(),
			PluginGroup:   proc.GetPluginGroup(),
			Platform: platform.Platform{
				OS:   criteria.OSType(proc.GetPlatform().GetOsType()),
				Arch: criteria.CPUArch(proc.GetPlatform().GetCpuArch()),
			},
			Generation: types.Generation(proc.GetGeneration()),
			Info: types.ProcessInfo{
				Pid:         int(proc.GetProcessInfo().GetPid()),
				Version:     proc.GetProcessInfo().GetVersion(),
				AgentID:     proc.GetProcessInfo().GetAgentId(),
				Trusteeship: proc.GetProcessInfo().GetTrusteeship(),
				Status:      types.ProcessStatus(proc.GetProcessInfo().GetStatus()),
			},
			Identity: types.ProcessIdentity{
				Name:       proc.GetProcessIdentity().GetName(),
				SetupPath:  proc.GetProcessIdentity().GetSetupPath(),
				PidPath:    proc.GetProcessIdentity().GetPidPath(),
				ConfigPath: proc.GetProcessIdentity().GetConfigPath(),
				LogPath:    proc.GetProcessIdentity().GetLogPath(),
				User:       proc.GetProcessIdentity().GetUser(),
			},
			Controller: types.ProcessController{
				StartCmd:   proc.GetProcessController().GetStartCmd(),
				StopCmd:    proc.GetProcessController().GetStopCmd(),
				RestartCmd: proc.GetProcessController().GetRestartCmd(),
				ReloadCmd:  proc.GetProcessController().GetReloadCmd(),
				KillCmd:    proc.GetProcessController().GetKillCmd(),
				VersionCmd: proc.GetProcessController().GetVersionCmd(),
				HealthCmd:  proc.GetProcessController().GetHealthCmd(),
			},
			Resource: types.ProcessResource{
				CPULimitPercent: proc.GetProcessResource().GetCpuLimitPercent(),
				MemLimitPercent: proc.GetProcessResource().GetMemLimitPercent(),
			},
			MonitorPolicy: types.ProcessMonitorPolicy{
				AutoType:       types.ProcessAutoType(proc.GetProcessMonitorPolicy().GetAutoType()),
				StartCheckSecs: proc.GetProcessMonitorPolicy().GetStartCheckSeconds(),
				StopCheckSecs:  proc.GetProcessMonitorPolicy().GetStopCheckSeconds(),
				OpTimeoutSecs:  proc.GetProcessMonitorPolicy().GetOperateTimeoutSeconds(),
			},
		}

		process[idx] = item
	}

	return process, total
}
