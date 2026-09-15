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

// DeployPolicyWorkflow records one policy's operation and acknowledged child creation.
type DeployPolicyWorkflow struct {
	TenantID       string
	WorkflowID     string
	OperationID    string
	TriggerID      string
	DeployPolicyID int64
	Operator       string
	OperateTime    time.Time
	Children       []DeployPolicyWorkflowChild
}

// DeployPolicyWorkflowChild records a child creation intent and its acknowledgement.
type DeployPolicyWorkflowChild struct {
	WorkflowID     string
	WorkflowDomain WorkflowDomain
	// Confirmed is false while creation is unknown, and never regresses after acknowledgement.
	Confirmed bool
}
