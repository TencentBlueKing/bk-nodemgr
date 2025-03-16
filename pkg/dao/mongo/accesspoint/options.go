/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package accesspoint

import "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"

// OptFn provides filtering options.
type OptFn = base.OptFn

// WithAccessPointID filters by accesspoint-id.
func WithAccessPointID(accessPointIDs ...int64) OptFn {
	return base.WithInt64Values("data.accesspoint_id", accessPointIDs...)
}

// WithoutAccessPointID filters by not contains accesspoint-id.
func WithoutAccessPointID(accessPointIDs ...int64) OptFn {
	return base.WithoutInt64Values("data.accesspoint_id", accessPointIDs...)
}

// WithNetworkAreaID filters by networkarea-id.
func WithNetworkAreaID(networkAreaIDs ...int64) OptFn {
	return base.WithInt64Values("data.networkarea_id", networkAreaIDs...)
}

// WithoutNetworkAreaID filters by not contains networkarea-id.
func WithoutNetworkAreaID(networkAreaIDs ...int64) OptFn {
	return base.WithoutInt64Values("data.networkarea_id", networkAreaIDs...)
}
