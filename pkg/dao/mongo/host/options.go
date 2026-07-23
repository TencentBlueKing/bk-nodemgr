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
	"go.mongodb.org/mongo-driver/bson"
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

// WithStaticBizID filters by biz-id.
func WithStaticBizID(bizIDs ...int64) OptFn {
	return base.WithValues(FieldKeyStaticBizID, bizIDs...)
}

// WithStaticSetID filters by topo set id.
func WithStaticSetID(setIDs ...int64) OptFn {
	return base.WithValues(FieldKeyStaticSetID, setIDs...)
}

// WithStaticModuleID filters by topo module id.
func WithStaticModuleID(moduleIDs ...int64) OptFn {
	return base.WithValues(FieldKeyStaticModuleID, moduleIDs...)
}

// WithStaticTopo filters by set and module ids within the same topo item.
func WithStaticTopo(topo ...types.HostTopo) OptFn {
	return func(f bson.D) bson.D {
		if len(topo) == 0 {
			return f
		}

		conditions := make(bson.A, 0, len(topo))
		for _, item := range topo {
			condition := base.WithElemMatch(
				FieldKeyStaticTopo,
				base.WithValues(FieldSubKeyStaticTopoItemSetID, item.SetID),
				base.WithValues(FieldSubKeyStaticTopoItemModuleID, item.ModuleID),
			)(bson.D{})
			conditions = append(conditions, condition)
		}

		return append(f, bson.E{Key: "$or", Value: conditions})
	}
}

// WithoutStaticBizID filters by not contains biz-id.
func WithoutStaticBizID(bizIDs ...int64) OptFn {
	return base.WithoutValues(FieldKeyStaticBizID, bizIDs...)
}

// WithStaticNetworkAreaID filters by network area id.
func WithStaticNetworkAreaID(networkAreaIDs ...int64) OptFn {
	return base.WithValues(FieldKeyStaticNetworkAreaID, networkAreaIDs...)
}

// WithoutStaticNetworkAreaID filters by not contains network area id.
func WithoutStaticNetworkAreaID(networkAreaIDs ...int64) OptFn {
	return base.WithoutValues(FieldKeyStaticNetworkAreaID, networkAreaIDs...)
}

// WithDynamicNetworkUnitID filters by network unit id.
func WithDynamicNetworkUnitID(networkUnitID ...int64) OptFn {
	return base.WithValues(FieldKeyDynamicNetworkUnitID, networkUnitID...)
}

// WithoutDynamicNetworkUnitID filters by not contains network unit id.
func WithoutDynamicNetworkUnitID(networkUnitID ...int64) OptFn {
	return base.WithoutValues(FieldKeyDynamicNetworkUnitID, networkUnitID...)
}

// WithFuzzyStaticHostName filters by host name.
func WithFuzzyStaticHostName(hostNames ...string) OptFn {
	return base.WithFuzzyValues(FieldKeyStaticHostName, hostNames...)
}

// WithoutFuzzyStaticHostName filters by not contains host name.
func WithoutFuzzyStaticHostName(hostNames ...string) OptFn {
	return base.WithoutFuzzyValues(FieldKeyStaticHostName, hostNames...)
}

// WithFuzzyStaticDeptName filters by dept name.
func WithFuzzyStaticDeptName(deptNames ...string) OptFn {
	return base.WithFuzzyValues(FieldKeyStaticDeptName, deptNames...)
}

// WithoutFuzzyStaticDeptName filters by not contains dept name.
func WithoutFuzzyStaticDeptName(deptNames ...string) OptFn {
	return base.WithoutFuzzyValues(FieldKeyStaticDeptName, deptNames...)
}

// WithStaticOSType filters by os type.
func WithStaticOSType(osTypes ...string) OptFn {
	return base.WithValues(FieldKeyStaticOSType, osTypes...)
}

// WithoutStaticOSType filters by not contains os type.
func WithoutStaticOSType(osTypes ...string) OptFn {
	return base.WithoutValues(FieldKeyStaticOSType, osTypes...)
}

// WithStaticAddressing filters by contains addressing.
func WithStaticAddressing(addressings ...types.Addressing) OptFn {
	strs := make([]string, len(addressings))
	for idx, addressing := range addressings {
		strs[idx] = string(addressing)
	}

	return base.WithValues(FieldKeyStaticAddressing, strs...)
}

// WithoutStaticAddressing filters by contains addressing.
func WithoutStaticAddressing(addressings ...types.Addressing) OptFn {
	strs := make([]string, len(addressings))
	for idx, addressing := range addressings {
		strs[idx] = string(addressing)
	}

	return base.WithoutValues(FieldKeyStaticAddressing, strs...)
}

// WithDynamicNodeOsType filters by os type.
func WithDynamicNodeOsType(osTypes ...string) OptFn {
	return base.WithValues(FieldKeyDynamicNodeOsType, osTypes...)
}

// WithoutDynamicNodeOsType filters by not contains os type.
func WithoutDynamicNodeOsType(osTypes ...string) OptFn {
	return base.WithoutValues(FieldKeyDynamicNodeOsType, osTypes...)
}

// WithDynamicCPUArch filters by arch.
func WithDynamicCPUArch(archs ...string) OptFn {
	return base.WithValues(FieldKeyDynamicNodeCPUArch, archs...)
}

// WithoutDynamicCPUArch filters by not contains arch.
func WithoutDynamicCPUArch(archs ...string) OptFn {
	return base.WithoutValues(FieldKeyDynamicNodeCPUArch, archs...)
}

