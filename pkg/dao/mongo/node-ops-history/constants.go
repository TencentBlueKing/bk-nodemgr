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

package nodeopshistory

const (
	// FieldKeyTenantID identifies the tenant ID field.
	FieldKeyTenantID = "data.tenant_id"

	// FieldKeyOperateID identifies the operation ID field.
	FieldKeyOperateID = "data.operate_id"

	// FieldKeyOperator identifies the operator field.
	FieldKeyOperator = "data.operator"

	// FieldKeyOperateTime identifies the operation timestamp field.
	FieldKeyOperateTime = "data.operate_time"

	// FieldKeyNodeRole identifies the archived node role field.
	FieldKeyNodeRole = "data.node_role"

	// FieldKeyBizIDs identifies the business IDs involved in the operation.
	FieldKeyBizIDs = "data.biz_ids"

	// FieldKeySnapshots identifies the embedded host snapshots.
	FieldKeySnapshots = "data.snapshots"
)
