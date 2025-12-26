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
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/google/uuid"
)

// IHandler is the interface for gse Handler.
type IHandler interface {
	IHandlerNode
	IHandlerProc
}

// IHandlerNode define the gse handler for node.
// nolint: interfacebloat
type IHandlerNode interface {
	// ListAgentInfo list agent detail information.
	// @param agentIDList given agent id list.
	// @return agentInfoList agent detail information list.
	ListAgentInfo(nCtx contextx.IContext, agentIDList ...string) ([]*types.AgentInfo, error)

	// ListAgentState list agent state information. agentState is a subset of agentInfo.
	// This method is more efficient than ListAgentInfo.
	// @param agentIDList given agent id list.
	// @return agentStateList agent state information list.
	ListAgentState(nCtx contextx.IContext, agentIDList ...string) ([]*types.AgentState, error)

	// ExecuteScript execute script on host.
	// @param scriptType given script type.
	// @param scriptContent given script content.
	// @param timeout given timeout.
	// @param endpoints given endpoint list with auth.
	// @return gse-task-id for this execution for further querying.
	ExecuteScript(nCtx contextx.IContext, scriptType types.ScriptType, scriptContent string, timeout time.Duration,
		endpoints ...*types.EndpointWithAuth) (string, error)

	// QueryScriptExecutionResult query script execution result.
	// @param taskID given task id.
	// @param endpoints given endpoint list.
	// @return script result list.
	QueryScriptExecutionResult(nCtx contextx.IContext, taskID string, endpoints ...*types.EndpointWithRestrict) (
		[]*types.ScriptResult, error)

	// TerminateScriptExecution terminate script execution.
	// @param taskID given task id
	// @param endpoints given endpoint list.
	// @return gse-task-id for this operation.
	TerminateScriptExecution(nCtx contextx.IContext, taskID string, endpoints ...*types.Endpoint) (string, error)

	// PushFile push files to target endpoints.
	// @param pushDetail given push file details.
	// @return gse-task-id for this pushing.
	PushFile(nCtx contextx.IContext, pushDetail ...*types.PushFileDetail) (string, error)

	// TransferFile transfer files from source to targets.
	// @param opts given options.
	// @param transfers given transfer details.
	// @return gse-task-id for this transferring.
	TransferFile(nCtx contextx.IContext, opts *types.TransferOptions, transfers ...*types.TransferDetail) (string, error)

	// QueryFileTransmissionResult query file transmission result.
	// @param taskID given task id.
	// @param endpoints given endpoint list.
	// @return file transmission result list.
	QueryFileTransmissionResult(nCtx contextx.IContext, taskID string, endpoints ...*types.Endpoint) (
		[]*types.TransferResult, error)

	// QueryPushFileFinalResult query push file final result.
	// @param taskID given task id.
	// @return  file transmission simple transfer result.
	QueryPushFileFinalResult(nCtx contextx.IContext, taskID string) (*types.SimpleTransferResult, error)

	// TerminateFileTransmission terminate file transmission.
	// @param taskID given task id.
	// @param endpoints given endpoint list.
	// @return gse-task-id for this operation.
	TerminateFileTransmission(nCtx contextx.IContext, taskID string, endpoints ...*types.Endpoint) (string, error)

	// OperateAgent operate agent.
	// @param operate given operate.
	// @param agentIDList given agent id list.
	// @return agent operate result.
	OperateAgent(nCtx contextx.IContext, operate types.OperateAgent, agentIDList ...string) (
		*types.OperateAgentResult, error)
}

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

// Handler this define the gse handler.
type Handler struct {
	cli *cli
}

// New initialize a new gse Handler.
func New(c *restclient.Capability, conf *Config) (*Handler, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	return &Handler{cli: cli}, nil
}

