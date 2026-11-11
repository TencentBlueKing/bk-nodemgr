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
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameInjectMainConfigDetail defines the action name.
	ActionNameInjectMainConfigDetail = "inject_main_config_detail"
)

// NewActionInjectMainConfigDetail get a new action.
func NewActionInjectMainConfigDetail(capability *Capability) action.Definition {
	return &actionInjectMainConfigDetail{
		daoPluginDeployment: capability.StoragePlugin,
		fileHandler:         capability.FileHandler,
	}
}

// ActParamInjectMainConfigDetail this is the param for inject main config detail.
type ActParamInjectMainConfigDetail struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

type actionInjectMainConfigDetail struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	fileHandler         file.IReleasePluginHandler
}

// Name returns the name of the action.
func (act *actionInjectMainConfigDetail) Name() string {
	return ActionNameInjectMainConfigDetail
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionInjectMainConfigDetail) DisplayNameZh() string {
	return "注入主配置详情"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionInjectMainConfigDetail) DisplayNameEn() string {
	return "Inject Main Config Detail"
}

// Version returns the version of the action.
func (act *actionInjectMainConfigDetail) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionInjectMainConfigDetail) Description() string {
	return "inject main config detail"
}

// Timeout returns the timeout of the action.
func (act *actionInjectMainConfigDetail) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionInjectMainConfigDetail) Tags() []action.Tag {
	return []action.Tag{}
}

// Queue returns the queue of the action.
func (act *actionInjectMainConfigDetail) Queue() action.Queue {
	return action.QueueDefault
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionInjectMainConfigDetail) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionInjectMainConfigDetail) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Do this func define what the action will do.
func (act *actionInjectMainConfigDetail) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamInjectMainConfigDetail)
	err = conv.MapToStruct(ctx.Data.Content, param)
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

	pluginRelease, err := act.fileHandler.GetReleasePlugin(std.Context(), types.ReleasePluginKey{
		Generation: std.DeployInfo().Process.Generation,
		Platform:   std.DeployInfo().Process.Platform,
		Version:    std.DeployInfo().Process.Info.Version,
		Name:       std.DeployInfo().Process.PluginPkgName,
	})
	if err != nil {
		return fmt.Errorf("failed to get plugin release info: %w", err)
	}

	mainConfigKeys := pluginRelease.MainConfigTemplateKeys()
	if len(mainConfigKeys) == 0 {
		std.InstanceData().Log().
			Zh("插件发布包中没有主配置模板, plugin-name(%s), version(%s)",
				std.DeployInfo().Process.PluginName, std.DeployInfo().Process.Info.Version).
			En("no main config template found in plugin release, plugin-name(%s), version(%s)",
				std.DeployInfo().Process.PluginName, std.DeployInfo().Process.Info.Version).
			Info()

		return nil
	}

	configDetails := conv.SliceToSlice(mainConfigKeys, func(key types.PluginPkgConfigTemplateKey) *types.PluginConfigDetail {
		return &types.PluginConfigDetail{
			Name:         key.Name,
			TemplateName: key.Name,
			IsMainConfig: key.IsMainConfig,
		}
	})
	if err := act.daoPluginDeployment.UpsertPluginDeploymentPluginConfConfigFilesDetail(
		std.Context(), std.Token(), configDetails...); err != nil {
		return fmt.Errorf("failed to upsert plugin deployment main config detail: %w", err)
	}

	std.InstanceData().Log().
		Zh("注入 %d 个主配置详情", len(configDetails)).
		En("inject %d main config details", len(configDetails)).
		Info()

	return nil
}
