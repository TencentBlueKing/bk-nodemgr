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
	"errors"
	"fmt"
	"regexp"
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/crypter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameOverwritePluginConfigForCompatibility defines the action name.
	ActionNameOverwritePluginConfigForCompatibility = "overwrite_plugin_config_for_compatibility"
)

var pluginConfigCompatibilityOverwriteRegex = regexp.MustCompile(`(?im)^([ \t]*[^#\r\n:]*dataid[^:\r\n]*:[ \t]*).*?([ \t]+#.*)?$`)

// NewActionOverwritePluginConfigForCompatibility new an action to overwrite plugin config for compatibility.
func NewActionOverwritePluginConfigForCompatibility(capability *Capability) action.Definition {
	return &actionOverwritePluginConfigForCompatibility{
		daoPluginDeployment: capability.StoragePlugin,
		daoProcessConfig:    capability.StoragePlugin,
	}
}

// ActParamOverwritePluginConfigForCompatibility defines the parameters for actionOverwritePluginConfigForCompatibility.
type ActParamOverwritePluginConfigForCompatibility struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

type actionOverwritePluginConfigForCompatibility struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoProcessConfig    pluginStg.IDaoProcessConfig
}

// Name returns the name of the action.
func (act *actionOverwritePluginConfigForCompatibility) Name() string {
	return ActionNameOverwritePluginConfigForCompatibility
}

// Version returns the version of the action.
func (act *actionOverwritePluginConfigForCompatibility) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actionOverwritePluginConfigForCompatibility) Description() string {
	return "overwrite plugin config for compatibility"
}

// Timeout returns the timeout of the action.
func (act *actionOverwritePluginConfigForCompatibility) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionOverwritePluginConfigForCompatibility) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionOverwritePluginConfigForCompatibility) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionOverwritePluginConfigForCompatibility) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionOverwritePluginConfigForCompatibility) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamOverwritePluginConfigForCompatibility)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	std := pluginUtils.NewPluginActionStandarder(act.daoPluginDeployment)
	if err = std.Initialize(ctx, param.PluginActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	if !std.DeployInfo().InstallOptions.EnableCompatibilityMode {
		std.InstanceData().Log().
			Zh("未启用兼容模式，跳过插件配置文件兼容性覆写操作").
			En("compatibility mode is disabled, skip overwriting plugin config for compatibility").
			Info()

		return nil
	}

	pluginConf, err := act.daoPluginDeployment.GetPluginDeploymentPluginConf(std.Context(), std.Token())
	if err != nil {
		return fmt.Errorf("failed to get plugin deployment plugin conf: %w", err)
	}
	if pluginConf == nil {
		return fmt.Errorf("plugin deployment plugin conf is nil")
	}

	for idx := range pluginConf.ConfigFilesDetail {
		pluginConf.ConfigFilesDetail[idx].Content = overwritePluginConfigContentForCompatibility(
			pluginConf.ConfigFilesDetail[idx].Content,
		)
	}

	if err := act.daoPluginDeployment.UpdatePluginDeploymentPluginConf(std.Context(), std.Token(), pluginConf); err != nil {
		return fmt.Errorf("failed to update plugin deployment plugin conf: %w", err)
	}

	configs := make([]*types.ProcessConfig, 0, len(pluginConf.ConfigFilesDetail))
	for _, detail := range pluginConf.ConfigFilesDetail {
		configs = append(configs, &types.ProcessConfig{
			Name:                detail.Name,
			TemplateName:        detail.TemplateName,
			ProcessName:         std.DeployInfo().Process.PluginName,
			HostID:              std.DeployInfo().Process.HostID,
			IsMainConfig:        detail.IsMainConfig,
			Content:             detail.Content,
			MD5:                 crypter.MD5Sum(detail.Content),
			FilePath:            detail.FilePath,
			CustomConfigContext: pluginConf.CustomConfigContext,
		})
	}
	if err = act.daoProcessConfig.UpsertProcessConfigs(std.Context(), configs...); err != nil {
		return fmt.Errorf("failed to upsert process configs: %w", err)
	}

	std.InstanceData().Log().
		Zh("对插件配置文件进行兼容性覆写操作成功完成").
		En("overwrite plugin config for compatibility action completed successfully").
		Info()

	return nil
}

func overwritePluginConfigContentForCompatibility(content string) string {
	return pluginConfigCompatibilityOverwriteRegex.ReplaceAllString(content, "${1}-1${2}")
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionOverwritePluginConfigForCompatibility) DisplayNameZh() string {
	return "对插件配置文件进行兼容性覆写"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionOverwritePluginConfigForCompatibility) DisplayNameEn() string {
	return "Overwrite Plugin Config for Compatibility"
}
