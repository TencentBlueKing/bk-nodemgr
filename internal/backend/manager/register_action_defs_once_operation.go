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
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

// registerOnceTriggerActions registers the action definitions for once trigger actions.
func (mgr *Manager) registerActionDefsOnceOperation() error {
	if err := mgr.registerActionDefCommon(); err != nil {
		return fmt.Errorf("failed to register action def common: %w", err)
	}

	if err := mgr.registerActionDefNode(); err != nil {
		return fmt.Errorf("failed to register action def node install: %w", err)
	}

	if err := mgr.registerActionDefSyncData(); err != nil {
		return fmt.Errorf("failed to register action def sync data: %w", err)
	}

	if err := mgr.registerActionDefPlugin(); err != nil {
		return fmt.Errorf("failed to register action defs plugin install: %w", err)
	}

	return nil
}

func (mgr *Manager) registerActionDefCommon() error {
	return mgr.workflowMgr.RegisterActions(
		node.NewActionWaitInstallerComplete(mgr.conf.StorageWorkflow),
	)
}

// registerActionDefNode registers the action definitions for node operations.
// nolint: lll
func (mgr *Manager) registerActionDefNode() error {
	return mgr.workflowMgr.RegisterActions(
		node.NewActionWaitGseReady(mgr.conf.GSEHandler, mgr.conf.StorageNode),
		node.NewActionTryReuseAgentID(mgr.conf.StorageTopo, mgr.conf.StorageNode),
		node.NewActionBindAgentHostRel(mgr.conf.CmdbHandler, mgr.conf.StorageTopo, mgr.conf.StorageNode),
		node.NewActionInstallNodeBySSH(mgr.conf.InstallerFileGroup, mgr.conf.StorageNode, mgr.conf.Provider, mgr.conf.StorageHostCredit, mgr.conf.HostPasswordVault),
		node.NewActionInstallNodeByWMI(mgr.conf.InstallerFileGroup, mgr.conf.StorageNode, mgr.conf.Provider, mgr.conf.StorageHostCredit, mgr.conf.HostPasswordVault),
		node.NewActionSyncNodeInfo(mgr.conf.GSEHandler, mgr.conf.StorageNode),
		node.NewActionPushHostIdentifier(mgr.conf.CmdbHandler, mgr.conf.StorageNode),
		node.NewActionRenderNodeDeployment(mgr.conf.StorageNode, mgr.conf.StorageTopo, mgr.conf.StorageTopo, mgr.conf.StorageRelease, mgr.conf.StorageConfigPolicy),
		node.NewActionUpsertHostToCMDB(mgr.conf.CmdbHandler, mgr.conf.StorageTopo, mgr.conf.StorageNode),
		node.NewActionUpdateHost(mgr.conf.StorageTopo, mgr.conf.StorageNode),
		node.NewActionTransferPkgToNode(mgr.conf.StorageNode, mgr.conf.FileHandler),
		node.NewActionDetectInfoBySSH(mgr.conf.StorageNode, mgr.conf.StorageRelease, mgr.conf.StorageHostCredit, mgr.conf.HostPasswordVault),
		node.NewActionDetectInfoByWMI(mgr.conf.StorageNode, mgr.conf.StorageRelease, mgr.conf.StorageHostCredit, mgr.conf.HostPasswordVault),
		node.NewActionUpgradeNode(mgr.conf.StorageNode, mgr.conf.GSEHandler, mgr.conf.Provider),
		node.NewActionCleanInstaller(mgr.conf.StorageNode, mgr.conf.GSEHandler),
		node.NewActionVersionCompatCheck(mgr.conf.StorageNode),
		node.NewActionReconfigNode(mgr.conf.StorageNode, mgr.conf.GSEHandler, mgr.conf.Provider),
		node.NewActionRestartNode(mgr.conf.StorageNode, mgr.conf.GSEHandler),
		node.NewActionSelectRelayHost(mgr.conf.StorageTopo, mgr.conf.StorageNode),
		node.NewActionEnsurePkgToRelay(mgr.conf.InstallerFileGroup, mgr.conf.StorageRelease, mgr.conf.StorageWorkflow, mgr.conf.StorageNode, mgr.conf.FileHandler, mgr.conf.ProxyMessager),
		node.NewActionPagentDetectInfoBySSH(mgr.conf.StorageWorkflow, mgr.conf.StorageNode, mgr.conf.StorageRelease, mgr.conf.StorageHostCredit, mgr.conf.HostPasswordVault, mgr.conf.ProxyMessager),
		node.NewActionInstallPagentBySSH(mgr.conf.ProxyMessager, mgr.conf.StorageNode, mgr.conf.StorageHostCredit, mgr.conf.StorageWorkflow, mgr.conf.HostPasswordVault),
		node.NewActionPagentDetectInfoByWMI(mgr.conf.StorageWorkflow, mgr.conf.StorageNode, mgr.conf.StorageRelease, mgr.conf.StorageHostCredit, mgr.conf.HostPasswordVault, mgr.conf.ProxyMessager),
		node.NewActionInstallPagentByWMI(mgr.conf.ProxyMessager, mgr.conf.StorageNode, mgr.conf.StorageHostCredit, mgr.conf.StorageWorkflow, mgr.conf.HostPasswordVault),
		node.NewActionEnableReleaseTransfer(mgr.conf.StorageNode),
		node.NewActionUpgradePagent(mgr.conf.StorageNode, mgr.conf.GSEHandler, mgr.conf.Provider),
	)
}

