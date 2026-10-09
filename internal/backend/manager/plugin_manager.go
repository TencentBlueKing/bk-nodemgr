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
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin"
	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/identifier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

// LaunchInstallPlugin launch a task to install plugin. returns the workflow-id.
func (mgr *Manager) LaunchInstallPlugin(
	nCtx contextx.IContext,
	param types.InstallPluginParam,
) (string, error) {

	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	if err = mgr.conf.StoragePlugin.CreatePluginWorkflow(nCtx, &types.PluginWorkflow{
		TenantID:        nCtx.TenantID(),
		WorkflowID:      workflowID,
		TriggerID:       triggerCtl.GetTriggerID(),
		Type:            param.Type,
		HostIDs:         param.HostIDs,
		BizIDs:          param.BizIDs,
		Operator:        param.Operator,
		DeployPolicyIDs: param.DeployPolicyIDs,
		OperateTime:     time.Now(),
		Status:          types.PluginWorkflowStatusRunning,
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
			return mgr.createInstallPluginOper(nCtx, param.Operator, triggerCtl, deploy)
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch install plugin task. err: %w", err)
	}

	if err = triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", err
	}

	needRollback = false

	return workflowID, nil
}

func (mgr *Manager) createInstallPluginOper(
	nCtx contextx.IContext,
	operator string,
	triggerCtl workflow.ITriggerCtl,
	deploy *types.PluginDeployment,
) error {

	if err := mgr.conf.StoragePlugin.CreatePluginDeployment(nCtx, deploy); err != nil {
		logger.G.Biz(nCtx).WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("plugin-token", deploy.Token).
			Error("failed to create plugin deployment.")

		return err
	}

	operationDef := mgr.getPluginInstallOperationDef(deploy, operator)

	operationParam := operationDef.DefaultParameters()

	operCtl, err := triggerCtl.CreateOperation(nCtx, operationDef, operationParam)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("plugin-token", deploy.Token).
			Error("failed to launch install plugin task.")

		return err
	}

	logger.G.Biz(nCtx).
		With("trigger-id", triggerCtl.GetTriggerID()).
		With("operation-id", operCtl.GetOperationID()).
		With("plugin-token", deploy.Token).
		Info("launched install plugin task.")

	return nil
}

func (mgr *Manager) getPluginInstallOperationDef(
	deploy *types.PluginDeployment,
	operator string,
) operation.Definition {

	return plugin.NewOperInstallPlugin(plugin.OperParamInstallPlugin{
		PluginActionStandardParam: pluginUtils.PluginActionStandardParam{
			Token:    deploy.Token,
			TenantID: deploy.Info.Process.TenantID,
			Operator: operator,
		},
	})
}

// LaunchUpgradePlugin launch a task to upgrade plugin. returns the workflow-id.
func (mgr *Manager) LaunchUpgradePlugin(nCtx contextx.IContext, param types.UpgradePluginParam) (string, error) {
	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	if err = mgr.conf.StoragePlugin.CreatePluginWorkflow(nCtx, &types.PluginWorkflow{
		TenantID:        nCtx.TenantID(),
		WorkflowID:      workflowID,
		TriggerID:       triggerCtl.GetTriggerID(),
		Type:            param.Type,
		HostIDs:         param.HostIDs,
		BizIDs:          param.BizIDs,
		Operator:        param.Operator,
		DeployPolicyIDs: param.DeployPolicyIDs,
		OperateTime:     time.Now(),
		Status:          types.PluginWorkflowStatusRunning,
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
			return mgr.createUpgradePluginOper(nCtx, param.Operator, triggerCtl, deploy)
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch upgrade plugin task. err: %w", err)
	}

	if err = triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", err
	}

	needRollback = false

	return workflowID, nil
}

func (mgr *Manager) createUpgradePluginOper(
	nCtx contextx.IContext, operator string, triggerCtl workflow.ITriggerCtl, deploy *types.PluginDeployment) error {

	if err := mgr.conf.StoragePlugin.CreatePluginDeployment(nCtx, deploy); err != nil {
		logger.G.Biz(nCtx).
			WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("plugin-token", deploy.Token).
			Error("failed to create plugin deployment.")

		return err
	}

	operationDef := mgr.getPluginUpgradeOperationDef(deploy, operator)

	operationParam := operationDef.DefaultParameters()

	operCtl, err := triggerCtl.CreateOperation(nCtx, operationDef, operationParam)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("plugin-token", deploy.Token).
			Error("failed to launch upgrade plugin task.")

		return err
	}

	logger.G.Biz(nCtx).
		With("trigger-id", triggerCtl.GetTriggerID()).
		With("operation-id", operCtl.GetOperationID()).
		With("plugin-token", deploy.Token).
		Info("launched upgrade plugin task.")

	return nil
}

