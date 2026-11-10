/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

//nolint:varnamelen // Keep the existing HandlerProc receiver name consistent across its methods.
package gse

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/batchexecutor"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ProcessInfoResult is the terminal query outcome for one plugin on one agent.
// Err means the process state was not confirmed; ProcessInfos must not be used in that case.
type ProcessInfoResult struct {
	PluginName   string
	AgentID      string
	ProcessInfos []types.ProcessInfo
	Err          error
}

type processInfoQueryTask struct {
	taskID  string
	pending map[string]ProcessInfoResult
	lastErr error
}

// QueryMultiProcessInfoManyStream queries process states incrementally without discarding completed results.
// Polling is bounded to one minute, comparable to the synchronous retry window, leaving sync callers time to persist results.
// Each distinct plugin/agent produces exactly one result, including failures and timeouts. Empty responses remain pending.
func (h *HandlerProc) QueryMultiProcessInfoManyStream(
	nCtx contextx.IContext, groups ...*types.ProcessAgentGroup,
) <-chan ProcessInfoResult {

	pending := make(map[string]ProcessInfoResult)
	for _, group := range groups {
		for _, agentID := range group.AgentIDList {
			key := fmt.Sprintf("%s:%s:%s", agentID, h.procNameSpace, group.PluginName)
			pending[key] = ProcessInfoResult{PluginName: group.PluginName, AgentID: agentID}
		}
	}

	// One terminal result per target bounds memory and permits timeout results to be sent even to a slow consumer.
	results := make(chan ProcessInfoResult, len(pending))
	go func() {
		defer close(results)
		queryCtx, cancel := contextx.WithTimeout(nCtx, time.Minute)
		defer cancel()

		tasks := make(chan *processInfoQueryTask)
		go func() {
			defer close(tasks)
			h.submitProcessInfoQueries(queryCtx, groups, pending, tasks, results)
		}()
		h.streamProcessInfoResults(queryCtx, tasks, results)
	}()

	return results
}

func (h *HandlerProc) streamProcessInfoResults(
	nCtx contextx.IContext, incoming <-chan *processInfoQueryTask, results chan<- ProcessInfoResult,
) {

	ticker := time.NewTicker(2 * time.Second) // nolint: mnd
	defer ticker.Stop()
	active := make([]*processInfoQueryTask, 0)
	for incoming != nil || len(active) > 0 {
		select {
		case task, ok := <-incoming:
			if !ok {
				incoming = nil

				continue
			}
			active = append(active, task)
			active = h.pollProcessInfoTasks(nCtx, active, results)
		case <-ticker.C:
			active = h.pollProcessInfoTasks(nCtx, active, results)
		case <-nCtx.Done():
			// Join the submitter before closing results; it also reports targets that were never submitted.
			if incoming != nil {
				for task := range incoming {
					active = append(active, task)
				}
			}
			for _, task := range active {
				task.fail(fmt.Errorf("process info query unfinished, taskID(%s): %w", task.taskID,
					errors.Join(nCtx.Err(), task.lastErr)), results)
			}

			return
		}
	}
}

func (h *HandlerProc) pollProcessInfoTasks(
	nCtx contextx.IContext, tasks []*processInfoQueryTask, results chan<- ProcessInfoResult,
) []*processInfoQueryTask {

	active := tasks[:0]
	for _, task := range tasks {
		resp, err := h.cli.getProcOperateResultV2(nCtx, &getProcOperateResultV2Req{TaskID: task.taskID})
		if err != nil {
			// A transient read failure does not discard results from this or other tasks.
			task.lastErr = err
			active = append(active, task)

			continue
		}
		h.emitProcessInfoResults(task.pending, resp, results)
		if len(task.pending) > 0 {
			active = append(active, task)
		}
	}

	return active
}

func (task *processInfoQueryTask) fail(err error, results chan<- ProcessInfoResult) {
	for _, result := range task.pending {
		result.Err = err
		results <- result
	}
}

func (h *HandlerProc) submitProcessInfoQueries(
	nCtx contextx.IContext, groups []*types.ProcessAgentGroup, pending map[string]ProcessInfoResult,
	tasks chan<- *processInfoQueryTask, results chan<- ProcessInfoResult,
) {

	var submitErr error
	for _, group := range groups {
		err := batchexecutor.Execute(nCtx, group.AgentIDList, func(nCtx contextx.IContext, agentIDs []string) error {
			task, uniqueAgentIDs := h.takeProcessInfoQueryBatch(group.PluginName, agentIDs, pending)
			if len(uniqueAgentIDs) == 0 {
				return nil
			}
			resp, err := h.cli.operateProcMulti(nCtx, &operateProcMultiReq{ProcOperateReq: []*procOperateReq{{
				Meta: procMeta{Namespace: h.procNameSpace, Name: group.PluginName,
					Labels: procInfoMetaLabels{ProcName: group.PluginName}},
				OpType: procOperateCodeStatus, AgentIDList: uniqueAgentIDs,
				Spec: procSpec{Identity: procSpecIdentity{ProcName: group.ProcessName}},
			}}})
			if err != nil {
				task.fail(fmt.Errorf("failed to submit process info query: %w", err), results)

				return nil
			}
			task.taskID = resp.TaskID
			tasks <- task

			return nil
		}, batchexecutor.WithBatchSize(queryMultiProcessInfoPageSize), batchexecutor.WithTimeout(time.Minute))
		if err != nil {
			submitErr = errors.Join(submitErr, err)
		}
	}
	for _, result := range pending {
		result.Err = fmt.Errorf("process info query was not submitted: %w", submitErr)
		results <- result
	}
}

func (h *HandlerProc) takeProcessInfoQueryBatch(
	pluginName string, agentIDs []string, pending map[string]ProcessInfoResult,
) (*processInfoQueryTask, []string) {

	task := &processInfoQueryTask{pending: make(map[string]ProcessInfoResult)}
	uniqueAgentIDs := make([]string, 0, len(agentIDs))
	for _, agentID := range agentIDs {
		key := fmt.Sprintf("%s:%s:%s", agentID, h.procNameSpace, pluginName)
		result, ok := pending[key]
		if !ok {
			continue
		}
		delete(pending, key)
		task.pending[key] = result
		uniqueAgentIDs = append(uniqueAgentIDs, agentID)
	}

	return task, uniqueAgentIDs
}

func (h *HandlerProc) emitProcessInfoResults(
	pending map[string]ProcessInfoResult, resp getProcOperateResultV2Resp, results chan<- ProcessInfoResult,
) {

	for key, item := range resp {
		result, ok := pending[key]
		if !ok || item.ErrorCode == procOperateResultCodeRunning {
			continue
		}
		delete(pending, key)
		if item.ErrorCode != procOperateResultCodeOK {
			result.Err = fmt.Errorf("process info query failed, code(%d): %s", item.ErrorCode, item.ErrorMsg)
			results <- result

			continue
		}
		infos, err := h.parseQueryMultiProcResult(getProcOperateResultV2Resp{key: item})
		if err != nil {
			result.Err = fmt.Errorf("failed to parse process info: %w", err)
			results <- result

			continue
		}
		result.ProcessInfos = make([]types.ProcessInfo, 0, len(infos[result.AgentID]))
		for _, info := range infos[result.AgentID] {
			result.ProcessInfos = append(result.ProcessInfos, types.ProcessInfo{
				AutoStart: info.IsAuto, AgentID: result.AgentID, Pid: info.Pid,
				Version: strings.TrimSpace(info.Version), Status: convPidToProcStatus(info.Pid), LastSyncAt: time.Now(),
			})
		}
		results <- result
	}
}
