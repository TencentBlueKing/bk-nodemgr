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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
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
	exactCond *ProcessExactConditions, fuzzyCond *ProcessFuzzyConditions) *types.ProcessCondition {

	condition := &types.ProcessCondition{}

	// exact conditions.
	if exactCond != nil {
		condition.ExactInclude = &types.ProcessExactFields{
			HostID:         exactCond.GetBkHostId(),
			BizID:          exactCond.GetBkBizId(),
			PluginGroup:    exactCond.GetPluginGroup(),
			NodeGeneration: exactCond.GetNodeGeneration(),
			PlatformOS:     exactCond.GetPlatformOs(),
			PlatformArch:   exactCond.GetPlatformArch(),
			InfoStatus: conv.SliceToSlice[string, types.ProcessStatus](exactCond.GetStatus(), func(status string) types.ProcessStatus {
				return types.ProcessStatus(status)
			}),
			InfoAgentID:   exactCond.GetAgentId(),
			InfoVersion:   exactCond.GetVersion(),
			PluginName:    exactCond.GetPluginName(),
			PluginPkgName: exactCond.GetPluginPkgName(),
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
func (x *ProcessListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage())
}

// ConvertProcessFromTypes converts process from types.
func (x *ProcessListResp) ConvertProcessFromTypes(total int64, process []*types.Process) {
	items := make([]*Process, len(process))
	for idx, proc := range process {
		item := newEmptyProcess()
		*item.TenantId = proc.TenantID
		*item.BkHostId = proc.HostID
		*item.BkBizId = proc.BizID
		*item.PluginName = proc.PluginName
		*item.PluginPkgName = proc.PluginPkgName
		*item.PluginGroup = proc.PluginGroup
		item.Platform.OsType = proc.Platform.OS.String()
		item.Platform.CpuArch = proc.Platform.Arch.String()
		*item.Generation = int64(proc.Generation)
		*item.ProcessInfo.Pid = int32(proc.Info.Pid)
		*item.ProcessInfo.Version = proc.Info.Version
		*item.ProcessInfo.AgentId = proc.Info.AgentID
		*item.ProcessInfo.AutoStart = proc.Info.AutoStart
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
		*item.ProcessMonitorPolicy.RestartType = proc.MonitorPolicy.RestartType.String()
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
		BkBizId:       new(int64),
		PluginName:    new(string),
		PluginPkgName: new(string),
		PluginGroup:   new(string),
		Platform: &Platform{
			OsType:  "",
			CpuArch: "",
		},
		Generation: new(int64),
		ProcessInfo: &ProcessInfo{
			Pid:       new(int32),
			Version:   new(string),
			AgentId:   new(string),
			AutoStart: new(bool),
			Status:    new(string),
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
			RestartType:           new(string),
			StartCheckSeconds:     new(int64),
			StopCheckSeconds:      new(int64),
			OperateTimeoutSeconds: new(int64),
		},
	}
}

// Validate check body.
func (x *GetProcessDistributionByHostIDReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *GetProcessDistributionByHostIDReq) AutoConvert() {}

// ConvertConditionsToTypes convert conditions.
func (x *GetProcessDistributionByHostIDReq) ConvertConditionsToTypes() *types.ProcessCondition {
	return convertProcessConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions())
}

// Validate check body.
func (x *GetProcessDistributionByPluginNameReq) Validate() error {
	return nil
}

// ConvertResultFromTypes convert result.
func (x *GetProcessDistributionByHostIDResp) ConvertResultFromTypes(data map[int64]int64) {
	x.Data = data
}

// AutoConvert auto convert.
func (x *GetProcessDistributionByPluginNameReq) AutoConvert() {}

// ConvertConditionsToTypes convert conditions.
func (x *GetProcessDistributionByPluginNameReq) ConvertConditionsToTypes() *types.ProcessCondition {
	return convertProcessConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions())
}

// ConvertResultFromTypes convert result.
func (x *GetProcessDistributionByPluginNameResp) ConvertResultFromTypes(data map[string]int64) {
	x.Data = data
}

// ===============================================================================
// DistinctProcessReq
// ===============================================================================

// Validate check body.
func (x *DistinctProcessReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *DistinctProcessReq) AutoConvert() {}

// ConvertConditionsToTypes convert conditions.
func (x *DistinctProcessReq) ConvertConditionsToTypes() *types.ProcessCondition {
	return convertProcessConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions())
}

// ConvertResultFromTypes convert result.
func (x *DistinctProcessResp) ConvertResultFromTypes(data *types.ProcessDistinctResult) {
	x.Data = &DistinctProcessResp_Data{
		OsType: formatRespSlice(conv.SliceToSlice[criteria.OSType, string](
			data.OSType, func(osType criteria.OSType) string { return osType.String() })),
		CpuArch: formatRespSlice(conv.SliceToSlice[criteria.CPUArch, string](
			data.CPUArch, func(cpuArch criteria.CPUArch) string { return cpuArch.String() })),
		Version: formatRespSlice(data.Version),
		Status: formatRespSlice(conv.SliceToSlice[types.ProcessStatus, string](
			data.Status, func(status types.ProcessStatus) string { return status.String() })),
		PluginName:    formatRespSlice(data.PluginName),
		PluginGroup:   formatRespSlice(data.PluginGroup),
		PluginPkgName: formatRespSlice(data.PluginPkgName),
	}
}