func (mgr *Manager) getPluginUpgradeOperationDef(deploy *types.PluginDeployment, operator string) operation.Definition {
	return plugin.NewOperUpgradePlugin(plugin.OperParamUpgradePlugin{
		PluginActionStandardParam: pluginUtils.PluginActionStandardParam{
			Token:    deploy.Token,
			TenantID: deploy.Info.Process.TenantID,
			Operator: operator,
		},
	})
}

// LaunchUninstallPlugin launch a task to uninstall plugin. returns the workflow-id.
func (mgr *Manager) LaunchUninstallPlugin(nCtx contextx.IContext, param types.UninstallPluginParam) (string, error) {
	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	if err = mgr.conf.StoragePlugin.CreatePluginWorkflow(nCtx, &types.PluginWorkflow{
		TenantID:        nCtx.TenantID(),
		WorkflowID:      workflowID,
		TriggerID:       triggerCtl.GetTriggerID(),
		Type:            param.Type,
		HostIDs:         param.HostIDs,
		BizIDs:          param.BizIDs,
		Operator:        param.Operator,
		DeployPolicyIDs: param.DeployPolicyIDs,
		OperateTime:     time.Now(),
		Status:          types.PluginWorkflowStatusRunning,
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
			return mgr.createUninstallPluginOper(nCtx, param.Operator, triggerCtl, deploy)
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch uninstall plugin task. err: %w", err)
	}

	if err = triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", err
	}

	needRollback = false

	return workflowID, nil
}

func (mgr *Manager) createUninstallPluginOper(
	nCtx contextx.IContext, operator string, triggerCtl workflow.ITriggerCtl, deploy *types.PluginDeployment) error {

	if deploy.PluginConf == nil {
		deploy.PluginConf = &types.PluginDeploymentPluginConf{}
	}
	deploy.PluginConf.RemoveAllConfigs = true

	if err := mgr.conf.StoragePlugin.CreatePluginDeployment(nCtx, deploy); err != nil {
		logger.G.Biz(nCtx).
			WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("plugin-token", deploy.Token).
			Error("failed to create plugin deployment.")

		return err
	}

	operationDef := mgr.getPluginUninstallOperationDef(deploy, operator)

	operationParam := operationDef.DefaultParameters()

	operCtl, err := triggerCtl.CreateOperation(nCtx, operationDef, operationParam)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("plugin-token", deploy.Token).
			Error("failed to launch uninstall plugin task.")

		return err
	}

	logger.G.Biz(nCtx).
		With("trigger-id", triggerCtl.GetTriggerID()).
		With("operation-id", operCtl.GetOperationID()).
		With("plugin-token", deploy.Token).
		Info("launched uninstall plugin task.")

	return nil
}

func (mgr *Manager) getPluginUninstallOperationDef(deploy *types.PluginDeployment, operator string) operation.Definition {
	return plugin.NewOperUninstallPlugin(plugin.OperParamUninstallPlugin{
		PluginActionStandardParam: pluginUtils.PluginActionStandardParam{
			Token:    deploy.Token,
			TenantID: deploy.Info.Process.TenantID,
			Operator: operator,
		},
	})
}