// ListAgentInfo list agent detail information.
func (h *Handler) ListAgentInfo(nCtx contextx.IContext, agentIDList ...string) ([]*types.AgentInfo, error) {
	req := ListAgentInfoReq{
		AgentIDList: agentIDList,
	}

	resp, err := h.cli.listAgentInfo(nCtx, &req)
	if err != nil {
		return nil, err
	}
	data := make([]*types.AgentInfo, len(resp))
	for idx, info := range resp {
		osType, _ := platfmt.NormalizeOS(info.BKOSType)
		arch, _ := platfmt.NormalizeArch(info.BKCPUArch)

		data[idx] = &types.AgentInfo{
			AgentState: types.AgentState{
				AgentID:        info.BKAgentID,
				CloudID:        info.BKCloudID,
				Version:        info.Version,
				NodeRole:       convRunModeToNodeRole(info.RunMode),
				NodeGeneration: detectGeneration(info.BKAgentID),
				NodeStatus:     info.StatusCode.ToNodeStatus(),
				ReportTime:     info.ReportTime,
			},
			OSType:         osType,
			Arch:           arch,
			ParentIP:       info.ParentIP,
			ParentPort:     info.ParentPort,
			CPURate:        info.CPURate,
			MemRate:        info.MemRate,
			CPUNum:         info.CPUNum,
			MemSize:        info.MemSize,
			StartTime:      info.StartTime,
			LastWorkTime:   info.LastWorkTime,
			ConnCycleTime:  info.ConnCycleTime,
			LastNodeStatus: info.LastStatusCode.ToNodeStatus(),
			Remark:         info.Remark,
		}
	}

	return data, nil
}

// notice: this is a special logic for generation.
func detectGeneration(agentID string) types.Generation {
	if strings.ContainsRune(agentID, ':') {
		return types.Generation1
	}

	return types.Generation2
}

// ListAgentState list agent state.
func (h *Handler) ListAgentState(nCtx contextx.IContext, agentIDList ...string) ([]*types.AgentState, error) {
	if nCtx == nil {
		return nil, errors.New("context is nil")
	}

	if len(agentIDList) == 0 {
		return nil, errors.New("agent id list is empty")
	}

	req := &ListAgentStateReq{
		AgentIDList: agentIDList,
	}
	resp, err := h.cli.listAgentState(nCtx, req)
	if err != nil {
		return nil, err
	}

	data := make([]*types.AgentState, len(resp))
	for idx, info := range resp {
		data[idx] = &types.AgentState{
			AgentID:        info.BKAgentID,
			CloudID:        info.BKCloudID,
			Version:        info.Version,
			NodeRole:       convRunModeToNodeRole(info.RunMode),
			NodeStatus:     info.StatusCode.ToNodeStatus(),
			NodeGeneration: detectGeneration(info.BKAgentID),
			ReportTime:     info.ReportTime,
		}
	}

	return data, nil
}

func convRunModeToNodeRole(runMode runMode) types.NodeRole {
	switch runMode {
	case RunModeAgent:
		return types.NodeRoleAgent
	case RunModeProxy:
		return types.NodeRoleProxy
	default:
		return types.NodeRoleBlank
	}
}

// ExecuteScript execute script.
func (h *Handler) ExecuteScript(nCtx contextx.IContext,
	scriptType types.ScriptType,
	scriptContent string,
	timeout time.Duration,
	endpoints ...*types.EndpointWithAuth) (string, error) {

	if nCtx == nil {
		return "", errors.New("context is nil")
	}

	if len(endpoints) == 0 {
		return "", errors.New("endpoints is empty")
	}

	eps := make([]*EndpointWithAuth, len(endpoints))
	for idx, endpoint := range endpoints {
		eps[idx] = &EndpointWithAuth{
			Endpoint: Endpoint{
				BKAgentID:     endpoint.AgentID,
				BKContainerID: endpoint.ContainerID,
			},
			User:     endpoint.User,
			Password: endpoint.Password,
		}
	}

	scriptExt, err := getScriptExt(scriptType)
	if err != nil {
		return "", err
	}

	scriptName := fmt.Sprintf("bk_gse_script_nodemgr_%s.%s", uuid.New().String(), scriptExt)
	storedDir := "/tmp/bknodemgr/"
	req := &AsyncExecuteScriptReq{
		Endpoints: eps,
		Scripts: []*ScriptDetail{
			{
				Name:      scriptName,
				StoredDir: storedDir,
				Content:   scriptContent,
			},
		},
		Atomics: []*ScriptAtomicTask{
			{
				ID:         0,
				Command:    filepath.Join(storedDir, scriptName),
				TimeoutSec: int(timeout.Seconds()),
			},
		},
		Relations: make([]*ScriptAtomicRelation, 0),
	}
	resp, err := h.cli.asyncExecuteScript(nCtx, req)
	if err != nil {
		return "", err
	}

	return resp.Result.TaskID, nil
}

