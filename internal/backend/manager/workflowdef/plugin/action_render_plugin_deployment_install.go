/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

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

	pluginployment "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	storageTopo "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameRenderPluginDeployment defines the action name.
	ActionNameRenderPluginDeployment = "render_plugin_deployment"
)

// NewActionRenderPluginDeployment ...
func NewActionRenderPluginDeployment(daoHost storageTopo.IStorageHost, daoPluginDeployment pluginployment.IDaoPluginDeployment) action.Definition {
	return &RenderPluginDeployment{
		daoHost:             daoHost,
		daoPluginDeployment: daoPluginDeployment,
	}
}

// ActParamRenderPluginDeployment ...
type ActParamRenderPluginDeployment struct {
	Token    string `json:"token"`
	TenantID string `json:"tenant_id"`
	Operator string `json:"operator"`
}

// RenderPluginDeployment ...
type RenderPluginDeployment struct {
	daoHost             storageTopo.IStorageHost
	daoPluginDeployment pluginployment.IDaoPluginDeployment
}

// Name returns the name of the action.
func (act *RenderPluginDeployment) Name() string {
	return ActionNameRenderPluginDeployment
}

// Version returns the version of the action.
func (act *RenderPluginDeployment) Version() string {
	return "1.0.0"
}

// Description returns the description of the action.
func (act *RenderPluginDeployment) Description() string {
	return "render plugin deployment"
}

// Timeout returns the timeout of the action.
func (act *RenderPluginDeployment) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *RenderPluginDeployment) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *RenderPluginDeployment) MaxRetryCount() uint {
	return 3
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *RenderPluginDeployment) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (act *RenderPluginDeployment) Do(ctx *action.InstanceContext) error {
	param := new(ActParamRenderPluginDeployment)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	nCtx := contextx.From(ctx.Ctx, contextx.WithTenantID(param.TenantID), contextx.WithBKUsername(param.Operator))
	info, err := act.daoPluginDeployment.GetPluginDeploymentInfo(nCtx, param.Token)
	if err != nil {
		return fmt.Errorf("failed to get plugin deployment info: %w", err)
	}

	defer func() {
		if storeErr := act.daoPluginDeployment.UpdatePluginDeploymentInfo(nCtx, param.Token, info); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	host, err := act.daoHost.GetHostByID(nCtx, info.Plugin.HostID)
	if err != nil {
		return fmt.Errorf("failed to get host by id. host-id(%d): %w", info.Plugin.HostID, err)
	}

	info = &types.PluginDeploymentInfo{
		BlockingActionName: ActionNameWaitInstallerComplete,
		Plugin: types.Plugin{
			Name:       info.Plugin.Name,
			HostID:     info.Plugin.HostID,
			Type:       info.Plugin.Type,
			Generation: host.Dynamic.NodeGeneration,
			Platform: platform.Platform{
				OS:   host.Dynamic.NodeOsType,
				Arch: host.Dynamic.NodeCPUArch,
			},
			Version: info.Plugin.Version,
		},
		InstallOptions:  types.PluginDeploymentInstallOptions{},
		TransferOptions: types.PluginDeploymentTransferOptions{},
		TargetVersion:   nil,
	}

	return nil
}
