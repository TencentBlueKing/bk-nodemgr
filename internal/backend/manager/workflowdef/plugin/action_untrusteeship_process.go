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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameUnTrusteeshipProcess the name of action untrusteeship process.
	ActionNameUnTrusteeshipProcess = "untrusteeship_process"
)

// NewActionUnTrusteeshipProcess new an action to untrusteeship process.
func NewActionUnTrusteeshipProcess(capability *Capability) action.Definition {
	return &actUnTrusteeshipProcess{
		daoPluginDeployment: capability.StoragePlugin,
		gseHandlerProc:      capability.GSEHandler,
	}
}

// ActionParamUnTrusteeshipProcess defines the parameters for actionUnTrusteeshipProcess.
type ActionParamUnTrusteeshipProcess struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// actUnTrusteeshipProcess ...
type actUnTrusteeshipProcess struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	gseHandlerProc      gse.IHandlerProc
}

// Name returns the name of the action.
func (act *actUnTrusteeshipProcess) Name() string {
	return ActionNameUnTrusteeshipProcess
}

// Version returns the version of the action.
func (act *actUnTrusteeshipProcess) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actUnTrusteeshipProcess) Description() string {
	return "untrusteeship process by gse."
}

// Timeout returns the timeout of the action.
func (act *actUnTrusteeshipProcess) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actUnTrusteeshipProcess) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actUnTrusteeshipProcess) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actUnTrusteeshipProcess) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actUnTrusteeshipProcess) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamUnTrusteeshipProcess)
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
	processSpec := std.DeployInfo().Process.ToProcessSpec()

	result, err := act.gseHandlerProc.UnTrusteeshipProcess(nCtx, processSpec)
	if err != nil {
		return fmt.Errorf("failed to untrusteeship process: %w", err)
	}

	std.InstanceData().LogI(fmt.Sprintf("successfully execute untrusteeship process operate, result(%s)", result))

	processInfo, err := act.gseHandlerProc.QueryProcessInfo(nCtx, processSpec.PluginName, processSpec.Identity.Name, processSpec.AgentID)
	if err != nil {
		return fmt.Errorf("failed to query process info: %w", err)
	}

	if processInfo.Trusteeship {
		std.InstanceData().LogI("process is still trusteeship by gse")

		return fmt.Errorf("process is still trusteeship by gse")
	}

	std.InstanceData().LogI(fmt.Sprintf("process running, pid(%d), version(%s), agent-id(%s), trusteeship(%t), status(%s)",
		processInfo.Pid, processInfo.Version, processInfo.AgentID, processInfo.Trusteeship, processInfo.Status))

	return nil
}