// QueryScriptExecutionResult query script execution result.
func (h *Handler) QueryScriptExecutionResult(nCtx contextx.IContext, taskID string,
	endpoints ...*types.EndpointWithRestrict) ([]*types.ScriptResult, error) {

	if nCtx == nil {
		return nil, errors.New("context is nil")
	}

	if len(endpoints) == 0 {
		return nil, errors.New("endpoints is empty")
	}

	conditions := make([]*ScriptEndpointCondition, len(endpoints))
	for idx, endpoint := range endpoints {
		conditions[idx] = &ScriptEndpointCondition{
			Endpoint: Endpoint{
				BKAgentID:     endpoint.AgentID,
				BKContainerID: endpoint.ContainerID,
			},
			Atomics: []*ScriptAtomicTaskCondition{
				{
					ID:     0,
					Offset: endpoint.Offset,
					Limit:  endpoint.Limit,
				},
			},
		}
	}

	req := &GetExecuteScriptResultReq{
		TaskID:     taskID,
		AgentTasks: conditions,
	}
	resp, err := h.cli.getExecuteScriptResult(nCtx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.ScriptResult, len(resp.Result))
	for idx, rst := range resp.Result {
		result[idx] = &types.ScriptResult{
			Endpoint: types.Endpoint{
				AgentID:     rst.BKAgentID,
				ContainerID: rst.BKContainerID,
			},
			Status:       rst.Status.ToScriptStatus(),
			ErrorCode:    rst.ErrorCode,
			ErrorMessage: rst.ErrorMessage,
			StartTime:    time.UnixMilli(rst.StartTime),
			EndTime:      time.UnixMilli(rst.EndTime),
			ExitCode:     rst.ExitCode,
			ScreenLog:    rst.ScreenLog,
		}
	}

	return result, nil
}

// TerminateScriptExecution terminate script execution.
func (h *Handler) TerminateScriptExecution(nCtx contextx.IContext, taskID string,
	endpoints ...*types.Endpoint) (string, error) {

	if nCtx == nil {
		return "", errors.New("context is nil")
	}

	if len(endpoints) == 0 {
		return "", errors.New("endpoints is empty")
	}

	eps := make([]*Endpoint, len(endpoints))
	for idx, endpoint := range endpoints {
		eps[idx] = &Endpoint{
			BKAgentID:     endpoint.AgentID,
			BKContainerID: endpoint.ContainerID,
		}
	}

	req := &AsyncTerminateExecuteScriptReq{
		TaskID:    taskID,
		Endpoints: eps,
	}
	resp, err := h.cli.asyncTerminateExecuteScript(nCtx, req)
	if err != nil {
		return "", err
	}

	return resp.Result.TaskID, nil
}

