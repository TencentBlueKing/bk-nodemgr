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

package plugin

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Debug starts one debug workflow for an existing plugin deployment.
func (h *handler) StartDebug(rCtx restserver.IContext) (any, error) {
	req := new(protoBackend.PluginStartDebugReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to start plugin debug, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	debugInfo := req.GetDebugInfo()
	scope, err := debugInfo.ConvertScopeToTypes()
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to start plugin debug, invalid scope")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	targets, err := h.cmdbHandler.GetTargetByScopeInstance(rCtx, scope)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to start plugin debug, failed to resolve scope")

		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	pluginName := debugInfo.GetPluginName()
	if authErr := h.authorizedPluginOperate(rCtx, pluginName); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).With("plugin-name", pluginName).Error("failed to start plugin debug, permission denied.")

		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	deployParams := make([]*types.PluginDeploymentParam, 0, len(targets))
	for _, target := range targets {
		param := &types.PluginDeploymentParam{
			BizID:               target.Host.Static.BizID,
			HostID:              target.Host.HostID,
			PluginName:          debugInfo.GetPluginName(),
			Version:             debugInfo.GetVersion(),
			ConfigTemplateName:  debugInfo.GetConfigTemplateName(),
			CustomConfigContext: make(map[string]any),
		}
		if debugInfo.GetCustomConfigContext() != nil {
			param.CustomConfigContext = debugInfo.GetCustomConfigContext().AsMap()
		}

		deployParams = append(deployParams, param)
	}
	pluginDeployments, _, _, err := types.NewPluginDeploymentsByParams(rCtx.TenantID(), types.PluginDeploymentTransferOptionsAll(), deployParams...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to start plugin debug, failed to generate plugin deployments.")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	hostIDs := conv.SliceToSlice(targets, func(target *types.Target) int64 {
		return target.Host.HostID
	})

	bizIDs := conv.SliceToSlice(targets, func(target *types.Target) int64 {
		return target.Host.Static.BizID
	})
	workflowID, err := h.pluginMgrIface.LaunchStartDebugPlugin(rCtx, types.StartDebugPluginParam{
		Type:              types.PluginWorkflowTypeDebug,
		HostIDs:           hostIDs,
		BizIDs:            bizIDs,
		Operator:          rCtx.BKUsername(),
		PluginDeployments: pluginDeployments,
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to start plugin debug")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	respData := &protoBackend.PluginStartDebugResp_Data{
		WorkflowId: workflowID,
	}

	logger.G.Biz(rCtx).With("workflow-id", workflowID).Info("launched plugin debug workflow")

	return respData, nil
}

// StopDebug stores a stop signal for one debug workflow.
func (h *handler) StopDebug(rCtx restserver.IContext) (any, error) {
	req := new(protoBackend.PluginStopDebugReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to stop plugin debug, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	workflow, err := h.daoPluginWorkflow.GetPluginWorkflow(rCtx, req.GetWorkflowId())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to stop plugin debug, failed to get workflow")

		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	if authErr := h.authorizer.Check(rCtx, auth.ActionPluginOperate, authRouter.BuildBizResources(workflow.BizIDs...)); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).With("workflow-id", req.GetWorkflowId()).Error("failed to stop plugin debug, permission denied")

		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	if err = h.pluginMgrIface.LaunchStopDebugPlugin(rCtx, types.StopDebugPluginParam{WorkflowID: req.GetWorkflowId()}); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to stop plugin debug")

		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	logger.G.Biz(rCtx).With("workflow-id", req.GetWorkflowId()).Info("requested plugin debug stop")

	return &protoBackend.PluginStopDebugResp_Data{}, nil
}
