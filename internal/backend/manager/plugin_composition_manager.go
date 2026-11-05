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

package manager

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin"
	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/globalsettings"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// LaunchComposePluginSubConfig launches source-to-target sub-config composition.
// Each operation orders one source before one target; separate operations may run concurrently.
func (mgr *Manager) LaunchComposePluginSubConfig(nCtx contextx.IContext, param types.ComposePluginSubConfigParam) (string, error) {
	if err := param.SourceOperation.Validate(); err != nil {
		return "", err
	}
	if err := validateCompositionDeployments(nCtx, param.PluginDeployments); err != nil {
		return "", err
	}
	triggerCtl, err := mgr.createOrderedPluginTrigger(
		nCtx,
		globalsettings.OperPluginApplySubConfigMaxConcurrencyNum,
		globalsettings.OperPluginApplySubConfigMaxConcurrencyNumDefault,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create composition trigger: %w", err)
	}
	workflowID := identifier.GenWorkflowID()
	if err := mgr.conf.StoragePlugin.CreatePluginWorkflow(nCtx, &types.PluginWorkflow{
		TenantID: nCtx.TenantID(), WorkflowID: workflowID, TriggerID: triggerCtl.GetTriggerID(),
		Type: types.PluginWorkflowTypeComposePluginSubConfig, HostIDs: param.HostIDs, BizIDs: param.BizIDs,
		Operator: param.Operator, DeployPolicyIDs: param.DeployPolicyIDs,
		OperateTime: time.Now(), Status: types.PluginWorkflowStatusRunning,
	}); err != nil {
		return "", fmt.Errorf("failed to create composition workflow: %w", err)
	}
	needRollback := true
	defer func() {
		if !needRollback {
			return
		}
		if err := triggerCtl.InactivateTrigger(nCtx); err != nil {
			logger.G.Biz(nCtx).WithErr(err).With("trigger-id", triggerCtl.GetTriggerID()).Error("failed to inactivate composition trigger")
		}
		if err := mgr.conf.StoragePlugin.UpdatePluginWorkflowStatus(nCtx, workflowID, types.PluginWorkflowStatusFailed); err != nil {
			logger.G.Biz(nCtx).WithErr(err).With("workflow-id", workflowID).Error("failed to mark composition workflow as failed")
		}
	}()
	for _, deploy := range param.PluginDeployments {
		if err := mgr.conf.StoragePlugin.CreatePluginDeployment(nCtx, deploy); err != nil {
			return "", fmt.Errorf("failed to create composition deployment: %w", err)
		}
		def := plugin.NewOperComposePluginSubConfig(plugin.OperParamComposePluginSubConfig{
			PluginActionStandardParam: pluginUtils.PluginActionStandardParam{
				Token: deploy.Token, TenantID: nCtx.TenantID(), Operator: param.Operator,
			}, SourceOperation: param.SourceOperation,
		})
		if _, err := triggerCtl.CreateOperation(nCtx, def, def.DefaultParameters()); err != nil {
			return "", fmt.Errorf("failed to create composition operation: %w", err)
		}
	}
	if err := triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", fmt.Errorf("failed to activate composition trigger: %w", err)
	}
	needRollback = false

	return workflowID, nil
}

func validateCompositionDeployments(nCtx contextx.IContext, deployments []*types.PluginDeployment) error {
	if len(deployments) == 0 {
		return fmt.Errorf("composition deployments are required")
	}
	for _, deploy := range deployments {
		if deploy == nil || deploy.Info == nil || deploy.Token == "" {
			return fmt.Errorf("invalid composition deployment")
		}
		info := deploy.Info
		if info.Process.TenantID != nCtx.TenantID() || info.Process.HostID <= 0 || info.Process.PluginName == "" {
			return fmt.Errorf("invalid composition source identity")
		}
		if info.ApplyPluginSubConfigOptions.PluginName == "" || info.ApplyPluginSubConfigOptions.Set == "" {
			return fmt.Errorf("target plugin name and config set are required")
		}
	}

	return nil
}
