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
	WorkflowID  int64
	TriggerID   string
	OperType    NodeWorkflowOperType
	TaskType    NodeWorkflowTaskType
	BizIDs      []int64
	ExecuteUser string
	ExecuteTime time.Time
	Status      NodeWorkflowStatus
}

// NodeWorkflowOperType represents the operation type of a node workflow.
type NodeWorkflowOperType string

const (
	// NodeWorkflowOperTypeInstall is the operation type for install.
	NodeWorkflowOperTypeInstall NodeWorkflowOperType = "install"
)

// Validate checks if the NodeWorkflowOperType is valid.
func (nwo NodeWorkflowOperType) Validate() error {
	switch nwo {
	case NodeWorkflowOperTypeInstall:
		return nil
	default:
		return fmt.Errorf("invalid node workflow oper type, oper-type(%s)", nwo)
	}
}

// NodeWorkflowStatus represents the status of a node workflow.
type NodeWorkflowStatus string

const (
	// NodeWorkflowStatusRunning is the status when the workflow is running.
	NodeWorkflowStatusRunning NodeWorkflowStatus = "running"
	// NodeWorkflowStatusSuccess is the status when the workflow is successful.
	NodeWorkflowStatusSuccess NodeWorkflowStatus = "success"
	// NodeWorkflowStatusFailed is the status when the workflow has failed.
	NodeWorkflowStatusFailed NodeWorkflowStatus = "failed"
)

// Validate checks if the NodeWorkflowStatus is valid.
func (nws NodeWorkflowStatus) Validate() error {
	switch nws {
	case NodeWorkflowStatusRunning, NodeWorkflowStatusSuccess, NodeWorkflowStatusFailed:
		return nil
	}

	return fmt.Errorf("invalid node workflow status, status(%s)", nws)
}

// NodeWorkflowTaskType represents the task type of a node workflow.
type NodeWorkflowTaskType string

const (
	// NodeWorkflowTaskTypeAgent is the task type for agent.
	NodeWorkflowTaskTypeAgent NodeWorkflowTaskType = "agent"
	// NodeWorkflowTaskTypePlugin is the task type for plugin.
	NodeWorkflowTaskTypePlugin NodeWorkflowTaskType = "plugin"
)
