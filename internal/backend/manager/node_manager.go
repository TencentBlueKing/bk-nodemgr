/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package manager

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

// LaunchInstallNode launch a task to install node.
func (mgr *Manager) LaunchInstallNode(nCtx contextx.IContext, param types.InstallNodeParam) (string, error) {
	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	areaIDs, unitIDs, nodeRoles := collectDeploymentIDs(param.NodeDeployments)
	if err = mgr.conf.StorageNode.CreateNodeWorkflow(nCtx, &types.NodeWorkflow{
		TenantID:        nCtx.TenantID(),
		WorkflowID:      workflowID,
		TriggerID:       triggerCtl.GetTriggerID(),
		Type:            param.Type,
		BizIDs:          param.BizIDs,
		NodeRoles:       nodeRoles,
		NetworkAreaIDs:  areaIDs,
		NetworkUnitIDs:  unitIDs,
		Operator:        param.Operator,
		DeployPolicyIDs: param.DeployPolicyIDs,
		OperateTime:     time.Now(),
		Status:          types.NodeWorkflowStatusRunning,
	}); err != nil {
		return "", err
	}

	gp := gopool.NewPool()
	for _, nodeDeploy := range param.NodeDeployments {
		deploy := nodeDeploy

		gp.Go(func() error {
			return mgr.createInstallNodeOper(nCtx, param.Operator, triggerCtl, deploy)
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch install node task. err: %w", err)
	}

	if err = triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", err
	}

	return workflowID, nil
}

// LaunchRetryNodeOperationFromLastInstance launch a task to retry operation from last instance.
func (mgr *Manager) LaunchRetryNodeOperationFromLastInstance(nCtx contextx.IContext, param types.RetryNodeWorkflowOperationParam) error {
	nodeWorkflow, err := mgr.conf.StorageNode.GetNodeWorkflow(nCtx, param.WorkflowID)
	if err != nil {
		return fmt.Errorf("failed to get node workflow: %w", err)
	}

	triggerCtl, err := mgr.workflowMgr.GetTrigger(nCtx, nodeWorkflow.TriggerID)
	if err != nil {
		return fmt.Errorf("failed to get trigger: %w", err)
	}

	if err := triggerCtl.UpdateOperationRetryFlag(nCtx, param.RetryMod, param.OperationIDs...); err != nil {
		return fmt.Errorf("failed to update operation retry flag: %w", err)
	}

	if err := mgr.conf.StorageNode.UpdateNodeWorkflowStatus(nCtx, param.WorkflowID, types.NodeWorkflowStatusRunning); err != nil {
		return fmt.Errorf("failed to update node workflow status: %w", err)
	}

	return triggerCtl.ActivateTrigger(nCtx)
}

// TerminateNodeOperationLastInstance terminate operation from last instance.
func (mgr *Manager) TerminateNodeOperationLastInstance(nCtx contextx.IContext, param types.TerminateNodeWorkflowOperationParam) error {
	nodeWorkflow, err := mgr.conf.StorageNode.GetNodeWorkflow(nCtx, param.WorkflowID)
	if err != nil {
		return fmt.Errorf("failed to get node workflow: %w", err)
	}

	triggerCtl, err := mgr.workflowMgr.GetTrigger(nCtx, nodeWorkflow.TriggerID)
	if err != nil {
		return fmt.Errorf("failed to get trigger: %w", err)
	}

	operationCtls, err := triggerCtl.ListOperation(nCtx, param.OperationIDs...)
	if err != nil {
		return fmt.Errorf("failed to list operation: %w", err)
	}

	gp := gopool.NewPool()
	for _, operationCtl := range operationCtls {
		opCtl := operationCtl

		gp.Go(func() error {
			instanceCtl, err := opCtl.GetLastOperationInstance(nCtx)
			if err != nil {
				return fmt.Errorf("failed to get last operation instance: %w", err)
			}

			return instanceCtl.TerminateOperationInstance(nCtx)
		})
	}

	if err := gp.Wait(); err != nil {
		return fmt.Errorf("failed to terminate node operation: %w", err)
	}

	return nil
}

// GetOperationManualInfoFromLastInstance get operation manual info from last instance.
func (mgr *Manager) GetOperationManualInfoFromLastInstance(nCtx contextx.IContext, param types.GetNodeWorklfowOperationManualInfoParam) (
	*types.NodeWorkflowOperationManualInfo, error) {

	operation, err := mgr.conf.StorageWorkflow.GetOperation(nCtx, param.OperationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get operation: %w", err)
	}

	if len(operation.InstanceIDs) == 0 {
		return nil, fmt.Errorf("operation has no instances")
	}

	lastOperInstID := operation.InstanceIDs[len(operation.InstanceIDs)-1]

	privateData, err := mgr.conf.StorageWorkflow.GetActionInstancePrivateData(nCtx, lastOperInstID, node.ActionNameGenManualBootstrapCommand)
	if err != nil {
		return nil, fmt.Errorf("failed to get action instance private data: %w", err)
	}

	// generate bootstrap commands.
	commands := make([]*types.NodeWorkflowOperationManualCommand, 0)
	raw, ok := privateData[types.PDKeyManualInstallBootstrapCommandBash]
	if ok {
		bootstrapCommand, err := conv.ToString(raw)
		if err != nil {
			return nil, fmt.Errorf("private data does not contain valid bootstrap command bash(%v): %w", raw, err)
		}

		commands = append(commands, &types.NodeWorkflowOperationManualCommand{
			Type:    types.NodeWorkflowOperationManualCommandTypeBash,
			Command: bootstrapCommand,
		})
	}
	raw, ok = privateData[types.PDKeyManualInstallBootstrapCommandBat]
	if ok {
		bootstrapCommand, err := conv.ToString(raw)
		if err != nil {
			return nil, fmt.Errorf("private data does not contain valid bootstrap command bat(%v): %w", raw, err)
		}

		commands = append(commands, &types.NodeWorkflowOperationManualCommand{
			Type:    types.NodeWorkflowOperationManualCommandTypeBat,
			Command: bootstrapCommand,
		})
	}
	if len(commands) == 0 {
		return nil, fmt.Errorf("private data does not contain bootstrap command")
	}

	return &types.NodeWorkflowOperationManualInfo{
		Commands: commands,
	}, nil
}

func (mgr *Manager) createInstallNodeOper(
	nCtx contextx.IContext,
	operator string,
	triggerCtl workflow.ITriggerCtl,
	deploy *types.NodeDeployment,
) error {

	if err := mgr.conf.StorageNode.CreateNodeDeployment(nCtx, deploy); err != nil {
		logger.G.Biz(nCtx).
			WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID(), "token", deploy.Token).
			Error("failed to create node deployment")

		return err
	}

	operationDef, err := mgr.getNodeInstallOperationDef(deploy, operator)
	if err != nil {
		return fmt.Errorf("failed to get node install operation def: %w", err)
	}

	operationParam := operationDef.DefaultParameters()

	operCtl, err := triggerCtl.CreateOperation(nCtx, operationDef, operationParam)
	if err != nil {
		logger.G.Biz(nCtx).
			WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID(), "operation-id", operCtl.GetOperationID(), "token", deploy.Token).
			Error("failed to launch install node task")

		return err
	}

	logger.G.Biz(nCtx).
		With("trigger-id", triggerCtl.GetTriggerID(), "operation-id", operCtl.GetOperationID(), "token", deploy.Token).
		Info("launched install node task")

	return nil
}

