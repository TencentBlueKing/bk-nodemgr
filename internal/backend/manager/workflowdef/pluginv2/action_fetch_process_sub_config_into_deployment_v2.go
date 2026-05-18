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
	"time"

	pluginV2Utils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/pluginv2/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameFetchProcessSubConfigIntoDeploymentV2 the name of action fetch process sub config into deployment.
	ActionNameFetchProcessSubConfigIntoDeploymentV2 = "fetch_process_sub_config_into_deployment_v2"
)

// NewActionFetchProcessSubConfigIntoDeploymentV2 new an action to fetch process sub config into deployment.
func NewActionFetchProcessSubConfigIntoDeploymentV2(capability *Capability) action.Definition {
	return &actionFetchProcessSubConfigIntoDeploymentV2{
		daoPluginDeployment: capability.StoragePlugin,
		daoProcessV2Config:  capability.StoragePlugin,
	}
}

// ActParamFetchProcessSubConfigIntoDeploymentV2 defines the parameters for actionFetchProcessSubConfigIntoDeploymentV2.
type ActParamFetchProcessSubConfigIntoDeploymentV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

type actionFetchProcessSubConfigIntoDeploymentV2 struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoProcessV2Config  pluginStg.IDaoProcessV2Config
}

// Name returns the name of the action.
func (act *actionFetchProcessSubConfigIntoDeploymentV2) Name() string {
	return ActionNameFetchProcessSubConfigIntoDeploymentV2
}

// Version returns the version of the action.
func (act *actionFetchProcessSubConfigIntoDeploymentV2) Version() string {
	return "1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionFetchProcessSubConfigIntoDeploymentV2) Description() string {
	return "fetch process sub config into deployment v2"
}

// Timeout returns the timeout of the action.
func (act *actionFetchProcessSubConfigIntoDeploymentV2) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionFetchProcessSubConfigIntoDeploymentV2) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionFetchProcessSubConfigIntoDeploymentV2) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionFetchProcessSubConfigIntoDeploymentV2) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionFetchProcessSubConfigIntoDeploymentV2) Do(ctx *action.InstanceContext) error {
	param := new(ActParamFetchProcessSubConfigIntoDeploymentV2)
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

	nCtx := std.Context()
	deployInfo := std.DeployInfo()
	processConfigs, _, err := act.daoProcessV2Config.ListProcessV2Configs(nCtx, types.UnlimitedPage(),
		&types.ProcessConfigCondition{
			ExactInclude: &types.ProcessConfigExactFields{
				ProcessName:  []string{deployInfo.Process.PluginName},
				HostID:       []int64{deployInfo.Process.HostID},
				IsMainConfig: []bool{false},
			},
		},
	)
	if err != nil {
		return err
	}

	if len(processConfigs) == 0 {
		std.InstanceData().Log().
			Zh("未找到 V2 进程子配置, process(%s), host(%d)",
				deployInfo.Process.PluginName, deployInfo.Process.HostID).
			En("no V2 process sub configs found for process(%s) on host(%d)",
				deployInfo.Process.PluginName, deployInfo.Process.HostID).
			Info()

		return nil
	}

	std.InstanceData().Log().
		Zh("获取 %d 个 V2 进程子配置用于部署", len(processConfigs)).
		En("get %d V2 process sub configs for deployment", len(processConfigs)).
		Info()

	configDetails := conv.SliceToSlice(processConfigs, func(processConfig *types.ProcessConfig) *types.PluginConfigDetail {
		return &types.PluginConfigDetail{
			Name:         processConfig.Name,
			Content:      processConfig.Content,
			IsMainConfig: processConfig.IsMainConfig,
			FilePath:     processConfig.FilePath,
		}
	})
	if err := act.daoPluginDeployment.UpsertPluginDeploymentPluginConfConfigFilesDetail(std.Context(), std.Token(), configDetails...); err != nil {
		return fmt.Errorf("failed to upsert plugin deployment plugin conf: %w", err)
	}

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionFetchProcessSubConfigIntoDeploymentV2) DisplayNameZh() string {
	return "获取 V2 进程配置到部署"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionFetchProcessSubConfigIntoDeploymentV2) DisplayNameEn() string {
	return "Fetch V2 Process Config into Deployment"
}
