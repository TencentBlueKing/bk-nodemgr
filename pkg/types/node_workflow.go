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
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// NodeWorkflow represents the workflow of a node.
type NodeWorkflow struct {
	TenantID       string
	WorkflowID     string
	TriggerID      string
	Type           NodeWorkflowType
	BizIDs         []int64
	NetworkAreaIDs []int64
	NetworkUnitIDs []int64
	NodeRoles      []NodeRole
	Operator       string
	OperateTime    time.Time
	FinishTime     time.Time
	Status         NodeWorkflowStatus
}

// NodeWorkflowType represents the operation type of a node workflow.
type NodeWorkflowType string

// NodeWorkflowTypeListToStringList converts the NodeWorkflowTypeList to a string list.
func NodeWorkflowTypeListToStringList(nodeWorkflowTypeList []NodeWorkflowType) []string {
	data := make([]string, len(nodeWorkflowTypeList))
	for i, nodeWorkflowType := range nodeWorkflowTypeList {
		data[i] = string(nodeWorkflowType)
	}

	return data
}

// StringListToNodeWorkflowTypeList converts the string list to a NodeWorkflowType list.
func StringListToNodeWorkflowTypeList(stringList []string) []NodeWorkflowType {
	data := make([]NodeWorkflowType, len(stringList))
	for i, str := range stringList {
		data[i] = NodeWorkflowType(str)
	}

	return data
}

const (
	// NodeWorkflowTypeInstallAgent is the operation type for install.
	NodeWorkflowTypeInstallAgent NodeWorkflowType = "install_agent"

	// NodeWorkflowTypeInstallProxy is the operation type for install proxy.
	NodeWorkflowTypeInstallProxy NodeWorkflowType = "install_proxy"

	// NodeWorkflowTypeUpgradeAgent is the operation type for upgrade.
	NodeWorkflowTypeUpgradeAgent NodeWorkflowType = "upgrade_agent"

	// NodeWorkflowTypeUpgradeProxy is the operation type for upgrade proxy.
	NodeWorkflowTypeUpgradeProxy NodeWorkflowType = "upgrade_proxy"

	// NodeWorkflowTypeReconfigAgent is the operation type for reconfig.
	NodeWorkflowTypeReconfigAgent NodeWorkflowType = "reconfig_agent"

	// NodeWorkflowTypeUninstallAgent is the operation type for reconfig.
	NodeWorkflowTypeUninstallAgent NodeWorkflowType = "uninstall_agent"

	// NodeWorkflowTypeUninstallProxy is the operation type for reconfig.
	NodeWorkflowTypeUninstallProxy NodeWorkflowType = "uninstall_proxy"

	// NodeWorkflowTypeReconfigProxy is the operation type for reconfig proxy.
	NodeWorkflowTypeReconfigProxy NodeWorkflowType = "reconfig_proxy"

	// NodeWorkflowTypeRestartAgent is the operation type for restart agent.
	NodeWorkflowTypeRestartAgent NodeWorkflowType = "restart_agent"

	// NodeWorkflowTypeRestartProxy is the operation type for restart proxy.
	NodeWorkflowTypeRestartProxy NodeWorkflowType = "restart_proxy"

	// NodeWorkflowTypeAssignProxyUnit is the operation type for assign proxy unit.
	NodeWorkflowTypeAssignProxyUnit NodeWorkflowType = "assign_proxy_unit"
)

// Validate checks if the NodeWorkflowType is valid.
func (nwo NodeWorkflowType) Validate() error {
	switch nwo {
	case NodeWorkflowTypeInstallAgent, NodeWorkflowTypeInstallProxy,
		NodeWorkflowTypeUpgradeAgent, NodeWorkflowTypeUpgradeProxy,
		NodeWorkflowTypeReconfigAgent, NodeWorkflowTypeReconfigProxy,
		NodeWorkflowTypeRestartAgent, NodeWorkflowTypeRestartProxy,
		NodeWorkflowTypeUninstallAgent, NodeWorkflowTypeUninstallProxy,
		NodeWorkflowTypeAssignProxyUnit:
		return nil
	default:
		return fmt.Errorf("invalid node workflow oper type, oper-type(%s)", nwo)
	}
}