func (mgr *Manager) getNodeInstallOperationDef(deploy *types.NodeDeployment, operator string) (operation.Definition, error) {
	// proxy use user selected install origin network unit id to select relay host.
	if deploy.Info.Host.Dynamic.NodeRole == types.NodeRoleProxy {
		return mgr.getNodeInstallOperationDefProxy(deploy, operator)
	}

	// agent use system selected relay host in its own network unit.
	return mgr.getNodeInstallOperationDefAgent(deploy, operator)
}

// agent install distinguish direct install or pagent install.
// nolint: gocognit, cyclop
func (mgr *Manager) getNodeInstallOperationDefAgent(deploy *types.NodeDeployment, operator string) (operation.Definition, error) {
	if deploy.Info.InstallOptions.IsManual {
		if deploy.Info.InstallOptions.DirectInstall {
			return node.NewOperInstallNodeByManual(node.OperParamInstallNodeByManual{
				Token:    deploy.Token,
				Operator: operator,
			}), nil
		}

		return node.NewOperInstallPagentByManual(node.OperParamInstallPagentByManual{
			Token:    deploy.Token,
			Operator: operator,
		}), nil
	}

	// direct install.
	if deploy.Info.InstallOptions.DirectInstall {
		switch criteria.OSType(deploy.Info.Host.Static.OSType) {
		case criteria.OSLinux, criteria.OSDarwin:
			switch deploy.Info.InstallOptions.InstallMethod {
			case types.NodeInstallMethodAuto, types.NodeInstallMethodSSH:
				return node.NewOperInstallNodeBySSH(node.OperParamInstallNodeBySSH{
					Token:    deploy.Token,
					Operator: operator,
				}), nil
			default:
				return nil, fmt.Errorf("install method is unsupported for this os type, install_method(%s), os_type(%s)",
					deploy.Info.InstallOptions.InstallMethod, deploy.Info.Host.Static.OSType)
			}
		case criteria.OSWindows:
			switch deploy.Info.InstallOptions.InstallMethod {
			case types.NodeInstallMethodAuto, types.NodeInstallMethodSSH:
				return node.NewOperInstallNodeByWindowsSSH(node.OperParamInstallNodeByWindowsSSH{
					Token:    deploy.Token,
					Operator: operator,
				}), nil
			case types.NodeInstallMethodWMI:
				return node.NewOperInstallNodeByWMI(node.OperParamInstallNodeByWMI{
					Token:    deploy.Token,
					Operator: operator,
				}), nil
			default:
				return nil, fmt.Errorf("install method is unsupported for this os type, install_method(%s), os_type(%s)",
					deploy.Info.InstallOptions.InstallMethod, deploy.Info.Host.Static.OSType)
			}

		default:
			switch deploy.Info.InstallOptions.InstallMethod {
			case types.NodeInstallMethodAuto, types.NodeInstallMethodSSH:
				return node.NewOperInstallNodeBySSH(node.OperParamInstallNodeBySSH{
					Token:    deploy.Token,
					Operator: operator,
				}), nil
			default:
				return nil, fmt.Errorf("install method is unsupported for this os type, install_method(%s), os_type(%s)",
					deploy.Info.InstallOptions.InstallMethod, deploy.Info.Host.Static.OSType)
			}
		}
	}

	// install by relay.
	switch criteria.OSType(deploy.Info.Host.Static.OSType) {
	case criteria.OSLinux, criteria.OSDarwin:
		switch deploy.Info.InstallOptions.InstallMethod {
		case types.NodeInstallMethodAuto, types.NodeInstallMethodSSH:
			return node.NewOperInstallPagentNodeBySSH(node.OperParamInstallPagentNodeBySSH{
				Token:    deploy.Token,
				Operator: operator,
			}), nil
		default:
			return nil, fmt.Errorf("install method is unsupported for this os type, install_method(%s), os_type(%s)",
				deploy.Info.InstallOptions.InstallMethod, deploy.Info.Host.Static.OSType)
		}

	case criteria.OSWindows:
		switch deploy.Info.InstallOptions.InstallMethod {
		case types.NodeInstallMethodAuto, types.NodeInstallMethodSSH:
			return node.NewOperInstallPagentNodeByWindowsSSH(node.OperParamInstallPagentNodeByWindowsSSH{
				Token:    deploy.Token,
				Operator: operator,
			}), nil
		case types.NodeInstallMethodWMI:
			return node.NewOperInstallPagentNodeByWMI(node.OperParamInstallPagentNodeByWMI{
				Token:    deploy.Token,
				Operator: operator,
			}), nil
		default:
			return nil, fmt.Errorf("install method is unsupported for this os type, install_method(%s), os_type(%s)",
				deploy.Info.InstallOptions.InstallMethod, deploy.Info.Host.Static.OSType)
		}

	default:
		switch deploy.Info.InstallOptions.InstallMethod {
		case types.NodeInstallMethodAuto, types.NodeInstallMethodSSH:
			return node.NewOperInstallPagentNodeBySSH(node.OperParamInstallPagentNodeBySSH{
				Token:    deploy.Token,
				Operator: operator,
			}), nil
		default:
			return nil, fmt.Errorf("install method is unsupported for this os type, install_method(%s), os_type(%s)",
				deploy.Info.InstallOptions.InstallMethod, deploy.Info.Host.Static.OSType)
		}
	}
}