// LaunchApplyPluginSubConfig launch a task to apply plugin subconfig. returns the workflow-id.
func (mgr *Manager) LaunchApplyPluginSubConfig(nCtx contextx.IContext, param types.ApplyPluginSubConfigParam) (string, error) {
	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	if err = mgr.conf.StoragePlugin.CreatePluginWorkflow(nCtx, &types.PluginWorkflow{
		TenantID:        nCtx.TenantID(),
		WorkflowID:      workflowID,
		TriggerID:       triggerCtl.GetTriggerID(),
		Type:            param.Type,
		HostIDs:         param.HostIDs,
		BizIDs:          param.BizIDs,
		Operator:        param.Operator,
		DeployPolicyIDs: param.DeployPolicyIDs,
		OperateTime:     time.Now(),
		Status:          types.PluginWorkflowStatusRunning,
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
			if err := mgr.conf.StoragePlugin.CreatePluginDeployment(nCtx, deploy); err != nil {
				logger.G.Biz(nCtx).
					WithErr(err).
					With("trigger-id", triggerCtl.GetTriggerID()).
					With("plugin-token", deploy.Token).
					Error("failed to create plugin deployment.")

				return err
			}

			operationDef := plugin.NewOperApplyPluginSubConfig(plugin.OperParamApplyPluginSubConfig{
				PluginActionStandardParam: pluginUtils.PluginActionStandardParam{
					Token:    deploy.Token,
					TenantID: deploy.Info.Process.TenantID,
					Operator: param.Operator,
				},
				OnlyPushSubConfig: true,
			})

			operationParam := operationDef.DefaultParameters()

			operCtl, err := triggerCtl.CreateOperation(nCtx, operationDef, operationParam)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).
					With("trigger-id", triggerCtl.GetTriggerID()).
					With("plugin-token", deploy.Token).
					Error("failed to launch apply plugin subconfig task.")

				return err
			}

			logger.G.Biz(nCtx).
				With("trigger-id", triggerCtl.GetTriggerID()).
				With("operation-id", operCtl.GetOperationID()).
				With("plugin-token", deploy.Token).
				Info("launched apply plugin subconfig task.")

			return nil
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch apply plugin subconfig task. err: %w", err)
	}

	if err = triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", err
	}

	needRollback = false

	return workflowID, nil
}

// LaunchRemovePluginSubConfig launch a task to remove plugin subconfig. returns the workflow-id.
func (mgr *Manager) LaunchRemovePluginSubConfig(nCtx contextx.IContext, param types.RemovePluginSubConfigParam) (string, error) {
	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	if err = mgr.conf.StoragePlugin.CreatePluginWorkflow(nCtx, &types.PluginWorkflow{
		TenantID:        nCtx.TenantID(),
		WorkflowID:      workflowID,
		TriggerID:       triggerCtl.GetTriggerID(),
		Type:            param.Type,
		HostIDs:         param.HostIDs,
		BizIDs:          param.BizIDs,
		Operator:        param.Operator,
		DeployPolicyIDs: param.DeployPolicyIDs,
		OperateTime:     time.Now(),
		Status:          types.PluginWorkflowStatusRunning,
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
			if err := mgr.conf.StoragePlugin.CreatePluginDeployment(nCtx, deploy); err != nil {
				logger.G.Biz(nCtx).
					WithErr(err).
					With("trigger-id", triggerCtl.GetTriggerID()).
					With("plugin-token", deploy.Token).
					Error("failed to create plugin deployment.")

				return err
			}

			operationDef := plugin.NewOperRemovePluginSubConfig(plugin.OperParamRemovePluginSubConfig{
				PluginActionStandardParam: pluginUtils.PluginActionStandardParam{
					Token:    deploy.Token,
					TenantID: deploy.Info.Process.TenantID,
					Operator: param.Operator,
				},
			})

			operationParam := operationDef.DefaultParameters()

			operCtl, err := triggerCtl.CreateOperation(nCtx, operationDef, operationParam)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).
					With("trigger-id", triggerCtl.GetTriggerID()).
					With("plugin-token", deploy.Token).
					Error("failed to launch remove plugin subconfig task.")

				return err
			}

			logger.G.Biz(nCtx).
				With("trigger-id", triggerCtl.GetTriggerID()).
				With("operation-id", operCtl.GetOperationID()).
				With("plugin-token", deploy.Token).
				Info("launched remove plugin subconfig task.")

			return nil
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch remove plugin subconfig task. err: %w", err)
	}

	if err = triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", err
	}

	needRollback = false

	return workflowID, nil
}

// LaunchStartProcess launch a task to start process. returns the workflow-id.
func (mgr *Manager) LaunchStartProcess(nCtx contextx.IContext, param types.StartProcessParam) (string, error) {
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
			return mgr.createStartProcessOper(nCtx, param.Operator, triggerCtl, deploy)
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch start process task. err: %w", err)
	}

	if err = triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", err
	}

	needRollback = false

	return workflowID, nil
}