// PushFile push file.
func (h *Handler) PushFile(nCtx contextx.IContext, pushDetail ...*types.PushFileDetail) (string, error) {
	if nCtx == nil {
		return "", errors.New("context is nil")
	}

	if len(pushDetail) == 0 {
		return "", errors.New("tasks detail is empty")
	}

	tasks := make([]*PushFileTask, len(pushDetail))
	for i, detail := range pushDetail {
		targetEndpoints := make([]*Endpoint, len(detail.Endpoints))
		for j, endpoint := range detail.Endpoints {
			targetEndpoints[j] = &Endpoint{
				BKAgentID:     endpoint.AgentID,
				BKContainerID: endpoint.ContainerID,
			}
		}

		if err := checkPathSafe(detail.StoreDir); err != nil {
			return "", err
		}

		// check file content size limit
		if len(detail.FileContent) > maxPushFileContentSize {
			return "", fmt.Errorf("file content size(%d bytes) exceeds the limit(%d bytes)", len(detail.FileContent), maxPushFileContentSize)
		}

		tasks[i] = &PushFileTask{
			FileName:    detail.FileName,
			StoreDir:    detail.StoreDir,
			FileContent: detail.FileContent,
			Owner:       detail.Owner,
			Endpoints:   targetEndpoints,
		}
	}

	req := &AsyncPushFileReq{
		Tasks: tasks,
	}

	resp, err := h.cli.asyncPushFile(nCtx, req)
	if err != nil {
		return "", err
	}

	return resp.Result.TaskID, nil
}

func checkPathSafe(dirPath string) error {
	if dirPath == "" {
		return errors.New("dirPath is empty")
	}

	cleanPath := filepath.Clean(dirPath)
	// compare the cleaned path with the original path
	if cleanPath != dirPath {
		return fmt.Errorf("dirPath is not a clean path, clean-path(%s), origin-path(%s)", cleanPath, dirPath)
	}

	cleanPath = strings.ToLower(cleanPath)

	if cleanPath == filepath.Clean("c:\\") || cleanPath == "/" {
		return fmt.Errorf("dirPath is dangerous, dirPath(%s)", cleanPath)
	}

	dangerousDirPrefixs := []string{
		"c:\\windows\\",
		"c:\\program files\\",
		"c:\\program files (x86)\\",
		"c:\\programs\\",
		"c:\\recovery\\",
		"/proc/",
		"/sys/",
		"/dev/",
	}

	for _, dangerousDir := range dangerousDirPrefixs {
		if strings.HasPrefix(cleanPath, strings.ToLower(filepath.Clean(dangerousDir))) {
			return fmt.Errorf("dirPath is dangerous, dirPath(%s)", cleanPath)
		}
	}

	return nil
}

// TransferFile transfer file.
func (h *Handler) TransferFile(nCtx contextx.IContext, opts *types.TransferOptions,
	transfers ...*types.TransferDetail) (string, error) {

	if nCtx == nil {
		return "", errors.New("context is nil")
	}

	if len(transfers) == 0 {
		return "", errors.New("transfers is empty")
	}

	tasks := make([]*TransferDetail, len(transfers))
	for i, transfer := range transfers {
		targetEndpoints := make([]*EndpointWithAuth, len(transfer.Target.Endpoints))
		for j, endpoint := range transfer.Target.Endpoints {
			targetEndpoints[j] = &EndpointWithAuth{
				Endpoint: Endpoint{
					BKAgentID:     endpoint.AgentID,
					BKContainerID: endpoint.ContainerID,
				},
				User: endpoint.User,
			}
		}

		tasks[i] = &TransferDetail{
			Source: &TransferSource{
				FileName:  transfer.Source.FileName,
				StoredDir: transfer.Source.StoredDir,
				Endpoint: EndpointWithAuth{
					Endpoint: Endpoint{
						BKAgentID:     transfer.Source.Endpoint.AgentID,
						BKContainerID: transfer.Source.Endpoint.ContainerID,
					},
					User: transfer.Source.Endpoint.User,
				},
			},
			Target: &TransferTarget{
				StoredDir: transfer.Target.StoredDir,
				Endpoints: targetEndpoints,
			},
		}
	}

	req := &AsyncTransferFileReq{
		TimeoutSec:        uint(opts.Timeout.Seconds()),
		AutoMkdir:         opts.AutoMkdir,
		UploadSpeed:       opts.UploadSpeedMBPerSec,
		DownloadSpeed:     opts.DownloadSpeedMBPerSec,
		KeepSourceSeeding: opts.KeepSourceSeeding,
		Tasks:             tasks,
	}
	resp, err := h.cli.asyncTransferFile(nCtx, req)
	if err != nil {
		return "", err
	}

	return resp.Result.TaskID, nil
}

