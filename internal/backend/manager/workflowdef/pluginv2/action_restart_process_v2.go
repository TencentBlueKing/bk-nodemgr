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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameRestartProcessV2 the name of action restart process.
	ActionNameRestartProcessV2 = "restart_process_v2"
)

// NewActionRestartProcessV2 new an action to restart process.
func NewActionRestartProcessV2(capability *Capability) action.Definition {
	return &actionRestartProcessV2{
		daoPluginDeployment: capability.StoragePlugin,
		gseHandlerProc:      capability.GSEHandler.NewHandlerProc(gse.WithProcNameSpace(procNameSpaceNodeMan)),
	}
}

// ActParamRestartProcessV2 defines the parameters for actionRestartProcessV2.
type ActParamRestartProcessV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

// actRestartProcess ...
type actionRestartProcessV2 struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	gseHandlerProc      gse.IHandlerProc
}

// Name returns the name of the action.
func (act *actionRestartProcessV2) Name() string {
	return ActionNameRestartProcessV2
}

// Version returns the version of the action.
func (act *actionRestartProcessV2) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actionRestartProcessV2) Description() string {
	return "restart process by gse v2."
}

// Timeout returns the timeout of the action.
func (act *actionRestartProcessV2) Timeout() time.Duration {
	return 5 * time.Minute // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionRestartProcessV2) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionRestartProcessV2) MaxRetryCount() uint {
	return 5 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionRestartProcessV2) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionRestartProcessV2) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamRestartProcessV2)
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

	std.InstanceData().Log().
		Zh("尝试执行重启插件进程, plugin-name(%s), host-id(%d), cmd(%s)",
			std.DeployInfo().Process.PluginName, std.DeployInfo().Process.HostID, std.DeployInfo().Process.Controller.RestartCmd).
		En("try to executed restart plugin process, plugin-name(%s), host-id(%d), cmd(%s)",
			std.DeployInfo().Process.PluginName, std.DeployInfo().Process.HostID, std.DeployInfo().Process.Controller.RestartCmd).
		Info()

	nCtx := std.Context()
	processSpec := std.DeployInfo().Process.ToProcessSpec()
	result, err := act.gseHandlerProc.TrusteeshipAndRestartProcess(nCtx, processSpec)
	if err != nil {
		return fmt.Errorf("failed to restart plugin process: %w", err)
	}

	std.InstanceData().Log().
		Zh("成功执行重启插件进程操作, result(%s)", result).
		En("successfully execute restart plugin process operation, result(%s)", result).
		Info()

	std.InstanceData().Log().
		Zh("等待进程运行").
		En("wait process running").
		Info()
	polling := retrier.NewPolling(retrier.PollingOpts{
		Timeout:  act.Timeout(),
		Interval: time.Second,
	})

	var processInfo *types.ProcessInfo
	err = polling.Do(std.Context(), func(_ int) error {
		processInfo, err = act.gseHandlerProc.QueryProcessInfo(std.Context(), processSpec.PluginName, processSpec.Identity.Name, processSpec.AgentID)
		if err != nil {
			return fmt.Errorf("failed to query process info, plugin-name(%s), process-identity-name(%s), agent-id(%s): %w",
				processSpec.PluginName, processSpec.Identity.Name, processSpec.AgentID, err)
		}

		if processInfo.Status != types.ProcessStatusRunning {
			std.InstanceData().Log().
				Zh("进程状态未运行, status(%s)", processInfo.Status).
				En("process status is not running, status(%s)", processInfo.Status).
				Info()

			return fmt.Errorf("process status is not running, status(%s)", processInfo.Status)
		}

		// update process info
		std.DeployInfo().Process.Info = *processInfo

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to wait process running: %w", err)
	}

	std.InstanceData().Log().
		Zh("进程运行中, pid(%d), version(%s), agent-id(%s), autostart(%t), status(%s)",
			processInfo.Pid, processInfo.Version, processInfo.AgentID, processInfo.AutoStart, processInfo.Status).
		En("process running, pid(%d), version(%s), agent-id(%s), autostart(%t), status(%s)",
			processInfo.Pid, processInfo.Version, processInfo.AgentID, processInfo.AutoStart, processInfo.Status).
		Info()

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionRestartProcessV2) DisplayNameZh() string {
	return "重启 V2 进程"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionRestartProcessV2) DisplayNameEn() string {
	return "Restart V2 Process"
}
