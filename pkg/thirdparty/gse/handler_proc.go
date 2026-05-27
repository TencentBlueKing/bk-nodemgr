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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandlerProc define the gse handler for process.
type IHandlerProc interface {
	// QueryProcessInfo order the gse_agent to trusteeship the process.
	// @param processName given process name.
	// @param agentID given agent id.
	// @return types.ProcessInfo
	QueryProcessInfo(nCtx contextx.IContext, pluginName string, processName string, agentID string) (*types.ProcessInfo, error)

	// QueryMultiProcessInfoMany query multiple process info for many agents.
	// @param procNameAgentIDMap given agent id list and process name mapping.
	// @return map[processName] -> []types.ProcessInfo
	QueryMultiProcessInfoMany(nCtx contextx.IContext, procAgentIDMap ...*types.ProcessAgentGroup) (map[string][]types.ProcessInfo, error)

	// TrusteeshipProcess order the gse_agent to trusteeship the process.
	TrusteeshipProcess(nCtx contextx.IContext, processSpec types.ProcessSpec) (string, error)

	// UnTrusteeshipProcess order the gse_agent to stop trusteeship the process.
	UnTrusteeshipProcess(nCtx contextx.IContext, processSpec types.ProcessSpec) (string, error)

	// TrusteeshipAndStartProcess order the gse_agent to trusteeship and start the process.
	TrusteeshipAndStartProcess(nCtx contextx.IContext, processSpec types.ProcessSpec) (string, error)

	// UnTrusteeshipAndStopProcess order the gse_agent to untrusteeship and stop the process.
	UnTrusteeshipAndStopProcess(nCtx contextx.IContext, processSpec types.ProcessSpec) (string, error)

	// TrusteeshipAndRestartProcess order the gse_agent to trusteeship and restart the process.
	TrusteeshipAndRestartProcess(nCtx contextx.IContext, processSpec types.ProcessSpec) (string, error)

	// TrusteeshipAndReloadProcess order the gse_agent to trusteeship and reload the process.
	TrusteeshipAndReloadProcess(nCtx contextx.IContext, processSpec types.ProcessSpec) (string, error)
}

// HandlerProc adapts GSE process APIs for Node Manager process operations.
type HandlerProc struct {
	cli           *cli
	procNameSpace string
}

const (
	// ProcNameSpace the namespace of proc.
	procNameSpaceNodemgr = "bk-nodemgr"
)

// NewHandlerProc new handler proc.
func (h *Handler) NewHandlerProc(opts ...HandlerProcOptionFn) IHandlerProc {
	hc := &HandlerProc{
		cli:           h.cli,
		procNameSpace: procNameSpaceNodemgr,
	}
	for _, opt := range opts {
		opt(hc)
	}

	return hc
}

// HandlerProcOptionFn configures HandlerProc behavior.
type HandlerProcOptionFn func(*HandlerProc)

// WithProcNameSpace set proc name space.
func WithProcNameSpace(procNameSpace string) HandlerProcOptionFn {
	return func(h *HandlerProc) {
		h.procNameSpace = procNameSpace
	}
}

// QueryProcessInfo order the gse_agent to trusteeship the process
// (trusteeshiping: when the managed process exits abnormally, the agent will automatically pull up the managed process;
// When the managed process resources exceed the limit, the agent will kill the managed process).
func (h *HandlerProc) QueryProcessInfo(
	nCtx contextx.IContext, pluginName string, processName string, agentID string,
) (*types.ProcessInfo, error) {

	operateProcReq := operateProcV2Req{
		Meta: procMeta{
			Namespace: h.procNameSpace,
			Name:      pluginName,
			Labels: procInfoMetaLabels{
				ProcName: pluginName,
			},
		},
		OpType:      procOperateCodeStatus,
		AgentIDList: []string{agentID},
		Spec: procSpec{
			Identity: procSpecIdentity{
				ProcName: processName,
			},
		},
	}

	procResult, err := h.operateProc(nCtx, &operateProcReq)
	if err != nil {
		return nil, fmt.Errorf("failed to operate proc: %w", err)
	}

	procInfoMap, err := h.parseQueryProcResult(procResult)
	if err != nil {
		return nil, fmt.Errorf("failed to query proc: %w", err)
	}

	if _, ok := procInfoMap[agentID]; !ok {
		return nil, fmt.Errorf("failed to query proc: agentID not found")
	}

	info := &types.ProcessInfo{
		AutoStart: procInfoMap[agentID].IsAuto,
		Pid:       procInfoMap[agentID].Pid,
		AgentID:   agentID,
		Version:   strings.TrimSpace(procInfoMap[agentID].Version),
		Status:    convPidToProcStatus(procInfoMap[agentID].Pid),
	}

	return info, nil
}

