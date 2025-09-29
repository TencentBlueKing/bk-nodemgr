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
func (x *NodeAgentInstallReq) Validate() error {
	switch {
	case x.GetDisableDefaultTargetVersion() && len(x.GetTargetVersion()) == 0:
		return errors.New("target_version can not be empty when disable_default_target_version is true")
	case !x.GetDisableDefaultTargetVersion() && len(x.GetTargetVersion()) > 0:
		return errors.New("target_version can not be set when disable_default_target_version is false")
	}

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
		NodeAgentInstallHosts:       hostsParam,
		NodeInstallTargetVersion:    targetVersion,
		DisableDefaultTargetVersion: x.GetDisableDefaultTargetVersion(),
	}
}

// ConvertWorkflowID convert workflow id.
func (x *NodeAgentInstallResp) ConvertWorkflowID(workflowID string) {
	x.Data = &NodeAgentInstallResp_Data{WorkflowId: workflowID}
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
// nolint: protogetter
func (x *AgentInstallCheckInfo) Validate() error {
	if x.BkHostInnerip == "" {
		return errors.New("bk_innerip can not be empty")
	}

	if x.GetBkBizId() < 0 {
		return errors.New("biz_id must be equal or greater than 0")
	}

	if x.GetBkNetworkunitId() < 0 {
		return errors.New("network_unit_id must be greater than or equal to 0")
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeAgentInstallCheckReq) AutoConvert() {
}

// ConvertAgentParamToTypes convert to types.
func (x *NodeAgentInstallCheckReq) ConvertAgentParamToTypes() []*types.NodeAgentInstallCheckInfo {
	infos := x.GetHost()
	if infos == nil {
		return nil
	}

	infoParam := make([]*types.NodeAgentInstallCheckInfo, len(infos))

	for idx, info := range infos {
		infoParam[idx] = &types.NodeAgentInstallCheckInfo{
			BizID:         info.GetBkBizId(),
			InnerIP:       info.GetBkHostInnerip(),
			NetworkUnitID: info.GetBkNetworkunitId(),
		}
	}

	return infoParam
}

// ConvertResultFromTypes convert result from types.
func (x *NodeAgentInstallCheckResp) ConvertResultFromTypes(result []*types.NodeAgentInstallCheckResult, num int) {
	if result == nil {
		return
	}

	installElig := make([]*NodeAgentInstallElig, len(result))
	for idx, status := range result {
		item := &NodeAgentInstallElig{
			InnerIp:    status.InnerIP,
			EligStatus: string(status.InstallEligibilitiy),
		}
		if status.DuplicateHostIDs != nil {
			item.DuplicateHostIds = status.DuplicateHostIDs
		}
		installElig[idx] = item
	}

	x.Data = &NodeAgentInstallCheckResp_Data{
		InstallEligs: installElig,
		TotalCount:   int64(num),
	}
}

// AutoConvert auto convert.
func (x *UploadAgentTemplateReq) AutoConvert() {
}

// Validate check body.
func (x *UploadAgentTemplateReq) Validate() error {
	return nil
}
