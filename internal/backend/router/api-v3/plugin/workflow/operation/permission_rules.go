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

package operation

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

func (h *handler) authorizedPluginHistoryView(rCtx restserver.IContext, workflowID string) error {
	pluginWorkflow, err := h.daoPluginWorkflow.GetPluginWorkflow(rCtx, workflowID)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to check workflow history view permission, failed to get the target plugin workflow")
		return resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resources := authRouter.BuildBizResources(pluginWorkflow.BizIDs...)
	if authErr := h.authorizer.Check(rCtx, auth.ActionPluginHistoryView, resources); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to check workflow history view permission, permission denied")
		return resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	return nil
}

func (h *handler) authorizedPluginInstallParamsView(rCtx restserver.IContext, workflowID string) error {
	workflow, err := h.daoPluginWorkflow.GetPluginWorkflow(rCtx, workflowID)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to check install params view permission, failed to get workflow")
		return resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}
	if workflow.TenantID != rCtx.TenantID() || len(workflow.BizIDs) == 0 {
		err := fmt.Errorf("invalid workflow authorization scope")
		logger.G.Biz(rCtx).WithErr(err).Error("failed to check install params view permission")

		return resterrf.ErrWrap(resterrf.PermissionDenied, err)
	}
	resources := authRouter.BuildBizResources(workflow.BizIDs...)
	actionResources := map[auth.Action][]types.AuthResource{
		auth.ActionPluginHistoryView: resources,
		auth.ActionPluginView:        resources,
	}
	if err := h.authorizer.CheckMany(rCtx, actionResources); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to check install params view permission, permission denied")
		return resterrf.ErrWrap(resterrf.PermissionDenied, err)
	}

	return nil
}

func (h *handler) authorizedPluginOperate(rCtx restserver.IContext, workflowID string) error {
	pluginWorkflow, err := h.daoPluginWorkflow.GetPluginWorkflow(rCtx, workflowID)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to check workflow operate permission, failed to get the target plugin workflow")
		return resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	bizIDs := pluginWorkflow.BizIDs
	resources := authRouter.BuildBizResources(bizIDs...)

	if authErr := h.authorizer.Check(rCtx, auth.ActionPluginOperate, resources); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to check workflow operate permission, permission denied")
		return resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	return nil
}
