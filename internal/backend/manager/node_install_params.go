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

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// GetNodeWorkflowInstallParams retrieves installation form inputs without launching or modifying any operation.
// The request boundary authorizes historical workflow access; this method additionally verifies operation ownership.
func (mgr *Manager) GetNodeWorkflowInstallParams(nCtx contextx.IContext, workflowID string, operationIDs []string) (
	*types.NodeWorkflowInstallParamsResult, error) {
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

	workflow, err := mgr.conf.StorageNode.GetNodeWorkflow(nCtx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node workflow: %w", err)
	}
	if workflow.TenantID != nCtx.TenantID() {
		return nil, fmt.Errorf("workflow does not belong to the current tenant")
	}
	if workflow.Type != types.NodeWorkflowTypeInstallAgent && workflow.Type != types.NodeWorkflowTypeInstallProxy {
		return nil, fmt.Errorf("only agent and proxy installation workflows support copying installation parameters")
	}

	operations, _, err := mgr.conf.StorageWorkflow.ListOperationByOperationID(nCtx, operationIDs...)
	if err != nil {
		return nil, fmt.Errorf("failed to list selected operations: %w", err)
	}
	tokens, err := installOperationTokens(operations, operationIDs, workflow.TriggerID)
	if err != nil {
		return nil, err
	}
	deployments, _, err := mgr.conf.StorageNode.ListNodeDeployment(nCtx, types.UnlimitedPage(), &types.NodeDeploymentCondition{
		ExactInclude: &types.NodeDeploymentExactFields{Token: tokens},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list installation deployments: %w", err)
	}
	byToken := make(map[string]*types.DeploymentInfo, len(deployments))
	for _, deployment := range deployments {
		if deployment == nil || deployment.Info == nil {
			continue
		}
		if deployment.Info.Host.TenantID != nCtx.TenantID() {
			return nil, fmt.Errorf("deployment does not belong to the current tenant")
		}
		byToken[deployment.Token] = deployment.Info
	}

	result := &types.NodeWorkflowInstallParamsResult{
		Type:                    workflow.Type,
		Items:                   make([]*types.NodeWorkflowInstallParams, 0, len(operationIDs)),
		UnavailableOperationIDs: make([]string, 0),
	}
	for i, id := range operationIDs {
		info := byToken[tokens[i]]
		if info == nil || info.Host.Static == nil || info.Host.Dynamic == nil {
			result.UnavailableOperationIDs = append(result.UnavailableOperationIDs, id)
			continue
		}
		result.Items = append(result.Items, recoverNodeInstallParams(id, info))
	}
	return result, nil
}

func installOperationTokens(operations []*operation.Operation, operationIDs []string, triggerID string) ([]string, error) {
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
		param := new(utils.NodeActionStandardParam)
		if err := conv.MapToStruct(op.Param.InitContent, param); err != nil {
			return nil, fmt.Errorf("failed to decode installation operation parameters: %w", err)
		}
		if param.Token == "" {
			return nil, fmt.Errorf("installation operation has no deployment token")
		}
		tokens[i] = param.Token
	}
	return tokens, nil
}

func recoverNodeInstallParams(operationID string, info *types.DeploymentInfo) *types.NodeWorkflowInstallParams {
	static, dynamic, options := info.Host.Static, info.Host.Dynamic, info.InstallOptions
	item := &types.NodeWorkflowInstallParams{
		OperationID: operationID,
		Host: types.NodeWorkflowInstallHost{
			HostID: info.Host.HostID, BizID: static.BizID, NetworkAreaID: static.NetworkAreaID,
			NetworkUnitID: dynamic.NetworkUnitID, Addressing: static.Addressing,
			InnerIPList: slices.Clone(static.InnerIPList), InnerIPV6List: slices.Clone(static.InnerIPV6List),
			OSType: static.OSType, LoginIP: dynamic.LoginIP, LoginPort: dynamic.LoginPort,
			LoginUser: dynamic.LoginUser, LoginMode: dynamic.LoginMode, InstallMethod: options.InstallMethod,
			IsSudo: options.IsSudo, ReRegister: options.ReRegister, InstallPreOrderedPlugins: options.InstallPreOrderedPlugins,
			RenewGSETask: options.RenewGSETask, RenewGSEProc: options.RenewGSEProc,
			ExportIP: dynamic.ExportIP, ExportIPV6: dynamic.ExportIPV6,
			AdvertiseIP: dynamic.AdvertiseIP, AdvertiseIPV6: dynamic.AdvertiseIPV6,
			ProxyTags: slices.Clone(dynamic.ProxyTags), ProxyInstallOriginUnitID: dynamic.ProxyInstallOriginUnitID,
			RelayDownloadPort: dynamic.RelayDownloadPort, RelayCallbackPort: dynamic.RelayCallbackPort,
			CPUArch: string(dynamic.NodeCPUArch),
		},
		TargetVersion: slices.Clone(info.TargetVersion), IsManual: options.IsManual, IsOffline: options.IsOffline,
		// Deployment records are runtime data, not complete original request snapshots.
		// IPv6 addresses were not persisted, so their absence is not an original empty selection.
		MissingFields: []string{"bk_host_innerip_v6"},
	}
	if options.InstallMethod == "" {
		item.MissingFields = append(item.MissingFields, "install_method")
	}
	if len(info.TargetVersion) == 0 {
		item.MissingFields = append(item.MissingFields, "target_version")
	}
	return item
}
