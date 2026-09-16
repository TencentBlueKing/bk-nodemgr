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

	pluginV2Utils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/pluginv2/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameDeleteProcessConfigRecordV2 the name of action delete process config record.
	ActionNameDeleteProcessConfigRecordV2 = "delete_process_config_record_v2"
)

// NewActionDeleteProcessConfigRecordV2 new an action to delete process config record.
func NewActionDeleteProcessConfigRecordV2(capability *Capability) action.Definition {
	return &actionDeleteProcessConfigRecordV2{
		daoPluginDeployment: capability.StoragePlugin,
		daoProcessConfig:    capability.StoragePlugin,
	}
}

// ActParamDeleteProcessConfigRecordV2 defines the parameters for actionDeleteProcessConfigRecordV2.
type ActParamDeleteProcessConfigRecordV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

type actionDeleteProcessConfigRecordV2 struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoProcessConfig    pluginStg.IDaoProcessConfig
}

// Name returns the name of the action.
func (act *actionDeleteProcessConfigRecordV2) Name() string {
	return ActionNameDeleteProcessConfigRecordV2
}

// Version returns the version of the action.
func (act *actionDeleteProcessConfigRecordV2) Version() string {
	return "1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionDeleteProcessConfigRecordV2) Description() string {
	return "delete process config record v2"
}

// Timeout returns the timeout of the action.
func (act *actionDeleteProcessConfigRecordV2) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionDeleteProcessConfigRecordV2) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionDeleteProcessConfigRecordV2) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionDeleteProcessConfigRecordV2) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionDeleteProcessConfigRecordV2) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamDeleteProcessConfigRecordV2)
	err = conv.MapToStruct(ctx.Data.Content, param)
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

	pluginConf, err := act.daoPluginDeployment.GetPluginDeploymentPluginConf(nCtx, std.Token())
	if err != nil {
		std.InstanceData().Log().
			Zh("获取插件部署配置失败, %v", err).
			En("failed to get plugin deployment config, %v", err).
			Error()

		return fmt.Errorf("get plugin deployment plugin config failed, err: %w", err)
	}

	processUniqueKey := &types.ProcessUniqueKey{
		HostID: deployInfo.Process.HostID,
		Name:   deployInfo.Process.PluginName,
	}
	if pluginConf.RemoveAllConfigs {
		if err := act.daoProcessConfig.DeleteProcessConfigsByProcessUniqueKey(nCtx, processUniqueKey); err != nil {
			return fmt.Errorf("failed to delete all process configs, host-id(%d), plugin-name(%s): %w",
				deployInfo.Process.HostID, deployInfo.Process.PluginName, err)
		}

		std.InstanceData().Log().
			Zh("删除全部配置存储记录成功, 主机ID(%d), 插件名(%s)",
				deployInfo.Process.HostID, deployInfo.Process.PluginName).
			En("delete all process config records succeeded, host-id(%d), plugin-name(%s)",
				deployInfo.Process.HostID, deployInfo.Process.PluginName).
			Info()

		return nil
	}

	if err := act.daoProcessConfig.DeleteProcessConfigs(nCtx, processUniqueKey, pluginConf.RemoveConfigFileName...); err != nil {
		return fmt.Errorf("failed to delete process configs, host-id(%d), plugin-name(%s): %w",
			deployInfo.Process.HostID, deployInfo.Process.PluginName, err)
	}

	std.InstanceData().Log().
		Zh("删除配置存储记录成功, 主机ID(%d), 插件名(%s), 目标配置文件名(%v)",
			deployInfo.Process.HostID, deployInfo.Process.PluginName, pluginConf.RemoveConfigFileName).
		En("delete process config record succeed, host-id(%d), plugin-name(%s), target config file names(%v)",
			deployInfo.Process.HostID, deployInfo.Process.PluginName, pluginConf.RemoveConfigFileName).
		Info()

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionDeleteProcessConfigRecordV2) DisplayNameZh() string {
	return "删除 V2 配置存储记录"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionDeleteProcessConfigRecordV2) DisplayNameEn() string {
	return "Delete V2 Process Config Record"
}
