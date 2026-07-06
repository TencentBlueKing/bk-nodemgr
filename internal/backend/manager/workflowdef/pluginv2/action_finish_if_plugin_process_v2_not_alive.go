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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/gse"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameFinishIfPluginProcessV2NotAlive finish if plugin process v2 not alive.
	ActionNameFinishIfPluginProcessV2NotAlive = "finish_if_plugin_process_not_alive_v2"
)

// NewActionFinishIfPluginProcessV2NotAlive finish if plugin process v2 not alive.
func NewActionFinishIfPluginProcessV2NotAlive(capability *Capability) action.Definition {
	return &actionFinishIfPluginProcessV2NotAlive{
		daoPluginDeployment:  capability.StoragePlugin,
		daoOperationInstance: capability.StorageWorkflow,
		daoActionInstance:    capability.StorageWorkflow,
		daoHost:              capability.StorageTopo,
		gseHandlerProc:       capability.GSEHandler.NewHandlerProc(gse.WithProcNameSpace(procNameSpaceNodeMan)),
	}
}

// ActParamFinishIfPluginProcessV2NotAlive defines the parameters for actionFinishIfPluginProcessV2NotAlive.
type ActParamFinishIfPluginProcessV2NotAlive struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

type actionFinishIfPluginProcessV2NotAlive struct {
	daoPluginDeployment  pluginStg.IDaoPluginDeployment
	daoOperationInstance workflow.IStorageOperationInstance
	daoActionInstance    workflow.IStorageActionInstance
	daoHost              topoStg.IStorageHost
	gseHandlerProc       gse.IHandlerProc
}

// Name returns the name of the action.
func (act *actionFinishIfPluginProcessV2NotAlive) Name() string {
	return ActionNameFinishIfPluginProcessV2NotAlive
}

// Version returns the version of the action.
func (act *actionFinishIfPluginProcessV2NotAlive) Version() string {
	return "1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionFinishIfPluginProcessV2NotAlive) Description() string {
	return "finish if plugin process v2 not alive"
}

// Timeout returns the timeout of the action.
func (act *actionFinishIfPluginProcessV2NotAlive) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionFinishIfPluginProcessV2NotAlive) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionFinishIfPluginProcessV2NotAlive) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionFinishIfPluginProcessV2NotAlive) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionFinishIfPluginProcessV2NotAlive) Do(ctx *action.InstanceContext) error {
	param := new(ActParamFinishIfPluginProcessV2NotAlive)
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
	deployInfo := std.DeployInfo()
	host, err := act.daoHost.GetHostByID(nCtx, deployInfo.Process.HostID)
	if err != nil {
		std.InstanceData().Log().
			Zh("获取主机信息失败, 主机id(%d), 错误(%s)", deployInfo.Process.HostID, err).
			En("fetch host info failed, host-id(%d), error(%s)", deployInfo.Process.HostID, err).
			Error()

		return err
	}

	pluginName := deployInfo.Process.PluginName
	processInfo, err := act.gseHandlerProc.QueryProcessInfo(nCtx, pluginName, pluginName, host.Dynamic.AgentID)
	if err != nil {
		return fmt.Errorf("failed to query process info: %w", err)
	}

	if processInfo.Status != types.ProcessStatusRunning {
		std.InstanceData().Log().
			Zh("V2 插件进程未运行, 跳过后续流程, 主机id(%d), 插件名(%s), 进程状态(%s)", host.HostID, pluginName, processInfo.Status).
			En("v2 plugin process is not running, host-id(%d), plugin-name(%s), process-status(%s)", host.HostID, pluginName, processInfo.Status).
			Info()

		std.InstanceData().Content = conv.StructToMapIgnoreError(param)

		if err := pluginV2Utils.FinishOperationInstance(std, act.daoOperationInstance, act.daoActionInstance); err != nil {
			return fmt.Errorf("failed to finish operation instance: %w", err)
		}
	}

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionFinishIfPluginProcessV2NotAlive) DisplayNameZh() string {
	return "检测 V2 插件进程是否已经存活，如果未存活则结束"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionFinishIfPluginProcessV2NotAlive) DisplayNameEn() string {
	return "check if plugin process is alive, if not alive then finish"
}
