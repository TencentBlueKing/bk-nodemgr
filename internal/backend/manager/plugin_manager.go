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

// IPluginManager defines the PluginManager interface.
type IPluginManager interface {
	// LaunchInstallPlugin launch a task to install plugin. returns the workflow-id.
	LaunchInstallPlugin(ctx contextx.IContext, param InstallPluginParam) (string, error)
}

// InstallPluginParam define the param of LaunchInstallPlugin.
type InstallPluginParam struct {
	Type              types.PluginWorkflowType
	HostIDs           []int64
	Operator          string
	PluginDeployments []*types.PluginDeployment
}

// LaunchInstallPlugin launch a task to install plugin. returns the workflow-id.
func (mgr *Manager) LaunchInstallPlugin(nCtx contextx.IContext, param InstallPluginParam) (string, error) {
	triggerCtl, err := mgr.workflowMgr.CreateTrigger(nCtx, trigger.CategoryOnce, trigger.NewMetadataOnce())
	if err != nil {
		return "", err
	}

	workflowID := identifier.GenWorkflowID()
	if err = mgr.conf.StoragePlugin.CreatePluginWorkflow(nCtx, &types.PluginWorkflow{
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

	if err = triggerCtl.RunTrigger(nCtx); err != nil {
		return "", err
	}

	return workflowID, nil
}

func (mgr *Manager) createInstallPluginOper(nCtx contextx.IContext, operator string, triggerCtl workflow.ITriggerCtl, deploy *types.PluginDeployment,
) error {

	if err := mgr.conf.StoragePlugin.CreatePluginDeployment(nCtx, deploy); err != nil {
		logger.G.Biz(nCtx).Error("failed to create plugin deployment. trigger-id(%s), plugin-token(%s), err(%v)",
			triggerCtl.GetTriggerID(), deploy.Token, err)

		return err
	}

	operationDef, err := mgr.getPluginInstallOperationDef(deploy, operator)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("failed to get plugin install operation definition")
	}

	operationParam := operationDef.DefaultParameters()
	operationParam.ExtraContent = map[string]any{}

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
		Error("launched install plugin task.")

	return nil
}

func (mgr *Manager) getPluginInstallOperationDef(deploy *types.PluginDeployment, operator string) (operation.Definition, error) {
	switch deploy.Info.Plugin.Type {
	case types.PluginTypeOfficial:
		return plugin.NewOperInstallPlugin(plugin.OperParamInstallPlugin{
			PluginActionStandardParam: pluginUtils.PluginActionStandardParam{
				Token:    deploy.Token,
				TenantID: deploy.Info.Plugin.TenantID,
				Operator: operator,
			},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported plugin type: %s", deploy.Info.Plugin.Type)
	}
}
