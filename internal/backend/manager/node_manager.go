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
	if err = mgr.conf.StorageNode.CreateNodeWorkflow(nCtx, &types.NodeWorkflow{
		TenantID:    nCtx.TenantID(),
		WorkflowID:  workflowID,
		TriggerID:   triggerCtl.GetTriggerID(),
		Type:        param.Type,
		BizIDs:      param.BizIDs,
		Operator:    param.Operator,
		OperateTime: time.Now(),
		Status:      types.NodeWorkflowStatusRunning,
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
		bootstrapCommand, ok := raw.(string)
		if !ok {
			return nil, fmt.Errorf("private data does not contain valid bootstrap command bash: %v", raw)
		}

		commands = append(commands, &types.NodeWorkflowOperationManualCommand{
			Type:    types.NodeWorkflowOperationManualCommandTypeBash,
			Command: bootstrapCommand,
		})
	}
	raw, ok = privateData[types.PDKeyManualInstallBootstrapCommandBat]
	if ok {
		bootstrapCommand, ok := raw.(string)
		if !ok {
			return nil, fmt.Errorf("private data does not contain valid bootstrap command bat: %v", raw)
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

	operationDef := mgr.getNodeInstallOperationDef(deploy, operator)

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

func (mgr *Manager) getNodeInstallOperationDef(deploy *types.NodeDeployment, operator string) operation.Definition {
	// proxy use user selected install origin network unit id to select relay host.
	if deploy.Info.Host.Dynamic.NodeRole == types.NodeRoleProxy {
		return mgr.getNodeInstallOperationDefProxy(deploy, operator)
	}

	// agent use system selected relay host in its own network unit.
	return mgr.getNodeInstallOperationDefAgent(deploy, operator)
}

// agent install distinguish direct install or pagent install.
func (mgr *Manager) getNodeInstallOperationDefAgent(deploy *types.NodeDeployment, operator string) operation.Definition {
	// direct install.
	if deploy.Info.InstallOptions.DirectInstall {
		if deploy.Info.InstallOptions.IsManual {
			return node.NewOperInstallNodeByManual(node.OperParamInstallNodeByManual{
				Token:    deploy.Token,
				Operator: operator,
			})
		}

		switch criteria.OSType(deploy.Info.Host.Static.OSType) {
		case criteria.OSLinux, criteria.OSDarwin:
			return node.NewOperInstallNodeBySSH(node.OperParamInstallNodeBySSH{
				Token:    deploy.Token,
				Operator: operator,
			})

		case criteria.OSWindows:
			return node.NewOperInstallNodeByWMI(node.OperParamInstallNodeByWMI{
				Token:    deploy.Token,
				Operator: operator,
			})

		default:
			return node.NewOperInstallNodeBySSH(node.OperParamInstallNodeBySSH{
				Token:    deploy.Token,
				Operator: operator,
			})
		}
	}

	// install by relay.
	if deploy.Info.InstallOptions.IsManual {
		return node.NewOperInstallPagentByManual(node.OperParamInstallPagentByManual{
			Token:    deploy.Token,
			Operator: operator,
		})
	}

	switch criteria.OSType(deploy.Info.Host.Static.OSType) {
	case criteria.OSLinux, criteria.OSDarwin:
		return node.NewOperInstallPagentNodeBySSH(node.OperParamInstallPagentNodeBySSH{
			Token:    deploy.Token,
			Operator: operator,
		})

	case criteria.OSWindows:
		return node.NewOperInstallPagentNodeByWMI(node.OperParamInstallPagentNodeByWMI{
			Token:    deploy.Token,
			Operator: operator,
		})

	default:
		return node.NewOperInstallPagentNodeBySSH(node.OperParamInstallPagentNodeBySSH{
			Token:    deploy.Token,
			Operator: operator,
		})
	}
}

// proxy install distinguish direct install or pagent install.
func (mgr *Manager) getNodeInstallOperationDefProxy(deploy *types.NodeDeployment, operator string) operation.Definition {
	if deploy.Info.InstallOptions.DirectInstall {
		return node.NewOperInstallNodeBySSH(node.OperParamInstallNodeBySSH{
			Token:    deploy.Token,
			Operator: operator,
		})
	}

	return node.NewOperInstallPagentNodeBySSH(node.OperParamInstallPagentNodeBySSH{
		Token:    deploy.Token,
		Operator: operator,
	})
}

// LaunchUpgradeNode launch a task to upgrade node. returns the workflow-id.
func (mgr *Manager) LaunchUpgradeNode(nCtx contextx.IContext, param types.UpgradeNodeParam) (string, error) {
	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	if err = mgr.conf.StorageNode.CreateNodeWorkflow(nCtx, &types.NodeWorkflow{
		TenantID:    nCtx.TenantID(),
		WorkflowID:  workflowID,
		TriggerID:   triggerCtl.GetTriggerID(),
		Type:        param.Type,
		BizIDs:      param.BizIDs,
		Operator:    param.Operator,
		OperateTime: time.Now(),
		Status:      types.NodeWorkflowStatusRunning,
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
	// proxy.
	if deploy.Info.Host.Dynamic.NodeRole == types.NodeRoleProxy {
		return node.NewOperUpgradeNode(node.OperParamUpgradeNode{
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
	return node.NewoperUpgradePagent(node.OperParamUpgradePagent{
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

	// if not direct link, use pagent. than we noly need to transfer installer.
	deploy.Info.TransferOptions.SelectDownloads = true
	deploy.Info.TransferOptions.EnableInstaller = true
}

// LaunchReconfigNode launch a task to reconfig node. returns the workflow-id.
func (mgr *Manager) LaunchReconfigNode(nCtx contextx.IContext, param types.ReconfigNodeParam) (string, error) {
	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	if err = mgr.conf.StorageNode.CreateNodeWorkflow(nCtx, &types.NodeWorkflow{
		TenantID:    nCtx.TenantID(),
		WorkflowID:  workflowID,
		TriggerID:   triggerCtl.GetTriggerID(),
		Type:        param.Type,
		BizIDs:      param.BizIDs,
		Operator:    param.Operator,
		OperateTime: time.Now(),
		Status:      types.NodeWorkflowStatusRunning,
	}); err != nil {
		return "", err
	}

	logger.G.Biz(nCtx).
		With("trigger-id", triggerCtl.GetTriggerID(), "node-deployments", len(param.NodeDeployments)).
		Info("launching reconfig node")

	gp := gopool.NewPool()
	for _, nodeDeploy := range param.NodeDeployments {
		deploy := nodeDeploy
		deploy.Info.TransferOptions.SelectDownloads = true
		deploy.Info.TransferOptions.EnableInstaller = true

		gp.Go(func() error {
			if err := mgr.conf.StorageNode.CreateNodeDeployment(nCtx, deploy); err != nil {
				logger.G.Biz(nCtx).
					WithErr(err).
					With("trigger-id", triggerCtl.GetTriggerID(), "token", deploy.Token).
					Error("failed to launch reconfig node, failed to create node deployment")

				return err
			}

			operationDef := node.NewOperReconfigNode(node.OperParamReconfigNode{
				Token:    deploy.Token,
				Operator: param.Operator,
			})
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
	if err = mgr.conf.StorageNode.CreateNodeWorkflow(nCtx, &types.NodeWorkflow{
		TenantID:    nCtx.TenantID(),
		WorkflowID:  workflowID,
		TriggerID:   triggerCtl.GetTriggerID(),
		Type:        param.Type,
		BizIDs:      param.BizIDs,
		Operator:    param.Operator,
		OperateTime: time.Now(),
		Status:      types.NodeWorkflowStatusRunning,
	}); err != nil {
		return "", err
	}

	logger.G.Biz(nCtx).
		With("trigger-id", triggerCtl.GetTriggerID(), "node-deployments", len(param.NodeDeployments)).
		Info("launching restart node")

	gp := gopool.NewPool()
	for _, nodeDeploy := range param.NodeDeployments {
		deploy := nodeDeploy
		deploy.Info.TransferOptions.SelectDownloads = true
		deploy.Info.TransferOptions.EnableInstaller = true

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
	if err = mgr.conf.StorageNode.CreateNodeWorkflow(nCtx, &types.NodeWorkflow{
		TenantID:    nCtx.TenantID(),
		WorkflowID:  workflowID,
		TriggerID:   triggerCtl.GetTriggerID(),
		Type:        param.Type,
		BizIDs:      param.BizIDs,
		Operator:    param.Operator,
		OperateTime: time.Now(),
		Status:      types.NodeWorkflowStatusRunning,
	}); err != nil {
		return "", err
	}

	logger.G.Biz(nCtx).
		With("trigger-id", triggerCtl.GetTriggerID(), "node-deployments", len(param.NodeDeployments)).
		Info("launching uninstall node")

	gp := gopool.NewPool()
	for _, nodeDeploy := range param.NodeDeployments {
		deploy := nodeDeploy
		deploy.Info.TransferOptions.SelectDownloads = true
		deploy.Info.TransferOptions.EnableInstaller = true

		gp.Go(func() error {
			if err := mgr.conf.StorageNode.CreateNodeDeployment(nCtx, deploy); err != nil {
				logger.G.Biz(nCtx).
					WithErr(err).
					With("trigger-id", triggerCtl.GetTriggerID(), "token", deploy.Token).
					Error("failed to launch uninstall node, failed to create node deployment")

				return err
			}

			operationDef := node.NewOperUninstallNode(node.OperParamUninstallNode{
				Token:    deploy.Token,
				Operator: param.Operator,
			})
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
