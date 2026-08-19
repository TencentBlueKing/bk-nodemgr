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
	// ActionNameTrusteeshipProcessToGseV2 the name of action trusteeship process.
	ActionNameTrusteeshipProcessToGseV2 = "trusteeship_process_to_gse_v2"
)

// NewActionTrusteeshipProcessV2 new an action to trusteeship process.
func NewActionTrusteeshipProcessV2(capability *Capability) action.Definition {
	return &actionTrusteeshipProcessV2{
		daoPluginDeployment: capability.StoragePlugin,
		gseHandlerProc:      capability.GSEHandler.NewHandlerProc(gse.WithProcNameSpace(procNameSpaceNodeMan)),
	}
}

// ActParamTrusteeshipProcessV2 defines the parameters for actionTrusteeshipProcess.
type ActParamTrusteeshipProcessV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

// actionTrusteeshipProcessV2 ...
type actionTrusteeshipProcessV2 struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	gseHandlerProc      gse.IHandlerProc
}

// Name returns the name of the action.
func (act *actionTrusteeshipProcessV2) Name() string {
	return ActionNameTrusteeshipProcessToGseV2
}

// Version returns the version of the action.
func (act *actionTrusteeshipProcessV2) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actionTrusteeshipProcessV2) Description() string {
	return "trusteeship process to gse v2."
}

// Timeout returns the timeout of the action.
func (act *actionTrusteeshipProcessV2) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionTrusteeshipProcessV2) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionTrusteeshipProcessV2) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionTrusteeshipProcessV2) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionTrusteeshipProcessV2) Do(ctx *action.InstanceContext) error {
	param := new(ActParamTrusteeshipProcessV2)
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

	result, err := act.gseHandlerProc.TrusteeshipProcess(nCtx, processSpec)
	if err != nil {
		return fmt.Errorf("failed to trusteeship process: %w", err)
	}

	std.InstanceData().Log().
		Zh("成功执行托管插件进程操作, result(%s)", result).
		En("successfully execute trusteeship plugin process operation, result(%s)", result).
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
	err = polling.Do(nCtx, func(_ int) error {
		processInfo, err = act.gseHandlerProc.QueryProcessInfo(nCtx, processSpec.PluginName, processSpec.Identity.Name, processSpec.AgentID)
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

		if !processInfo.AutoStart {
			std.InstanceData().Log().
				Zh("进程未被 GSE 托管").
				En("process is not trusteeship by gse").
				Info()

			return fmt.Errorf("process is not trusteeship by gse")
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
func (act *actionTrusteeshipProcessV2) DisplayNameZh() string {
	return "托管 V2 进程到 GSE"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionTrusteeshipProcessV2) DisplayNameEn() string {
	return "Trusteeship V2 Process to GSE"
}
