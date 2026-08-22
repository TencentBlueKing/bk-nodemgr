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
	// ActionNameStopProcessV2 the name of action stop process.
	ActionNameStopProcessV2 = "stop_process_v2"
)

// NewActionStopProcessV2 new an action to stop process.
func NewActionStopProcessV2(capability *Capability) action.Definition {
	return &actionStopProcessV2{
		daoPluginDeployment: capability.StoragePlugin,
		daoProcess:          capability.StoragePlugin,
		gseHandlerProc:      capability.GSEHandler.NewHandlerProc(gse.WithProcNameSpace(procNameSpaceNodeMan)),
	}
}

// ActParamStopProcessV2 defines the parameters for actionStopProcess.
type ActParamStopProcessV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

// actionStopProcessV2 ...
type actionStopProcessV2 struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoProcess          pluginStg.IDaoProcess
	gseHandlerProc      gse.IHandlerProc
}

// Name returns the name of the action.
func (act *actionStopProcessV2) Name() string {
	return ActionNameStopProcessV2
}

// Version returns the version of the action.
func (act *actionStopProcessV2) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actionStopProcessV2) Description() string {
	return "stop process by gse v2."
}

// Timeout returns the timeout of the action.
func (act *actionStopProcessV2) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionStopProcessV2) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionStopProcessV2) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionStopProcessV2) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionStopProcessV2) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamStopProcessV2)
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
		Zh("尝试执行停止 V2 插件进程, plugin-name(%s), host-id(%d), cmd(%s)",
			std.DeployInfo().Process.PluginName, std.DeployInfo().Process.HostID, std.DeployInfo().Process.Controller.StopCmd).
		En("try to executed stop V2 plugin process, plugin-name(%s), host-id(%d), cmd(%s)",
			std.DeployInfo().Process.PluginName, std.DeployInfo().Process.HostID, std.DeployInfo().Process.Controller.StopCmd).
		Info()

	nCtx := std.Context()
	processSpec := std.DeployInfo().Process.ToProcessSpec()

	result, err := act.gseHandlerProc.UnTrusteeshipAndStopProcess(nCtx, processSpec)
	if err != nil {
		return fmt.Errorf("failed to stop V2 plugin process: %w", err)
	}

	std.InstanceData().Log().
		Zh("成功执行停止 V2 插件进程操作, result(%s)", result).
		En("successfully execute stop V2 plugin process operation, result(%s)", result).
		Info()

	std.InstanceData().Log().
		Zh("等待 V2 进程停止").
		En("wait V2 process stopped").
		Info()
	polling := retrier.NewPolling(retrier.PollingOpts{
		Timeout:  act.Timeout(),
		Interval: time.Second,
	})

	var processInfo *types.ProcessInfo
	err = polling.Do(nCtx, func(_ int) error {
		processInfo, err = act.gseHandlerProc.QueryProcessInfo(nCtx, processSpec.PluginName, processSpec.Identity.Name, processSpec.AgentID)
		if err != nil {
			return fmt.Errorf("failed to query process info: %w", err)
		}

		if processInfo.Status != types.ProcessStatusStopped {
			std.InstanceData().Log().
				Zh("V2 进程状态未停止, status(%s)", processInfo.Status).
				En("V2 process status is not stopped, status(%s)", processInfo.Status).
				Info()

			return fmt.Errorf("V2 process status is not stopped, status(%s)", processInfo.Status)
		}

		if processInfo.AutoStart {
			std.InstanceData().Log().
				Zh("V2 进程仍被 GSE 托管").
				En("V2 process is still trusteeship by gse").
				Info()

			return fmt.Errorf("V2 process is still trusteeship by gse")
		}

		// update process info
		std.DeployInfo().Process.Info = *processInfo

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to wait V2 process stopped: %w", err)
	}

	std.InstanceData().Log().
		Zh("V2 进程已停止, pid(%d), version(%s), agent-id(%s), autostart(%t), status(%s)",
			processInfo.Pid, processInfo.Version, processInfo.AgentID, processInfo.AutoStart, processInfo.Status).
		En("V2 process stopped, pid(%d), version(%s), agent-id(%s), autostart(%t), status(%s)",
			processInfo.Pid, processInfo.Version, processInfo.AgentID, processInfo.AutoStart, processInfo.Status).
		Info()

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionStopProcessV2) DisplayNameZh() string {
	return "停止 V2 进程"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionStopProcessV2) DisplayNameEn() string {
	return "Stop V2 Process"
}
