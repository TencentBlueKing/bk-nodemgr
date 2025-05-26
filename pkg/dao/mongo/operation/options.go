/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operation ...
package operation

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// OptFn provides filtering options.
type OptFn = base.OptFn

// WithTriggerID filter by trigger id.
func WithTriggerID(triggerID ...string) OptFn {
	return base.WithValues(FieldKeyTriggerID, triggerID...)
}

// WithOperationID ...
func WithOperationID(ids ...string) OptFn {
	return base.WithValues(FieldKeyOperationID, ids...)
}

// WithEmptyOperation filter by operation_instance.
func WithEmptyOperation() OptFn {
	return base.WithValues(FieldKeyOperationInstanceEmpty, true)
}
