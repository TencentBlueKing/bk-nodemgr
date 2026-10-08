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

	pluginv2 "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/pluginv2"
	pluginV2Utils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/pluginv2/utils"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

// LaunchPluginEnsurePluginV2 launch a task to ensure plugin v2. returns the workflow-id.
func (mgr *Manager) LaunchPluginEnsurePluginV2(
	nCtx contextx.IContext,
	param types.InstallPluginParam,
) (string, error) {

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	if err = mgr.conf.StoragePlugin.CreatePluginWorkflow(nCtx, &types.PluginWorkflow{
		TenantID:    nCtx.TenantID(),
		WorkflowID:  workflowID,
		TriggerID:   triggerCtl.GetTriggerID(),
		Type:        param.Type,
		HostIDs:     param.HostIDs,
		BizIDs:      param.BizIDs,
		Operator:    param.Operator,
		OperateTime: time.Now(),
		Status:      types.PluginWorkflowStatusRunning,
	}); err != nil {
		return "", err
	}

	needRollback := true
	defer func() {
		if !needRollback {
			return
		}

		if err := triggerCtl.InactivateTrigger(nCtx); err != nil {
			logger.G.Biz(nCtx).WithErr(err).
				With("trigger-id", triggerCtl.GetTriggerID()).
				Error("failed to inactivate trigger after launch failure")
		}

		if err := mgr.conf.StoragePlugin.UpdatePluginWorkflowStatus(nCtx, workflowID, types.PluginWorkflowStatusFailed); err != nil {
			logger.G.Biz(nCtx).WithErr(err).
				With("workflow-id", workflowID, "trigger-id", triggerCtl.GetTriggerID()).
				Error("failed to mark plugin workflow as failed after launch failure")
		}
	}()

	gp := gopool.NewPool()
	for _, pluginDeploy := range param.PluginDeployments {
		deploy := pluginDeploy

		gp.Go(func() error {
			return mgr.createPluginEnsurePluginV2Oper(nCtx, param.Operator, triggerCtl, deploy)
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch ensure plugin v2 task. err: %w", err)
	}

	if err = triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", err
	}

	needRollback = false

	return workflowID, nil
}

func (mgr *Manager) createPluginEnsurePluginV2Oper(
	nCtx contextx.IContext,
	operator string,
	triggerCtl workflow.ITriggerCtl,
	deploy *types.PluginDeployment,
) error {

	if err := mgr.conf.StoragePlugin.CreatePluginDeployment(nCtx, deploy); err != nil {
		logger.G.Biz(nCtx).WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("pluginv2-token", deploy.Token).
			Error("failed to create pluginv2 deployment.")

		return err
	}

	operationDef := mgr.getPluginEnsurePluginV2OperationDef(deploy, operator)

	operationParam := operationDef.DefaultParameters()

	operCtl, err := triggerCtl.CreateOperation(nCtx, operationDef, operationParam)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("pluginv2-token", deploy.Token).
			Error("failed to launch ensure plugin v2 task, failed to create operation.")

		return err
	}

	logger.G.Biz(nCtx).
		With("trigger-id", triggerCtl.GetTriggerID()).
		With("operation-id", operCtl.GetOperationID()).
		With("pluginv2-token", deploy.Token).
		Info("launched ensure plugin v2 task.")

	return nil
}

func (mgr *Manager) getPluginEnsurePluginV2OperationDef(
	deploy *types.PluginDeployment,
	operator string,
) operation.Definition {

	return pluginv2.NewOperPluginEnsurePluginV2(pluginv2.OperParamPluginEnsurePluginV2{
		PluginActionStandardParam: pluginV2Utils.PluginActionStandardParam{
			Token:    deploy.Token,
			TenantID: deploy.Info.Process.TenantID,
			Operator: operator,
		},
	})
}

