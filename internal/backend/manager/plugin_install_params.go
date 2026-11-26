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
	"slices"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// GetPluginWorkflowInstallParams retrieves historical installation form inputs without executing an operation.
func (mgr *Manager) GetPluginWorkflowInstallParams(nCtx contextx.IContext, workflowID string, operationIDs []string) (
	*types.PluginWorkflowInstallParamsResult, error) {
	if nCtx == nil || strings.TrimSpace(workflowID) == "" || len(operationIDs) == 0 {
		return nil, fmt.Errorf("context, workflow_id and operation_ids are required")
	}
	selected := make(map[string]struct{}, len(operationIDs))
	for _, id := range operationIDs {
		if strings.TrimSpace(id) == "" {
			return nil, fmt.Errorf("operation_id must not be empty")
		}
		if _, exists := selected[id]; exists {
			return nil, fmt.Errorf("duplicate operation_id")
		}
		selected[id] = struct{}{}
	}

	workflow, err := mgr.conf.StoragePlugin.GetPluginWorkflow(nCtx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin workflow: %w", err)
	}
	if workflow.TenantID != nCtx.TenantID() {
		return nil, fmt.Errorf("workflow does not belong to the current tenant")
	}
	if workflow.Type != types.PluginWorkflowTypeInstall {
		return nil, fmt.Errorf("only plugin installation workflows support copying installation parameters")
	}

	operations, _, err := mgr.conf.StorageWorkflow.ListOperationByOperationID(nCtx, operationIDs...)
	if err != nil {
		return nil, fmt.Errorf("failed to list selected operations: %w", err)
	}
	tokens, err := pluginInstallOperationTokens(operations, operationIDs, workflow.TriggerID, nCtx.TenantID())
	if err != nil {
		return nil, fmt.Errorf("failed to resolve installation operations: %w", err)
	}
	deployments, _, err := mgr.conf.StoragePlugin.ListPluginDeployment(nCtx, types.UnlimitedPage(), &types.PluginDeploymentCondition{
		ExactInclude: &types.PluginDeploymentExactFields{Token: tokens},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list installation deployments: %w", err)
	}
	return recoverPluginInstallParams(workflow, operationIDs, tokens, deployments)
}

func pluginInstallOperationTokens(operations []*operation.Operation, operationIDs []string, triggerID, tenantID string) ([]string, error) {
	byID := make(map[string]*operation.Operation, len(operations))
	for _, op := range operations {
		if op == nil || op.TriggerID != triggerID {
			return nil, fmt.Errorf("selected operation does not belong to the workflow")
		}
		byID[op.OperationID] = op
	}
	tokens := make([]string, len(operationIDs))
	for i, id := range operationIDs {
		op := byID[id]
		if op == nil {
			return nil, fmt.Errorf("selected operation does not belong to the workflow")
		}
		param := new(utils.PluginActionStandardParam)
		if err := conv.MapToStruct(op.Param.InitContent, param); err != nil {
			return nil, fmt.Errorf("failed to decode installation operation parameters: %w", err)
		}
		if param.Token == "" || param.TenantID != tenantID {
			return nil, fmt.Errorf("invalid installation operation parameters")
		}
		tokens[i] = param.Token
	}
	return tokens, nil
}

func recoverPluginInstallParams(workflow *types.PluginWorkflow, operationIDs, tokens []string,
	deployments []*types.PluginDeployment) (*types.PluginWorkflowInstallParamsResult, error) {
	byToken := make(map[string]*types.PluginDeployment, len(deployments))
	for _, deployment := range deployments {
		if deployment == nil || deployment.Info == nil {
			continue
		}
		process := deployment.Info.Process
		if process.TenantID != workflow.TenantID || !slices.Contains(workflow.BizIDs, process.BizID) {
			return nil, fmt.Errorf("deployment does not belong to the workflow scope")
		}
		byToken[deployment.Token] = deployment
	}

	result := &types.PluginWorkflowInstallParamsResult{
		Type:                    workflow.Type,
		Items:                   make([]*types.PluginWorkflowInstallParams, 0, len(operationIDs)),
		UnavailableOperationIDs: make([]string, 0),
	}
	for i, id := range operationIDs {
		deployment := byToken[tokens[i]]
		if deployment == nil {
			result.UnavailableOperationIDs = append(result.UnavailableOperationIDs, id)
			continue
		}
		result.Items = append(result.Items, recoverPluginInstallParam(id, deployment))
	}
	return result, nil
}

func recoverPluginInstallParam(operationID string, deployment *types.PluginDeployment) *types.PluginWorkflowInstallParams {
	process, options := deployment.Info.Process, deployment.Info.InstallOptions
	item := &types.PluginWorkflowInstallParams{
		OperationID:   operationID,
		HostID:        process.HostID,
		PluginName:    process.PluginName,
		Version:       options.Version,
		CustomSpec:    options.CustomSpec,
		MissingFields: make([]string, 0),
	}
	if process.HostID == 0 {
		item.MissingFields = append(item.MissingFields, "bk_host_id")
	}
	if process.PluginName == "" {
		item.MissingFields = append(item.MissingFields, "plugin_name")
	}
	if options.Version == "" {
		item.MissingFields = append(item.MissingFields, "version")
	}
	if deployment.PluginConf == nil {
		item.MissingFields = append(item.MissingFields, "custom_config_context")
		return item
	}
	item.CustomConfigContext = deployment.PluginConf.CustomConfigContext
	return item
}