func (mgr *Manager) createStartProcessOper(
	nCtx contextx.IContext, operator string, triggerCtl workflow.ITriggerCtl, deploy *types.PluginDeployment) error {

	if err := mgr.conf.StoragePlugin.CreatePluginDeployment(nCtx, deploy); err != nil {
		logger.G.Biz(nCtx).
			WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("plugin-token", deploy.Token).
			Error("failed to create plugin deployment.")

		return err
	}

	operationDef := mgr.getPluginStartOperationDef(deploy, operator)

	operationParam := operationDef.DefaultParameters()

	operCtl, err := triggerCtl.CreateOperation(nCtx, operationDef, operationParam)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("plugin-token", deploy.Token).
			Error("failed to launch start process task.")

		return err
	}

	logger.G.Biz(nCtx).
		With("trigger-id", triggerCtl.GetTriggerID()).
		With("operation-id", operCtl.GetOperationID()).
		With("plugin-token", deploy.Token).
		Info("launched start process task.")

	return nil
}

func (mgr *Manager) getPluginStartOperationDef(deploy *types.PluginDeployment, operator string) operation.Definition {
	return plugin.NewOperStartProcess(plugin.OperParamStartProcess{
		PluginActionStandardParam: pluginUtils.PluginActionStandardParam{
			Token:    deploy.Token,
			TenantID: deploy.Info.Process.TenantID,
			Operator: operator,
		},
	})
}

// LaunchRestartProcess launch a task to restart process. returns the workflow-id.
func (mgr *Manager) LaunchRestartProcess(nCtx contextx.IContext, param types.RestartProcessParam) (string, error) {
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
			return mgr.createRestartProcessOper(nCtx, param.Operator, triggerCtl, deploy)
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch restart process task. err: %w", err)
	}

	if err = triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", err
	}

	needRollback = false

	return workflowID, nil
}

func (mgr *Manager) createRestartProcessOper(
	nCtx contextx.IContext, operator string, triggerCtl workflow.ITriggerCtl, deploy *types.PluginDeployment) error {

	if err := mgr.conf.StoragePlugin.CreatePluginDeployment(nCtx, deploy); err != nil {
		logger.G.Biz(nCtx).
			WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("plugin-token", deploy.Token).
			Error("failed to create plugin deployment.")

		return err
	}

	operationDef := mgr.getPluginRestartOperationDef(deploy, operator)

	operationParam := operationDef.DefaultParameters()

	operCtl, err := triggerCtl.CreateOperation(nCtx, operationDef, operationParam)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("plugin-token", deploy.Token).
			Error("failed to launch restart process task.")

		return err
	}

	logger.G.Biz(nCtx).
		With("trigger-id", triggerCtl.GetTriggerID()).
		With("operation-id", operCtl.GetOperationID()).
		With("plugin-token", deploy.Token).
		Info("launched restart process task.")

	return nil
}

func (mgr *Manager) getPluginRestartOperationDef(deploy *types.PluginDeployment, operator string) operation.Definition {
	return plugin.NewOperRestartProcess(plugin.OperParamRestartProcess{
		PluginActionStandardParam: pluginUtils.PluginActionStandardParam{
			Token:    deploy.Token,
			TenantID: deploy.Info.Process.TenantID,
			Operator: operator,
		},
	})
}

// LaunchMigrateFromV2 launch a task to migrate plugin process from v2. returns the workflow-id.
func (mgr *Manager) LaunchMigrateFromV2(
	nCtx contextx.IContext,
	param types.MigrateFromV2Param,
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
			return mgr.createMigrateFromV2Oper(nCtx, param.Operator, triggerCtl, deploy)
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch migrate plugin process from v2 task. err: %w", err)
	}

	if err = triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", err
	}

	needRollback = false

	return workflowID, nil
}

func (mgr *Manager) createMigrateFromV2Oper(
	nCtx contextx.IContext, operator string, triggerCtl workflow.ITriggerCtl, deploy *types.PluginDeployment) error {

	if err := mgr.conf.StoragePlugin.CreatePluginDeployment(nCtx, deploy); err != nil {
		logger.G.Biz(nCtx).
			WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("plugin-token", deploy.Token).
			Error("failed to create plugin deployment.")

		return err
	}

	operationDef := mgr.getPluginMigrateFromV2OperationDef(deploy, operator)

	operationParam := operationDef.DefaultParameters()

	operCtl, err := triggerCtl.CreateOperation(nCtx, operationDef, operationParam)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("plugin-token", deploy.Token).
			Error("failed to launch migrate plugin process from v2 task.")

		return err
	}

	logger.G.Biz(nCtx).
		With("trigger-id", triggerCtl.GetTriggerID()).
		With("operation-id", operCtl.GetOperationID()).
		With("plugin-token", deploy.Token).
		Info("launched migrate plugin process from v2 task.")

	return nil
}

