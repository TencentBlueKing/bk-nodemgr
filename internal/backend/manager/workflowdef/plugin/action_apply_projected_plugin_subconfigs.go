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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

// ActionNameApplyProjectedPluginSubConfigs identifies sequential projection target application.
const ActionNameApplyProjectedPluginSubConfigs = "apply_projected_plugin_subconfigs"

// NewActionApplyProjectedPluginSubConfigs creates the projection target action.
func NewActionApplyProjectedPluginSubConfigs(capability *Capability) action.Definition {
	return &actionApplyProjectedPluginSubConfigs{
		storagePlugin: capability.StoragePlugin, storageAction: capability.StorageWorkflow, pluginManager: capability.PluginIface,
	}
}

type actionApplyProjectedPluginSubConfigs struct {
	storagePlugin pluginStg.IStorage
	storageAction workflow.IStorageActionInstance
	pluginManager managerIface.IPluginManager
}

// Name returns the action name.
func (act *actionApplyProjectedPluginSubConfigs) Name() string {
	return ActionNameApplyProjectedPluginSubConfigs
}

// DisplayNameZh returns the action display name.
func (act *actionApplyProjectedPluginSubConfigs) DisplayNameZh() string {
	return "逐项下发投射插件子配置"
}

// DisplayNameEn returns the action display name.
func (act *actionApplyProjectedPluginSubConfigs) DisplayNameEn() string {
	return "Apply Projected Plugin Sub-configs"
}

// Version returns the action version.
func (act *actionApplyProjectedPluginSubConfigs) Version() string { return "1.0.0" }

// Description returns the action description.
func (act *actionApplyProjectedPluginSubConfigs) Description() string { return act.DisplayNameEn() }

// Timeout returns the total sequential target wait timeout.
func (act *actionApplyProjectedPluginSubConfigs) Timeout() time.Duration { return 10 * time.Minute } // nolint: mnd

// Tags returns action tags.
func (act *actionApplyProjectedPluginSubConfigs) Tags() []action.Tag { return []action.Tag{} }

// MaxRetryCount prevents automatic replay after an uncertain child outcome.
func (act *actionApplyProjectedPluginSubConfigs) MaxRetryCount() uint { return 0 }

// DelayFn returns the standard retry delay.
func (act *actionApplyProjectedPluginSubConfigs) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Do launches and waits for each target before processing the next source.
func (act *actionApplyProjectedPluginSubConfigs) Do(ctx *action.InstanceContext) error {
	param, err := parseProjectionParam(ctx)
	if err != nil {
		return err
	}
	target := &actionApplyComposedPluginSubConfig{storagePlugin: act.storagePlugin, pluginManager: act.pluginManager}
	refs := make([]types.SubWorkflowRef, 0, len(param.SourceTokens))
	for _, token := range param.SourceTokens {
		std, err := projectionSource(ctx, act.storagePlugin, param, token)
		if err != nil {
			return err
		}
		workflowID, err := target.launchTarget(std)
		if err != nil {
			return fmt.Errorf("failed to launch projection target for source(%s): %w", token, err)
		}
		if workflowID == "" {
			return fmt.Errorf("projection target workflow id is empty, source(%s)", token)
		}
		refs = append(refs, types.SubWorkflowRef{WorkflowID: workflowID, WorkflowDomain: types.WorkflowDomainPlugin})
		if err := saveProjectionChildren(act.storageAction, &action.InstanceContext{Ctx: std.Context(), Data: ctx.Data}, act.Name(), refs); err != nil {
			return err
		}
		if err := waitCompositionChild(act.storagePlugin, std, workflowID, act.Timeout()); err != nil {
			return fmt.Errorf("failed to apply projection target for source(%s): %w", token, err)
		}
	}

	return nil
}
