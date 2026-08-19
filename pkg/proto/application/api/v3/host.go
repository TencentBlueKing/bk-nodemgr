/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
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
func (x *TopoHostListReq) ConvertPageToTypes() (types.Page, error) {
	return convPageToTypes(x.GetPage())
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
		*item.Info.ExportIpV6 = host.Dynamic.ExportIPV6
		*item.Info.AdvertiseIp = host.Dynamic.AdvertiseIP
		*item.Info.AdvertiseIpV6 = host.Dynamic.AdvertiseIPV6
		*item.Info.RelayCallbackPort = host.Dynamic.RelayCallbackPort
		*item.Info.RelayDownloadPort = host.Dynamic.RelayDownloadPort
		*item.Info.BkAddressing = string(host.Static.Addressing)

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
			Addressing:    types.Addressing(info.GetBkAddressing()),
		}
		host.Dynamic = &types.HostDynamic{
			AgentID:           state.GetBkAgentId(),
			NodeRole:          types.NodeRole(state.GetNodeRole()),
			NodeStatus:        types.NodeStatus(state.GetNodeStatus()),
			NodeVersion:       state.GetNodeVersion(),
			NodeGeneration:    types.Generation(state.GetNodeGeneration()),
			NetworkUnitID:     info.GetBkNetworkunitId(),
			NodeOsType:        criteria.OSType(info.GetOsType()),
			NodeCPUArch:       criteria.CPUArch(info.GetCpuArch()),
			ProxyTags:         types.StringListToProxyTagList(state.GetProxyTags()),
			LoginIP:           info.GetLoginIp(),
			LoginPort:         info.GetLoginPort(),
			LoginUser:         info.GetLoginUser(),
			LoginMode:         types.LoginMode(info.GetLoginMode()),
			LoginCreditValid:  info.GetLoginCreditValid(),
			ExportIP:          info.GetExportIp(),
			ExportIPV6:        info.GetExportIpV6(),
			AdvertiseIP:       info.GetAdvertiseIp(),
			AdvertiseIPV6:     info.GetAdvertiseIpV6(),
			RelayCallbackPort: info.GetRelayCallbackPort(),
			RelayDownloadPort: info.GetRelayDownloadPort(),
		}

		result[idx] = host
	}

	return data.GetTotal(), result
}

// Validate check body.
func (x *TopoHostSelectHostIDReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *TopoHostSelectHostIDReq) AutoConvert() {
}

// ConvertConditionsToTypes convert conditions to types.
func (x *TopoHostSelectHostIDReq) ConvertConditionsToTypes() *types.HostCondition {
	return convertHostConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions(), x.GetExactExcludeConditions())
}

// ConvertHostID converts host IDs to the proto response.
func (x *TopoHostSelectHostIDResp) ConvertHostID(hosts []int64) {
	x.Data = &TopoHostSelectHostIDResp_Data{
		Items: hosts,
	}
}

// Validate validates the request.
func (x *TopoHostDistinctReq) Validate() error {
	return nil
}

