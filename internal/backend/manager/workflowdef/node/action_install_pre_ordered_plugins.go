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
// nolint: perfsprint,funlen,gocognit,cyclop,gocyclo
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

	gp := gopool.NewPool()
	subWorkflowRefChan := make(chan types.SubWorkflowRef, 2) // nolint: mnd
	gp.Go(func() error {
		execute, workflowID, err := act.installPreOrderedPlugin(std)
		if err != nil {
			return err
		}

		if execute {
			subWorkflowRefChan <- types.SubWorkflowRef{
				WorkflowID:     workflowID,
				WorkflowDomain: types.WorkflowDomainPlugin,
			}
		}

		return nil
	})

	gp.Go(func() error {
		execute, workflowID, err := act.installPreOrderedPluginV2(std)
		if err != nil {
			return err
		}

		if execute {
			subWorkflowRefChan <- types.SubWorkflowRef{
				WorkflowID:     workflowID,
				WorkflowDomain: types.WorkflowDomainPlugin,
			}
		}

		return nil
	})

	if err = gp.Wait(); err != nil {
		std.InstanceData().Log().
			Zh("安装预设插件失败, 错误(%v)", err).
			En("install pre-ordered plugins failed, error(%v)", err).
			Error()
	}
	close(subWorkflowRefChan)

	subWorkflowRefs := make([]types.SubWorkflowRef, 0)
	for subWorkflowRef := range subWorkflowRefChan {
		subWorkflowRefs = append(subWorkflowRefs, subWorkflowRef)
	}

	if err := act.waitWorkflow(std, subWorkflowRefs); err != nil {
		return err
	}

	return nil
}

func (act *actionInstallPreOrderedPlugins) installPreOrderedPlugin(std *nodeUtils.NodeActionStandarder) (bool, string, error) {
	nCtx := std.Context()
	deployInfo := std.DeployInfo()

	// Skip this action if InstallPreOrderedPlugins is disabled.
	if !deployInfo.InstallOptions.InstallPreOrderedPlugins {
		std.InstanceData().Log().
			Zh("未开启安装预设插件, 跳过此操作").
			En("install pre-ordered plugins is disabled, skip this action").
			Info()

		return false, "", nil
	}

	preOrderedPlugins := getPreOrderedPlugins()
	if len(preOrderedPlugins) == 0 {
		std.InstanceData().Log().
			Zh("无预置插件需要安装, 跳过此操作").
			En("no pre-ordered plugins to install, skip this action").
			Info()

		return false, "", nil
	}

	preOrderedPluginsName := preOrderedPlugins[deployInfo.Host.Dynamic.NodeRole]
	std.InstanceData().Log().
		Zh("开始安装预置插件(%v)", preOrderedPluginsName).
		En("start to install pre-ordered plugins(%v)", preOrderedPluginsName).
		Info()

	if std.DeployInfo().InstallOptions.IsOffline {
		std.InstanceData().Log().
			Zh("当前为离线安装模式").
			En("offline install mode").
			Info()
	}

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
				BizID:      deployInfo.Host.Static.BizID,
				PluginName: name,
				Version:    version,
				IsOffline:  deployInfo.InstallOptions.IsOffline,
			})

			return nil
		})
	}
	if err := gp.Wait(); err != nil {
		return false, "", err
	}

	// if in offline mode or indirect unit or proxy install, disable select downloads.
	pluginTransferOpts := types.DefaultPluginDeploymentTransferOptions()
	if deployInfo.InstallOptions.IsOffline || !deployInfo.InstallOptions.DirectInstall || deployInfo.Host.Dynamic.NodeRole == types.NodeRoleProxy {
		pluginTransferOpts.SelectDownloads = false
	}

	// create plugin deployments.
	pluginDeployments, hostIDs, bizIDs, err := types.NewPluginDeploymentsByParams(nCtx.TenantID(), pluginTransferOpts, deployParams...)
	if err != nil {
		return false, "", fmt.Errorf("failed to create plugin deployments by params: %w", err)
	}

	workflowID, err := act.pluginMgrIface.LaunchInstallPlugin(nCtx, types.InstallPluginParam{
		Type:              types.PluginWorkflowTypeInstall,
		HostIDs:           hostIDs,
		BizIDs:            bizIDs,
		PluginDeployments: pluginDeployments,
		Operator:          nCtx.BKUsername(),
	})
	if err != nil {
		return false, "", fmt.Errorf("failed to launch install pre-ordered plugins workflow: %w", err)
	}

	return true, workflowID, nil
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

