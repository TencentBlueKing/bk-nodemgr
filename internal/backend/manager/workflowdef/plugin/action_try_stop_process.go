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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameTryStopProcess the name of action try stop process.
	ActionNameTryStopProcess = "try_stop_process"
)

// NewActionTryStopProcess new an action to try stop process.
func NewActionTryStopProcess(capability *Capability) action.Definition {
	return &actTryStopProcess{
		daoPluginDeployment: capability.StoragePlugin,
		daoProcess:          capability.StoragePlugin,
		daoHost:             capability.StorageTopo,
		gseHandlerProc:      capability.GSEHandler.NewHandlerProc(),
	}
}

// ActionParamTryStopProcess defines the parameters for actionTryStopProcess.
type ActionParamTryStopProcess struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// actTryStopProcess ...
type actTryStopProcess struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoProcess          pluginStg.IDaoProcess
	daoHost             topoStg.IStorageHost
	gseHandlerProc      gse.IHandlerProc
}

// Name returns the name of the action.
func (act *actTryStopProcess) Name() string {
	return ActionNameTryStopProcess
}

// Version returns the version of the action.
func (act *actTryStopProcess) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actTryStopProcess) Description() string {
	return "try stop process by gse."
}

// Timeout returns the timeout of the action.
func (act *actTryStopProcess) Timeout() time.Duration {
	return 1 * time.Minute // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actTryStopProcess) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actTryStopProcess) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actTryStopProcess) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: funlen
func (act *actTryStopProcess) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamTryStopProcess)
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

	std.InstanceData().Log().
		Zh("尝试从数据库获取进程信息").
		En("try get process info from database").
		Info()

	nCtx := std.Context()
	exist, err := act.daoProcess.ExistProcess(nCtx, std.DeployInfo().Process.HostID, std.DeployInfo().Process.PluginName)
	if err != nil {
		return fmt.Errorf("failed to check process existence: %w", err)
	}

	if !exist {
		std.InstanceData().Log().
			Zh("进程在数据库中不存在, 无需停止进程").
			En("process not exist in database, no need to stop the process.").
			Info()

		return nil
	}

	dbProcess, err := pluginUtils.GetActualExistingProcess(
		nCtx,
		act.daoProcess,
		act.daoHost,
		std.DeployInfo().Process.HostID,
		std.DeployInfo().Process.PluginName,
	)
	if err != nil {
		return fmt.Errorf("failed to get process info from database: %w", err)
	}

	hostProcessInfo, err := act.gseHandlerProc.QueryProcessInfo(nCtx, std.DeployInfo().Process.PluginName,
		dbProcess.Identity.Name, dbProcess.Info.AgentID)
	if err != nil {
		std.InstanceData().Log().
			Zh("查询主机上进程信息失败, agent-id(%s), plugin-name(%s), program-name(%s): %s",
				dbProcess.Info.AgentID, std.DeployInfo().Process.PluginName,
				dbProcess.Identity.Name, err.Error()).
			En("failed to query process info from gse, agent-id(%s), plugin-name(%s), program-name(%s): %s",
				dbProcess.Info.AgentID, std.DeployInfo().Process.PluginName,
				dbProcess.Identity.Name, err.Error()).
			Error()

		return fmt.Errorf("failed to query process info from gse: %w", err)
	}

	if hostProcessInfo.Status != types.ProcessStatusRunning {
		std.InstanceData().Log().
			Zh("主机(%d)上进程状态(%s)不为运行中, 无需停止进程",
				dbProcess.HostID, hostProcessInfo.Status).
			En("host(%d) process status(%s) is not running, no need to stop the process",
				dbProcess.HostID, hostProcessInfo.Status).
			Info()

		return nil
	}

	std.InstanceData().Log().
		Zh("主机上进程状态为运行中, 尝试执行停止插件进程, plugin-name(%s), host-id(%d), cmd(%s)",
			dbProcess.PluginName, dbProcess.HostID, dbProcess.Controller.StopCmd).
		En("process status recorded in database is running, try to executed stop plugin process, "+
			"plugin-name(%s), host-id(%d), cmd(%s)",
			dbProcess.PluginName, dbProcess.HostID, dbProcess.Controller.StopCmd).
		Info()

	std.InstanceData().Log().
		Zh("主机上进程状态, pid(%d), version(%s), agent-id(%s), autostart(%t), status(%s)",
			hostProcessInfo.Pid, hostProcessInfo.Version, hostProcessInfo.AgentID,
			hostProcessInfo.AutoStart, hostProcessInfo.Status).
		En("process record in database, pid(%d), version(%s), agent-id(%s), autostart(%t), status(%s)",
			hostProcessInfo.Pid, hostProcessInfo.Version, hostProcessInfo.AgentID,
			hostProcessInfo.AutoStart, hostProcessInfo.Status).
		Info()

	processSpec := dbProcess.ToProcessSpec()
	result, err := act.gseHandlerProc.UnTrusteeshipAndStopProcess(nCtx, processSpec)
	if err != nil {
		std.InstanceData().Log().
			Zh("执行停止插件进程操作失败, result(%s), err(%s)", result, err.Error()).
			En("failed to execute stop plugin process operation, result(%s), err(%s)", result, err.Error()).
			Warn()
	} else {
		std.InstanceData().Log().
			Zh("成功执行停止插件进程操作, result(%s)", result).
			En("successfully execute stop plugin process operation, result(%s)", result).
			Info()
	}

	std.InstanceData().Log().
		Zh("检查进程状态").
		En("check process status").
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
				Zh("进程状态仍在运行").
				En("process status is still running").
				Info()

			return errors.New("process status is still running")
		}

		if processInfo.AutoStart {
			std.InstanceData().Log().
				Zh("进程自动启动为 true, 进程将被 GSE 再次重启").
				En("process autostart is true, process will be restart by gse again").
				Info()

			return errors.New("process autostart is true, process will be restart by gse again")
		}

		// update process info
		std.DeployInfo().Process.Info = *processInfo

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to wait process no running: %w", err)
	}

	std.InstanceData().Log().
		Zh("进程已停止, pid(%d), version(%s), agent-id(%s), autostart(%t), status(%s)",
			processInfo.Pid, processInfo.Version, processInfo.AgentID, processInfo.AutoStart, processInfo.Status).
		En("process stopped, pid(%d), version(%s), agent-id(%s), autostart(%t), status(%s)",
			processInfo.Pid, processInfo.Version, processInfo.AgentID, processInfo.AutoStart, processInfo.Status).
		Info()

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actTryStopProcess) DisplayNameZh() string {
	return "尝试停止进程"
}

// DisplayNameEn returns the English display name of the action.
func (act *actTryStopProcess) DisplayNameEn() string {
	return "Try Stop Process"
}
