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
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
)

func (c *cli) updateScheduleWorkflowInterval(ctx contextx.IContext, req *protoBackend.UpdateScheduleWorkflowIntervalReq) (
	*protoBackend.UpdateScheduleWorkflowIntervalResp, error) {

	resp := new(protoBackend.UpdateScheduleWorkflowIntervalResp)
	err := c.client.Post().SubResourcef("/schedule/workflow/interval/update").
		WithContext(ctx).WithHeaders(c.getHeader(ctx)).Body(req).
		EnableLogBody().EnableLogResponse().Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to request schedule workflow interval update: %w", err)
	}
	if resp.GetCode() != CodeOK {
		return nil, buildBackendResponseError("update schedule workflow interval", resp, resp.GetError())
	}

	return resp, nil
}

func (c *cli) listScheduleWorkflow(ctx contextx.IContext, req *protoBackend.ListScheduleWorkflowReq) (*protoBackend.ListScheduleWorkflowResp, error) {
	resp := new(protoBackend.ListScheduleWorkflowResp)
	err := c.client.Post().SubResourcef("/schedule/workflow/list").
		WithContext(ctx).WithHeaders(c.getHeader(ctx)).Body(req).
		EnableLogBody().EnableLogResponse().Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to request schedule workflow list: %w", err)
	}
	if resp.GetCode() != CodeOK {
		return nil, buildBackendResponseError("list schedule workflow", resp, resp.GetError())
	}
	if resp.GetData() == nil {
		return nil, fmt.Errorf("list schedule workflow returned empty data, request-id(%s)", resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listScheduleWorkflowOperation(ctx contextx.IContext, req *protoBackend.ScheduleWorkflowOperationListReq) (
	*protoBackend.ScheduleWorkflowOperationListResp, error) {

	resp := new(protoBackend.ScheduleWorkflowOperationListResp)
	err := c.client.Post().SubResourcef("/schedule/workflow/operation/list").
		WithContext(ctx).WithHeaders(c.getHeader(ctx)).Body(req).
		EnableLogBody().EnableLogResponse().Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to request schedule workflow operations: %w", err)
	}
	if resp.GetCode() != CodeOK {
		return nil, buildBackendResponseError("list schedule workflow operation", resp, resp.GetError())
	}
	if resp.GetData() == nil {
		return nil, fmt.Errorf("list schedule workflow operation returned empty data, request-id(%s)", resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) listScheduleWorkflowOperationInstance(
	ctx contextx.IContext, req *protoBackend.ScheduleWorkflowOperationInstanceListReq,
) (*protoBackend.ScheduleWorkflowOperationInstanceListResp, error) {

	resp := new(protoBackend.ScheduleWorkflowOperationInstanceListResp)
	err := c.client.Post().SubResourcef("/schedule/workflow/operation/instance/list").
		WithContext(ctx).WithHeaders(c.getHeader(ctx)).Body(req).
		EnableLogBody().EnableLogResponse().Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to request schedule workflow instances: %w", err)
	}
	if resp.GetCode() != CodeOK {
		return nil, buildBackendResponseError("list schedule workflow operation instance", resp, resp.GetError())
	}
	if resp.GetData() == nil {
		return nil, fmt.Errorf("list schedule workflow operation instance returned empty data, request-id(%s)", resp.GetRequestId())
	}

	return resp, nil
}

func (c *cli) getScheduleWorkflowOperationInstanceLog(
	ctx contextx.IContext, req *protoBackend.ScheduleWorkflowOperationInstanceLogGetReq,
) (*protoBackend.ScheduleWorkflowOperationInstanceLogGetResp, error) {

	resp := new(protoBackend.ScheduleWorkflowOperationInstanceLogGetResp)
	err := c.client.Post().SubResourcef("/schedule/workflow/operation/instance/log/get").
		WithContext(ctx).WithHeaders(c.getHeader(ctx)).Body(req).
		EnableLogBody().EnableLogResponse().Do().Into(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to request schedule workflow instance logs: %w", err)
	}
	if resp.GetCode() != CodeOK {
		return nil, buildBackendResponseError("get schedule workflow operation instance log", resp, resp.GetError())
	}
	if resp.GetData() == nil {
		return nil, fmt.Errorf("get schedule workflow operation instance log returned empty data, request-id(%s)", resp.GetRequestId())
	}

	return resp, nil
}
