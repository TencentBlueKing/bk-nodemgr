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
			ProxyTags:                types.ProxyTagListToStringList(host.ProxyTags),
			ProxyInstallOriginUnitId: host.ProxyInstallOriginUnitID,
			CreditExpiredIntervalSec: host.CreditExpiredIntervalSec,
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
}

// Validate check body.
// nolint: protogetter
func (x *NodeProxyInstallHost) Validate() error {
	if x.GetBkHostInnerip() == "" && x.GetBkHostInneripV6() == "" {
		return errors.New("bk_innerip and bk_inneripv6 can not be empty at the same time")
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
		hostsParam[idx] = &NodeProxyUpgradeReq_Host{
			BkHostId:                  host.HostID,
			Force:                     host.Force,
			GracefulRestartTimeoutSec: int64(host.GracefulRestartTimeout.Seconds()),
		}
	}

	targetVersion := make([]*TargetVersion, len(upgradeParam.TargetVersion))
	for idx, version := range upgradeParam.TargetVersion {
		targetVersion[idx] = &TargetVersion{
			Version: version.Version,
			CpuArch: string(version.CPUArch),
			OsType:  string(version.OsType),
		}
	}

	x.TargetVersion = targetVersion
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

	if x.GetForce() && x.GetGracefulRestartTimeoutSec() <= 0 {
		return errors.New("graceful_restart_timeout_sec must be > 0 when force is true")
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeProxyUpgradeReq_Host) AutoConvert() {
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

	if x.GetForce() && x.GetGracefulRestartTimeoutSec() <= 0 {
		return errors.New("graceful_restart_timeout_sec must be > 0 when force is true")
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

	if x.GetForce() && x.GetGracefulRestartTimeoutSec() <= 0 {
		return errors.New("graceful_restart_timeout_sec must be > 0 when force is true")
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
				LoginIP:     host.GetLoginIp(),
				LoginPort:   host.GetLoginPort(),
				LoginUser:   host.GetLoginUser(),
				LoginMode:   types.LoginMode(host.GetLoginMode()),
				ExportIP:    host.GetExportIp(),
				AdvertiseIP: host.GetAdvertiseIp(),
				ProxyTags:   types.StringListToProxyTagList(host.GetProxyTags()),
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
			BkHostId:    host.HostID,
			LoginIp:     host.LoginIP,
			LoginPort:   host.LoginPort,
			LoginUser:   host.LoginUser,
			LoginMode:   string(host.LoginMode),
			ExportIp:    host.ExportIP,
			AdvertiseIp: host.AdvertiseIP,
			ProxyTags:   types.ProxyTagListToStringList(host.ProxyTags),
		}
	}

	x.Host = hostsParam
}

// ConvertHostFieldsToTypes convert host dynamic fields to types.
func (x *NodeProxyUpdateReq) ConvertHostFieldsToTypes() types.HostDynamicFields {
	return types.HostDynamicFields{
		LoginIP:     true,
		LoginPort:   true,
		LoginUser:   true,
		LoginMode:   true,
		ExportIP:    true,
		AdvertiseIP: true,
		ProxyTags:   true,
	}
}

// Validate check body.
// nolint: protogetter
func (x *NodeProxyUpdateHost) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id must be >= 0")
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
