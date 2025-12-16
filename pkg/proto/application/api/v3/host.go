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
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
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
	return convertHostConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions(), nil)
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
		item.Info.BkHostInneripList = host.Static.InnerIPList
		item.Info.BkHostInneripV6List = host.Static.InnerIPV6List
		item.Info.BkHostOuteripList = host.Static.OuterIPList
		item.Info.BkHostOuteripV6List = host.Static.OuterIPV6List
		*item.Info.BkMac = host.Static.Mac
		*item.Info.OsType = host.Static.OSType
		*item.Info.CpuArch = string(host.Dynamic.NodeCPUArch)
		*item.Info.LoginIp = host.Dynamic.LoginIP
		*item.Info.LoginPort = host.Dynamic.LoginPort
		*item.Info.LoginUser = host.Dynamic.LoginUser
		*item.Info.LoginMode = string(host.Dynamic.LoginMode)
		*item.Info.LoginCreditValid = host.Dynamic.LoginCreditValid
		*item.Info.ExportIp = host.Dynamic.ExportIP
		*item.Info.AdvertiseIp = host.Dynamic.AdvertiseIP

		*item.State.NodeRole = string(host.Dynamic.NodeRole)
		*item.State.NodeStatus = string(host.Dynamic.NodeStatus)
		*item.State.NodeVersion = host.Dynamic.NodeVersion
		*item.State.NodeGeneration = int64(host.Dynamic.NodeGeneration)
		*item.State.BkAgentId = host.Dynamic.AgentID
		item.State.ProxyTags = types.ProxyTagListToStringList(host.Dynamic.ProxyTags)

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
			InnerIPList:   info.GetBkHostInneripList(),
			InnerIPV6List: info.GetBkHostInneripV6List(),
			OuterIPList:   info.GetBkHostOuteripList(),
			OuterIPV6List: info.GetBkHostOuteripV6List(),
			Mac:           info.GetBkMac(),
			OSType:        info.GetOsType(),
		}
		host.Dynamic = &types.HostDynamic{
			AgentID:          state.GetBkAgentId(),
			NodeRole:         types.NodeRole(state.GetNodeRole()),
			NodeStatus:       types.NodeStatus(state.GetNodeStatus()),
			NodeVersion:      state.GetNodeVersion(),
			NodeGeneration:   types.Generation(state.GetNodeGeneration()),
			NetworkUnitID:    info.GetBkNetworkunitId(),
			NodeOsType:       criteria.OSType(info.GetOsType()),
			NodeCPUArch:      criteria.CPUArch(info.GetCpuArch()),
			ProxyTags:        types.StringListToProxyTagList(state.GetProxyTags()),
			LoginIP:          info.GetLoginIp(),
			LoginPort:        info.GetLoginPort(),
			LoginUser:        info.GetLoginUser(),
			LoginMode:        types.LoginMode(info.GetLoginMode()),
			LoginCreditValid: info.GetLoginCreditValid(),
			ExportIP:         info.GetExportIp(),
			AdvertiseIP:      info.GetAdvertiseIp(),
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
	return convertHostConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions(), nil)
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

// Validate check body.
func (x *TopoHostSimpleListReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *TopoHostSimpleListReq) AutoConvert() {
}

// ConvertConditionsToTypes convert conditions to types.
func (x *TopoHostSimpleListReq) ConvertConditionsToTypes() *types.HostCondition {
	return convertHostConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions(), x.GetExactExcludeConditions())
}

func (x *TopoHostSimpleListReq) ConvertFieldSelectionToTypes() *types.HostFieldSelection {
	if x.GetFieldSelection() == nil {
		return nil
	}

	return &types.HostFieldSelection{
		EnableFieldBkHostID:        x.GetFieldSelection().GetEnableFieldBkHostId(),
		EnableFieldBkBizID:         x.GetFieldSelection().GetEnableFieldBkBizId(),
		EnableFieldBkHostInnerip:   x.GetFieldSelection().GetEnableFieldBkHostInnerip(),
		EnableFieldBkHostInneripV6: x.GetFieldSelection().GetEnableFieldBkHostInneripV6(),
		EnableFieldBkNetworkunitID: x.GetFieldSelection().GetEnableFieldBkNetworkunitId(),
	}
}

// ConvertHostsFromTypes convert types to proto.
func (x *TopoHostSimpleListResp) ConvertHostSelectedFiledFromTypes(total int64, hosts []*types.HostSelectedFields) {
	if hosts == nil {
		return
	}

	items := make([]*HostSelectedFields, len(hosts))
	for idx, host := range hosts {
		items[idx] = &HostSelectedFields{
			BkHostId:            &host.BkHostID,
			BkBizId:             &host.BkBizID,
			BkHostInneripList:   host.BkHostInnerip,
			BkHostInneripV6List: host.BkHostInneripV6,
			BkNetworkunitId:     &host.BkNetworkunitID,
		}
	}

	x.Data = &TopoHostSimpleListResp_Data{
		Total: total,
		Items: items,
	}
}

