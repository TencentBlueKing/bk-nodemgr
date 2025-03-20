/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package v3

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate check body.
func (x *TopoHostListReq) Validate() error {
	return validateTopoPage(x.GetPage())
}

// AutoConvert auto convert.
func (x *TopoHostListReq) AutoConvert() {
}

// ConvertPageToTypes convert page to types.
func (x *TopoHostListReq) ConvertPageToTypes(maxLimit int) types.Page {
	return generatePage(x.GetPage(), maxLimit)
}

// ConvertConditionsToTypes convert conditions to types.
func (x *TopoHostListReq) ConvertConditionsToTypes() *types.HostCondition {
	// exact conditions.
	if exactCond := x.GetExactIncludeConditions(); exactCond != nil {
		conditions := &types.HostCondition{
			Type: types.ConditionTypeExactInclude,
		}
		conditions.Exact = &types.HostExactFields{
			HostID:        exactCond.GetBkHostId(),
			BizID:         exactCond.GetBkBizId(),
			NetworkAreaID: exactCond.GetBkNetworkareaId(),
			NetworkUnitID: exactCond.GetBkNetworkunitId(),
			OSType:        exactCond.GetBkOsType(),
			NodeRole: func(source []string) []types.NodeRole {
				target := make([]types.NodeRole, len(source))
				for idx, s := range source {
					target[idx] = types.NodeRole(s)
				}

				return target
			}(exactCond.GetNodeRole()),
			NodeStatus: func(source []string) []types.NodeStatus {
				target := make([]types.NodeStatus, len(source))
				for idx, s := range source {
					target[idx] = types.NodeStatus(s)
				}

				return target
			}(exactCond.GetNodeStatus()),
			NodeVersion: exactCond.GetNodeVersion(),
			AgentID:     exactCond.GetBkAgentId(),
		}

		return conditions
	}

	// fuzzy conditions.
	if fuzzyCond := x.GetFuzzyIncludeConditions(); fuzzyCond != nil {
		conditions := &types.HostCondition{
			Type: types.ConditionTypeFuzzyInclude,
		}
		conditions.Fuzzy = &types.HostFuzzyFields{
			HostName:  fuzzyCond.GetBkHostName(),
			DeptName:  fuzzyCond.GetDeptName(),
			InnerIP:   fuzzyCond.GetBkHostInnerip(),
			InnerIPV6: fuzzyCond.GetBkHostInneripV6(),
			OuterIP:   fuzzyCond.GetBkHostOuterip(),
			OuterIPV6: fuzzyCond.GetBkHostOuteripV6(),
		}

		return conditions
	}

	// default empty conditions.
	return &types.HostCondition{
		Type: types.ConditionTypeExactInclude,
	}
}

// ConvertConditionsFromTypes convert types to proto.
func (x *TopoHostListReq) ConvertConditionsFromTypes(condition *types.HostCondition) error {
	if condition == nil {
		return nil
	}

	switch condition.Type {
	case types.ConditionTypeExactInclude:
		if condition.Exact != nil {
			x.ExactIncludeConditions = &TopoHostListReq_ExactConditions{
				BkHostId:        condition.Exact.HostID,
				BkBizId:         condition.Exact.BizID,
				BkNetworkareaId: condition.Exact.NetworkAreaID,
				BkNetworkunitId: condition.Exact.NetworkUnitID,
				BkOsType:        condition.Exact.OSType,
				NodeRole:        types.NodeRoleListToStringList(condition.Exact.NodeRole),
				NodeStatus:      types.NodeStatusListToStringList(condition.Exact.NodeStatus),
				NodeVersion:     condition.Exact.NodeVersion,
				BkAgentId:       condition.Exact.AgentID,
			}
		}

	case types.ConditionTypeFuzzyInclude:
		if condition.Fuzzy != nil {
			x.FuzzyIncludeConditions = &TopoHostListReq_FuzzyConditions{
				BkHostName:      condition.Fuzzy.HostName,
				DeptName:        condition.Fuzzy.DeptName,
				BkHostInnerip:   condition.Fuzzy.InnerIP,
				BkHostInneripV6: condition.Fuzzy.InnerIPV6,
				BkHostOuterip:   condition.Fuzzy.OuterIP,
				BkHostOuteripV6: condition.Fuzzy.OuterIPV6,
			}
		}

	default:
		return fmt.Errorf("unknown condition type: %s", condition.Type)
	}

	return nil
}

