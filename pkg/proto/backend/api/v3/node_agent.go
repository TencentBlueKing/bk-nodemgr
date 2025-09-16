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

	_, err := conv.SliceToMap(x.GetTargetVersion(), func(v *NodeAgentInstallReq_TargetVersion) string {
		return fmt.Sprintf("%s:%s", v.GetOsType(), v.GetCpuArch())
	})
	if err != nil {
		return err
	}

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
func (x *NodeAgentInstallReq_Host) Validate() error {
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
	hosts := x.GetHost()
	for idx := range hosts {
		hosts[idx].AutoConvert()
	}
}

// AutoConvert auto convert.
func (x *NodeAgentInstallReq_Host) AutoConvert() {
	if x.BkNetworkunitId == nil {
		x.BkNetworkunitId = new(int64)
		*x.BkNetworkunitId = -1
	}

	if x.BkBizId == nil {
		x.BkBizId = new(int64)
		*x.BkBizId = -1
	}

	if x.LoginPort == nil {
		x.LoginPort = new(int64)
		*x.LoginPort = -1
	}

	if x.BkHostId == nil {
		x.BkHostId = new(int64)
		*x.BkHostId = -1
	}
}

// ConvertWorkflowID convert workflow id.
func (x *NodeAgentInstallResp) ConvertWorkflowID(workflowID string) {
	x.Data = &NodeAgentInstallResp_Data{WorkflowId: workflowID}
}

// ConvertHostParamFromTypes convert host param from types.
func (x *NodeAgentInstallReq) ConvertHostParamFromTypes(installParam *types.NodeAgentInstallParam) {
	hostsParam := make([]*NodeAgentInstallReq_Host, len(installParam.NodeAgentInstallHosts))
	for idx, host := range installParam.NodeAgentInstallHosts {
		hostsParam[idx] = &NodeAgentInstallReq_Host{
			BkBizId:         &host.BizID,
			BkHostInnerip:   host.InnerIP,
			BkHostInneripV6: host.InnerIPV6,
			BkAddressing:    string(host.Addressing),
			LoginIp:         host.LoginIP,
			LoginPort:       &host.LoginPort,
			LoginUser:       host.LoginUser,
			LoginMode:       string(host.LoginMode),
			LoginPassword:   host.LoginPassword,
			LoginKeyFile:    host.LoginKeyFile,
			BkNetworkunitId: &host.NetworkUnitID,
			OsType:          host.OSType,
		}
	}

	targetVersion := make([]*NodeAgentInstallReq_TargetVersion, len(installParam.NodeInstallTargetVersion))
	for idx, version := range installParam.NodeInstallTargetVersion {
		targetVersion[idx] = &NodeAgentInstallReq_TargetVersion{
			Version: version.Version,
			CpuArch: string(version.CPUArch),
			OsType:  string(version.OsType),
		}
	}

	x.TargetVersion = targetVersion
	x.Host = hostsParam
	x.DisableDefaultTargetVersion = installParam.DisableDefaultTargetVersion
}

// ConvertResultToComm ...
func (x *NodeAgentInstallResp) ConvertResultToComm() string {
	return x.GetData().GetWorkflowId()
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

// Validate check body.
// nolint: protogetter
func (x *NodeAgentUpgradeReq_Host) Validate() error {
	if x.GetBkHostId() < 0 {
		return errors.New("bk_host_id must be equal or greater than 0")
	}

	if x.GetTargetVersion() == "" {
		return errors.New("target_version can not be empty")
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeAgentUpgradeReq) AutoConvert() {
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

	return nil
}

// AutoConvert auto convert.
func (x *NodeAgentReconfigReq) AutoConvert() {
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

	return nil
}

// AutoConvert auto convert.
func (x *NodeAgentRestartReq) AutoConvert() {
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

	for _, host := range hosts {
		if host.GetBkNetworkunitId() < 0 {
			return errors.New("bk_networkunit_id is required")
		}

		if host.GetBkHostInnerip() == "" {
			return errors.New("bk_innerip is required")
		}

		if host.GetBkBizId() < 0 {
			return errors.New("bk_biz_id is required")
		}
	}

	return nil
}

// AutoConvert auto convert.
func (x *NodeAgentInstallCheckReq) AutoConvert() {
}

// ConvertHostParamFromTypes convert host param from types.
func (x *NodeAgentInstallCheckReq) ConvertHostParamFromTypes(checkParam []*types.NodeAgentInstallCheckInfo) {
	hostsParam := make([]*NodeAgentInstallCheckReq_Host, len(checkParam))
	for idx, host := range checkParam {
		hostsParam[idx] = &NodeAgentInstallCheckReq_Host{
			BkBizId:         host.BizID,
			BkHostInnerip:   host.InnerIP,
			BkNetworkunitId: host.NetworkUnitID,
		}
	}

	x.Host = hostsParam
}

// ConvertResultFromTypes convert result from types.
func (x *NodeAgentInstallCheckResp) ConvertResultFromTypes(result []*types.NodeAgentInstallCheckResult, total int) {
	items := make([]*NodeAgentInstallEligibility, 0, total)
	for _, status := range result {
		item := &NodeAgentInstallEligibility{
			InnerIp:           status.InnerIP,
			EligibilityStatus: string(status.InstallEligibilitiy),
		}
		if status.DuplicateHostIDs != nil {
			item.DuplicateHostIds = status.DuplicateHostIDs
		}
		items = append(items, item)
	}

	x.Data = &NodeAgentInstallCheckResp_Data{
		TotalCount:    int64(total),
		Eligibilities: items,
	}
}

// ConvertResultToTypes convert result to types.
func (x *NodeAgentInstallCheckResp) ConvertResultToTypes() []*types.NodeAgentInstallCheckResult {
	if x.GetData() == nil {
		return nil
	}

	data := x.GetData()

	items := make([]*types.NodeAgentInstallCheckResult, len(data.GetEligibilities()))
	for idx, item := range data.GetEligibilities() {
		items[idx] = &types.NodeAgentInstallCheckResult{
			InnerIP:             item.GetInnerIp(),
			InstallEligibilitiy: types.NodeAgentInstallEligibility(item.GetEligibilityStatus()),
			DuplicateHostIDs:    item.GetDuplicateHostIds(),
		}
	}

	return items
}
