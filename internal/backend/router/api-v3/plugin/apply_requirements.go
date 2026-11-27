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
 * either express or implied. See the License for the specific
 * language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package plugin

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ApplyRequirements distributes applicable dependency files and restarts plugin processes.
func (h *handler) ApplyRequirements(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginApplyRequirementsReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to apply plugin requirements, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	pluginNames := conv.SliceUnique(conv.SliceToSlice(req.GetPlugin(), func(item *protoBackend.PluginApplyRequirementsReq_ApplyInfo) string {
		return item.GetPluginName()
	}))
	if authErr := h.authorizedPluginOperate(rCtx, pluginNames...); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).
			With("plugin-name", pluginNames).
			Error("failed to apply plugin requirements, permission denied")

		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	hostTopoMapping, err := h.daoHost.GetHostTopoRelationMapping(rCtx, req.GetHostIDs()...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to apply plugin requirements, failed to get host topo mapping")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	pluginDeployments, hostIDs, bizIDs, err := types.NewPluginDeploymentsByParams(
		rCtx.TenantID(),
		types.PluginDeploymentTransferOptionsOnlyTransferInstaller(),
		req.ConvertParamToTypesWithHostTopoMapping(hostTopoMapping)...,
	)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to apply plugin requirements, failed to generate plugin deployments")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	workflowID, err := h.pluginMgrIface.LaunchApplyPluginRequirements(
		rCtx,
		types.ApplyPluginRequirementsParam{
			Type:              types.PluginWorkflowTypeApplyRequirements,
			HostIDs:           hostIDs,
			BizIDs:            bizIDs,
			Operator:          rCtx.BKUsername(),
			PluginDeployments: pluginDeployments,
		},
	)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to apply plugin requirements")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	logger.G.Biz(rCtx).With("workflow-id", workflowID).Info("launched apply plugin requirements workflow")

	return &protoBackend.PluginApplyRequirementsResp_Data{WorkflowId: workflowID}, nil
}
