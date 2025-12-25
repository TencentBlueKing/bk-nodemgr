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
	// ActionNameReloadProcess the name of action reload process.
	ActionNameReloadProcess = "reload_process"
)

// NewActionReloadProcess new an action to reload process.
func NewActionReloadProcess(capability *Capability) action.Definition {
	return &actionReloadProcess{
		daoPluginDeployment: capability.StoragePlugin,
		gseHandlerProc:      capability.GSEHandler,
	}
}

// ActParamReloadProcess defines the parameters for actionReloadProcess.
type ActParamReloadProcess struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// actionReloadProcess ...
type actionReloadProcess struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	gseHandlerProc      gse.IHandlerProc
}

// Name returns the name of the action.
func (act *actionReloadProcess) Name() string {
	return ActionNameReloadProcess
}

// Version returns the version of the action.
func (act *actionReloadProcess) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actionReloadProcess) Description() string {
	return "reload process by gse."
}

// Timeout returns the timeout of the action.
func (act *actionReloadProcess) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionReloadProcess) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionReloadProcess) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionReloadProcess) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionReloadProcess) Do(ctx *action.InstanceContext) error {
	param := new(ActParamReloadProcess)
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

	std.InstanceData().LogI(fmt.Sprintf("try to executed reload plugin process, plugin-name(%s), host-id(%d), cmd(%s)",
		std.DeployInfo().Process.PluginName, std.DeployInfo().Process.HostID, std.DeployInfo().Process.Controller.RestartCmd))

	nCtx := std.Context()
	processSpec := std.DeployInfo().Process.ToProcessSpec()
	result, err := act.gseHandlerProc.TrusteeshipAndReloadProcess(nCtx, processSpec)
	if err != nil {
		return fmt.Errorf("failed to reload plugin process: %w", err)
	}

	std.InstanceData().LogI(fmt.Sprintf("successfully execute reload plugin process operation, result(%s)", result))

	std.InstanceData().LogI("wait process running")
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
			std.InstanceData().LogI(fmt.Sprintf("process status is not running, status(%s)", processInfo.Status))

			return fmt.Errorf("process status is not running, status(%s)", processInfo.Status)
		}

		// update process info
		std.DeployInfo().Process.Info = *processInfo

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to wait process running: %w", err)
	}

	std.InstanceData().LogI(fmt.Sprintf("process running, pid(%d), version(%s), agent-id(%s), autostart(%t), status(%s)",
		processInfo.Pid, processInfo.Version, processInfo.AgentID, processInfo.AutoStart, processInfo.Status))

	return nil
}
