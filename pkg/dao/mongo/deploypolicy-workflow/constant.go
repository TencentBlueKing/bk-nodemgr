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

package deploypolicyworkflow

// Field keys address fields inside the common document envelope.
const (
	FieldKeyWorkflowID     = "data.workflow_id"
	FieldKeyOperationID    = "data.operation_id"
	FieldKeyDeployPolicyID = "data.deploy_policy_id"
	FieldKeyChildren       = "data.children"
	FieldKeyOperator       = "data.operator"
	FieldKeyOperateTime    = "data.operate_time"
	FieldKeyFinishTime     = "data.finish_time"
	FieldKeyStatus         = "data.status"
)
