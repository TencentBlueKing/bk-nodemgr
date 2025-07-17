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
)

// NodeWorkflow represents the workflow of a node.
type NodeWorkflow struct {
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
		NodeWorkflowTypeRestartAgent, NodeWorkflowTypeRestartProxy:
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
	State       OperationState
}

// OperationState ...
type OperationState string

// NodeWorkflowOperationStatusList.
const (
	// StateInit operation instance state init.
	StateInit OperationState = "init"

	// StateLaunched operation instance state launched.
	StateLaunched OperationState = "launched"

	// StateRunning operation instance state running.
	StateRunning OperationState = "running"

	// StateSuccess operation instance state success.
	StateSuccess OperationState = "success"

	// StateFailed operation instance state failed.
	StateFailed OperationState = "failed"

	// StateTimeout operation instance state timeout.
	StateTimeout OperationState = "timeout"

	// StateTerminated operation instance state terminated.
	StateTerminated OperationState = "terminated"
)

// WorkflowOperationStatusListToStringList converts a node status list to a string list.
func WorkflowOperationStatusListToStringList(operationStatusList []OperationState) []string {
	data := make([]string, len(operationStatusList))
	for idx, operationStatus := range operationStatusList {
		data[idx] = string(operationStatus)
	}

	return data
}

// StringListToWorkflowOperationStatusList converts a string list to a node status list.
func StringListToWorkflowOperationStatusList(stringList []string) []OperationState {
	data := make([]OperationState, len(stringList))
	for idx, operationStatus := range stringList {
		data[idx] = OperationState(operationStatus)
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

// OperationSummary ...
type OperationSummary struct {
	TotalDuration int64
	LastStatus    string
}
