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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandlerNodeOps provides access to node out-of-band settings and operation history.
type IHandlerNodeOps interface {
	// ListNodeOpsHistory lists authorized node ops history within specified tenant in contextx,
	// aggregated across backend-sized pages.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param page describes the page info when listing.
	// @param condition the filter conditions.
	// @return the authorized node ops history items in the requested range, the total count with filter and error.
	ListNodeOpsHistory(nCtx contextx.IContext, page types.Page, condition *types.NodeOpsHistoryCondition) ([]*types.NodeOpsHistory, int64, error)

	// CountNodeOpsHistory counts authorized node ops history matching the conditions.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param condition the filter conditions.
	// @return the authorized node ops history count with filter and error.
	CountNodeOpsHistory(nCtx contextx.IContext, condition *types.NodeOpsHistoryCondition) (int64, error)

	// GetNodeOpsHistory gets all authorized archived host snapshots for an operation.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param operateID the operation ID.
	// @return the authorized archived host snapshot list and error.
	GetNodeOpsHistory(nCtx contextx.IContext, operateID string) ([]*types.NodeOpsHistoryHostSnapshot, error)

	// UpdateAgentOpsFields updates node agent ops fields.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param hosts the host list with ops fields.
	// @return the error.
	UpdateAgentOpsFields(nCtx contextx.IContext, hosts []*types.Host) error

	// UpdateProxyOpsFields updates node proxy ops fields.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param hosts the host list with ops fields.
	// @return the error.
	UpdateProxyOpsFields(nCtx contextx.IContext, hosts []*types.Host) error

	// ListAgentOpsFields lists node agent ops fields.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param hostIDs the host ID list whose current ops fields are requested; duplicated IDs are deduplicated.
	// @return the hosts with current ops fields and error; the call fails when hostIDs is empty or any host is unmatched.
	ListAgentOpsFields(nCtx contextx.IContext, hostIDs []int64) ([]*types.Host, error)

	// ListProxyOpsFields lists node proxy ops fields.
	// @param nCtx contextx.IContext, contains tenant-id and username.
	// @param hostIDs the host ID list whose current ops fields are requested; duplicated IDs are deduplicated.
	// @return the hosts with current ops fields and error; the call fails when hostIDs is empty or any host is unmatched.
	ListProxyOpsFields(nCtx contextx.IContext, hostIDs []int64) ([]*types.Host, error)
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

// UpdateAgentOpsFields updates node agent ops fields.
func (h *Handler) UpdateAgentOpsFields(nCtx contextx.IContext, hosts []*types.Host) error {
	req := new(protoBackend.NodeOpsUpdateFieldsAgentReq)
	req.ConvertParamFromTypes(hosts)

	_, err := h.cli.updateNodeOpsFieldsAgent(nCtx, req)
	if err != nil {
		return err
	}

	return nil
}

// UpdateProxyOpsFields updates node proxy ops fields.
func (h *Handler) UpdateProxyOpsFields(nCtx contextx.IContext, hosts []*types.Host) error {
	req := new(protoBackend.NodeOpsUpdateFieldsProxyReq)
	req.ConvertParamFromTypes(hosts)

	_, err := h.cli.updateNodeOpsFieldsProxy(nCtx, req)
	if err != nil {
		return err
	}

	return nil
}

// ListAgentOpsFields lists node agent ops fields for the requested hosts,
// deduplicating host IDs and querying them in backend-sized batches.
func (h *Handler) ListAgentOpsFields(nCtx contextx.IContext, hostIDs []int64) ([]*types.Host, error) {
	targets := conv.SliceUnique(hostIDs)
	if len(targets) == 0 {
		return nil, fmt.Errorf("no host IDs provided for listing agent node ops fields")
	}

	req := new(protoBackend.NodeOpsFieldsListAgentReq)
	executor := pageexecutor.NewPageExecutor[*types.Host](req.PageLimit(), req.PageTimeout())
	fn := func(nCtx contextx.IContext, page types.Page) ([]*types.Host, error) {
		req.BkHostId = targets[page.Offset : page.Offset+page.Limit]
		resp, err := h.cli.listNodeOpsFieldsAgent(nCtx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to request agent node ops fields page: %w", err)
		}

		return resp.ConvertHostsToTypes(), nil
	}

	result, err := executor.Execute(nCtx, types.Page{Offset: 0, Limit: len(targets)}, fn)
	if err != nil {
		return nil, fmt.Errorf("failed to list agent node ops fields: %w", err)
	}

	return result.Items, nil
}

// ListProxyOpsFields lists node proxy ops fields for the requested hosts,
// deduplicating host IDs and querying them in backend-sized batches.
func (h *Handler) ListProxyOpsFields(nCtx contextx.IContext, hostIDs []int64) ([]*types.Host, error) {
	targets := conv.SliceUnique(hostIDs)
	if len(targets) == 0 {
		return nil, fmt.Errorf("no host IDs provided for listing proxy node ops fields")
	}

	req := new(protoBackend.NodeOpsFieldsListProxyReq)
	executor := pageexecutor.NewPageExecutor[*types.Host](req.PageLimit(), req.PageTimeout())
	fn := func(nCtx contextx.IContext, page types.Page) ([]*types.Host, error) {
		req.BkHostId = targets[page.Offset : page.Offset+page.Limit]
		resp, err := h.cli.listNodeOpsFieldsProxy(nCtx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to request proxy node ops fields page: %w", err)
		}

		return resp.ConvertHostsToTypes(), nil
	}

	result, err := executor.Execute(nCtx, types.Page{Offset: 0, Limit: len(targets)}, fn)
	if err != nil {
		return nil, fmt.Errorf("failed to list proxy node ops fields: %w", err)
	}

	return result.Items, nil
}
