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
func (x *NodeProxyInstallReq_Host) Validate() error {
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
func (x *NodeProxyInstallReq_Host) AutoConvert() {
	if x.BkHostId == nil {
		x.BkHostId = new(int64)
		*x.BkHostId = -1
	}
}

// ConvertAgentParamToTypes convert to types.
func (x *NodeProxyInstallReq) ConvertAgentParamToTypes() *types.NodeProxyInstallParam {
	hosts := x.GetHost()
	if hosts == nil {
		return nil
	}

	hostsParam := make([]*types.NodeProxyInstallHost, len(hosts))

	for idx, host := range hosts {
		hostsParam[idx] = &types.NodeProxyInstallHost{
			BizID:         host.GetBkBizId(),
			NetworkUnitID: host.GetBkNetworkunitId(),
			HostID:        host.GetBkHostId(),
			Addressing:    types.Addressing(host.GetBkAddressing()),
			InnerIP:       host.GetBkHostInnerip(),
			InnerIPV6:     host.GetBkHostInneripV6(),
			OSType:        host.GetOsType(),
			LoginIP:       host.GetLoginIp(),
			LoginPort:     int64(host.GetLoginPort()),
			LoginUser:     host.GetLoginUser(),
			LoginMode:     types.LoginMode(host.GetLoginMode()),
			LoginPassword: host.GetLoginPassword(),
			LoginKeyFile:  host.GetLoginKeyFile(),
			ExportIP:      host.GetExportIp(),
			AdvertiseIP:   host.GetAdvertiseIp(),
			ReRegister:    host.GetReRegister(),
			ProxyTags:     types.StringListToProxyTagList(host.GetProxyTags()),
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
