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
	"maps"
	"time"

	managerIface "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/iface"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/pluginaction"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

// ActionNameApplyComposedPluginSubConfig applies target configs after source completion.
const ActionNameApplyComposedPluginSubConfig = "apply_composed_plugin_subconfig"

// NewActionApplyComposedPluginSubConfig creates the target config action.
func NewActionApplyComposedPluginSubConfig(capability *Capability) action.Definition {
	return &actionApplyComposedPluginSubConfig{
		storagePlugin: capability.StoragePlugin,
		storageAction: capability.StorageWorkflow,
		pluginManager: capability.PluginIface,
	}
}

type actionApplyComposedPluginSubConfig struct {
	storagePlugin pluginStg.IStorage
	storageAction workflow.IStorageActionInstance
	pluginManager managerIface.IPluginManager
}

// Name returns the action name.
func (act *actionApplyComposedPluginSubConfig) Name() string {
	return ActionNameApplyComposedPluginSubConfig
}

// DisplayNameZh returns the action display name.
func (act *actionApplyComposedPluginSubConfig) DisplayNameZh() string {
	return "下发组合插件子配置"
}

// DisplayNameEn returns the action display name.
func (act *actionApplyComposedPluginSubConfig) DisplayNameEn() string {
	return "Apply Composed Plugin Sub-config"
}

// Version returns the action version.
func (act *actionApplyComposedPluginSubConfig) Version() string { return "1.0.0" }

// Description returns the action description.
func (act *actionApplyComposedPluginSubConfig) Description() string { return act.DisplayNameEn() }

// Timeout returns the child workflow wait timeout.
func (act *actionApplyComposedPluginSubConfig) Timeout() time.Duration { return 10 * time.Minute } // nolint: mnd

// Tags returns action tags.
func (act *actionApplyComposedPluginSubConfig) Tags() []action.Tag { return []action.Tag{} }

// MaxRetryCount returns the action retry count.
func (act *actionApplyComposedPluginSubConfig) MaxRetryCount() uint { return 3 } // nolint: mnd

// DelayFn returns the standard retry delay.
func (act *actionApplyComposedPluginSubConfig) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Do reads the source binding and applies the target plugin sub-config.
func (act *actionApplyComposedPluginSubConfig) Do(ctx *action.InstanceContext) error {
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

	workflowID, err := act.launchTarget(std)
	if err != nil {
		return err
	}

	return saveAndWaitCompositionChild(act, act.storagePlugin, act.storageAction, std, workflowID)
}

func (act *actionApplyComposedPluginSubConfig) launchTarget(std *pluginaction.PluginActionStandarder) (string, error) {
	info := std.DeployInfo()
	options := info.ApplyPluginSubConfigOptions
	source, err := act.storagePlugin.GetProcess(std.Context(), info.Process.HostID, info.Process.PluginName)
	if err != nil {
		return "", fmt.Errorf("failed to get current source process: %w", err)
	}
	if source == nil || source.TenantID != std.Context().TenantID() || source.HostID != info.Process.HostID ||
		source.PluginName != info.Process.PluginName {

		return "", fmt.Errorf("source process identity mismatch")
	}
	customContext, err := composedCustomContext(options, source)
	if err != nil {
		return "", err
	}
	deploy := types.NewPluginDeployment(&types.PluginDeploymentInfo{
		Process:      types.Process{TenantID: source.TenantID, HostID: source.HostID, BizID: source.BizID, PluginName: options.PluginName},
		ConfigSource: info.ConfigSource,
	}, &types.PluginDeploymentPluginConf{
		Set: options.Set, ConfigFilesDetail: options.ConfigFilesDetail, CustomConfigContext: customContext,
	})

	return act.pluginManager.LaunchApplyPluginSubConfig(std.Context(), types.ApplyPluginSubConfigParam{
		Type: types.PluginWorkflowTypeApplyPluginSubConfig, HostIDs: []int64{source.HostID}, BizIDs: []int64{source.BizID},
		Operator: std.Operator(), PluginDeployments: []*types.PluginDeployment{deploy},
	})
}

func composedCustomContext(options types.PluginDeploymentApplyPluginSubConfigOptions, source *types.Process) (map[string]any, error) {
	if options.PluginName == "" || options.Set == "" {
		return nil, fmt.Errorf("target plugin name and config set are required")
	}
	customContext := make(map[string]any, len(options.CustomConfigContext))
	maps.Copy(customContext, options.CustomConfigContext)
	mapping := options.PluginInfoFieldMapping
	if mapping.BindIP != "" {
		if source.BindIP == "" {
			return nil, fmt.Errorf("requested source bind IP is unavailable")
		}
		if _, exists := customContext[mapping.BindIP]; exists {
			return nil, fmt.Errorf("mapped custom context key already exists: %s", mapping.BindIP)
		}
		customContext[mapping.BindIP] = source.BindIP
	}
	if mapping.BindPort != "" {
		if source.BindPort <= 0 {
			return nil, fmt.Errorf("requested source bind port is unavailable")
		}
		if _, exists := customContext[mapping.BindPort]; exists {
			return nil, fmt.Errorf("mapped custom context key already exists: %s", mapping.BindPort)
		}
		customContext[mapping.BindPort] = source.BindPort
	}
	blacklist := getBlacklistKeys()
	for key := range customContext {
		if _, exists := blacklist[key]; exists {
			return nil, fmt.Errorf("the key is reserved, cannot be used, custom-context-key(%s)", key)
		}
	}

	return customContext, nil
}
