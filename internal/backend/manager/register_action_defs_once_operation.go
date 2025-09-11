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
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/syncdata"
)

// registerOnceTriggerActions registers the action definitions for once trigger actions.
func (mgr *Manager) registerActionDefsOnceOperation() error {
	if err := mgr.registerActionDefNodeInstall(); err != nil {
		return fmt.Errorf("register action def node install failed, err: %w", err)
	}

	if err := mgr.registerActionDefSyncData(); err != nil {
		return fmt.Errorf("register action def sync data failed, err: %w", err)
	}

	return nil
}

// registerActionDefNodeInstall registers the action definitions for node installation operations.
// nolint: lll
func (mgr *Manager) registerActionDefNodeInstall() error {
	return mgr.workflowMgr.RegisterActions(
		node.NewActionTryReuseAgentID(mgr.conf.StorageTopo, mgr.conf.StorageNode, mgr.logger),
		node.NewActionBindAgentHostRel(mgr.conf.CmdbHandler, mgr.conf.StorageTopo, mgr.conf.StorageNode, mgr.logger),
		node.NewActionInstallNodeBySSH(mgr.conf.InstallerFileGroup, mgr.logger, mgr.conf.StorageNode, mgr.conf.Provider, mgr.conf.StorageHostCredit, mgr.conf.HostPasswordVault),
		node.NewActionInstallNodeByWMI(mgr.conf.InstallerFileGroup, mgr.logger, mgr.conf.StorageNode, mgr.conf.Provider, mgr.conf.StorageHostCredit, mgr.conf.HostPasswordVault),
		node.NewActionWaitGseReady(mgr.conf.GSEHandler, mgr.conf.StorageNode, mgr.logger),
		node.NewActionSyncNodeInfo(mgr.conf.GSEHandler, mgr.conf.StorageNode, mgr.logger),
		node.NewActionPushHostIdentifier(mgr.conf.CmdbHandler, mgr.conf.StorageNode, mgr.logger),
		node.NewActionRenderNodeDeployment(mgr.conf.StorageNode, mgr.conf.StorageTopo, mgr.conf.StorageTopo, mgr.conf.StorageRelease, mgr.conf.StorageConfigPolicy, mgr.logger),
		node.NewActionUpsertHostToCMDB(mgr.conf.CmdbHandler, mgr.conf.StorageTopo, mgr.conf.StorageNode),
		node.NewActionWaitInstallerComplete(mgr.conf.StorageOperInst, mgr.logger),
		node.NewActionUpdateHost(mgr.conf.StorageTopo, mgr.conf.StorageNode, mgr.logger),
		node.NewActionTransferPkgToNode(mgr.conf.StorageNode, mgr.conf.FileHandler, mgr.logger),
		node.NewActionDetectInfoBySSH(mgr.logger, mgr.conf.StorageNode, mgr.conf.StorageRelease, mgr.conf.StorageHostCredit, mgr.conf.HostPasswordVault),
		node.NewActionDetectInfoByWMI(mgr.logger, mgr.conf.StorageNode, mgr.conf.StorageRelease, mgr.conf.StorageHostCredit, mgr.conf.HostPasswordVault),
		node.NewActionUpgradeNode(mgr.conf.StorageNode, mgr.conf.GSEHandler, mgr.logger, mgr.conf.Provider),
		node.NewActionCleanInstaller(mgr.conf.StorageNode, mgr.conf.GSEHandler, mgr.logger),
		node.NewActionVersionCompatCheck(mgr.conf.StorageNode, mgr.logger),
		node.NewActionReconfigNode(mgr.conf.StorageNode, mgr.conf.GSEHandler, mgr.logger, mgr.conf.Provider),
		node.NewActionRestartNode(mgr.conf.StorageNode, mgr.conf.GSEHandler, mgr.logger),
		node.NewActionSelectRelayHost(mgr.conf.StorageTopo, mgr.conf.StorageNode, mgr.logger),
		node.NewActionEnsurePkgToRelay(mgr.conf.InstallerFileGroup, mgr.conf.StorageRelease, mgr.conf.StorageOperInst, mgr.conf.StorageNode, mgr.conf.FileHandler, mgr.conf.ProxyMessager, mgr.logger),
		node.NewActionPagentDetectInfoBySSH(mgr.logger, mgr.conf.StorageOperInst, mgr.conf.StorageNode, mgr.conf.StorageRelease, mgr.conf.StorageHostCredit, mgr.conf.HostPasswordVault, mgr.conf.ProxyMessager),
		node.NewActionInstallPagentBySSH(mgr.conf.ProxyMessager, mgr.conf.StorageNode, mgr.conf.StorageHostCredit, mgr.conf.StorageOperInst, mgr.conf.HostPasswordVault, mgr.logger),
		node.NewActionPagentDetectInfoByWMI(mgr.logger, mgr.conf.StorageOperInst, mgr.conf.StorageNode, mgr.conf.StorageRelease, mgr.conf.StorageHostCredit, mgr.conf.HostPasswordVault, mgr.conf.ProxyMessager),
		node.NewActionInstallPagentByWMI(mgr.conf.ProxyMessager, mgr.conf.StorageNode, mgr.conf.StorageHostCredit, mgr.conf.StorageOperInst, mgr.conf.HostPasswordVault, mgr.logger),
	)
}

// registerActionDefSyncData registers the action definitions for sync data operations.
func (mgr *Manager) registerActionDefSyncData() error {
	return mgr.workflowMgr.RegisterActions(
		syncdata.NewActionSyncBusinessFromCMDB(mgr.conf.CmdbHandler, mgr.conf.StorageTopo, mgr.logger),
		syncdata.NewActionSyncHostFromCMDB(mgr.conf.CmdbHandler, mgr.conf.StorageTopo),
		syncdata.NewActionSyncNetworkAreaFromCMDB(mgr.conf.CmdbHandler, mgr.conf.StorageTopo),
		syncdata.NewActionGenOperSyncHost(mgr.conf.StorageTopo, mgr.workflowMgr),
		syncdata.NewActionSyncAgentState(mgr.conf.GSEHandler, mgr.conf.StorageTopo, mgr.logger),
		syncdata.NewActionGenOperSyncAgentState(mgr.conf.StorageTopo, mgr.workflowMgr),
		syncdata.NewActionWatchCMDBResource(mgr.conf.Cache, mgr.conf.CmdbHandler, mgr.conf.StorageTopo),
	)
}
