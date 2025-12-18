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
	"strings"

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
	exactIncludeCond, fuzzyIncludeCond, _, err := convertHostConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactIncludeCond
	x.FuzzyIncludeConditions = fuzzyIncludeCond

	return nil
}

// ConvertHostsFromTypes convert types to proto.
func (x *TopoHostListResp) ConvertHostsFromTypes(total int64, hosts []*types.Host, hostCredits map[string]bool) {
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
		item.Info.BkHostInneripList = host.Static.InnerIPList
		item.Info.BkHostInneripV6List = host.Static.InnerIPV6List
		item.Info.BkHostOuteripList = host.Static.OuterIPList
		item.Info.BkHostOuteripV6List = host.Static.OuterIPV6List
		*item.Info.BkMac = host.Static.Mac
		*item.Info.OsType = host.Static.OSType
		if host.Dynamic.NodeOsType.Validate() != nil && host.Dynamic.NodeOsType != criteria.OSUnknown {
			*item.Info.OsType = string(host.Dynamic.NodeOsType)
		}
		*item.Info.CpuArch = string(host.Dynamic.NodeCPUArch)
		*item.Info.LoginIp = host.Dynamic.LoginIP
		*item.Info.LoginPort = host.Dynamic.LoginPort
		*item.Info.LoginUser = host.Dynamic.LoginUser
		*item.Info.LoginMode = string(host.Dynamic.LoginMode)
		*item.Info.ExportIp = host.Dynamic.ExportIP
		*item.Info.AdvertiseIp = host.Dynamic.AdvertiseIP
		*item.Info.LoginCreditValid = false
		if hostCredits != nil {
			if valid, ok := hostCredits[host.Dynamic.LoginCreditID]; ok && valid {
				*item.Info.LoginCreditValid = true
			}
		}

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
func (x *TopoHostSelectHostIDReq) Validate() error {
	return nil
}

// AutoConvert automatically converts the request to types.
func (x *TopoHostSelectHostIDReq) AutoConvert() {
}

// ConvertConditionsToTypes convert conditions to types for simple list.
func (x *TopoHostSelectHostIDReq) ConvertConditionsToTypes() *types.HostCondition {
	return convertHostConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions(), x.GetExactExcludeConditions())
}

// ConvertConditionsFromTypes convert types to proto for simple list.
func (x *TopoHostSelectHostIDReq) ConvertConditionsFromTypes(condition *types.HostCondition) error {
	exactCond, fuzzyCond, exactExcCond, err := convertHostConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond
	x.FuzzyIncludeConditions = fuzzyCond
	x.ExactExcludeConditions = exactExcCond

	return nil
}

// ConvertHostIDFromTypes convert types to proto for select host id list.
func (x *TopoHostSelectHostIDResp) ConvertHostIDFromTypes(hosts []*types.Host) {
	if hosts == nil {
		return
	}
	items := make([]int64, len(hosts))

	for idx, host := range hosts {
		items[idx] = host.HostID
	}

	x.Data = &TopoHostSelectHostIDResp_Data{
		Items: items,
	}
}

// ConvertHostIDToTypes convert proto to types for select host id list.
func (x *TopoHostSelectHostIDResp) ConvertHostIDToTypes() []int64 {
	data := x.GetData()
	if data == nil {
		return nil
	}

	return data.GetItems()
}

// Validate validates the request.
func (x *TopoHostSelectInnerIPReq) Validate() error {
	return nil
}

// AutoConvert automatically converts the request to types.
func (x *TopoHostSelectInnerIPReq) AutoConvert() {
}

// ConvertConditionsToTypes convert conditions to types.
func (x *TopoHostSelectInnerIPReq) ConvertConditionsToTypes() *types.HostCondition {
	return convertHostConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions(), x.GetExactExcludeConditions())
}

// ConvertConditionsFromTypes convert types to proto.
func (x *TopoHostSelectInnerIPReq) ConvertConditionsFromTypes(condition *types.HostCondition) error {
	exactCond, fuzzyCond, exactExcCond, err := convertHostConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond
	x.FuzzyIncludeConditions = fuzzyCond
	x.ExactExcludeConditions = exactExcCond

	return nil
}

