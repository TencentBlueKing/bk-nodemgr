/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodedeployment

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
)

// OptFn option of find.
type OptFn = base.OptFn

// WithToken set token.
func WithToken(token ...string) OptFn {
	return base.WithStringValues(FieldKeyToken, token...)
}

// WithInfoInnerIP set info inner ip.
func WithInfoInnerIP(ips ...string) OptFn {
	return base.WithStringValues(FieldKeyInfoInnerIPList, ips...)
}

// WithInfoInnerIPV6 set info inner ipv6.
func WithInfoInnerIPV6(ips ...string) OptFn {
	return base.WithStringValues(FieldKeyInfoInnerIPV6List, ips...)
}

// WithInfoBizID set info biz id.
func WithInfoBizID(ids ...int64) OptFn {
	return base.WithInt64Values(FieldKeyInfoBizID, ids...)
}

// WithInfoNetworkAreaID set info network area id.
func WithInfoNetworkAreaID(ids ...int64) OptFn {
	return base.WithInt64Values(FieldKeyInfoNetworkAreaID, ids...)
}

// WithInfoNetworkUnitID set info network unit id.
func WithInfoNetworkUnitID(ids ...int64) OptFn {
	return base.WithInt64Values(FieldKeyInfoNetworkUnitID, ids...)
}

// WithInfoNodeVersion set info node version.
func WithInfoNodeVersion(version ...string) OptFn {
	return base.WithStringValues(FieldKeyInfoNodeVersion, version...)
}
