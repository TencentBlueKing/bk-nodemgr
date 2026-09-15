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

import "time"

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
	Children            []DeployPolicyWorkflowChild
}

// DeployPolicyWorkflowChild records a child creation intent and its acknowledgement.
type DeployPolicyWorkflowChild struct {
	WorkflowID     string
	WorkflowDomain WorkflowDomain
	// Confirmed is false while creation is unknown, and never regresses after acknowledgement.
	Confirmed bool
}