// ConvertInnerIPFromTypes convert types to proto for select inner ip list.
func (x *TopoHostSelectInnerIPResp) ConvertInnerIPFromTypes(hosts []*types.Host) {
	if hosts == nil {
		return
	}

	items := make([]string, len(hosts))
	for idx, host := range hosts {
		items[idx] = strings.Join(host.Static.InnerIPList, types.IpSeparator)
	}

	x.Data = &TopoHostSelectInnerIPResp_Data{
		Items: items,
	}
}

// ConvertInnerIPToTypes convert proto to types for select inner ip list.
func (x *TopoHostSelectInnerIPResp) ConvertInnerIPToTypes() []string {
	data := x.GetData()
	if data == nil {
		return nil
	}

	return data.GetItems()
}

// Validate validates the request.
func (x *TopoHostSelectInnerIPV6Req) Validate() error {
	return nil
}

// AutoConvert automatically converts the request to types.
func (x *TopoHostSelectInnerIPV6Req) AutoConvert() {
}

// ConvertConditionsToTypes convert conditions to types.
func (x *TopoHostSelectInnerIPV6Req) ConvertConditionsToTypes() *types.HostCondition {
	return convertHostConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions(), x.GetExactExcludeConditions())
}

// ConvertConditionsFromTypes convert types to proto.
func (x *TopoHostSelectInnerIPV6Req) ConvertConditionsFromTypes(condition *types.HostCondition) error {
	exactCond, fuzzyCond, exactExcCond, err := convertHostConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond
	x.FuzzyIncludeConditions = fuzzyCond
	x.ExactExcludeConditions = exactExcCond

	return nil
}

// ConvertInnerIPV6FromTypes convert types to proto for select inner ipv6 list.
func (x *TopoHostSelectInnerIPV6Resp) ConvertInnerIPV6FromTypes(hosts []*types.Host) {
	if hosts == nil {
		return
	}

	items := make([]string, len(hosts))
	for idx, host := range hosts {
		items[idx] = strings.Join(host.Static.InnerIPV6List, types.IpSeparator)
	}

	x.Data = &TopoHostSelectInnerIPV6Resp_Data{
		Items: items,
	}
}

// ConvertInnerIPV6ToTypes convert proto to types for select inner ipv6 list.
func (x *TopoHostSelectInnerIPV6Resp) ConvertInnerIPV6ToTypes() []string {
	data := x.GetData()
	if data == nil {
		return nil
	}

	return data.GetItems()
}

// Validate validates the request.
func (x *TopoHostSelectNetWorkareaIDAndInnerIPReq) Validate() error {
	return nil
}

// AutoConvert automatically converts the request to types.
func (x *TopoHostSelectNetWorkareaIDAndInnerIPReq) AutoConvert() {
}

// ConvertConditionsToTypes convert conditions to types.
func (x *TopoHostSelectNetWorkareaIDAndInnerIPReq) ConvertConditionsToTypes() *types.HostCondition {
	return convertHostConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions(), x.GetExactExcludeConditions())
}

// ConvertConditionsFromTypes convert types to proto.
func (x *TopoHostSelectNetWorkareaIDAndInnerIPReq) ConvertConditionsFromTypes(condition *types.HostCondition) error {
	exactCond, fuzzyCond, exactExcCond, err := convertHostConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond
	x.FuzzyIncludeConditions = fuzzyCond
	x.ExactExcludeConditions = exactExcCond

	return nil
}

// ConvertNetWorkareaIDAndInnerIPFromTypes convert types to proto for select networkarea id and inner ip list.
func (x *TopoHostSelectNetWorkareaIDAndInnerIPResp) ConvertNetWorkareaIDAndInnerIPFromTypes(hosts []*types.Host) {
	if hosts == nil {
		return
	}

	items := make([]string, len(hosts))
	for idx, host := range hosts {
		ips := strings.Join(host.Static.InnerIPList, types.IpSeparator)
		items[idx] = fmt.Sprintf("%d:%s", host.Static.NetworkAreaID, ips)
	}

	x.Data = &TopoHostSelectNetWorkareaIDAndInnerIPResp_Data{
		Items: items,
	}
}