func (mgr *Manager) getPluginMigrateFromV2OperationDef(
	deploy *types.PluginDeployment,
	operator string,
) operation.Definition {

	return plugin.NewOperMigratePluginProcessFromV2(plugin.OperParamMigratePluginProcessFromV2{
		PluginActionStandardParam: pluginUtils.PluginActionStandardParam{
			Token:    deploy.Token,
			TenantID: deploy.Info.Process.TenantID,
			Operator: operator,
		},
	})
}

// LaunchStopProcess launch a task to stop process. returns the workflow-id.
func (mgr *Manager) LaunchStopProcess(nCtx contextx.IContext, param types.StopProcessParam) (string, error) {
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
			return mgr.createStopProcessOper(nCtx, param.Operator, triggerCtl, deploy)
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch stop process task. err: %w", err)
	}

	if err = triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", err
	}

	needRollback = false

	return workflowID, nil
}

func (mgr *Manager) createStopProcessOper(
	nCtx contextx.IContext, operator string, triggerCtl workflow.ITriggerCtl, deploy *types.PluginDeployment) error {

	if err := mgr.conf.StoragePlugin.CreatePluginDeployment(nCtx, deploy); err != nil {
		logger.G.Biz(nCtx).
			WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("plugin-token", deploy.Token).
			Error("failed to create plugin deployment.")

		return err
	}

	operationDef := mgr.getProcessStopOperationDef(deploy, operator)

	operationParam := operationDef.DefaultParameters()

	operCtl, err := triggerCtl.CreateOperation(nCtx, operationDef, operationParam)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("plugin-token", deploy.Token).
			Error("failed to launch stop process task.")

		return err
	}

	logger.G.Biz(nCtx).
		With("trigger-id", triggerCtl.GetTriggerID()).
		With("operation-id", operCtl.GetOperationID()).
		With("plugin-token", deploy.Token).
		Info("launched stop process task.")

	return nil
}

func (mgr *Manager) getProcessStopOperationDef(deploy *types.PluginDeployment, operator string) operation.Definition {
	return plugin.NewOperStopProcess(plugin.OperParamStopProcess{
		PluginActionStandardParam: pluginUtils.PluginActionStandardParam{
			Token:    deploy.Token,
			TenantID: deploy.Info.Process.TenantID,
			Operator: operator,
		},
	})
}

// LaunchStartDebugPlugin launches a task to debug one prepared plugin deployment. returns the workflow-id.
func (mgr *Manager) LaunchStartDebugPlugin(nCtx contextx.IContext, param types.StartDebugPluginParam) (string, error) {
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
			return mgr.createStartDebugProcessOper(nCtx, param.Operator, triggerCtl, deploy)
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch start debug plugin. err: %w", err)
	}

	if err = triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", err
	}

	needRollback = false

	return workflowID, nil
}

func (mgr *Manager) createStartDebugProcessOper(
	nCtx contextx.IContext, operator string, triggerCtl workflow.ITriggerCtl, deploy *types.PluginDeployment) error {

	if err := mgr.conf.StoragePlugin.CreatePluginDeployment(nCtx, deploy); err != nil {
		logger.G.Biz(nCtx).
			WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("plugin-token", deploy.Token).
			Error("failed to create plugin deployment.")

		return err
	}

	operationDef := mgr.getProcessStartDebugOperationDef(deploy, operator)

	operationParam := operationDef.DefaultParameters()

	operCtl, err := triggerCtl.CreateOperation(nCtx, operationDef, operationParam)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).
			With("trigger-id", triggerCtl.GetTriggerID()).
			With("plugin-token", deploy.Token).
			Error("failed to launch start debug process task.")

		return err
	}

	logger.G.Biz(nCtx).
		With("trigger-id", triggerCtl.GetTriggerID()).
		With("operation-id", operCtl.GetOperationID()).
		With("plugin-token", deploy.Token).
		Info("launched start debug process task.")

	return nil
}

