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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameCheckPluginProcessAlive the name of action check plugin process alive.
	ActionNameCheckPluginProcessAlive = "check_plugin_process_alive"
)

// NewActionCheckPluginProcessAlive new an action to check plugin process alive.
func NewActionCheckPluginProcessAlive(capability *Capability) action.Definition {
	return &actionCheckPluginProcessAlive{
		daoPluginDeployment: capability.StoragePlugin,
		daoProcess:          capability.StoragePlugin,
		gseHandlerProc:      capability.GSEHandler.NewHandlerProc(),
	}
}

// ActParamCheckPluginProcessAlive defines the parameters for actionCheckPluginProcessAlive.
type ActParamCheckPluginProcessAlive struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

type actionCheckPluginProcessAlive struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoProcess          pluginStg.IDaoProcess
	gseHandlerProc      gse.IHandlerProc
}

// Name returns the name of the action.
func (act *actionCheckPluginProcessAlive) Name() string {
	return ActionNameCheckPluginProcessAlive
}

// Version returns the version of the action.
func (act *actionCheckPluginProcessAlive) Version() string {
	return "1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionCheckPluginProcessAlive) Description() string {
	return "check plugin process alive"
}

// Timeout returns the timeout of the action.
func (act *actionCheckPluginProcessAlive) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionCheckPluginProcessAlive) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionCheckPluginProcessAlive) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionCheckPluginProcessAlive) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionCheckPluginProcessAlive) Do(ctx *action.InstanceContext) error {
	param := new(ActParamCheckPluginProcessAlive)
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
	processSpec := deployInfo.Process.ToProcessSpec()
	processInfo, err := act.gseHandlerProc.QueryProcessInfo(nCtx, processSpec.PluginName, processSpec.Identity.Name, processSpec.AgentID)
	if err != nil {
		return fmt.Errorf("failed to query process info: %w", err)
	}

	if processInfo.Status != types.ProcessStatusRunning {
		std.InstanceData().Log().
			Zh("进程状态未运行, status(%s)", processInfo.Status).
			En("process status is not running, status(%s)", processInfo.Status).
			Info()

		return fmt.Errorf("process status is not running, status(%s)", processInfo.Status)
	}

	if deployInfo.Process.Info.Version != processInfo.Version {
		std.InstanceData().Log().
			Zh("进程版本不匹配, 记录版本(%s), 实际版本(%s)",
				deployInfo.Process.Info.Version, processInfo.Version).
			En("process version mismatch, record-version(%s), actual-version(%s)",
				deployInfo.Process.Info.Version, processInfo.Version).
			Info()
	}

	std.InstanceData().Log().
		Zh("检查插件进程存活成功, plugin-name(%s), host-id(%d), status(%s), version(%s)",
			deployInfo.Process.PluginName, deployInfo.Process.HostID, processInfo.Status, processInfo.Version).
		En("check plugin process alive succeed, plugin-name(%s), host-id(%d), status(%s), version(%s)",
			deployInfo.Process.PluginName, deployInfo.Process.HostID, processInfo.Status, processInfo.Version).
		Info()

	// update process info
	deployInfo.Process.Info = *processInfo

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionCheckPluginProcessAlive) DisplayNameZh() string {
	return "检查插件进程存活"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionCheckPluginProcessAlive) DisplayNameEn() string {
	return "Check Plugin Process Alive"
}
