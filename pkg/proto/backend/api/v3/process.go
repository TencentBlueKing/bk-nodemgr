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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
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
func (x *ProcessListReq) ConvertPageToTypes(maxLimit int) (types.Page, error) {
	return convPageToTypes(x.GetPage(), maxLimit)
}

// ConvertConditionFromTypes converts condition from types.
func (x *ProcessListReq) ConvertConditionFromTypes(condition *types.ProcessCondition) {
	if condition == nil {
		return
	}

	// exact conditions.
	if condition.ExactInclude != nil {
		x.ExactIncludeConditions = convExactIncludeConditionsFromTypes(condition)
	}

	// fuzzy conditions.
	if condition.FuzzyInclude != nil {
		x.FuzzyIncludeConditions = convFuzzyConditionsFromTypes(condition)
	}
}

func convFuzzyConditionsFromTypes(condition *types.ProcessCondition) *ProcessFuzzyConditions {
	return &ProcessFuzzyConditions{
		Name:          condition.FuzzyInclude.Name,
		PluginPkgName: condition.FuzzyInclude.PkgName,
	}
}

func convExactIncludeConditionsFromTypes(condition *types.ProcessCondition) *ProcessExactConditions {
	return &ProcessExactConditions{
		BkHostId:       condition.ExactInclude.HostID,
		PluginGroup:    condition.ExactInclude.PluginGroup,
		NodeGeneration: condition.ExactInclude.NodeGeneration,
		PlatformOs:     condition.ExactInclude.PlatformOS,
		PlatformArch:   condition.ExactInclude.PlatformArch,
		Status: conv.SliceToSlice[types.ProcessStatus, string](condition.ExactInclude.InfoStatus, func(status types.ProcessStatus) string {
			return string(status)
		}),
		AgentId:       condition.ExactInclude.InfoAgentID,
		Version:       condition.ExactInclude.InfoVersion,
		PluginName:    condition.ExactInclude.PluginName,
		PluginPkgName: condition.ExactInclude.PluginPkgName,
	}
}

