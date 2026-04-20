/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package workflow

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

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
