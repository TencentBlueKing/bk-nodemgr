/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
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

// Upgrade defines the handler to upgrade plugin.
func (h *handler) Upgrade(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.PluginUpgradeReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upgrade plugin, failed to decode request body.")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	pluginName := conv.SliceToSlice(req.GetPlugin(), func(item *protoBackend.PluginOperateFullInfo) string {
		return item.GetPluginName()
	})
	if authErr := h.authorizedPluginOperate(rCtx, pluginName...); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).
			With("plugin-name", pluginName).
			Error("failed to upgrade plugin, permission denied.")

		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	hostBizMapping, err := h.daoHost.GetHostBizMapping(rCtx, req.GetHostIDs())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upgrade plugin, failed to get host biz mapping.")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	pluginDeployments, hostIDs, bizIDs, err := types.NewPluginDeploymentsByParams(
		rCtx.TenantID(), types.DefaultPluginDeploymentTransferOptions(), req.ConvertParamToTypesWithHostBizMapping(hostBizMapping)...)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upgrade plugin, failed to generate plugin deployments.")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	workflowID, err := h.pluginMgrIface.LaunchUpgradePlugin(rCtx, types.UpgradePluginParam{
		Type:              types.PluginWorkflowTypeUpgrade,
		HostIDs:           hostIDs,
		BizIDs:            bizIDs,
		Operator:          rCtx.BKUsername(),
		PluginDeployments: pluginDeployments,
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to upgrade plugin.")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	respData := &protoBackend.PluginUpgradeResp_Data{
		WorkflowId: workflowID,
	}

	logger.G.Biz(rCtx).With("workflow-id", workflowID).Info("launched upgrade plugin workflow")

	return respData, nil
}