func getPreOrderedPlugins() map[types.NodeRole][]string {
	// TODO: 接入配置管理
	return map[types.NodeRole][]string{
		types.NodeRoleAgent: {"bkmonitorbeat"},
		types.NodeRoleProxy: {"bkmonitorbeat", "bk-nodemgr-relay"},
	}
}

func (act *actionInstallPreOrderedPlugins) installPreOrderedPluginV2(std *nodeUtils.NodeActionStandarder) (bool, string, error) {
	nCtx := std.Context()
	deployInfo := std.DeployInfo()

	//// Skip this action if EnableCompatibilityMode is disabled.
	// v2 plugin is special, it not controlled by InstallPreOrderedPlugins.
	if !deployInfo.InstallOptions.EnableCompatibilityMode || deployInfo.InstallOptions.InstallPreOrderedPlugins {
		std.InstanceData().Log().
			Zh("未开启安装预设 V2 插件, 跳过此操作").
			En("install pre-ordered v2 plugins is disabled, skip this action").
			Info()

		return false, "", nil
	}

	preOrderedPluginV2s := getPreOrderedPluginV2s(deployInfo.Host.Dynamic.NodeRole)
	if len(preOrderedPluginV2s) == 0 {
		std.InstanceData().Log().
			Zh("无预置 V2 插件需要安装, 跳过此操作").
			En("no pre-ordered v2 plugins to install, skip this action").
			Info()

		return false, "", nil
	}

	std.InstanceData().Log().
		Zh("开始安装预置插件(%v)", preOrderedPluginV2s).
		En("start to install pre-ordered v2 plugins(%v)", preOrderedPluginV2s).
		Info()

	if std.DeployInfo().InstallOptions.IsOffline {
		std.InstanceData().Log().
			Zh("当前为离线安装模式").
			En("offline install mode").
			Info()
	}

	deployParams := make([]*types.PluginDeploymentParam, 0, len(preOrderedPluginV2s))
	gp := gopool.NewPool()
	for _, pluginName := range preOrderedPluginV2s {
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
				BizID:      deployInfo.Host.Static.BizID,
				PluginName: name,
				Version:    version,
				IsOffline:  deployInfo.InstallOptions.IsOffline,
			})

			return nil
		})
	}
	if err := gp.Wait(); err != nil {
		return false, "", err
	}

	// if in offline mode or indirect unit or proxy install, disable select downloads.
	pluginTransferOpts := types.DefaultPluginDeploymentTransferOptions()
	if deployInfo.InstallOptions.IsOffline || !deployInfo.InstallOptions.DirectInstall || deployInfo.Host.Dynamic.NodeRole == types.NodeRoleProxy {
		pluginTransferOpts.SelectDownloads = false
	}

	// create plugin deployments.
	pluginDeployments, hostIDs, bizIDs, err := types.NewPluginDeploymentsByParams(nCtx.TenantID(),
		pluginTransferOpts, deployParams...)
	if err != nil {
		return false, "", fmt.Errorf("failed to create plugin deployments by params: %w", err)
	}

	workflowID, err := act.pluginMgrIface.LaunchPluginEnsurePluginV2(nCtx, types.InstallPluginParam{
		Type:              types.PluginWorkflowTypePluginEnsureV2,
		HostIDs:           hostIDs,
		BizIDs:            bizIDs,
		PluginDeployments: pluginDeployments,
		Operator:          nCtx.BKUsername(),
	})
	if err != nil {
		return false, "", fmt.Errorf("failed to launch ensure plugin v2 workflow for pre-ordered plugins: %w", err)
	}

	return true, workflowID, nil
}

