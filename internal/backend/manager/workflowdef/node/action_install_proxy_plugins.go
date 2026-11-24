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

package node

import (
	"errors"
	"fmt"
	"time"

	managerIface "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/iface"
	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	pluginStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/plugin"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/file"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

// ActionNameInstallProxyPlugins defines the action name for installing Proxy-related plugins.
const ActionNameInstallProxyPlugins = "install_proxy_plugins"

// NewActionInstallProxyPlugins creates an action that installs only Proxy-related plugins.
func NewActionInstallProxyPlugins(capability *Capability) action.Definition {
	return &actionInstallProxyPlugins{
		fileHandler:           capability.FileHandler,
		storageNodeDeployment: capability.StorageNode,
		storageHost:           capability.StorageTopo,
		storagePluginWorkflow: capability.StoragePlugin,
		storageActionInstance: capability.StorageWorkflow,
		pluginMgrIface:        capability.PluginIface,
	}
}

type actionInstallProxyPlugins struct {
	fileHandler           file.IReleasePluginHandler
	storageNodeDeployment nodeStg.IDaoNodeDeployment
	storageHost           topoStg.IStorageHost
	storagePluginWorkflow pluginStg.IDaoPluginWorkflow
	storageActionInstance workflow.IStorageActionInstance
	pluginMgrIface        managerIface.IPluginManager
}

// Name returns the name of the action.
func (act *actionInstallProxyPlugins) Name() string {
	return ActionNameInstallProxyPlugins
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionInstallProxyPlugins) DisplayNameZh() string {
	return "安装 Proxy 相关插件"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionInstallProxyPlugins) DisplayNameEn() string {
	return "Install Proxy Plugins"
}

// Version returns the version of the action.
func (act *actionInstallProxyPlugins) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionInstallProxyPlugins) Description() string {
	return "Install Proxy-related plugins on target node."
}

// Timeout returns the timeout of the action.
func (act *actionInstallProxyPlugins) Timeout() time.Duration {
	return 10 * time.Minute // nolint: mnd
}

// Tags returns the tags of the action.
func (act *actionInstallProxyPlugins) Tags() []action.Tag {
	return []action.Tag{}
}

// Queue returns the queue of the action.
func (act *actionInstallProxyPlugins) Queue() action.Queue {
	return action.QueueDefault
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionInstallProxyPlugins) MaxRetryCount() uint {
	return 3 // nolint: mnd
}

// DelayFn returns the retry delay function.
func (act *actionInstallProxyPlugins) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Do installs the relay plugin without installing pre-ordered monitoring plugins.
func (act *actionInstallProxyPlugins) Do(ctx *action.InstanceContext) (err error) {
	param := new(nodeUtils.NodeActionStandardParam)
	if err := conv.MapToStruct(ctx.Data.Content, param); err != nil {
		return fmt.Errorf("failed to decode Proxy plugin install parameters: %w", err)
	}

	std := nodeUtils.NewNodeActionStandarder(act.storageNodeDeployment, act.storageHost)
	if err := std.Initialize(ctx, *param); err != nil {
		return fmt.Errorf("failed to initialize Proxy plugin installation: %w", err)
	}
	defer func() {
		if storeErr := std.Save(); storeErr != nil {
			err = errors.Join(storeErr, err)
		}
	}()

	workflowID, err := act.installProxyPlugins(std)
	if err != nil {
		return fmt.Errorf("failed to install Proxy plugins: %w", err)
	}

	if err := act.waitWorkflow(std, workflowID); err != nil {
		return fmt.Errorf("failed to wait for Proxy plugin installation: %w", err)
	}

	return nil
}

