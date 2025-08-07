/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package configpolicy

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// OptFn provides filtering options.
type OptFn = base.OptFn

// WithConfigPolicyID filters by config policy id.
func WithConfigPolicyID(configPolicyIDs ...int64) OptFn {
	return base.WithValues(FieldKeyConfigPolicyID, configPolicyIDs...)
}

// WithBizID filters by biz id.
func WithBizID(bizIDs ...int64) OptFn {
	return base.WithValues(FieldKeyBizID, bizIDs...)
}

// WithNodeRole filters by node role.
func WithNodeRole(nodeRoles ...types.NodeRole) OptFn {
	return base.WithValues(FieldKeyNodeRole, types.NodeRoleListToStringList(nodeRoles)...)
}

// WithEnabled filters by enabled.
func WithEnabled(enabled ...bool) OptFn {
	return base.WithValues(FieldKeyEnabled, enabled...)
}

// WithFuzzyConfigPolicyName filters by config policy name.
func WithFuzzyConfigPolicyName(names ...string) OptFn {
	return base.WithFuzzyValues(FieldKeyConfigPolicyName, names...)
}
