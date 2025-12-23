/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package manager

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/schedule"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata"
)

func (mgr *Manager) registerDefinitions() error {
	if err := mgr.registerDefNode(); err != nil {
		return fmt.Errorf("failed to register def node: %w", err)
	}

	if err := mgr.registerDefSyncData(); err != nil {
		return fmt.Errorf("failed to register def sync data: %w", err)
	}

	if err := mgr.registerDefPlugin(); err != nil {
		return fmt.Errorf("failed to register def plugin: %w", err)
	}

	if err := mgr.registerDefSchedule(); err != nil {
		return fmt.Errorf("failed to register def schedule: %w", err)
	}

	return nil
}

// registerDefNode registers the definitions for node.
// nolint: lll
func (mgr *Manager) registerDefNode() error {
	nodeCap := &node.Capability{
		CMDBHandler:         mgr.conf.CmdbHandler,
		GSEHandler:          mgr.conf.GSEHandler,
		FileHandler:         mgr.conf.FileHandler,
		StorageTopo:         mgr.conf.StorageTopo,
		StorageRelease:      mgr.conf.StorageRelease,
		StorageNode:         mgr.conf.StorageNode,
		StoragePlugin:       mgr.conf.StoragePlugin,
		StorageWorkflow:     mgr.conf.StorageWorkflow,
		StorageHostCredit:   mgr.conf.StorageHostCredit,
		StorageConfigPolicy: mgr.conf.StorageConfigPolicy,
		DiscoverProvider:    mgr.conf.Provider,
		HostPasswordVault:   mgr.conf.HostPasswordVault,
		Cache:               mgr.conf.Cache,
		InstallerFileGroup:  mgr.conf.InstallerFileGroup,
		ProxyMessager:       mgr.conf.ProxyMessager,
		PluginIface:         mgr,
	}

	// register action defs.
	if err := mgr.workflowMgr.RegisterActions(
		node.NewActionWaitInstallerComplete(nodeCap),
		node.NewActionWaitGseReady(nodeCap),
		node.NewActionTryReuseAgentID(nodeCap),
		node.NewActionBindAgentHostRel(nodeCap),
		node.NewActionInstallNodeBySSH(nodeCap),
		node.NewActionInstallNodeByWMI(nodeCap),
		node.NewActionSyncNodeInfo(nodeCap),
		node.NewActionPushHostIdentifier(nodeCap),
		node.NewActionRenderNodeDeployment(nodeCap),
		node.NewActionUpsertHostToCMDB(nodeCap),
		node.NewActionUpdateHost(nodeCap),
		node.NewActionTransferPkgToNode(nodeCap),
		node.NewActionDetectInfoBySSH(nodeCap),
		node.NewActionDetectInfoByWMI(nodeCap),
		node.NewActionUpgradeNode(nodeCap),
		node.NewActionCleanInstaller(nodeCap),
		node.NewActionVersionCompatCheck(nodeCap),
		node.NewActionReconfigNode(nodeCap),
		node.NewActionRestartNode(nodeCap),
		node.NewActionSelectRelayHost(nodeCap),
		node.NewActionEnsurePkgToRelay(nodeCap),
		node.NewActionPagentDetectInfoBySSH(nodeCap),
		node.NewActionInstallPagentBySSH(nodeCap),
		node.NewActionPagentDetectInfoByWMI(nodeCap),
		node.NewActionInstallPagentByWMI(nodeCap),
		node.NewActionEnableReleaseTransfer(nodeCap),
		node.NewActionUpgradePagent(nodeCap),
		node.NewActionUninstallNode(nodeCap),
		node.NewActionResetNodeDynamic(nodeCap),
		node.NewActionInstallPreOrderedPlugins(nodeCap),
		node.NewActionGenManualCommand(nodeCap),
		node.NewActionWaitDetectInfoByManual(nodeCap),
		node.NewActionInstallNodeByManual(nodeCap),
	); err != nil {
		return err
	}

	// register operation extra executions.
	if err := mgr.workflowMgr.RegisterOperExtraExecutions(
		node.NewOperationExtraExecution(nodeCap),
	); err != nil {
		return err
	}

	return nil
}

