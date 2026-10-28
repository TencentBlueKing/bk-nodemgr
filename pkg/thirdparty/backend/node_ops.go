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

package backend

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/pageexecutor"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandlerNodeOps provides access to node out-of-band operation history.
type IHandlerNodeOps interface {
	// ListNodeOpsHistory returns the requested range of authorized metadata in backend-sized batches.
	ListNodeOpsHistory(nCtx contextx.IContext, page types.Page, condition *types.NodeOpsHistoryCondition) ([]*types.NodeOpsHistory, int64, error)

	// CountNodeOpsHistory counts authorized operations matching the conditions.
	CountNodeOpsHistory(nCtx contextx.IContext, condition *types.NodeOpsHistoryCondition) (int64, error)

	// GetNodeOpsHistory returns all authorized archived snapshots for an operation.
	GetNodeOpsHistory(nCtx contextx.IContext, operateID string) ([]*types.NodeOpsHistoryHostSnapshot, error)
}

// ListNodeOpsHistory returns the requested range of authorized metadata in backend-sized batches.
func (h *Handler) ListNodeOpsHistory(nCtx contextx.IContext, page types.Page, condition *types.NodeOpsHistoryCondition) (
	[]*types.NodeOpsHistory, int64, error) {

	req := new(protoBackend.NodeOpsHistoryListReq)
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return nil, 0, fmt.Errorf("failed to convert node ops history conditions: %w", err)
	}

	var (
		total     int64
		histories []*types.NodeOpsHistory
	)
	executor := pageexecutor.NewPageExecutor[*types.NodeOpsHistory](req.PageLimit(), req.PageTimeout())
	fn := func(nCtx contextx.IContext, page types.Page) ([]*types.NodeOpsHistory, error) {
		req.Page = convertPage(page)
		resp, err := h.cli.listNodeOpsHistory(nCtx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to request node ops history page: %w", err)
		}

		histories, total = resp.ConvertNodeOpsHistoriesToTypes()

		return histories, nil
	}

	result, err := executor.Execute(nCtx, page, fn)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list node ops history: %w", err)
	}

	return result.Items, total, nil
}

// CountNodeOpsHistory counts authorized operations matching the conditions.
func (h *Handler) CountNodeOpsHistory(nCtx contextx.IContext, condition *types.NodeOpsHistoryCondition) (int64, error) {
	// The backend validates pagination even when only_count is enabled.
	req := &protoBackend.NodeOpsHistoryListReq{
		OnlyCount: true,
		Page:      convertPage(types.SingleItemPage()),
	}
	if err := req.ConvertConditionsFromTypes(condition); err != nil {
		return 0, fmt.Errorf("failed to convert node ops history conditions: %w", err)
	}

	resp, err := h.cli.listNodeOpsHistory(nCtx, req)
	if err != nil {
		return 0, fmt.Errorf("failed to count node ops history: %w", err)
	}

	return resp.GetData().GetTotal(), nil
}

// GetNodeOpsHistory returns all authorized archived snapshots for an operation.
func (h *Handler) GetNodeOpsHistory(nCtx contextx.IContext, operateID string) ([]*types.NodeOpsHistoryHostSnapshot, error) {
	req := &protoBackend.NodeOpsHistoryDetailReq{
		OperateId: operateID,
	}
	resp, err := h.cli.getNodeOpsHistory(nCtx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get node ops history: %w", err)
	}

	return resp.ConvertSnapshotsToTypes(), nil
}
