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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	managerIface "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/iface"
	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	releaseStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameInstallPreOrderedPlugins defines the action name.
	ActionNameInstallPreOrderedPlugins = "install_pre_ordered_plugins"

	pollingInterval = 10 * time.Second
)

// NewActionInstallPreOrderedPlugins get a new action.
func NewActionInstallPreOrderedPlugins(capability *Capability) action.Definition {
	return &actionInstallPreOrderedPlugins{
		storagePkg:            capability.StorageRelease,
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		storagePluginWorkflow: capability.StoragePlugin,
		storageActionInstance: capability.StorageWorkflow,

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
	storageHost           topoStg.IStorageHost
	storagePluginWorkflow pluginStg.IDaoPluginWorkflow
	storageActionInstance workflow.IStorageActionInstance

	pluginMgrIface managerIface.IPluginManager
}

// Name returns the name of the action.
func (act *actionInstallPreOrderedPlugins) Name() string {
	return ActionNameInstallPreOrderedPlugins
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionInstallPreOrderedPlugins) DisplayNameZh() string {
	return "安装预置插件"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionInstallPreOrderedPlugins) DisplayNameEn() string {
	return "Install Pre-ordered Plugins"
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
	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment, act.storageHost)
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
		std.InstanceData().Log().
			Zh("无预置插件需要安装, 跳过此操作").
			En("no pre-ordered plugins to install, skip this action").
			Info()

		return nil
	}

	preOrderedPluginsName := preOrderedPlugins[deployInfo.Host.Dynamic.NodeRole]
	std.InstanceData().Log().
		Zh("开始安装预置插件(%v)", preOrderedPluginsName).
		En("start to install pre-ordered plugins(%v)", preOrderedPluginsName).
		Info()

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

	subWorkflowRefs := []types.SubWorkflowRef{{
		WorkflowID:     workflowID,
		WorkflowDomain: types.WorkflowDomainPlugin,
	}}
	serializedSubWorkflowRefs, err := serializeSubWorkflowRefs(subWorkflowRefs)
	if err != nil {
		return fmt.Errorf("failed to serialize sub workflow refs: %w", err)
	}

	std.InstanceData().PrivateData[types.PDKeySubWorkflowRefs] = serializedSubWorkflowRefs
	if err = act.saveSubWorkflowRefs(
		nCtx,
		std.InstanceData().OperationInstanceID,
		serializedSubWorkflowRefs,
	); err != nil {
		return fmt.Errorf("failed to save sub workflow refs to private data: %w", err)
	}

	std.InstanceData().Log().
		Zh("成功启动预置插件安装工作流, workflow-id(%s)", workflowID).
		En("succeed to launch install pre-ordered plugins workflow, workflow-id(%s)", workflowID).
		Info()

	std.InstanceData().Log().
		Zh("等待工作流完成").
		En("wait workflow finish").
		Info()
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
			std.InstanceData().Log().
				Zh("预置插件安装工作流仍在运行中...").
				En("install pre-ordered plugins workflow is still running...").
				Info()

			return fmt.Errorf("plugin workflow is still running, workflow-id(%s)", workflowID)
		}

		std.InstanceData().Log().
			Zh("预置插件安装工作流已完成, 状态: %s", workflowStatus).
			En("install pre-ordered plugins workflow finished with status: %s", workflowStatus).
			Info()

		return nil
	})
	if err != nil {
		return fmt.Errorf("install pre-ordered plugins workflow polling failed: %w", err)
	}

	std.InstanceData().Log().
		Zh("预置插件安装工作流已完成").
		En("install pre-ordered plugins workflow finished").
		Info()
	switch workflowStatus {
	case types.PluginWorkflowStatusSuccess:
		std.InstanceData().Log().
			Zh("预置插件安装成功").
			En("install pre-ordered plugins succeeded").
			Info()

		return nil
	case types.PluginWorkflowStatusFailed:
		std.InstanceData().Log().
			Zh("预置插件安装失败").
			En("install pre-ordered plugins failed").
			Info()

		return errors.New("install pre-ordered plugins failed")
	case types.PluginWorkflowStatusPartialFailed:
		std.InstanceData().Log().
			Zh("预置插件安装部分失败").
			En("install pre-ordered plugins partially failed").
			Info()

		return errors.New("install pre-ordered plugins partially failed")
	default:
		return fmt.Errorf("unknown plugin workflow status: %s", workflowStatus)
	}
}

func (act *actionInstallPreOrderedPlugins) saveSubWorkflowRefs(
	nCtx contextx.IContext,
	operInstID string,
	serializedRefs string,
) error {

	return act.storageActionInstance.UpsertActionInstancePrivateData(
		nCtx,
		operInstID,
		ActionNameInstallPreOrderedPlugins,
		map[string]any{
			types.PDKeySubWorkflowRefs: serializedRefs,
		},
	)
}

func serializeSubWorkflowRefs(refs []types.SubWorkflowRef) (string, error) {
	data, err := json.Marshal(refs)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func getPreOrderedPlugins() map[types.NodeRole][]string {
	// TODO: 接入配置管理
	return map[types.NodeRole][]string{
		types.NodeRoleAgent: {"bkmonitorbeat"},
		types.NodeRoleProxy: {"bkmonitorbeat", "bk-nodemgr-relay"},
	}
}
