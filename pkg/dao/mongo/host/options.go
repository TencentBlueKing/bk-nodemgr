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
	return base.WithValues(FieldKeyHostID, hostIDs...)
}

// WithoutHostID filters by not contains host-id.
func WithoutHostID(hostIDs ...int64) OptFn {
	return base.WithoutValues(FieldKeyHostID, hostIDs...)
}

// WithBizID filters by biz-id.
func WithBizID(bizIDs ...int64) OptFn {
	return base.WithValues(FieldKeyStaticBizID, bizIDs...)
}

// WithoutBizID filters by not contains biz-id.
func WithoutBizID(bizIDs ...int64) OptFn {
	return base.WithoutValues(FieldKeyStaticBizID, bizIDs...)
}

// WithNetworkAreaID filters by network area id.
func WithNetworkAreaID(networkAreaIDs ...int64) OptFn {
	return base.WithValues(FieldKeyStaticNetworkAreaID, networkAreaIDs...)
}

// WithoutNetworkAreaID filters by not contains network area id.
func WithoutNetworkAreaID(networkAreaIDs ...int64) OptFn {
	return base.WithoutValues(FieldKeyStaticNetworkAreaID, networkAreaIDs...)
}

// WithNetworkUnitID filters by network unit id.
func WithNetworkUnitID(networkUnitID ...int64) OptFn {
	return base.WithValues(FieldKeyDynamicNetworkUnitID, networkUnitID...)
}

// WithoutNetworkUnitID filters by not contains network unit id.
func WithoutNetworkUnitID(networkUnitID ...int64) OptFn {
	return base.WithoutValues(FieldKeyDynamicNetworkUnitID, networkUnitID...)
}

// WithFuzzyHostName filters by host name.
func WithFuzzyHostName(hostNames ...string) OptFn {
	return base.WithFuzzyValues(FieldKeyStaticHostName, hostNames...)
}

// WithoutFuzzyHostName filters by not contains host name.
func WithoutFuzzyHostName(hostNames ...string) OptFn {
	return base.WithoutFuzzyValues(FieldKeyStaticHostName, hostNames...)
}

// WithFuzzyDeptName filters by dept name.
func WithFuzzyDeptName(deptNames ...string) OptFn {
	return base.WithFuzzyValues(FieldKeyStaticDeptName, deptNames...)
}

// WithoutFuzzyDeptName filters by not contains dept name.
func WithoutFuzzyDeptName(deptNames ...string) OptFn {
	return base.WithoutFuzzyValues(FieldKeyStaticDeptName, deptNames...)
}

// WithOSType filters by os type.
func WithOSType(osTypes ...string) OptFn {
	return base.WithValues(FieldKeyStaticOSType, osTypes...)
}

// WithoutOSType filters by not contains os type.
func WithoutOSType(osTypes ...string) OptFn {
	return base.WithoutValues(FieldKeyStaticOSType, osTypes...)
}

// WithArch filters by arch.
func WithArch(archs ...string) OptFn {
	return base.WithValues(FieldKeyDynamicNodeCPUArch, archs...)
}

// WithoutArch filters by not contains arch.
func WithoutArch(archs ...string) OptFn {
	return base.WithoutValues(FieldKeyDynamicNodeCPUArch, archs...)
}

// WithNodeRole filters by node role.
func WithNodeRole(roles ...types.NodeRole) OptFn {
	str := make([]string, len(roles))
	for idx, role := range roles {
		str[idx] = string(role)
	}

	return base.WithValues(FieldKeyDynamicNodeRole, str...)
}

// WithoutNodeRole filters by not contains node role.
func WithoutNodeRole(roles ...types.NodeRole) OptFn {
	str := make([]string, len(roles))
	for idx, role := range roles {
		str[idx] = string(role)
	}

	return base.WithoutValues(FieldKeyDynamicNodeRole, str...)
}

// WithNodeStatus filters by node status.
func WithNodeStatus(statuses ...types.NodeStatus) OptFn {
	str := make([]string, len(statuses))
	for idx, status := range statuses {
		str[idx] = string(status)
	}

	return base.WithValues(FieldKeyDynamicNodeStatus, str...)
}

// WithoutNodeStatus filters by not contains node status.
func WithoutNodeStatus(statuses ...types.NodeStatus) OptFn {
	str := make([]string, len(statuses))
	for idx, status := range statuses {
		str[idx] = string(status)
	}

	return base.WithoutValues(FieldKeyDynamicNodeStatus, str...)
}

// WithDynamicProxyTags filters by proxy tag.
func WithDynamicProxyTags(tags ...types.ProxyTag) OptFn {
	str := make([]string, len(tags))
	for idx, tag := range tags {
		str[idx] = string(tag)
	}

	return base.WithValues(FieldKeyDynamicProxyTags, str...)
}

// WithDynamicProxyAccessDisabled filters by not contains proxy access disabled.
func WithDynamicProxyAccessDisabled(bools ...bool) OptFn {
	return base.WithValues(FieldKeyDynamicProxyAccessDisabled, bools...)
}