func (h *HandlerProc) parseQueryProcResult(operateProcResultResp getProcOperateResultV2Resp) (map[string]processInfo, error) {
	procInfoMap := make(map[string]processInfo)
	for key, item := range operateProcResultResp {
		// notice: this key is formated as: agentID:namespace:procName
		keys := strings.Split(key, ":")
		agentID := keys[0]

		result := queryProcessContent{}
		err := json.Unmarshal([]byte(item.Content), &result)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal operate proc result: %w", err)
		}

		if len(result.Process) != 1 {
			return nil, fmt.Errorf("failed to parse operate proc result: this result process has invalid length: %d", len(result.Process))
		}

		if len(result.Process[0].Instance) == 0 {
			return nil, fmt.Errorf("failed to parse operate proc result: this result process instance has invalid length: %d",
				len(result.Process[0].Instance))
		}

		// notice: this is can be sure that the length of the result is 1.
		procInfoMap[agentID] = result.Process[0].Instance[0]
	}

	return procInfoMap, nil
}

// QueryMultiProcessInfoMany query multiple process info for many agents.
//
// Name mapping in GSE:
//   - meta.name / labels.procName: nodemgr pluginName.
//   - spec.identity.procName: programName, the real OS process name from process identity.
//
// The returned map is keyed by meta.name, so callers can use the key as pluginName.
func (h *HandlerProc) QueryMultiProcessInfoMany(
	nCtx contextx.IContext, procNameAgentIDMap ...*types.ProcessAgentGroup) (map[string][]types.ProcessInfo, error) {

	if len(procNameAgentIDMap) == 0 {
		return make(map[string][]types.ProcessInfo), nil
	}

	processInfoMap := make(map[string][]types.ProcessInfo)
	for _, item := range procNameAgentIDMap {
		for start := 0; start < len(item.AgentIDList); start += queryMultiProcessInfoPageSize {
			end := min(start+queryMultiProcessInfoPageSize, len(item.AgentIDList))
			batchProcessInfoMap, err := h.queryMultiProcessInfoMany(nCtx, &types.ProcessAgentGroup{
				PluginName:  item.PluginName,
				ProcessName: item.ProcessName,
				AgentIDList: item.AgentIDList[start:end],
			})
			if err != nil {
				return nil, err
			}

			for pluginName, infos := range batchProcessInfoMap {
				processInfoMap[pluginName] = append(processInfoMap[pluginName], infos...)
			}
		}
	}

	return processInfoMap, nil
}

