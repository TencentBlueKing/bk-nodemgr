/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cipher

import "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"

// OptFn provides filtering options.
type OptFn = base.OptFn

// WithName filters by name.
func WithName(names ...string) OptFn {
	return base.WithValues(FieldKeyName, names...)
}

// WithoutName filters by not contains name.
func WithoutName(names ...string) OptFn {
	return base.WithoutValues(FieldKeyName, names...)
}

// WithKeyType filters by key-type.
func WithKeyType(keyTypes ...string) OptFn {
	return base.WithValues(FieldKeyKeyType, keyTypes...)
}

// WithoutKeyType filters by not contains key-type.
func WithoutKeyType(keyTypes ...string) OptFn {
	return base.WithoutValues(FieldKeyKeyType, keyTypes...)
}
