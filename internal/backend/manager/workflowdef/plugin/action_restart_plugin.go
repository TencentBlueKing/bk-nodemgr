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
	// ActionNameRestartPluginProcess the name of action restart plugin process.
	ActionNameRestartPluginProcess = "restart_plugin_process"
)

// NewActionRestartPluginProcess new an action to restart plugin process.
func NewActionRestartPluginProcess(capability *Capability) action.Definition {
	return &actionRestartPluginProcess{
		daoPluginDeployment: capability.StoragePlugin,
		gseHandlerProc:      capability.GSEHandler,
	}
}

// ActParamRestartPluginProcess defines the parameters for actionRestartPluginProcess.
type ActParamRestartPluginProcess struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

type actionRestartPluginProcess struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	gseHandlerProc      gse.IHandlerProc
}

// Name returns the name of the action.
func (act *actionRestartPluginProcess) Name() string {
	return ActionNameRestartPluginProcess
}

// Version returns the version of the action.
func (act *actionRestartPluginProcess) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actionRestartPluginProcess) Description() string {
	return "restart plugin process"
}

// Timeout returns the timeout of the action.
func (act *actionRestartPluginProcess) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionRestartPluginProcess) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionRestartPluginProcess) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionRestartPluginProcess) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionRestartPluginProcess) Do(ctx *action.InstanceContext) error {
	param := new(ActParamRestartPluginProcess)
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

	processSpec := types.ProcessSpec{
		PluginName:    std.DeployInfo().Process.PluginName,
		AgentID:       std.DeployInfo().Process.Info.AgentID,
		Identity:      std.DeployInfo().Process.Identity,
		Controller:    std.DeployInfo().Process.Controller,
		Resource:      std.DeployInfo().Process.Resource,
		MonitorPolicy: std.DeployInfo().Process.MonitorPolicy,
	}

	std.InstanceData().LogI(fmt.Sprintf("try to executed the operation of restart process by cmd(%s)", std.DeployInfo().Process.Controller.RestartCmd))

	cmdOut, err := act.gseHandlerProc.RestartProcess(std.Context(), processSpec)
	if err != nil {
		std.InstanceData().LogE(fmt.Sprintf("failed to restart process, cmdOut(%s), error(%s)", cmdOut, err.Error()))
		return err
	}
	std.InstanceData().LogI(fmt.Sprintf("succeed to executed the operation of restart process, cmd-output(%s)", cmdOut))

	std.InstanceData().LogI("wait process running")
	expoBackoff := retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault())

	var processInfo *types.ProcessInfo
	err = expoBackoff.Do(std.Context(), func(_ int) error {
		processInfo, err = act.gseHandlerProc.QueryProcessInfo(std.Context(), processSpec.PluginName, processSpec.Identity.Name, processSpec.AgentID)
		if err != nil {
			return fmt.Errorf("failed to query process info: %w", err)
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

	std.InstanceData().LogI(fmt.Sprintf("process running, info(%+v)", processInfo))

	return nil
}