func (h *HandlerProc) queryMultiProcessInfoMany(
	nCtx contextx.IContext, procNameAgentIDMap ...*types.ProcessAgentGroup) (map[string][]types.ProcessInfo, error) {

	if len(procNameAgentIDMap) == 0 {
		return make(map[string][]types.ProcessInfo), nil
	}

	operateProcReqs := make([]*procOperateReq, 0, len(procNameAgentIDMap))
	for _, item := range procNameAgentIDMap {
		operateProcReqs = append(operateProcReqs, &procOperateReq{
			Meta: procMeta{
				Namespace: h.procNameSpace,
				Name:      item.PluginName,
				Labels: procInfoMetaLabels{
					ProcName: item.PluginName,
				},
			},
			OpType:      procOperateCodeStatus,
			AgentIDList: item.AgentIDList,
			Spec: procSpec{
				Identity: procSpecIdentity{
					ProcName: item.ProcessName,
				},
			},
		})
	}

	procResult, err := h.operateProcMulti(nCtx, &operateProcMultiReq{
		ProcOperateReq: operateProcReqs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to operate proc multi: %w", err)
	}

	procInfoMap, err := h.parseQueryMultiProcResult(procResult)
	if err != nil {
		return nil, fmt.Errorf("failed to query proc: %w", err)
	}

	processInfoMap := make(map[string][]types.ProcessInfo)
	for agentID, infos := range procInfoMap {
		for _, info := range infos {
			// For this GSE API, info.ProcessName carries meta.name/pluginName,
			// not spec.identity.procName/programName.
			processInfoMap[info.ProcessName] = append(processInfoMap[info.ProcessName], types.ProcessInfo{
				AutoStart: info.IsAuto,
				AgentID:   agentID,
				Pid:       info.Pid,
				Version:   strings.TrimSpace(info.Version),
				Status:    convPidToProcStatus(info.Pid),
			})
		}
	}

	return processInfoMap, nil
}

func (h *HandlerProc) parseQueryMultiProcResult(operateProcResultResp getProcOperateResultV2Resp) (map[string][]processInfo, error) {
	procInfoMap := make(map[string][]processInfo)
	for key, item := range operateProcResultResp {
		// notice: this key is formated as: agentID:namespace:procName
		keys := strings.Split(key, ":")
		agentID := keys[0]

		result := queryProcessContent{}
		err := json.Unmarshal([]byte(item.Content), &result)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal operate proc result: %w", err)
		}

		if len(result.Process) != 1 {
			return nil, fmt.Errorf("failed to parse operate proc result: this result process has invalid length: %d", len(result.Process))
		}

		if len(result.Process[0].Instance) == 0 {
			return nil, fmt.Errorf("failed to parse operate proc result: this result process instance has invalid length: %d",
				len(result.Process[0].Instance))
		}

		if _, ok := procInfoMap[agentID]; !ok {
			procInfoMap[agentID] = make([]processInfo, 0)
		}

		procInfoMap[agentID] = append(procInfoMap[agentID], result.Process[0].Instance...)
	}

	return procInfoMap, nil
}

func convPidToProcStatus(pid int) types.ProcessStatus {
	if pid == -1 {
		return types.ProcessStatusStopped
	}

	if pid > 1 {
		return types.ProcessStatusRunning
	}

	return types.ProcessStatusUnknown
}

// TrusteeshipProcess order the gse_agent to trusteeship the process
// (trusteeship: when the managed process exits abnormally, the agent will automatically pull up the managed process;
// When the managed process resources exceed the limit, the agent will kill the managed process).
func (h *HandlerProc) TrusteeshipProcess(nCtx contextx.IContext, processSpec types.ProcessSpec) (string, error) {
	// The lower level will be based on the value issued by the upper level, so it must be checked here.
	if err := processSpec.Validate(); err != nil {
		return "", fmt.Errorf("failed to trusteeship process: %w", err)
	}

	operateProcReq := operateProcV2Req{
		Meta: procMeta{
			Namespace: h.procNameSpace,
			Name:      processSpec.PluginName,
			Labels: procInfoMetaLabels{
				ProcName: processSpec.PluginName,
			},
		},
		OpType:      procOperateCodeTrusteeship,
		AgentIDList: []string{processSpec.AgentID},
	}

	var err error
	operateProcReq.Spec, err = convProcessSpecFromType(processSpec)
	if err != nil {
		return "", fmt.Errorf("failed to parse process spec: %w", err)
	}

	procResult, err := h.operateProc(nCtx, &operateProcReq)
	if err != nil {
		return "", fmt.Errorf("failed to operate proc: %w", err)
	}

	controlProcResultMap, err := h.parseControlProcResult(procResult)
	if err != nil {
		return "", fmt.Errorf("failed to query proc: %w", err)
	}

	controlProcResult, ok := controlProcResultMap[processSpec.AgentID]
	if !ok {
		return "", fmt.Errorf("failed to parse control proc result: %w", err)
	}

	return controlProcResult.CmdOut, controlProcResult.Err
}