// registerActionDefSyncData registers the action definitions for sync data operations.
func (mgr *Manager) registerActionDefSyncData() error {
	return mgr.workflowMgr.RegisterActions(
		syncdata.NewActionSyncBusinessFromCMDB(mgr.conf.CmdbHandler, mgr.conf.StorageTopo),
		syncdata.NewActionSyncHostFromCMDB(mgr.conf.CmdbHandler, mgr.conf.StorageTopo),
		syncdata.NewActionSyncNetworkAreaFromCMDB(mgr.conf.CmdbHandler, mgr.conf.StorageTopo),
		syncdata.NewActionGenOperSyncHost(mgr.conf.StorageTopo, mgr.workflowMgr),
		syncdata.NewActionSyncAgentState(mgr.conf.GSEHandler, mgr.conf.StorageTopo),
		syncdata.NewActionGenOperSyncAgentState(mgr.conf.StorageTopo, mgr.workflowMgr),
		syncdata.NewActionSyncAgentInfo(mgr.conf.GSEHandler, mgr.conf.StorageTopo),
		syncdata.NewActionSyncAliveHostAgentInfo(mgr.conf.StorageTopo, mgr.workflowMgr),
		syncdata.NewActionWatchCMDBResource(mgr.conf.Cache, mgr.conf.CmdbHandler, mgr.conf.StorageTopo),
	)
}

// registerActionDefPlugin registers the action definitions for plugin operations.
// nolint: lll
func (mgr *Manager) registerActionDefPlugin() error {
	actionDefs := []action.Definition{
		plugin.NewActionTransferPluginPkgToNode(mgr.conf.StoragePlugin, mgr.conf.StorageTopo, mgr.conf.FileHandler),
		plugin.NewActionRenderPluginDeployment(mgr.conf.StorageTopo, mgr.conf.StoragePlugin),
		plugin.NewActionWaitInstallerComplete(mgr.conf.StorageWorkflow),
		plugin.NewActionInstallPlugin(mgr.conf.StorageTopo, mgr.conf.StoragePlugin, mgr.conf.Provider, mgr.conf.GSEHandler),
	}

	if err := mgr.workflowMgr.RegisterActions(actionDefs...); err != nil {
		return fmt.Errorf("failed to reigster actions: %w", err)
	}

	return nil
}