func (mgr *Manager) getProcessStartDebugOperationDef(deploy *types.PluginDeployment, operator string) operation.Definition {
	return plugin.NewOperDebugPlugin(plugin.OperParamDebugPlugin{
		PluginActionStandardParam: pluginUtils.PluginActionStandardParam{
			Token:    deploy.Token,
			TenantID: deploy.Info.Process.TenantID,
			Operator: operator,
		},
	})
}

// LaunchStopDebugPlugin stores a stop signal for a debug workflow.
func (mgr *Manager) LaunchStopDebugPlugin(nCtx contextx.IContext, param types.StopDebugPluginParam) error {
	workflow, err := mgr.conf.StoragePlugin.GetPluginWorkflow(nCtx, param.WorkflowID)
	if err != nil {
		return fmt.Errorf("failed to get debug workflow: %w", err)
	}

	if workflow.Type != types.PluginWorkflowTypeDebug {
		return fmt.Errorf("workflow is not a debug workflow")
	}

	if workflow.Status.IsFinished() {
		return errors.New("workflow is already finished")
	}

	operations, _, err := mgr.conf.StorageWorkflow.ListOperation(
		nCtx,
		types.UnlimitedPage(),
		&types.OperationCondition{ExactInclude: &types.OperationExactFields{TriggerID: []string{workflow.TriggerID}}},
	)
	if err != nil {
		return fmt.Errorf("failed to list debug operation: %w", err)
	}

	if len(operations) == 0 {
		return fmt.Errorf("debug workflow requires at least one operation")
	}

	gp := gopool.NewPool()
	for _, item := range operations {
		oper := item
		gp.Go(func() error {
			operInstID := oper.GetLastInstanceID()
			if operInstID == "" {
				return fmt.Errorf("no operation instance found, operation-id(%s)", oper.OperationID)
			}

			if err := mgr.conf.StorageWorkflow.UpsertActionInstancePrivateData(
				nCtx,
				operInstID,
				plugin.ActionNameRunDebugPlugin,
				map[string]any{types.PDKeyDebugStopSignal: true},
			); err != nil {
				return fmt.Errorf("failed to store debug stop signal, operation-instance-id(%s): %w", operInstID, err)
			}

			return nil
		})
	}

	if err := gp.Wait(); err != nil {
		return err
	}

	logger.G.Biz(nCtx).With("workflow-id", param.WorkflowID).Info("stored plugin debug stop signal")

	return nil
}

// LaunchRetryPluginOperationFromLastInstance launch a task to retry operation from last instance.
func (mgr *Manager) LaunchRetryPluginOperationFromLastInstance(nCtx contextx.IContext, param types.RetryPluginWorkflowOperationParam) error {
	nodeWorkflow, err := mgr.conf.StoragePlugin.GetPluginWorkflow(nCtx, param.WorkflowID)
	if err != nil {
		return fmt.Errorf("failed to get plugin workflow: %w", err)
	}

	triggerCtl, err := mgr.workflowMgr.GetTrigger(nCtx, nodeWorkflow.TriggerID)
	if err != nil {
		return fmt.Errorf("failed to get trigger: %w", err)
	}

	if err := triggerCtl.UpdateOperationRetryFlag(nCtx, param.RetryMod, param.OperationIDs...); err != nil {
		return fmt.Errorf("failed to update operation retry flag: %w", err)
	}

	if err := mgr.conf.StoragePlugin.UpdatePluginWorkflowStatus(nCtx, param.WorkflowID, types.PluginWorkflowStatusRunning); err != nil {
		return fmt.Errorf("failed to update plugin workflow status: %w", err)
	}

	return triggerCtl.ActivateTrigger(nCtx)
}

// TerminatePluginOperationLastInstance terminate operation from last instance.
func (mgr *Manager) TerminatePluginOperationLastInstance(nCtx contextx.IContext, param types.TerminatePluginWorkflowOperationParam) error {
	nodeWorkflow, err := mgr.conf.StoragePlugin.GetPluginWorkflow(nCtx, param.WorkflowID)
	if err != nil {
		return fmt.Errorf("failed to get plugin workflow: %w", err)
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
		return fmt.Errorf("failed to terminate plugin operation: %w", err)
	}

	return nil
}