// UnTrusteeshipProcess order the gse_agent to untrusteeship the process.
func (h *HandlerProc) UnTrusteeshipProcess(nCtx contextx.IContext, processSpec types.ProcessSpec) (string, error) {
	// The lower level will be based on the value issued by the upper level, so it must be checked here.
	if err := processSpec.Validate(); err != nil {
		return "", fmt.Errorf("failed to untrusteeship process: %w", err)
	}

	operateProcReq := operateProcV2Req{
		Meta: procMeta{
			Namespace: h.procNameSpace,
			Name:      processSpec.PluginName,
			Labels: procInfoMetaLabels{
				ProcName: processSpec.PluginName,
			},
		},
		OpType:      procOperateCodeUnTrusteeship,
		AgentIDList: []string{processSpec.AgentID},
	}

	var err error
	operateProcReq.Spec, err = convProcessSpecFromType(processSpec)
	if err != nil {
		return "", fmt.Errorf("failed to parse process spec: %w", err)
	}

	procResult, err := h.operateProc(nCtx, &operateProcReq)
	if err != nil {
		return "", fmt.Errorf("failed to operate proc: %w", err)
	}

	controlProcResultMap, err := h.parseControlProcResult(procResult)
	if err != nil {
		return "", fmt.Errorf("failed to query proc: %w", err)
	}

	controlProcResult, ok := controlProcResultMap[processSpec.AgentID]
	if !ok {
		return "", fmt.Errorf("failed to parse control proc result: %w", err)
	}

	return controlProcResult.CmdOut, controlProcResult.Err
}

// TrusteeshipAndStartProcess starts the process.
func (h *HandlerProc) TrusteeshipAndStartProcess(nCtx contextx.IContext, processSpec types.ProcessSpec) (string, error) {
	// The lower level will be based on the value issued by the upper level, so it must be checked here.
	if err := processSpec.Validate(); err != nil {
		return "", fmt.Errorf("failed to start process: %w", err)
	}

	operateProcReq := operateProcV2Req{
		Meta: procMeta{
			Namespace: h.procNameSpace,
			Name:      processSpec.PluginName,
			Labels: procInfoMetaLabels{
				ProcName: processSpec.PluginName,
			},
		},
		OpType:      procOperateCodeStart,
		AgentIDList: []string{processSpec.AgentID},
	}

	var err error
	operateProcReq.Spec, err = convProcessSpecFromType(processSpec)
	if err != nil {
		return "", fmt.Errorf("failed to parse process spec: %w", err)
	}

	procResult, err := h.operateProc(nCtx, &operateProcReq)
	if err != nil {
		return "", fmt.Errorf("failed to operate proc: %w", err)
	}

	controlProcResultMap, err := h.parseControlProcResult(procResult)
	if err != nil {
		return "", fmt.Errorf("failed to query proc: %w", err)
	}

	controlProcResult, ok := controlProcResultMap[processSpec.AgentID]
	if !ok {
		return "", fmt.Errorf("failed to parse control proc result: %w", err)
	}

	if controlProcResult.Code == procOperateResultCodeProcAlreadyRunning {
		return controlProcResult.CmdOut, nil
	}

	return controlProcResult.CmdOut, controlProcResult.Err
}

// UnTrusteeshipAndStopProcess stops the process.
func (h *HandlerProc) UnTrusteeshipAndStopProcess(nCtx contextx.IContext, processSpec types.ProcessSpec) (string, error) {
	// The lower level will be based on the value issued by the upper level, so it must be checked here.
	if err := processSpec.Validate(); err != nil {
		return "", fmt.Errorf("failed to stop process: %w", err)
	}

	operateProcReq := operateProcV2Req{
		Meta: procMeta{
			Namespace: h.procNameSpace,
			Name:      processSpec.PluginName,
			Labels: procInfoMetaLabels{
				ProcName: processSpec.PluginName,
			},
		},
		OpType:      procOperateCodeStop,
		AgentIDList: []string{processSpec.AgentID},
	}

	var err error
	operateProcReq.Spec, err = convProcessSpecFromType(processSpec)
	if err != nil {
		return "", fmt.Errorf("failed to parse process spec: %w", err)
	}

	procResult, err := h.operateProc(nCtx, &operateProcReq)
	if err != nil {
		return "", fmt.Errorf("failed to operate proc: %w", err)
	}

	controlProcResultMap, err := h.parseControlProcResult(procResult)
	if err != nil {
		return "", fmt.Errorf("failed to query proc: %w", err)
	}

	controlProcResult, ok := controlProcResultMap[processSpec.AgentID]
	if !ok {
		return "", fmt.Errorf("failed to parse control proc result: %w", err)
	}

	if controlProcResult.Code == procOperateResultCodeProcNotRunning {
		return controlProcResult.CmdOut, nil
	}

	return controlProcResult.CmdOut, controlProcResult.Err
}