func getPreOrderedPluginV2s(nodeRole types.NodeRole) []string {
	// TODO: 接入配置管理
	m := map[types.NodeRole][]string{
		types.NodeRoleAgent: {"bkmonitorbeat"},
		types.NodeRoleProxy: {"bkmonitorbeat"},
	}

	preOrderedPluginV2s, ok := m[nodeRole]
	if !ok {
		return []string{}
	}

	return preOrderedPluginV2s
}

func (act *actionInstallPreOrderedPlugins) waitWorkflow(std *nodeUtils.NodeActionStandarder,
	subWorkflowRefs []types.SubWorkflowRef) error {

	nCtx := std.Context()
	serializedSubWorkflowRefs, err := types.SerializeSubWorkflowRefs(subWorkflowRefs)
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
		Zh("成功启动预置插件安装工作流, workflow(%+v)", subWorkflowRefs).
		En("succeed to launch install pre-ordered plugins workflow, workflow(%+v)", subWorkflowRefs).
		Info()

	std.InstanceData().Log().
		Zh("等待工作流完成").
		En("wait workflow finish").
		Info()

	gp := gopool.NewPool()
	for idx := range subWorkflowRefs {
		subWorkflowRef := subWorkflowRefs[idx]
		gp.Go(func() error {
			return act.checkPluginWorkflowStatus(std, subWorkflowRef)
		})
	}
	if err := gp.Wait(); err != nil {
		return fmt.Errorf("failed to wait workflow: %w", err)
	}

	return nil
}

func (act *actionInstallPreOrderedPlugins) checkPluginWorkflowStatus(
	std *nodeUtils.NodeActionStandarder, subWorkflowRef types.SubWorkflowRef) error {

	workflowID := subWorkflowRef.WorkflowID

	polling := retrier.NewPolling(retrier.PollingOpts{
		Timeout:  act.Timeout(),
		Interval: pollingInterval,
	})

	nCtx := std.Context()

	var workflowStatus types.PluginWorkflowStatus
	err := polling.Do(nCtx, func(_ int) error {
		var err error
		workflowStatus, err = act.storagePluginWorkflow.GetPluginWorkflowStatus(nCtx, workflowID)
		if err != nil {
			return fmt.Errorf("failed to get plugin workflow status: %w", err)
		}

		if workflowStatus == types.PluginWorkflowStatusRunning {
			std.InstanceData().Log().
				Zh("预置插件安装工作流仍在运行中, workflow-id: %s", workflowID).
				En("install pre-ordered plugins workflow is still running, workflow-id: %s", workflowID).
				Info()

			return fmt.Errorf("plugin workflow is still running, workflow-id(%s)", workflowID)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("install pre-ordered plugins workflow polling failed: %w", err)
	}

	std.InstanceData().Log().
		Zh("预置插件安装工作流已结束, workflow-id: %s, 状态: %s", workflowID, workflowStatus).
		En("install pre-ordered plugins workflow finished, workflow-id: %s, status: %s", workflowID, workflowStatus).
		Info()

	switch workflowStatus {
	case types.PluginWorkflowStatusSuccess:
		std.InstanceData().Log().
			Zh("预置插件安装成功, workflow-id: %s", workflowID).
			En("install pre-ordered plugins succeeded, workflow-id: %s", workflowID).
			Info()

		return nil
	case types.PluginWorkflowStatusFailed:
		std.InstanceData().Log().
			Zh("预置插件安装失败, workflow-id: %s", workflowID).
			En("install pre-ordered plugins failed, workflow-id: %s", workflowID).
			Info()

		return errors.New("install pre-ordered plugins failed")
	case types.PluginWorkflowStatusPartialFailed:
		std.InstanceData().Log().
			Zh("预置插件安装部分失败, workflow-id: %s", workflowID).
			En("install pre-ordered plugins partially failed, workflow-id: %s", workflowID).
			Info()

		return errors.New("install pre-ordered plugins partially failed")
	default:
		return fmt.Errorf("unknown plugin workflow status: %s", workflowStatus)
	}
}
