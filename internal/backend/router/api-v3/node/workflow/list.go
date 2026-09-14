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

package workflow

import (
	"fmt"

	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ListNodeWorkflow lists workflows with permission filtering.
func (h *handler) ListNodeWorkflow(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeWorkflowListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list node workflow, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	// Convert conditions
	cond, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list node workflow, failed to convert conditions")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	page := types.UnlimitedPage()
	if !req.GetOnlyCount() {
		page, err = req.ConvertPageToTypes()
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list node workflow, invalid page info")
			return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
		}
	}

	if cond.ExactInclude != nil && len(cond.ExactInclude.WorkflowID) > 0 {
		workflows, num, err := h.listNodeWorkflowsByIDs(rCtx, page, cond)
		if err != nil {
			return nil, err
		}
		if req.GetOnlyCount() {
			workflows = nil
		}
		resp := new(protoBackend.NodeWorkflowListResp)
		resp.ConvertNodeWorkflowsFromTypes(num, workflows)

		return resp.GetData(), nil
	}

	// Extract requested BizIDs
	var bizIDs []int64
	if exactCond := cond.ExactInclude; exactCond != nil {
		bizIDs = exactCond.BizID
	}

	// Permission narrowing: filter by authorized business IDs
	narrowedBizIDs, scopeIsAny, authErr := h.narrowAuthorizedBizIDsForWorkflowList(rCtx, bizIDs, cond)
	if authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to list node workflow, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	// Update condition with narrowed BizIDs
	cond = narrowWorkflowConditionByBiz(cond, narrowedBizIDs, scopeIsAny)

	// only count.
	if req.GetOnlyCount() {
		num, err := h.daoNodeWorkflow.CountNodeWorkflow(rCtx, cond)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list node workflow, failed to count workflow")
			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.NodeWorkflowListResp)
		resp.ConvertNodeWorkflowsFromTypes(num, nil)

		return resp.GetData(), nil
	}

	workflows, num, err := h.daoNodeWorkflow.ListNodeWorkflow(rCtx, page, cond)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list node workflow")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.NodeWorkflowListResp)

	resp.ConvertNodeWorkflowsFromTypes(num, workflows)

	return resp.GetData(), nil
}

func (h *handler) listNodeWorkflowsByIDs(
	rCtx restserver.IContext, page types.Page, cond *types.NodeWorkflowCondition,
) ([]*types.NodeWorkflow, int64, error) {
	// Authorize the complete matching ID set before exposing a page or its total.
	workflows, num, err := h.daoNodeWorkflow.ListNodeWorkflow(rCtx, types.UnlimitedPage(), cond)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list node workflow")
		return nil, 0, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}
	if num != int64(len(workflows)) {
		err := fmt.Errorf("incomplete node workflow read: expected %d, got %d", num, len(workflows))
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list node workflow")

		return nil, 0, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	for _, workflow := range workflows {
		// Keep each workflow's roles paired with its own businesses, including empty resources.
		if err := h.authorizer.CheckMany(rCtx, authRouter.BuildNodeWorkflowHistoryResources(workflow)); err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list node workflow, permission denied")
			return nil, 0, resterrf.ErrWrap(resterrf.PermissionDenied, err)
		}
	}

	start := min(page.Offset, len(workflows))
	end := start + min(page.Limit, len(workflows)-start)

	return workflows[start:end], num, nil
}

// DistinctNodeWorkflow returns distinct workflow field values.
func (h *handler) DistinctNodeWorkflow(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeWorkflowDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct node workflow, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	cond, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct node workflow, failed to convert conditions")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	result, err := h.daoNodeWorkflow.DistinctNodeWorkflow(
		rCtx,
		types.NewNodeWorkflowDistinctRequestAllSet(),
		cond)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct host. failed to distinct host fields: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.NodeWorkflowDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}
