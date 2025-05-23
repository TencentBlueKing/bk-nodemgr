/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package trigger ...
package trigger

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigger"
)

// OptFn provides filtering options.
type OptFn = base.OptFn

// WithState filters by state.
func WithState(states ...trigger.State) OptFn {
	return base.WithStringValues(FieldKeyState, trigger.StateListToStringList(states)...)
}

// WithCategory filter by category.
func WithCategory(category ...trigger.Category) OptFn {
	return base.WithStringValues(FieldKeyCategory, trigger.CategoryListToStringList(category)...)
}

// WithTriggerID filter by trigger id.
func WithTriggerID(triggerID ...string) OptFn {
	return base.WithStringValues(FieldKeyTriggerID, triggerID...)
}
