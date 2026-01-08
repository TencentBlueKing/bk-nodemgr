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

	managerIface "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/iface"
	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	releaseStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameInstallPreOrderedPlugins defines the action name.
	ActionNameInstallPreOrderedPlugins = "install_pre_ordered_plugins"

	privateDataKeyPluginWorkflowID = "plugin_workflow_id"
	pollingInterval                = 10 * time.Second
)

// NewActionInstallPreOrderedPlugins get a new action.
func NewActionInstallPreOrderedPlugins(capability *Capability) action.Definition {
	return &actionInstallPreOrderedPlugins{
		storagePkg:            capability.StorageRelease,
		storageNodeDeployment: capability.StorageNode,
		storagePluginWorkflow: capability.StoragePlugin,

		pluginMgrIface: capability.PluginIface,
	}
}

// ActParamInstallPreOrderedPlugins ...
type ActParamInstallPreOrderedPlugins struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

// InstallPreOrderedPluginsParams this struct defines the parameters for installing agent.
type InstallPreOrderedPluginsParams struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionInstallPreOrderedPlugins struct {
	storagePkg            releaseStg.IPlugin
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storagePluginWorkflow pluginStg.IDaoPluginWorkflow

	pluginMgrIface managerIface.IPluginManager
}

// Name returns the name of the action.
func (act *actionInstallPreOrderedPlugins) Name() string {
	return ActionNameInstallPreOrderedPlugins
}

// Version returns the version of the action.
func (act *actionInstallPreOrderedPlugins) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionInstallPreOrderedPlugins) Description() string {
	return "Install pre-ordered plugins on target node."
}

// Timeout returns the timeout of the action.
func (act *actionInstallPreOrderedPlugins) Timeout() time.Duration {
	return 3 * time.Minute // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionInstallPreOrderedPlugins) Tags() []action.Tag {
	return []action.Tag{}
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionInstallPreOrderedPlugins) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (act *actionInstallPreOrderedPlugins) DelayFn() func() {
	return func() {
		time.Sleep(5 * time.Second) // nolint: mnd
	}
}

// Do this func define what the action will do.
// To ensure readability, this action uses fmt.Sprintf to concatenate characters.
// nolint: perfsprint,funlen
// NOCC: golint/fnsize(func design is not suitable for splitting).
func (act *actionInstallPreOrderedPlugins) Do(ctx *action.InstanceContext) error {
	param := new(InstallPreOrderedPluginsParams)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	// initialize standard data.
	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment)
	if err = std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	nCtx := std.Context()
	tenantID := nCtx.TenantID()
	deployInfo := std.DeployInfo()

	preOrderedPlugins := getPreOrderedPlugins()
	if len(preOrderedPlugins) == 0 {
		std.InstanceData().LogI("no pre-ordered plugins to install, skip this action")
		return nil
	}

	preOrderedPluginsName := preOrderedPlugins[deployInfo.Host.Dynamic.NodeRole]
	std.InstanceData().LogI(fmt.Sprintf("start to install pre-ordered plugins(%v)", preOrderedPluginsName))

	deployParams := make([]*types.PluginDeploymentParam, 0, len(preOrderedPluginsName))
	gp := gopool.NewPool()
	for _, pluginName := range preOrderedPluginsName {
		name := pluginName
		gp.Go(func() error {
			version, err := act.storagePkg.GetReleasePluginDefaultVersion(
				nCtx,
				name,
				deployInfo.Host.Dynamic.NodeGeneration,
				platform.NewPlatform(deployInfo.Host.Dynamic.NodeOsType, deployInfo.Host.Dynamic.NodeCPUArch),
			)
			if err != nil {
				return fmt.Errorf("failed to get default version for plugin, plugin-name(%s): %w", name, err)
			}

			deployParams = append(deployParams, &types.PluginDeploymentParam{
				HostID:     deployInfo.Host.HostID,
				PluginName: name,
				Version:    version,
			})

			return nil
		})
	}
	if err = gp.Wait(); err != nil {
		return err
	}

	pluginDeployments, hostIDs, err := types.NewPluginDeploymentsByParams(tenantID, types.DefaultPluginDeploymentTransferOptions(), deployParams...)
	if err != nil {
		return fmt.Errorf("failed to create plugin deployments by params: %w", err)
	}

	workflowID, err := act.pluginMgrIface.LaunchInstallPlugin(nCtx, types.InstallPluginParam{
		Type:              types.PluginWorkflowTypeInstall,
		HostIDs:           hostIDs,
		PluginDeployments: pluginDeployments,
		Operator:          nCtx.BKUsername(),
	})
	if err != nil {
		return fmt.Errorf("failed to launch install pre-ordered plugins workflow: %w", err)
	}

	std.InstanceData().PrivateData[privateDataKeyPluginWorkflowID] = workflowID

	std.InstanceData().LogI(fmt.Sprintf("succeed to launch install pre-ordered plugins workflow, workflow-id(%s)", workflowID))

	std.InstanceData().LogI("wait workflow finish")
	polling := retrier.NewPolling(retrier.PollingOpts{
		Timeout:  act.Timeout(),
		Interval: pollingInterval,
	})

	var workflowStatus types.PluginWorkflowStatus
	err = polling.Do(nCtx, func(_ int) error {
		workflowStatus, err = act.storagePluginWorkflow.GetPluginWorkflowStatus(nCtx, workflowID)
		if err != nil {
			return fmt.Errorf("failed to get plugin workflow status: %w", err)
		}

		if workflowStatus == types.PluginWorkflowStatusRunning {
			std.InstanceData().LogI("install pre-ordered plugins workflow is still running...")
			return fmt.Errorf("plugin workflow is still running, workflow-id(%s)", workflowID)
		}

		std.InstanceData().LogI(fmt.Sprintf("install pre-ordered plugins workflow finished with status: %s", workflowStatus))

		return nil
	})
	if err != nil {
		return fmt.Errorf("install pre-ordered plugins workflow polling failed: %w", err)
	}

	std.InstanceData().LogI("install pre-ordered plugins workflow finished")
	switch workflowStatus {
	case types.PluginWorkflowStatusSuccess:
		std.InstanceData().LogI("install pre-ordered plugins succeeded")
		return nil
	case types.PluginWorkflowStatusFailed:
		std.InstanceData().LogI("install pre-ordered plugins failed")
		return errors.New("install pre-ordered plugins failed")
	case types.PluginWorkflowStatusPartialFailed:
		std.InstanceData().LogI("install pre-ordered plugins partially failed")
		return errors.New("install pre-ordered plugins partially failed")
	default:
		return fmt.Errorf("unknown plugin workflow status: %s", workflowStatus)
	}
}

func getPreOrderedPlugins() map[types.NodeRole][]string {
	// TODO: 接入配置管理
	return map[types.NodeRole][]string{
		types.NodeRoleAgent: {"bkmonitorbeat"},
		types.NodeRoleProxy: {"bkmonitorbeat", "bk-nodemgr-relay"},
	}
}