func newEmptyHost() *Host {
	return &Host{
		TenantId: new(string),
		BkHostId: new(int64),
		Info: &HostInfo{
			BkBizId:             new(int64),
			BkNetworkareaId:     new(int64),
			BkNetworkareaName:   new(string),
			BkNetworkunitId:     new(int64),
			BkNetworkunitName:   new(string),
			BkHostName:          new(string),
			DeptName:            new(string),
			BkHostInneripList:   make([]string, 0),
			BkHostInneripV6List: make([]string, 0),
			BkHostOuteripList:   make([]string, 0),
			BkHostOuteripV6List: make([]string, 0),
			BkMac:               new(string),
			OsType:              new(string),
			CpuArch:             new(string),
			LoginIp:             new(string),
			LoginPort:           new(int64),
			LoginUser:           new(string),
			LoginMode:           new(string),
			LoginCreditValid:    new(bool),
			ExportIp:            new(string),
			AdvertiseIp:         new(string),
		},
		State: &HostState{
			NodeRole:       new(string),
			NodeStatus:     new(string),
			NodeVersion:    new(string),
			NodeGeneration: new(int64),
			BkAgentId:      new(string),
			ProxyTags:      make([]string, 0),
		},
	}
}

func convertHostConditionsToTypes(
	exactCond *TopoHostExactConditions, fuzzyCond *TopoHostFuzzyConditions, exactExcCond *TopoHostExactConditions) *types.HostCondition {

	condition := &types.HostCondition{}

	// exact conditions.
	if exactCond != nil {
		condition.StaticExactInclude = &types.HostStaticExactFields{
			HostID:        exactCond.GetBkHostId(),
			BizID:         exactCond.GetBkBizId(),
			NetworkAreaID: exactCond.GetBkNetworkareaId(),
		}
		condition.DynamicExactInclude = &types.HostDynamicExactFields{
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
		condition.StaticFuzzyInclude = &types.HostStaticFuzzyFields{
			HostName:  fuzzyCond.GetBkHostName(),
			DeptName:  fuzzyCond.GetDeptName(),
			InnerIP:   fuzzyCond.GetBkHostInnerip(),
			InnerIPV6: fuzzyCond.GetBkHostInneripV6(),
			OuterIP:   fuzzyCond.GetBkHostOuterip(),
			OuterIPV6: fuzzyCond.GetBkHostOuteripV6(),
		}
	}

	if exactExcCond != nil {
		condition.StaticExactExclude = &types.HostStaticExactFields{
			HostID:        exactExcCond.GetBkHostId(),
			BizID:         exactExcCond.GetBkBizId(),
			NetworkAreaID: exactExcCond.GetBkNetworkareaId(),
		}
		condition.DynamicExactExclude = &types.HostDynamicExactFields{
			NetworkUnitID:  exactExcCond.GetBkNetworkunitId(),
			OSType:         exactExcCond.GetOsType(),
			NodeRole:       types.StringListToNodeRoleList(exactExcCond.GetNodeRole()),
			NodeStatus:     types.StringListToNodeStatusList(exactExcCond.GetNodeStatus()),
			NodeVersion:    exactExcCond.GetNodeVersion(),
			NodeGeneration: exactExcCond.GetNodeGeneration(),
			AgentID:        exactExcCond.GetBkAgentId(),
		}
	}

	return condition
}

func convertHostConditionsFromTypes(
	condition *types.HostCondition) (*TopoHostExactConditions, *TopoHostFuzzyConditions, error) {

	if condition == nil {
		return nil, nil, nil
	}

	exactCond := new(TopoHostExactConditions)
	fuzzyCond := new(TopoHostFuzzyConditions)

	if condition.StaticExactInclude != nil {
		exactCond.BkHostId = condition.StaticExactInclude.HostID
		exactCond.BkBizId = condition.StaticExactInclude.BizID
		exactCond.BkNetworkareaId = condition.StaticExactInclude.NetworkAreaID
	}

	if condition.DynamicExactInclude != nil {
		exactCond.BkNetworkunitId = condition.DynamicExactInclude.NetworkUnitID
		exactCond.OsType = condition.DynamicExactInclude.OSType
		exactCond.NodeRole = types.NodeRoleListToStringList(condition.DynamicExactInclude.NodeRole)
		exactCond.NodeStatus = types.NodeStatusListToStringList(condition.DynamicExactInclude.NodeStatus)
		exactCond.NodeVersion = condition.DynamicExactInclude.NodeVersion
		exactCond.NodeGeneration = condition.DynamicExactInclude.NodeGeneration
		exactCond.BkAgentId = condition.DynamicExactInclude.AgentID
	}

	if condition.StaticFuzzyInclude != nil {
		fuzzyCond.BkHostName = condition.StaticFuzzyInclude.HostName
		fuzzyCond.DeptName = condition.StaticFuzzyInclude.DeptName
		fuzzyCond.BkHostInnerip = condition.StaticFuzzyInclude.InnerIP
		fuzzyCond.BkHostInneripV6 = condition.StaticFuzzyInclude.InnerIPV6
		fuzzyCond.BkHostOuterip = condition.StaticFuzzyInclude.OuterIP
		fuzzyCond.BkHostOuteripV6 = condition.StaticFuzzyInclude.OuterIPV6
	}

	return exactCond, fuzzyCond, nil
}