// ConvertNetWorkareaIDAndInnerIPToTypes convert proto to types for select networkarea id and inner ip list.
func (x *TopoHostSelectNetWorkareaIDAndInnerIPResp) ConvertNetWorkareaIDAndInnerIPToTypes() []string {
	data := x.GetData()
	if data == nil {
		return nil
	}

	return data.GetItems()
}

// Validate validates the request.
func (x *TopoHostSelectNetWorkareaIDAndInnerIPV6Req) Validate() error {
	return nil
}

// AutoConvert automatically converts the request to types.
func (x *TopoHostSelectNetWorkareaIDAndInnerIPV6Req) AutoConvert() {
}

// ConvertConditionsToTypes convert conditions to types.
func (x *TopoHostSelectNetWorkareaIDAndInnerIPV6Req) ConvertConditionsToTypes() *types.HostCondition {
	return convertHostConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions(), x.GetExactExcludeConditions())
}

// ConvertConditionsFromTypes convert types to proto.
func (x *TopoHostSelectNetWorkareaIDAndInnerIPV6Req) ConvertConditionsFromTypes(condition *types.HostCondition) error {
	exactCond, fuzzyCond, exactExcCond, err := convertHostConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactCond
	x.FuzzyIncludeConditions = fuzzyCond
	x.ExactExcludeConditions = exactExcCond

	return nil
}

// ConvertNetWorkareaIDAndInnerIPV6FromTypes convert types to proto for select networkarea id and inner ipv6 list.
func (x *TopoHostSelectNetWorkareaIDAndInnerIPV6Resp) ConvertNetWorkareaIDAndInnerIPV6FromTypes(hosts []*types.Host) {
	if hosts == nil {
		return
	}

	items := make([]string, len(hosts))
	for idx, host := range hosts {
		ips := strings.Join(host.Static.InnerIPV6List, types.IpSeparator)
		items[idx] = fmt.Sprintf("%d:%s", host.Static.NetworkAreaID, ips)
	}

	x.Data = &TopoHostSelectNetWorkareaIDAndInnerIPV6Resp_Data{
		Items: items,
	}
}

