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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameStartProcessV2 the name of action start process.
	ActionNameStartProcessV2 = "start_process_v2"
)

// NewActionStartProcessV2 new an action to start process.
func NewActionStartProcessV2(capability *Capability) action.Definition {
	return &actionStartProcessV2{
		daoPluginDeployment: capability.StoragePlugin,
		gseHandlerProc:      capability.GSEHandler.NewHandlerProc(gse.WithProcNameSpace(procNameSpaceNodeMan)),
	}
}

// ActParamStartProcessV2 defines the parameters for actionStartProcess.
type ActParamStartProcessV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

// actionStartProcessV2 ...
type actionStartProcessV2 struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	gseHandlerProc      gse.IHandlerProc
}

// Name returns the name of the action.
func (act *actionStartProcessV2) Name() string {
	return ActionNameStartProcessV2
}

// Version returns the version of the action.
func (act *actionStartProcessV2) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actionStartProcessV2) Description() string {
	return "start process by gse v2."
}

// Timeout returns the timeout of the action.
func (act *actionStartProcessV2) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionStartProcessV2) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionStartProcessV2) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionStartProcessV2) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionStartProcessV2) Do(ctx *action.InstanceContext) error {
	param := new(ActParamStartProcessV2)
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
		Zh("尝试执行启动插件进程, plugin-name(%s), host-id(%d), cmd(%s)",
			std.DeployInfo().Process.PluginName, std.DeployInfo().Process.HostID, std.DeployInfo().Process.Controller.StartCmd).
		En("try to executed start plugin process, plugin-name(%s), host-id(%d), cmd(%s)",
			std.DeployInfo().Process.PluginName, std.DeployInfo().Process.HostID, std.DeployInfo().Process.Controller.StartCmd).
		Info()

	nCtx := std.Context()
	processSpec := std.DeployInfo().Process.ToProcessSpec()

	result, err := act.gseHandlerProc.TrusteeshipAndStartProcess(nCtx, processSpec)
	if err != nil {
		return fmt.Errorf("failed to start plugin process: %w", err)
	}

	std.InstanceData().Log().
		Zh("成功执行启动插件进程操作, result(%s)", result).
		En("successfully execute start plugin process operation, result(%s)", result).
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
func (act *actionStartProcessV2) DisplayNameZh() string {
	return "启动 V2 进程"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionStartProcessV2) DisplayNameEn() string {
	return "Start V2 Process"
}