// NodeWorkflowStatus represents the status of a node workflow.
type NodeWorkflowStatus string

// NodeWorkflowStatusListToStringList converts the NodeWorkflowStatusList to a string list.
func NodeWorkflowStatusListToStringList(status []NodeWorkflowStatus) []string {
	data := make([]string, len(status))
	for i, s := range status {
		data[i] = string(s)
	}

	return data
}

// StringListToNodeWorkflowStatusList converts the string list to a NodeWorkflowStatus list.
func StringListToNodeWorkflowStatusList(stringList []string) []NodeWorkflowStatus {
	data := make([]NodeWorkflowStatus, len(stringList))
	for i, str := range stringList {
		data[i] = NodeWorkflowStatus(str)
	}

	return data
}

const (
	// NodeWorkflowStatusRunning is the status when the workflow is running.
	NodeWorkflowStatusRunning NodeWorkflowStatus = "running"

	// NodeWorkflowStatusSuccess is the status when the workflow is successful.
	NodeWorkflowStatusSuccess NodeWorkflowStatus = "success"

	// NodeWorkflowStatusFailed is the status when the workflow has failed.
	NodeWorkflowStatusFailed NodeWorkflowStatus = "failed"

	// NodeWorkflowStatusPartialFailed is the status when the workflow has partially failed.
	NodeWorkflowStatusPartialFailed NodeWorkflowStatus = "partial_failed"
)

// Validate checks if the NodeWorkflowStatus is valid.
func (nws NodeWorkflowStatus) Validate() error {
	switch nws {
	case NodeWorkflowStatusRunning,
		NodeWorkflowStatusSuccess,
		NodeWorkflowStatusFailed,
		NodeWorkflowStatusPartialFailed:
		return nil
	}

	return fmt.Errorf("invalid node workflow status, status(%s)", nws)
}

// GetFinishedNodeWorkflowStatus returns the finished node workflow status.
func GetFinishedNodeWorkflowStatus() []NodeWorkflowStatus {
	return []NodeWorkflowStatus{
		NodeWorkflowStatusSuccess,
		NodeWorkflowStatusFailed,
		NodeWorkflowStatusPartialFailed,
	}
}

// NodeWorkflowListOperationResult operation list result.
type NodeWorkflowListOperationResult struct {
	HostID        int64
	BizID         int64
	NetworkAreaID int64
	NetworkUnitID int64
	InnerIPList   []string
	InnerIPV6List []string
	NodeVersion   string

	OperationID     string
	OperInstanceIDs []string
	Operator        string
	CreateTime      time.Time

	// LastInstanceBriefData is the last instance brief data.
	// currently support life-cycle and latest action.
	LastInstanceBriefData *operation.InstanceBriefData
}

// NodeWorkflowOperationRetryParam validates the retry param.
type NodeWorkflowOperationRetryParam struct {
	WorkflowID   string
	OperationIDs []string
	RetryMode    operation.RetryMode
}

// NodeWorkflowOperationTerminateParam validates the terminate param.
type NodeWorkflowOperationTerminateParam struct {
	WorkflowID   string
	OperationIDs []string
}

// NodeWorkflowOperationManualCommandType describes the manual command type.
type NodeWorkflowOperationManualCommandType string

const (
	// NodeWorkflowOperationManualCommandTypeBash the bash command type.
	NodeWorkflowOperationManualCommandTypeBash NodeWorkflowOperationManualCommandType = "bash"
	// NodeWorkflowOperationManualCommandTypeBat the bat command type.
	NodeWorkflowOperationManualCommandTypeBat NodeWorkflowOperationManualCommandType = "bat"
)

// NodeWorkflowOperationManualCommand describes the manual command.
type NodeWorkflowOperationManualCommand struct {
	Type    NodeWorkflowOperationManualCommandType
	Command string
}

// NodeWorkflowOperationManualInfo describes the manual info.
type NodeWorkflowOperationManualInfo struct {
	NetworkPolicies []*NetworkPolicy
	Commands        []*NodeWorkflowOperationManualCommand
}
