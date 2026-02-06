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
		gseHandlerProc:      capability.GSEHandler,
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
	return 30 * time.Second // nolint: mnd
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

	std.InstanceData().LogI("try get process info from database")

	nCtx := std.Context()
	exist, err := act.daoProcess.ExistProcess(nCtx, std.DeployInfo().Process.HostID, std.DeployInfo().Process.PluginName)
	if err != nil {
		return fmt.Errorf("failed to check process existence: %w", err)
	}

	if !exist {
		std.InstanceData().LogI("process not exist, no need to stop the process.")

		return nil
	}

	process, err := act.daoProcess.GetProcess(nCtx, std.DeployInfo().Process.HostID, std.DeployInfo().Process.PluginName)
	if err != nil {
		return fmt.Errorf("failed to get process info: %w", err)
	}

	if process.Info.Status != types.ProcessStatusRunning {
		std.InstanceData().LogI(
			fmt.Sprintf("process status recorded in database is not running(%s), no need to stop the process", process.Info.Status))

		return nil
	}

	std.InstanceData().LogI(
		fmt.Sprintf("process status recorded in database is running, try to executed stop plugin process, plugin-name(%s), host-id(%d), cmd(%s)",
			process.PluginName, process.HostID, process.Controller.StopCmd),
		fmt.Sprintf("process record in database, pid(%d), version(%s), agent-id(%s), autostart(%t), status(%s)",
			process.Info.Pid, process.Info.Version, process.Info.AgentID, process.Info.AutoStart, process.Info.Status),
	)

	processSpec := process.ToProcessSpec()

	result, err := act.gseHandlerProc.UnTrusteeshipAndStopProcess(nCtx, processSpec)
	if err != nil {
		std.InstanceData().LogW(fmt.Sprintf("failed to execute stop plugin process operation, result(%s), err(%s)", result, err.Error()))
	} else {
		std.InstanceData().LogI(fmt.Sprintf("successfully execute stop plugin process operation, result(%s)", result))
	}

	std.InstanceData().LogI("check process status")
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
			std.InstanceData().LogI("process status is still running")

			return errors.New("process status is still running")
		}

		if processInfo.AutoStart {
			std.InstanceData().LogI("process autostart is true, process will be restart by gse again")

			return errors.New("process autostart is true, process will be restart by gse again")
		}

		// update process info
		std.DeployInfo().Process.Info = *processInfo

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to wait process no running: %w", err)
	}

	std.InstanceData().LogI(fmt.Sprintf("process stopped, pid(%d), version(%s), agent-id(%s), autostart(%t), status(%s)",
		processInfo.Pid, processInfo.Version, processInfo.AgentID, processInfo.AutoStart, processInfo.Status))

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actTryStopProcess) DisplayNameZh() string { return act.Name() }

// DisplayNameEn returns the English display name of the action.
func (act *actTryStopProcess) DisplayNameEn() string { return act.Name() }
