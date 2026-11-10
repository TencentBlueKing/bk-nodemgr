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
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/pluginaction"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

// ActionNamePrepareSubConfigSource prepares and waits for the source plugin.
const ActionNamePrepareSubConfigSource = "prepare_subconfig_source"

// NewActionPrepareSubConfigSource creates the source preparation action.
func NewActionPrepareSubConfigSource(capability *Capability) action.Definition {
	return &actionPrepareSubConfigSource{
		storagePlugin: capability.StoragePlugin,
		storageAction: capability.StorageWorkflow,
		pluginManager: capability.PluginIface,
	}
}

type actionPrepareSubConfigSource struct {
	storagePlugin pluginStg.IStorage
	storageAction workflow.IStorageActionInstance
	pluginManager managerIface.IPluginManager
}

// Name returns the action name.
func (act *actionPrepareSubConfigSource) Name() string { return ActionNamePrepareSubConfigSource }

// DisplayNameZh returns the action display name.
func (act *actionPrepareSubConfigSource) DisplayNameZh() string { return "准备子配置源插件" }

// DisplayNameEn returns the action display name.
func (act *actionPrepareSubConfigSource) DisplayNameEn() string { return "Prepare Sub-config Source" }

// Version returns the action version.
func (act *actionPrepareSubConfigSource) Version() string { return "1.0.0" }

// Description returns the action description.
func (act *actionPrepareSubConfigSource) Description() string { return act.DisplayNameEn() }

// Timeout returns the child workflow wait timeout.
func (act *actionPrepareSubConfigSource) Timeout() time.Duration { return 10 * time.Minute } // nolint: mnd

// Tags returns action tags.
func (act *actionPrepareSubConfigSource) Tags() []action.Tag { return []action.Tag{} }

// MaxRetryCount returns the action retry count.
func (act *actionPrepareSubConfigSource) MaxRetryCount() uint { return 3 } // nolint: mnd

// DelayFn returns the standard retry delay.
func (act *actionPrepareSubConfigSource) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Do prepares the source plugin and waits for its deployment workflow.
func (act *actionPrepareSubConfigSource) Do(ctx *action.InstanceContext) error {
	param := new(OperParamComposePluginSubConfig)
	if err := conv.MapToStruct(ctx.Data.Content, param); err != nil {
		return fmt.Errorf("failed to parse composition parameters: %w", err)
	}
	if err := param.SourceOperation.Validate(); err != nil {
		return err
	}
	std := pluginaction.NewPluginActionStandarder(act.storagePlugin)
	if err := std.Initialize(ctx, param.PluginActionStandardParam); err != nil {
		return err
	}

	workflowID, err := act.launchSource(std, param.SourceOperation)
	if err != nil {
		return err
	}

	return saveAndWaitCompositionChild(act, act.storagePlugin, act.storageAction, std, workflowID)
}

func (act *actionPrepareSubConfigSource) launchSource(std *pluginaction.PluginActionStandarder,
	sourceOperation types.PluginSubConfigSourceOperation) (string, error) {

	if sourceOperation == types.PluginSubConfigSourceNone {
		return "", nil
	}
	// Only the copied options value is changed here. Child launch persists a new token;
	// child actions load that stored deployment rather than mutating these pointers.
	info := *std.DeployInfo()
	info.ApplyPluginSubConfigOptions = types.PluginDeploymentApplyPluginSubConfigOptions{}
	conf, err := act.storagePlugin.GetPluginDeploymentPluginConf(std.Context(), std.Token())
	if err != nil {
		return "", fmt.Errorf("failed to get source plugin config: %w", err)
	}
	deploy := types.NewPluginDeployment(&info, conf)
	switch sourceOperation {
	case types.PluginSubConfigSourceInstall:
		return act.pluginManager.LaunchInstallPlugin(std.Context(), types.InstallPluginParam{
			Type: types.PluginWorkflowTypeInstall, HostIDs: []int64{info.Process.HostID}, BizIDs: []int64{info.Process.BizID},
			Operator: std.Operator(), PluginDeployments: []*types.PluginDeployment{deploy},
		})
	case types.PluginSubConfigSourceUpgrade:
		return act.pluginManager.LaunchUpgradePlugin(std.Context(), types.UpgradePluginParam{
			Type: types.PluginWorkflowTypeUpgrade, HostIDs: []int64{info.Process.HostID}, BizIDs: []int64{info.Process.BizID},
			Operator: std.Operator(), PluginDeployments: []*types.PluginDeployment{deploy},
		})
	default:
		return "", fmt.Errorf("unsupported source operation: %s", sourceOperation)
	}
}