// proxy install distinguish direct install or relay install, and manual or automated.
func (mgr *Manager) getNodeInstallOperationDefProxy(deploy *types.NodeDeployment, operator string) (operation.Definition, error) {
	if deploy.Info.InstallOptions.IsOffline {
		switch deploy.Info.InstallOptions.InstallMethod {
		case types.NodeInstallMethodAuto, types.NodeInstallMethodSSH:
			return node.NewOperInstallProxyByOffline(node.OperParamInstallProxyByOffline{
				Token:    deploy.Token,
				Operator: operator,
			}), nil
		default:
			return nil, fmt.Errorf("install method is unsupported for this os type, install_method(%s), os_type(%s)",
				deploy.Info.InstallOptions.InstallMethod, deploy.Info.Host.Static.OSType)
		}
	}

	if deploy.Info.InstallOptions.IsManual {
		if deploy.Info.InstallOptions.DirectInstall {
			switch deploy.Info.InstallOptions.InstallMethod {
			case types.NodeInstallMethodAuto, types.NodeInstallMethodSSH:
				return node.NewOperInstallNodeByManual(node.OperParamInstallNodeByManual{
					Token:    deploy.Token,
					Operator: operator,
				}), nil
			default:
				return nil, fmt.Errorf("install method is unsupported for this os type, install_method(%s), os_type(%s)",
					deploy.Info.InstallOptions.InstallMethod, deploy.Info.Host.Static.OSType)
			}
		}

		switch deploy.Info.InstallOptions.InstallMethod {
		case types.NodeInstallMethodAuto, types.NodeInstallMethodSSH:
			return node.NewOperInstallPagentByManual(node.OperParamInstallPagentByManual{
				Token:    deploy.Token,
				Operator: operator,
			}), nil
		default:
			return nil, fmt.Errorf("install method is unsupported for this os type, install_method(%s), os_type(%s)",
				deploy.Info.InstallOptions.InstallMethod, deploy.Info.Host.Static.OSType)
		}
	}

	if deploy.Info.InstallOptions.DirectInstall {
		switch deploy.Info.InstallOptions.InstallMethod {
		case types.NodeInstallMethodAuto, types.NodeInstallMethodSSH:
			return node.NewOperInstallNodeBySSH(node.OperParamInstallNodeBySSH{
				Token:    deploy.Token,
				Operator: operator,
			}), nil
		default:
			return nil, fmt.Errorf("install method is unsupported for this os type, install_method(%s), os_type(%s)",
				deploy.Info.InstallOptions.InstallMethod, deploy.Info.Host.Static.OSType)
		}
	}
	switch deploy.Info.InstallOptions.InstallMethod {
	case types.NodeInstallMethodAuto, types.NodeInstallMethodSSH:
		return node.NewOperInstallProxyBySSH(node.OperParamInstallProxyBySSH{
			Token:    deploy.Token,
			Operator: operator,
		}), nil
	default:
		return nil, fmt.Errorf("install method is unsupported for this os type, install_method(%s), os_type(%s)",
			deploy.Info.InstallOptions.InstallMethod, deploy.Info.Host.Static.OSType)
	}
}

