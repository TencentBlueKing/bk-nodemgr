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
	"time"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameMarkPluginProcessStopped defines the action name, it marks the plugin process
	// status as stopped when the plugin is delivered but not started.
	ActionNameMarkPluginProcessStopped = "mark_plugin_process_stopped"
)

// NewActionMarkPluginProcessStopped new an action that marks the plugin process as stopped.
func NewActionMarkPluginProcessStopped(capability *Capability) action.Definition {
	return &actMarkPluginProcessStopped{
		daoPluginDeployment: capability.StoragePlugin,
	}
}

// ActionParamMarkPluginProcessStopped ...
type ActionParamMarkPluginProcessStopped struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// actMarkPluginProcessStopped ...
type actMarkPluginProcessStopped struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
}

// Name returns the name of the action.
func (act *actMarkPluginProcessStopped) Name() string {
	return ActionNameMarkPluginProcessStopped
}

// Version returns the version of the action.
func (act *actMarkPluginProcessStopped) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actMarkPluginProcessStopped) Description() string {
	return "mark the plugin process status as stopped."
}

// Timeout returns the timeout of the action.
func (act *actMarkPluginProcessStopped) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actMarkPluginProcessStopped) Tags() []action.Tag {
	return []action.Tag{}
}

// Queue returns the queue of the action.
func (act *actMarkPluginProcessStopped) Queue() action.Queue {
	return action.QueueDefault
}

// MaxRetryCount returns the max retry count of the action.
func (act *actMarkPluginProcessStopped) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actMarkPluginProcessStopped) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Do this func define what the action will do.
func (act *actMarkPluginProcessStopped) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActionParamMarkPluginProcessStopped)
	if err = conv.MapToStruct(ctx.Data.Content, param); err != nil {
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

	deployInfo := std.DeployInfo()
	deployInfo.Process.Info.Pid = 0
	deployInfo.Process.Info.Status = types.ProcessStatusStopped

	std.InstanceData().Log().
		Zh("成功标记进程为已停止状态(仅下发插件, 不启动进程), plugin-name(%s), host-id(%d)",
			deployInfo.Process.PluginName, deployInfo.Process.HostID).
		En("succeed to mark process as stopped(deliver plugin only, not start process), plugin-name(%s), host-id(%d)",
			deployInfo.Process.PluginName, deployInfo.Process.HostID).
		Info()

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actMarkPluginProcessStopped) DisplayNameZh() string {
	return "标记进程为停止状态"
}

// DisplayNameEn returns the English display name of the action.
func (act *actMarkPluginProcessStopped) DisplayNameEn() string {
	return "Mark Process Stopped"
}
