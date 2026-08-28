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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameDeleteProcessConfigRecord the name of action delete process config record.
	ActionNameDeleteProcessConfigRecord = "delete_process_config_record"
)

// NewActionDeleteProcessConfigRecord new an action to delete process config record.
func NewActionDeleteProcessConfigRecord(capability *Capability) action.Definition {
	return &actionDeleteProcessConfigRecord{
		daoPluginDeployment:   capability.StoragePlugin,
		daoProcessConfig:      capability.StoragePlugin,
		storageActionInstance: capability.StorageWorkflow,
	}
}

// ActParamDeleteProcessConfigRecord defines the parameters for actionDeleteProcessConfigRecord.
type ActParamDeleteProcessConfigRecord struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

type actionDeleteProcessConfigRecord struct {
	daoPluginDeployment   pluginStg.IDaoPluginDeployment
	daoProcessConfig      pluginStg.IDaoProcessConfig
	storageActionInstance workflow.IStorageActionInstance
}

// Name returns the name of the action.
func (act *actionDeleteProcessConfigRecord) Name() string {
	return ActionNameDeleteProcessConfigRecord
}

// Version returns the version of the action.
func (act *actionDeleteProcessConfigRecord) Version() string {
	return "1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionDeleteProcessConfigRecord) Description() string {
	return "delete process config record"
}

// Timeout returns the timeout of the action.
func (act *actionDeleteProcessConfigRecord) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionDeleteProcessConfigRecord) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionDeleteProcessConfigRecord) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionDeleteProcessConfigRecord) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionDeleteProcessConfigRecord) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamDeleteProcessConfigRecord)
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
func (act *actionDeleteProcessConfigRecord) DisplayNameZh() string {
	return "删除配置存储记录"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionDeleteProcessConfigRecord) DisplayNameEn() string {
	return "Delete Process Config Record"
}
