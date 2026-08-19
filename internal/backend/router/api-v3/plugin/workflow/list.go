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

// Package workflow describes the workflow router.
package workflow

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ListPluginWorkflow lists plugin workflows with permission filtering.
func (h *handler) ListPluginWorkflow(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginWorkflowListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin workflow, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	condition, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin workflow, failed to convert conditions")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	narrowedBizIDs, scopeIsAny, authErr := h.narrowAuthorizedBizIDsForPluginHistoryView(rCtx, condition.ExactInclude.BizID)
	if authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to list plugin workflow, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}
	condition = narrowPluginConditionByBiz(condition, narrowedBizIDs, scopeIsAny)

	// only count.
	if req.GetOnlyCount() {
		num, err := h.daoPluginWorkflow.CountPluginWorkflow(rCtx, condition)
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin workflow, failed to count workflow")
			return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
		}

		resp := new(protoBackend.PluginWorkflowListResp)
		resp.ConvertPluginWorkflowsFromTypes(num, nil)

		return resp.GetData(), nil
	}

	page, err := req.ConvertPageToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin workflow, invalid page info")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	workflows, total, err := h.daoPluginWorkflow.ListPluginWorkflow(rCtx, page, condition)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list plugin workflow")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PluginWorkflowListResp)

	resp.ConvertPluginWorkflowsFromTypes(total, workflows)

	return resp.GetData(), nil
}

// DistinctPluginWorkflow workflow distinct.
func (h *handler) DistinctPluginWorkflow(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginWorkflowDistinctReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct plugin workflow, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	cond, err := req.ConvertConditionsToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct plugin workflow, failed to convert conditions")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	result, err := h.daoPluginWorkflow.DistinctPluginWorkflow(
		rCtx,
		types.NewPluginWorkflowDistinctRequestAllSet(),
		cond)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to distinct plugin workflow. failed to distinct plugin workflow fields: %v", err)
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	resp := new(protoBackend.PluginWorkflowDistinctResp)
	resp.ConvertResultFromTypes(result)

	return resp.GetData(), nil
}
