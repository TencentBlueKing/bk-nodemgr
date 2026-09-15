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

// Package iface defines the manager interfaces.
package iface

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// INodeManager defines the NodeManager interface.
type INodeManager interface {
	// LaunchInstallNode launch a task to install node. returns the workflow-id.
	LaunchInstallNode(ctx contextx.IContext, param types.InstallNodeParam) (string, error)

	// LaunchUpgradeNode launch a task to upgrade node. returns the workflow-id.
	LaunchUpgradeNode(ctx contextx.IContext, param types.UpgradeNodeParam) (string, error)

	// LaunchReconfigNode launch a task to reconfig node. returns the workflow-id.
	LaunchReconfigNode(ctx contextx.IContext, param types.ReconfigNodeParam) (string, error)

	// LaunchRestartNode launch a task to restart node. returns the workflow-id.
	LaunchRestartNode(ctx contextx.IContext, param types.RestartNodeParam) (string, error)

	// LaunchUninstallNode launch a task to uninstall node. returns the workflow-id.
	LaunchUninstallNode(ctx contextx.IContext, param types.UninstallNodeParam) (string, error)

	// LaunchAssignProxyUnit launch a task to assign proxy unit. returns the workflow-id.
	LaunchAssignProxyUnit(ctx contextx.IContext, param types.AssignProxyUnitParam) (string, error)

	// AssignAgentNetworkUnit assigns a network unit to agent hosts.
	AssignAgentNetworkUnit(nCtx contextx.IContext, param types.NodeAgentAssignUnitParam) (*types.NodeAgentAssignUnitResult, error)

	// LaunchRetryOperationFromLastInstance launch a task to retry operation from last instance.
	LaunchRetryNodeOperationFromLastInstance(ctx contextx.IContext, param types.RetryNodeWorkflowOperationParam) error

	// TerminateOperationLastInstance terminate operation from last instance.
	TerminateNodeOperationLastInstance(ctx contextx.IContext, param types.TerminateNodeWorkflowOperationParam) error

	// GetOperationManualInfoFromLastInstance get operation manual info from last instance.
	GetOperationManualInfoFromLastInstance(nCtx contextx.IContext, param types.GetNodeWorklfowOperationManualInfoParam) (
		*types.NodeWorkflowOperationManualInfo, error)
}

// IPluginManager defines the PluginManager interface.
type IPluginManager interface {
	// LaunchRetryPluginOperationFromLastInstance launch a task to retry operation from last instance.
	LaunchRetryPluginOperationFromLastInstance(nCtx contextx.IContext, param types.RetryPluginWorkflowOperationParam) error

	// TerminatePluginOperationLastInstance terminate operation from last instance.
	TerminatePluginOperationLastInstance(nCtx contextx.IContext, param types.TerminatePluginWorkflowOperationParam) error

	iPluginManagerPlugin
	iPluginManagerPluginV2
}

// iPluginManagerPlugin defines the PluginManager sub interface for.
// nolint: interfacebloat
type iPluginManagerPlugin interface {
	// LaunchInstallPlugin launch a task to install plugin. returns the workflow-id.
	LaunchInstallPlugin(ctx contextx.IContext, param types.InstallPluginParam) (string, error)

	// LaunchUpgradePlugin launch a task to upgrade plugin. returns the workflow-id.
	LaunchUpgradePlugin(nCtx contextx.IContext, param types.UpgradePluginParam) (string, error)

	// LaunchUninstallPlugin launch a task to uninstall plugin. returns the workflow-id.
	LaunchUninstallPlugin(nCtx contextx.IContext, param types.UninstallPluginParam) (string, error)

	// LaunchApplyPluginSubConfig launch a task to apply plugin subconfig. returns the workflow-id.
	LaunchApplyPluginSubConfig(nCtx contextx.IContext, param types.ApplyPluginSubConfigParam) (string, error)

	// LaunchRemovePluginSubConfig launch a task to remove plugin subconfig. returns the workflow-id.
	LaunchRemovePluginSubConfig(nCtx contextx.IContext, param types.RemovePluginSubConfigParam) (string, error)

	// LaunchStartProcess launch a task to start process. returns the workflow-id.
	LaunchStartProcess(nCtx contextx.IContext, param types.StartProcessParam) (string, error)

	// LaunchRestartProcess launch a task to restart process. returns the workflow-id.
	LaunchRestartProcess(nCtx contextx.IContext, param types.RestartProcessParam) (string, error)

	// LaunchMigrateFromV2 launch a task to migrate plugin process from v2. returns the workflow-id.
	LaunchMigrateFromV2(nCtx contextx.IContext, param types.MigrateFromV2Param) (string, error)

	// LaunchStopProcess launch a task to stop process. returns the workflow-id.
	LaunchStopProcess(nCtx contextx.IContext, param types.StopProcessParam) (string, error)

	// LaunchStartDebugPlugin launches a task to debug one prepared plugin deployment.
	LaunchStartDebugPlugin(nCtx contextx.IContext, param types.StartDebugPluginParam) (string, error)

	// LaunchStopDebugPlugin stores a stop signal for a debug plugin workflow.
	LaunchStopDebugPlugin(nCtx contextx.IContext, param types.StopDebugPluginParam) error
}

