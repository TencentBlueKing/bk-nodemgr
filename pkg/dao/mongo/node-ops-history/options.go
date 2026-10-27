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

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// OptFn provides archive filtering options.
type OptFn = base.OptFn

// WithOperateID filters exact operation IDs.
func WithOperateID(ids ...string) OptFn {
	return base.WithValues(FieldKeyOperateID, ids...)
}

// WithoutOperateID excludes exact operation IDs.
func WithoutOperateID(ids ...string) OptFn {
	return base.WithoutValues(FieldKeyOperateID, ids...)
}

// WithOperator filters exact operator names.
func WithOperator(operators ...string) OptFn {
	return base.WithValues(FieldKeyOperator, operators...)
}

// WithoutOperator excludes exact operator names.
func WithoutOperator(operators ...string) OptFn {
	return base.WithoutValues(FieldKeyOperator, operators...)
}

// WithNodeRole filters exact node roles.
func WithNodeRole(roles ...string) OptFn {
	return base.WithValues(FieldKeyNodeRole, roles...)
}

// WithoutNodeRole excludes exact node roles.
func WithoutNodeRole(roles ...string) OptFn {
	return base.WithoutValues(FieldKeyNodeRole, roles...)
}

// WithOperateTimeRange filters operation timestamps.
func WithOperateTimeRange(r types.TimeRange) OptFn {
	return base.WithTimeRange(FieldKeyOperateTime, r.StartTime, r.EndTime)
}

// WithBizID filters operations containing the supplied business IDs.
func WithBizID(ids ...int64) OptFn {
	return base.WithValues(FieldKeyBizIDs, ids...)
}

// WithoutBizID excludes operations containing the supplied business IDs.
func WithoutBizID(ids ...int64) OptFn {
	return base.WithoutValues(FieldKeyBizIDs, ids...)
}
