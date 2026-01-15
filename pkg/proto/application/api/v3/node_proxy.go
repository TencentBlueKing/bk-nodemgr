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
			ProxyTags:                types.StringListToProxyTagList(host.GetProxyTags()),
			ProxyInstallOriginUnitID: host.GetProxyInstallOriginUnitId(),
			CreditExpiredIntervalSec: host.GetCreditExpiredIntervalSec(),
			RelayDownloadPort:        host.GetRelayDownloadPort(),
			RelayCallbackPort:        host.GetRelayCallbackPort(),
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
		Hosts:         hostsParam,
		TargetVersion: targetVersion,
		IsManual:      x.GetIsManual(),
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
			Force:                  host.GetForce(),
			GracefulRestartTimeout: time.Duration(host.GetGracefulRestartTimeoutSec()) * time.Second,
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

	return &types.NodeProxyUpgradeParam{
		Hosts:         hostsParam,
		TargetVersion: targetVersion,
	}
}

// Validate check body.
// nolint: protogetter
func (x *NodeProxyUpgradeHost) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id must be >= 0")
	}

	if x.GetForce() && x.GetGracefulRestartTimeoutSec() <= 0 {
		return errors.New("graceful_restart_timeout_sec must be > 0 when force is true")
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeProxyUpgradeHost) AutoConvert() {
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

	if x.GetForce() && x.GetGracefulRestartTimeoutSec() <= 0 {
		return errors.New("graceful_restart_timeout_sec must be > 0 when force is true")
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

	if x.GetForce() && x.GetGracefulRestartTimeoutSec() <= 0 {
		return errors.New("graceful_restart_timeout_sec must be > 0 when force is true")
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

	if err := types.LoginMode(x.GetLoginMode()).Validate(); err != nil {
		return err
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
