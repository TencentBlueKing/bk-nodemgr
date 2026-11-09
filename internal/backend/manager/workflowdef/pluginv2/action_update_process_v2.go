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
	// ActionNameUpdateProcessV2 the name of action update process.
	ActionNameUpdateProcessV2 = "update_process_v2"
)

// NewActionUpdateProcessV2 new an action to update the v2 plugin process.
func NewActionUpdateProcessV2(capability *Capability) action.Definition {
	return &actionUpdateProcessV2{
		daoPluginDeployment: capability.StoragePlugin,
		daoProcessV2:        capability.StoragePlugin,
	}
}

// ActParamUpdateProcessV2 defines the parameters for actionUpdateProcessV2.
type ActParamUpdateProcessV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

// actionUpdateProcessV2 completes the v2 plugin process record at the end of the install chain,
// when the process identity, controller and runtime info are all rendered and reported.
type actionUpdateProcessV2 struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoProcessV2        pluginStg.IDaoProcessV2
}

// Name returns the name of the action.
func (act *actionUpdateProcessV2) Name() string {
	return ActionNameUpdateProcessV2
}

// Version returns the version of the action.
func (act *actionUpdateProcessV2) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actionUpdateProcessV2) Description() string {
	return "update process v2"
}

// Timeout returns the timeout of the action.
func (act *actionUpdateProcessV2) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionUpdateProcessV2) Tags() []action.Tag {
	return []action.Tag{}
}

// Queue returns the queue of the action.
func (act *actionUpdateProcessV2) Queue() action.Queue {
	return action.QueueDefault
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionUpdateProcessV2) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionUpdateProcessV2) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Do this func define what the action will do.
func (act *actionUpdateProcessV2) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamUpdateProcessV2)
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

	nCtx := std.Context()
	process := std.DeployInfo().Process
	process.TenantID = nCtx.TenantID()

	if err = act.daoProcessV2.UpdateProcessV2(nCtx, &process); err != nil {
		return fmt.Errorf("failed to update process v2, plugin-name(%s), host-id(%d): %w", process.PluginName, process.HostID, err)
	}

	std.InstanceData().Log().
		Zh("成功更新 V2 进程信息, plugin-name(%s), host-id(%d)", process.PluginName, process.HostID).
		En("succeed to update v2 process info, plugin-name(%s), host-id(%d)", process.PluginName, process.HostID).
		Info()

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionUpdateProcessV2) DisplayNameZh() string {
	return "更新 V2 进程"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionUpdateProcessV2) DisplayNameEn() string {
	return "Update V2 Process"
}
