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
	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNamePrepareSubConfigSource prepares and waits for the source plugin.
	ActionNamePrepareSubConfigSource = "prepare_subconfig_source"
	// ActionNameApplyComposedPluginSubConfig applies target configs after source completion.
	ActionNameApplyComposedPluginSubConfig = "apply_composed_plugin_subconfig"
)

// NewActionPrepareSubConfigSource creates the source preparation action.
func NewActionPrepareSubConfigSource(capability *Capability) action.Definition {
	return &actionComposePluginSubConfig{
		storagePlugin: capability.StoragePlugin,
		storageAction: capability.StorageWorkflow,
		pluginManager: capability.PluginIface,
		name:          ActionNamePrepareSubConfigSource,
	}
}

// NewActionApplyComposedPluginSubConfig creates the target config action.
func NewActionApplyComposedPluginSubConfig(capability *Capability) action.Definition {
	return &actionComposePluginSubConfig{
		storagePlugin: capability.StoragePlugin,
		storageAction: capability.StorageWorkflow,
		pluginManager: capability.PluginIface,
		name:          ActionNameApplyComposedPluginSubConfig,
	}
}

type actionComposePluginSubConfig struct {
	storagePlugin pluginStg.IStorage
	storageAction workflow.IStorageActionInstance
	pluginManager managerIface.IPluginManager
	name          string
}

// Name returns the action name.
func (act *actionComposePluginSubConfig) Name() string { return act.name }

// DisplayNameZh returns the action display name.
func (act *actionComposePluginSubConfig) DisplayNameZh() string {
	if act.name == ActionNamePrepareSubConfigSource {
		return "准备子配置源插件"
	}

	return "下发组合插件子配置"
}

// DisplayNameEn returns the action display name.
func (act *actionComposePluginSubConfig) DisplayNameEn() string {
	if act.name == ActionNamePrepareSubConfigSource {
		return "Prepare Sub-config Source"
	}

	return "Apply Composed Plugin Sub-config"
}

// Version returns the action version.
func (act *actionComposePluginSubConfig) Version() string { return "1.0.0" }

// Description returns the action description.
func (act *actionComposePluginSubConfig) Description() string { return act.DisplayNameEn() }

// Timeout returns the child workflow wait timeout.
func (act *actionComposePluginSubConfig) Timeout() time.Duration { return 10 * time.Minute } // nolint: mnd

// Tags returns action tags.
func (act *actionComposePluginSubConfig) Tags() []action.Tag { return []action.Tag{} }

// MaxRetryCount returns the action retry count.
func (act *actionComposePluginSubConfig) MaxRetryCount() uint { return 3 } // nolint: mnd

// DelayFn returns the standard retry delay.
func (act *actionComposePluginSubConfig) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Do launches one child workflow and waits before the next composition action can run.
func (act *actionComposePluginSubConfig) Do(ctx *action.InstanceContext) error {
	param := new(OperParamComposePluginSubConfig)
	if err := conv.MapToStruct(ctx.Data.Content, param); err != nil {
		return fmt.Errorf("failed to parse composition parameters: %w", err)
	}
	if err := param.SourceOperation.Validate(); err != nil {
		return err
	}
	std := pluginUtils.NewPluginActionStandarder(act.storagePlugin)
	if err := std.Initialize(ctx, param.PluginActionStandardParam); err != nil {
		return err
	}

	workflowID, err := act.launchChild(std, param.SourceOperation)
	if err != nil {
		return err
	}
	// Empty workflow ID means no source operation was requested.
	if workflowID == "" {
		return nil
	}
	refs := []types.SubWorkflowRef{{WorkflowID: workflowID, WorkflowDomain: types.WorkflowDomainPlugin}}
	serialized, err := types.SerializeSubWorkflowRefs(refs)
	if err != nil {
		return fmt.Errorf("failed to serialize composition child references: %w", err)
	}
	ctx.Data.PrivateData[types.PDKeySubWorkflowRefs] = serialized
	if err := act.storageAction.UpsertActionInstancePrivateData(std.Context(), ctx.Data.OperationInstanceID,
		act.Name(), map[string]any{types.PDKeySubWorkflowRefs: serialized}); err != nil {
		return fmt.Errorf("failed to save composition child references: %w", err)
	}
	for _, ref := range refs {
		if err := act.waitChild(std, ref.WorkflowID); err != nil {
			return err
		}
	}

	return nil
}

func (act *actionComposePluginSubConfig) launchChild(std *pluginUtils.PluginActionStandarder,
	sourceOperation types.PluginSubConfigSourceOperation) (string, error) {

	if act.name == ActionNameApplyComposedPluginSubConfig {
		return act.launchTarget(std)
	}
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

func (act *actionComposePluginSubConfig) launchTarget(std *pluginUtils.PluginActionStandarder) (string, error) {
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

func (act *actionComposePluginSubConfig) waitChild(std *pluginUtils.PluginActionStandarder, workflowID string) error {
	var status types.PluginWorkflowStatus
	polling := retrier.NewPolling(retrier.PollingOpts{Timeout: act.Timeout(), Interval: 10 * time.Second}) // nolint: mnd
	if err := polling.Do(std.Context(), func(_ int) error {
		var err error
		status, err = act.storagePlugin.GetPluginWorkflowStatus(std.Context(), workflowID)
		if err != nil {
			return fmt.Errorf("failed to get child workflow status: %w", err)
		}
		if status == types.PluginWorkflowStatusRunning {
			return fmt.Errorf("child workflow is still running: %s", workflowID)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to wait for child workflow(%s): %w", workflowID, err)
	}
	if status != types.PluginWorkflowStatusSuccess {
		return fmt.Errorf("child workflow(%s) ended with status(%s)", workflowID, status)
	}

	return nil
}
