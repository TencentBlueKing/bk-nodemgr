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

package plugin

import (
	"fmt"
	"time"

	managerIface "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/iface"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

// ActionNamePrepareProjectionSources identifies batch source preparation.
const ActionNamePrepareProjectionSources = "prepare_projection_sources"

// NewActionPrepareProjectionSources creates the projection source preparation action.
func NewActionPrepareProjectionSources(capability *Capability) action.Definition {
	return &actionPrepareProjectionSources{
		storagePlugin: capability.StoragePlugin, storageAction: capability.StorageWorkflow, pluginManager: capability.PluginIface,
	}
}

type actionPrepareProjectionSources struct {
	storagePlugin pluginStg.IStorage
	storageAction workflow.IStorageActionInstance
	pluginManager managerIface.IPluginManager
}

// Name returns the action name.
func (act *actionPrepareProjectionSources) Name() string { return ActionNamePrepareProjectionSources }

// DisplayNameZh returns the action display name.
func (act *actionPrepareProjectionSources) DisplayNameZh() string { return "准备投射源插件" }

// DisplayNameEn returns the action display name.
func (act *actionPrepareProjectionSources) DisplayNameEn() string {
	return "Prepare Projection Sources"
}

// Version returns the action version.
func (act *actionPrepareProjectionSources) Version() string { return "1.0.0" }

// Description returns the action description.
func (act *actionPrepareProjectionSources) Description() string { return act.DisplayNameEn() }

// Timeout returns the child workflow wait timeout.
func (act *actionPrepareProjectionSources) Timeout() time.Duration { return 10 * time.Minute } // nolint: mnd

// Tags returns action tags.
func (act *actionPrepareProjectionSources) Tags() []action.Tag { return []action.Tag{} }

// MaxRetryCount disables automatic relaunch while children may still be running.
func (act *actionPrepareProjectionSources) MaxRetryCount() uint { return 0 }

// DelayFn returns the standard retry delay.
func (act *actionPrepareProjectionSources) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Do launches a mature batch source workflow, then waits before any target is applied.
func (act *actionPrepareProjectionSources) Do(ctx *action.InstanceContext) error {
	param, err := parseProjectionParam(ctx)
	if err != nil {
		return err
	}
	deployments := make([]*types.PluginDeployment, 0, len(param.SourceTokens))
	bizIDs := make([]int64, 0, len(param.SourceTokens))
	for _, token := range param.SourceTokens {
		std, err := projectionSource(ctx, act.storagePlugin, param, token)
		if err != nil {
			return err
		}
		if param.SourceOperation == types.PluginSubConfigSourceNone {
			continue
		}
		info := *std.DeployInfo()
		info.ApplyPluginSubConfigOptions = types.PluginDeploymentApplyPluginSubConfigOptions{}
		conf, err := act.storagePlugin.GetPluginDeploymentPluginConf(std.Context(), token)
		if err != nil {
			return fmt.Errorf("failed to get projection source config: %w", err)
		}
		deployments = append(deployments, types.NewPluginDeployment(&info, conf))
		bizIDs = append(bizIDs, info.Process.BizID)
	}
	if param.SourceOperation == types.PluginSubConfigSourceNone {
		return nil
	}
	nCtx := contextx.From(ctx.Ctx, contextx.WithTenantID(param.TenantID), contextx.WithBKUsername(param.Operator))
	bizIDs = conv.SliceUnique(bizIDs)
	var workflowID string
	switch param.SourceOperation {
	case types.PluginSubConfigSourceNone:
		return nil
	case types.PluginSubConfigSourceInstall:
		workflowID, err = act.pluginManager.LaunchInstallPlugin(nCtx, types.InstallPluginParam{
			Type: types.PluginWorkflowTypeInstall, HostIDs: []int64{param.HostID}, BizIDs: bizIDs,
			Operator: param.Operator, PluginDeployments: deployments,
		})
	case types.PluginSubConfigSourceUpgrade:
		workflowID, err = act.pluginManager.LaunchUpgradePlugin(nCtx, types.UpgradePluginParam{
			Type: types.PluginWorkflowTypeUpgrade, HostIDs: []int64{param.HostID}, BizIDs: bizIDs,
			Operator: param.Operator, PluginDeployments: deployments,
		})
	}
	if err != nil {
		return fmt.Errorf("failed to launch projection sources: %w", err)
	}
	if workflowID == "" {
		return fmt.Errorf("projection source workflow id is empty")
	}
	if err := saveProjectionChildren(act.storageAction, &action.InstanceContext{Ctx: nCtx, Data: ctx.Data}, act.Name(), []types.SubWorkflowRef{
		{WorkflowID: workflowID, WorkflowDomain: types.WorkflowDomainPlugin},
	}); err != nil {
		return err
	}

	return waitPluginSubConfigChild(nCtx, act.storagePlugin, workflowID, act.Timeout())
}
