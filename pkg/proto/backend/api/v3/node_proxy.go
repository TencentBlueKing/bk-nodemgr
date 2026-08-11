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

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
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

// ConvertParamFromTypes convert param from types.
func (x *NodeProxyInstallReq) ConvertParamFromTypes(installParam *types.NodeProxyInstallParam) {
	hostsParam := make([]*NodeProxyInstallHost, len(installParam.Hosts))
	for idx, host := range installParam.Hosts {
		hostsParam[idx] = &NodeProxyInstallHost{
			BkBizId:                  host.BizID,
			BkNetworkunitId:          host.NetworkUnitID,
			BkHostId:                 &host.HostID,
			BkAddressing:             string(host.Addressing),
			BkHostInnerip:            host.InnerIP,
			BkHostInneripV6:          host.InnerIPV6,
			OsType:                   host.OSType,
			LoginIp:                  host.LoginIP,
			LoginPort:                host.LoginPort,
			LoginUser:                host.LoginUser,
			LoginMode:                string(host.LoginMode),
			LoginPassword:            host.LoginPassword,
			LoginKeyFile:             host.LoginKeyFile,
			ExportIp:                 host.ExportIP,
			AdvertiseIp:              host.AdvertiseIP,
			ReRegister:               host.ReRegister,
			RenewGseTask:             host.RenewGSETask,
			RenewGseProc:             host.RenewGSEProc,
			InstallPreOrderedPlugins: host.InstallPreOrderedPlugins,
			ProxyTags:                types.ProxyTagListToStringList(host.ProxyTags),
			ProxyInstallOriginUnitId: host.ProxyInstallOriginUnitID,
			CreditExpiredIntervalSec: host.CreditExpiredIntervalSec,
			RelayDownloadPort:        host.RelayDownloadPort,
			RelayCallbackPort:        host.RelayCallbackPort,
			CpuArch:                  host.CPUArch,
			InstallMethod:            string(host.InstallMethod),
		}
	}

	targetVersion := make([]*TargetVersion, len(installParam.TargetVersion))
	for idx, version := range installParam.TargetVersion {
		targetVersion[idx] = &TargetVersion{
			Version: version.Version,
			CpuArch: string(version.CPUArch),
			OsType:  string(version.OsType),
		}
	}

	x.TargetVersion = targetVersion
	x.Host = hostsParam
	x.IsManual = installParam.IsManual
	x.IsOffline = installParam.IsOffline
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
	if err := types.NodeInstallMethod(x.GetInstallMethod()).Validate(); err != nil {
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

// ConvertWorkflowID convert workflow id.
func (x *NodeProxyInstallResp) ConvertWorkflowID(workflowID string) {
	x.Data = &NodeProxyInstallResp_Data{WorkflowId: workflowID}
}

// GetWorkflowID get workflow id.
func (x *NodeProxyInstallResp) GetWorkflowID() string {
	if x.GetData() != nil {
		return x.GetData().GetWorkflowId()
	}

	return ""
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

// ConvertParamFromTypes convert host param from types.
func (x *NodeProxyUpgradeReq) ConvertParamFromTypes(upgradeParam *types.NodeProxyUpgradeParam) {
	hostsParam := make([]*NodeProxyUpgradeReq_Host, len(upgradeParam.Hosts))
	for idx, host := range upgradeParam.Hosts {
		networkUnitID := host.NetworkUnitID
		hostsParam[idx] = &NodeProxyUpgradeReq_Host{
			BkHostId:                  host.HostID,
			BkNetworkunitId:           &networkUnitID,
			CpuArch:                   host.CPUArch,
			Force:                     host.Force,
			GracefulRestartTimeoutSec: int64(host.GracefulRestartTimeout.Seconds()),
			TargetVersion:             host.TargetVersion,
		}
	}

	x.Host = hostsParam
}

// AutoConvert auto convert.
func (x *NodeProxyUpgradeReq) AutoConvert() {
	hosts := x.GetHost()
	for idx := range hosts {
		hosts[idx].AutoConvert()
	}
}

// Validate check body.
// nolint: protogetter
func (x *NodeProxyUpgradeReq_Host) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id must be >= 0")
	}

	if !x.GetForce() && x.GetGracefulRestartTimeoutSec() <= 0 {
		return errors.New("graceful_restart_timeout_sec must be > 0 when force is false")
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeProxyUpgradeReq_Host) AutoConvert() {
	if x.BkNetworkunitId == nil {
		x.BkNetworkunitId = new(int64)
		*x.BkNetworkunitId = -1
	}
}

// ConvertWorkflowID convert workflow id.
func (x *NodeProxyUpgradeResp) ConvertWorkflowID(workflowID string) {
	x.Data = &NodeProxyUpgradeResp_Data{WorkflowId: workflowID}
}

// GetWorkflowID get workflow id.
func (x *NodeProxyUpgradeResp) GetWorkflowID() string {
	if x.GetData() != nil {
		return x.GetData().GetWorkflowId()
	}

	return ""
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
	hostsParam := make([]*NodeProxyRestartReq_Host, len(restartParam.Hosts))
	for idx, host := range restartParam.Hosts {
		hostsParam[idx] = &NodeProxyRestartReq_Host{
			BkHostId:                  host.HostID,
			Force:                     host.Force,
			GracefulRestartTimeoutSec: int64(host.GracefulRestartTimeout.Seconds()),
		}
	}

	x.Host = hostsParam
}

// Validate check body.
// nolint: protogetter
func (x *NodeProxyRestartReq_Host) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id must be >= 0")
	}

	if !x.GetForce() && x.GetGracefulRestartTimeoutSec() <= 0 {
		return errors.New("graceful_restart_timeout_sec must be > 0 when force is false")
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeProxyRestartReq_Host) AutoConvert() {
}

// ConvertWorkflowID convert workflow id.
func (x *NodeProxyRestartResp) ConvertWorkflowID(workflowID string) {
	x.Data = &NodeProxyRestartResp_Data{WorkflowId: workflowID}
}

// GetWorkflowID get workflow id.
func (x *NodeProxyRestartResp) GetWorkflowID() string {
	if x.GetData() != nil {
		return x.GetData().GetWorkflowId()
	}

	return ""
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

// ConvertParamFromTypes convert host param from types.
func (x *NodeProxyReconfigReq) ConvertParamFromTypes(reconfigParam *types.NodeProxyReconfigParam) {
	hostsParam := make([]*NodeProxyReconfigReq_Host, len(reconfigParam.Hosts))
	for idx, host := range reconfigParam.Hosts {
		hostsParam[idx] = &NodeProxyReconfigReq_Host{
			BkHostId:                  host.HostID,
			Force:                     host.Force,
			GracefulRestartTimeoutSec: int64(host.GracefulRestartTimeout.Seconds()),
		}
	}

	x.Host = hostsParam
}

// Validate check body.
// nolint: protogetter
func (x *NodeProxyReconfigReq_Host) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id must be >= 0")
	}

	if !x.GetForce() && x.GetGracefulRestartTimeoutSec() <= 0 {
		return errors.New("graceful_restart_timeout_sec must be > 0 when force is false")
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeProxyReconfigReq_Host) AutoConvert() {
}

// ConvertWorkflowID convert workflow id.
func (x *NodeProxyReconfigResp) ConvertWorkflowID(workflowID string) {
	x.Data = &NodeProxyReconfigResp_Data{WorkflowId: workflowID}
}

// GetWorkflowID get workflow id.
func (x *NodeProxyReconfigResp) GetWorkflowID() string {
	if x.GetData() != nil {
		return x.GetData().GetWorkflowId()
	}

	return ""
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

// ConvertHostToTypes convert host to types.
func (x *NodeProxyUpdateReq) ConvertHostToTypes() []*types.Host {
	hosts := make([]*types.Host, 0, len(x.GetHost()))
	for _, host := range x.GetHost() {
		hosts = append(hosts, &types.Host{
			HostID: host.GetBkHostId(),
			Dynamic: &types.HostDynamic{
				LoginIP:           host.GetLoginIp(),
				LoginPort:         host.GetLoginPort(),
				LoginUser:         host.GetLoginUser(),
				LoginMode:         types.LoginMode(host.GetLoginMode()),
				ExportIP:          host.GetExportIp(),
				ExportIPV6:        host.GetExportIpV6(),
				AdvertiseIP:       host.GetAdvertiseIp(),
				AdvertiseIPV6:     host.GetAdvertiseIpV6(),
				ProxyTags:         types.StringListToProxyTagList(host.GetProxyTags()),
				RelayDownloadPort: host.GetRelayDownloadPort(),
				RelayCallbackPort: host.GetRelayCallbackPort(),
			},
		})
	}

	return hosts
}

// ConvertParamFromTypes convert host param from types.
func (x *NodeProxyUpdateReq) ConvertParamFromTypes(updateParam *types.NodeProxyUpdateParam) {
	hostsParam := make([]*NodeProxyUpdateHost, len(updateParam.Hosts))
	for idx, host := range updateParam.Hosts {
		hostsParam[idx] = &NodeProxyUpdateHost{
			BkHostId:          host.HostID,
			LoginIp:           host.LoginIP,
			LoginPort:         host.LoginPort,
			LoginUser:         host.LoginUser,
			LoginMode:         string(host.LoginMode),
			ExportIp:          host.ExportIP,
			ExportIpV6:        host.ExportIPV6,
			AdvertiseIp:       host.AdvertiseIP,
			AdvertiseIpV6:     host.AdvertiseIPV6,
			ProxyTags:         types.ProxyTagListToStringList(host.ProxyTags),
			RelayDownloadPort: host.RelayDownloadPort,
			RelayCallbackPort: host.RelayCallbackPort,
		}
	}

	x.Host = hostsParam
}

// ConvertHostFieldsToTypes convert host dynamic fields to types.
func (x *NodeProxyUpdateReq) ConvertHostFieldsToTypes() types.HostDynamicFields {
	hasLoginMode := false
	for _, host := range x.GetHost() {
		if host.GetLoginMode() != "" {
			hasLoginMode = true
			break
		}
	}

	return types.HostDynamicFields{
		LoginIP:           true,
		LoginPort:         true,
		LoginUser:         true,
		LoginMode:         hasLoginMode,
		ExportIP:          true,
		ExportIPV6:        true,
		AdvertiseIP:       true,
		AdvertiseIPV6:     true,
		ProxyTags:         true,
		RelayDownloadPort: true,
		RelayCallbackPort: true,
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

// Validate check body.
func (x *NodeProxyUninstallReq_Host) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id must be equal or greater than 0")
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeProxyUninstallReq) AutoConvert() {
}

// ConvertParamFromTypes convert host param from types.
func (x *NodeProxyUninstallReq) ConvertParamFromTypes(restartParam *types.NodeProxyUninstallParam) {
	hostsParam := make([]*NodeProxyUninstallReq_Host, len(restartParam.Hosts))
	for idx, host := range restartParam.Hosts {
		hostsParam[idx] = &NodeProxyUninstallReq_Host{
			BkHostId: host.HostID,
		}
	}

	x.Host = hostsParam
}

// ConvertWorkflowID convert workflow id.
func (x *NodeProxyUninstallResp) ConvertWorkflowID(workflowID string) {
	x.Data = &NodeProxyUninstallResp_Data{WorkflowId: workflowID}
}

// GetWorkflowID get workflow id.
func (x *NodeProxyUninstallResp) GetWorkflowID() string {
	data := x.GetData()
	if data == nil {
		return ""
	}

	return data.GetWorkflowId()
}

// Validate checks that the request body is valid.
func (x *NodeProxyAssignUnitReq) Validate() error {
	if len(x.GetBkHostId()) == 0 {
		return errors.New("bk_host_id can not be empty")
	}

	if x.GetBkNetworkunitId() < 0 {
		return errors.New("bk_networkunit_id must be greater than or equal to 0")
	}

	return nil
}

// AutoConvert is a no-op for assign unit requests.
func (x *NodeProxyAssignUnitReq) AutoConvert() {}

// ConvertParamFromTypes converts param from types.
func (x *NodeProxyAssignUnitReq) ConvertParamFromTypes(param *types.NodeProxyAssignUnitParam) {
	x.BkHostId = param.HostIDs
	x.BkNetworkunitId = param.NetworkUnitID
}

// Validate checks that the multi-assignment request body is valid.
func (x *NodeProxyAssignUnitMultiReq) Validate() error {
	if len(x.GetItems()) == 0 {
		return errors.New("items can not be empty")
	}

	seenHostIDs := make(map[int64]struct{})
	for _, item := range x.GetItems() {
		if item == nil {
			return errors.New("item can not be nil")
		}
		if len(item.GetBkHostId()) == 0 {
			return errors.New("bk_host_id can not be empty")
		}
		if item.GetBkNetworkunitId() < 0 {
			return errors.New("bk_networkunit_id must be greater than or equal to 0")
		}
		for _, hostID := range item.GetBkHostId() {
			if _, ok := seenHostIDs[hostID]; ok {
				return fmt.Errorf("host-id(%d) appears in multiple items", hostID)
			}
			seenHostIDs[hostID] = struct{}{}
		}
	}

	return nil
}

// AutoConvert is a no-op for multi-assignment requests.
func (x *NodeProxyAssignUnitMultiReq) AutoConvert() {}

// ConvertParamFromTypes converts a multi-assignment parameter.
func (x *NodeProxyAssignUnitMultiReq) ConvertParamFromTypes(param *types.NodeProxyAssignUnitMultiParam) {
	x.Items = make([]*NodeProxyAssignUnitMultiReq_Item, len(param.Items))
	for idx, item := range param.Items {
		x.Items[idx] = &NodeProxyAssignUnitMultiReq_Item{
			BkHostId:        item.HostIDs,
			BkNetworkunitId: item.NetworkUnitID,
		}
	}
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
func (x *NodeProxyInstallCheckReq_Host) Validate() error {
	if x.GetBkBizId() < 0 {
		return errors.New("bk_biz_id is required")
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
func (x *NodeProxyInstallCheckReq_Host) AutoConvert() {
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

// ConvertToTypeHost converts request host to types.Host.
func (x *NodeProxyInstallCheckReq_Host) ConvertToTypeHost() *types.Host {
	return &types.Host{
		HostID: x.GetBkHostId(),
		Static: &types.HostStatic{
			BizID: x.GetBkBizId(),
		},
	}
}

// ConvertHostsToTypes converts request hosts to types.Host list.
func (x *NodeProxyInstallCheckReq) ConvertHostsToTypes() []*types.Host {
	hosts := make([]*types.Host, 0, len(x.GetHost()))
	for _, host := range x.GetHost() {
		hosts = append(hosts, host.ConvertToTypeHost())
	}

	return hosts
}

// ConvertParamFromTypes convert host param from types.
func (x *NodeProxyInstallCheckReq) ConvertParamFromTypes(checkParam []*types.NodeProxyInstallCheckParam) {
	hostsParam := make([]*NodeProxyInstallCheckReq_Host, len(checkParam))
	for idx, host := range checkParam {
		hostsParam[idx] = &NodeProxyInstallCheckReq_Host{
			BkHostId:            &host.HostID,
			BkBizId:             &host.BizID,
			BkNetworkunitId:     &host.NetworkUnitID,
			BkHostInneripList:   host.InnerIPList,
			BkHostInneripV6List: host.InnerIPV6List,
		}
	}

	x.Host = hostsParam
}

// ConvertResultFromTypes convert result from types.
func (x *NodeProxyInstallCheckResp) ConvertResultFromTypes(results []*types.NodeProxyInstallCheckResult) {
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

// ConvertResultToTypes convert result to types.
func (x *NodeProxyInstallCheckResp) ConvertResultToTypes() []*types.NodeProxyInstallCheckResult {
	if x.GetData() == nil {
		return nil
	}

	data := x.GetData()

	items := make([]*types.NodeProxyInstallCheckResult, len(data.GetResults()))
	for idx, result := range data.GetResults() {
		item := &types.NodeProxyInstallCheckResult{
			Status:    types.NodeProxyInstallCheckStatus(result.GetStatus()),
			MessageEn: result.GetMessageEn(),
			MessageZh: result.GetMessageZh(),
			Category:  result.GetCategory(),
		}

		if matched := result.GetMatched(); matched != nil {
			item.Matched = &types.NodeProxyInstallCheckMatchedItem{
				HostID:        matched.GetBkHostId(),
				BizID:         matched.GetBkBizId(),
				NetworkAreaID: matched.GetBkNetworkareaId(),
				NetworkUnitID: matched.GetBkNetworkunitId(),
				OsType:        criteria.OSType(matched.GetOsType()),
				NodeRole:      types.NodeRole(matched.GetNodeRole()),
				InnerIPList:   matched.GetBkHostInneripList(),
				InnerIPV6List: matched.GetBkHostInneripV6List(),
			}
		}

		items[idx] = item
	}

	return items
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
func (x *NodeProxyUpgradeCheckReq_Host) Validate() error {
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
func (x *NodeProxyUpgradeCheckReq_Host) AutoConvert() {
	if x.BkNetworkunitId == nil {
		x.BkNetworkunitId = new(int64)
		*x.BkNetworkunitId = -1
	}

	if x.BkHostId == nil {
		x.BkHostId = new(int64)
		*x.BkHostId = -1
	}
}

// ConvertParamFromTypes converts type params to proto request hosts.
func (x *NodeProxyUpgradeCheckReq) ConvertParamFromTypes(params []*types.NodeProxyUpgradeCheckParam, targetVersions []*types.TargetVersion) {
	hosts := make([]*NodeProxyUpgradeCheckReq_Host, len(params))
	for idx, p := range params {
		networkUnitID := p.NetworkUnitID
		hosts[idx] = &NodeProxyUpgradeCheckReq_Host{
			BkHostId:        &p.HostID,
			BkNetworkunitId: &networkUnitID,
			CpuArch:         p.CPUArch,
		}
	}
	x.Host = hosts

	versions := make([]*TargetVersion, len(targetVersions))
	for idx, v := range targetVersions {
		versions[idx] = &TargetVersion{
			Version: v.Version,
			CpuArch: string(v.CPUArch),
			OsType:  string(v.OsType),
		}
	}
	x.TargetVersion = versions
}

// ConvertResultFromTypes converts type results to proto response data.
func (x *NodeProxyUpgradeCheckResp) ConvertResultFromTypes(results []*types.NodeProxyUpgradeCheckResult) {
	items := make([]*NodeProxyUpgradeCheckResult, len(results))
	for idx, result := range results {
		item := &NodeProxyUpgradeCheckResult{
			Status: string(result.Status),
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

// ConvertResultToTypes converts proto response data to type results.
func (x *NodeProxyUpgradeCheckResp) ConvertResultToTypes() []*types.NodeProxyUpgradeCheckResult {
	if x.GetData() == nil {
		return nil
	}

	items := make([]*types.NodeProxyUpgradeCheckResult, len(x.GetData().GetResults()))
	for idx, result := range x.GetData().GetResults() {
		item := &types.NodeProxyUpgradeCheckResult{
			Status: types.NodeProxyUpgradeCheckStatus(result.GetStatus()),
		}

		if matched := result.GetMatched(); matched != nil {
			item.Matched = &types.NodeProxyUpgradeCheckMatchedItem{
				HostID:        matched.GetBkHostId(),
				BizID:         matched.GetBkBizId(),
				NetworkAreaID: matched.GetBkNetworkareaId(),
				NetworkUnitID: matched.GetBkNetworkunitId(),
				OsType:        criteria.OSType(matched.GetOsType()),
				NodeRole:      types.NodeRole(matched.GetNodeRole()),
				InnerIPList:   matched.GetBkHostInneripList(),
				InnerIPV6List: matched.GetBkHostInneripV6List(),
			}
		}

		items[idx] = item
	}

	return items
}

// AutoConvert auto convert.
func (x *NodeProxyUpdateOpsFieldsReq) AutoConvert() {}

// Validate validates the update ops fields request.
func (x *NodeProxyUpdateOpsFieldsReq) Validate() error {
	if len(x.GetHosts()) == 0 {
		return errors.New("hosts can not be empty")
	}

	for _, h := range x.GetHosts() {
		if h.GetBkHostId() <= 0 {
			return errors.New("bk_host_id must be greater than 0")
		}
	}

	return nil
}

// ConvertParamFromTypes converts update ops fields hosts from types.
func (x *NodeProxyUpdateOpsFieldsReq) ConvertParamFromTypes(hosts []*types.Host) {
	if hosts == nil {
		return
	}

	items := make([]*HostOpsInfo, len(hosts))
	for idx, host := range hosts {
		item := &HostOpsInfo{}
		if host != nil {
			item.BkHostId = host.HostID
			if host.Dynamic != nil {
				item.OpsConsoleHostId = host.Dynamic.OpsConsoleHostID
				item.OpsOutBandType = host.Dynamic.OpsOutBandType
				item.OpsOutBandProtocol = host.Dynamic.OpsOutBandProtocol
				item.OpsBmcIp = host.Dynamic.OpsBMCIP
				item.OpsBmcPort = host.Dynamic.OpsBMCPort
			}
		}
		items[idx] = item
	}

	x.Hosts = items
}

// ConvertHostToTypes converts update ops hosts to types hosts.
func (x *NodeProxyUpdateOpsFieldsReq) ConvertHostToTypes() []*types.Host {
	hosts := make([]*types.Host, 0, len(x.GetHosts()))
	for _, host := range x.GetHosts() {
		typeHost := &types.Host{
			HostID: host.GetBkHostId(),
			Dynamic: &types.HostDynamic{
				OpsConsoleHostID:   host.GetOpsConsoleHostId(),
				OpsOutBandType:     host.GetOpsOutBandType(),
				OpsOutBandProtocol: host.GetOpsOutBandProtocol(),
				OpsBMCIP:           host.GetOpsBmcIp(),
				OpsBMCPort:         host.GetOpsBmcPort(),
			},
		}

		hosts = append(hosts, typeHost)
	}

	return hosts
}

// GetHostIDs extracts the host IDs from the update ops fields request.
func (x *NodeProxyUpdateOpsFieldsReq) GetHostIDs() []int64 {
	hostIDs := conv.SliceToSlice(x.GetHosts(), func(h *HostOpsInfo) int64 {
		return h.GetBkHostId()
	})

	return hostIDs
}
