/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package backend

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandlerProcess defines the backend Handler for process.
type IHandlerProcess interface {
	// ListProcesses lists processes by page and conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return process list with page and the total count with filter and error.
	ListProcesses(nCtx contextx.IContext, page types.Page, condition *types.ProcessCondition) ([]*types.Process, int64, error)

	// CountProcesses counts processes by conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the process count with filter and error.
	CountProcesses(nCtx contextx.IContext, condition *types.ProcessCondition) (int64, error)

	// GetProcessDistributionByHostID gets process distribution by host id.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the process distribution by host id and error.
	GetProcessDistributionByHostID(nCtx contextx.IContext, condition *types.ProcessCondition) (map[int64]int64, error)

	// GetProcessDistributionByPluginName gets process distribution by plugin name.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the process distribution by plugin name and error.
	GetProcessDistributionByPluginName(nCtx contextx.IContext, condition *types.ProcessCondition) (map[string]int64, error)

	// DistinctProcess distinct process by conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param selector the distinct selector.
	// @param condition the filter conditions.
	// @return the process distinct result and error.
	DistinctProcess(nCtx contextx.IContext, selector types.ProcessDistinctSelector, condition *types.ProcessCondition) (
		*types.ProcessDistinctResult, error)
}

// CountProcesses count processes.
func (h *Handler) CountProcesses(nCtx contextx.IContext, condition *types.ProcessCondition) (int64, error) {
	req := new(protoBackend.ProcessListReq)
	req.ConvertConditionFromTypes(condition)
	req.OnlyCount = true

	resp, err := h.cli.listProcesses(nCtx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// ListProcesses list processes.
func (h *Handler) ListProcesses(nCtx contextx.IContext, page types.Page, condition *types.ProcessCondition) ([]*types.Process, int64, error) {
	req := new(protoBackend.ProcessListReq)
	req.ConvertConditionFromTypes(condition)
	req.Page = convertPage(page)
	req.OnlyCount = false

	resp, err := h.cli.listProcesses(nCtx, req)
	if err != nil {
		return nil, 0, err
	}

	processes, total := resp.ConvertProcessToTypes()

	return processes, total, nil
}

// GetProcessDistributionByHostID get process distribution by host id.
func (h *Handler) GetProcessDistributionByHostID(nCtx contextx.IContext, condition *types.ProcessCondition) (map[int64]int64, error) {
	req := new(protoBackend.GetProcessDistributionByHostIDReq)
	req.ConvertConditionFromTypes(condition)

	resp, err := h.cli.getProcessDistributionByHostID(nCtx, req)
	if err != nil {
		return nil, err
	}

	return resp.GetData(), nil
}

// GetProcessDistributionByPluginName get process distribution by plugin name.
func (h *Handler) GetProcessDistributionByPluginName(nCtx contextx.IContext, condition *types.ProcessCondition) (map[string]int64, error) {
	req := new(protoBackend.GetProcessDistributionByPluginNameReq)
	req.ConvertConditionFromTypes(condition)

	resp, err := h.cli.getProcessDistributionByPluginName(nCtx, req)
	if err != nil {
		return nil, err
	}

	return resp.GetData(), nil
}

// DistinctProcess distinct process by conditions.
func (h *Handler) DistinctProcess(nCtx contextx.IContext, selector types.ProcessDistinctSelector, condition *types.ProcessCondition) (
	*types.ProcessDistinctResult, error) {

	req := new(protoBackend.ProcessDistinctReq)
	req.ConvertSelectorFromTypes(selector)
	req.ConvertConditionFromTypes(condition)

	resp, err := h.cli.distinctProcess(nCtx, req)
	if err != nil {
		return nil, err
	}

	result, err := resp.ConvertResultToTypes()
	if err != nil {
		return nil, err
	}

	return result, nil
}
