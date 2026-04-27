/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// ===============================================================================
// Node Manager Params
// ===============================================================================

// InstallNodeParam install node param.
type InstallNodeParam struct {
	Type            NodeWorkflowType
	BizIDs          []int64
	Operator        string
	NodeDeployments []*NodeDeployment
}

// UpgradeNodeParam upgrade node param.
type UpgradeNodeParam struct {
	Type            NodeWorkflowType
	BizIDs          []int64
	Operator        string
	NodeDeployments []*NodeDeployment
}

// ReconfigNodeParam reconfig node param.
type ReconfigNodeParam struct {
	Type            NodeWorkflowType
	BizIDs          []int64
	Operator        string
	NodeDeployments []*NodeDeployment
}

// RestartNodeParam restart node param.
type RestartNodeParam struct {
	Type            NodeWorkflowType
	BizIDs          []int64
	Operator        string
	NodeDeployments []*NodeDeployment
}

// UninstallNodeParam uninstall node param.
type UninstallNodeParam struct {
	Type            NodeWorkflowType
	BizIDs          []int64
	Operator        string
	NodeDeployments []*NodeDeployment
}

// AssignProxyUnitParam assign proxy unit param.
type AssignProxyUnitParam struct {
	Type            NodeWorkflowType
	BizIDs          []int64
	Operator        string
	NodeDeployments []*NodeDeployment
}

// RetryNodeWorkflowOperationParam retry node workflow operation param.
type RetryNodeWorkflowOperationParam struct {
	WorkflowID   string
	RetryMod     operation.RetryMode
	OperationIDs []string
}

// TerminateNodeWorkflowOperationParam terminate node workflow operation param.
type TerminateNodeWorkflowOperationParam struct {
	WorkflowID   string
	OperationIDs []string
}

// GetNodeWorklfowOperationManualInfoParam get node workflow operation manual info param.
type GetNodeWorklfowOperationManualInfoParam struct {
	WorkflowID  string
	OperationID string
}

// ===============================================================================
// Plugin Manager Params
// ===============================================================================

// InstallPluginParam define the param of LaunchInstallPlugin.
type InstallPluginParam struct {
	Type              PluginWorkflowType
	HostIDs           []int64
	BizIDs            []int64
	Operator          string
	PluginDeployments []*PluginDeployment
}

// UpgradePluginParam define the param of LaunchUpgradePlugin.
type UpgradePluginParam struct {
	Type              PluginWorkflowType
	HostIDs           []int64
	BizIDs            []int64
	Operator          string
	PluginDeployments []*PluginDeployment
}

// UninstallPluginParam define the param of LaunchUninstallPlugin.
type UninstallPluginParam struct {
	Type              PluginWorkflowType
	HostIDs           []int64
	BizIDs            []int64
	Operator          string
	PluginDeployments []*PluginDeployment
}

// ApplyPluginSubConfigParam define the param of LaunchApplyPluginSubConfig.
type ApplyPluginSubConfigParam struct {
	Type              PluginWorkflowType
	HostIDs           []int64
	BizIDs            []int64
	Operator          string
	PluginDeployments []*PluginDeployment
}

// RetryPluginWorkflowOperationParam retry node workflow operation param.
type RetryPluginWorkflowOperationParam struct {
	WorkflowID   string
	RetryMod     operation.RetryMode
	OperationIDs []string
}

// TerminatePluginWorkflowOperationParam terminate node workflow operation param.
type TerminatePluginWorkflowOperationParam struct {
	WorkflowID   string
	OperationIDs []string
}

// ===============================================================================
// DeployPolicy Manager Params
// ===============================================================================

// ExecuteDeployPolicyParam execute deploy policy param.
type ExecuteDeployPolicyParam struct {
	DeployPolicyIDs []int64
	Operator        string
}
