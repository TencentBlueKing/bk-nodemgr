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
	TenantID    string
	WorkflowID  string
	TriggerID   string
	Type        NodeWorkflowType
	BizIDs      []int64
	Operator    string
	OperateTime time.Time
	FinishTime  time.Time
	Status      NodeWorkflowStatus
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
)

// Validate checks if the NodeWorkflowType is valid.
func (nwo NodeWorkflowType) Validate() error {
	switch nwo {
	case NodeWorkflowTypeInstallAgent, NodeWorkflowTypeInstallProxy,
		NodeWorkflowTypeUpgradeAgent, NodeWorkflowTypeUpgradeProxy,
		NodeWorkflowTypeReconfigAgent, NodeWorkflowTypeReconfigProxy,
		NodeWorkflowTypeRestartAgent, NodeWorkflowTypeRestartProxy,
		NodeWorkflowTypeUninstallAgent, NodeWorkflowTypeUninstallProxy:
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

// NodeWorkflowOperationStatus ...
type NodeWorkflowOperationStatus struct {
	Index       int
	OperationID string
	TriggerID   string
	State       NodeWorkflowOperationState
}

// NodeWorkflowOperationState defines the state of node workflow operation.
type NodeWorkflowOperationState string

const (
	// NodeWorkflowOperationStateInit node workflow operation state init.
	NodeWorkflowOperationStateInit NodeWorkflowOperationState = "init"

	// NodeWorkflowOperationStateLaunched node workflow operation state launched.
	NodeWorkflowOperationStateLaunched NodeWorkflowOperationState = "launched"

	// NodeWorkflowOperationStateRunning node workflow operation state running.
	NodeWorkflowOperationStateRunning NodeWorkflowOperationState = "running"

	// NodeWorkflowOperationStateSuccess node workflow operation state success.
	NodeWorkflowOperationStateSuccess NodeWorkflowOperationState = "success"

	// NodeWorkflowOperationStateFailed node workflow operation state failed.
	NodeWorkflowOperationStateFailed NodeWorkflowOperationState = "failed"

	// NodeWorkflowOperationStateTimeout node workflow operation state timeout.
	NodeWorkflowOperationStateTimeout NodeWorkflowOperationState = "timeout"

	// NodeWorkflowOperationStateTerminated node workflow operation state terminated.
	NodeWorkflowOperationStateTerminated NodeWorkflowOperationState = "terminated"
)

// WorkflowOperationStatusListToStringList converts a node status list to a string list.
func WorkflowOperationStatusListToStringList(operationStatusList []NodeWorkflowOperationState) []string {
	data := make([]string, len(operationStatusList))
	for idx, operationStatus := range operationStatusList {
		data[idx] = string(operationStatus)
	}

	return data
}

// StringListToWorkflowOperationStatusList converts a string list to a node status list.
func StringListToWorkflowOperationStatusList(stringList []string) []NodeWorkflowOperationState {
	data := make([]NodeWorkflowOperationState, len(stringList))
	for idx, operationStatus := range stringList {
		data[idx] = NodeWorkflowOperationState(operationStatus)
	}

	return data
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
	NodeVersion     string
	NetworkAreaID   int64
	InnerIP         string
	InnerIPV6       string
	BizID           int64
	Operator        string
	OperationID     string
	OperInstanceIDs []string
}

// OperationSummary ...
type OperationSummary struct {
	TotalDuration int64
	LastStatus    NodeWorkflowOperationState
}

// InstanceStatusToNodeWorkflowOperationState converts the instance status to operation state.
func InstanceStatusToNodeWorkflowOperationState(status operation.State) (NodeWorkflowOperationState, error) {
	switch status {
	case operation.StateInit:
		return NodeWorkflowOperationStateInit, nil
	case operation.StateLaunched:
		return NodeWorkflowOperationStateLaunched, nil
	case operation.StateRunning:
		return NodeWorkflowOperationStateRunning, nil
	case operation.StateSuccess:
		return NodeWorkflowOperationStateSuccess, nil
	case operation.StateFailed:
		return NodeWorkflowOperationStateFailed, nil
	case operation.StateTimeout:
		return NodeWorkflowOperationStateTimeout, nil
	case operation.StateTerminated:
		return NodeWorkflowOperationStateTerminated, nil
	default:
		return "", fmt.Errorf("invalid operation instance state. state(%s)", status)
	}
}