// ConvertProcessFromTypes converts process from types.
func (x *ProcessListResp) ConvertProcessFromTypes(total int64, process []*types.Process) {
	items := make([]*ProcessListResp_Process, len(process))
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

func newEmptyProcess() *ProcessListResp_Process {
	return &ProcessListResp_Process{
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
				Pid:       int(proc.GetProcessInfo().GetPid()),
				Version:   proc.GetProcessInfo().GetVersion(),
				AgentID:   proc.GetProcessInfo().GetAgentId(),
				AutoStart: proc.GetProcessInfo().GetAutoStart(),
				Status:    types.ProcessStatus(proc.GetProcessInfo().GetStatus()),
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

// ConvertConditionFromTypes convert conditions.
func (x *GetProcessDistributionByHostIDReq) ConvertConditionFromTypes(condition *types.ProcessCondition) {
	if condition == nil {
		return
	}

	// exact conditions.
	if condition.ExactInclude != nil {
		x.ExactIncludeConditions = convExactIncludeConditionsFromTypes(condition)
	}

	// fuzzy conditions.
	if condition.FuzzyInclude != nil {
		x.FuzzyIncludeConditions = convFuzzyConditionsFromTypes(condition)
	}
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

// ConvertConditionFromTypes convert conditions.
func (x *GetProcessDistributionByPluginNameReq) ConvertConditionFromTypes(condition *types.ProcessCondition) {
	if condition == nil {
		return
	}

	// exact conditions.
	if condition.ExactInclude != nil {
		x.ExactIncludeConditions = convExactIncludeConditionsFromTypes(condition)
	}

	// fuzzy conditions.
	if condition.FuzzyInclude != nil {
		x.FuzzyIncludeConditions = convFuzzyConditionsFromTypes(condition)
	}
}

// ===============================================================================
// ProcessDistinct
// ===============================================================================

// Validate check body.
func (x *ProcessDistinctReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *ProcessDistinctReq) AutoConvert() {}

// ConvertSelectorToTypes convert selector.
func (x *ProcessDistinctReq) ConvertSelectorToTypes() types.ProcessDistinctSelector {
	if x.GetSelector() == nil {
		return types.ProcessDistinctSelector{}
	}

	selector := x.GetSelector()

	return types.ProcessDistinctSelector{
		OSType:        selector.GetOsType(),
		CPUArch:       selector.GetCpuArch(),
		Version:       selector.GetVersion(),
		Status:        selector.GetStatus(),
		PluginName:    selector.GetPluginName(),
		PluginGroup:   selector.GetPluginGroup(),
		PluginPkgName: selector.GetPluginPkgName(),
	}
}

// ConvertSelectorFromTypes convert selector.
func (x *ProcessDistinctReq) ConvertSelectorFromTypes(selector types.ProcessDistinctSelector) {
	x.Selector = &ProcessDistinctSelector{
		OsType:        selector.OSType,
		CpuArch:       selector.CPUArch,
		Version:       selector.Version,
		Status:        selector.Status,
		PluginName:    selector.PluginName,
		PluginGroup:   selector.PluginGroup,
		PluginPkgName: selector.PluginPkgName,
	}
}

// ConvertConditionsToTypes convert conditions.
func (x *ProcessDistinctReq) ConvertConditionsToTypes() *types.ProcessCondition {
	return convertProcessConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions())
}

// ConvertConditionFromTypes convert conditions.
func (x *ProcessDistinctReq) ConvertConditionFromTypes(condition *types.ProcessCondition) {
	if condition == nil {
		return
	}

	exactCond, fuzzyCond := convertProcessConditionsFromTypes(condition)

	x.ExactIncludeConditions = exactCond
	x.FuzzyIncludeConditions = fuzzyCond

	return
}

func convertProcessConditionsFromTypes(condition *types.ProcessCondition) (*ProcessExactConditions, *ProcessFuzzyConditions) {
	if condition == nil {
		return nil, nil
	}

	exactCond := &ProcessExactConditions{}
	fuzzyCond := &ProcessFuzzyConditions{}

	if condition.ExactInclude != nil {
		exactCond = convExactIncludeConditionsFromTypes(condition)
	}

	if condition.FuzzyInclude != nil {
		fuzzyCond = convFuzzyConditionsFromTypes(condition)
	}

	return exactCond, fuzzyCond
}

// ConvertResultFromTypes convert result.
func (x *ProcessDistinctResp) ConvertResultFromTypes(result *types.ProcessDistinctResult) {
	x.Data = &ProcessDistinctResp_Data{
		OsType: formatRespSlice(conv.SliceToSlice[criteria.OSType, string](
			result.OSType, func(osType criteria.OSType) string { return osType.String() })),
		CpuArch: formatRespSlice(conv.SliceToSlice[criteria.CPUArch, string](
			result.CPUArch, func(cpuArch criteria.CPUArch) string { return cpuArch.String() })),
		Version: formatRespSlice(result.Version),
		Status: formatRespSlice(conv.SliceToSlice[types.ProcessStatus, string](
			result.Status, func(status types.ProcessStatus) string { return status.String() })),
		PluginName:    formatRespSlice(result.PluginName),
		PluginGroup:   formatRespSlice(result.PluginGroup),
		PluginPkgName: formatRespSlice(result.PluginPkgName),
	}
}

// ConvertResultToTypes convert result to types.
func (x *ProcessDistinctResp) ConvertResultToTypes() (*types.ProcessDistinctResult, error) {
	if x.GetData() == nil {
		return nil, fmt.Errorf("distinct process failed, get empty data")
	}

	var err error
	result := new(types.ProcessDistinctResult)
	result.OSType, err = conv.SliceToSliceWithError[string, criteria.OSType](
		x.GetData().GetOsType(),
		func(s string) (criteria.OSType, error) {
			osType := criteria.OSType(s)
			if err := osType.Validate(); err != nil {
				return "", err
			}

			return osType, nil
		})
	if err != nil {
		return nil, fmt.Errorf("failed to convert os type: %w", err)
	}
	result.CPUArch, err = conv.SliceToSliceWithError[string, criteria.CPUArch](
		x.GetData().GetCpuArch(),
		func(s string) (criteria.CPUArch, error) {
			cpuArch := criteria.CPUArch(s)
			if err := cpuArch.Validate(); err != nil {
				return "", err
			}

			return cpuArch, nil
		})
	if err != nil {
		return nil, fmt.Errorf("failed to convert cpu arch: %w", err)
	}
	result.Version = x.GetData().GetVersion()
	result.Status, err = conv.SliceToSliceWithError[string, types.ProcessStatus](
		x.GetData().GetStatus(),
		func(s string) (types.ProcessStatus, error) {
			processStatus := types.ProcessStatus(s)
			if err := processStatus.Validate(); err != nil {
				return "", err
			}

			return processStatus, nil
		})
	if err != nil {
		return nil, fmt.Errorf("failed to convert status: %w", err)
	}
	result.PluginName = x.GetData().GetPluginName()
	result.PluginGroup = x.GetData().GetPluginGroup()
	result.PluginPkgName = x.GetData().GetPluginPkgName()

	return result, nil
}
