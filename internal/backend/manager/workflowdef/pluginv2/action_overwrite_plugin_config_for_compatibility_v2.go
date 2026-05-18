/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pluginv2

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	pluginV2Utils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/pluginv2/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/crypter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameOverwritePluginConfigForCompatibilityV2 defines the action name.
	ActionNameOverwritePluginConfigForCompatibilityV2 = "overwrite_plugin_config_for_compatibility_v2"
)

var pluginConfigCompatibilityOverwriteRegex = regexp.MustCompile(`(?im)^([ \t]*[^#\r\n:]*dataid[^:\r\n]*:[ \t]*).*?([ \t]+#.*)?$`)

// NewActionOverwritePluginConfigForCompatibilityV2 new an action to overwrite plugin config for compatibility.
func NewActionOverwritePluginConfigForCompatibilityV2(capability *Capability) action.Definition {
	return &actionOverwritePluginConfigForCompatibilityV2{
		daoPluginDeployment: capability.StoragePlugin,
		daoProcessV2Config:  capability.StoragePlugin,
	}
}

// ActParamOverwritePluginConfigForCompatibilityV2 defines the parameters for actionOverwritePluginConfigForCompatibilityV2.
type ActParamOverwritePluginConfigForCompatibilityV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

type actionOverwritePluginConfigForCompatibilityV2 struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoProcessV2Config  pluginStg.IDaoProcessV2Config
}

// Name returns the name of the action.
func (act *actionOverwritePluginConfigForCompatibilityV2) Name() string {
	return ActionNameOverwritePluginConfigForCompatibilityV2
}

// Version returns the version of the action.
func (act *actionOverwritePluginConfigForCompatibilityV2) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actionOverwritePluginConfigForCompatibilityV2) Description() string {
	return "overwrite plugin config for compatibility v2"
}

// Timeout returns the timeout of the action.
func (act *actionOverwritePluginConfigForCompatibilityV2) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionOverwritePluginConfigForCompatibilityV2) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionOverwritePluginConfigForCompatibilityV2) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionOverwritePluginConfigForCompatibilityV2) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionOverwritePluginConfigForCompatibilityV2) Do(ctx *action.InstanceContext) error {
	param := new(ActParamOverwritePluginConfigForCompatibilityV2)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	std := pluginV2Utils.NewPluginActionStandarder(act.daoPluginDeployment)
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
			Name:         detail.Name,
			ProcessName:  std.DeployInfo().Process.PluginName,
			HostID:       std.DeployInfo().Process.HostID,
			IsMainConfig: detail.IsMainConfig,
			Content:      detail.Content,
			MD5:          crypter.MD5Sum(detail.Content),
			FilePath:     detail.FilePath,
		})
	}
	if err = act.daoProcessV2Config.UpsertProcessV2Configs(std.Context(), configs...); err != nil {
		return fmt.Errorf("failed to upsert V2 process configs: %w", err)
	}

	std.InstanceData().Log().
		Zh("对 V2 插件配置文件进行兼容性覆写操作成功完成").
		En("overwrite V2 plugin config for compatibility action completed successfully").
		Info()

	return nil
}

func overwritePluginConfigContentForCompatibility(content string) string {
	return pluginConfigCompatibilityOverwriteRegex.ReplaceAllString(content, "${1}-1${2}")
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionOverwritePluginConfigForCompatibilityV2) DisplayNameZh() string {
	return "对 V2 插件配置文件进行兼容性覆写"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionOverwritePluginConfigForCompatibilityV2) DisplayNameEn() string {
	return "Overwrite V2 Plugin Config for Compatibility"
}