// QueryFileTransmissionResult query file transmission result.
func (h *Handler) QueryFileTransmissionResult(nCtx contextx.IContext, taskID string,
	endpoints ...*types.Endpoint) ([]*types.TransferResult, error) {

	if nCtx == nil {
		return nil, errors.New("context is nil")
	}

	eps := make([]*Endpoint, len(endpoints))
	for idx, endpoint := range endpoints {
		eps[idx] = &Endpoint{
			BKAgentID:     endpoint.AgentID,
			BKContainerID: endpoint.ContainerID,
		}
	}

	req := &GetTransferFileResultReq{
		TaskID:    taskID,
		Endpoints: eps,
	}
	resp, err := h.cli.getTransferFileResult(nCtx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.TransferResult, len(resp.Result))
	for idx, rst := range resp.Result {
		result[idx] = &types.TransferResult{
			Source: types.Endpoint{
				AgentID:     rst.Content.SourceAgentID,
				ContainerID: rst.Content.SourceContainerID,
			},
			Target: types.Endpoint{
				AgentID:     rst.Content.DestinationAgentID,
				ContainerID: rst.Content.DestinationContainerID,
			},
			Mode:           types.TransferMode(rst.Content.Mode),
			Progress:       rst.Content.Progress,
			SpeedKBPerSec:  rst.Content.Speed,
			SizeBytes:      rst.Content.Size,
			SourceDir:      rst.Content.SourceFileDir,
			SourceFileName: rst.Content.SourceFileName,
			TargetDir:      rst.Content.DestFileDir,
			TargetFileName: rst.Content.DestFileName,
			ErrorCode:      rst.ErrorCode,
			ErrorMessage:   rst.ErrorMessage,
			StatusCode:     types.TransferStatus(rst.Content.Status),
			StatusMessage:  rst.Content.StatusInfo,
			StartTime:      time.UnixMilli(rst.Content.StartTime),
			EndTime:        time.UnixMilli(rst.Content.EndTime),
		}
	}

	return result, nil
}

