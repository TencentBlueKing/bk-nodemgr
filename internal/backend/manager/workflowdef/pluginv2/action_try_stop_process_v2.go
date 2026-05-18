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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameTryStopProcessV2 the name of action try stop process.
	ActionNameTryStopProcessV2 = "try_stop_process_v2"
)

// NewActionTryStopProcessV2 new an action to try stop process.
func NewActionTryStopProcessV2(capability *Capability) action.Definition {
	return &actionTryStopProcessV2{
		daoPluginDeployment: capability.StoragePlugin,
		daoProcessV2:        capability.StoragePlugin,
		daoHost:             capability.StorageTopo,
		gseHandlerProc:      capability.GSEHandler.NewHandlerProc(gse.WithProcNameSpace(procNameSpaceNodeMan)),
	}
}

// ActParamTryStopProcessV2 defines the parameters for actionTryStopProcess.
type ActParamTryStopProcessV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

// actionTryStopProcessV2 ...
type actionTryStopProcessV2 struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoProcessV2        pluginStg.IDaoProcessV2
	daoHost             topoStg.IStorageHost
	gseHandlerProc      gse.IHandlerProc
}

// Name returns the name of the action.
func (act *actionTryStopProcessV2) Name() string {
	return ActionNameTryStopProcessV2
}

// Version returns the version of the action.
func (act *actionTryStopProcessV2) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actionTryStopProcessV2) Description() string {
	return "try stop process by gse v2."
}

// Timeout returns the timeout of the action.
func (act *actionTryStopProcessV2) Timeout() time.Duration {
	return 1 * time.Minute // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionTryStopProcessV2) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionTryStopProcessV2) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionTryStopProcessV2) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen
func (act *actionTryStopProcessV2) Do(ctx *action.InstanceContext) error {
	param := new(ActParamTryStopProcessV2)
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

	std.InstanceData().Log().
		Zh("尝试从数据库获取进程信息").
		En("try get process info from database").
		Info()

	nCtx := std.Context()
	exist, err := act.daoProcessV2.ExistProcessV2(nCtx, std.DeployInfo().Process.HostID, std.DeployInfo().Process.PluginName)
	if err != nil {
		return fmt.Errorf("failed to check process v2 existence: %w", err)
	}

	if !exist {
		std.InstanceData().Log().
			Zh("V2 进程不存在, 无需停止进程").
			En("V2 process not exist, no need to stop the process.").
			Info()

		return nil
	}

	process, err := pluginV2Utils.GetActualExistingProcess(
		nCtx,
		act.daoProcessV2,
		act.daoHost,
		std.DeployInfo().Process.HostID,
		std.DeployInfo().Process.PluginName,
	)
	if err != nil {
		return fmt.Errorf("failed to get process v2 info: %w", err)
	}

	if process.Info.Status != types.ProcessStatusRunning {
		std.InstanceData().Log().
			Zh("数据库中记录的 V2 进程状态未运行(%s), 无需停止进程", process.Info.Status).
			En("V2 process status recorded in database is not running(%s), no need to stop the process", process.Info.Status).
			Info()

		return nil
	}

	std.InstanceData().Log().
		Zh("数据库中记录的 V2 进程状态为运行中, 尝试执行停止插件进程, plugin-name(%s), host-id(%d), cmd(%s)",
			process.PluginName, process.HostID, process.Controller.StopCmd).
		En("V2 process status recorded in database is running, try to executed stop plugin process, plugin-name(%s), host-id(%d), cmd(%s)",
			process.PluginName, process.HostID, process.Controller.StopCmd).
		Info()
	std.InstanceData().Log().
		Zh("数据库中的 V2 进程记录, pid(%d), version(%s), agent-id(%s), autostart(%t), status(%s)",
			process.Info.Pid, process.Info.Version, process.Info.AgentID, process.Info.AutoStart, process.Info.Status).
		En("V2 process record in database, pid(%d), version(%s), agent-id(%s), autostart(%t), status(%s)",
			process.Info.Pid, process.Info.Version, process.Info.AgentID, process.Info.AutoStart, process.Info.Status).
		Info()

	processSpec := process.ToProcessSpec()

	result, err := act.gseHandlerProc.UnTrusteeshipAndStopProcess(nCtx, processSpec)
	if err != nil {
		std.InstanceData().Log().
			Zh("执行停止 V2 插件进程操作失败, result(%s), err(%s)", result, err.Error()).
			En("failed to execute stop V2 plugin process operation, result(%s), err(%s)", result, err.Error()).
			Warn()
	} else {
		std.InstanceData().Log().
			Zh("成功执行停止 V2 插件进程操作, result(%s)", result).
			En("successfully execute stop V2 plugin process operation, result(%s)", result).
			Info()
	}

	std.InstanceData().Log().
		Zh("检查 V2 进程状态").
		En("check V2 process status").
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

		if processInfo.Status == types.ProcessStatusRunning {
			std.InstanceData().Log().
				Zh("V2 进程状态仍在运行").
				En("V2 process status is still running").
				Info()

			return errors.New("V2 process status is still running")
		}

		if processInfo.AutoStart {
			std.InstanceData().Log().
				Zh("V2 进程自动启动为 true, 进程将被 GSE 再次重启").
				En("V2 process autostart is true, process will be restart by gse again").
				Info()

			return errors.New("V2 process autostart is true, process will be restart by gse again")
		}

		// update process info
		std.DeployInfo().Process.Info = *processInfo

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to wait V2 process no running: %w", err)
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
func (act *actionTryStopProcessV2) DisplayNameZh() string {
	return "尝试停止 V2 进程"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionTryStopProcessV2) DisplayNameEn() string {
	return "Try Stop V2 Process"
}