// registerDefSyncData registers the definitions for sync data.
func (mgr *Manager) registerDefSyncData() error {
	syncdataCap := &syncdata.Capability{
		CMDBHandler:         mgr.conf.CmdbHandler,
		GSEHandler:          mgr.conf.GSEHandler,
		FileHandler:         mgr.conf.FileHandler,
		UserManagerHandler:  mgr.conf.UserManagerHandler,
		StorageTopo:         mgr.conf.StorageTopo,
		StorageRelease:      mgr.conf.StorageRelease,
		StorageNode:         mgr.conf.StorageNode,
		StorageWorkflow:     mgr.conf.StorageWorkflow,
		StoragePlugin:       mgr.conf.StoragePlugin,
		StorageHostCredit:   mgr.conf.StorageHostCredit,
		StorageConfigPolicy: mgr.conf.StorageConfigPolicy,
		StorageTenant:       mgr.conf.StorageTenant,
		DiscoverProvider:    mgr.conf.Provider,
		Cache:               mgr.conf.Cache,
		WorkflowCtl:         mgr.workflowMgr,
	}

	// register action defs.
	if err := mgr.workflowMgr.RegisterActions(
		syncdata.NewActionSyncBusiness(syncdataCap),
		syncdata.NewActionSyncHost(syncdataCap),
		syncdata.NewActionSyncTenant(syncdataCap),
		syncdata.NewActionSyncNetworkArea(syncdataCap),
		syncdata.NewActionGenOperSyncHost(syncdataCap),
		syncdata.NewActionSyncAgentState(syncdataCap),
		syncdata.NewActionGenOperSyncAgentState(syncdataCap),
		syncdata.NewActionSyncAgentInfo(syncdataCap),
		syncdata.NewActionGenOperSyncAgentInfo(syncdataCap),
		syncdata.NewActionWatchCMDBResource(syncdataCap),
		syncdata.NewActionSyncAlivePluginProcessInfo(syncdataCap),
		syncdata.NewActionGenOperSyncAlivePluginProcessInfo(syncdataCap),
	); err != nil {
		return err
	}

	return nil
}

// registerDefPlugin registers the definitions for plugin.
// nolint: lll
func (mgr *Manager) registerDefPlugin() error {
	pluginCap := &plugin.Capability{
		CMDBHandler:         mgr.conf.CmdbHandler,
		GSEHandler:          mgr.conf.GSEHandler,
		FileHandler:         mgr.conf.FileHandler,
		StorageTopo:         mgr.conf.StorageTopo,
		StorageRelease:      mgr.conf.StorageRelease,
		StorageNode:         mgr.conf.StorageNode,
		StorageWorkflow:     mgr.conf.StorageWorkflow,
		StoragePlugin:       mgr.conf.StoragePlugin,
		StorageHostCredit:   mgr.conf.StorageHostCredit,
		StorageConfigPolicy: mgr.conf.StorageConfigPolicy,
		DiscoverProvider:    mgr.conf.Provider,
	}

	// register action defs.
	if err := mgr.workflowMgr.RegisterActions(
		plugin.NewActionTransferPluginPkgToNode(pluginCap),
		plugin.NewActionRenderPluginDeployment(pluginCap),
		plugin.NewActionRenderPluginConfig(pluginCap),
		plugin.NewActionWaitPluginInstallerComplete(pluginCap),
		plugin.NewActionInstallPlugin(pluginCap),
		plugin.NewActionUpsertProcess(pluginCap),
		plugin.NewActionPushPluginConfig(pluginCap),
		plugin.NewActionCheckPluginProcessAlive(pluginCap),
		plugin.NewActionEnsureAndUpdatePluginConfigDetails(pluginCap),
		plugin.NewActionUpdateProcess(pluginCap),
		plugin.NewActionStartProcess(pluginCap),
		plugin.NewActionStopProcess(pluginCap),
		plugin.NewActionRestartProcess(pluginCap),
		plugin.NewActionReloadProcess(pluginCap),
		plugin.NewActionTrusteeshipProcess(pluginCap),
		plugin.NewActionUnTrusteeshipProcess(pluginCap),
		plugin.NewActionVerifyPluginAvailability(pluginCap),
	); err != nil {
		return err
	}

	return nil
}

// registerDefPlugin registers the definitions for plugin.
// nolint: lll
func (mgr *Manager) registerDefSchedule() error {
	scheduleCap := &schedule.Capability{
		StorageWorkflow: mgr.conf.StorageWorkflow,
	}

	// register operation extra executions.
	if err := mgr.workflowMgr.RegisterOperExtraExecutions(
		schedule.NewOperationExtraExecution(scheduleCap),
	); err != nil {
		return err
	}

	return nil
}
