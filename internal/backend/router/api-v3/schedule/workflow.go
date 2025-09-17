/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package schedule describes the node router.
package schedule

import (
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

// EnableScheduleWorkflow enable schedule workflow.
func (h *handler) EnableScheduleWorkflow(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.EnableScheduleWorkflowReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to enable schedule workflow, failed to decode request body. err: %v", err)

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	schedule, err := h.storageWorkflow.GetScheduleWorkflow(ctx, req.GetWorkflowId())
	if err != nil {
		h.logger.Errorf("failed to get schedule workflow. workflow-id(%s): %v", req.GetWorkflowId(), err)
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	err = h.storageWorkflow.UpdateTriggerState(ctx, schedule.TriggerID, trigger.StateRunning)
	if err != nil {
		h.logger.Errorf("failed to enable schedule workflow. workflow-id(%s), trigger-id(%s): %v",
			schedule.WorkflowID, schedule.TriggerID, err)

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	return &protoBackend.EnableScheduleWorkflowResp_Data{}, nil
}

// DisableScheduleWorkflow disable schedule workflow.
func (h *handler) DisableScheduleWorkflow(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.DisableScheduleWorkflowReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to disable schedule workflow, failed to decode request body. err: %v", err)

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	schedule, err := h.storageWorkflow.GetScheduleWorkflow(ctx, req.GetWorkflowId())
	if err != nil {
		h.logger.Errorf("failed to get schedule workflow. workflow-id(%s): %v", req.GetWorkflowId(), err)
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	err = h.storageWorkflow.UpdateTriggerState(ctx, schedule.TriggerID, trigger.StateTerminated)
	if err != nil {
		h.logger.Errorf("failed to disable schedule workflow. workflow-id(%s), trigger-id(%s): %v",
			schedule.WorkflowID, schedule.TriggerID, err)

		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	return &protoBackend.DisableScheduleWorkflowResp_Data{}, nil
}

// ListScheduleWorkflow list schedule workflow.
func (h *handler) ListScheduleWorkflow(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.ListScheduleWorkflowReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to list schedule workflow, failed to decode request body. err: %v", err)

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if req.GetOnlyCount() {
		cnt, err := h.storageWorkflow.CountScheduleWorkflow(ctx, req.ConvertConditionsToTypes())
		if err != nil {
			h.logger.ErrorCtxf(ctx, "failed to count schedule workflow, failed to decode request body. err: %v", err)
			return nil, resterrf.ErrWrap(resterrf.Aborted, err)
		}

		resp := new(protoBackend.ListScheduleWorkflowResp)
		resp.ConvertScheduleWorkflowsFromTypes(cnt, nil)

		return resp.GetData(), nil
	}

	schedule, cnt, err := h.storageWorkflow.ListScheduleWorkflow(
		ctx,
		req.ConvertPageToTypes(maxScheduleWorkflowLimit),
		req.ConvertConditionsToTypes(),
	)
	if err != nil {
		h.logger.Errorf("failed to list schedule workflow. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.Aborted, err)
	}

	resp := new(protoBackend.ListScheduleWorkflowResp)
	resp.ConvertScheduleWorkflowsFromTypes(cnt, schedule)

	return resp.GetData(), nil
}
