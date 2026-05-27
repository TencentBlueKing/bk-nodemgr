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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameFinishIfPluginProcessV2Alive finish if plugin process v2 alive.
	ActionNameFinishIfPluginProcessV2Alive = "finish_if_plugin_process_running_v2"
)

// NewActionFinishIfPluginProcessV2Alive finish if plugin process v2 alive.
func NewActionFinishIfPluginProcessV2Alive(capability *Capability) action.Definition {
	return &actionFinishIfPluginProcessV2Alive{
		daoPluginDeployment:  capability.StoragePlugin,
		daoOperationInstance: capability.StorageWorkflow,
		daoActionInstance:    capability.StorageWorkflow,
	}
}

// ActParamFinishIfPluginProcessV2Alive defines the parameters for actionFinishIfPluginProcessV2Alive.
type ActParamFinishIfPluginProcessV2Alive struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

type actionFinishIfPluginProcessV2Alive struct {
	daoPluginDeployment  pluginStg.IDaoPluginDeployment
	daoOperationInstance workflow.IStorageOperationInstance
	daoActionInstance    workflow.IStorageActionInstance
}

// Name returns the name of the action.
func (act *actionFinishIfPluginProcessV2Alive) Name() string {
	return ActionNameFinishIfPluginProcessV2Alive
}

// Version returns the version of the action.
func (act *actionFinishIfPluginProcessV2Alive) Version() string {
	return "1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionFinishIfPluginProcessV2Alive) Description() string {
	return "finish if plugin process v2 alive"
}

// Timeout returns the timeout of the action.
func (act *actionFinishIfPluginProcessV2Alive) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionFinishIfPluginProcessV2Alive) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionFinishIfPluginProcessV2Alive) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionFinishIfPluginProcessV2Alive) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionFinishIfPluginProcessV2Alive) Do(ctx *action.InstanceContext) error {
	param := new(ActParamFinishIfPluginProcessV2Alive)
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
	if deployInfo.Process.Info.Status == types.ProcessStatusRunning {
		std.InstanceData().Log().
			Zh("V2 插件进程已正常运行, 跳过后续流程, 主机id(%d), 插件名(%s), 进程状态(%s)",
				deployInfo.Process.HostID, deployInfo.Process.PluginName, deployInfo.Process.Info.Status).
			En("v2 plugin process is already running, host-id(%d), plugin-name(%s), process-status(%s)",
				deployInfo.Process.HostID, deployInfo.Process.PluginName, deployInfo.Process.Info.Status).
			Info()

		std.InstanceData().Content = conv.StructToMapIgnoreError(param)

		operInstanceID := std.InstanceData().OperationInstanceID
		operationInstance, err := act.daoOperationInstance.GetOperationInstanceFullData(nCtx, operInstanceID)
		if err != nil {
			std.InstanceData().Log().
				Zh("获取 Operation Instance 失败, operation-instance-id(%s), err(%v)", operInstanceID, err).
				En("failed to get operation instance, operation-instance-id(%s), err(%v)", operInstanceID, err).
				Error()

			return fmt.Errorf("get operation instance, operation-instance-id(%s), err(%v)", operInstanceID, err)
		}

		for actionName, actionInstData := range operationInstance.ActionInstanceDataMap {
			if actionInstData.Index <= std.InstanceData().Index {
				continue
			}

			// mark action as skipped
			actionInstData.Lifecycle.State = action.StateSkipped

			err := act.daoActionInstance.UpdateOperInstActionStatus(nCtx, operInstanceID, actionName, action.StateSkipped)
			if err != nil {
				std.InstanceData().Log().
					Zh("更新 Action Instance 失败, operation-instance-id(%s), action-name(%s), err(%v)",
						operInstanceID, actionName, err).
					En("failed to update action instance, operation-instance-id(%s), action-name(%s), err(%v)",
						operInstanceID, actionName, err).
					Error()

				return fmt.Errorf("update action instance, operation-instance-id(%s), action-name(%s), err(%v)",
					operInstanceID, actionName, err)
			}
		}
	}

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionFinishIfPluginProcessV2Alive) DisplayNameZh() string {
	return "检测 V2 插件进程是否已经存活，如果存活则结束"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionFinishIfPluginProcessV2Alive) DisplayNameEn() string {
	return "check if plugin process is alive, if alive then finish"
}