// ConvertHostsFromTypes convert types to proto.
func (x *TopoHostListResp) ConvertHostsFromTypes(total int64, hosts []*types.Host, mapping types.TopoNameMapping) {
	items := make([]*Host, len(hosts))
	for idx, host := range hosts {
		item := newEmptyHost()
		*item.TenantId = host.TenantID
		*item.BkHostId = host.HostID

		*item.Info.BkBizId = host.Static.BizID
		*item.Info.BkNetworkareaId = host.Static.NetworkAreaID
		*item.Info.BkNetworkareaName = mapping.GetNetworkAreaName(host.Static.NetworkAreaID)
		*item.Info.BkNetworkunitId = host.Dynamic.NetworkUnitID
		*item.Info.BkNetworkunitName = mapping.GetNetworkUnitName(host.Dynamic.NetworkUnitID)
		*item.Info.BkHostName = host.Static.HostName
		*item.Info.DeptName = host.Static.DeptName
		*item.Info.BkHostInnerip = host.Static.InnerIP
		*item.Info.BkHostInneripV6 = host.Static.InnerIPV6
		*item.Info.BkHostOuterip = host.Static.OuterIP
		*item.Info.BkHostOuteripV6 = host.Static.OuterIPV6
		*item.Info.BkMac = host.Static.Mac
		*item.Info.BkOsType = host.Static.OSType
		*item.Info.BkOsTypeName = mapping.GetOsTypeName(host.Static.OSType)

		*item.State.NodeRole = string(host.Dynamic.NodeRole)
		*item.State.NodeStatus = string(host.Dynamic.NodeStatus)
		*item.State.NodeVersion = host.Dynamic.NodeVersion
		*item.State.BkAgentId = host.Dynamic.AgentID

		items[idx] = item
	}

	x.Data = &TopoHostListResp_Data{
		Total: total,
		Items: items,
	}
}

// ConvertHostsToTypes convert proto to types.
func (x *TopoHostListResp) ConvertHostsToTypes() (int64, []*types.Host) {
	data := x.GetData()
	if data == nil {
		return 0, nil
	}

	items := data.GetItems()
	result := make([]*types.Host, len(items))
	for idx, item := range items {
		host := &types.Host{
			TenantID: item.GetTenantId(),
			HostID:   item.GetBkHostId(),
			Static:   &types.HostStatic{},
			Dynamic:  &types.HostDynamic{},
		}

		info := item.GetInfo()
		state := item.GetState()
		if info == nil || state == nil {
			continue
		}
		host.Static = &types.HostStatic{
			BizID:         info.GetBkBizId(),
			NetworkAreaID: info.GetBkNetworkareaId(),
			HostName:      info.GetBkHostName(),
			DeptName:      info.GetDeptName(),
			InnerIP:       info.GetBkHostInnerip(),
			InnerIPV6:     info.GetBkHostInneripV6(),
			OuterIP:       info.GetBkHostOuterip(),
			OuterIPV6:     info.GetBkHostOuteripV6(),
			Mac:           info.GetBkMac(),
			OSType:        info.GetBkOsType(),
		}
		host.Dynamic = &types.HostDynamic{
			AgentID:       state.GetBkAgentId(),
			NodeRole:      types.NodeRole(state.GetNodeRole()),
			NodeStatus:    types.NodeStatus(state.GetNodeStatus()),
			NodeVersion:   state.GetNodeVersion(),
			NetworkUnitID: info.GetBkNetworkunitId(),
		}

		result[idx] = host
	}

	return data.GetTotal(), result
}

func newEmptyHost() *Host {
	return &Host{
		TenantId: new(string),
		BkHostId: new(int64),
		Info: &HostInfo{
			BkBizId:           new(int64),
			BkNetworkareaId:   new(int64),
			BkNetworkareaName: new(string),
			BkNetworkunitId:   new(int64),
			BkNetworkunitName: new(string),
			BkHostName:        new(string),
			DeptName:          new(string),
			BkHostInnerip:     new(string),
			BkHostInneripV6:   new(string),
			BkHostOuterip:     new(string),
			BkHostOuteripV6:   new(string),
			BkMac:             new(string),
			BkOsType:          new(string),
			BkOsTypeName:      new(string),
		},
		State: &HostState{
			NodeRole:    new(string),
			NodeStatus:  new(string),
			NodeVersion: new(string),
			BkAgentId:   new(string),
		},
	}
}
