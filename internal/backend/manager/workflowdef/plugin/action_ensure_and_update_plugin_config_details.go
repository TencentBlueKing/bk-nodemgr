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
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameEnsureAndUpdatePluginConfigDetails defines the action name.
	ActionNameEnsureAndUpdatePluginConfigDetails = "ensure_and_update_plugin_config_details"
)

// NewActionEnsureAndUpdatePluginConfigDetails ...
func NewActionEnsureAndUpdatePluginConfigDetails(capability *Capability) action.Definition {
	return &actionEnsureAndUpdatePluginConfigDetails{
		daoPluginDeployment: capability.StoragePlugin,
		daoPluginRelease:    capability.StorageRelease,
	}
}

// ActParamEnsureAndUpdatePluginConfigDetails ...
type ActParamEnsureAndUpdatePluginConfigDetails struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// actionEnsureAndUpdatePluginConfigDetails ...
type actionEnsureAndUpdatePluginConfigDetails struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoPluginRelease    release.IPlugin
}

// Name returns the name of the action.
func (act *actionEnsureAndUpdatePluginConfigDetails) Name() string {
	return ActionNameEnsureAndUpdatePluginConfigDetails
}

// Version returns the version of the action.
func (act *actionEnsureAndUpdatePluginConfigDetails) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actionEnsureAndUpdatePluginConfigDetails) Description() string {
	return "ensure and update plugin config details"
}

// Timeout returns the timeout of the action.
func (act *actionEnsureAndUpdatePluginConfigDetails) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionEnsureAndUpdatePluginConfigDetails) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionEnsureAndUpdatePluginConfigDetails) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionEnsureAndUpdatePluginConfigDetails) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionEnsureAndUpdatePluginConfigDetails) Do(ctx *action.InstanceContext) error {
	param := new(ActParamEnsureAndUpdatePluginConfigDetails)
	err := conv.MapToStruct(ctx.Data.Content, param)
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

	pluginRelease, err := act.daoPluginRelease.GetEnabledReleasePlugin(std.Context(), std.DeployInfo().Process.PluginName,
		std.DeployInfo().Process.Generation, std.DeployInfo().Process.Platform, std.DeployInfo().Process.Info.Version)
	if err != nil {
		return fmt.Errorf("failed to get plugin release info: %w", err)
	}

	configFiles, err := act.daoPluginDeployment.GetPluginDeploymentPluginConfConfigFilesDetail(std.Context(), std.Token())
	if err != nil {
		return fmt.Errorf("failed to get plugin config files detail: %w", err)
	}

	configIsMainMap := make(map[string]bool)
	for _, item := range pluginRelease.ConfigTemplates {
		if item.IsMainConfig && len(configFiles) == 0 {
			configFiles = append(configFiles, &types.PluginConfigDetail{
				Name:         item.Name,
				IsMainConfig: item.IsMainConfig,
			})
		}

		configIsMainMap[item.Name] = item.IsMainConfig
	}

	for _, cfg := range configFiles {
		if isMain, ok := configIsMainMap[cfg.Name]; ok {
			cfg.IsMainConfig = isMain
		}
	}

	if err := std.UpdatePluginConfConfigFilesDetail(configFiles...); err != nil {
		return err
	}

	return nil
}
