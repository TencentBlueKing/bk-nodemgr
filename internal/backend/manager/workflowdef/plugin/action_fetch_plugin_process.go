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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameFetchPluginProcess the name of action fetch plugin process.
	ActionNameFetchPluginProcess = "fetch_plugin_process"
)

// NewActionFetchPluginProcess new an action to fetch plugin process.
func NewActionFetchPluginProcess(capability *Capability) action.Definition {
	return &actionFetchPluginProcess{
		daoPluginDeployment: capability.StoragePlugin,
		daoProcess:          capability.StoragePlugin,
	}
}

// ActParamFetchPluginProcess defines the parameters for actionFetchPluginProcess.
type ActParamFetchPluginProcess struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

type actionFetchPluginProcess struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoProcess          pluginStg.IDaoProcess
}

// Name returns the name of the action.
func (act *actionFetchPluginProcess) Name() string {
	return ActionNameFetchPluginProcess
}

// Version returns the version of the action.
func (act *actionFetchPluginProcess) Version() string {
	return "1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionFetchPluginProcess) Description() string {
	return "fetch plugin process"
}

// Timeout returns the timeout of the action.
func (act *actionFetchPluginProcess) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionFetchPluginProcess) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionFetchPluginProcess) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionFetchPluginProcess) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionFetchPluginProcess) Do(ctx *action.InstanceContext) error {
	param := new(ActParamFetchPluginProcess)
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
	deployInfo := std.DeployInfo()
	process, err := act.daoProcess.GetProcess(nCtx, deployInfo.Process.HostID, deployInfo.Process.PluginName)
	if err != nil {
		std.InstanceData().LogE(fmt.Sprintf("failed to get process, process-name(%s), host-id(%d): %v",
			deployInfo.Process.PluginName, deployInfo.Process.HostID, err))

		return fmt.Errorf("failed to get plugin process, process-name(%s), host-id(%d): %w",
			deployInfo.Process.PluginName, deployInfo.Process.HostID, err)
	}

	std.InstanceData().LogI(fmt.Sprintf("fetch plugin process succeed, plugin-name(%s), host-id(%d)",
		deployInfo.Process.PluginName, deployInfo.Process.HostID))

	deployInfo.Process = *process

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionFetchPluginProcess) DisplayNameZh() string { return act.Name() }

// DisplayNameEn returns the English display name of the action.
func (act *actionFetchPluginProcess) DisplayNameEn() string { return act.Name() }
