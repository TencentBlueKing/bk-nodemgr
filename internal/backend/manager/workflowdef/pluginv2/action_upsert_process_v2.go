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
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameUpsertProcessV2 the name of action upsert process.
	ActionNameUpsertProcessV2 = "upsert_process_v2"
)

// NewActionUpsertProcessV2 new an action to upsert the v2 plugin process.
func NewActionUpsertProcessV2(capability *Capability) action.Definition {
	return &actionUpsertProcessV2{
		daoPluginDeployment: capability.StoragePlugin,
		daoProcessV2:        capability.StoragePlugin,
	}
}

// ActParamUpsertProcessV2 defines the parameters for actionUpsertProcessV2.
type ActParamUpsertProcessV2 struct {
	pluginV2Utils.PluginActionStandardParam `json:",inline"`
}

// actionUpsertProcessV2 creates the v2 plugin process record at the beginning of the install chain,
// and the record is completed by the update process action once the process is started.
type actionUpsertProcessV2 struct {
	daoPluginDeployment pluginStg.IDaoPluginDeployment
	daoProcessV2        pluginStg.IDaoProcessV2
}

// Name returns the name of the action.
func (act *actionUpsertProcessV2) Name() string {
	return ActionNameUpsertProcessV2
}

// Version returns the version of the action.
func (act *actionUpsertProcessV2) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *actionUpsertProcessV2) Description() string {
	return "upsert process v2"
}

// Timeout returns the timeout of the action.
func (act *actionUpsertProcessV2) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionUpsertProcessV2) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionUpsertProcessV2) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionUpsertProcessV2) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Do this func define what the action will do.
func (act *actionUpsertProcessV2) Do(ctx *action.InstanceContext) (err error) {
	param := new(ActParamUpsertProcessV2)
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
	deployProcess := std.DeployInfo().Process
	process := types.Process{
		TenantID:   nCtx.TenantID(),
		HostID:     deployProcess.HostID,
		BizID:      deployProcess.BizID,
		PluginName: deployProcess.PluginName,
		Platform:   platfmt.UnknownPlatform(),
		Info: types.ProcessInfo{
			Version: std.DeployInfo().InstallOptions.Version,
			Status:  types.ProcessStatusInit,
		},
	}

	exist, err := act.daoProcessV2.ExistProcessV2(nCtx, process.HostID, process.PluginName)
	if err != nil {
		return fmt.Errorf("failed to check process v2 exist: %w", err)
	}

	if exist {
		if err = act.daoProcessV2.UpdateProcessInfoV2(nCtx, process.HostID, process.PluginName, &process.Info); err != nil {
			return fmt.Errorf("failed to update process v2 info: %w", err)
		}

		std.InstanceData().Log().
			Zh("指定 V2 进程已存在, plugin-name(%s), host-id(%d)", process.PluginName, process.HostID).
			En("the specified v2 process exist, plugin-name(%s), host-id(%d)", process.PluginName, process.HostID).
			Info()

		return nil
	}

	if err = act.daoProcessV2.CreateProcessV2(nCtx, &process); err != nil {
		return fmt.Errorf("failed to create process v2: %w", err)
	}

	std.InstanceData().Log().
		Zh("成功创建 V2 进程, plugin-name(%s), host-id(%d)", process.PluginName, process.HostID).
		En("successfully create v2 process, plugin-name(%s), host-id(%d)", process.PluginName, process.HostID).
		Info()

	return nil
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionUpsertProcessV2) DisplayNameZh() string {
	return "创建或更新 V2 进程"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionUpsertProcessV2) DisplayNameEn() string {
	return "Upsert V2 Process"
}
