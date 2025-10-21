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
	// ActionNameUpdatePluginProcess defines the action name.
	ActionNameUpdatePluginProcess = "trusteeship_plugin"
)

// NewActionUpdatePluginProcess ...
func NewActionUpdatePluginProcess(capability *Capability) action.Definition {
	return &actUpdatePluginProcess{
		daoPluginDeployment: capability.StoragePlugin,
		daoProcess:          capability.StoragePlugin,
		gseHandlerProc:      capability.GSEHandler,
	}
}

// ActionParamUpdatePluginProcess ...
type ActionParamUpdatePluginProcess struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// actUpdatePluginProcess ...
type actUpdatePluginProcess struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoProcess          pluginStg.IDaoProcess
	gseHandlerProc      gse.IHandlerProc
}

// Name returns the name of the action.
func (act *actUpdatePluginProcess) Name() string {
	return ActionNameUpdatePluginProcess
}

// Version returns the version of the action.
func (act *actUpdatePluginProcess) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actUpdatePluginProcess) Description() string {
	return "trusteeship plugin to gse."
}

// Timeout returns the timeout of the action.
func (act *actUpdatePluginProcess) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actUpdatePluginProcess) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actUpdatePluginProcess) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actUpdatePluginProcess) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actUpdatePluginProcess) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamUpdatePluginProcess)
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
	err = act.daoProcess.UpdateProcess(nCtx, std.DeployInfo().Process.HostID, std.DeployInfo().Process.PluginName, &std.DeployInfo().Process)
	if err != nil {
		return fmt.Errorf("update process err:%v", err)
	}

	return nil
}
