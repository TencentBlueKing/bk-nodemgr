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

// CountProcesses count processes.
func (h *Handler) CountProcesses(ctx contextx.IContext, condition *types.ProcessCondition) (int64, error) {
	req := new(protoBackend.ProcessListReq)
	req.ConvertConditionFromTypes(condition)
	req.OnlyCount = true

	resp, err := h.cli.listProcesses(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.GetData().GetTotal(), nil
}

// ListProcesses list processes.
func (h *Handler) ListProcesses(ctx contextx.IContext, page types.Page, condition *types.ProcessCondition) ([]*types.Process, int64, error) {
	req := new(protoBackend.ProcessListReq)
	req.ConvertConditionFromTypes(condition)
	req.Page = convertPage(page)
	req.OnlyCount = false

	resp, err := h.cli.listProcesses(ctx, req)
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