// WithNodeVersion filters by node version.
func WithNodeVersion(versions ...string) OptFn {
	return base.WithValues(FieldKeyDynamicNodeVersion, versions...)
}

// WithNodeGeneration filters by node generation.
func WithNodeGeneration(generations ...int64) OptFn {
	return base.WithValues(FieldKeyDynamicNodeGeneration, generations...)
}

// WithoutNodeVersion filters by not contains node version.
func WithoutNodeVersion(versions ...string) OptFn {
	return base.WithoutValues(FieldKeyDynamicNodeVersion, versions...)
}

// WithAgentID filters by contains agent id.
func WithAgentID(agentIDs ...string) OptFn {
	return base.WithValues(FieldKeyDynamicAgentID, agentIDs...)
}

// WithStaticAddressing filters by contains addressing.
func WithStaticAddressing(addressings ...types.Addressing) OptFn {
	strs := make([]string, len(addressings))
	for idx, addressing := range addressings {
		strs[idx] = string(addressing)
	}

	return base.WithValues(FieldKeyStaticAddressing, strs...)
}

// WithStaticInnerIPList filters by contains inner ip list.
func WithStaticInnerIPList(ips ...string) OptFn {
	return base.WithValues(FieldKeyStaticInnerIPList, ips...)
}

// WithFuzzyStaticInnerIPList filters by contains inner ip.
func WithFuzzyStaticInnerIPList(ips ...string) OptFn {
	return base.WithFuzzyValues(FieldKeyStaticInnerIPList, ips...)
}

// WithoutFuzzyStaticInnerIPList filters by not contains inner ip list.
func WithoutFuzzyStaticInnerIPList(ips ...string) OptFn {
	return base.WithoutFuzzyValues(FieldKeyStaticInnerIPList, ips...)
}

// WithoutFuzzyStaticInnerIPV6List filters by not contains inner ip v6 list.
func WithoutFuzzyStaticInnerIPV6List(ips ...string) OptFn {
	return base.WithoutFuzzyValues(FieldKeyStaticInnerIPV6List, ips...)
}

// WithStaticInnerIPV6List filters by contains inner ip v6 list.
func WithStaticInnerIPV6List(ips ...string) OptFn {
	return base.WithValues(FieldKeyStaticInnerIPV6List, ips...)
}

// WithFuzzyStaticInnerIPV6List filters by contains inner ip v6 list.
func WithFuzzyStaticInnerIPV6List(ips ...string) OptFn {
	return base.WithFuzzyValues(FieldKeyStaticInnerIPV6List, ips...)
}

// WithStaticOuterIPList filters by contains outer ip list.
func WithStaticOuterIPList(ips ...string) OptFn {
	return base.WithValues(FieldKeyStaticOuterIPList, ips...)
}

// WithFuzzyStaticOuterIPList filters by contains outer ip v6 list.
func WithFuzzyStaticOuterIPList(ips ...string) OptFn {
	return base.WithFuzzyValues(FieldKeyStaticOuterIPList, ips...)
}

// WithStaticOuterIPV6List filters by contains outer ip v6 list.
func WithStaticOuterIPV6List(ips ...string) OptFn {
	return base.WithValues(FieldKeyStaticOuterIPV6List, ips...)
}

// WithFuzzyStaticOuterIPV6List filters by contains outer ip v6 list.
func WithFuzzyStaticOuterIPV6List(ips ...string) OptFn {
	return base.WithFuzzyValues(FieldKeyStaticOuterIPV6List, ips...)
}

// WithoutFuzzyStaticOuterIPList filters by not contains outer ip list.
func WithoutFuzzyStaticOuterIPList(ips ...string) OptFn {
	return base.WithoutFuzzyValues(FieldKeyStaticOuterIPList, ips...)
}

// WithoutFuzzyStaticOuterIPV6List filters by not contains outer ip v6 list.
func WithoutFuzzyStaticOuterIPV6List(ips ...string) OptFn {
	return base.WithoutFuzzyValues(FieldKeyStaticOuterIPV6List, ips...)
}

// WithoutAgentID filters by not contains agent id.
func WithoutAgentID(agentIDs ...string) OptFn {
	return base.WithoutValues(FieldKeyDynamicAgentID, agentIDs...)
}

// WithoutNodeGeneration filters by not contains node generation.
func WithoutNodeGeneration(generations ...int64) OptFn {
	return base.WithoutValues(FieldKeyDynamicNodeGeneration, generations...)
}

// WithoutStaticAddressing filters by contains addressing.
func WithoutStaticAddressing(addressings ...types.Addressing) OptFn {
	strs := make([]string, len(addressings))
	for idx, addressing := range addressings {
		strs[idx] = string(addressing)
	}

	return base.WithoutValues(FieldKeyStaticAddressing, strs...)
}

// WithoutStaticInnerIPList filters by contains inner ip list.
func WithoutStaticInnerIPList(ips ...string) OptFn {
	return base.WithoutValues(FieldKeyStaticInnerIPList, ips...)
}
