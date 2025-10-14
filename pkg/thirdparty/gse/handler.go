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
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/google/uuid"
)

// IHandler is the interface for gse Handler.
type IHandler interface {
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
		osType, _ := platform.NormalizeOS(info.BKOSType)
		arch, _ := platform.NormalizeArch(info.BKCPUArch)

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
		TimeoutSec:    uint(opts.Timeout.Seconds()),
		AutoMkdir:     opts.AutoMkdir,
		UploadSpeed:   opts.UploadSpeedMBPerSec,
		DownloadSpeed: opts.DownloadSpeedMBPerSec,
		Tasks:         tasks,
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
