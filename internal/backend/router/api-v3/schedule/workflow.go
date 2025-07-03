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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

// EnableScheduleWorkflow enable schedule workflow.
func (h *handler) EnableScheduleWorkflow(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to enable schedule workflow, failed to get request context. err: %v", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoBackend.EnableScheduleWorkflowReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to enable schedule workflow, failed to decode request body. err: %v", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	schedule, err := h.storageScheduleWorkflow.GetScheduleWorkflow(sCtx, req.GetWorkflowId())
	if err != nil {
		h.logger.Errorf("failed to get schedule workflow. workflow-id(%s), err: %v", req.GetWorkflowId(), err)
		return nil, errf.ErrWrap(errf.Aborted, err)
	}

	err = h.storageTrigger.UpdateTriggerState(sCtx, schedule.TriggerID, trigger.StateRunning)
	if err != nil {
		h.logger.Errorf("failed to enable schedule workflow. workflow-id(%s), trigger-id(%s), err: %v",
			schedule.WorkflowID, schedule.TriggerID, err)

		return nil, errf.ErrWrap(errf.Aborted, err)
	}

	return &protoBackend.EnableScheduleWorkflowResp_Data{}, nil
}

// DisableScheduleWorkflow disable schedule workflow.
func (h *handler) DisableScheduleWorkflow(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to disable schedule workflow, failed to get request context. err: %v", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoBackend.DisableScheduleWorkflowReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to disable schedule workflow, failed to decode request body. err: %v", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	schedule, err := h.storageScheduleWorkflow.GetScheduleWorkflow(sCtx, req.GetWorkflowId())
	if err != nil {
		h.logger.Errorf("failed to get schedule workflow. workflow-id(%s), err: %v", req.GetWorkflowId(), err)
		return nil, errf.ErrWrap(errf.Aborted, err)
	}

	err = h.storageTrigger.UpdateTriggerState(sCtx, schedule.TriggerID, trigger.StateTerminated)
	if err != nil {
		h.logger.Errorf("failed to disable schedule workflow. workflow-id(%s), trigger-id(%s), err: %v",
			schedule.WorkflowID, schedule.TriggerID, err)

		return nil, errf.ErrWrap(errf.Aborted, err)
	}

	return &protoBackend.DisableScheduleWorkflowResp_Data{}, nil
}

// ListScheduleWorkflow list schedule workflow.
func (h *handler) ListScheduleWorkflow(ctx *rest.Context) (interface{}, error) {
	sCtx, err := ctx.GetContext()
	if err != nil {
		h.logger.Errorf("failed to list schedule workflow, failed to get request context. err: %v", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	req := new(protoBackend.ListScheduleWorkflowReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(sCtx, "failed to list schedule workflow, failed to decode request body. err: %v", err)

		return nil, errf.ErrWrap(errf.InvalidParameter, err)
	}

	if req.GetOnlyCount() {
		cnt, err := h.storageScheduleWorkflow.CountScheduleWorkflow(sCtx, req.ConvertConditionsToTypes())
		if err != nil {
			h.logger.ErrorCtxf(sCtx, "failed to count schedule workflow, failed to decode request body. err: %v", err)
			return nil, errf.ErrWrap(errf.Aborted, err)
		}

		resp := new(protoBackend.ListScheduleWorkflowResp)
		resp.ConvertScheduleWorkflowsFromTypes(cnt, nil)

		return resp.GetData(), nil
	}

	schedule, cnt, err := h.storageScheduleWorkflow.ListScheduleWorkflow(
		sCtx,
		req.ConvertPageToTypes(maxScheduleWorkflowLimit),
		req.ConvertConditionsToTypes(),
	)
	if err != nil {
		h.logger.Errorf("failed to list schedule workflow. err: %v", err)
		return nil, errf.ErrWrap(errf.Aborted, err)
	}

	resp := new(protoBackend.ListScheduleWorkflowResp)
	resp.ConvertScheduleWorkflowsFromTypes(cnt, schedule)

	return resp.GetData(), nil
}