// LaunchUninstallPluginV2 launch a task to uninstall pluginv2. returns the workflow-id.
func (mgr *Manager) LaunchUninstallPluginV2(nCtx contextx.IContext, param types.UninstallPluginParam) (string, error) {
	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	if err = mgr.conf.StoragePlugin.CreatePluginWorkflow(nCtx, &types.PluginWorkflow{
		TenantID:    nCtx.TenantID(),
		WorkflowID:  workflowID,
		TriggerID:   triggerCtl.GetTriggerID(),
		Type:        param.Type,
		HostIDs:     param.HostIDs,
		BizIDs:      param.BizIDs,
		Operator:    param.Operator,
		OperateTime: time.Now(),
		Status:      types.PluginWorkflowStatusRunning,
	}); err != nil {
		return "", err
	}

	needRollback := true
	defer func() {
		if !needRollback {
			return
		}

		if err := triggerCtl.InactivateTrigger(nCtx); err != nil {
			logger.G.Biz(nCtx).WithErr(err).
				With("trigger-id", triggerCtl.GetTriggerID()).
				Error("failed to inactivate trigger after launch failure")
		}

		if err := mgr.conf.StoragePlugin.UpdatePluginWorkflowStatus(nCtx, workflowID, types.PluginWorkflowStatusFailed); err != nil {
			logger.G.Biz(nCtx).WithErr(err).
				With("workflow-id", workflowID, "trigger-id", triggerCtl.GetTriggerID()).
				Error("failed to mark plugin workflow as failed after launch failure")
		}
	}()

	gp := gopool.NewPool()
	for _, pluginDeploy := range param.PluginDeployments {
		deploy := pluginDeploy

		gp.Go(func() error {
			return mgr.createUninstallPluginV2Oper(nCtx, param.Operator, triggerCtl, deploy)
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch uninstall pluginv2 task. err: %w", err)
	}

	if err = triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", err
	}

	needRollback = false

	return workflowID, nil
}

func (mgr *Manager) createUninstallPluginV2Oper(
	nCtx contextx.IContext, operator string, triggerCtl workflow.ITriggerCtl, deploy *types.PluginDeployment) error {

	if deploy.PluginConf == nil {
		deploy.PluginConf = &types.PluginDeploymentPluginConf{}
	}
	deploy.PluginConf.RemoveAllConfigs = true
	if err := mgr.conf.StoragePlugin.CreatePluginDeployment(nCtx, deploy); err != nil {
		logger.G.Biz(nCtx).
			WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("pluginv2-token", deploy.Token).
			Error("failed to create pluginv2 deployment.")

		return err
	}

	operationDef := mgr.getPluginUninstallV2OperationDef(deploy, operator)

	operationParam := operationDef.DefaultParameters()

	operCtl, err := triggerCtl.CreateOperation(nCtx, operationDef, operationParam)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("pluginv2-token", deploy.Token).
			Error("failed to launch uninstall pluginv2 task, failed to create operation.")

		return err
	}

	logger.G.Biz(nCtx).
		With("trigger-id", triggerCtl.GetTriggerID()).
		With("operation-id", operCtl.GetOperationID()).
		With("pluginv2-token", deploy.Token).
		Info("launched uninstall pluginv2 task.")

	return nil
}

func (mgr *Manager) getPluginUninstallV2OperationDef(deploy *types.PluginDeployment, operator string) operation.Definition {
	return pluginv2.NewOperUninstallPluginV2(pluginv2.OperParamUninstallPluginV2{
		PluginActionStandardParam: pluginV2Utils.PluginActionStandardParam{
			Token:    deploy.Token,
			TenantID: deploy.Info.Process.TenantID,
			Operator: operator,
		},
	})
}