// TrusteeshipAndRestartProcess restarts the process.
func (h *HandlerProc) TrusteeshipAndRestartProcess(nCtx contextx.IContext, processSpec types.ProcessSpec) (string, error) {
	// The lower level will be based on the value issued by the upper level, so it must be checked here.
	if err := processSpec.Validate(); err != nil {
		return "", fmt.Errorf("failed to restart process: %w", err)
	}

	operateProcReq := operateProcV2Req{
		Meta: procMeta{
			Namespace: h.procNameSpace,
			Name:      processSpec.PluginName,
			Labels: procInfoMetaLabels{
				ProcName: processSpec.PluginName,
			},
		},
		OpType:      procOperateCodeRestart,
		AgentIDList: []string{processSpec.AgentID},
	}

	var err error
	operateProcReq.Spec, err = convProcessSpecFromType(processSpec)
	if err != nil {
		return "", fmt.Errorf("failed to parse process spec: %w", err)
	}

	procResult, err := h.operateProc(nCtx, &operateProcReq)
	if err != nil {
		return "", fmt.Errorf("failed to operate proc: %w", err)
	}

	controlProcResultMap, err := h.parseControlProcResult(procResult)
	if err != nil {
		return "", fmt.Errorf("failed to query proc: %w", err)
	}

	controlProcResult, ok := controlProcResultMap[processSpec.AgentID]
	if !ok {
		return "", fmt.Errorf("failed to parse control proc result: %w", err)
	}

	return controlProcResult.CmdOut, controlProcResult.Err
}

// TrusteeshipAndReloadProcess reload the process.
func (h *HandlerProc) TrusteeshipAndReloadProcess(nCtx contextx.IContext, processSpec types.ProcessSpec) (string, error) {
	// The lower level will be based on the value issued by the upper level, so it must be checked here.
	if err := processSpec.Validate(); err != nil {
		return "", fmt.Errorf("failed to reload process: %w", err)
	}

	operateProcReq := operateProcV2Req{
		Meta: procMeta{
			Namespace: h.procNameSpace,
			Name:      processSpec.PluginName,
			Labels: procInfoMetaLabels{
				ProcName: processSpec.PluginName,
			},
		},
		OpType:      procOperateCodeReload,
		AgentIDList: []string{processSpec.AgentID},
	}

	var err error
	operateProcReq.Spec, err = convProcessSpecFromType(processSpec)
	if err != nil {
		return "", fmt.Errorf("failed to parse process spec: %w", err)
	}

	procResult, err := h.operateProc(nCtx, &operateProcReq)
	if err != nil {
		return "", fmt.Errorf("failed to operate proc: %w", err)
	}

	controlProcResultMap, err := h.parseControlProcResult(procResult)
	if err != nil {
		return "", fmt.Errorf("failed to query proc: %w", err)
	}

	controlProcResult, ok := controlProcResultMap[processSpec.AgentID]
	if !ok {
		return "", fmt.Errorf("failed to parse control proc result: %w", err)
	}

	return controlProcResult.CmdOut, controlProcResult.Err
}

func convProcessSpecFromType(processSpec types.ProcessSpec) (procSpec, error) {
	spec := procSpec{
		Identity: procSpecIdentity{
			ProcName:   processSpec.Identity.Name,
			SetupPath:  processSpec.Identity.SetupPath,
			PidPath:    processSpec.Identity.PidPath,
			ConfigPath: processSpec.Identity.ConfigPath,
			LogPath:    processSpec.Identity.LogPath,
			User:       processSpec.Identity.User,
		},
		Control: procSpecControl{
			StartCmd:   processSpec.Controller.StartCmd,
			StopCmd:    processSpec.Controller.StopCmd,
			RestartCmd: processSpec.Controller.RestartCmd,
			ReloadCmd:  processSpec.Controller.ReloadCmd,
			KillCmd:    processSpec.Controller.KillCmd,
			VersionCmd: processSpec.Controller.VersionCmd,
			HealthCmd:  processSpec.Controller.HealthCmd,
		},
		Resource: procSpecResource{
			CPU: processSpec.Resource.CPULimitPercent,
			Mem: processSpec.Resource.MemLimitPercent,
		},
		MonitorPolicy: procSpecMonitorPolicy{
			StartCheckSecs: processSpec.MonitorPolicy.StartCheckSecs,
			StopCheckSecs:  processSpec.MonitorPolicy.StopCheckSecs,
			OpTimeoutSecs:  processSpec.MonitorPolicy.OpTimeoutSecs,
		},
	}

	var err error
	spec.MonitorPolicy.AutoType, err = convAutoTypeFromType(processSpec.MonitorPolicy.RestartType)
	if err != nil {
		return spec, fmt.Errorf("failed to parse auto type: %w", err)
	}

	return spec, nil
}

