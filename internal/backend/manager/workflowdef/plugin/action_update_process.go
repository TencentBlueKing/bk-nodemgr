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
	// ActionNameUpdateProcess defines the action name.
	ActionNameUpdateProcess = "update_process"
)

// NewActionUpdateProcess ...
func NewActionUpdateProcess(capability *Capability) action.Definition {
	return &actUpdateProcess{
		daoPluginDeployment: capability.StoragePlugin,
		daoProcess:          capability.StoragePlugin,
		gseHandlerProc:      capability.GSEHandler,
	}
}

// ActionParamUpdateProcess ...
type ActionParamUpdateProcess struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// actUpdateProcess ...
type actUpdateProcess struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoProcess          pluginStg.IDaoProcess
	gseHandlerProc      gse.IHandlerProc
}

// Name returns the name of the action.
func (act *actUpdateProcess) Name() string {
	return ActionNameUpdateProcess
}

// Version returns the version of the action.
func (act *actUpdateProcess) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actUpdateProcess) Description() string {
	return "trusteeship plugin to gse."
}

// Timeout returns the timeout of the action.
func (act *actUpdateProcess) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actUpdateProcess) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actUpdateProcess) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actUpdateProcess) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actUpdateProcess) Do(ctx *action.InstanceContext) error {
	param := new(ActionParamUpdateProcess)
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
		return fmt.Errorf("failed to update process, plugin-name(%s), host-id(%d): %w",
			std.DeployInfo().Process.PluginName, std.DeployInfo().Process.HostID, err)
	}

	std.InstanceData().Log().
		Zh("成功更新进程信息，plugin-name(%s), host-id(%d)",
			std.DeployInfo().Process.PluginName,
			std.DeployInfo().Process.HostID).
		En("succeed to update process info, plugin-name(%s), host-id(%d)",
			std.DeployInfo().Process.PluginName,
			std.DeployInfo().Process.HostID).
		Info()

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actUpdateProcess) DisplayNameZh() string {
	return "更新进程"
}

// DisplayNameEn returns the English display name of the action.
func (act *actUpdateProcess) DisplayNameEn() string {
	return "Update Process"
}
