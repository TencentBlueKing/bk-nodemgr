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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameUnTrusteeshipProcessToGseV2 the name of action untrusteeship process.
	ActionNameUnTrusteeshipProcessToGseV2 = "untrusteeship_process_to_gse_v2"
)

// NewActionUnTrusteeshipProcessV2 new an action to untrusteeship process.
func NewActionUnTrusteeshipProcessV2(capability *Capability) action.Definition {
	return &actionUnTrusteeshipProcessV2{
		daoPluginDeployment: capability.StoragePlugin,
		gseHandlerProc:      capability.GSEHandler.NewHandlerProc(gse.WithProcNameSpace(procNameSpaceNodeMan)),
	}
}

// ActParamUnTrusteeshipProcessV2 defines the parameters for actionUnTrusteeshipProcess.
type ActParamUnTrusteeshipProcessV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

// actionUnTrusteeshipProcessV2 ...
type actionUnTrusteeshipProcessV2 struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	gseHandlerProc      gse.IHandlerProc
}

// Name returns the name of the action.
func (act *actionUnTrusteeshipProcessV2) Name() string {
	return ActionNameUnTrusteeshipProcessToGseV2
}

// Version returns the version of the action.
func (act *actionUnTrusteeshipProcessV2) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actionUnTrusteeshipProcessV2) Description() string {
	return "cancel trusteeship process to gse v2."
}

// Timeout returns the timeout of the action.
func (act *actionUnTrusteeshipProcessV2) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionUnTrusteeshipProcessV2) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionUnTrusteeshipProcessV2) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionUnTrusteeshipProcessV2) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionUnTrusteeshipProcessV2) Do(ctx *action.InstanceContext) error {
	param := new(ActParamUnTrusteeshipProcessV2)
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

	nCtx := std.Context()
	processSpec := std.DeployInfo().Process.ToProcessSpec()

	result, err := act.gseHandlerProc.UnTrusteeshipProcess(nCtx, processSpec)
	if err != nil {
		return fmt.Errorf("failed to untrusteeship process: %w", err)
	}

	std.InstanceData().Log().
		Zh("成功执行取消进程托管操作, result(%s)", result).
		En("successfully execute untrusteeship process operation, result(%s)", result).
		Info()

	processInfo, err := act.gseHandlerProc.QueryProcessInfo(nCtx, processSpec.PluginName, processSpec.Identity.Name, processSpec.AgentID)
	if err != nil {
		return fmt.Errorf("failed to query process info: %w", err)
	}

	if processInfo.AutoStart {
		std.InstanceData().Log().
			Zh("进程仍被 GSE 托管").
			En("process is still trusteeship by gse").
			Info()

		return fmt.Errorf("process is still trusteeship by gse")
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
func (act *actionUnTrusteeshipProcessV2) DisplayNameZh() string {
	return "取消 V2 GSE 进程托管"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionUnTrusteeshipProcessV2) DisplayNameEn() string {
	return "Untrusteeship V2 Process to GSE"
}
