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

	pluginUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin/utils"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameRenderPluginDeployment defines the action name.
	ActionNameRenderPluginDeployment = "render_plugin_deployment"
)

// NewActionRenderPluginDeployment ...
func NewActionRenderPluginDeployment(capability *Capability) action.Definition {
	return &RenderPluginDeployment{
		daoHost:             capability.StorageTopo,
		daoPluginDeployment: capability.StoragePlugin,
	}
}

// ActParamRenderPluginDeployment ...
type ActParamRenderPluginDeployment struct {
	pluginUtils.PluginActionStandardParam `json:",inline"`
}

// RenderPluginDeployment ...
type RenderPluginDeployment struct {
	daoHost             topoStg.IStorageHost
	daoPluginDeployment pluginStg.IDaoPluginDeployment
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
	return 3 // nolint: mnd
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
	host, err := act.daoHost.GetHostByID(nCtx, std.DeployInfo().Plugin.Dynamic.HostID)
	if err != nil {
		return fmt.Errorf("failed to get host by id. host-id(%d): %w", std.DeployInfo().Plugin.Dynamic.HostID, err)
	}

	std.DeployInfo().BlockingActionName = ActionNameRenderPluginDeployment
	std.DeployInfo().Plugin.Dynamic.Generation = host.Dynamic.NodeGeneration
	std.DeployInfo().Plugin.Dynamic.Platform = platfmt.Platform{
		OS:   host.Dynamic.NodeOsType,
		Arch: host.Dynamic.NodeCPUArch,
	}

	return nil
}
