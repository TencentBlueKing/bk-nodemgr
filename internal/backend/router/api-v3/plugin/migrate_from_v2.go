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

// Package plugin defines the router to handle the plugin request.
package plugin

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// MigrateFromV2 defines the handler to migrate plugin process from v2.
func (h *handler) MigrateFromV2(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginMigrateFromV2Req)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to migrate plugin from v2, failed to decode request body.")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	pluginName := conv.SliceToSlice(req.GetPlugin(), func(item *protoBackend.PluginMigrateFromV2Req_MigrateFromV2Info) string {
		return item.GetPluginName()
	})
	if authErr := h.authorizedPluginOperate(rCtx, pluginName...); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).
			With("plugin-name", pluginName).
			Error("failed to migrate plugin from v2, permission denied.")

		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	hostTopoMapping, err := h.daoHost.GetHostTopoRelationMapping(rCtx, req.GetHostIDs()...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to migrate plugin from v2, failed to get host topo mapping.")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	pluginDeployments, hostIDs, bizIDs, err := types.NewPluginDeploymentsByParams(
		rCtx.TenantID(), types.PluginDeploymentTransferOptionsOnlyTransferInstaller(), req.ConvertParamToTypesWithHostTopoMapping(hostTopoMapping)...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to migrate plugin from v2, failed to generate plugin deployments.")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	workflowID, err := h.pluginMgrIface.LaunchMigrateFromV2(rCtx, types.MigrateFromV2Param{
		Type:              types.PluginWorkflowTypeMigrateV2,
		HostIDs:           hostIDs,
		BizIDs:            bizIDs,
		Operator:          rCtx.BKUsername(),
		PluginDeployments: pluginDeployments,
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to migrate plugin from v2.")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	respData := &protoBackend.PluginMigrateFromV2Resp_Data{
		WorkflowId: workflowID,
	}

	logger.G.Biz(rCtx).With("workflow-id", workflowID).Info("launched migrate plugin from v2 workflow")

	return respData, nil
}
