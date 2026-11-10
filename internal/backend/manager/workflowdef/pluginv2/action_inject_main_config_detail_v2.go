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

package pluginv2

import (
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/pluginv2/pluginv2action"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameInjectMainConfigDetailV2 defines the action name.
	ActionNameInjectMainConfigDetailV2 = "inject_main_config_detail_v2"
)

// NewActionInjectMainConfigDetailV2 get a new action.
func NewActionInjectMainConfigDetailV2(capability *Capability) action.Definition {
	return &actionInjectMainConfigDetailV2{
		daoPluginDeployment: capability.StoragePlugin,
		fileHandler:         capability.FileHandler,
	}
}

// ActParamInjectMainConfigDetailV2 this is the param for inject main config detail v2.
type ActParamInjectMainConfigDetailV2 struct {
	pluginv2action.PluginActionStandardParam `json:",inline"`
}

type actionInjectMainConfigDetailV2 struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	fileHandler         file.IReleasePluginHandler
}

// Name returns the name of the action.
func (act *actionInjectMainConfigDetailV2) Name() string {
	return ActionNameInjectMainConfigDetailV2
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionInjectMainConfigDetailV2) DisplayNameZh() string {
	return "注入 V2 插件主配置详情"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionInjectMainConfigDetailV2) DisplayNameEn() string {
	return "Inject V2 Plugin Main Config Detail"
}

// Version returns the version of the action.
func (act *actionInjectMainConfigDetailV2) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionInjectMainConfigDetailV2) Description() string {
	return "inject plugin main config detail v2"
}

// Timeout returns the timeout of the action.
func (act *actionInjectMainConfigDetailV2) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionInjectMainConfigDetailV2) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionInjectMainConfigDetailV2) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionInjectMainConfigDetailV2) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Do this func define what the action will do.
func (act *actionInjectMainConfigDetailV2) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamInjectMainConfigDetailV2)
	err = conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := pluginv2action.NewPluginActionStandarder(act.daoPluginDeployment)
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
			Zh("V2 插件发布包中没有主配置模板, plugin-name(%s), version(%s)",
				std.DeployInfo().Process.PluginName, std.DeployInfo().Process.Info.Version).
			En("no main config template found in v2 plugin release, plugin-name(%s), version(%s)",
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
		return fmt.Errorf("failed to upsert plugin deployment main config detail v2: %w", err)
	}

	std.InstanceData().Log().
		Zh("注入 %d 个 V2 插件主配置详情", len(configDetails)).
		En("inject %d v2 plugin main config details", len(configDetails)).
		Info()

	return nil
}
