/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package host

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// OptFn provides filtering options.
type OptFn = base.OptFn

// WithHostID filters by host-id.
func WithHostID(hostIDs ...int64) OptFn {
	return base.WithInt64Values("data.host_id", hostIDs...)
}

// WithoutHostID filters by not contains host-id.
func WithoutHostID(hostIDs ...int64) OptFn {
	return base.WithoutInt64Values("data.host_id", hostIDs...)
}

// WithBizID filters by biz-id.
func WithBizID(bizIDs ...int64) OptFn {
	return base.WithInt64Values("data.static.biz_id", bizIDs...)
}

// WithoutBizID filters by not contains biz-id.
func WithoutBizID(bizIDs ...int64) OptFn {
	return base.WithoutInt64Values("data.static.biz_id", bizIDs...)
}

// WithNetworkAreaID filters by network area id.
func WithNetworkAreaID(networkAreaIDs ...int64) OptFn {
	return base.WithInt64Values("data.static.network_area_id", networkAreaIDs...)
}

// WithoutNetworkAreaID filters by not contains network area id.
func WithoutNetworkAreaID(networkAreaIDs ...int64) OptFn {
	return base.WithoutInt64Values("data.static.network_area_id", networkAreaIDs...)
}

// WithFuzzyHostName filters by host name.
func WithFuzzyHostName(hostNames ...string) OptFn {
	return base.WithFuzzyValues("data.static.host_name", hostNames...)
}

// WithoutFuzzyHostName filters by not contains host name.
func WithoutFuzzyHostName(hostNames ...string) OptFn {
	return base.WithoutFuzzyValues("data.static.host_name", hostNames...)
}

// WithFuzzyDeptName filters by dept name.
func WithFuzzyDeptName(deptNames ...string) OptFn {
	return base.WithFuzzyValues("data.static.dept_name", deptNames...)
}

// WithoutFuzzyDeptName filters by not contains dept name.
func WithoutFuzzyDeptName(deptNames ...string) OptFn {
	return base.WithoutFuzzyValues("data.static.dept_name", deptNames...)
}

// WithFuzzyInnerIP filters by inner ip.
func WithFuzzyInnerIP(innerIPs ...string) OptFn {
	return base.WithFuzzyValues("data.static.inner_ip", innerIPs...)
}

// WithoutFuzzyInnerIP filters by not contains inner ip.
func WithoutFuzzyInnerIP(innerIPs ...string) OptFn {
	return base.WithoutFuzzyValues("data.static.inner_ip", innerIPs...)
}

// WithFuzzyInnerIPV6 filters by inner ipv6.
func WithFuzzyInnerIPV6(innerIPV6s ...string) OptFn {
	return base.WithFuzzyValues("data.static.inner_ipv6", innerIPV6s...)
}

// WithoutFuzzyInnerIPV6 filters by not contains inner ip.
func WithoutFuzzyInnerIPV6(innerIPV6s ...string) OptFn {
	return base.WithoutFuzzyValues("data.static.inner_ipv6", innerIPV6s...)
}

// WithFuzzyOuterIP filters by outer ip.
func WithFuzzyOuterIP(outerIps ...string) OptFn {
	return base.WithFuzzyValues("data.static.outer_ip", outerIps...)
}

// WithoutFuzzyOuterIP filters by not contains outer ip.
func WithoutFuzzyOuterIP(outerIps ...string) OptFn {
	return base.WithoutFuzzyValues("data.static.outer_ip", outerIps...)
}

// WithFuzzyOuterIPV6 filters by outer ip.
func WithFuzzyOuterIPV6(outerIPV6s ...string) OptFn {
	return base.WithFuzzyValues("data.static.outer_ipv6", outerIPV6s...)
}

// WithoutFuzzyOuterIPV6 filters by not contains outer ip.
func WithoutFuzzyOuterIPV6(outerIPV6s ...string) OptFn {
	return base.WithoutFuzzyValues("data.static.outer_ipv6", outerIPV6s...)
}

// WithOSType filters by os type.
func WithOSType(osTypes ...string) OptFn {
	return base.WithStringValues("data.static.os_type", osTypes...)
}

// WithoutOSType filters by not contains os type.
func WithoutOSType(osTypes ...string) OptFn {
	return base.WithoutStringValues("data.static.os_type", osTypes...)
}

// WithNodeRole filters by node role.
func WithNodeRole(roles ...types.NodeRole) OptFn {
	str := make([]string, len(roles))
	for idx, role := range roles {
		str[idx] = string(role)
	}

	return base.WithStringValues("data.dynamic.node_role", str...)
}

// WithoutNodeRole filters by not contains node role.
func WithoutNodeRole(roles ...types.NodeRole) OptFn {
	str := make([]string, len(roles))
	for idx, role := range roles {
		str[idx] = string(role)
	}

	return base.WithoutStringValues("data.dynamic.node_role", str...)
}

// WithNodeStatus filters by node status.
func WithNodeStatus(statuses ...types.NodeStatus) OptFn {
	str := make([]string, len(statuses))
	for idx, status := range statuses {
		str[idx] = string(status)
	}

	return base.WithStringValues("data.dynamic.node_status", str...)
}

// WithoutNodeStatus filters by not contains node status.
func WithoutNodeStatus(statuses ...types.NodeStatus) OptFn {
	str := make([]string, len(statuses))
	for idx, status := range statuses {
		str[idx] = string(status)
	}

	return base.WithoutStringValues("data.dynamic.node_status", str...)
}

// WithNodeVersion filters by node version.
func WithNodeVersion(versions ...string) OptFn {
	return base.WithStringValues("data.dynamic.node_version", versions...)
}

// WithoutNodeVersion filters by not contains node version.
func WithoutNodeVersion(versions ...string) OptFn {
	return base.WithoutStringValues("data.dynamic.node_version", versions...)
}

// WithAgentID filters by contains agent id.
func WithAgentID(agentIDs ...string) OptFn {
	return base.WithStringValues("data.dynamic.agent_id", agentIDs...)
}

// WithoutAgentID filters by not contains agent id.
func WithoutAgentID(agentIDs ...string) OptFn {
	return base.WithoutStringValues("data.dynamic.agent_id", agentIDs...)
}
