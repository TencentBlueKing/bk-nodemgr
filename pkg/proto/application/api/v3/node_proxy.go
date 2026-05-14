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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// Validate check body.
func (x *NodeProxyInstallReq) Validate() error {
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

// AutoConvert auto convert.
func (x *NodeProxyInstallReq) AutoConvert() {
	hosts := x.GetHost()
	for idx := range hosts {
		hosts[idx].AutoConvert()
	}
}

// Validate check body.
// nolint: protogetter
func (x *NodeProxyInstallHost) Validate() error {
	if len(x.GetBkHostInnerip()) == 0 && len(x.GetBkHostInneripV6()) == 0 {
		return errors.New("bk_innerip and bk_inneripv6 can not be empty at the same time")
	}

	if err := validateIPList(x.GetBkHostInnerip(), "bk_host_innerip"); err != nil {
		return err
	}
	if err := validateIPList(x.GetBkHostInneripV6(), "bk_host_innerip_v6"); err != nil {
		return err
	}

	if err := types.Addressing(x.GetBkAddressing()).Validate(); err != nil {
		return err
	}

	if x.GetBkBizId() < 0 {
		return errors.New("biz_id must be >= 0")
	}

	if x.GetOsType() == "" {
		return errors.New("os_type can not be empty")
	}

	if x.GetBkNetworkunitId() < 0 {
		return errors.New("bk_networkunit_id must be >= 0")
	}

	if x.GetLoginIp() == "" {
		return errors.New("login_ip can not be empty")
	}

	if x.GetLoginPort() < 0 {
		return errors.New("login_port must be >= 0")
	}

	if x.GetLoginUser() == "" {
		return errors.New("login_user can not be empty")
	}

	if x.GetRelayCallbackPort() < 0 {
		return errors.New("relay_callback_port must be >= 0")
	}

	if x.GetRelayDownloadPort() < 0 {
		return errors.New("relay_download_port must be >= 0")
	}

	if err := types.LoginMode(x.GetLoginMode()).Validate(); err != nil {
		return err
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeProxyInstallHost) AutoConvert() {
	if x.BkHostId == nil {
		x.BkHostId = new(int64)
		*x.BkHostId = -1
	}
}

// ConvertProxyParamToTypes convert to types.
func (x *NodeProxyInstallReq) ConvertProxyParamToTypes() *types.NodeProxyInstallParam {
	hosts := x.GetHost()
	if hosts == nil {
		return nil
	}

	hostsParam := make([]*types.NodeProxyInstallHost, len(hosts))

	for idx, host := range hosts {
		hostsParam[idx] = &types.NodeProxyInstallHost{
			BizID:                    host.GetBkBizId(),
			NetworkUnitID:            host.GetBkNetworkunitId(),
			HostID:                   host.GetBkHostId(),
			Addressing:               types.Addressing(host.GetBkAddressing()),
			InnerIP:                  host.GetBkHostInnerip(),
			InnerIPV6:                host.GetBkHostInneripV6(),
			OSType:                   host.GetOsType(),
			LoginIP:                  host.GetLoginIp(),
			LoginPort:                int64(host.GetLoginPort()),
			LoginUser:                host.GetLoginUser(),
			LoginMode:                types.LoginMode(host.GetLoginMode()),
			LoginPassword:            host.GetLoginPassword(),
			LoginKeyFile:             host.GetLoginKeyFile(),
			ExportIP:                 host.GetExportIp(),
			AdvertiseIP:              host.GetAdvertiseIp(),
			ReRegister:               host.GetReRegister(),
			InstallPreOrderedPlugins: host.GetInstallPreOrderedPlugins(),
			ProxyTags:                types.StringListToProxyTagList(host.GetProxyTags()),
			ProxyInstallOriginUnitID: host.GetProxyInstallOriginUnitId(),
			CreditExpiredIntervalSec: host.GetCreditExpiredIntervalSec(),
			RelayDownloadPort:        host.GetRelayDownloadPort(),
			RelayCallbackPort:        host.GetRelayCallbackPort(),
			CPUArch:                  host.GetCpuArch(),
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

	return &types.NodeProxyInstallParam{
		Hosts:                   hostsParam,
		TargetVersion:           targetVersion,
		IsManual:                x.GetIsManual(),
		IsOffline:               x.GetIsOffline(),
		EnableCompatibilityMode: x.GetEnableCompatibilityMode(),
	}
}

// ConvertWorkflowID convert workflow id.
func (x *NodeProxyInstallResp) ConvertWorkflowID(workflowID string) {
	x.Data = &NodeProxyInstallResp_Data{WorkflowId: workflowID}
}

// Validate check body.
func (x *NodeProxyUpgradeReq) Validate() error {
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

// AutoConvert auto convert.
func (x *NodeProxyUpgradeReq) AutoConvert() {
	hosts := x.GetHost()
	for idx := range hosts {
		hosts[idx].AutoConvert()
	}
}

// ConvertParamToTypes convert to types.
func (x *NodeProxyUpgradeReq) ConvertParamToTypes() *types.NodeProxyUpgradeParam {
	hosts := x.GetHost()
	if hosts == nil {
		return nil
	}

	hostsParam := make([]*types.NodeProxyUpgradeHost, len(hosts))

	for idx, host := range hosts {
		hostsParam[idx] = &types.NodeProxyUpgradeHost{
			HostID:                 host.GetBkHostId(),
			NetworkUnitID:          host.GetBkNetworkunitId(),
			CPUArch:                host.GetCpuArch(),
			Force:                  host.GetForce(),
			TargetVersion:          host.GetTargetVersion(),
			GracefulRestartTimeout: time.Duration(host.GetGracefulRestartTimeoutSec()) * time.Second,
		}
	}

	return &types.NodeProxyUpgradeParam{
		Hosts: hostsParam,
	}
}

// Validate check body.
// nolint: protogetter
func (x *NodeProxyUpgradeHost) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id must be >= 0")
	}

	if !x.GetForce() && x.GetGracefulRestartTimeoutSec() <= 0 {
		return errors.New("graceful_restart_timeout_sec must be > 0 when force is false")
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeProxyUpgradeHost) AutoConvert() {
	if x.BkNetworkunitId == nil {
		x.BkNetworkunitId = new(int64)
		*x.BkNetworkunitId = -1
	}
}

// ConvertWorkflowID convert workflow id.
func (x *NodeProxyUpgradeResp) ConvertWorkflowID(workflowID string) {
	x.Data = &NodeProxyUpgradeResp_Data{WorkflowId: workflowID}
}

// Validate check body.
func (x *NodeProxyRestartReq) Validate() error {
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

// AutoConvert auto convert.
func (x *NodeProxyRestartReq) AutoConvert() {
	hosts := x.GetHost()
	for idx := range hosts {
		hosts[idx].AutoConvert()
	}
}

// ConvertParamFromTypes convert host param from types.
func (x *NodeProxyRestartReq) ConvertParamFromTypes(restartParam *types.NodeProxyRestartParam) {
	hostsParam := make([]*NodeProxyRestartHost, len(restartParam.Hosts))
	for idx, host := range restartParam.Hosts {
		hostsParam[idx] = &NodeProxyRestartHost{
			BkHostId:                  host.HostID,
			Force:                     host.Force,
			GracefulRestartTimeoutSec: int64(host.GracefulRestartTimeout.Seconds()),
		}
	}

	x.Host = hostsParam
}

// ConvertParamToTypes convert to types.
func (x *NodeProxyRestartReq) ConvertParamToTypes() *types.NodeProxyRestartParam {
	hosts := x.GetHost()
	if hosts == nil {
		return nil
	}

	hostsParam := make([]*types.NodeProxyRestartHost, len(hosts))

	for idx, host := range hosts {
		hostsParam[idx] = &types.NodeProxyRestartHost{
			HostID:                 host.GetBkHostId(),
			Force:                  host.GetForce(),
			GracefulRestartTimeout: time.Duration(host.GetGracefulRestartTimeoutSec()) * time.Second,
		}
	}

	return &types.NodeProxyRestartParam{
		Hosts: hostsParam,
	}
}

// Validate check body.
// nolint: protogetter
func (x *NodeProxyRestartHost) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id must be >= 0")
	}

	if !x.GetForce() && x.GetGracefulRestartTimeoutSec() <= 0 {
		return errors.New("graceful_restart_timeout_sec must be > 0 when force is false")
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeProxyRestartHost) AutoConvert() {
}

// ConvertWorkflowID convert workflow id.
func (x *NodeProxyRestartResp) ConvertWorkflowID(workflowID string) {
	x.Data = &NodeProxyRestartResp_Data{WorkflowId: workflowID}
}

// Validate check body.
func (x *NodeProxyReconfigReq) Validate() error {
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

// AutoConvert auto convert.
func (x *NodeProxyReconfigReq) AutoConvert() {
	hosts := x.GetHost()
	for idx := range hosts {
		hosts[idx].AutoConvert()
	}
}

// ConvertParamToTypes convert to types.
func (x *NodeProxyReconfigReq) ConvertParamToTypes() *types.NodeProxyReconfigParam {
	hosts := x.GetHost()
	if hosts == nil {
		return nil
	}

	hostsParam := make([]*types.NodeProxyReconfigHost, len(hosts))

	for idx, host := range hosts {
		hostsParam[idx] = &types.NodeProxyReconfigHost{
			HostID:                 host.GetBkHostId(),
			Force:                  host.GetForce(),
			GracefulRestartTimeout: time.Duration(host.GetGracefulRestartTimeoutSec()) * time.Second,
		}
	}

	return &types.NodeProxyReconfigParam{
		Hosts: hostsParam,
	}
}

// Validate check body.
// nolint: protogetter
func (x *NodeProxyReconfigHost) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id must be >= 0")
	}

	if !x.GetForce() && x.GetGracefulRestartTimeoutSec() <= 0 {
		return errors.New("graceful_restart_timeout_sec must be > 0 when force is false")
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeProxyReconfigHost) AutoConvert() {
}

// ConvertWorkflowID convert workflow id.
func (x *NodeProxyReconfigResp) ConvertWorkflowID(workflowID string) {
	x.Data = &NodeProxyReconfigResp_Data{WorkflowId: workflowID}
}

// Validate check body.
func (x *NodeProxyUpdateReq) Validate() error {
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

// AutoConvert auto convert.
func (x *NodeProxyUpdateReq) AutoConvert() {
	hosts := x.GetHost()
	for idx := range hosts {
		hosts[idx].AutoConvert()
	}
}

// ConvertParamToTypes convert host param from types.
func (x *NodeProxyUpdateReq) ConvertParamToTypes() *types.NodeProxyUpdateParam {
	hosts := x.GetHost()
	if hosts == nil {
		return nil
	}

	hostsParam := make([]*types.NodeProxyUpdateHost, len(hosts))

	for idx, host := range hosts {
		hostsParam[idx] = &types.NodeProxyUpdateHost{
			HostID:            host.GetBkHostId(),
			LoginIP:           host.GetLoginIp(),
			LoginPort:         host.GetLoginPort(),
			LoginUser:         host.GetLoginUser(),
			LoginMode:         types.LoginMode(host.GetLoginMode()),
			ExportIP:          host.GetExportIp(),
			AdvertiseIP:       host.GetAdvertiseIp(),
			ProxyTags:         types.StringListToProxyTagList(host.GetProxyTags()),
			RelayDownloadPort: host.GetRelayDownloadPort(),
			RelayCallbackPort: host.GetRelayCallbackPort(),
		}
	}

	return &types.NodeProxyUpdateParam{
		Hosts: hostsParam,
	}
}

// Validate check body.
// nolint: protogetter
func (x *NodeProxyUpdateHost) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id must be >= 0")
	}

	if x.GetLoginIp() == "" {
		return errors.New("login_ip can not be empty")
	}

	if x.GetLoginPort() < 0 {
		return errors.New("login_port must be >= 0")
	}

	if x.GetLoginUser() == "" {
		return errors.New("login_user can not be empty")
	}

	if x.GetRelayCallbackPort() < 0 {
		return errors.New("relay_callback_port must be >= 0")
	}

	if x.GetRelayDownloadPort() < 0 {
		return errors.New("relay_download_port must be >= 0")
	}

	if x.GetLoginMode() != "" {
		if err := types.LoginMode(x.GetLoginMode()).Validate(); err != nil {
			return err
		}
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeProxyUpdateHost) AutoConvert() {
}

// AutoConvert auto convert.
func (x *NodeProxyUninstallReq) AutoConvert() {
}

// Validate check body.
func (x *NodeProxyUninstallReq) Validate() error {
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
func (x *NodeProxyUninstallReq) ConvertParamToTypes() *types.NodeProxyUninstallParam {
	hosts := x.GetHost()
	if hosts == nil {
		return nil
	}

	hostsParam := make([]*types.NodeProxyUninstallHost, len(hosts))

	for idx, host := range hosts {
		hostsParam[idx] = &types.NodeProxyUninstallHost{
			HostID: host.GetBkHostId(),
		}
	}

	return &types.NodeProxyUninstallParam{
		Hosts: hostsParam,
	}
}

// Validate check body.
// nolint: protogetter
func (x *NodeProxyUninstallHost) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id must be equal or greater than 0")
	}

	return nil
}

// ConvertWorkflowID convert workflow id.
func (x *NodeProxyUninstallResp) ConvertWorkflowID(workflowID string) {
	x.Data = &NodeProxyUninstallResp_Data{WorkflowId: workflowID}
}

// Validate check body.
func (x *NodeProxyInstallCheckReq) Validate() error {
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
func (x *ProxyInstallCheckInfo) Validate() error {
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
func (x *NodeProxyInstallCheckReq) AutoConvert() {
	hosts := x.GetHost()
	for idx := range hosts {
		hosts[idx].AutoConvert()
	}
}

// AutoConvert auto convert.
func (x *ProxyInstallCheckInfo) AutoConvert() {
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
func (x *NodeProxyInstallCheckReq) ConvertParamToTypes() []*types.NodeProxyInstallCheckParam {
	infos := x.GetHost()
	if infos == nil {
		return nil
	}

	infoParam := make([]*types.NodeProxyInstallCheckParam, len(infos))
	for idx, info := range infos {
		infoParam[idx] = &types.NodeProxyInstallCheckParam{
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
func (x *NodeProxyInstallCheckResp) ConvertResultFromTypes(results []*types.NodeProxyInstallCheckResult) {
	if results == nil {
		return
	}

	items := make([]*NodeProxyInstallCheckResult, len(results))
	for idx, result := range results {
		item := &NodeProxyInstallCheckResult{
			Status:    string(result.Status),
			MessageEn: result.MessageEn,
			MessageZh: result.MessageZh,
			Category:  result.Category,
		}

		if result.Matched != nil {
			item.Matched = &NodeProxyInstallCheckMatchedItem{
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

	x.Data = &NodeProxyInstallCheckResp_Data{
		Results: items,
	}
}

// Validate checks body.
func (x *NodeProxyUpgradeCheckReq) Validate() error {
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
func (x *ProxyUpgradeCheckInfo) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id must be equal or greater than 0")
	}

	return nil
}

// AutoConvert auto-converts default values.
func (x *NodeProxyUpgradeCheckReq) AutoConvert() {
	hosts := x.GetHost()
	for idx := range hosts {
		hosts[idx].AutoConvert()
	}
}

// AutoConvert auto-converts default values.
func (x *ProxyUpgradeCheckInfo) AutoConvert() {
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
func (x *NodeProxyUpgradeCheckReq) ConvertParamToTypes() ([]*types.NodeProxyUpgradeCheckParam, []*types.TargetVersion) {
	hosts := x.GetHost()
	params := make([]*types.NodeProxyUpgradeCheckParam, len(hosts))
	for idx, info := range hosts {
		params[idx] = &types.NodeProxyUpgradeCheckParam{
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
	proxyUpgradeCheckResultCategoryNormalUpgrade = "normal_upgrade"
	proxyUpgradeCheckResultCategoryNeedConfirm   = "need_confirm"
	proxyUpgradeCheckResultCategoryError         = "error"
)

// ConvertResultFromTypes converts type results to proto response with category and messages.
// nolint: cyclop
func (x *NodeProxyUpgradeCheckResp) ConvertResultFromTypes(results []*types.NodeProxyUpgradeCheckResult) {
	if results == nil {
		return
	}

	items := make([]*NodeProxyUpgradeCheckResult, len(results))
	for idx, result := range results {
		item := &NodeProxyUpgradeCheckResult{
			Status: string(result.Status),
		}

		switch result.Status {
		case types.NodeProxyUpgradeCheckStatusHostNotFound:
			item.MessageEn = "Host does not exist"
			item.MessageZh = "该主机不存在"
			item.Category = proxyUpgradeCheckResultCategoryError

		case types.NodeProxyUpgradeCheckStatusNetworkUnitNotFound:
			item.MessageEn = "Target networkunit does not exist"
			item.MessageZh = "目标管控单元不存在"
			item.Category = proxyUpgradeCheckResultCategoryError

		case types.NodeProxyUpgradeCheckStatusNetworkUnitMismatch:
			item.MessageEn = "Target networkunit does not belong to the same networkarea as the host"
			item.MessageZh = "目标管控单元不属于该主机所在的管控区域"
			item.Category = proxyUpgradeCheckResultCategoryError

		case types.NodeProxyUpgradeCheckStatusVersionNotFound:
			item.MessageEn = "No matching version found for host os_type and cpu_arch"
			item.MessageZh = "未找到匹配该主机操作系统和CPU架构的目标版本"
			item.Category = proxyUpgradeCheckResultCategoryError

		case types.NodeProxyUpgradeCheckStatusNodeStatusNotAllowed:
			item.MessageEn = "Host current status does not allow upgrade"
			item.MessageZh = "主机当前状态不允许升级"
			item.Category = proxyUpgradeCheckResultCategoryError

		case types.NodeProxyUpgradeCheckStatusCPUArchMissing:
			item.MessageEn = "Host cpu_arch is missing"
			item.MessageZh = "主机CPU架构缺失"
			item.Category = proxyUpgradeCheckResultCategoryError

		case types.NodeProxyUpgradeCheckStatusNetworkUnitChanged:
			item.MessageEn = "Upgrade will change the host's networkunit, please confirm"
			item.MessageZh = "升级将变更主机的管控单元，请确认"
			item.Category = proxyUpgradeCheckResultCategoryNeedConfirm

		case types.NodeProxyUpgradeCheckStatusNormalUpgrade:
			item.MessageEn = "Upgrade Proxy"
			item.MessageZh = "升级Proxy"
			item.Category = proxyUpgradeCheckResultCategoryNormalUpgrade

		default:
			item.MessageEn = fmt.Sprintf("Unknown error %s", result.Status)
			item.MessageZh = fmt.Sprintf("未知错误 %s", result.Status)
			item.Category = proxyUpgradeCheckResultCategoryError
		}

		if result.Matched != nil {
			item.Matched = &NodeProxyUpgradeCheckMatchedItem{
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

	x.Data = &NodeProxyUpgradeCheckResp_Data{
		Results: items,
	}
}

// Validate check body.
func (x *NodeProxyAssignUnitReq) Validate() error {
	if len(x.GetBkHostId()) == 0 {
		return errors.New("bk_host_id can not be empty")
	}

	if x.GetBkNetworkunitId() <= 0 {
		return errors.New("bk_networkunit_id must be > 0")
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeProxyAssignUnitReq) AutoConvert() {}

// ConvertResult converts the assign unit result to the response.
// ConvertResult converts the assign unit result from types struct.
func (x *NodeProxyAssignUnitResp) ConvertResult(result *types.NodeProxyAssignUnitResult) {
	x.Data = &NodeProxyAssignUnitResp_Data{
		SuccessCount:  result.SuccessCount,
		FailedCount:   result.FailedCount,
		FailedReasons: result.FailedReasons,
		WorkflowId:    result.WorkflowID,
	}
}

// AutoConvert auto convert.
func (x *UploadProxyInstallTemplateReq) AutoConvert() {
}

// Validate check body.
func (x *UploadProxyInstallTemplateReq) Validate() error {
	return nil
}
