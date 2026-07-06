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
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNamePushPluginConfigV2 the name of the action definition.
	ActionNamePushPluginConfigV2 = "push_plugin_config_v2"
)

// NewActionPushPluginConfigV2 new an action.
func NewActionPushPluginConfigV2(capability *Capability) action.Definition {
	return &actionPushPluginConfigV2{
		daoPluginDeployment: capability.StoragePlugin,
		daoHost:             capability.StorageTopo,
		gseHandler:          capability.GSEHandler,
	}
}

// ActParamPushPluginConfigV2 defines the parameters for actionPushPluginConfigV2.
type ActParamPushPluginConfigV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

type actionPushPluginConfigV2 struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoHost             topoStg.IStorageHost
	gseHandler          gse.IHandler
}

// Name returns the name of the action.
func (act *actionPushPluginConfigV2) Name() string {
	return ActionNamePushPluginConfigV2
}

// Version returns the version of the action.
func (act *actionPushPluginConfigV2) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actionPushPluginConfigV2) Description() string {
	return "push plugin config by gse v2"
}

// Timeout returns the timeout of the action.
func (act *actionPushPluginConfigV2) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionPushPluginConfigV2) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionPushPluginConfigV2) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionPushPluginConfigV2) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionPushPluginConfigV2) Do(ctx *action.InstanceContext) error {
	param := new(ActParamPushPluginConfigV2)
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

	host, err := act.daoHost.GetHostByID(std.Context(), std.DeployInfo().Process.HostID)
	if err != nil {
		return err
	}

	if err = pluginV2Utils.EnsureHostLoginUser(std, host); err != nil {
		return err
	}

	pluginConf, err := act.daoPluginDeployment.GetPluginDeploymentPluginConfConfigFilesDetail(std.Context(), std.Token())
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

		if err = pluginV2Utils.CheckDirPathSafe(std.DeployInfo().BaseRuntime.SubConfigDir, std.DeployInfo().Process.Platform.OS); err != nil {
			return fmt.Errorf("check config store dir safe failed, dir(%s): %w", std.DeployInfo().BaseRuntime.SubConfigDir, err)
		}

		tasks = append(tasks, &types.PushFileDetail{
			FileName:    pluginConfDetail.Name,
			FileContent: pluginConfDetail.Content,
			StoreDir:    std.DeployInfo().BaseRuntime.SubConfigDir,
			Owner:       host.Dynamic.LoginUser,
			Endpoints:   endpoints,
		})

		std.InstanceData().Log().
			Zh("准备推送插件配置文件(%s)到主机(%d), 目录(%s)",
				pluginConfDetail.Name, host.HostID, std.DeployInfo().BaseRuntime.SubConfigDir).
			En("prepare to push plugin config file(%s) to host(%d) in dir(%s)",
				pluginConfDetail.Name, host.HostID, std.DeployInfo().BaseRuntime.SubConfigDir).
			Info()
	}

	if len(tasks) == 0 {
		std.InstanceData().Log().
			Zh("无需推送插件配置").
			En("no plugin config need to push").
			Info()

		return nil
	}

	std.InstanceData().Log().
		Zh("开始推送 %d 个插件配置文件到主机(%d)", len(tasks), host.HostID).
		En("start to push %d plugin config files to host(%d)", len(tasks), host.HostID).
		Info()

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

	std.InstanceData().Log().
		Zh("推送插件配置全部完成, task-id(%s)", taskID).
		En("push plugin config all done, task-id(%s).", taskID).
		Info()

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionPushPluginConfigV2) DisplayNameZh() string {
	return "推送 V2 插件配置"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionPushPluginConfigV2) DisplayNameEn() string {
	return "Push V2 Plugin Config"
}
