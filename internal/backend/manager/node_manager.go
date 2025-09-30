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

// INodeManager defines the NodeManager interface.
type INodeManager interface {

	// LaunchInstallNode launch a task to install node. returns the workflow-id.
	LaunchInstallNode(ctx contextx.IContext, param InstallNodeParam) (string, error)

	// LaunchUpgradeNode launch a task to upgrade node. returns the workflow-id.
	LaunchUpgradeNode(ctx contextx.IContext, param UpgradeNodeParam) (string, error)

	// LaunchReconfigNode launch a task to reconfig node. returns the workflow-id.
	LaunchReconfigNode(ctx contextx.IContext, param ReconfigNodeParam) (string, error)

	// LaunchRestartNode launch a task to restart node. returns the workflow-id.
	LaunchRestartNode(ctx contextx.IContext, param RestartNodeParam) (string, error)

	// RetryOperationNode launch a task to retry operation instance
	RetryOperationNode(ctx contextx.IContext, param RetryOperationNodeParam) ([]string, error)
}

// InstallNodeParam install node param.
type InstallNodeParam struct {
	Type            types.NodeWorkflowType
	BizIDs          []int64
	Operator        string
	NodeDeployments []*types.NodeDeployment
}

// UpgradeNodeParam upgrade node param.
type UpgradeNodeParam struct {
	Type            types.NodeWorkflowType
	BizIDs          []int64
	Operator        string
	NodeDeployments []*types.NodeDeployment
}

// ReconfigNodeParam reconfig node param.
type ReconfigNodeParam struct {
	Type            types.NodeWorkflowType
	BizIDs          []int64
	Operator        string
	NodeDeployments []*types.NodeDeployment
}

// RestartNodeParam restart node param.
type RestartNodeParam struct {
	Type            types.NodeWorkflowType
	BizIDs          []int64
	Operator        string
	NodeDeployments []*types.NodeDeployment
}

// RetryOperationNodeParam retry node param.
type RetryOperationNodeParam struct {
	WorkflowID string

	RetryMod     types.NodeOperationRetryMode
	OperationIDs []string
}

// LaunchInstallNode launch a task to install node.
func (mgr *Manager) LaunchInstallNode(ctx contextx.IContext, param InstallNodeParam) (string, error) {
	triggerCtl, err := mgr.workflowMgr.CreateTrigger(ctx, trigger.CategoryOnce, &trigger.MetadataOnce{})
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	if err = mgr.conf.StorageNode.CreateNodeWorkflow(ctx, &types.NodeWorkflow{
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
			return mgr.createInstallNodeOper(ctx, param.Operator, triggerCtl, deploy)
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch install node task. err: %w", err)
	}

	if err = triggerCtl.RunTrigger(ctx); err != nil {
		return "", err
	}

	return workflowID, nil
}

// RetryOperationNode launch a task to retry operation instance.
func (mgr *Manager) RetryOperationNode(ctx contextx.IContext, param RetryOperationNodeParam) ([]string, error) {
	instanceIDs := make([]string, 0)

	nodeWorkflow, err := mgr.conf.StorageNode.GetNodeWorkflow(ctx, param.WorkflowID)
	if err != nil {
		return nil, fmt.Errorf("get trigger failed: %w", err)
	}

	triggerCtl, err := mgr.workflowMgr.GetTrigger(ctx, nodeWorkflow.TriggerID)
	if err != nil {
		return nil, fmt.Errorf("get trigger failed: %w", err)
	}

	operCtls, err := triggerCtl.ListOperation(ctx, param.OperationIDs...)
	if err != nil {
		return nil, fmt.Errorf("get operation failed: %w", err)
	}

	err = mgr.conf.StorageNode.UpdateNodeWorkflowStatus(ctx, param.WorkflowID, types.NodeWorkflowStatusRunning)
	if err != nil {
		return nil, fmt.Errorf("update node workflow status failed: %w", err)
	}

	for _, operCtl := range operCtls {
		instanceCtl, err := operCtl.CreateRetryOperationInstance(ctx, param.RetryMod)
		if err != nil {
			return nil, fmt.Errorf("create operation instance failed: %w", err)
		}

		if err := instanceCtl.LaunchOperationInstance(ctx); err != nil {
			return nil, fmt.Errorf("launch operation instance failed: %w", err)
		}

		instanceIDs = append(instanceIDs, instanceCtl.GetOperationInstanceID())
	}

	return instanceIDs, nil
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
	operationParam.ExtraContent = convNodeDeploymentInfoToMap(deploy.Info)

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
	// proxy.
	// user select install origin form upstream or current or server.
	if deploy.Info.Host.Dynamic.NodeRole == types.NodeRoleProxy {
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

	// system select install origin form server or proxy.
	// direct agent.
	if deploy.Info.InstallOptions.DirectInstall {
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

	// pagent under proxy.
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

// LaunchUpgradeNode launch a task to upgrade node. returns the workflow-id.
func (mgr *Manager) LaunchUpgradeNode(nCtx contextx.IContext, param UpgradeNodeParam) (string, error) {
	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, &trigger.MetadataOnce{})
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	if err = mgr.conf.StorageNode.CreateNodeWorkflow(nCtx, &types.NodeWorkflow{
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
			operationParam.ExtraContent = convNodeDeploymentInfoToMap(deploy.Info)

			if err := mgr.conf.StorageNode.CreateNodeDeployment(nCtx, deploy); err != nil {
				logger.G.Biz(nCtx).
					WithErr(err).
					With("trigger-id", triggerCtl.GetTriggerID(), "token", deploy.Token).
					Error("failed to launch upgrade node, failed to create node deployment")

				return err
			}

			operationParam.ExtraContent = convNodeDeploymentInfoToMap(deploy.Info)

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

	if err = triggerCtl.RunTrigger(nCtx); err != nil {
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
func (mgr *Manager) LaunchReconfigNode(nCtx contextx.IContext, param ReconfigNodeParam) (string, error) {
	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, &trigger.MetadataOnce{})
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	if err = mgr.conf.StorageNode.CreateNodeWorkflow(nCtx, &types.NodeWorkflow{
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
			operationParam.ExtraContent = convNodeDeploymentInfoToMap(deploy.Info)

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

	if err = triggerCtl.RunTrigger(nCtx); err != nil {
		return "", err
	}

	return workflowID, nil
}

// LaunchRestartNode launch a task to restart node. returns the workflow-id.
func (mgr *Manager) LaunchRestartNode(nCtx contextx.IContext, param RestartNodeParam) (string, error) {
	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, &trigger.MetadataOnce{})
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	if err = mgr.conf.StorageNode.CreateNodeWorkflow(nCtx, &types.NodeWorkflow{
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
			operationParam.ExtraContent = convNodeDeploymentInfoToMap(deploy.Info)

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

	if err = triggerCtl.RunTrigger(nCtx); err != nil {
		return "", err
	}

	return workflowID, nil
}

func convNodeDeploymentInfoToMap(info *types.DeploymentInfo) map[string]any {
	return map[string]any{
		"networkarea_id": info.Host.Static.NetworkAreaID,
		"biz_id":         info.Host.Static.BizID,
		"inner_ip":       info.Host.Static.InnerIP,
		"inner_ipv6":     info.Host.Static.InnerIPV6,
		"node_version":   info.Host.Dynamic.NodeVersion,
	}
}