// ConvertNetWorkareaIDAndInnerIPV6ToTypes convert proto to types for select networkarea id and inner ipv6 list.
func (x *TopoHostSelectNetWorkareaIDAndInnerIPV6Resp) ConvertNetWorkareaIDAndInnerIPV6ToTypes() []string {
	data := x.GetData()
	if data == nil {
		return nil
	}

	return data.GetItems()
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
	exactIncludeCond, fuzzyIncludeCond, _, err := convertHostConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactIncludeCond
	x.FuzzyIncludeConditions = fuzzyIncludeCond

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

// ConvertResultFromTypes converts the result from types.
func (x *TopoGetHostDistributionByNodeRoleResp) ConvertResultFromTypes(result map[string]int64) {
	if result == nil {
		return
	}

	x.Data = result
}

// ConvertResultToTypes converts the response to types.
func (x *TopoGetHostDistributionByNodeRoleResp) ConvertResultToTypes() map[string]int64 {
	if x.GetData() == nil {
		return map[string]int64{}
	}

	data := x.GetData()

	return data
}

// ConvertResultFromTypes converts the result from types.
func (x *TopoGetHostDistributionByNetworkAreaIDResp) ConvertResultFromTypes(result map[int64]int64) {
	if result == nil {
		return
	}

	x.Data = result
}

// ConvertResultToTypes converts the response to types.
func (x *TopoGetHostDistributionByNetworkAreaIDResp) ConvertResultToTypes() map[int64]int64 {
	if x.GetData() == nil {
		return map[int64]int64{}
	}

	data := x.GetData()

	return data
}

func newEmptyHost() *Host {
	return &Host{
		TenantId: new(string),
		BkHostId: new(int64),
		Info: &HostInfo{
			BkBizId:             new(int64),
			BkNetworkareaId:     new(int64),
			BkNetworkunitId:     new(int64),
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
	exactIncludeCond *TopoHostExactConditions, fuzzyIncludeCond *TopoHostFuzzyConditions, exactExcludeCond *TopoHostExactConditions) *types.HostCondition {

	condition := &types.HostCondition{}

	// exact conditions.
	if exactIncludeCond != nil {
		condition.StaticExactInclude = &types.HostStaticExactFields{
			HostID:        exactIncludeCond.GetBkHostId(),
			BizID:         exactIncludeCond.GetBkBizId(),
			NetworkAreaID: exactIncludeCond.GetBkNetworkareaId(),
		}

		condition.DynamicExactInclude = &types.HostDynamicExactFields{
			NetworkUnitID:  exactIncludeCond.GetBkNetworkunitId(),
			OSType:         exactIncludeCond.GetOsType(),
			Arch:           exactIncludeCond.GetArch(),
			NodeRole:       types.StringListToNodeRoleList(exactIncludeCond.GetNodeRole()),
			NodeStatus:     types.StringListToNodeStatusList(exactIncludeCond.GetNodeStatus()),
			NodeVersion:    exactIncludeCond.GetNodeVersion(),
			NodeGeneration: exactIncludeCond.GetNodeGeneration(),
			AgentID:        exactIncludeCond.GetBkAgentId(),
		}
	}

	// fuzzy conditions.
	if fuzzyIncludeCond != nil {
		condition.StaticFuzzyInclude = &types.HostStaticFuzzyFields{
			HostName:  fuzzyIncludeCond.GetBkHostName(),
			DeptName:  fuzzyIncludeCond.GetDeptName(),
			InnerIP:   fuzzyIncludeCond.GetBkHostInnerip(),
			InnerIPV6: fuzzyIncludeCond.GetBkHostInneripV6(),
			OuterIP:   fuzzyIncludeCond.GetBkHostOuterip(),
			OuterIPV6: fuzzyIncludeCond.GetBkHostOuteripV6(),
		}
	}

	// exact exclude conditions.
	if exactExcludeCond != nil {
		condition.StaticExactExclude = &types.HostStaticExactFields{
			HostID:        exactExcludeCond.GetBkHostId(),
			BizID:         exactExcludeCond.GetBkBizId(),
			NetworkAreaID: exactExcludeCond.GetBkNetworkareaId(),
		}
		condition.DynamicExactExclude = &types.HostDynamicExactFields{
			NetworkUnitID:  exactExcludeCond.GetBkNetworkunitId(),
			OSType:         exactExcludeCond.GetOsType(),
			Arch:           exactExcludeCond.GetArch(),
			NodeRole:       types.StringListToNodeRoleList(exactExcludeCond.GetNodeRole()),
			NodeStatus:     types.StringListToNodeStatusList(exactExcludeCond.GetNodeStatus()),
			NodeVersion:    exactExcludeCond.GetNodeVersion(),
			NodeGeneration: exactExcludeCond.GetNodeGeneration(),
			AgentID:        exactExcludeCond.GetBkAgentId(),
		}

	}

	return condition
}

func convertHostConditionsFromTypes(
	condition *types.HostCondition) (*TopoHostExactConditions, *TopoHostFuzzyConditions, *TopoHostExactConditions, error) {

	exactIncludeCond := new(TopoHostExactConditions)
	fuzzyIncludeCond := new(TopoHostFuzzyConditions)
	exactExcludeCond := new(TopoHostExactConditions)

	if condition == nil {
		return exactIncludeCond, fuzzyIncludeCond, exactExcludeCond, nil
	}

	if condition.StaticExactInclude != nil {
		exactIncludeCond.BkHostId = condition.StaticExactInclude.HostID
		exactIncludeCond.BkBizId = condition.StaticExactInclude.BizID
		exactIncludeCond.BkNetworkareaId = condition.StaticExactInclude.NetworkAreaID
	}

	if condition.StaticFuzzyInclude != nil {
		fuzzyIncludeCond.BkHostName = condition.StaticFuzzyInclude.HostName
		fuzzyIncludeCond.DeptName = condition.StaticFuzzyInclude.DeptName
		fuzzyIncludeCond.BkHostInnerip = condition.StaticFuzzyInclude.InnerIP
		fuzzyIncludeCond.BkHostInneripV6 = condition.StaticFuzzyInclude.InnerIPV6
		fuzzyIncludeCond.BkHostOuterip = condition.StaticFuzzyInclude.OuterIP
		fuzzyIncludeCond.BkHostOuteripV6 = condition.StaticFuzzyInclude.OuterIPV6
	}

	if condition.DynamicExactInclude != nil {
		exactIncludeCond.BkNetworkunitId = condition.DynamicExactInclude.NetworkUnitID
		exactIncludeCond.OsType = condition.DynamicExactInclude.OSType
		exactIncludeCond.Arch = condition.DynamicExactInclude.Arch
		exactIncludeCond.NodeRole = types.NodeRoleListToStringList(condition.DynamicExactInclude.NodeRole)
		exactIncludeCond.NodeStatus = types.NodeStatusListToStringList(condition.DynamicExactInclude.NodeStatus)
		exactIncludeCond.NodeVersion = condition.DynamicExactInclude.NodeVersion
		exactIncludeCond.NodeGeneration = condition.DynamicExactInclude.NodeGeneration
		exactIncludeCond.BkAgentId = condition.DynamicExactInclude.AgentID
	}

	if condition.StaticExactExclude != nil {
		exactExcludeCond.BkHostId = condition.StaticExactExclude.HostID
		exactExcludeCond.BkBizId = condition.StaticExactExclude.BizID
		exactExcludeCond.BkNetworkareaId = condition.StaticExactExclude.NetworkAreaID
	}

	if condition.DynamicExactExclude != nil {
		exactExcludeCond.BkNetworkunitId = condition.DynamicExactExclude.NetworkUnitID
		exactExcludeCond.OsType = condition.DynamicExactExclude.OSType
		exactExcludeCond.Arch = condition.DynamicExactExclude.Arch
		exactExcludeCond.NodeRole = types.NodeRoleListToStringList(condition.DynamicExactExclude.NodeRole)
		exactExcludeCond.NodeStatus = types.NodeStatusListToStringList(condition.DynamicExactExclude.NodeStatus)
		exactExcludeCond.NodeVersion = condition.DynamicExactExclude.NodeVersion
		exactExcludeCond.NodeGeneration = condition.DynamicExactExclude.NodeGeneration
		exactExcludeCond.BkAgentId = condition.DynamicExactExclude.AgentID
	}

	return exactIncludeCond, fuzzyIncludeCond, exactExcludeCond, nil
}

// Validate validates the request.
func (x *TopoGetHostDistributionByNodeRoleReq) Validate() error {
	return nil
}

// AutoConvert automatically converts the request to types.
func (x *TopoGetHostDistributionByNodeRoleReq) AutoConvert() {
}

// ConvertConditionsToTypes converts the request to types.
func (x *TopoGetHostDistributionByNodeRoleReq) ConvertConditionsToTypes() *types.HostCondition {
	return convertHostConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions(), nil)
}

// ConvertConditionsFromTypes converts the request to types.
func (x *TopoGetHostDistributionByNodeRoleReq) ConvertConditionsFromTypes(condition *types.HostCondition) error {
	exactIncludeCond, fuzzyIncludeCond, _, err := convertHostConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactIncludeCond
	x.FuzzyIncludeConditions = fuzzyIncludeCond

	return nil
}

// Validate validates the request.
func (x *TopoGetHostDistributionByNetworkAreaIDReq) Validate() error {
	return nil
}

// AutoConvert automatically converts the request to types.
func (x *TopoGetHostDistributionByNetworkAreaIDReq) AutoConvert() {
}

// ConvertConditionsToTypes converts the request to types.
func (x *TopoGetHostDistributionByNetworkAreaIDReq) ConvertConditionsToTypes() *types.HostCondition {
	return convertHostConditionsToTypes(x.GetExactIncludeConditions(), x.GetFuzzyIncludeConditions(), nil)
}

// ConvertConditionsFromTypes converts the request to types.
func (x *TopoGetHostDistributionByNetworkAreaIDReq) ConvertConditionsFromTypes(condition *types.HostCondition) error {
	exactIncludeCond, fuzzyIncludeCond, _, err := convertHostConditionsFromTypes(condition)
	if err != nil {
		return err
	}

	x.ExactIncludeConditions = exactIncludeCond
	x.FuzzyIncludeConditions = fuzzyIncludeCond

	return nil
}
