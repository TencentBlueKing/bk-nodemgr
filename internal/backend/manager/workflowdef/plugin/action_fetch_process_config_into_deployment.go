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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameFetchProcessSubConfigIntoDeployment the name of action fetch process sub config into deployment.
	ActionNameFetchProcessSubConfigIntoDeployment = "fetch_process_sub_config_into_deployment"
)

// NewActionFetchProcessSubConfigIntoDeployment new an action to fetch process sub config into deployment.
func NewActionFetchProcessSubConfigIntoDeployment(capability *Capability) action.Definition {
	return &actionFetchProcessSubConfigIntoDeployment{
		daoPluginDeployment: capability.StoragePlugin,
		daoProcessConfig:    capability.StoragePlugin,
	}
}

// ActParamFetchProcessSubConfigIntoDeployment defines the parameters for actionFetchProcessSubConfigIntoDeployment.
type ActParamFetchProcessSubConfigIntoDeployment struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

type actionFetchProcessSubConfigIntoDeployment struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoProcessConfig    pluginStg.IDaoProcessConfig
}

// Name returns the name of the action.
func (act *actionFetchProcessSubConfigIntoDeployment) Name() string {
	return ActionNameFetchProcessSubConfigIntoDeployment
}

// Version returns the version of the action.
func (act *actionFetchProcessSubConfigIntoDeployment) Version() string {
	return "1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionFetchProcessSubConfigIntoDeployment) Description() string {
	return "fetch process sub config into deployment"
}

// Timeout returns the timeout of the action.
func (act *actionFetchProcessSubConfigIntoDeployment) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionFetchProcessSubConfigIntoDeployment) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionFetchProcessSubConfigIntoDeployment) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionFetchProcessSubConfigIntoDeployment) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionFetchProcessSubConfigIntoDeployment) Do(ctx *action.InstanceContext) error {
	param := new(ActParamFetchProcessSubConfigIntoDeployment)
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

	nCtx := std.Context()
	deployInfo := std.DeployInfo()
	processConfigs, _, err := act.daoProcessConfig.ListProcessConfigs(nCtx, types.UnlimitedPage(),
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
		std.InstanceData().LogI(fmt.Sprintf("no process sub configs found for process(%s) on host(%d)",
			deployInfo.Process.PluginName, deployInfo.Process.HostID))

		return nil
	}

	std.InstanceData().LogI(fmt.Sprintf("get %d process sub configs for deployment", len(processConfigs)))

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
