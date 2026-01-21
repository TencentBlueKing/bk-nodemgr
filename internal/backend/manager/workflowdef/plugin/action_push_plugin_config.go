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
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNamePushPluginConfig the name of the action definition.
	ActionNamePushPluginConfig = "push_plugin_config"
)

// NewActionPushPluginConfig new an action.
func NewActionPushPluginConfig(capability *Capability) action.Definition {
	return &actionPushPluginConfig{
		daoPluginDeployment: capability.StoragePlugin,
		daoHost:             capability.StorageTopo,
		gseHandler:          capability.GSEHandler,
	}
}

// ActParamPushPluginConfig defines the parameters for actionPushPluginConfig.
type ActParamPushPluginConfig struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

type actionPushPluginConfig struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoHost             topoStg.IStorageHost
	gseHandler          gse.IHandler
}

// Name returns the name of the action.
func (act *actionPushPluginConfig) Name() string {
	return ActionNamePushPluginConfig
}

// Version returns the version of the action.
func (act *actionPushPluginConfig) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actionPushPluginConfig) Description() string {
	return "push plugin config by gse"
}

// Timeout returns the timeout of the action.
func (act *actionPushPluginConfig) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionPushPluginConfig) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionPushPluginConfig) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionPushPluginConfig) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionPushPluginConfig) Do(ctx *action.InstanceContext) error {
	param := new(ActParamPushPluginConfig)
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

	host, err := act.daoHost.GetHostByID(std.Context(), std.DeployInfo().Process.HostID)
	if err != nil {
		return err
	}

	if host.Dynamic.LoginUser == "" {
		return fmt.Errorf("host login user is empty, host-id(%d)", host.HostID)
	}

	pluginConf, err := act.daoPluginDeployment.GetPluginDeploymentPluginConfConfigFilesDetail(std.Context(), std.Token())
	if err != nil {
		return err
	}

	pluginDeployConf, err := deployconstant.GetPluginDeployConf(std.DeployInfo().Process.Generation, std.DeployInfo().Process.Platform.OS)
	if err != nil {
		return err
	}

	endpoints := []*types.Endpoint{{AgentID: host.Dynamic.AgentID}}
	tasks := make([]*types.PushFileDetail, 0, len(pluginConf))
	for _, pluginConfDetail := range pluginConf {
		// notice: push plugin config only push sub config files, not main config file
		if pluginConfDetail == nil || pluginConfDetail.IsMainConfig {
			continue
		}

		storeDir := pluginDeployConf.GenerateDefaultSubConfigDir(
			std.DeployInfo().Process.PluginGroup, std.DeployInfo().Process.PluginName, pluginConfDetail.FilePath,
		)

		if err = pluginUtils.CheckDirPathSafe(storeDir, std.DeployInfo().Process.Platform.OS); err != nil {
			return fmt.Errorf("check config store dir safe failed, dir(%s): %w", storeDir, err)
		}

		tasks = append(tasks, &types.PushFileDetail{
			FileName:    pluginConfDetail.Name,
			FileContent: pluginConfDetail.Content,
			StoreDir:    storeDir,
			Owner:       host.Dynamic.LoginUser,
			Endpoints:   endpoints,
		})

		std.InstanceData().LogI(fmt.Sprintf("prepare to push plugin config file(%s) to host(%d) in dir(%s)",
			pluginConfDetail.Name, host.HostID, storeDir))
	}

	if len(tasks) == 0 {
		std.InstanceData().LogI("no plugin config need to push")
		return nil
	}

	std.InstanceData().LogI(fmt.Sprintf("start to push %d plugin config files to host(%d)", len(tasks), host.HostID))

	taskID, err := act.gseHandler.PushFile(std.Context(), tasks...)
	if err != nil {
		return err
	}

	result, err := act.gseHandler.QueryPushFileFinalResult(std.Context(), taskID)
	if err != nil {
		return err
	}

	if !result.Terminated {
		return fmt.Errorf("push config not terminated, task-id(%s)", taskID)
	}

	if result.ErrorCode != 0 {
		return fmt.Errorf("push config failed, task-id(%s), err-code(%d), err-msg(%s)", taskID, result.ErrorCode, result.ErrorMessage)
	}

	std.InstanceData().LogI(fmt.Sprintf("push plugin config all done, task-id(%s).", taskID))

	return nil
}
