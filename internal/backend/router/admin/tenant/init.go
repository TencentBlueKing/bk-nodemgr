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

package tenant

import (
	"fmt"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	pkgTenant "github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	initWorkflowSyncBizAndHost      = "sync_biz_and_host"
	initWorkflowSyncNetworkArea     = "sync_networkarea"
	initWorkflowEnsureDefaultPlugin = "ensure_default_plugin"
	initWorkflowSyncSharedReleases  = "sync_shared_releases"
)

// Init handles the tenant initialization request.
func (h *handler) Init(rCtx restserver.IContext) (any, error) {
	req := new(initReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to init tenant, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := validateInitTenantMode(req.TenantID); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to init tenant, invalid tenant mode")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := validateInitTenantAuth(rCtx, req.TenantID); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to init tenant, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, err)
	}

	targetTenant, err := h.getTargetTenant(rCtx, req.TenantID)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to init tenant, failed to get tenant from user manager")
		return nil, resterrf.ErrWrap(resterrf.ThirdpartyRequestFailed, err)
	}
	if targetTenant == nil {
		err = fmt.Errorf("tenant(%s) does not exist", req.TenantID)
		logger.G.Biz(rCtx).WithErr(err).Error("failed to init tenant, tenant does not exist")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}
	if !targetTenant.Enabled {
		err = fmt.Errorf("tenant(%s) is disabled", req.TenantID)
		logger.G.Biz(rCtx).WithErr(err).Error("failed to init tenant, tenant is disabled")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := h.ensureTargetTenant(rCtx, targetTenant); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to init tenant, failed to ensure tenant storage")
		return nil, resterrf.ErrWrap(resterrf.DBExecCmdFailed, err)
	}

	triggeredWorkflows, err := h.launchInitWorkflows(rCtx, req.TenantID)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to init tenant, failed to launch sync workflows")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	return &initResp{TriggeredWorkflows: triggeredWorkflows}, nil
}

func validateInitTenantMode(targetTenantID string) error {
	if targetTenantID == pkgTenant.SystemTenantID {
		return fmt.Errorf("system tenant cannot be initialized by tenant init")
	}

	if pkgTenant.GetMode() == pkgTenant.ModeSingle && targetTenantID != pkgTenant.SingleModeTenantID {
		return fmt.Errorf("single tenant mode only supports initializing tenant(%s)", pkgTenant.SingleModeTenantID)
	}

	return nil
}

func validateInitTenantAuth(rCtx restserver.IContext, targetTenantID string) error {
	if pkgTenant.GetMode() != pkgTenant.ModeMultiple {
		return nil
	}

	if rCtx.TenantID() != pkgTenant.SystemTenantID {
		return fmt.Errorf("only system tenant can initialize tenant(%s)", targetTenantID)
	}

	return nil
}

func (h *handler) getTargetTenant(rCtx restserver.IContext, tenantID string) (*types.Tenant, error) {
	if h.userManagerHandler == nil {
		return nil, fmt.Errorf("user manager handler is nil")
	}

	tenants, err := h.userManagerHandler.ListALLTenants(rCtx)
	if err != nil {
		return nil, err
	}

	for _, tenant := range tenants {
		if tenant == nil || tenant.ID != tenantID {
			continue
		}

		return tenant, nil
	}

	return nil, nil
}

func (h *handler) ensureTargetTenant(rCtx restserver.IContext, targetTenant *types.Tenant) error {
	if h.tenantStorage == nil {
		return fmt.Errorf("tenant storage is nil")
	}

	return h.tenantStorage.EnsureReservedTenant(rCtx, &types.Tenant{
		ID:      targetTenant.ID,
		Name:    targetTenant.Name,
		Enabled: targetTenant.Enabled,
	})
}

func (h *handler) launchInitWorkflows(rCtx restserver.IContext, tenantID string) ([]string, error) {
	if h.syncManager == nil {
		return nil, fmt.Errorf("sync manager is nil")
	}

	operator := rCtx.BKUsername()
	if operator == "" {
		operator = rCtx.LoginName()
	}
	tenantCtx := contextx.From(rCtx, contextx.WithTenantID(tenantID), contextx.WithBKUsername(operator))

	workflowLaunchers := []struct {
		name   string
		launch func(contextx.IContext) (string, error)
	}{
		{name: initWorkflowSyncBizAndHost, launch: h.syncManager.LaunchSyncBizAndHost},
		{name: initWorkflowSyncNetworkArea, launch: h.syncManager.LaunchSyncNetworkArea},
		{name: initWorkflowEnsureDefaultPlugin, launch: h.syncManager.LaunchEnsureDefaultPlugin},
		{name: initWorkflowSyncSharedReleases, launch: h.syncManager.LaunchSyncSharedReleases},
	}

	triggeredWorkflows := make([]string, 0, len(workflowLaunchers))
	failedReasons := make([]string, 0)
	for _, launcher := range workflowLaunchers {
		if _, err := launcher.launch(tenantCtx); err != nil {
			failedReasons = append(failedReasons, fmt.Sprintf("%s: %v", launcher.name, err))
			continue
		}

		triggeredWorkflows = append(triggeredWorkflows, launcher.name)
	}

	if len(failedReasons) > 0 {
		return triggeredWorkflows, fmt.Errorf("failed to launch init workflows: %s", strings.Join(failedReasons, "; "))
	}

	return triggeredWorkflows, nil
}