// iPluginManagerPluginV2 defines the PluginManager sub interface for v2.
type iPluginManagerPluginV2 interface {
	// LaunchPluginEnsurePluginV2 launch a task to ensure v2 plugin. returns the workflow-id.
	LaunchPluginEnsurePluginV2(ctx contextx.IContext, param types.InstallPluginParam) (string, error)

	// LaunchUninstallPluginV2 launch a task to uninstall v2 plugin. returns the workflow-id.
	LaunchUninstallPluginV2(nCtx contextx.IContext, param types.UninstallPluginParam) (string, error)

	// LaunchStopPluginV2 launch a task to stop v2 plugin. returns the workflow-id.
	LaunchStopPluginV2(nCtx contextx.IContext, param types.StopProcessParam) (string, error)
}

// ISyncManager defines the SyncManager interface.
type ISyncManager interface {
	// LaunchSyncBizAndHost launch a task to sync biz and host. returns the trigger-id.
	LaunchSyncBizAndHost(ctx contextx.IContext) (string, error)

	// LaunchSyncHostByBizID launch a task to sync host by biz-id. returns the trigger-id.
	LaunchSyncHostByBizID(ctx contextx.IContext, bizID int64) (string, error)

	// LaunchSyncNetworkArea launch a task to sync networkarea. returns the trigger-id.
	LaunchSyncNetworkArea(ctx contextx.IContext) (string, error)

	// LaunchSyncAgentState launch a task to sync agent state from gse. returns the workflow-id.
	LaunchSyncAgentState(ctx contextx.IContext, hostIDs ...int64) (string, error)

	// LaunchSyncAllAgentState launch a task to sync all agent state from gse. returns the workflow-id.
	LaunchSyncAllAgentState(ctx contextx.IContext) (string, error)

	// LaunchSyncAgentInfo launch a task to sync agent info from gse. returns the workflow-id.
	LaunchSyncAgentInfo(ctx contextx.IContext, hostIDs ...int64) (string, error)

	// LaunchSyncCorrectAgentID launch a task to correct agent id. returns the workflow-id.
	LaunchSyncCorrectAgentID(ctx contextx.IContext, hostIDs ...int64) (string, error)

	// LaunchSyncAliveHostAgentInfo launch a task to sync alive host agent info. returns the trigger-id.
	LaunchSyncAliveHostAgentInfo(ctx contextx.IContext) (string, error)

	// LaunchSyncAlivePluginProcessInfo launch a task to sync alive plugin process info. returns the workflow-id.
	LaunchSyncAlivePluginProcessInfo(ctx contextx.IContext, hostIDs ...int64) (string, error)

	// LaunchSyncAllAlivePluginProcessInfo launch a task to sync all alive plugin process info. returns the workflow-id.
	LaunchSyncAllAlivePluginProcessInfo(ctx contextx.IContext) (string, error)
}

// IDeployPolicyManager defines the DeployPolicyManager interface.
type IDeployPolicyManager interface {
	// LaunchExecuteDeployPolicy launch a task to execute deploy policy. returns the trigger-id.
	LaunchExecuteDeployPolicy(ctx contextx.IContext, param types.ExecuteDeployPolicyParam) (string, error)

	// LaunchExecuteDeployPolicyWorkflow launches one deploy policy and returns its workflow ID.
	LaunchExecuteDeployPolicyWorkflow(ctx contextx.IContext, param types.ExecuteDeployPolicyParam) (string, error)
}

// IPackageManager defines package workflow manager methods.
type IPackageManager interface {
	// LaunchPackageImportPluginV3Pkg launches package import plugin v3 pkg workflow.
	LaunchPackageImportPluginV3Pkg(ctx contextx.IContext, param types.PackageImportParam) (string, error)
	// LaunchPackageImportPluginV2Pkg launches the package import plugin v2 pkg workflow.
	LaunchPackageImportPluginV2Pkg(ctx contextx.IContext, param types.PackageImportParam) (string, error)
	// LaunchPackageImportExternalPluginV2Pkg launches the package import external plugin v2 pkg workflow.
	LaunchPackageImportExternalPluginV2Pkg(ctx contextx.IContext, param types.PackageImportParam) (string, error)
	// LaunchPackageExportPlugin launches a plugin package export workflow.
	LaunchPackageExportPlugin(ctx contextx.IContext, param types.PackageExportParam) (string, error)
}
