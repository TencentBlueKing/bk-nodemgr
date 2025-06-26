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
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate check body.
func (x *TopoHostListReq) Validate() error {
	return validatePage(x.GetPage())
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
	return convertHostConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions())
}

// ConvertConditionsFromTypes convert types to proto.
func (x *TopoHostListReq) ConvertConditionsFromTypes(condition *types.HostCondition) error {
	exactCond, fuzzyCond, err := convertHostConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond
	x.FuzzyIncludeConditions = fuzzyCond

	return nil
}

// ConvertHostsFromTypes convert types to proto.
func (x *TopoHostListResp) ConvertHostsFromTypes(total int64, hosts []*types.Host) {
	items := make([]*Host, len(hosts))
	for idx, host := range hosts {
		item := newEmptyHost()
		*item.TenantId = host.TenantID
		*item.BkHostId = host.HostID

		*item.Info.BkBizId = host.Static.BizID
		*item.Info.BkNetworkareaId = host.Static.NetworkAreaID
		*item.Info.BkNetworkunitId = host.Dynamic.NetworkUnitID
		*item.Info.BkHostName = host.Static.HostName
		*item.Info.DeptName = host.Static.DeptName
		*item.Info.BkHostInnerip = host.Static.InnerIP
		*item.Info.BkHostInneripV6 = host.Static.InnerIPV6
		*item.Info.BkHostOuterip = host.Static.OuterIP
		*item.Info.BkHostOuteripV6 = host.Static.OuterIPV6
		*item.Info.BkMac = host.Static.Mac
		*item.Info.OsType = host.Static.OSType

		*item.State.NodeRole = string(host.Dynamic.NodeRole)
		*item.State.NodeStatus = string(host.Dynamic.NodeStatus)
		*item.State.NodeVersion = host.Dynamic.NodeVersion
		*item.State.NodeGeneration = int64(host.Dynamic.NodeGeneration)
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
			OSType:        info.GetOsType(),
		}
		host.Dynamic = &types.HostDynamic{
			AgentID:        state.GetBkAgentId(),
			NodeRole:       types.NodeRole(state.GetNodeRole()),
			NodeStatus:     types.NodeStatus(state.GetNodeStatus()),
			NodeVersion:    state.GetNodeVersion(),
			NodeGeneration: types.Generation(state.GetNodeGeneration()),
			NetworkUnitID:  info.GetBkNetworkunitId(),
		}

		result[idx] = host
	}

	return data.GetTotal(), result
}

// Validate validates the request.
func (x *TopoHostDistinctReq) Validate() error {
	return nil
}

// AutoConvert automatically converts the request to types.
func (x *TopoHostDistinctReq) AutoConvert() {
}

// ConvertConditionsToTypes converts the request to types.
func (x *TopoHostDistinctReq) ConvertConditionsToTypes() *types.HostCondition {
	return convertHostConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions())
}

// ConvertConditionsFromTypes converts the request to types.
func (x *TopoHostDistinctReq) ConvertConditionsFromTypes(condition *types.HostCondition) error {
	exactCond, fuzzyCond, err := convertHostConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond
	x.FuzzyIncludeConditions = fuzzyCond

	return nil
}

// ConvertResultFromTypes converts the result from types.
func (x *TopoHostDistinctResp) ConvertResultFromTypes(result *types.HostDistinctResult) {
	if result == nil {
		return
	}

	x.Data = &TopoHostDistinctResp_Data{
		BkBizId:         formatRespSlice(result.BizID),
		NodeRole:        formatRespSlice(types.NodeRoleListToStringList(result.NodeRole)),
		NodeStatus:      formatRespSlice(types.NodeStatusListToStringList(result.NodeStatus)),
		NodeVersion:     formatRespSlice(result.NodeVersion),
		DeptName:        formatRespSlice(result.DeptName),
		OsType:          formatRespSlice(result.OSType),
		Arch:            formatRespSlice(result.Arch),
		Addressing:      formatRespSlice(result.Addressing),
		BkNetworkareaId: formatRespSlice(result.NetworkAreaID),
		BkNetworkunitId: formatRespSlice(result.NetworkUnitID),
	}
}

// ConvertResultToTypes converts the response to types.
func (x *TopoHostDistinctResp) ConvertResultToTypes() *types.HostDistinctResult {
	if x.GetData() == nil {
		return &types.HostDistinctResult{}
	}

	data := x.GetData()

	return &types.HostDistinctResult{
		BizID:         data.GetBkBizId(),
		NodeRole:      types.StringListToNodeRoleList(data.GetNodeRole()),
		NodeStatus:    types.StringListToNodeStatusList(data.GetNodeStatus()),
		NodeVersion:   data.GetNodeVersion(),
		DeptName:      data.GetDeptName(),
		OSType:        data.GetOsType(),
		Arch:          data.GetArch(),
		Addressing:    data.GetAddressing(),
		NetworkAreaID: data.GetBkNetworkareaId(),
		NetworkUnitID: data.GetBkNetworkunitId(),
	}
}

