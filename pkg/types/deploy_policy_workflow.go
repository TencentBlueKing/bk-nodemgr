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

package types

import (
	"fmt"
	"time"
)

// DeployPolicyWorkflow records one policy's operation and associated child workflows.
type DeployPolicyWorkflow struct {
	TenantID       string
	WorkflowID     string
	OperationID    string
	TriggerID      string
	DeployPolicyID int64
	Operator       string
	OperateTime    time.Time
	FinishTime     time.Time
	Status         DeployPolicyWorkflowStatus
	Children       []DeployPolicyWorkflowChild
}

// DeployPolicyWorkflowStatus represents the status of a deploy policy workflow.
type DeployPolicyWorkflowStatus string

const (
	// DeployPolicyWorkflowStatusRunning is the status while the operation is unfinished.
	DeployPolicyWorkflowStatusRunning DeployPolicyWorkflowStatus = "running"
	// DeployPolicyWorkflowStatusSuccess is the status when the operation succeeds.
	DeployPolicyWorkflowStatusSuccess DeployPolicyWorkflowStatus = "success"
	// DeployPolicyWorkflowStatusFailed is the status when the operation fails or times out.
	DeployPolicyWorkflowStatusFailed DeployPolicyWorkflowStatus = "failed"
	// DeployPolicyWorkflowStatusPartialFailed is the status when the operation is terminated.
	DeployPolicyWorkflowStatusPartialFailed DeployPolicyWorkflowStatus = "partial_failed"
)

// Validate checks if the deploy policy workflow status is valid.
func (status DeployPolicyWorkflowStatus) Validate() error {
	switch status {
	case DeployPolicyWorkflowStatusRunning, DeployPolicyWorkflowStatusSuccess,
		DeployPolicyWorkflowStatusFailed, DeployPolicyWorkflowStatusPartialFailed:
		return nil
	}

	return fmt.Errorf("invalid deploy policy workflow status, status(%s)", status)
}

// GetFinishedDeployPolicyWorkflowStatus returns the finished workflow statuses.
func GetFinishedDeployPolicyWorkflowStatus() []DeployPolicyWorkflowStatus {
	return []DeployPolicyWorkflowStatus{
		DeployPolicyWorkflowStatusSuccess,
		DeployPolicyWorkflowStatusFailed,
		DeployPolicyWorkflowStatusPartialFailed,
	}
}

// DeployPolicyWorkflowExactConditions defines exact workflow filters.
type DeployPolicyWorkflowExactConditions struct {
	WorkflowID     []string
	DeployPolicyID []int64
	Status         []DeployPolicyWorkflowStatus
	Operator       []string
}

// DeployPolicyWorkflowCondition defines deploy policy workflow query conditions.
type DeployPolicyWorkflowCondition struct {
	OperateTimeRange *TimeRange
	ExactInclude     *DeployPolicyWorkflowExactConditions
	ExactExclude     *DeployPolicyWorkflowExactConditions
}

// DeployPolicyWorkflowChild identifies an associated child workflow.
type DeployPolicyWorkflowChild struct {
	WorkflowID     string
	WorkflowDomain WorkflowDomain
}