func (act *actionInstallProxyPlugins) installProxyPlugins(std *nodeUtils.NodeActionStandarder) (string, error) {
	nCtx := std.Context()
	deployInfo := std.DeployInfo()
	const pluginName = "bk-nodemgr-relay"

	std.InstanceData().Log().
		Zh("开始安装 Proxy 相关插件(%s)", pluginName).
		En("start to install Proxy plugins(%s)", pluginName).
		Info()

	version, err := act.fileHandler.GetReleasePluginDefaultVersion(
		nCtx,
		pluginName,
		deployInfo.Host.Dynamic.NodeGeneration,
		platform.NewPlatform(deployInfo.Host.Dynamic.NodeOsType, deployInfo.Host.Dynamic.NodeCPUArch),
	)
	if err != nil {
		return "", fmt.Errorf("failed to get default version for plugin, plugin-name(%s): %w", pluginName, err)
	}

	deployParam := &types.PluginDeploymentParam{
		HostID:                  deployInfo.Host.HostID,
		BizID:                   deployInfo.Host.Static.BizID,
		PluginName:              pluginName,
		Version:                 version,
		IsOffline:               deployInfo.InstallOptions.IsOffline,
		EnableCompatibilityMode: deployInfo.InstallOptions.EnableCompatibilityMode,
		RenewSubConfig:          true,
	}
	// Proxy plugins require all transfer endpoints rather than selective downloads.
	pluginDeployments, hostIDs, bizIDs, err := types.NewPluginDeploymentsByParams(
		nCtx.TenantID(), types.PluginDeploymentTransferOptionsAll(), deployParam)
	if err != nil {
		return "", fmt.Errorf("failed to create plugin deployments by params: %w", err)
	}

	workflowID, err := act.pluginMgrIface.LaunchInstallPlugin(nCtx, types.InstallPluginParam{
		Type:              types.PluginWorkflowTypeInstall,
		HostIDs:           hostIDs,
		BizIDs:            bizIDs,
		PluginDeployments: pluginDeployments,
		Operator:          nCtx.BKUsername(),
	})
	if err != nil {
		return "", fmt.Errorf("failed to launch install Proxy plugins workflow: %w", err)
	}

	return workflowID, nil
}

func (act *actionInstallProxyPlugins) waitWorkflow(std *nodeUtils.NodeActionStandarder, workflowID string) error {
	subWorkflowRefs := []types.SubWorkflowRef{{
		WorkflowID:     workflowID,
		WorkflowDomain: types.WorkflowDomainPlugin,
	}}
	serializedRefs, err := types.SerializeSubWorkflowRefs(subWorkflowRefs)
	if err != nil {
		return fmt.Errorf("failed to serialize sub workflow refs: %w", err)
	}

	std.InstanceData().PrivateData[types.PDKeySubWorkflowRefs] = serializedRefs
	if err := act.storageActionInstance.UpsertActionInstancePrivateData(
		std.Context(),
		std.InstanceData().OperationInstanceID,
		ActionNameInstallProxyPlugins,
		map[string]any{types.PDKeySubWorkflowRefs: serializedRefs},
	); err != nil {
		return fmt.Errorf("failed to save sub workflow refs to private data: %w", err)
	}

	std.InstanceData().Log().
		Zh("成功启动 Proxy 插件安装工作流，等待工作流完成, workflow-id(%s)", workflowID).
		En("launched Proxy plugin installation workflow, waiting for completion, workflow-id(%s)", workflowID).
		Info()

	return act.checkPluginWorkflowStatus(std, workflowID)
}

func (act *actionInstallProxyPlugins) checkPluginWorkflowStatus(std *nodeUtils.NodeActionStandarder, workflowID string) error {
	polling := retrier.NewPolling(retrier.PollingOpts{
		Timeout:  act.Timeout(),
		Interval: 10 * time.Second, // nolint: mnd
	})
	nCtx := std.Context()
	var workflowStatus types.PluginWorkflowStatus
	if err := polling.Do(nCtx, func(_ int) error {
		var err error
		workflowStatus, err = act.storagePluginWorkflow.GetPluginWorkflowStatus(nCtx, workflowID)
		if err != nil {
			return fmt.Errorf("failed to get plugin workflow status: %w", err)
		}

		if workflowStatus == types.PluginWorkflowStatusRunning {
			std.InstanceData().Log().
				Zh("Proxy 插件安装工作流仍在运行中, workflow-id(%s)", workflowID).
				En("Proxy plugin installation workflow is still running, workflow-id(%s)", workflowID).
				Info()

			return fmt.Errorf("plugin workflow is still running, workflow-id(%s)", workflowID)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("Proxy plugin installation workflow polling failed: %w", err)
	}

	std.InstanceData().Log().
		Zh("Proxy 插件安装工作流已结束, workflow-id(%s), 状态(%s)", workflowID, workflowStatus).
		En("Proxy plugin installation workflow finished, workflow-id(%s), status(%s)", workflowID, workflowStatus).
		Info()

	switch workflowStatus {
	case types.PluginWorkflowStatusSuccess:
		return nil
	case types.PluginWorkflowStatusFailed:
		return errors.New("install Proxy plugins failed")
	case types.PluginWorkflowStatusPartialFailed:
		return errors.New("install Proxy plugins partially failed")
	default:
		return fmt.Errorf("unknown plugin workflow status: %s", workflowStatus)
	}
}