// QueryPushFileFinalResult query push file final result.
// nolint:mnd
func (h *Handler) QueryPushFileFinalResult(nCtx contextx.IContext, taskID string) (*types.SimpleTransferResult, error) {
	var (
		err         error
		queryResult []*types.TransferResult
		dst         *types.SimpleTransferResult
	)

	expoBackoffOpts := retrier.ExpoBackoffOpts{
		MaxRetries:    5,
		BaseDelay:     time.Second,
		MaxDelay:      5 * time.Second,
		JitterPercent: 0.2,
	}
	expoBackoff := retrier.NewExpoBackoff(expoBackoffOpts)
	err = expoBackoff.Do(nCtx, func(_ int) error {
		queryResult, err = h.QueryFileTransmissionResult(nCtx, taskID)
		if err != nil {
			return fmt.Errorf("failed to get operate proc result: %w", err)
		}

		if len(queryResult) != 1 {
			return fmt.Errorf("unexpected task(%s) transfer result count(%d), results(%+v)", taskID, len(queryResult), queryResult)
		}

		dst = types.ConvertTransferResultToSimple(queryResult[0])
		if !dst.Terminated {
			return fmt.Errorf("task(%s) is not terminated yet", taskID)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query transfer results: %w", err)
	}

	return dst, nil
}

// TerminateFileTransmission terminate file transmission.
func (h *Handler) TerminateFileTransmission(nCtx contextx.IContext, taskID string, endpoints ...*types.Endpoint) (
	string, error) {

	if nCtx == nil {
		return "", errors.New("context is nil")
	}

	if len(endpoints) == 0 {
		return "", errors.New("endpoints is empty")
	}

	eps := make([]*Endpoint, len(endpoints))
	for idx, endpoint := range endpoints {
		eps[idx] = &Endpoint{
			BKAgentID:     endpoint.AgentID,
			BKContainerID: endpoint.ContainerID,
		}
	}

	req := &AsyncTerminateTransferFileReq{
		TaskID:    taskID,
		Endpoints: eps,
	}
	resp, err := h.cli.asyncTerminateTransferFile(nCtx, req)
	if err != nil {
		return "", err
	}

	return resp.Result.TaskID, nil
}

// OperateAgent operate agent.
func (h *Handler) OperateAgent(nCtx contextx.IContext, operate types.OperateAgent, agentIDList ...string) (
	*types.OperateAgentResult, error) {

	if nCtx == nil {
		return nil, errors.New("context is nil")
	}

	if len(agentIDList) == 0 {
		return nil, errors.New("agent id list is empty")
	}

	req := &OperateAgentReq{
		CurrentVersion:    operate.CurrentAgentVersion,
		TargetVersionSign: operate.TargetAgentVersionSign,
		Timeout:           int(operate.Timeout.Seconds()),
		Force:             operate.Force,
		Remark:            operate.Remark,
		AgentIDList:       agentIDList,
	}
	switch operate.Type {
	case types.OperateAgentTypeRestart:
		req.Type = operateAgentTypeRestart

	default:
		req.Type = operateAgentTypeUnknown
	}

	resp, err := h.cli.operateAgent(nCtx, req)
	if err != nil {
		return nil, err
	}

	result := &types.OperateAgentResult{
		MissingAgentIDs: resp.MissingAgentIDList,
	}

	return result, nil
}

const (
	// ProcNameSpace the namespace of proc.
	procNameSpace = "bk-nodemgr"
)

// QueryProcessInfo order the gse_agent to trusteeship the process
// (trusteeshiping: when the managed process exits abnormally, the agent will automatically pull up the managed process;
// When the managed process resources exceed the limit, the agent will kill the managed process).
func (h *Handler) QueryProcessInfo(nCtx contextx.IContext, pluginName string, processName string, agentID string) (*types.ProcessInfo, error) {
	operateProcReq := operateProcV2Req{
		Meta: procMeta{
			Namespace: procNameSpace,
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

func (h *Handler) parseQueryProcResult(operateProcResultResp getProcOperateResultV2Resp) (map[string]processInfo, error) {
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
func (h *Handler) QueryMultiProcessInfoMany(
	nCtx contextx.IContext, procNameAgentIDMap ...*types.ProcessAgentGroup) (map[string][]types.ProcessInfo, error) {

	if len(procNameAgentIDMap) == 0 {
		return make(map[string][]types.ProcessInfo), nil
	}

	operateProcReqs := make([]*procOperateReq, 0, len(procNameAgentIDMap))
	for _, item := range procNameAgentIDMap {
		operateProcReqs = append(operateProcReqs, &procOperateReq{
			Meta: procMeta{
				Namespace: procNameSpace,
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

func (h *Handler) parseQueryMultiProcResult(operateProcResultResp getProcOperateResultV2Resp) (map[string][]processInfo, error) {
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
func (h *Handler) TrusteeshipProcess(nCtx contextx.IContext, processSpec types.ProcessSpec) (string, error) {
	// The lower level will be based on the value issued by the upper level, so it must be checked here.
	if err := processSpec.Validate(); err != nil {
		return "", fmt.Errorf("failed to trusteeship process: %w", err)
	}

	operateProcReq := operateProcV2Req{
		Meta: procMeta{
			Namespace: procNameSpace,
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
func (h *Handler) UnTrusteeshipProcess(nCtx contextx.IContext, processSpec types.ProcessSpec) (string, error) {
	// The lower level will be based on the value issued by the upper level, so it must be checked here.
	if err := processSpec.Validate(); err != nil {
		return "", fmt.Errorf("failed to untrusteeship process: %w", err)
	}

	operateProcReq := operateProcV2Req{
		Meta: procMeta{
			Namespace: procNameSpace,
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
func (h *Handler) TrusteeshipAndStartProcess(nCtx contextx.IContext, processSpec types.ProcessSpec) (string, error) {
	// The lower level will be based on the value issued by the upper level, so it must be checked here.
	if err := processSpec.Validate(); err != nil {
		return "", fmt.Errorf("failed to start process: %w", err)
	}

	operateProcReq := operateProcV2Req{
		Meta: procMeta{
			Namespace: procNameSpace,
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

	return controlProcResult.CmdOut, controlProcResult.Err
}

// UnTrusteeshipAndStopProcess stops the process.
func (h *Handler) UnTrusteeshipAndStopProcess(nCtx contextx.IContext, processSpec types.ProcessSpec) (string, error) {
	// The lower level will be based on the value issued by the upper level, so it must be checked here.
	if err := processSpec.Validate(); err != nil {
		return "", fmt.Errorf("failed to stop process: %w", err)
	}

	operateProcReq := operateProcV2Req{
		Meta: procMeta{
			Namespace: procNameSpace,
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

	return controlProcResult.CmdOut, controlProcResult.Err
}

// TrusteeshipAndRestartProcess restarts the process.
func (h *Handler) TrusteeshipAndRestartProcess(nCtx contextx.IContext, processSpec types.ProcessSpec) (string, error) {
	// The lower level will be based on the value issued by the upper level, so it must be checked here.
	if err := processSpec.Validate(); err != nil {
		return "", fmt.Errorf("failed to restart process: %w", err)
	}

	operateProcReq := operateProcV2Req{
		Meta: procMeta{
			Namespace: procNameSpace,
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
func (h *Handler) TrusteeshipAndReloadProcess(nCtx contextx.IContext, processSpec types.ProcessSpec) (string, error) {
	// The lower level will be based on the value issued by the upper level, so it must be checked here.
	if err := processSpec.Validate(); err != nil {
		return "", fmt.Errorf("failed to reload process: %w", err)
	}

	operateProcReq := operateProcV2Req{
		Meta: procMeta{
			Namespace: procNameSpace,
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
	spec.MonitorPolicy.AutoType, err = convAutoTypeFromType(processSpec.MonitorPolicy.AutoType)
	if err != nil {
		return spec, fmt.Errorf("failed to parse auto type: %w", err)
	}

	return spec, nil
}

func convAutoTypeFromType(autoType types.ProcessAutoType) (procSpecMonitorPolicyAutoType, error) {
	if err := autoType.Validate(); err != nil {
		return 0, fmt.Errorf("failed to parse auto type: %w", err)
	}

	switch autoType {
	case types.ProcessAutoTypeTrusteeship:
		return procSpecMonitorPolicyAutoTypeTrusteeship, nil
	case types.ProcessAutoTypeOnce:
		return procSpecMonitorPolicyAutoTypeOnce, nil
	default:
		return 0, fmt.Errorf("unsupport auto type: %s", autoType)
	}
}

type controlProcessResult struct {
	Err    error
	CmdOut string
}

func (h *Handler) parseControlProcResult(operateProcResultResp getProcOperateResultV2Resp) (map[string]controlProcessResult, error) {
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
			Err:    controlErr,
			CmdOut: content.Value[0].Result,
		}
	}

	return procControlResult, nil
}

func (h *Handler) operateProc(nCtx contextx.IContext, operateProcReq *operateProcV2Req) (getProcOperateResultV2Resp, error) {
	operateProcResp, err := h.cli.operateProcV2(nCtx, operateProcReq)
	if err != nil {
		return nil, fmt.Errorf("failed to operate proc: %w", err)
	}

	return h.queryOperateProcResult(nCtx, operateProcResp.TaskID)
}

func (h *Handler) operateProcMulti(nCtx contextx.IContext, operateProcReq *operateProcMultiReq) (getProcOperateResultV2Resp, error) {
	operateProcMultiResp, err := h.cli.operateProcMulti(nCtx, operateProcReq)
	if err != nil {
		return nil, fmt.Errorf("failed to operate proc: %w", err)
	}

	return h.queryOperateProcResult(nCtx, operateProcMultiResp.TaskID)
}

func (h *Handler) queryOperateProcResult(nCtx contextx.IContext, taskID string) (getProcOperateResultV2Resp, error) {
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
