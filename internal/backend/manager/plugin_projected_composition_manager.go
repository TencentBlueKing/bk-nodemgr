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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

// LaunchComposeProjectedPluginSubConfigs launches one batch operation for a single execution host.
// Target updates are serialized only within this batch, not across independent calls.
func (mgr *Manager) LaunchComposeProjectedPluginSubConfigs(
	nCtx contextx.IContext,
	param types.ComposeProjectedPluginSubConfigsParam,
) (string, error) {

	if err := validateProjectionDeployments(nCtx, param); err != nil {
		return "", err
	}
	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", fmt.Errorf("failed to create projection trigger: %w", err)
	}
	workflowID := identifier.GenWorkflowID()
	bizIDs := make([]int64, 0, len(param.PluginDeployments))
	for _, deploy := range param.PluginDeployments {
		bizIDs = append(bizIDs, deploy.Info.Process.BizID)
	}
	if err := mgr.conf.StoragePlugin.CreatePluginWorkflow(nCtx, &types.PluginWorkflow{
		TenantID: nCtx.TenantID(), WorkflowID: workflowID, TriggerID: triggerCtl.GetTriggerID(),
		Type: types.PluginWorkflowTypeComposeProjectedPluginSubConfigs, HostIDs: []int64{param.HostID}, BizIDs: conv.SliceUnique(bizIDs),
		Operator: param.Operator, DeployPolicyIDs: param.DeployPolicyIDs,
		OperateTime: time.Now(), Status: types.PluginWorkflowStatusRunning,
	}); err != nil {
		return "", fmt.Errorf("failed to create projection workflow: %w", err)
	}
	needRollback := true
	defer func() {
		if !needRollback {
			return
		}
		if err := triggerCtl.InactivateTrigger(nCtx); err != nil {
			logger.G.Biz(nCtx).WithErr(err).With("trigger-id", triggerCtl.GetTriggerID()).Error("failed to inactivate projection trigger")
		}
		if err := mgr.conf.StoragePlugin.UpdatePluginWorkflowStatus(nCtx, workflowID, types.PluginWorkflowStatusFailed); err != nil {
			logger.G.Biz(nCtx).WithErr(err).With("workflow-id", workflowID).Error("failed to mark projection workflow as failed")
		}
	}()
	tokens := make([]string, 0, len(param.PluginDeployments))
	for _, deploy := range param.PluginDeployments {
		if err := mgr.conf.StoragePlugin.CreatePluginDeployment(nCtx, deploy); err != nil {
			return "", fmt.Errorf("failed to create projection deployment: %w", err)
		}
		tokens = append(tokens, deploy.Token)
	}
	def := plugin.NewOperComposeProjectedPluginSubConfigs(plugin.OperParamComposeProjectedPluginSubConfigs{
		TenantID: nCtx.TenantID(), HostID: param.HostID, Operator: param.Operator,
		SourceTokens: tokens, SourceOperation: param.SourceOperation,
	})
	if _, err := triggerCtl.CreateOperation(nCtx, def, def.DefaultParameters()); err != nil {
		return "", fmt.Errorf("failed to create projection operation: %w", err)
	}
	if err := triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", fmt.Errorf("failed to activate projection trigger: %w", err)
	}
	needRollback = false

	return workflowID, nil
}

func validateProjectionDeployments(nCtx contextx.IContext, param types.ComposeProjectedPluginSubConfigsParam) error {
	if nCtx.TenantID() == "" || param.HostID <= 0 {
		return fmt.Errorf("projection tenant and execution host are required")
	}
	if err := param.SourceOperation.Validate(); err != nil {
		return err
	}
	if err := validateCompositionDeployments(nCtx, param.PluginDeployments); err != nil {
		return err
	}
	sources := make(map[string]bool, len(param.PluginDeployments))
	tokens := make(map[string]bool, len(param.PluginDeployments))
	targetSets := make(map[string]map[string]bool)
	for _, deploy := range param.PluginDeployments {
		info := deploy.Info
		if info.Process.HostID != param.HostID {
			return fmt.Errorf("projection sources must share the execution host")
		}
		if sources[info.Process.PluginName] || tokens[deploy.Token] {
			return fmt.Errorf("duplicate projection source identity or token")
		}
		sources[info.Process.PluginName], tokens[deploy.Token] = true, true
		options := info.ApplyPluginSubConfigOptions
		if targetSets[options.PluginName] == nil {
			targetSets[options.PluginName] = make(map[string]bool)
		}
		if targetSets[options.PluginName][options.Set] {
			return fmt.Errorf("duplicate projection target config set")
		}
		targetSets[options.PluginName][options.Set] = true
	}

	return nil
}
