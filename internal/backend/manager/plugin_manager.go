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
func (mgr *Manager) LaunchInstallPlugin(nCtx contextx.IContext, param types.InstallPluginParam) (string, error) {
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
		Operator:    param.Operator,
		OperateTime: time.Now(),
		Status:      types.PluginWorkflowStatusRunning,
	}); err != nil {
		return "", err
	}

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

	return workflowID, nil
}

func (mgr *Manager) createInstallPluginOper(
	nCtx contextx.IContext, operator string, triggerCtl workflow.ITriggerCtl, deploy *types.PluginDeployment) error {

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
			With("operation-id", operCtl.GetOperationID()).
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

func (mgr *Manager) getPluginInstallOperationDef(deploy *types.PluginDeployment, operator string) operation.Definition {
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
		TenantID:    nCtx.TenantID(),
		WorkflowID:  workflowID,
		TriggerID:   triggerCtl.GetTriggerID(),
		Type:        param.Type,
		HostIDs:     param.HostIDs,
		Operator:    param.Operator,
		OperateTime: time.Now(),
		Status:      types.PluginWorkflowStatusRunning,
	}); err != nil {
		return "", err
	}

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
			With("operation-id", operCtl.GetOperationID()).
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

// LaunchApplyPluginSubConfig launch a task to apply plugin subconfig. returns the workflow-id.
func (mgr *Manager) LaunchApplyPluginSubConfig(nCtx contextx.IContext, param types.ApplyPluginSubConfigParam) (string, error) {
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
		Operator:    param.Operator,
		OperateTime: time.Now(),
		Status:      types.PluginWorkflowStatusRunning,
	}); err != nil {
		return "", err
	}

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
			})

			operationParam := operationDef.DefaultParameters()

			operCtl, err := triggerCtl.CreateOperation(nCtx, operationDef, operationParam)
			if err != nil {
				logger.G.Biz(nCtx).WithErr(err).
					With("trigger-id", triggerCtl.GetTriggerID()).
					With("operation-id", operCtl.GetOperationID()).
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
		})
	}

	if err := gp.Wait(); err != nil {
		return "", fmt.Errorf("failed to launch apply plugin subconfig task. err: %w", err)
	}

	if err = triggerCtl.ActivateTrigger(nCtx); err != nil {
		return "", err
	}

	return workflowID, nil
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
