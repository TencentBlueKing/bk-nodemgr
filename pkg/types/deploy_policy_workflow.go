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

// DeployPolicyExecutionParam carries the engine operation identity and execution context across retries.
type DeployPolicyExecutionParam struct {
	OperationID         string           `json:"operation_id"`
	TriggerID           string           `json:"trigger_id"`
	OperationInstanceID string           `json:"operation_instance_id"`
	WorkflowIDs         map[int64]string `json:"workflow_ids"`
}

// DeployPolicyWorkflow records one policy's execution and acknowledged child creation.
type DeployPolicyWorkflow struct {
	TenantID            string
	WorkflowID          string
	OperationID         string
	TriggerID           string
	DeployPolicyID      int64
	Operator            string
	OperateTime         time.Time
	OperationInstanceID string
	AttemptCount        int64
	// AttemptStatus is empty until the first attempt starts; it is not a final dispatch status.
	AttemptStatus DeployPolicyWorkflowAttemptStatus
	AttemptError  string
	Children      []DeployPolicyWorkflowChild
}

// DeployPolicyWorkflowAttemptStatus describes an attempt, not the final execution outcome.
type DeployPolicyWorkflowAttemptStatus string

const (
	// DeployPolicyWorkflowAttemptRunning indicates an attempt has started.
	DeployPolicyWorkflowAttemptRunning DeployPolicyWorkflowAttemptStatus = "running"
	// DeployPolicyWorkflowAttemptSuccess indicates the current attempt returned successfully.
	DeployPolicyWorkflowAttemptSuccess DeployPolicyWorkflowAttemptStatus = "success"
	// DeployPolicyWorkflowAttemptFailed indicates the current attempt failed and may be retried.
	DeployPolicyWorkflowAttemptFailed DeployPolicyWorkflowAttemptStatus = "failed"
)

// Validate checks whether the attempt status is supported.
func (status DeployPolicyWorkflowAttemptStatus) Validate() error {
	switch status {
	case DeployPolicyWorkflowAttemptRunning, DeployPolicyWorkflowAttemptSuccess, DeployPolicyWorkflowAttemptFailed:
		return nil
	default:
		return fmt.Errorf("invalid deploy policy workflow attempt status: %s", status)
	}
}

// DeployPolicyWorkflowChild records a child creation intent and its acknowledgement.
type DeployPolicyWorkflowChild struct {
	WorkflowID     string
	WorkflowDomain WorkflowDomain
	// Confirmed is false while creation is unknown, and never regresses after acknowledgement.
	Confirmed bool
}
