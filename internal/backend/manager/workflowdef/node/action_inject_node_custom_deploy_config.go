/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package node

import (
	"errors"
	"fmt"
	"time"

	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/deployconstant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameInjectNodeCustomDeployConfig defines the action name.
	ActionNameInjectNodeCustomDeployConfig = "inject_node_custom_deploy_config"
)

// NewActionInjectNodeCustomDeployConfig get a new action.
func NewActionInjectNodeCustomDeployConfig(capability *Capability) action.Definition {
	return &actionInjectNodeCustomDeployConfig{
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		storageDomainGse:      capability.StorageTopo,
	}
}

// ActParamInjectNodeCustomDeployConfig this is the param for render deployment.
type ActParamInjectNodeCustomDeployConfig struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionInjectNodeCustomDeployConfig struct {
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	storageDomainGse      topoStg.IStorageDomainGse
}

// Name returns the name of the action.
func (act *actionInjectNodeCustomDeployConfig) Name() string {
	return ActionNameInjectNodeCustomDeployConfig
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionInjectNodeCustomDeployConfig) DisplayNameZh() string {
	return "注入节点自定义部署配置"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionInjectNodeCustomDeployConfig) DisplayNameEn() string {
	return "Inject Node Custom Deploy Config"
}

// Version returns the version of the action.
func (act *actionInjectNodeCustomDeployConfig) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionInjectNodeCustomDeployConfig) Description() string {
	return "inject node custom deploy config"
}

// Timeout returns the timeout of the action.
func (act *actionInjectNodeCustomDeployConfig) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionInjectNodeCustomDeployConfig) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionInjectNodeCustomDeployConfig) MaxRetryCount() uint {
	return 1 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionInjectNodeCustomDeployConfig) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
// nolint: lll
func (act *actionInjectNodeCustomDeployConfig) Do(ctx *action.InstanceContext) error {
	param := new(ActParamInjectNodeCustomDeployConfig)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment, act.storageHost)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	osType := std.DeployInfo().Host.Dynamic.NodeOsType
	deployConstant, err := deployconstant.GetNodeDeployConf(std.DeployInfo().Host.Dynamic.NodeGeneration, osType)
	if err != nil {
		return fmt.Errorf("failed to get node deploy conf: %w", err)
	}

	customDeployConfig, err := act.storageDomainGse.GetNetworkUnitCustomDeployConfig(
		std.Context(), std.DeployInfo().Host.Dynamic.NetworkUnitID, std.DeployInfo().Host.Dynamic.NodeOsType)
	if err != nil {
		return fmt.Errorf("failed to get network unit custom deploy config: %w", err)
	}

	nodeRole := std.DeployInfo().Host.Dynamic.NodeRole

	// update deploy constant with custom deploy config if custom deploy config exists
	if customDeployConfig != nil {
		// override BaseDeployDir and generate corresponding configuration
		if customDeployConfig.NodeRuntime.BaseDeployDir != "" {
			deployConstant.BaseDeployDir = customDeployConfig.NodeRuntime.BaseDeployDir
		}

		// override BaseWorkDir and generate corresponding configuration
		if customDeployConfig.InstallerRuntime.BaseWorkDir != "" {
			deployConstant.BaseWorkDir = customDeployConfig.InstallerRuntime.BaseWorkDir
		}
	}

	if std.DeployInfo().InstallerRuntime.BaseWorkDir != "" {
		deployConstant.BaseWorkDir = std.DeployInfo().InstallerRuntime.BaseWorkDir
	}

	// installer workdir priority: user specified in info > network unit custom deploy config > deploy constant default.
	std.DeployInfo().InstallerRuntime.BaseWorkDir = deployConstant.BaseWorkDir
	std.DeployInfo().InstallerRuntime.WorkDir = deployConstant.GenerateWorkDir()

	std.DeployInfo().BaseRuntime = types.DeploymentBaseRuntime{
		BaseDeployDir: deployConstant.BaseDeployDir,
		DeployDir:     deployConstant.GenerateDeployDir(),
		HomeDir:       deployConstant.GenerateNodeHomeDir(nodeRole),
		DataIPC:       conv.NonEmptyOr(customDeployConfig.NodeRuntime.DataIPC, deployConstant.GenerateDataIPCPath(nodeRole)),
		PluginIPC:     conv.NonEmptyOr(customDeployConfig.NodeRuntime.PluginIPC, deployConstant.GeneratePluginIPCPath(nodeRole)),
		LogDir:        conv.NonEmptyOr(customDeployConfig.NodeRuntime.LogDir, deployConstant.LogDir),
	}

	std.InstanceData().Log().
		Zh("注入节点自定义部署配置成功, 当前部署配置为(%+v), Installer部署配置为(%+v)", std.DeployInfo().BaseRuntime, std.DeployInfo().InstallerRuntime).
		En("inject node custom deploy config successfully, current deploy config(%+v), Installer deploy config(%+v)", std.DeployInfo().BaseRuntime, std.DeployInfo().InstallerRuntime).
		Info()

	return nil
}
