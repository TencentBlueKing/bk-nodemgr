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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameDeleteProcessV2 the name of action delete process.
	ActionNameDeleteProcessV2 = "delete_process_v2"
)

// NewActionDeleteProcessV2 new an action to delete process.
func NewActionDeleteProcessV2(capability *Capability) action.Definition {
	return &actionDeleteProcessV2{
		daoPluginDeployment: capability.StoragePlugin,
		daoProcess:          capability.StoragePlugin,
	}
}

// ActParamDeleteProcessV2 defines the parameters for actionDeleteProcessV2.
type ActParamDeleteProcessV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

type actionDeleteProcessV2 struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoProcess          pluginStg.IDaoProcess
}

// Name returns the name of the action.
func (act *actionDeleteProcessV2) Name() string {
	return ActionNameDeleteProcessV2
}

// Version returns the version of the action.
func (act *actionDeleteProcessV2) Version() string {
	return "1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionDeleteProcessV2) Description() string {
	return "delete process v2"
}

// Timeout returns the timeout of the action.
func (act *actionDeleteProcessV2) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionDeleteProcessV2) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionDeleteProcessV2) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionDeleteProcessV2) DelayFn(_ int) func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *actionDeleteProcessV2) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamDeleteProcessV2)
	err = conv.MapToStruct(ctx.Data.Content, param)
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

	deployInfo := std.DeployInfo()
	if err = act.daoProcess.DeleteProcess(std.Context(), deployInfo.Process.HostID, deployInfo.Process.PluginName); err != nil {
		return fmt.Errorf("failed to delete process, plugin-name(%s), host-id(%d): %w",
			deployInfo.Process.PluginName, deployInfo.Process.HostID, err)
	}

	std.InstanceData().Log().
		Zh("成功删除 V2 进程信息, plugin-name(%s), host-id(%d)",
			deployInfo.Process.PluginName, deployInfo.Process.HostID).
		En("succeed to delete V2 process info, plugin-name(%s), host-id(%d)",
			deployInfo.Process.PluginName, deployInfo.Process.HostID).
		Info()

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionDeleteProcessV2) DisplayNameZh() string {
	return "删除 V2 进程"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionDeleteProcessV2) DisplayNameEn() string {
	return "Delete V2 Process"
}
