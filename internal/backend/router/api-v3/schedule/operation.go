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

package schedule

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (h *handler) checkScheduleManagePermission(rCtx restserver.IContext) error {
	resources := auth.BuildPlatformResources(types.PlatformResourceSchedulerWorkflow)
	if err := h.authorizer.Check(rCtx, auth.ActionPlatformManage, resources); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query schedule workflow, permission denied")
		return resterrf.ErrWrap(resterrf.PermissionDenied, err)
	}

	return nil
}

// ListOperation lists root operations or a specific instance's direct children.
func (h *handler) ListOperation(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ScheduleWorkflowOperationListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list schedule operations, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	if err := h.checkScheduleManagePermission(rCtx); err != nil {
		return nil, err
	}
	page := types.UnlimitedPage()
	if !req.GetOnlyCount() {
		var err error
		page, err = req.ConvertPageToTypes()
		if err != nil {
			logger.G.Biz(rCtx).WithErr(err).Error("failed to list schedule operations, invalid page")
			return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
		}
	}
	operations, total, err := h.storageWorkflow.ListScheduledWorkflowOperation(rCtx, page,
		req.GetWorkflowId(), req.GetParentOperInstId(), req.GetOnlyCount())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list schedule operations")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}
	resp := new(protoBackend.ScheduleWorkflowOperationListResp)
	resp.ConvertResultFromTypes(total, operations)

	return resp.GetData(), nil
}

// ListOperationInstance lists instances after checking every operation and instance identifier.
func (h *handler) ListOperationInstance(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ScheduleWorkflowOperationInstanceListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list schedule instances, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	if err := h.checkScheduleManagePermission(rCtx); err != nil {
		return nil, err
	}
	instances, total, err := h.storageWorkflow.ListScheduledWorkflowOperationInstance(rCtx,
		req.GetOperationId(), req.GetOperInstId(), req.GetOnlyCount())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list schedule instances")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}
	resp := new(protoBackend.ScheduleWorkflowOperationInstanceListResp)
	resp.ConvertResultFromTypes(total, instances)

	return resp.GetData(), nil
}

// GetOperationInstanceLog gets action and extra-execution logs in a scheduled task chain.
func (h *handler) GetOperationInstanceLog(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.ScheduleWorkflowOperationInstanceLogGetReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get schedule instance logs, invalid request")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	if err := h.checkScheduleManagePermission(rCtx); err != nil {
		return nil, err
	}
	instance, err := h.storageWorkflow.GetScheduledWorkflowOperationInstanceLog(rCtx, req.GetOperInstId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to get schedule instance logs")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}
	resp := new(protoBackend.ScheduleWorkflowOperationInstanceLogGetResp)
	resp.ConvertResultFromTypes(instance)

	return resp.GetData(), nil
}
