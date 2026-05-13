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
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate check body.
func (x *NodeAgentInstallReq) Validate() error {
	_, err := conv.SliceToMap(x.GetTargetVersion(), func(v *TargetVersion) string {
		return fmt.Sprintf("%s:%s", v.GetOsType(), v.GetCpuArch())
	})
	if err != nil {
		return err
	}

	hosts := x.GetInfo()
	if len(hosts) == 0 {
		return errors.New("host can not be empty")
	}

	for idx := range hosts {
		if err := hosts[idx].Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate check body.
// nolint: protogetter
func (x *AgentInstallInfo) Validate() error {
	if len(x.BkHostInnerip) == 0 && len(x.BkHostInneripV6) == 0 {
		return errors.New("bk_innerip and bk_inneripv6 can not be empty at the same time")
	}

	if err := validateIPList(x.BkHostInnerip, "bk_host_innerip"); err != nil {
		return err
	}
	if err := validateIPList(x.BkHostInneripV6, "bk_host_innerip_v6"); err != nil {
		return err
	}

	if x.GetBkBizId() < 0 {
		return errors.New("biz_id must be equal or greater than 0")
	}

	if x.OsType == "" {
		return errors.New("os_type can not be empty")
	}

	if err := types.Addressing(x.GetBkAddressing()).Validate(); err != nil {
		return err
	}

	if x.GetBkNetworkunitId() < 0 {
		return errors.New("network_unit_id must be greater than or equal to 0")
	}

	if x.LoginIp == "" {
		return errors.New("login_ip can not be empty")
	}

	if x.GetLoginPort() < 0 {
		return errors.New("login_port must be greater than 0")
	}

	if len(x.GetLoginUser()) == 0 {
		return errors.New("login_user can not be empty")
	}

	if err := types.LoginMode(x.GetLoginMode()).Validate(); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeAgentInstallReq) AutoConvert() {
	for _, host := range x.GetInfo() {
		if host == nil {
			continue
		}

		if host.BkNetworkunitId == nil {
			host.BkNetworkunitId = new(int64)
			*host.BkNetworkunitId = -1
		}

		if host.BkBizId == nil {
			host.BkBizId = new(int64)
			*host.BkBizId = -1
		}

		if host.LoginPort == nil {
			host.LoginPort = new(int64)
			*host.LoginPort = -1
		}

		if host.BkHostId == nil {
			host.BkHostId = new(int64)
			*host.BkHostId = -1
		}
	}
}

// ConvertAgentParamToTypes ...
func (x *NodeAgentInstallReq) ConvertAgentParamToTypes() *types.NodeAgentInstallParam {
	hosts := x.GetInfo()
	if hosts == nil {
		return nil
	}

	hostsParam := make([]*types.NodeAgentInstallHost, len(hosts))

	for idx, host := range hosts {
		hostsParam[idx] = &types.NodeAgentInstallHost{
			HostID:                   host.GetBkHostId(),
			BizID:                    host.GetBkBizId(),
			InnerIP:                  host.GetBkHostInnerip(),
			InnerIPV6:                host.GetBkHostInneripV6(),
			Addressing:               types.Addressing(host.GetBkAddressing()),
			LoginIP:                  host.GetLoginIp(),
			LoginPort:                host.GetLoginPort(),
			LoginUser:                host.GetLoginUser(),
			LoginMode:                types.LoginMode(host.GetLoginMode()),
			LoginPassword:            host.GetLoginPassword(),
			LoginKeyFile:             host.GetLoginKeyFile(),
			NetworkUnitID:            host.GetBkNetworkunitId(),
			OSType:                   host.GetOsType(),
			ReRegister:               host.GetReRegister(),
			InstallPreOrderedPlugins: host.GetInstallPreOrderedPlugins(),
		}
	}

	versions := x.GetTargetVersion()

	targetVersion := make([]*types.TargetVersion, len(versions))
	for idx, version := range versions {
		targetVersion[idx] = &types.TargetVersion{
			Version: version.GetVersion(),
			CPUArch: criteria.CPUArch(version.GetCpuArch()),
			OsType:  criteria.OSType(version.GetOsType()),
		}
	}

	return &types.NodeAgentInstallParam{
		NodeAgentInstallHosts:    hostsParam,
		NodeInstallTargetVersion: targetVersion,
		IsManual:                 x.GetIsManual(),
		EnableCompatibilityMode:  x.GetEnableCompatibilityMode(),
	}
}

// ConvertWorkflowID convert workflow id.
func (x *NodeAgentInstallResp) ConvertWorkflowID(workflowID string) {
	x.Data = &NodeAgentInstallResp_Data{WorkflowId: workflowID}
}

// AutoConvert auto convert.
func (x *NodeAgentUpgradeReq) AutoConvert() {
	hosts := x.GetHost()
	for idx := range hosts {
		hosts[idx].AutoConvert()
	}
}

// Validate check body.
func (x *NodeAgentUpgradeReq) Validate() error {
	hosts := x.GetHost()
	if len(hosts) == 0 {
		return errors.New("host can not be empty")
	}

	for idx := range hosts {
		if err := hosts[idx].Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ConvertParamToTypes convert to types.
func (x *NodeAgentUpgradeReq) ConvertParamToTypes() *types.NodeAgentUpgradeParam {
	hosts := x.GetHost()
	if hosts == nil {
		return nil
	}

	hostsParam := make([]*types.NodeAgentUpgradeHost, len(hosts))

	for idx, host := range hosts {
		hostsParam[idx] = &types.NodeAgentUpgradeHost{
			HostID:                 host.GetBkHostId(),
			NetworkUnitID:          host.GetBkNetworkunitId(),
			CPUArch:                host.GetCpuArch(),
			Force:                  host.GetForce(),
			GracefulRestartTimeout: time.Duration(host.GetGracefulRestartTimeoutSec()) * time.Second,
			TargetVersion:          host.GetTargetVersion(),
		}
	}

	return &types.NodeAgentUpgradeParam{
		Hosts: hostsParam,
	}
}

// Validate check body.
// nolint: protogetter
func (x *NodeAgentUpgradeHost) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id must be equal or greater than 0")
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeAgentUpgradeHost) AutoConvert() {
	if x.BkNetworkunitId == nil {
		x.BkNetworkunitId = new(int64)
		*x.BkNetworkunitId = -1
	}
}

// ConvertWorkflowID convert workflow id.
func (x *NodeAgentUpgradeResp) ConvertWorkflowID(workflowID string) {
	x.Data = &NodeAgentUpgradeResp_Data{WorkflowId: workflowID}
}

// Validate check body.
func (x *NodeAgentReconfigReq) Validate() error {
	hosts := x.GetHost()
	if len(hosts) == 0 {
		return errors.New("host can not be empty")
	}

	for idx := range hosts {
		if err := hosts[idx].Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate check body.
func (x *NodeAgentReconfigHost) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id must be equal or greater than 0")
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeAgentReconfigReq) AutoConvert() {
}

// ConvertParamToTypes convert to types.
func (x *NodeAgentReconfigReq) ConvertParamToTypes() *types.NodeAgentReconfigParam {
	hosts := x.GetHost()
	if hosts == nil {
		return nil
	}

	hostsParam := make([]*types.NodeAgentReconfigHost, len(hosts))

	for idx, host := range hosts {
		hostsParam[idx] = &types.NodeAgentReconfigHost{
			HostID:                 host.GetBkHostId(),
			Force:                  host.GetForce(),
			GracefulRestartTimeout: time.Duration(host.GetGracefulRestartTimeoutSec()) * time.Second,
		}
	}

	return &types.NodeAgentReconfigParam{
		Hosts: hostsParam,
	}
}

// ConvertWorkflowID convert workflow id.
func (x *NodeAgentReconfigResp) ConvertWorkflowID(workflowID string) {
	x.Data = &NodeAgentReconfigResp_Data{WorkflowId: workflowID}
}

// Validate check body.
func (x *NodeAgentRestartReq) Validate() error {
	hosts := x.GetHost()
	if len(hosts) == 0 {
		return errors.New("host can not be empty")
	}

	for idx := range hosts {
		if err := hosts[idx].Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate check body.
func (x *NodeAgentRestartHost) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id must be equal or greater than 0")
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeAgentRestartReq) AutoConvert() {
}

// ConvertParamToTypes convert to types.
func (x *NodeAgentRestartReq) ConvertParamToTypes() *types.NodeAgentRestartParam {
	hosts := x.GetHost()
	if hosts == nil {
		return nil
	}

	hostsParam := make([]*types.NodeAgentRestartHost, len(hosts))

	for idx, host := range hosts {
		hostsParam[idx] = &types.NodeAgentRestartHost{
			HostID:                 host.GetBkHostId(),
			Force:                  host.GetForce(),
			GracefulRestartTimeout: time.Duration(host.GetGracefulRestartTimeoutSec()) * time.Second,
		}
	}

	return &types.NodeAgentRestartParam{
		Hosts: hostsParam,
	}
}

// ConvertParamFromTypes convert host param from types.
func (x *NodeAgentRestartReq) ConvertParamFromTypes(restartParam *types.NodeAgentRestartParam) {
	hostsParam := make([]*NodeAgentRestartHost, len(restartParam.Hosts))
	for idx, host := range restartParam.Hosts {
		hostsParam[idx] = &NodeAgentRestartHost{
			BkHostId:                  host.HostID,
			Force:                     host.Force,
			GracefulRestartTimeoutSec: int64(host.GracefulRestartTimeout.Seconds()),
		}
	}

	x.Host = hostsParam
}

// ConvertWorkflowID convert workflow id.
func (x *NodeAgentRestartResp) ConvertWorkflowID(workflowID string) {
	x.Data = &NodeAgentRestartResp_Data{WorkflowId: workflowID}
}

// Validate check body.
func (x *NodeAgentInstallCheckReq) Validate() error {
	hosts := x.GetHost()
	if len(hosts) == 0 {
		return errors.New("host can not be empty")
	}

	for idx := range hosts {
		if err := hosts[idx].Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate check body.
func (x *AgentInstallCheckInfo) Validate() error {
	if x.GetBkBizId() < 0 {
		return errors.New("biz_id must be equal or greater than 0")
	}

	if len(x.GetBkHostInneripList()) == 0 && len(x.GetBkHostInneripV6List()) == 0 {
		return errors.New("bk_host_innerip_list and bk_host_innerip_v6_list can not be both empty")
	}

	if err := validateIPList(x.GetBkHostInneripList(), "bk_host_innerip_list"); err != nil {
		return err
	}
	if err := validateIPList(x.GetBkHostInneripV6List(), "bk_host_innerip_v6_list"); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeAgentInstallCheckReq) AutoConvert() {
	hosts := x.GetHost()
	for idx := range hosts {
		hosts[idx].AutoConvert()
	}
}

// AutoConvert auto convert.
func (x *AgentInstallCheckInfo) AutoConvert() {
	if x.BkNetworkunitId == nil {
		x.BkNetworkunitId = new(int64)
		*x.BkNetworkunitId = -1
	}

	if x.BkBizId == nil {
		x.BkBizId = new(int64)
		*x.BkBizId = -1
	}

	if x.BkHostId == nil {
		x.BkHostId = new(int64)
		*x.BkHostId = -1
	}
}

// ConvertParamToTypes convert to types.
func (x *NodeAgentInstallCheckReq) ConvertParamToTypes() []*types.NodeAgentInstallCheckParam {
	infos := x.GetHost()
	if infos == nil {
		return nil
	}

	infoParam := make([]*types.NodeAgentInstallCheckParam, len(infos))
	for idx, info := range infos {
		infoParam[idx] = &types.NodeAgentInstallCheckParam{
			BizID:         info.GetBkBizId(),
			HostID:        info.GetBkHostId(),
			NetworkUnitID: info.GetBkNetworkunitId(),
			InnerIPList:   info.GetBkHostInneripList(),
			InnerIPV6List: info.GetBkHostInneripV6List(),
		}
	}

	return infoParam
}

// ConvertResultFromTypes convert result from types.
func (x *NodeAgentInstallCheckResp) ConvertResultFromTypes(results []*types.NodeAgentInstallCheckResult) {
	if results == nil {
		return
	}

	items := make([]*NodeAgentInstallCheckResult, len(results))
	for idx, result := range results {
		item := &NodeAgentInstallCheckResult{
			Status:    string(result.Status),
			MessageEn: result.MessageEn,
			MessageZh: result.MessageZh,
			Category:  result.Category,
		}

		if result.Matched != nil {
			item.Matched = &NodeAgentInstallCheckMatchedItem{
				BkHostId:            &result.Matched.HostID,
				BkBizId:             &result.Matched.BizID,
				BkNetworkareaId:     &result.Matched.NetworkAreaID,
				BkNetworkunitId:     &result.Matched.NetworkUnitID,
				OsType:              string(result.Matched.OsType),
				NodeRole:            string(result.Matched.NodeRole),
				BkHostInneripList:   result.Matched.InnerIPList,
				BkHostInneripV6List: result.Matched.InnerIPV6List,
			}
		}

		items[idx] = item
	}

	x.Data = &NodeAgentInstallCheckResp_Data{
		Results: items,
	}
}

// AutoConvert auto convert.
func (x *UploadAgentInstallTemplateReq) AutoConvert() {
}

// Validate check body.
func (x *UploadAgentInstallTemplateReq) Validate() error {
	return nil
}

// AutoConvert auto convert.
func (x *NodeAgentUninstallReq) AutoConvert() {
}

// Validate check body.
func (x *NodeAgentUninstallReq) Validate() error {
	hosts := x.GetHost()
	if len(hosts) == 0 {
		return errors.New("host can not be empty")
	}

	for idx := range hosts {
		if err := hosts[idx].Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ConvertParamToTypes convert to types.
func (x *NodeAgentUninstallReq) ConvertParamToTypes() *types.NodeAgentUninstallParam {
	hosts := x.GetHost()
	if hosts == nil {
		return nil
	}

	hostsParam := make([]*types.NodeAgentUninstallHost, len(hosts))

	for idx, host := range hosts {
		hostsParam[idx] = &types.NodeAgentUninstallHost{
			HostID: host.GetBkHostId(),
		}
	}

	return &types.NodeAgentUninstallParam{
		Hosts: hostsParam,
	}
}

// Validate check body.
// nolint: protogetter
func (x *NodeAgentUninstallHost) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id must be equal or greater than 0")
	}

	return nil
}

// ConvertWorkflowID convert workflow id.
func (x *NodeAgentUninstallResp) ConvertWorkflowID(workflowID string) {
	x.Data = &NodeAgentUninstallResp_Data{WorkflowId: workflowID}
}

// AutoConvert is a no-op for assign unit requests.
func (x *NodeAgentAssignUnitReq) AutoConvert() {}

// Validate checks that the request body is valid.
func (x *NodeAgentAssignUnitReq) Validate() error {
	if len(x.GetBkHostId()) == 0 {
		return errors.New("bk_host_id can not be empty")
	}
	if x.GetBkNetworkunitId() < 0 {
		return errors.New("bk_networkunit_id must be greater than or equal to 0")
	}
	return nil
}

// ConvertResult populates the response from result values.
// ConvertResult converts the assign unit result from types struct.
func (x *NodeAgentAssignUnitResp) ConvertResult(result *types.NodeAgentAssignUnitResult) {
	x.Data = &NodeAgentAssignUnitResp_Data{
		SuccessCount:  result.SuccessCount,
		FailedCount:   result.FailedCount,
		FailedReasons: result.FailedReasons,
	}
}

// Validate checks body.
func (x *NodeAgentUpgradeCheckReq) Validate() error {
	hosts := x.GetHost()
	if len(hosts) == 0 {
		return errors.New("host can not be empty")
	}

	for idx := range hosts {
		if err := hosts[idx].Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate checks body.
func (x *AgentUpgradeCheckInfo) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id must be equal or greater than 0")
	}

	return nil
}

// AutoConvert auto-converts default values.
func (x *NodeAgentUpgradeCheckReq) AutoConvert() {
	hosts := x.GetHost()
	for idx := range hosts {
		hosts[idx].AutoConvert()
	}
}

// AutoConvert auto-converts default values.
func (x *AgentUpgradeCheckInfo) AutoConvert() {
	if x.BkNetworkunitId == nil {
		x.BkNetworkunitId = new(int64)
		*x.BkNetworkunitId = -1
	}

	if x.BkHostId == nil {
		x.BkHostId = new(int64)
		*x.BkHostId = -1
	}
}

// ConvertParamToTypes converts proto request to type params and target versions.
func (x *NodeAgentUpgradeCheckReq) ConvertParamToTypes() ([]*types.NodeAgentUpgradeCheckParam, []*types.TargetVersion) {
	hosts := x.GetHost()
	params := make([]*types.NodeAgentUpgradeCheckParam, len(hosts))
	for idx, info := range hosts {
		params[idx] = &types.NodeAgentUpgradeCheckParam{
			HostID:        info.GetBkHostId(),
			NetworkUnitID: info.GetBkNetworkunitId(),
			CPUArch:       info.GetCpuArch(),
		}
	}

	rawVersions := x.GetTargetVersion()
	versions := make([]*types.TargetVersion, len(rawVersions))
	for idx, v := range rawVersions {
		versions[idx] = &types.TargetVersion{
			Version: v.GetVersion(),
			OsType:  criteria.OSType(v.GetOsType()),
			CPUArch: criteria.CPUArch(v.GetCpuArch()),
		}
	}

	return params, versions
}

const (
	agentUpgradeCheckResultCategoryNormalUpgrade = "normal_upgrade"
	agentUpgradeCheckResultCategoryNeedConfirm   = "need_confirm"
	agentUpgradeCheckResultCategoryError         = "error"
)

// ConvertResultFromTypes converts type results to proto response with category and messages.
// nolint: cyclop
func (x *NodeAgentUpgradeCheckResp) ConvertResultFromTypes(results []*types.NodeAgentUpgradeCheckResult) {
	if results == nil {
		return
	}

	items := make([]*NodeAgentUpgradeCheckResult, len(results))
	for idx, result := range results {
		item := &NodeAgentUpgradeCheckResult{
			Status: string(result.Status),
		}

		switch result.Status {
		case types.NodeAgentUpgradeCheckStatusHostNotFound:
			item.MessageEn = "Host does not exist"
			item.MessageZh = "该主机不存在"
			item.Category = agentUpgradeCheckResultCategoryError

		case types.NodeAgentUpgradeCheckStatusNetworkUnitNotFound:
			item.MessageEn = "Target networkunit does not exist"
			item.MessageZh = "目标管控单元不存在"
			item.Category = agentUpgradeCheckResultCategoryError

		case types.NodeAgentUpgradeCheckStatusNetworkUnitMismatch:
			item.MessageEn = "Target networkunit does not belong to the same networkarea as the host"
			item.MessageZh = "目标管控单元不属于该主机所在的管控区域"
			item.Category = agentUpgradeCheckResultCategoryError

		case types.NodeAgentUpgradeCheckStatusVersionNotFound:
			item.MessageEn = "No matching version found for host os_type and cpu_arch"
			item.MessageZh = "未找到匹配该主机操作系统和CPU架构的目标版本"
			item.Category = agentUpgradeCheckResultCategoryError

		case types.NodeAgentUpgradeCheckStatusNodeStatusNotAllowed:
			item.MessageEn = "Host current status does not allow upgrade"
			item.MessageZh = "主机当前状态不允许升级"
			item.Category = agentUpgradeCheckResultCategoryError

		case types.NodeAgentUpgradeCheckStatusCPUArchMissing:
			item.MessageEn = "Host cpu_arch is missing"
			item.MessageZh = "主机CPU架构缺失"
			item.Category = agentUpgradeCheckResultCategoryError

		case types.NodeAgentUpgradeCheckStatusNetworkUnitChanged:
			item.MessageEn = "Upgrade will change the host's networkunit, please confirm"
			item.MessageZh = "升级将变更主机的管控单元，请确认"
			item.Category = agentUpgradeCheckResultCategoryNeedConfirm

		case types.NodeAgentUpgradeCheckStatusNormalUpgrade:
			item.MessageEn = "Upgrade Agent"
			item.MessageZh = "升级Agent"
			item.Category = agentUpgradeCheckResultCategoryNormalUpgrade

		default:
			item.MessageEn = fmt.Sprintf("Unknown error %s", result.Status)
			item.MessageZh = fmt.Sprintf("未知错误 %s", result.Status)
			item.Category = agentUpgradeCheckResultCategoryError
		}

		if result.Matched != nil {
			item.Matched = &NodeAgentUpgradeCheckMatchedItem{
				BkHostId:            &result.Matched.HostID,
				BkBizId:             &result.Matched.BizID,
				BkNetworkareaId:     &result.Matched.NetworkAreaID,
				BkNetworkunitId:     &result.Matched.NetworkUnitID,
				OsType:              string(result.Matched.OsType),
				NodeRole:            string(result.Matched.NodeRole),
				BkHostInneripList:   result.Matched.InnerIPList,
				BkHostInneripV6List: result.Matched.InnerIPV6List,
			}
		}

		items[idx] = item
	}

	x.Data = &NodeAgentUpgradeCheckResp_Data{
		Results: items,
	}
}
