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
	"maps"
	"time"

	pluginV2Utils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/pluginv2/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/renderer"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/crypter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameRenderPluginConfigV2 defines the action name.
	ActionNameRenderPluginConfigV2 = "render_plugin_config_v2"
)

// NewActionRenderPluginConfigV2 ...
func NewActionRenderPluginConfigV2(capability *Capability) action.Definition {
	return &actionRenderPluginConfigV2{
		daoPluginDeployment: capability.StoragePlugin,
		daoProcessV2Config:  capability.StoragePlugin,
	}
}

// ActParamRenderPluginConfigV2 ...
type ActParamRenderPluginConfigV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

// actionRenderPluginConfigV2 ...
type actionRenderPluginConfigV2 struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoProcessV2Config  pluginStg.IDaoProcessV2Config
}

// Name returns the name of the action.
func (act *actionRenderPluginConfigV2) Name() string {
	return ActionNameRenderPluginConfigV2
}

// Version returns the version of the action.
func (act *actionRenderPluginConfigV2) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actionRenderPluginConfigV2) Description() string {
	return "render plugin config v2"
}

// Timeout returns the timeout of the action.
func (act *actionRenderPluginConfigV2) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionRenderPluginConfigV2) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionRenderPluginConfigV2) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionRenderPluginConfigV2) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionRenderPluginConfigV2) Do(ctx *action.InstanceContext) error {
	param := new(ActParamRenderPluginConfigV2)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := pluginV2Utils.NewPluginActionStandarder(act.daoPluginDeployment)
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
			Name:         detail.Name,
			ProcessName:  std.DeployInfo().Process.PluginName,
			HostID:       std.DeployInfo().Process.HostID,
			IsMainConfig: detail.IsMainConfig,
			Content:      detail.Content,
			MD5:          crypter.MD5Sum(detail.Content),
			FilePath:     detail.FilePath,
		}
	})

	if err = act.daoProcessV2Config.UpsertProcessV2Configs(std.Context(), configs...); err != nil {
		return fmt.Errorf("failed to upsert process v2 configs: %w", err)
	}

	std.InstanceData().Log().
		Zh("渲染v2插件配置操作成功完成").
		En("render v2 plugin config action completed successfully").
		Info()

	return nil
}

func (act *actionRenderPluginConfigV2) renderPluginConf(std *pluginV2Utils.PluginActionStandarder,
	pluginConf *types.PluginDeploymentPluginConf) error {

	customContext := pluginConf.CustomConfigContext
	systemContext := pluginConf.SystemConfigContext
	renderContext := make(map[string]any)

	std.InstanceData().Log().
		Zh("插件模板类型为(%s)", pluginConf.TemplateRenderer).
		En("plugin template type is(%s)", pluginConf.TemplateRenderer).
		Info()

	// V2 plugin only supports Jinja2 template renderer currently, and the render context is composed of both system context and custom context.
	switch pluginConf.TemplateRenderer {
	case types.TemplateRendererTypeJinja2:
		maps.Copy(renderContext, systemContext)
		maps.Copy(renderContext, customContext)
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
func (act *actionRenderPluginConfigV2) DisplayNameZh() string {
	return "渲染 V2 插件配置"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionRenderPluginConfigV2) DisplayNameEn() string {
	return "Render V2 Plugin Config"
}