// LaunchUpgradeNode launch a task to upgrade node. returns the workflow-id.
func (mgr *Manager) LaunchUpgradeNode(nCtx contextx.IContext, param types.UpgradeNodeParam) (string, error) {
	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	areaIDs, unitIDs, nodeRoles := collectDeploymentIDs(param.NodeDeployments)
	if err = mgr.conf.StorageNode.CreateNodeWorkflow(nCtx, &types.NodeWorkflow{
		TenantID:        nCtx.TenantID(),
		WorkflowID:      workflowID,
		TriggerID:       triggerCtl.GetTriggerID(),
		Type:            param.Type,
		BizIDs:          param.BizIDs,
		NodeRoles:       nodeRoles,
		NetworkAreaIDs:  areaIDs,
		NetworkUnitIDs:  unitIDs,
		Operator:        param.Operator,
		DeployPolicyIDs: param.DeployPolicyIDs,
		OperateTime:     time.Now(),
		Status:          types.NodeWorkflowStatusRunning,
	}); err != nil {
		return "", err
	}

	logger.G.Biz(nCtx).
		With("trigger-id", triggerCtl.GetTriggerID(), "node-deployments", len(param.NodeDeployments)).
		Info("launching upgrade node task")

	gp := gopool.NewPool()
	for _, nodeDeploy := range param.NodeDeployments {
		deploy := nodeDeploy

		gp.Go(func() error {
			operationDef := mgr.getUpgradeOperationDef(deploy, param.Operator)
			enablePagentInstaller(deploy)

			operationParam := operationDef.DefaultParameters()

			if err := mgr.conf.StorageNode.CreateNodeDeployment(nCtx, deploy); err != nil {
				logger.G.Biz(nCtx).
					WithErr(err).
					With("trigger-id", triggerCtl.GetTriggerID(), "token", deploy.Token).
					Error("failed to launch upgrade node, failed to create node deployment")

				return err
			}

			operCtl, err := triggerCtl.CreateOperation(nCtx, operationDef, operationParam)
			if err != nil {
				logger.G.Biz(nCtx).
					WithErr(err).
					With("trigger-id", triggerCtl.GetTriggerID(), "operation-id", operCtl.GetOperationID(), "token", deploy.Token).
					Error("failed to launch upgrade node, failed to create operation")

				return err
			}

			logger.G.Biz(nCtx).
				With("trigger-id", triggerCtl.GetTriggerID(), "operation-id", operCtl.GetOperationID(), "token", deploy.Token).
				Info("launched upgrade node")

			return nil
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch upgrade node task. err: %w", err)
	}

	if err = triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", err
	}

	return workflowID, nil
}

func (mgr *Manager) getUpgradeOperationDef(deploy *types.NodeDeployment, operator string) operation.Definition {
	// proxy node, use its own relay callback.
	if deploy.Info.Host.Dynamic.NodeRole == types.NodeRoleProxy {
		return node.NewOperUpgradeProxy(node.OperParamUpgradeProxy{
			Token:    deploy.Token,
			Operator: operator,
		})
	}

	// direct link, use agent.
	if deploy.Info.UpgradeOptions.DirectLink {
		return node.NewOperUpgradeNode(node.OperParamUpgradeNode{
			Token:    deploy.Token,
			Operator: operator,
		})
	}

	// indirect link, use pagent.
	return node.NewOperUpgradePagent(node.OperParamUpgradePagent{
		Token:    deploy.Token,
		Operator: operator,
	})
}

func enablePagentInstaller(deploy *types.NodeDeployment) {
	if deploy.Info.Host.Dynamic.NodeRole == types.NodeRoleProxy {
		return
	}

	if deploy.Info.UpgradeOptions.DirectLink {
		return
	}

	// if not direct link, use pagent. than we only need to transfer installer.
	deploy.Info.TransferOptions = types.DeploymentTransferOptionsOnlyTransferInstaller()
}

// LaunchReconfigNode launch a task to reconfig node. returns the workflow-id.
func (mgr *Manager) LaunchReconfigNode(nCtx contextx.IContext, param types.ReconfigNodeParam) (string, error) {
	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	areaIDs, unitIDs, nodeRoles := collectDeploymentIDs(param.NodeDeployments)
	if err = mgr.conf.StorageNode.CreateNodeWorkflow(nCtx, &types.NodeWorkflow{
		TenantID:       nCtx.TenantID(),
		WorkflowID:     workflowID,
		TriggerID:      triggerCtl.GetTriggerID(),
		Type:           param.Type,
		BizIDs:         param.BizIDs,
		NodeRoles:      nodeRoles,
		NetworkAreaIDs: areaIDs,
		NetworkUnitIDs: unitIDs,
		Operator:       param.Operator,
		OperateTime:    time.Now(),
		Status:         types.NodeWorkflowStatusRunning,
	}); err != nil {
		return "", err
	}

	logger.G.Biz(nCtx).
		With("trigger-id", triggerCtl.GetTriggerID(), "node-deployments", len(param.NodeDeployments)).
		Info("launching reconfig node")

	gp := gopool.NewPool()
	for _, nodeDeploy := range param.NodeDeployments {
		deploy := nodeDeploy
		deploy.Info.TransferOptions = types.DeploymentTransferOptionsOnlyTransferInstaller()

		gp.Go(func() error {
			if err := mgr.conf.StorageNode.CreateNodeDeployment(nCtx, deploy); err != nil {
				logger.G.Biz(nCtx).
					WithErr(err).
					With("trigger-id", triggerCtl.GetTriggerID(), "token", deploy.Token).
					Error("failed to launch reconfig node, failed to create node deployment")

				return err
			}

			operationDef := mgr.getReconfigOperationDef(deploy, param.Operator)
			operationParam := operationDef.DefaultParameters()

			operCtl, err := triggerCtl.CreateOperation(nCtx, operationDef, operationParam)
			if err != nil {
				logger.G.Biz(nCtx).
					WithErr(err).
					With("trigger-id", triggerCtl.GetTriggerID(), "operation-id", operCtl.GetOperationID(), "token", deploy.Token).
					Error("failed to launch reconfig node, failed to create operation")

				return err
			}

			logger.G.Biz(nCtx).
				With("trigger-id", triggerCtl.GetTriggerID(), "operation-id", operCtl.GetOperationID(), "token", deploy.Token).
				Info("launched reconfig node")

			return nil
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch reconfig node task. err: %w", err)
	}

	if err = triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", err
	}

	return workflowID, nil
}

// LaunchRestartNode launch a task to restart node. returns the workflow-id.
func (mgr *Manager) LaunchRestartNode(nCtx contextx.IContext, param types.RestartNodeParam) (string, error) {
	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	areaIDs, unitIDs, nodeRoles := collectDeploymentIDs(param.NodeDeployments)
	if err = mgr.conf.StorageNode.CreateNodeWorkflow(nCtx, &types.NodeWorkflow{
		TenantID:       nCtx.TenantID(),
		WorkflowID:     workflowID,
		TriggerID:      triggerCtl.GetTriggerID(),
		Type:           param.Type,
		BizIDs:         param.BizIDs,
		NodeRoles:      nodeRoles,
		NetworkAreaIDs: areaIDs,
		NetworkUnitIDs: unitIDs,
		Operator:       param.Operator,
		OperateTime:    time.Now(),
		Status:         types.NodeWorkflowStatusRunning,
	}); err != nil {
		return "", err
	}

	logger.G.Biz(nCtx).
		With("trigger-id", triggerCtl.GetTriggerID(), "node-deployments", len(param.NodeDeployments)).
		Info("launching restart node")

	gp := gopool.NewPool()
	for _, nodeDeploy := range param.NodeDeployments {
		deploy := nodeDeploy
		deploy.Info.TransferOptions = types.DeploymentTransferOptionsOnlyTransferInstaller()

		gp.Go(func() error {
			if err := mgr.conf.StorageNode.CreateNodeDeployment(nCtx, deploy); err != nil {
				logger.G.Biz(nCtx).
					WithErr(err).
					With("trigger-id", triggerCtl.GetTriggerID(), "token", deploy.Token).
					Error("failed to launch restart node, failed to create node deployment")

				return err
			}

			operationDef := node.NewOperRestartNode(node.OperParamRestartNode{
				Token:    deploy.Token,
				Operator: param.Operator,
			})
			operationParam := operationDef.DefaultParameters()

			operCtl, err := triggerCtl.CreateOperation(nCtx, operationDef, operationParam)
			if err != nil {
				logger.G.Biz(nCtx).
					WithErr(err).
					With("trigger-id", triggerCtl.GetTriggerID(), "operation-id", operCtl.GetOperationID(), "token", deploy.Token).
					Error("failed to launch restart node, failed to create operation")

				return err
			}

			logger.G.Biz(nCtx).
				With("trigger-id", triggerCtl.GetTriggerID(), "operation-id", operCtl.GetOperationID(), "token", deploy.Token).
				Info("launched restart node")

			return nil
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch restart node task. err: %w", err)
	}

	if err = triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", err
	}

	return workflowID, nil
}

// LaunchUninstallNode launch a task to uninstall node. returns the workflow-id.
func (mgr *Manager) LaunchUninstallNode(nCtx contextx.IContext, param types.UninstallNodeParam) (string, error) {
	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	areaIDs, unitIDs, nodeRoles := collectDeploymentIDs(param.NodeDeployments)
	if err = mgr.conf.StorageNode.CreateNodeWorkflow(nCtx, &types.NodeWorkflow{
		TenantID:        nCtx.TenantID(),
		WorkflowID:      workflowID,
		TriggerID:       triggerCtl.GetTriggerID(),
		Type:            param.Type,
		BizIDs:          param.BizIDs,
		NodeRoles:       nodeRoles,
		NetworkAreaIDs:  areaIDs,
		NetworkUnitIDs:  unitIDs,
		Operator:        param.Operator,
		DeployPolicyIDs: param.DeployPolicyIDs,
		OperateTime:     time.Now(),
		Status:          types.NodeWorkflowStatusRunning,
	}); err != nil {
		return "", err
	}

	logger.G.Biz(nCtx).
		With("trigger-id", triggerCtl.GetTriggerID(), "node-deployments", len(param.NodeDeployments)).
		Info("launching uninstall node")

	gp := gopool.NewPool()
	for _, nodeDeploy := range param.NodeDeployments {
		deploy := nodeDeploy
		deploy.Info.TransferOptions = types.DeploymentTransferOptionsOnlyTransferInstaller()

		gp.Go(func() error {
			if err := mgr.conf.StorageNode.CreateNodeDeployment(nCtx, deploy); err != nil {
				logger.G.Biz(nCtx).
					WithErr(err).
					With("trigger-id", triggerCtl.GetTriggerID(), "token", deploy.Token).
					Error("failed to launch uninstall node, failed to create node deployment")

				return err
			}

			operationDef := mgr.getUninstallOperationDef(deploy, param.Operator)
			operationParam := operationDef.DefaultParameters()

			operCtl, err := triggerCtl.CreateOperation(nCtx, operationDef, operationParam)
			if err != nil {
				logger.G.Biz(nCtx).
					WithErr(err).
					With("trigger-id", triggerCtl.GetTriggerID(), "operation-id", operCtl.GetOperationID(), "token", deploy.Token).
					Error("failed to launch uninstall node, failed to create operation")

				return err
			}

			logger.G.Biz(nCtx).
				With("trigger-id", triggerCtl.GetTriggerID(), "operation-id", operCtl.GetOperationID(), "token", deploy.Token).
				Info("launched uninstall node")

			return nil
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch uninstall node task: %w", err)
	}

	if err = triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", err
	}

	return workflowID, nil
}

// LaunchAssignProxyUnit launch a task to assign proxy unit. returns the workflow-id.
func (mgr *Manager) LaunchAssignProxyUnit(nCtx contextx.IContext, param types.AssignProxyUnitParam) (string, error) {
	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	areaIDs, unitIDs, nodeRoles := collectDeploymentIDs(param.NodeDeployments)
	if err = mgr.conf.StorageNode.CreateNodeWorkflow(nCtx, &types.NodeWorkflow{
		TenantID:       nCtx.TenantID(),
		WorkflowID:     workflowID,
		TriggerID:      triggerCtl.GetTriggerID(),
		Type:           param.Type,
		BizIDs:         param.BizIDs,
		NodeRoles:      nodeRoles,
		NetworkAreaIDs: areaIDs,
		NetworkUnitIDs: unitIDs,
		Operator:       param.Operator,
		OperateTime:    time.Now(),
		Status:         types.NodeWorkflowStatusRunning,
	}); err != nil {
		return "", err
	}

	logger.G.Biz(nCtx).
		With("trigger-id", triggerCtl.GetTriggerID(), "node-deployments", len(param.NodeDeployments)).
		Info("launching assign proxy unit")

	gp := gopool.NewPool()
	for _, nodeDeploy := range param.NodeDeployments {
		deploy := nodeDeploy

		gp.Go(func() error {
			if err := mgr.conf.StorageNode.CreateNodeDeployment(nCtx, deploy); err != nil {
				logger.G.Biz(nCtx).
					WithErr(err).
					With("trigger-id", triggerCtl.GetTriggerID(), "token", deploy.Token).
					Error("failed to launch assign proxy unit, failed to create node deployment")

				return err
			}

			operationDef := node.NewOperAssignProxyUnit(node.OperParamAssignProxyUnit{
				Token:             deploy.Token,
				Operator:          param.Operator,
				NetworkUnitID:     deploy.Info.Host.Dynamic.NetworkUnitID,
				RelayCallbackPort: deploy.Info.Host.Dynamic.RelayCallbackPort,
				RelayDownloadPort: deploy.Info.Host.Dynamic.RelayDownloadPort,
				ProxyTags:         types.ProxyTagListToStringList(deploy.Info.Host.Dynamic.ProxyTags),
			})
			operationParam := operationDef.DefaultParameters()

			operCtl, err := triggerCtl.CreateOperation(nCtx, operationDef, operationParam)
			if err != nil {
				logger.G.Biz(nCtx).
					WithErr(err).
					With("trigger-id", triggerCtl.GetTriggerID(), "operation-id", operCtl.GetOperationID(), "token", deploy.Token).
					Error("failed to launch assign proxy unit, failed to create operation")

				return err
			}

			logger.G.Biz(nCtx).
				With("trigger-id", triggerCtl.GetTriggerID(), "operation-id", operCtl.GetOperationID(), "token", deploy.Token).
				Info("launched assign proxy unit")

			return nil
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch assign proxy unit task: %w", err)
	}

	if err = triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", err
	}

	return workflowID, nil
}

func (mgr *Manager) getReconfigOperationDef(deploy *types.NodeDeployment, operator string) operation.Definition {
	// proxy node, use its own relay callback.
	if deploy.Info.Host.Dynamic.NodeRole == types.NodeRoleProxy {
		return node.NewOperReconfigProxy(node.OperParamReconfigProxy{
			Token:    deploy.Token,
			Operator: operator,
		})
	}

	// direct link, use direct link.
	if deploy.Info.ReconfigOptions.DirectLink {
		return node.NewOperReconfigNode(node.OperParamReconfigNode{
			Token:    deploy.Token,
			Operator: operator,
		})
	}

	// indirect link, use pagent.
	return node.NewOperReconfigPagent(node.OperParamReconfigPagent{
		Token:    deploy.Token,
		Operator: operator,
	})
}

func (mgr *Manager) getUninstallOperationDef(deploy *types.NodeDeployment, operator string) operation.Definition {
	if deploy.Info.Host.Dynamic.NodeRole == types.NodeRoleProxy {
		return mgr.getProxyUninstallOperationDef(deploy, operator)
	}

	// direct link, use direct link.
	if deploy.Info.UninstallOptions.DirectLink {
		return node.NewOperUninstallNode(node.OperParamUninstallNode{
			Token:    deploy.Token,
			Operator: operator,
		})
	}

	// indirect link, use pagent.
	return node.NewOperUninstallPagent(node.OperParamUninstallPagent{
		Token:    deploy.Token,
		Operator: operator,
	})
}

func (mgr *Manager) getProxyUninstallOperationDef(deploy *types.NodeDeployment, operator string) operation.Definition {
	if deploy.Info.UninstallOptions.SkipReport {
		return node.NewOperUninstallNodeSkipReport(node.OperParamUninstallNode{
			Token:    deploy.Token,
			Operator: operator,
		})
	}

	if deploy.Info.UninstallOptions.DirectLink {
		return node.NewOperUninstallNode(node.OperParamUninstallNode{
			Token:    deploy.Token,
			Operator: operator,
		})
	}

	return node.NewOperUninstallPagent(node.OperParamUninstallPagent{
		Token:    deploy.Token,
		Operator: operator,
	})
}

func collectDeploymentIDs(deployments []*types.NodeDeployment) ([]int64, []int64, []types.NodeRole) {
	var areaIDs, unitIDs []int64
	var nodeRoles []types.NodeRole

	for _, deploy := range deployments {
		if deploy.Info == nil {
			continue
		}

		if deploy.Info.Host.Static != nil {
			areaIDs = append(areaIDs, deploy.Info.Host.Static.NetworkAreaID)
		}

		if deploy.Info.Host.Dynamic != nil {
			unitIDs = append(unitIDs, deploy.Info.Host.Dynamic.NetworkUnitID)
			nodeRoles = append(nodeRoles, deploy.Info.Host.Dynamic.NodeRole)
		}
	}

	return conv.SliceUnique(areaIDs), conv.SliceUnique(unitIDs), conv.SliceUnique(nodeRoles)
}