// AutoConvert automatically converts the request to types.
func (x *TopoHostDistinctReq) AutoConvert() {
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
			ExportIpV6:          new(string),
			AdvertiseIp:         new(string),
			AdvertiseIpV6:       new(string),
			RelayCallbackPort:   new(int64),
			RelayDownloadPort:   new(int64),
			BkAddressing:        new(string),
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
			SetID:         exactCond.GetBkSetId(),
			ModuleID:      exactCond.GetBkModuleId(),
			NetworkAreaID: exactCond.GetBkNetworkareaId(),
			InnerIP:       exactCond.GetBkHostInnerip(),
			InnerIPV6:     exactCond.GetBkHostInneripV6(),
		}
		condition.DynamicExactInclude = &types.HostDynamicExactFields{
			NetworkUnitID:  exactCond.GetBkNetworkunitId(),
			OSType:         exactCond.GetOsType(),
			NodeRole:       types.StringListToNodeRoleList(exactCond.GetNodeRole()),
			NodeStatus:     types.StringListToNodeStatusList(exactCond.GetNodeStatus()),
			NodeVersion:    exactCond.GetNodeVersion(),
			NodeGeneration: exactCond.GetNodeGeneration(),
			AgentID:        exactCond.GetBkAgentId(),
			ProxyTags:      types.StringListToProxyTagList(exactCond.GetProxyTags()),
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
			SetID:         exactExcCond.GetBkSetId(),
			ModuleID:      exactExcCond.GetBkModuleId(),
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
			ProxyTags:      types.StringListToProxyTagList(exactExcCond.GetProxyTags()),
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
		exactCond.BkSetId = condition.StaticExactInclude.SetID
		exactCond.BkModuleId = condition.StaticExactInclude.ModuleID
		exactCond.BkNetworkareaId = condition.StaticExactInclude.NetworkAreaID
		exactCond.BkHostInnerip = condition.StaticExactInclude.InnerIP
		exactCond.BkHostInneripV6 = condition.StaticExactInclude.InnerIPV6
	}

	if condition.DynamicExactInclude != nil {
		exactCond.BkNetworkunitId = condition.DynamicExactInclude.NetworkUnitID
		exactCond.OsType = condition.DynamicExactInclude.OSType
		exactCond.NodeRole = types.NodeRoleListToStringList(condition.DynamicExactInclude.NodeRole)
		exactCond.NodeStatus = types.NodeStatusListToStringList(condition.DynamicExactInclude.NodeStatus)
		exactCond.NodeVersion = condition.DynamicExactInclude.NodeVersion
		exactCond.NodeGeneration = condition.DynamicExactInclude.NodeGeneration
		exactCond.BkAgentId = condition.DynamicExactInclude.AgentID
		exactCond.ProxyTags = types.ProxyTagListToStringList(condition.DynamicExactInclude.ProxyTags)
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

// Validate check body.
func (x *TopoHostSelectInnerIPReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *TopoHostSelectInnerIPReq) AutoConvert() {
}

// ConvertConditionsToTypes convert conditions to types.
func (x *TopoHostSelectInnerIPReq) ConvertConditionsToTypes() *types.HostCondition {
	return convertHostConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions(), x.GetExactExcludeConditions())
}

// ConvertInnerIP convert types to proto.
func (x *TopoHostSelectInnerIPResp) ConvertInnerIP(items []string) {
	x.Data = &TopoHostSelectInnerIPResp_Data{
		Items: items,
	}
}

// Validate check body.
func (x *TopoHostSelectInnerIPV6Req) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *TopoHostSelectInnerIPV6Req) AutoConvert() {
}

// ConvertConditionsToTypes convert conditions to types.
func (x *TopoHostSelectInnerIPV6Req) ConvertConditionsToTypes() *types.HostCondition {
	return convertHostConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions(), x.GetExactExcludeConditions())
}

// ConvertInnerIPV6 convert types to proto.
func (x *TopoHostSelectInnerIPV6Resp) ConvertInnerIPV6(items []string) {
	x.Data = &TopoHostSelectInnerIPV6Resp_Data{
		Items: items,
	}
}

// Validate check body.
func (x *TopoHostSelectNetWorkareaIDAndInnerIPReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *TopoHostSelectNetWorkareaIDAndInnerIPReq) AutoConvert() {
}

// ConvertConditionsToTypes convert conditions to types.
func (x *TopoHostSelectNetWorkareaIDAndInnerIPReq) ConvertConditionsToTypes() *types.HostCondition {
	return convertHostConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions(), x.GetExactExcludeConditions())
}

// ConvertNetWorkareaIDAndInnerIP convert types to proto.
func (x *TopoHostSelectNetWorkareaIDAndInnerIPResp) ConvertNetWorkareaIDAndInnerIP(items []string) {
	x.Data = &TopoHostSelectNetWorkareaIDAndInnerIPResp_Data{
		Items: items,
	}
}

// Validate check body.
func (x *TopoHostSelectNetWorkareaIDAndInnerIPV6Req) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *TopoHostSelectNetWorkareaIDAndInnerIPV6Req) AutoConvert() {
}

// ConvertConditionsToTypes convert conditions to types.
func (x *TopoHostSelectNetWorkareaIDAndInnerIPV6Req) ConvertConditionsToTypes() *types.HostCondition {
	return convertHostConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions(), x.GetExactExcludeConditions())
}

// ConvertNetWorkareaIDAndInnerIPV6 convert types to proto.
func (x *TopoHostSelectNetWorkareaIDAndInnerIPV6Resp) ConvertNetWorkareaIDAndInnerIPV6(items []string) {
	x.Data = &TopoHostSelectNetWorkareaIDAndInnerIPV6Resp_Data{
		Items: items,
	}
}