func newEmptyHost() *Host {
	return &Host{
		TenantId: new(string),
		BkHostId: new(int64),
		Info: &HostInfo{
			BkBizId:         new(int64),
			BkNetworkareaId: new(int64),
			BkNetworkunitId: new(int64),
			BkHostName:      new(string),
			DeptName:        new(string),
			BkHostInnerip:   new(string),
			BkHostInneripV6: new(string),
			BkHostOuterip:   new(string),
			BkHostOuteripV6: new(string),
			BkMac:           new(string),
			OsType:          new(string),
		},
		State: &HostState{
			NodeRole:       new(string),
			NodeStatus:     new(string),
			NodeVersion:    new(string),
			NodeGeneration: new(int64),
			BkAgentId:      new(string),
		},
	}
}

func convertHostConditionsToTypes(
	exactCond *TopoHostExactConditions, fuzzyCond *TopoHostFuzzyConditions) *types.HostCondition {

	condition := &types.HostCondition{}

	// exact conditions.
	if exactCond != nil {
		condition.ExactInclude = &types.HostExactFields{
			HostID:         exactCond.GetBkHostId(),
			BizID:          exactCond.GetBkBizId(),
			NetworkAreaID:  exactCond.GetBkNetworkareaId(),
			NetworkUnitID:  exactCond.GetBkNetworkunitId(),
			OSType:         exactCond.GetOsType(),
			NodeRole:       types.StringListToNodeRoleList(exactCond.GetNodeRole()),
			NodeStatus:     types.StringListToNodeStatusList(exactCond.GetNodeStatus()),
			NodeVersion:    exactCond.GetNodeVersion(),
			NodeGeneration: exactCond.GetNodeGeneration(),
			AgentID:        exactCond.GetBkAgentId(),
		}
	}

	// fuzzy conditions.
	if fuzzyCond != nil {
		condition.FuzzyInclude = &types.HostFuzzyFields{
			HostName:  fuzzyCond.GetBkHostName(),
			DeptName:  fuzzyCond.GetDeptName(),
			InnerIP:   fuzzyCond.GetBkHostInnerip(),
			InnerIPV6: fuzzyCond.GetBkHostInneripV6(),
			OuterIP:   fuzzyCond.GetBkHostOuterip(),
			OuterIPV6: fuzzyCond.GetBkHostOuteripV6(),
		}
	}

	return condition
}

func convertHostConditionsFromTypes(
	condition *types.HostCondition) (*TopoHostExactConditions, *TopoHostFuzzyConditions, error) {

	if condition == nil {
		return nil, nil, nil
	}

	var exactCond *TopoHostExactConditions
	var fuzzyCond *TopoHostFuzzyConditions

	if condition.ExactInclude != nil {
		exactCond = &TopoHostExactConditions{
			BkHostId:        condition.ExactInclude.HostID,
			BkBizId:         condition.ExactInclude.BizID,
			BkNetworkareaId: condition.ExactInclude.NetworkAreaID,
			BkNetworkunitId: condition.ExactInclude.NetworkUnitID,
			OsType:          condition.ExactInclude.OSType,
			NodeRole:        types.NodeRoleListToStringList(condition.ExactInclude.NodeRole),
			NodeStatus:      types.NodeStatusListToStringList(condition.ExactInclude.NodeStatus),
			NodeVersion:     condition.ExactInclude.NodeVersion,
			NodeGeneration:  condition.ExactInclude.NodeGeneration,
			BkAgentId:       condition.ExactInclude.AgentID,
		}
	}

	if condition.FuzzyInclude != nil {
		fuzzyCond = &TopoHostFuzzyConditions{
			BkHostName:      condition.FuzzyInclude.HostName,
			DeptName:        condition.FuzzyInclude.DeptName,
			BkHostInnerip:   condition.FuzzyInclude.InnerIP,
			BkHostInneripV6: condition.FuzzyInclude.InnerIPV6,
			BkHostOuterip:   condition.FuzzyInclude.OuterIP,
			BkHostOuteripV6: condition.FuzzyInclude.OuterIPV6,
		}
	}

	if condition.ExactExclude != nil || condition.FuzzyExclude != nil {
		return nil, nil, errors.New("exact-exclude and fuzzy-exclude not supported")
	}

	return exactCond, fuzzyCond, nil
}
