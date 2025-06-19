/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operinstdata ...
package operinstdata

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operation"
)

// OptFn option of find.
type OptFn = base.OptFn

// WithOperInstID filter by operation instance id.
func WithOperInstID(id ...string) OptFn {
	return base.WithStringValues(FieldKeyOperInstID, id...)
}

// WithOperationID filter by operation id.
func WithOperationID(id ...string) OptFn {
	return base.WithStringValues(FieldKeyOperationID, id...)
}

// WithTriggerID filter by trigger id.
func WithTriggerID(triggerID ...string) OptFn {
	return base.WithStringValues(FieldKeyTriggerID, triggerID...)
}

// WithState filter by state.
func WithState(states ...operation.State) OptFn {
	return base.WithStringValues(FieldKeyState, operation.StateListToStringList(states)...)
}
