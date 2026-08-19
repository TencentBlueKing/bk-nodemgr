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

package manager

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/dpmgr"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/deploypolicy"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/pkg"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/plugin"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/pluginv2"
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

	if err := mgr.registerDefPluginV2(); err != nil {
		return fmt.Errorf("failed to register def plugin v2: %w", err)
	}

	if err := mgr.registerDefPackage(); err != nil {
		return fmt.Errorf("failed to register def package import: %w", err)
	}

	if err := mgr.registerDefSchedule(); err != nil {
		return fmt.Errorf("failed to register def schedule: %w", err)
	}

	if err := mgr.registerDefDeployPolicy(); err != nil {
		return fmt.Errorf("failed to register def deploy policy: %w", err)
	}

	return nil
}

func (mgr *Manager) registerDefPackage() error {
	capability := &pkg.Capability{
		FileHandler:    mgr.conf.FileHandler,
		StorageRelease: mgr.conf.StorageRelease,
		StoragePackage: mgr.conf.StoragePackage,
	}

	return mgr.workflowMgr.RegisterActions(
		pkg.NewActionPackageImportPluginV3PkgFetchAndUpload(capability),
		pkg.NewActionPackagePublishPluginV3Pkg(capability),
		pkg.NewActionPackageReleasePluginEnable(capability),
		pkg.NewActionPackageReleasePluginHidden(capability),
	)
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
		FileCache:           mgr.conf.FileCache,
		ProxyMessager:       mgr.conf.ProxyMessager,
		PluginIface:         mgr,
	}

	// register action defs.
	if err := mgr.workflowMgr.RegisterActions(
		node.NewActionWaitInstallerComplete(nodeCap),
		node.NewActionWaitGseReady(nodeCap),
		node.NewActionWaitGseNotAlive(nodeCap),
		node.NewActionTryReuseAgentID(nodeCap),
		node.NewActionBindAgentHostRel(nodeCap),
		node.NewActionInstallNodeBySSH(nodeCap),
		node.NewActionInstallProxyBySSH(nodeCap),
		node.NewActionInstallNodeByWMI(nodeCap),
		node.NewActionInstallNodeByWindowsSSH(nodeCap),
		node.NewActionInstallNodeByWindowsAuto(nodeCap),
		node.NewActionSyncNodeInfo(nodeCap),
		node.NewActionPushHostIdentifier(nodeCap),
		node.NewActionRenderNodeDeployment(nodeCap),
		node.NewActionUpsertHostToCMDB(nodeCap),
		node.NewActionUpdateHost(nodeCap),
		node.NewActionTransferPkgToNode(nodeCap),
		node.NewActionDetectInfoBySSH(nodeCap),
		node.NewActionDetectInfoByWMI(nodeCap),
		node.NewActionDetectInfoByWindowsSSH(nodeCap),
		node.NewActionDetectInfoByWindowsAuto(nodeCap),
		node.NewActionUpgradeNode(nodeCap),
		node.NewActionCheckSelfRelayPluginAlive(nodeCap),
		node.NewActionUpgradeProxy(nodeCap),
		node.NewActionCleanInstaller(nodeCap),
		node.NewActionVersionCompatCheck(nodeCap),
		node.NewActionReconfigNode(nodeCap),
		node.NewActionReconfigProxy(nodeCap),
		node.NewActionRestartNode(nodeCap),
		node.NewActionSelectRelayHost(nodeCap),
		node.NewActionEnsurePkgToRelay(nodeCap),
		node.NewActionPagentDetectInfoBySSH(nodeCap),
		node.NewActionInstallPagentBySSH(nodeCap),
		node.NewActionPagentDetectInfoByWindowsSSH(nodeCap),
		node.NewActionInstallPagentByWindowsSSH(nodeCap),
		node.NewActionPagentDetectInfoByWMI(nodeCap),
		node.NewActionInstallPagentByWMI(nodeCap),
		node.NewActionUpgradePagent(nodeCap),
		node.NewActionUninstallNode(nodeCap),
		node.NewActionUninstallNodeSkipReport(nodeCap),
		node.NewActionReconfigPagent(nodeCap),
		node.NewActionUninstallPagent(nodeCap),
		node.NewActionResetNodeDynamic(nodeCap),
		node.NewActionInstallPreOrderedPlugins(nodeCap),
		node.NewActionGenManualCommand(nodeCap),
		node.NewActionWaitDetectInfoByManual(nodeCap),
		node.NewActionInstallNodeByManual(nodeCap),
		node.NewActionResolveOfflineDetectInfo(nodeCap),
		node.NewActionWaitOfflineManualInstall(nodeCap),
		node.NewActionInjectNodeCustomDeployConfig(nodeCap),
		node.NewActionAssignProxyInfo(nodeCap),
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
		SyncIface:           mgr,
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
		syncdata.NewActionSyncCorrectAgentID(syncdataCap),
		syncdata.NewActionWatchCMDBResource(syncdataCap),
		syncdata.NewActionSyncAlivePluginProcessInfo(syncdataCap),
		syncdata.NewActionGenOperSyncAlivePluginProcessInfo(syncdataCap),
	); err != nil {
		return err
	}

	return nil
}

