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
	if x.BkHostInnerip == "" && x.BkHostInneripV6 == "" {
		return errors.New("bk_innerip and bk_inneripv6 can not be empty at the same time")
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
			HostID:        host.GetBkHostId(),
			BizID:         host.GetBkBizId(),
			InnerIP:       host.GetBkHostInnerip(),
			InnerIPV6:     host.GetBkHostInneripV6(),
			Addressing:    types.Addressing(host.GetBkAddressing()),
			LoginIP:       host.GetLoginIp(),
			LoginPort:     int64(host.GetLoginPort()),
			LoginUser:     host.GetLoginUser(),
			LoginMode:     types.LoginMode(host.GetLoginMode()),
			LoginPassword: host.GetLoginPassword(),
			LoginKeyFile:  host.GetLoginKeyFile(),
			NetworkUnitID: host.GetBkNetworkunitId(),
			OSType:        host.GetOsType(),
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
	}
}

// ConvertWorkflowID convert workflow id.
func (x *NodeAgentInstallResp) ConvertWorkflowID(workflowID string) {
	x.Data = &NodeAgentInstallResp_Data{WorkflowId: workflowID}
}

// AutoConvert auto convert.
func (x *NodeAgentUpgradeReq) AutoConvert() {
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

	if x.GetTargetVersion() == "" {
		return errors.New("target_version can not be empty")
	}

	return nil
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

	for _, ip := range x.GetBkHostInneripList() {
		if ip == "" {
			return errors.New("bk_host_innerip_list can not contain empty string")
		}
	}

	for _, ipv6 := range x.GetBkHostInneripV6List() {
		if ipv6 == "" {
			return errors.New("bk_host_innerip_v6_list can not contain empty string")
		}
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

const (
	agentInstallCheckResultCategoryNormalInstall            = "normal_install"
	agentInstallCheckResultCategoryRegisterToCMDBAndInstall = "register_to_cmdb_and_install"
	agentInstallCheckResultCategoryNeedConfirm              = "need_confirm"
	agentInstallCheckResultCategoryError                    = "error"
)

// ConvertResultFromTypes convert result from types.
// nolint: cyclop
func (x *NodeAgentInstallCheckResp) ConvertResultFromTypes(requests []*AgentInstallCheckInfo, results []*types.NodeAgentInstallCheckResult) {
	if requests == nil || results == nil || len(requests) != len(results) {
		return
	}

	total := len(requests)

	items := make([]*NodeAgentInstallCheckResult, total)
	for idx := range total {
		request := requests[idx]
		result := results[idx]

		item := &NodeAgentInstallCheckResult{
			Status: string(result.Status),
		}

		switch result.Status {
		case types.NodeAgentInstallCheckStatusDuplicatedInnerIP:
			item.MessageEn = "Inner IPV4 already exists in networkarea, will reinstall this host"
			item.MessageZh = "内网IPV4在该管控区域下已经存在, 将重装该主机"
			item.Category = agentInstallCheckResultCategoryNeedConfirm

		case types.NodeAgentInstallCheckStatusDuplicatedInnerIPV6:
			item.MessageEn = "Inner IPV6 already exists in networkarea, will reinstall this host"
			item.MessageZh = "内网IPV6在该管控区域下已经存在, 将重装该主机"
			item.Category = agentInstallCheckResultCategoryNeedConfirm

		case types.NodeAgentInstallCheckStatusHostNotFound:
			item.MessageEn = "Host does not exist"
			item.MessageZh = "该主机不存在"
			item.Category = agentInstallCheckResultCategoryError

		case types.NodeAgentInstallCheckStatusNetworkUnitNotFound:
			item.MessageEn = "Networkunit does not exist"
			item.MessageZh = "所属管控单元不存在"
			item.Category = agentInstallCheckResultCategoryError

		case types.NodeAgentInstallCheckStatusMismatchedInnerIP:
			item.MessageEn = "Inner IPV4 does not match CMDB configuration"
			item.MessageZh = "内网IPV4与CMDB配置不符"
			item.Category = agentInstallCheckResultCategoryError

		case types.NodeAgentInstallCheckStatusMismatchedInnerIPV6:
			item.MessageEn = "Inner IPV6 does not match CMDB configuration"
			item.MessageZh = "内网IPV6与CMDB配置不符"
			item.Category = agentInstallCheckResultCategoryError

		case types.NodeAgentInstallCheckStatusMismatchedBizID:
			item.MessageEn = "Business ID does not match CMDB configuration"
			item.MessageZh = "所属业务与CMDB配置不符"
			item.Category = agentInstallCheckResultCategoryError

		case types.NodeAgentInstallCheckStatusMismatchedNetworkAreaID:
			item.MessageEn = "Networkarea ID does not match CMDB configuration"
			item.MessageZh = "所属管控区域与CMDB配置不符"
			item.Category = agentInstallCheckResultCategoryError

		case types.NodeAgentInstallCheckStatusInvalidNodeRole:
			item.MessageEn = "Node role does not allow Agent installation, please uninstall the node first"
			item.MessageZh = "节点角色不允许安装Agent, 请先卸载节点"
			item.Category = agentInstallCheckResultCategoryError

		case types.NodeAgentInstallCheckStatusNetworkUnitNotSupportInstall:
			item.MessageEn = "Networkunit lacks available installation proxy nodes"
			item.MessageZh = "所属管控单元缺少可用的安装代理proxy节点"
			item.Category = agentInstallCheckResultCategoryError

		case types.NodeAgentInstallCheckStatusRegisterToCMDBAndInstall:
			item.MessageEn = "Import node to CMDB and install Agent"
			item.MessageZh = "将节点导入CMDB并安装Agent"
			item.Category = agentInstallCheckResultCategoryRegisterToCMDBAndInstall

		case types.NodeAgentInstallCheckStatusNormalInstall:
			if result.Matched != nil && request.GetBkNetworkunitId() != result.Matched.NetworkUnitID {
				item.MessageEn = "Install Agent into networkunit"
				item.MessageZh = "安装节点到新的管控单元"
				item.Category = agentInstallCheckResultCategoryNeedConfirm
			} else {
				item.MessageEn = "Install Agent"
				item.MessageZh = "安装Agent"
				item.Category = agentInstallCheckResultCategoryNormalInstall
			}

		default:
			item.MessageEn = fmt.Sprintf("Unknown error %s", result.Status)
			item.MessageZh = fmt.Sprintf("未知错误 %s", result.Status)
			item.Category = agentInstallCheckResultCategoryError
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
func (x *UploadAgentTemplateReq) AutoConvert() {
}

// Validate check body.
func (x *UploadAgentTemplateReq) Validate() error {
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
func (x *NodeAgentAssignUnitResp) ConvertResult(successCount, failedCount int64, failedReasons []string) {
	x.Data = &NodeAgentAssignUnitResp_Data{
		SuccessCount:  successCount,
		FailedCount:   failedCount,
		FailedReasons: failedReasons,
	}
}
