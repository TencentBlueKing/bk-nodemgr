/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operation provides the dao for operation related constants.
package operation

const (
	// FieldKeyOperationID the operation_id field key.
	FieldKeyOperationID = "data.operation_id"

	// FieldKeyTriggerID the trigger_id field key.
	FieldKeyTriggerID = "data.trigger_id"

	// FieldKeyBizID the biz_id field key.
	FieldKeyBizID = "data.extras.biz_id"

	// FieldKeyAreaID the area_id field key.
	FieldKeyAreaID = "data.extras.area_id"

	// FieldKeyIPV4 the ipv4 field key.
	FieldKeyIPV4 = "data.extras.inner_ip"

	// FieldKeyIPV6 the ipv6 field key.
	FieldKeyIPV6 = "data.extras.inner_ip_v6"

	// FieldKeyOperationInstanceEmpty the oper_inst_empty field key.
	FieldKeyOperationInstanceEmpty = "data.oper_inst_empty"
)
