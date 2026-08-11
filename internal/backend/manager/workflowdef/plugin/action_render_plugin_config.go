/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package plugin

import (
	"errors"
	"fmt"
	"maps"
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/renderer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/crypter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameRenderPluginConfig defines the action name.
	ActionNameRenderPluginConfig = "render_plugin_config"
)

// NewActionRenderPluginConfig ...
func NewActionRenderPluginConfig(capability *Capability) action.Definition {
	return &actionRenderPluginConfig{
		daoPluginDeployment: capability.StoragePlugin,
		daoProcessConfig:    capability.StoragePlugin,
	}
}

// ActParamRenderPluginConfig ...
type ActParamRenderPluginConfig struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// actionRenderPluginConfig ...
type actionRenderPluginConfig struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoProcessConfig    pluginStg.IDaoProcessConfig
}

// Name returns the name of the action.
func (act *actionRenderPluginConfig) Name() string {
	return ActionNameRenderPluginConfig
}

// Version returns the version of the action.
func (act *actionRenderPluginConfig) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actionRenderPluginConfig) Description() string {
	return "render plugin config"
}

// Timeout returns the timeout of the action.
func (act *actionRenderPluginConfig) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionRenderPluginConfig) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionRenderPluginConfig) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionRenderPluginConfig) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionRenderPluginConfig) Do(ctx *action.InstanceContext) error {
	param := new(ActParamRenderPluginConfig)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := pluginUtils.NewPluginActionStandarder(act.daoPluginDeployment)
	if err = std.Initialize(ctx, param.PluginActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	pluginConf, err := act.daoPluginDeployment.GetPluginDeploymentPluginConf(std.Context(), std.Token())
	if err != nil {
		return fmt.Errorf("failed to get plugin deployment plugin conf: %w", err)
	}

	if pluginConf == nil {
		return fmt.Errorf("plugin deployment plugin conf is nil")
	}

	if err := act.renderPluginConf(std, pluginConf); err != nil {
		return fmt.Errorf("failed to render plugin conf: %w", err)
	}

	if err := act.daoPluginDeployment.UpdatePluginDeploymentPluginConf(std.Context(), std.Token(), pluginConf); err != nil {
		return fmt.Errorf("failed to update plugin deployment plugin conf: %w", err)
	}

	configs := conv.SliceToSlice(pluginConf.ConfigFilesDetail, func(detail *types.PluginConfigDetail) *types.ProcessConfig {
		return &types.ProcessConfig{
			Name:                detail.Name,
			ProcessName:         std.DeployInfo().Process.PluginName,
			HostID:              std.DeployInfo().Process.HostID,
			IsMainConfig:        detail.IsMainConfig,
			Content:             detail.Content,
			MD5:                 crypter.MD5Sum(detail.Content),
			FilePath:            detail.FilePath,
			CustomConfigContext: pluginConf.CustomConfigContext,
		}
	})

	if err = act.daoProcessConfig.UpsertProcessConfigs(std.Context(), configs...); err != nil {
		return fmt.Errorf("failed to upsert process configs: %w", err)
	}

	std.InstanceData().Log().
		Zh("渲染插件配置操作成功完成").
		En("render plugin config action completed successfully").
		Info()

	return nil
}

func (act *actionRenderPluginConfig) renderPluginConf(std *pluginUtils.PluginActionStandarder, pluginConf *types.PluginDeploymentPluginConf) error {
	customContext := pluginConf.CustomConfigContext
	systemContext := pluginConf.SystemConfigContext
	renderContext := make(map[string]any)

	std.InstanceData().Log().
		Zh("插件模板类型为(%s)", pluginConf.TemplateRenderer).
		En("plugin template type is(%s)", pluginConf.TemplateRenderer).
		Info()

	switch pluginConf.TemplateRenderer {
	case types.TemplateRendererTypeJinja2:
		maps.Copy(renderContext, systemContext)
		maps.Copy(renderContext, customContext)
	case types.TemplateRendererTypeGoTemplate:
		maps.Copy(renderContext, systemContext)
		renderContext[keyCustomContext] = customContext
	default:
		return fmt.Errorf("unsupported template renderer type: %s", pluginConf.TemplateRenderer)
	}

	renderHandler, err := renderer.NewRenderer(pluginConf.TemplateRenderer)
	if err != nil {
		return fmt.Errorf("failed to create template renderer: %w", err)
	}

	for idx := range pluginConf.ConfigFilesDetail {
		pluginConf.ConfigFilesDetail[idx].Content, err = renderHandler.Render(pluginConf.ConfigFilesDetail[idx].Content, renderContext)
		if err != nil {
			logger.G.Sys().WithErr(err).Error("failed to render config template")
			return fmt.Errorf("failed to render sub config template: %w", err)
		}

		std.InstanceData().Log().
			Zh("渲染插件配置成功, plugin-name(%s), platform(%s), version(%s), config-file-name(%s)",
				std.DeployInfo().Process.PluginName, std.DeployInfo().Process.Platform.String(),
				std.DeployInfo().Process.Info.Version, pluginConf.ConfigFilesDetail[idx].Name).
			En("rendered plugin config success, plugin-name(%s), platform(%s), version(%s), config-file-name(%s)",
				std.DeployInfo().Process.PluginName, std.DeployInfo().Process.Platform.String(),
				std.DeployInfo().Process.Info.Version, pluginConf.ConfigFilesDetail[idx].Name).
			Info()
	}

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionRenderPluginConfig) DisplayNameZh() string {
	return "渲染插件配置"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionRenderPluginConfig) DisplayNameEn() string {
	return "Render Plugin Config"
}