// LaunchStopPluginV2 launch a task to stop pluginv2. returns the workflow-id.
func (mgr *Manager) LaunchStopPluginV2(nCtx contextx.IContext, param types.StopProcessParam) (string, error) {
	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	if err = mgr.conf.StoragePlugin.CreatePluginWorkflow(nCtx, &types.PluginWorkflow{
		TenantID:    nCtx.TenantID(),
		WorkflowID:  workflowID,
		TriggerID:   triggerCtl.GetTriggerID(),
		Type:        param.Type,
		HostIDs:     param.HostIDs,
		BizIDs:      param.BizIDs,
		Operator:    param.Operator,
		OperateTime: time.Now(),
		Status:      types.PluginWorkflowStatusRunning,
	}); err != nil {
		return "", err
	}

	needRollback := true
	defer func() {
		if !needRollback {
			return
		}

		if err := triggerCtl.InactivateTrigger(nCtx); err != nil {
			logger.G.Biz(nCtx).WithErr(err).
				With("trigger-id", triggerCtl.GetTriggerID()).
				Error("failed to inactivate trigger after launch failure")
		}

		if err := mgr.conf.StoragePlugin.UpdatePluginWorkflowStatus(nCtx, workflowID, types.PluginWorkflowStatusFailed); err != nil {
			logger.G.Biz(nCtx).WithErr(err).
				With("workflow-id", workflowID, "trigger-id", triggerCtl.GetTriggerID()).
				Error("failed to mark plugin workflow as failed after launch failure")
		}
	}()

	gp := gopool.NewPool()
	for _, pluginDeploy := range param.PluginDeployments {
		deploy := pluginDeploy

		gp.Go(func() error {
			return mgr.createStopPluginV2Oper(nCtx, param.Operator, triggerCtl, deploy)
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch stop pluginv2 task. err: %w", err)
	}

	if err = triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", err
	}

	needRollback = false

	return workflowID, nil
}

func (mgr *Manager) createStopPluginV2Oper(
	nCtx contextx.IContext, operator string, triggerCtl workflow.ITriggerCtl, deploy *types.PluginDeployment) error {

	if err := mgr.conf.StoragePlugin.CreatePluginDeployment(nCtx, deploy); err != nil {
		logger.G.Biz(nCtx).
			WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("pluginv2-token", deploy.Token).
			Error("failed to create pluginv2 deployment.")

		return err
	}

	operationDef := mgr.getPluginStopV2OperationDef(deploy, operator)

	operationParam := operationDef.DefaultParameters()

	operCtl, err := triggerCtl.CreateOperation(nCtx, operationDef, operationParam)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("pluginv2-token", deploy.Token).
			Error("failed to launch stop pluginv2 task, failed to create operation.")

		return err
	}

	logger.G.Biz(nCtx).
		With("trigger-id", triggerCtl.GetTriggerID()).
		With("operation-id", operCtl.GetOperationID()).
		With("pluginv2-token", deploy.Token).
		Info("launched stop pluginv2 task.")

	return nil
}

func (mgr *Manager) getPluginStopV2OperationDef(deploy *types.PluginDeployment, operator string) operation.Definition {
	return pluginv2.NewOperStopPluginV2(pluginv2.OperParamStopPluginV2{
		PluginActionStandardParam: pluginV2Utils.PluginActionStandardParam{
			Token:    deploy.Token,
			TenantID: deploy.Info.Process.TenantID,
			Operator: operator,
		},
	})
}

// LaunchRetryPluginOperationFromLastInstanceV2 launch a task to retry operation from last instance.
func (mgr *Manager) LaunchRetryPluginOperationFromLastInstanceV2(nCtx contextx.IContext, param types.RetryPluginWorkflowOperationParam) error {
	nodeWorkflow, err := mgr.conf.StoragePlugin.GetPluginWorkflow(nCtx, param.WorkflowID)
	if err != nil {
		return fmt.Errorf("failed to get pluginv2 workflow: %w", err)
	}

	triggerCtl, err := mgr.workflowMgr.GetTrigger(nCtx, nodeWorkflow.TriggerID)
	if err != nil {
		return fmt.Errorf("failed to get trigger: %w", err)
	}

	if err := triggerCtl.UpdateOperationRetryFlag(nCtx, param.RetryMod, param.OperationIDs...); err != nil {
		return fmt.Errorf("failed to update operation retry flag: %w", err)
	}

	if err := mgr.conf.StoragePlugin.UpdatePluginWorkflowStatus(nCtx, param.WorkflowID, types.PluginWorkflowStatusRunning); err != nil {
		return fmt.Errorf("failed to update pluginv2 workflow status: %w", err)
	}

	return triggerCtl.ActivateTrigger(nCtx)
}

// TerminatePluginOperationLastInstanceV2 terminate operation from last instance.
func (mgr *Manager) TerminatePluginOperationLastInstanceV2(nCtx contextx.IContext, param types.TerminatePluginWorkflowOperationParam) error {
	nodeWorkflow, err := mgr.conf.StoragePlugin.GetPluginWorkflow(nCtx, param.WorkflowID)
	if err != nil {
		return fmt.Errorf("failed to get pluginv2 workflow: %w", err)
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
		return fmt.Errorf("failed to terminate pluginv2 operation: %w", err)
	}

	return nil
}