// WithDynamicNodeRole filters by node role.
func WithDynamicNodeRole(roles ...types.NodeRole) OptFn {
	str := make([]string, len(roles))
	for idx, role := range roles {
		str[idx] = string(role)
	}

	return base.WithValues(FieldKeyDynamicNodeRole, str...)
}

// WithoutDynamicNodeRole filters by not contains node role.
func WithoutDynamicNodeRole(roles ...types.NodeRole) OptFn {
	str := make([]string, len(roles))
	for idx, role := range roles {
		str[idx] = string(role)
	}

	return base.WithoutValues(FieldKeyDynamicNodeRole, str...)
}

// WithDynamicNodeStatus filters by node status.
func WithDynamicNodeStatus(statuses ...types.NodeStatus) OptFn {
	str := make([]string, len(statuses))
	for idx, status := range statuses {
		str[idx] = string(status)
	}

	return base.WithValues(FieldKeyDynamicNodeStatus, str...)
}

// WithoutDynamicNodeStatus filters by not contains node status.
func WithoutDynamicNodeStatus(statuses ...types.NodeStatus) OptFn {
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

// WithoutDynamicProxyTags filters by not contains proxy tag.
func WithoutDynamicProxyTags(tags ...types.ProxyTag) OptFn {
	str := make([]string, len(tags))
	for idx, tag := range tags {
		str[idx] = string(tag)
	}

	return base.WithoutValues(FieldKeyDynamicProxyTags, str...)
}

// WithDynamicProxyAccessDisabled filters by not contains proxy access disabled.
func WithDynamicProxyAccessDisabled(bools ...bool) OptFn {
	return base.WithValues(FieldKeyDynamicProxyAccessDisabled, bools...)
}

// WithDynamicNodeVersion filters by node version.
func WithDynamicNodeVersion(versions ...string) OptFn {
	return base.WithValues(FieldKeyDynamicNodeVersion, versions...)
}

// WithoutDynamicNodeVersion filters by not contains node version.
func WithoutDynamicNodeVersion(versions ...string) OptFn {
	return base.WithoutValues(FieldKeyDynamicNodeVersion, versions...)
}

// WithDynamicNodeGeneration filters by node generation.
func WithDynamicNodeGeneration(generations ...int64) OptFn {
	return base.WithValues(FieldKeyDynamicNodeGeneration, generations...)
}

// WithoutDynamicNodeGeneration filters by not contains node generation.
func WithoutDynamicNodeGeneration(generations ...int64) OptFn {
	return base.WithoutValues(FieldKeyDynamicNodeGeneration, generations...)
}

// WithDynamicAgentID filters by contains agent id.
func WithDynamicAgentID(agentIDs ...string) OptFn {
	return base.WithValues(FieldKeyDynamicAgentID, agentIDs...)
}

// WithoutDynamicAgentID filters by not contains agent id.
func WithoutDynamicAgentID(agentIDs ...string) OptFn {
	return base.WithoutValues(FieldKeyDynamicAgentID, agentIDs...)
}

// WithDynamicAgentIDNotEmpty filters by non-empty agent id.
func WithDynamicAgentIDNotEmpty() OptFn {
	return base.WithGreaterThanValue(FieldKeyDynamicAgentID, "")
}

// WithStaticInnerIPList filters by contains inner ip list.
func WithStaticInnerIPList(ips ...string) OptFn {
	return base.WithValues(FieldKeyStaticInnerIPList, ips...)
}

// WithoutFuzzyStaticInnerIPList filters by not contains inner ip list.
func WithoutFuzzyStaticInnerIPList(ips ...string) OptFn {
	return base.WithoutFuzzyValues(FieldKeyStaticInnerIPList, ips...)
}

// WithFuzzyStaticInnerIPList filters by contains inner ip.
func WithFuzzyStaticInnerIPList(ips ...string) OptFn {
	return base.WithFuzzyValues(FieldKeyStaticInnerIPList, ips...)
}

// WithoutFuzzyStaticInnerIPV6List filters by not contains inner ip v6 list.
func WithoutFuzzyStaticInnerIPV6List(ips ...string) OptFn {
	return base.WithoutFuzzyValues(FieldKeyStaticInnerIPV6List, ips...)
}

// WithoutStaticInnerIPList filters by contains inner ip list.
func WithoutStaticInnerIPList(ips ...string) OptFn {
	return base.WithoutValues(FieldKeyStaticInnerIPList, ips...)
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

// WithoutFuzzyStaticOuterIPList filters by not contains outer ip list.
func WithoutFuzzyStaticOuterIPList(ips ...string) OptFn {
	return base.WithoutFuzzyValues(FieldKeyStaticOuterIPList, ips...)
}

// WithStaticOuterIPV6List filters by contains outer ip v6 list.
func WithStaticOuterIPV6List(ips ...string) OptFn {
	return base.WithValues(FieldKeyStaticOuterIPV6List, ips...)
}

// WithFuzzyStaticOuterIPV6List filters by contains outer ip v6 list.
func WithFuzzyStaticOuterIPV6List(ips ...string) OptFn {
	return base.WithFuzzyValues(FieldKeyStaticOuterIPV6List, ips...)
}

// WithoutFuzzyStaticOuterIPV6List filters by not contains outer ip v6 list.
func WithoutFuzzyStaticOuterIPV6List(ips ...string) OptFn {
	return base.WithoutFuzzyValues(FieldKeyStaticOuterIPV6List, ips...)
}