func convAutoTypeFromType(autoType types.ProcessRestartType) (procSpecMonitorPolicyAutoType, error) {
	if err := autoType.Validate(); err != nil {
		return 0, fmt.Errorf("failed to parse auto type: %w", err)
	}

	switch autoType {
	case types.ProcessRestartTypeAuto:
		return procSpecMonitorPolicyAutoTypeTrusteeship, nil
	case types.ProcessRestartTypeManual:
		return procSpecMonitorPolicyAutoTypeOnce, nil
	default:
		return 0, fmt.Errorf("unsupport auto type: %s", autoType)
	}
}

type controlProcessResult struct {
	Code   procOperateResultCode
	Err    error
	CmdOut string
}

func (h *HandlerProc) parseControlProcResult(
	operateProcResultResp getProcOperateResultV2Resp,
) (map[string]controlProcessResult, error) {

	procControlResult := make(map[string]controlProcessResult)
	for key, item := range operateProcResultResp {
		// notice: this key is formated as: agentID:namespace:procName
		keys := strings.Split(key, ":")
		agentID := keys[0]

		content := controlProcessContent{}
		err := json.Unmarshal([]byte(item.Content), &content)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal operate proc content: %w", err)
		}

		if len(content.Value) != 1 {
			return nil, fmt.Errorf("failed to parse operate proc content: this content value has invalid length: %d", len(content.Value))
		}

		var controlErr error
		if item.ErrorCode != procOperateResultCodeOK {
			controlErr = fmt.Errorf("operate proc failed: %s", item.ErrorMsg)
		}

		// notice: this is can be sure that the length of the content is 1.
		procControlResult[agentID] = controlProcessResult{
			Code:   item.ErrorCode,
			Err:    controlErr,
			CmdOut: content.Value[0].Result,
		}
	}

	return procControlResult, nil
}

func (h *HandlerProc) operateProc(nCtx contextx.IContext, operateProcReq *operateProcV2Req) (getProcOperateResultV2Resp, error) {
	operateProcResp, err := h.cli.operateProcV2(nCtx, operateProcReq)
	if err != nil {
		return nil, fmt.Errorf("failed to operate proc: %w", err)
	}

	return h.queryOperateProcResult(nCtx, operateProcResp.TaskID)
}

func (h *HandlerProc) operateProcMulti(nCtx contextx.IContext, operateProcReq *operateProcMultiReq) (getProcOperateResultV2Resp, error) {
	operateProcMultiResp, err := h.cli.operateProcMulti(nCtx, operateProcReq)
	if err != nil {
		return nil, fmt.Errorf("failed to operate proc: %w", err)
	}

	return h.queryOperateProcResult(nCtx, operateProcMultiResp.TaskID)
}

func (h *HandlerProc) queryOperateProcResult(nCtx contextx.IContext, taskID string) (getProcOperateResultV2Resp, error) {
	operateProcResultReq := getProcOperateResultV2Req{
		TaskID: taskID,
	}

	var (
		operateProcResultResp getProcOperateResultV2Resp
		err                   error
	)
	// nolint: mnd
	expoBackoffOpts := retrier.ExpoBackoffOpts{
		MaxRetries:    5,
		BaseDelay:     time.Second,
		MaxDelay:      5 * time.Second,
		JitterPercent: 0.2,
	}
	expoBackoff := retrier.NewExpoBackoff(expoBackoffOpts)
	err = expoBackoff.Do(nCtx, func(_ int) error {
		operateProcResultResp, err = h.cli.getProcOperateResultV2(nCtx, &operateProcResultReq)
		if err != nil {
			return fmt.Errorf("failed to get operate proc result: %w", err)
		}

		for _, item := range operateProcResultResp {
			switch item.ErrorCode {
			case procOperateResultCodeOK:
				continue
			case procOperateResultCodeRunning:
				// continue to retry.
				return fmt.Errorf("proc operate task is running, taskID(%s)", taskID)
			default:
				// other code is error, but that means that the task is finished.
				continue
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get operate proc result: %w", err)
	}

	return operateProcResultResp, nil
}
