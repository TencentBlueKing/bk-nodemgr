/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package networkunit

import "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"

// OptFn provides filtering options.
type OptFn = base.OptFn

// WithNetworkUnitID filters by networkunit-id.
func WithNetworkUnitID(networkUnitIDs ...int64) OptFn {
	return base.WithInt64Values("data.networkunit_id", networkUnitIDs...)
}

// WithoutNetworkUnitID filters by not contains networkunit-id.
func WithoutNetworkUnitID(networkUnitIDs ...int64) OptFn {
	return base.WithoutInt64Values("data.networkunit_id", networkUnitIDs...)
}

// WithNetworkAreaID filters by networkarea-id.
func WithNetworkAreaID(networkAreaIDs ...int64) OptFn {
	return base.WithInt64Values("data.networkarea_id", networkAreaIDs...)
}

// WithoutNetworkAreaID filters by not contains networkarea-id.
func WithoutNetworkAreaID(networkAreaIDs ...int64) OptFn {
	return base.WithoutInt64Values("data.networkarea_id", networkAreaIDs...)
}

// WithIsDirect filters by is_direct.
func WithIsDirect(isDirect ...bool) OptFn {
	return base.WithValues("data.is_direct", isDirect...)
}

// WithoutIsDirect filters by not contains is_direct.
func WithoutIsDirect(isDirect ...bool) OptFn {
	return base.WithoutValues("data.is_direct", isDirect...)
}