// registerDefDeployPolicy registers the definitions for deploy policy.
func (mgr *Manager) registerDefDeployPolicy() error {
	dpMgr := dpmgr.NewHandler(&dpmgr.Config{
		DaoProcess:            mgr.conf.StoragePlugin,
		DaoPlugin:             mgr.conf.StoragePlugin,
		DaoHost:               mgr.conf.StorageTopo,
		DomainDeployPolicyMgr: mgr.conf.StorageDeployPolicy,
		CmdbHandler:           mgr.conf.CmdbHandler,
		NodeManager:           mgr,
		PluginManager:         mgr,
	})

	deployPolicyCap := &deploypolicy.Capability{
		DPMgr:               dpMgr,
		StorageDeployPolicy: mgr.conf.StorageDeployPolicy,
		WorkflowCtl:         mgr.workflowMgr,
	}

	// register action defs.
	if err := mgr.workflowMgr.RegisterActions(
		deploypolicy.NewActionGenOperExecuteDeployPolicy(deployPolicyCap),
		deploypolicy.NewActionExecuteDeployPolicy(deployPolicyCap),
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
		PluginIface:         mgr,
	}

	// register action defs.
	if err := mgr.workflowMgr.RegisterActions(
		plugin.NewActionTransferPluginPkgToNode(pluginCap),
		plugin.NewActionRenderPluginDeployment(pluginCap),
		plugin.NewActionRenderPluginConfig(pluginCap),
		plugin.NewActionOverwritePluginConfigForCompatibility(pluginCap),
		plugin.NewActionWaitPluginInstallerComplete(pluginCap),
		plugin.NewActionInstallPlugin(pluginCap),
		plugin.NewActionUpgradePlugin(pluginCap),
		plugin.NewActionUninstallPlugin(pluginCap),
		plugin.NewActionUpsertProcess(pluginCap),
		plugin.NewActionPushPluginConfig(pluginCap),
		plugin.NewActionFetchPluginProcess(pluginCap),
		plugin.NewActionCheckPluginProcessAlive(pluginCap),
		plugin.NewActionEnsureAndUpdatePluginConfigDetails(pluginCap),
		plugin.NewActionUpdateProcess(pluginCap),
		plugin.NewActionDeleteProcess(pluginCap),
		plugin.NewActionStartProcess(pluginCap),
		plugin.NewActionStopProcess(pluginCap),
		plugin.NewActionRestartProcess(pluginCap),
		plugin.NewActionReloadProcess(pluginCap),
		plugin.NewActionTrusteeshipProcess(pluginCap),
		plugin.NewActionUnTrusteeshipProcess(pluginCap),
		plugin.NewActionTryStopProcess(pluginCap),
		plugin.NewActionVerifyPluginAvailability(pluginCap),
		plugin.NewActionFetchProcessSubConfigIntoDeployment(pluginCap),
		plugin.NewActionInjectPluginCustomDeployConfig(pluginCap),
		plugin.NewActionStopPluginV2Process(pluginCap),
	); err != nil {
		return err
	}

	return nil
}

// registerDefPluginV2 registers the definitions for plugin v2.
// nolint: lll
func (mgr *Manager) registerDefPluginV2() error {
	pluginCap := &pluginv2.Capability{
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
		pluginv2.NewActionTransferPluginPkgToNodeV2(pluginCap),
		pluginv2.NewActionRenderPluginDeploymentV2(pluginCap),
		pluginv2.NewActionRenderPluginConfigV2(pluginCap),
		pluginv2.NewActionWaitPluginInstallerCompleteV2(pluginCap),
		pluginv2.NewActionInstallPluginV2(pluginCap),
		pluginv2.NewActionUpgradePluginV2(pluginCap),
		pluginv2.NewActionUninstallPluginV2(pluginCap),
		pluginv2.NewActionPushPluginConfigV2(pluginCap),
		pluginv2.NewActionFetchPluginProcessV2(pluginCap),
		pluginv2.NewActionCheckPluginProcessAliveV2(pluginCap),
		pluginv2.NewActionEnsureAndUpdatePluginConfigDetailsV2(pluginCap),
		pluginv2.NewActionStartProcessV2(pluginCap),
		pluginv2.NewActionStopProcessV2(pluginCap),
		pluginv2.NewActionRestartProcessV2(pluginCap),
		pluginv2.NewActionReloadProcessV2(pluginCap),
		pluginv2.NewActionTrusteeshipProcessV2(pluginCap),
		pluginv2.NewActionUnTrusteeshipProcessV2(pluginCap),
		pluginv2.NewActionVerifyPluginAvailabilityV2(pluginCap),
		pluginv2.NewActionInjectPluginBaseRuntimeV2(pluginCap),
		pluginv2.NewActionFinishIfPluginProcessV2Alive(pluginCap),
		pluginv2.NewActionFinishIfPluginProcessV2NotAlive(pluginCap),
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
